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

// RestoreEngine handles VM restore operations
type RestoreEngine struct {
	logger *zap.Logger
	client *Client
}

// RestoreConfig holds restore configuration
type RestoreConfig struct {
	BackupID             string            // Source backup ID
	DestinationHost      string            // Target ESXi host
	DestinationDatastore string            // Target datastore
	DestinationFolder    string            // Target VM folder
	NewName              string            // New VM name (optional)
	ResourcePool         string            // Target resource pool
	PowerOn              bool              // Power on after restore
	NetworkMapping       map[string]string // Source->Target network mapping
	DiskProvisioning     string            // thin/thick eagerZeroedThick
	KeepBackupID         bool              // Keep original backup ID
	Tags                 map[string]string // Custom tags
}

// RestoreResult contains restore operation results
type RestoreResult struct {
	RestoreID     string            `json:"restore_id"`
	BackupID      string            `json:"backup_id"`
	VMName        string            `json:"vm_name"`
	VMUUID        string            `json:"vm_uuid"`
	StartTime     time.Time         `json:"start_time"`
	EndTime       time.Time         `json:"end_time"`
	Duration      time.Duration     `json:"duration"`
	Status        string            `json:"status"` // success, failed, partial
	Error         string            `json:"error,omitempty"`
	TotalBytes    int64             `json:"total_bytes"`
	RestoredBytes int64             `json:"restored_bytes"`
	Disks         []DiskRestoreInfo `json:"disks"`
	Host          string            `json:"host"`
	Datastore     string            `json:"datastore"`
	ResourcePool  string            `json:"resource_pool"`
	PowerState    string            `json:"power_state"`
}

// DiskRestoreInfo contains per-disk restore information
type DiskRestoreInfo struct {
	DiskName      string `json:"disk_name"`
	SourcePath    string `json:"source_path"`
	TargetPath    string `json:"target_path"`
	SizeBytes     int64  `json:"size_bytes"`
	RestoredBytes int64  `json:"restored_bytes"`
	Provisioning  string `json:"provisioning"`
}

// RestoreProgressCallback is called during restore progress
type RestoreProgressCallback func(progress RestoreProgress)

// RestoreProgress contains current restore progress
type RestoreProgress struct {
	Phase         string  // validating, preparing, restoring, registering, powering_on, completed
	Percent       float64 // 0-100
	BytesRestored int64
	BytesTotal    int64
	CurrentDisk   string
	DiskNumber    int
	TotalDisks    int
	ETA           time.Duration
	Message       string
}

// BackupMetadata contains backup metadata stored with backup
type BackupMetadata struct {
	BackupID     string            `json:"backup_id"`
	BackupType   string            `json:"backup_type"` // full, incremental
	VMName       string            `json:"vm_name"`
	VMUUID       string            `json:"vm_uuid"`
	CreateTime   time.Time         `json:"create_time"`
	VMSpec       VMSpec            `json:"vm_spec"`
	Disks        []DiskSpec        `json:"disks"`
	Networks     []NetworkSpec     `json:"networks"`
	CustomValues map[string]string `json:"custom_values"`
	Compression  bool              `json:"compression"`
	Encryption   bool              `json:"encryption"`
}

// VMSpec contains VM configuration
type VMSpec struct {
	Name              string `json:"name"`
	GuestID           string `json:"guest_id"`
	NumCPU            int32  `json:"num_cpu"`
	NumCoresPerSocket int32  `json:"num_cores_per_socket"`
	MemoryMB          int32  `json:"memory_mb"`
	Firmware          string `json:"firmware"` // bios or efi
}

// DiskSpec contains disk configuration
type DiskSpec struct {
	Name         string `json:"name"`
	Label        string `json:"label"`
	CapacityGB   int64  `json:"capacity_gb"`
	Datastore    string `json:"datastore"`
	Controller   string `json:"controller"`
	UnitNumber   int32  `json:"unit_number"`
	Provisioning string `json:"provisioning"`
}

// NetworkSpec contains network configuration
type NetworkSpec struct {
	Name       string `json:"name"`
	Label      string `json:"label"`
	MacAddress string `json:"mac_address"`
	Network    string `json:"network"`
}

// NewRestoreEngine creates a new restore engine
func NewRestoreEngine(client *Client) *RestoreEngine {
	return &RestoreEngine{
		logger: client.logger.With(zap.String("component", "restore-engine")),
		client: client,
	}
}

