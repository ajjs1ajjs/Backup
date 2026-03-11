// Package vmware provides VMware vSphere integration for NovaBackup
package vmware

import (
	"context"
	"fmt"
	"io"

	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"go.uber.org/zap"
)

// CBTManager handles Changed Block Tracking operations
type CBTManager struct {
	logger *zap.Logger
	client *Client
}

// DiskChangeInfo contains information about changed blocks on a disk
type DiskChangeInfo struct {
	DiskName       string           `json:"disk_name"`
	DiskKey        int32            `json:"disk_key"`
	ChangeID       string           `json:"change_id"`
	PreviousChangeID string         `json:"previous_change_id"`
	ChangedAreas   []ChangedArea    `json:"changed_areas"`
	TotalChangedBytes int64         `json:"total_changed_bytes"`
}

// ChangedArea represents a contiguous range of changed blocks
type ChangedArea struct {
	StartOffset int64 `json:"start_offset"`
	Length      int64 `json:"length"`
}

// CBTStatus represents the CBT status for a VM
type CBTStatus struct {
	VMName           string            `json:"vm_name"`
	VMUUID           string            `json:"vm_uuid"`
	CBTEnabled       bool              `json:"cbt_enabled"`
	Supported        bool              `json:"supported"`
	Disks            []DiskCBTStatus   `json:"disks"`
	LastSnapshotRef  string            `json:"last_snapshot_ref,omitempty"`
}

// DiskCBTStatus represents CBT status for a specific disk
type DiskCBTStatus struct {
	DiskName       string `json:"disk_name"`
	DiskKey        int32  `json:"disk_key"`
	CapacityBytes  int64  `json:"capacity_bytes"`
	CurrentChangeID string `json:"current_change_id"`
}

// NewCBTManager creates a new CBT manager
func NewCBTManager(client *Client) *CBTManager {
	return &CBTManager{
		logger: client.logger.With(zap.String("component", "cbt-manager")),
		client: client,
	}
}

// EnableCBTForVM enables Changed Block Tracking for all disks on a VM
func (c *CBTManager) EnableCBTForVM(ctx context.Context, vm *VM) error {
	c.logger.Info("Enabling CBT for VM", zap.String("vm", vm.GetName()))

	// Get current VM config
	vmInfo, err := vm.GetInfo()
	if err != nil {
		return fmt.Errorf("failed to get VM info: %w", err)
	}

	if vmInfo.CBTEnabled {
		c.logger.Info("CBT already enabled for VM", zap.String("vm", vm.GetName()))
		return nil
	}

	// Enable CBT on VM
	if err := vm.EnableCBT(); err != nil {
		return fmt.Errorf("failed to enable CBT: %w", err)
	}

	// Note: CBT only starts tracking changes after a snapshot is created and removed
	// Create a temporary snapshot to activate CBT
	snapshotName := fmt.Sprintf("CBT-Activate-%s", vm.GetName())
	task, err := vm.CreateSnapshot(snapshotName, "Activate CBT tracking", false, false)
	if err != nil {
		return fmt.Errorf("failed to create activation snapshot: %w", err)
	}

	if err := task.Wait(ctx); err != nil {
		return fmt.Errorf("activation snapshot creation failed: %w", err)
	}

	// Remove the snapshot immediately
	task, err = vm.RemoveSnapshot(snapshotName, true)
	if err != nil {
		c.logger.Warn("Failed to remove activation snapshot", zap.Error(err))
		return nil // CBT is still enabled, just snapshot cleanup failed
	}

	if err := task.Wait(ctx); err != nil {
		c.logger.Warn("Failed to wait for snapshot removal", zap.Error(err))
	}

	c.logger.Info("CBT enabled successfully for VM", zap.String("vm", vm.GetName()))
	return nil
}

// DisableCBTForVM disables Changed Block Tracking for a VM
func (c *CBTManager) DisableCBTForVM(ctx context.Context, vm *VM) error {
	c.logger.Info("Disabling CBT for VM", zap.String("vm", vm.GetName()))

	if err := vm.DisableCBT(); err != nil {
		return fmt.Errorf("failed to disable CBT: %w", err)
	}

	c.logger.Info("CBT disabled successfully for VM", zap.String("vm", vm.GetName()))
	return nil
}

