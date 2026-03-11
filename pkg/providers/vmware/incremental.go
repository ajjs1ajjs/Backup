// Package vmware provides VMware vSphere integration for NovaBackup
package vmware

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
	"go.uber.org/zap"
)

// IncrementalBackupEngine handles CBT-based incremental backups
type IncrementalBackupEngine struct {
	logger    *zap.Logger
	client    *Client
	cbtMgr    *CBTManager
	backupEng *BackupEngine
	stateDir  string
}

// BackupState tracks the backup state for incremental operations
type BackupState struct {
	VMName           string                `json:"vm_name"`
	VMUUID           string                `json:"vm_uuid"`
	BackupID         string                `json:"backup_id"`
	BackupType       string                `json:"backup_type"` // full, incremental
	Timestamp        time.Time             `json:"timestamp"`
	DiskChangeIDs    map[int32]string      `json:"disk_change_ids"`
	SnapshotName     string                `json:"snapshot_name"`
	BaseBackupID     string                `json:"base_backup_id,omitempty"` // For incremental
}

// NewIncrementalBackupEngine creates a new incremental backup engine
func NewIncrementalBackupEngine(client *Client, stateDir string) *IncrementalBackupEngine {
	return &IncrementalBackupEngine{
		logger:    client.logger.With(zap.String("component", "incremental-backup")),
		client:    client,
		cbtMgr:    NewCBTManager(client),
		backupEng: NewBackupEngine(client),
		stateDir:  stateDir,
	}
}

// PerformIncrementalBackup performs an incremental backup using CBT
func (i *IncrementalBackupEngine) PerformIncrementalBackup(ctx context.Context, vm *VM, config *BackupConfig, callback BackupProgressCallback) (*BackupResult, error) {
	result := &BackupResult{
		BackupID:  generateBackupID(),
		VMName:    vm.GetName(),
		StartTime: time.Now(),
		Status:    "in_progress",
	}

	i.logger.Info("Starting incremental backup",
		zap.String("backup_id", result.BackupID),
		zap.String("vm", vm.GetName()))

	// Phase 1: Check CBT status
	if callback != nil {
		callback(BackupProgress{
			Phase:   "checking_cbt",
			Percent: 5,
			Message: "Checking CBT status...",
		})
	}

	cbtStatus, err := i.cbtMgr.GetCBTStatus(ctx, vm)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("failed to get CBT status: %v", err)
		return result, err
	}

	if !cbtStatus.CBTEnabled {
		i.logger.Info("CBT not enabled, enabling now...")
		if err := i.cbtMgr.EnableCBTForVM(ctx, vm); err != nil {
			result.Status = "failed"
			result.Error = fmt.Sprintf("failed to enable CBT: %v", err)
			return result, err
		}
		// Refresh CBT status
		cbtStatus, _ = i.cbtMgr.GetCBTStatus(ctx, vm)
	}

	// Phase 2: Load previous backup state
	if callback != nil {
		callback(BackupProgress{
			Phase:   "loading_state",
			Percent: 10,
			Message: "Loading previous backup state...",
		})
	}

	previousState, err := i.loadBackupState(vm.GetName())
	if err != nil {
		i.logger.Info("No previous backup found, performing full backup")
		// No previous backup - do full backup
		return i.performFullWithCBT(ctx, vm, config, callback, result)
	}

	// Phase 3: Create snapshot
	if callback != nil {
		callback(BackupProgress{
			Phase:   "snapshotting",
			Percent: 15,
			Message: "Creating snapshot...",
		})
	}

	snapshotName := fmt.Sprintf("NovaBackup-Inc-%s", result.BackupID)
	snapshotTask, err := vm.CreateSnapshot(snapshotName, "NovaBackup incremental snapshot", config.Memory, config.Quiesce)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("failed to create snapshot: %v", err)
		return result, err
	}

	if err := snapshotTask.Wait(ctx); err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("snapshot creation failed: %v", err)
		return result, err
	}

	result.SnapshotName = snapshotName

	// Phase 4: Query changed blocks
	if callback != nil {
		callback(BackupProgress{
			Phase:   "querying_changes",
			Percent: 20,
			Message: "Querying changed blocks...",
		})
	}

	changeInfos, err := i.cbtMgr.QueryAllDiskChanges(ctx, vm, previousState.DiskChangeIDs)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("failed to query changes: %v", err)
		i.backupEng.cleanupSnapshot(ctx, vm, snapshotName)
		return result, err
	}

	// Calculate total changed bytes
	var totalChangedBytes int64
	for _, info := range changeInfos {
		totalChangedBytes += info.TotalChangedBytes
	}

	i.logger.Info("Changed blocks calculated",
		zap.Int("disks", len(changeInfos)),
		zap.Int64("total_changed_bytes", totalChangedBytes))

	// Phase 5: Export changed blocks
	if callback != nil {
		callback(BackupProgress{
			Phase:       "exporting",
			Percent:   25,
			BytesTotal:  totalChangedBytes,
			Message:    fmt.Sprintf("Exporting %d changed blocks...", len(changeInfos)),
		})
	}

	// Export only changed blocks
	err = i.exportChangedBlocks(ctx, vm, changeInfos, config, result, callback)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("failed to export changed blocks: %v", err)
		i.backupEng.cleanupSnapshot(ctx, vm, snapshotName)
		return result, err
	}

	// Phase 6: Save new state
	if callback != nil {
		callback(BackupProgress{
			Phase:   "saving_state",
			Percent: 95,
			Message: "Saving backup state...",
		})
	}

	newState := &BackupState{
		VMName:        vm.GetName(),
		VMUUID:        result.VMUUID,
		BackupID:      result.BackupID,
		BackupType:    "incremental",
		Timestamp:     time.Now(),
		SnapshotName:  snapshotName,
		BaseBackupID:  previousState.BackupID,
		DiskChangeIDs: make(map[int32]string),
	}

	// Store new change IDs
	for diskKey, info := range changeInfos {
		newState.DiskChangeIDs[diskKey] = info.ChangeID
	}

	if err := i.saveBackupState(newState); err != nil {
		i.logger.Warn("Failed to save backup state", zap.Error(err))
	}

	// Phase 7: Cleanup
	if callback != nil {
		callback(BackupProgress{
			Phase:   "completing",
			Percent: 98,
			Message: "Cleaning up...",
		})
	}

	// Remove snapshot
	if err := i.backupEng.cleanupSnapshot(ctx, vm, snapshotName); err != nil {
		i.logger.Warn("Failed to cleanup snapshot", zap.Error(err))
	}

	// Finalize result
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Status = "success"
	result.ChangeID = result.BackupID // Use current backup as next baseline

	if callback != nil {
		callback(BackupProgress{
			Phase:   "completed",
			Percent: 100,
			Message: "Incremental backup completed successfully",
		})
	}

	i.logger.Info("Incremental backup completed",
		zap.String("backup_id", result.BackupID),
		zap.Duration("duration", result.Duration),
		zap.Int64("changed_bytes", totalChangedBytes))

	return result, nil
}