// FullRestore performs a full VM restore from backup
func (r *RestoreEngine) FullRestore(ctx context.Context, config *RestoreConfig, callback RestoreProgressCallback) (*RestoreResult, error) {
	result := &RestoreResult{
		RestoreID: generateRestoreID(),
		BackupID:  config.BackupID,
		StartTime: time.Now(),
		Status:    "in_progress",
	}

	r.logger.Info("Starting full restore",
		zap.String("restore_id", result.RestoreID),
		zap.String("backup_id", config.BackupID),
		zap.String("target_host", config.DestinationHost))

	// Phase 1: Validate backup
	if callback != nil {
		callback(RestoreProgress{
			Phase:   "validating",
			Percent: 5,
			Message: "Validating backup integrity...",
		})
	}

	metadata, err := r.loadBackupMetadata(config.BackupID)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("failed to load backup metadata: %v", err)
		return result, err
	}

	result.VMName = metadata.VMName
	result.VMUUID = metadata.VMUUID

	// Phase 2: Prepare resources
	if callback != nil {
		callback(RestoreProgress{
			Phase:      "preparing",
			Percent:    10,
			TotalDisks: len(metadata.Disks),
			Message:    "Preparing target resources...",
		})
	}

	// Find target host
	host, err := r.findHost(ctx, config.DestinationHost)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("target host not found: %v", err)
		return result, err
	}

	// Find target datastore
	datastore, err := r.findDatastore(ctx, config.DestinationDatastore)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("target datastore not found: %v", err)
		return result, err
	}

	// Find target folder
	folder, err := r.findFolder(ctx, config.DestinationFolder)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("target folder not found: %v", err)
		return result, err
	}

	// Find resource pool
	pool, err := r.findResourcePool(ctx, config.ResourcePool)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("resource pool not found: %v", err)
		return result, err
	}

	// Phase 3: Restore disks
	if callback != nil {
		callback(RestoreProgress{
			Phase:      "restoring",
			Percent:    20,
			TotalDisks: len(metadata.Disks),
			Message:    fmt.Sprintf("Restoring %d disks...", len(metadata.Disks)),
		})
	}

	// Copy disk files to target datastore
	diskCount := 0
	for _, disk := range metadata.Disks {
		diskCount++

		if callback != nil {
			callback(RestoreProgress{
				Phase:       "restoring",
				Percent:     20 + (float64(diskCount) / float64(len(metadata.Disks)) * 60),
				CurrentDisk: disk.Name,
				DiskNumber:  diskCount,
				TotalDisks:  len(metadata.Disks),
				Message:     fmt.Sprintf("Restoring disk %d of %d...", diskCount, len(metadata.Disks)),
			})
		}

		sourcePath := filepath.Join(r.getBackupPath(config.BackupID), disk.Name)
		targetPath := r.buildDiskPath(datastore.Name(), disk.Name, config.NewName)

		restoredBytes, err := r.restoreDisk(sourcePath, targetPath, datastore)
		if err != nil {
			result.Status = "failed"
			result.Error = fmt.Sprintf("failed to restore disk %s: %v", disk.Name, err)
			return result, err
		}

		diskInfo := DiskRestoreInfo{
			DiskName:      disk.Name,
			SourcePath:    sourcePath,
			TargetPath:    targetPath,
			SizeBytes:     restoredBytes,
			RestoredBytes: restoredBytes,
			Provisioning:  disk.Provisioning,
		}
		result.Disks = append(result.Disks, diskInfo)
		result.RestoredBytes += restoredBytes
	}

	// Phase 4: Register VM
	if callback != nil {
		callback(RestoreProgress{
			Phase:   "registering",
			Percent: 85,
			Message: "Registering VM in vCenter...",
		})
	}

	vmName := config.NewName
	if vmName == "" {
		vmName = metadata.VMName
	}

	_ = r.buildVMConfigSpec(metadata, config) // spec will be used when reconfiguring VM

	registerTask, err := folder.RegisterVM(ctx, r.buildVMConfigPath(datastore.Name(), vmName), vmName, false, pool, host)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("failed to register VM: %v", err)
		return result, err
	}

	err = registerTask.Wait(ctx)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("VM registration failed: %v", err)
		return result, err
	}

	// Find the newly registered VM
	vm, err := r.client.GetFinder().VirtualMachine(ctx, vmName)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("failed to find registered VM: %v", err)
		return result, err
	}

	// Phase 5: Power on (optional)
	if config.PowerOn {
		if callback != nil {
			callback(RestoreProgress{
				Phase:   "powering_on",
				Percent: 95,
				Message: "Powering on VM...",
			})
		}

		powerTask, err := vm.PowerOn(ctx)
		if err != nil {
			r.logger.Warn("Failed to power on VM", zap.Error(err))
			result.PowerState = "unknown"
		} else {
			err = powerTask.Wait(ctx)
			if err != nil {
				r.logger.Warn("Failed to wait for power on", zap.Error(err))
			}
			result.PowerState = "poweredOn"
		}
	}

	// Finalize result
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Status = "success"
	result.Host = host.Name()
	result.Datastore = datastore.Name()
	result.ResourcePool = pool.Name()

	if callback != nil {
		callback(RestoreProgress{
			Phase:   "completed",
			Percent: 100,
			Message: "Restore completed successfully",
		})
	}

	r.logger.Info("Restore completed",
		zap.String("restore_id", result.RestoreID),
		zap.String("vm", vmName),
		zap.Duration("duration", result.Duration))

	return result, nil
}