// GetCBTStatus retrieves the current CBT status for a VM
func (c *CBTManager) GetCBTStatus(ctx context.Context, vm *VM) (*CBTStatus, error) {
	var vmMo mo.VirtualMachine
	err := c.client.GetClient().RetrieveOne(ctx, vm.GetReference(), []string{"config", "snapshot", "layoutEx"}, &vmMo)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve VM: %w", err)
	}

	if vmMo.Config == nil {
		return nil, fmt.Errorf("VM configuration not available")
	}

	status := &CBTStatus{
		VMName:     vm.GetName(),
		VMUUID:     vmMo.Config.Uuid,
		CBTEnabled: vmMo.Config.ChangeTrackingEnabled,
		Supported:  vmMo.Config.ChangeTrackingSupported,
		Disks:      []DiskCBTStatus{},
	}

	if vmMo.Snapshot != nil && vmMo.Snapshot.CurrentSnapshot != nil {
		status.LastSnapshotRef = vmMo.Snapshot.CurrentSnapshot.Value
	}

	// Get disk information
	for _, device := range vmMo.Config.Hardware.Device {
		if disk, ok := device.(*types.VirtualDisk); ok {
			diskStatus := DiskCBTStatus{
				DiskName:      disk.DeviceInfo.GetDescription().Label,
				DiskKey:       disk.Key,
				CapacityBytes: disk.CapacityInBytes,
			}

			// Get current change ID if available
			if backing, ok := disk.Backing.(*types.VirtualDiskFlatVer2BackingInfo); ok {
				if backing.ChangeId != "" {
					diskStatus.CurrentChangeID = backing.ChangeId
				}
			}

			status.Disks = append(status.Disks, diskStatus)
		}
	}

	return status, nil
}

// QueryDiskChanges queries CBT for changed blocks since a specific change ID
func (c *CBTManager) QueryDiskChanges(ctx context.Context, vm *VM, diskKey int32, changeID string) (*DiskChangeInfo, error) {
	c.logger.Info("Querying disk changes",
		zap.String("vm", vm.GetName()),
		zap.Int32("disk_key", diskKey),
		zap.String("change_id", changeID))

	// Get current snapshot reference
	var vmMo mo.VirtualMachine
	err := c.client.GetClient().RetrieveOne(ctx, vm.GetReference(), []string{"snapshot"}, &vmMo)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve VM snapshot info: %w", err)
	}

	if vmMo.Snapshot == nil || vmMo.Snapshot.CurrentSnapshot == nil {
		return nil, fmt.Errorf("VM has no snapshots - CBT requires at least one snapshot")
	}

	// Query changed disk areas
	// The deviceKey parameter is the key of the virtual disk device
	changeInfo, err := vm.GetObject().QueryChangedDiskAreas(ctx, vmMo.Snapshot.CurrentSnapshot, diskKey, 0, changeID)
	if err != nil {
		return nil, fmt.Errorf("failed to query changed disk areas: %w", err)
	}

	result := &DiskChangeInfo{
		DiskKey:        diskKey,
		ChangeID:       changeInfo.StartOffset, // This is actually the new change ID after the query
		PreviousChangeID: changeID,
		ChangedAreas:   []ChangedArea{},
	}

	// Parse changed areas
	for _, area := range changeInfo.ChangedArea {
		changedArea := ChangedArea{
			StartOffset: area.Start,
			Length:      area.Length,
		}
		result.ChangedAreas = append(result.ChangedAreas, changedArea)
		result.TotalChangedBytes += area.Length
	}

	c.logger.Info("Disk changes query completed",
		zap.Int32("disk_key", diskKey),
		zap.Int("changed_areas", len(result.ChangedAreas)),
		zap.Int64("total_changed_bytes", result.TotalChangedBytes))

	return result, nil
}