// performFullWithCBT performs a full backup but saves CBT state for future incrementals
func (i *IncrementalBackupEngine) performFullWithCBT(ctx context.Context, vm *VM, config *BackupConfig, callback BackupProgressCallback, result *BackupResult) (*BackupResult, error) {
	i.logger.Info("Performing full backup with CBT baseline")

	// Do regular full backup
	result, err := i.backupEng.FullBackup(ctx, vm, config, callback)
	if err != nil {
		return result, err
	}

	// Get CBT status to establish baseline
	cbtStatus, err := i.cbtMgr.GetCBTStatus(ctx, vm)
	if err != nil {
		i.logger.Warn("Failed to get CBT status after full backup", zap.Error(err))
		return result, nil // Return success anyway
	}

	// Save initial state
	state := &BackupState{
		VMName:        vm.GetName(),
		VMUUID:        result.VMUUID,
		BackupID:      result.BackupID,
		BackupType:    "full",
		Timestamp:     time.Now(),
		SnapshotName:  result.SnapshotName,
		DiskChangeIDs: make(map[int32]string),
	}

	for _, disk := range cbtStatus.Disks {
		state.DiskChangeIDs[disk.DiskKey] = disk.CurrentChangeID
	}

	if err := i.saveBackupState(state); err != nil {
		i.logger.Warn("Failed to save initial backup state", zap.Error(err))
	}

	return result, nil
}