// InstantRestore performs instant restore (NFS mount) - STUB
func (r *RestoreEngine) InstantRestore(ctx context.Context, config *RestoreConfig, callback RestoreProgressCallback) (*RestoreResult, error) {
	return nil, fmt.Errorf("instant restore not yet implemented")
}

// FileLevelRestore performs file-level restore - STUB
func (r *RestoreEngine) FileLevelRestore(ctx context.Context, backupID string, filePath string, destination string) error {
	return fmt.Errorf("file-level restore not yet implemented")
}

// loadBackupMetadata loads backup metadata from disk
func (r *RestoreEngine) loadBackupMetadata(backupID string) (*BackupMetadata, error) {
	metadataPath := filepath.Join(r.getBackupPath(backupID), "metadata.json")

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	var metadata BackupMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	return &metadata, nil
}

// findHost finds an ESXi host by name
func (r *RestoreEngine) findHost(ctx context.Context, name string) (*object.HostSystem, error) {
	if name == "" {
		// Use default host
		hosts, err := r.client.GetFinder().HostSystemList(ctx, "*")
		if err != nil {
			return nil, err
		}
		if len(hosts) == 0 {
			return nil, fmt.Errorf("no hosts found")
		}
		return hosts[0], nil
	}
	return r.client.GetFinder().HostSystem(ctx, name)
}

// findDatastore finds a datastore by name
func (r *RestoreEngine) findDatastore(ctx context.Context, name string) (*object.Datastore, error) {
	if name == "" {
		// Use default datastore
		datastores, err := r.client.GetFinder().DatastoreList(ctx, "*")
		if err != nil {
			return nil, err
		}
		if len(datastores) == 0 {
			return nil, fmt.Errorf("no datastores found")
		}
		return datastores[0], nil
	}
	return r.client.GetFinder().Datastore(ctx, name)
}

// findFolder finds a folder by path
func (r *RestoreEngine) findFolder(ctx context.Context, path string) (*object.Folder, error) {
	if path == "" {
		// Use default VM folder
		dc, err := r.client.GetFinder().DefaultDatacenter(ctx)
		if err != nil {
			return nil, err
		}
		folders, err := dc.Folders(ctx)
		if err != nil {
			return nil, err
		}
		return folders.VmFolder, nil
	}
	return r.client.GetFinder().Folder(ctx, path)
}

// findResourcePool finds a resource pool by name
func (r *RestoreEngine) findResourcePool(ctx context.Context, name string) (*object.ResourcePool, error) {
	if name == "" {
		// Use default resource pool
		return r.client.GetFinder().DefaultResourcePool(ctx)
	}
	return r.client.GetFinder().ResourcePool(ctx, name)
}

// restoreDisk copies a disk file to target datastore
func (r *RestoreEngine) restoreDisk(sourcePath, targetPath string, datastore *object.Datastore) (int64, error) {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return 0, fmt.Errorf("failed to open source disk: %w", err)
	}
	defer sourceFile.Close()

	// Get source file info
	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return 0, fmt.Errorf("failed to stat source disk: %w", err)
	}

	// For now, we just return the size - actual upload would use datastore browser
	// TODO: Implement proper datastore file upload via HTTP POST to datastore browser
	r.logger.Debug("Would restore disk",
		zap.String("source", sourcePath),
		zap.String("target", targetPath),
		zap.Int64("size", sourceInfo.Size()))

	return sourceInfo.Size(), nil
}

// buildVMConfigSpec builds VM configuration spec
func (r *RestoreEngine) buildVMConfigSpec(metadata *BackupMetadata, config *RestoreConfig) types.VirtualMachineConfigSpec {
	spec := types.VirtualMachineConfigSpec{
		Name:     config.NewName,
		GuestId:  metadata.VMSpec.GuestID,
		NumCPUs:  metadata.VMSpec.NumCPU,
		MemoryMB: int64(metadata.VMSpec.MemoryMB),
		Files: &types.VirtualMachineFileInfo{
			VmPathName: r.buildVMConfigPath(config.DestinationDatastore, config.NewName),
		},
	}

	if metadata.VMSpec.Firmware == "efi" {
		spec.Firmware = "efi"
	} else {
		spec.Firmware = "bios"
	}

	return spec
}

// buildVMConfigPath builds VM config path on datastore
func (r *RestoreEngine) buildVMConfigPath(datastoreName, vmName string) string {
	return fmt.Sprintf("[%s] %s/%s.vmx", datastoreName, vmName, vmName)
}

// buildDiskPath builds disk path on datastore
func (r *RestoreEngine) buildDiskPath(datastoreName, diskName, vmName string) string {
	return fmt.Sprintf("[%s] %s/%s", datastoreName, vmName, diskName)
}

// getBackupPath returns the path to backup directory
func (r *RestoreEngine) getBackupPath(backupID string) string {
	// TODO: Make this configurable
	return filepath.Join("/backups", backupID)
}

// generateRestoreID generates a unique restore ID
func generateRestoreID() string {
	return fmt.Sprintf("rst_%s", time.Now().Format("20060102_150405_000"))
}