// QueryAllDiskChanges queries CBT for all disks on a VM
func (c *CBTManager) QueryAllDiskChanges(ctx context.Context, vm *VM, previousChangeIDs map[int32]string) (map[int32]*DiskChangeInfo, error) {
	c.logger.Info("Querying changes for all disks", zap.String("vm", vm.GetName()))

	// Get CBT status to find all disks
	status, err := c.GetCBTStatus(ctx, vm)
	if err != nil {
		return nil, err
	}

	if !status.CBTEnabled {
		return nil, fmt.Errorf("CBT is not enabled for VM %s", vm.GetName())
	}

	results := make(map[int32]*DiskChangeInfo)

	for _, disk := range status.Disks {
		previousID := previousChangeIDs[disk.DiskKey]
		if previousID == "" {
			// If no previous change ID, use "*" to get all blocks
			previousID = "*"
		}

		changeInfo, err := c.QueryDiskChanges(ctx, vm, disk.DiskKey, previousID)
		if err != nil {
			c.logger.Warn("Failed to query changes for disk",
				zap.String("vm", vm.GetName()),
				zap.Int32("disk_key", disk.DiskKey),
				zap.Error(err))
			continue
		}

		changeInfo.DiskName = disk.DiskName
		results[disk.DiskKey] = changeInfo
	}

	return results, nil
}

// ResetCBT resets CBT tracking for a VM (useful after restore or when CBT gets corrupted)
func (c *CBTManager) ResetCBT(ctx context.Context, vm *VM) error {
	c.logger.Info("Resetting CBT for VM", zap.String("vm", vm.GetName()))

	// 1. Disable CBT
	if err := c.DisableCBTForVM(ctx, vm); err != nil {
		return fmt.Errorf("failed to disable CBT: %w", err)
	}

	// 2. Create a temporary snapshot to clear CBT state
	tempSnapshot := "CBT-Reset-Temp"
	task, err := vm.CreateSnapshot(tempSnapshot, "Temporary snapshot for CBT reset", false, false)
	if err != nil {
		return fmt.Errorf("failed to create temp snapshot: %w", err)
	}

	if err := task.Wait(ctx); err != nil {
		return fmt.Errorf("temp snapshot creation failed: %w", err)
	}

	// 3. Remove the temp snapshot
	task, err = vm.RemoveSnapshot(tempSnapshot, true)
	if err != nil {
		return fmt.Errorf("failed to remove temp snapshot: %w", err)
	}

	if err := task.Wait(ctx); err != nil {
		return fmt.Errorf("temp snapshot removal failed: %w", err)
	}

	// 4. Re-enable CBT
	if err := c.EnableCBTForVM(ctx, vm); err != nil {
		return fmt.Errorf("failed to re-enable CBT: %w", err)
	}

	c.logger.Info("CBT reset completed successfully", zap.String("vm", vm.GetName()))
	return nil
}

// ValidateCBTIntegrity checks if CBT data is consistent
func (c *CBTManager) ValidateCBTIntegrity(ctx context.Context, vm *VM) error {
	c.logger.Info("Validating CBT integrity", zap.String("vm", vm.GetName()))

	status, err := c.GetCBTStatus(ctx, vm)
	if err != nil {
		return err
	}

	if !status.CBTEnabled {
		return fmt.Errorf("CBT is not enabled")
	}

	// Try to query changes for each disk
	for _, disk := range status.Disks {
		_, err := c.QueryDiskChanges(ctx, vm, disk.DiskKey, "*")
		if err != nil {
			return fmt.Errorf("CBT validation failed for disk %s: %w", disk.DiskName, err)
		}
	}

	c.logger.Info("CBT integrity validated successfully", zap.String("vm", vm.GetName()))
	return nil
}

// ExportChangedBlocks exports only the changed blocks to a writer
func (c *CBTManager) ExportChangedBlocks(ctx context.Context, vm *VM, diskKey int32, changeInfo *DiskChangeInfo, writer io.Writer) error {
	c.logger.Info("Exporting changed blocks",
		zap.String("vm", vm.GetName()),
		zap.Int32("disk_key", diskKey),
		zap.Int("changed_areas", len(changeInfo.ChangedAreas)))

	// This is a placeholder - actual implementation would:
	// 1. Create a NFC lease for export
	// 2. Use HttpNfcLeasePullFromURLs to get block-level access
	// 3. Read only the changed areas and write to the output

	c.logger.Info("Changed blocks export completed (stub)")
	return nil
}