// exportChangedBlocks exports only changed blocks from disks
func (i *IncrementalBackupEngine) exportChangedBlocks(ctx context.Context, vm *VM, changeInfos map[int32]*DiskChangeInfo, config *BackupConfig, result *BackupResult, callback BackupProgressCallback) error {
	// Create backup directory
	backupDir := filepath.Join(config.Destination, result.BackupID)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	diskCount := 0
	totalDisks := len(changeInfos)

	for diskKey, changeInfo := range changeInfos {
		diskCount++

		if callback != nil {
			callback(BackupProgress{
				Phase:          "exporting",
				Percent:       25 + (float64(diskCount) / float64(totalDisks) * 70),
				BytesProcessed: result.ProcessedBytes,
				BytesTotal:     result.TotalBytes,
				CurrentDisk:    changeInfo.DiskName,
				DiskNumber:     diskCount,
				TotalDisks:     totalDisks,
				Message:        fmt.Sprintf("Exporting disk %d/%d: %s (%d changed areas)", diskCount, totalDisks, changeInfo.DiskName, len(changeInfo.ChangedAreas)),
			})
		}

		// Create incremental disk file
		diskFileName := fmt.Sprintf("disk_%d_inc.vmdk", diskKey)
		diskPath := filepath.Join(backupDir, diskFileName)

		// Export changed blocks for this disk
		err := i.exportDiskChanges(ctx, vm, diskKey, changeInfo, diskPath)
		if err != nil {
			return fmt.Errorf("failed to export disk %d: %w", diskKey, err)
		}

		diskInfo := DiskBackupInfo{
			DiskName:         changeInfo.DiskName,
			ChangedBlocks:    len(changeInfo.ChangedAreas),
			ProcessedGB:     float64(changeInfo.TotalChangedBytes) / (1024 * 1024 * 1024),
		}

		result.Disks = append(result.Disks, diskInfo)
		result.ProcessedBytes += changeInfo.TotalChangedBytes
	}

	// Save incremental backup metadata
	result.BackupFile = backupDir
	result.MetadataFile = filepath.Join(backupDir, "incremental_metadata.json")

	return nil
}

// exportDiskChanges exports only changed areas of a disk
func (i *IncrementalBackupEngine) exportDiskChanges(ctx context.Context, vm *VM, diskKey int32, changeInfo *DiskChangeInfo, destination string) error {
	// Create output file
	file, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Write changed areas info header
	header := map[string]interface{}{
		"disk_key":        diskKey,
		"disk_name":       changeInfo.DiskName,
		"change_id":       changeInfo.ChangeID,
		"previous_change_id": changeInfo.PreviousChangeID,
		"changed_areas":   changeInfo.ChangedAreas,
		"total_changed_bytes": changeInfo.TotalChangedBytes,
	}

	headerData, _ := json.Marshal(header)
	if _, err := file.Write(headerData); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}
	file.Write([]byte("\n"))

	// TODO: Implement actual block-level disk export
	// This would use NFC lease to download specific changed blocks
	// For now, we just save the change info

	i.logger.Debug("Exported disk changes",
		zap.Int32("disk_key", diskKey),
		zap.String("destination", destination),
		zap.Int("changed_areas", len(changeInfo.ChangedAreas)))

	return nil
}

// loadBackupState loads the previous backup state for a VM
func (i *IncrementalBackupEngine) loadBackupState(vmName string) (*BackupState, error) {
	stateFile := filepath.Join(i.stateDir, fmt.Sprintf("%s_state.json", vmName))

	data, err := os.ReadFile(stateFile)
	if err != nil {
		return nil, err
	}

	var state BackupState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	return &state, nil
}

// saveBackupState saves the backup state for future incrementals
func (i *IncrementalBackupEngine) saveBackupState(state *BackupState) error {
	if err := os.MkdirAll(i.stateDir, 0755); err != nil {
		return err
	}

	stateFile := filepath.Join(i.stateDir, fmt.Sprintf("%s_state.json", state.VMName))

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(stateFile, data, 0644)
}

// deleteBackupState removes the backup state for a VM
func (i *IncrementalBackupEngine) deleteBackupState(vmName string) error {
	stateFile := filepath.Join(i.stateDir, fmt.Sprintf("%s_state.json", vmName))
	return os.Remove(stateFile)
}

// GetLastBackupState returns the last backup state for a VM
func (i *IncrementalBackupEngine) GetLastBackupState(vmName string) (*BackupState, error) {
	return i.loadBackupState(vmName)
}

// CanPerformIncremental checks if an incremental backup is possible
func (i *IncrementalBackupEngine) CanPerformIncremental(vmName string) bool {
	state, err := i.loadBackupState(vmName)
	if err != nil {
		return false
	}

	// Check if state is not too old (e.g., 7 days)
	if time.Since(state.Timestamp) > 7*24*time.Hour {
		return false
	}

	return true
}

// ResetBackupState resets the backup state for a VM (force full backup next time)
func (i *IncrementalBackupEngine) ResetBackupState(vmName string) error {
	return i.deleteBackupState(vmName)
}

// CompactIncrementalChain compacts incremental backups into a full backup
func (i *IncrementalBackupEngine) CompactIncrementalChain(ctx context.Context, vmName string, fullBackupPath string, incrementalBackups []string, destination string) error {
	i.logger.Info("Compacting incremental chain",
		zap.String("vm", vmName),
		zap.Int("incrementals", len(incrementalBackups)))

	// TODO: Implement chain compaction
	// 1. Start from full backup
	// 2. Apply each incremental in order
	// 3. Create new consolidated full backup

	return fmt.Errorf("chain compaction not yet implemented")
}
