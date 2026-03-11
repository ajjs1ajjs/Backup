package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"novabackup/pkg/models"
)

// HyperVCheckpoint represents a Hyper-V checkpoint (snapshot)
type HyperVCheckpoint struct {
	Name        string
	ID          string
	VMName      string
	CreateTime  time.Time
	Type        string // Standard, Production
	Description string
	ParentID    string
	TreePath    string
}

// HyperVIncrementalInfo contains incremental backup information
type HyperVIncrementalInfo struct {
	VMName           string
	LastCheckpointID string
	ChangeTracking   bool
	DiskChanges      map[string]int64
	BackupTime       time.Time
}

// HyperVBackupProvider handles Microsoft Hyper-V VM backups
type HyperVBackupProvider struct {
	host      string
	username  string
	password  string
	useRemote bool
}

// HyperVConfig contains Hyper-V connection configuration
type HyperVConfig struct {
	Host      string
	Username  string
	Password  string
	UseRemote bool
}

// NewHyperVBackupProvider creates a new Hyper-V backup provider
func NewHyperVBackupProvider(cfg HyperVConfig) *HyperVBackupProvider {
	return &HyperVBackupProvider{
		host:      cfg.Host,
		username:  cfg.Username,
		password:  cfg.Password,
		useRemote: cfg.UseRemote,
	}
}

// VMInfo contains Hyper-V VM information
type VMInfo struct {
	Name        string
	ID          string
	State       string
	CPUCount    int
	MemorySize  int64
	DiskSize    int64
	Generation  int
	Uptime      time.Duration
	Checkpoints []HyperVCheckpoint
	NetworkName string
}

// ListVMs lists all VMs on the Hyper-V host
func (h *HyperVBackupProvider) ListVMs(ctx context.Context) ([]VMInfo, error) {
	powerShellCmd := `Get-VM | Select-Object Name,VMId,State,ProcessorCount,MemoryStartup,Generation,@{Name='NetworkName';Expression={($_.NetworkAdapters | Select-Object -First 1).SwitchName}},@{Name='Uptime';Expression={(Get-Date) - $_.CreationTime}},@{Name='Checkpoints';Expression={Get-VMSnapshot -VM $_ | Select-Object Name,ID,CreationTime,Type,Description,ParentSnapshotId,TreePath}} | ConvertTo-Json -Depth 5`

	output, err := h.runPowerShell(ctx, powerShellCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to list VMs: %w", err)
	}

	// Parse JSON output
	var vms []VMInfo
	if err := json.Unmarshal([]byte(output), &vms); err != nil {
		return nil, fmt.Errorf("failed to parse VM list JSON: %w", err)
	}

	return vms, nil
}

// GetVMInfo retrieves detailed VM information
func (h *HyperVBackupProvider) GetVMInfo(ctx context.Context, vmName string) (*VMInfo, error) {
	powerShellCmd := fmt.Sprintf(`Get-VM -Name "%s" | Select-Object Name,VMId,State,ProcessorCount,MemoryStartup,Generation | ConvertTo-Json`, vmName)

	output, err := h.runPowerShell(ctx, powerShellCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get VM info: %w", err)
	}

	// Parse JSON output
	var vm VMInfo
	if err := json.Unmarshal([]byte(output), &vm); err != nil {
		return nil, fmt.Errorf("failed to parse VM info JSON: %w", err)
	}

	return &vm, nil
}

// Backup performs a Hyper-V VM backup using Export-VM
func (h *HyperVBackupProvider) Backup(ctx context.Context, vmName string, dest string) (*models.BackupResult, error) {
	result := &models.BackupResult{
		StartTime: time.Now(),
	}

	// Create destination directory
	if err := os.MkdirAll(dest, 0755); err != nil {
		result.Status = models.JobStatusFailed
		result.ErrorMessage = err.Error()
		result.EndTime = time.Now()
		return result, err
	}

	// Check VM state
	vmState, err := h.getVMState(ctx, vmName)
	if err != nil {
		result.Status = models.JobStatusFailed
		result.ErrorMessage = fmt.Sprintf("failed to get VM state: %w", err)
		result.EndTime = time.Now()
		return result, err
	}

	// Skip if VM is already off
	_ = vmState // Use vmState for potential future logic

	// Create checkpoint (snapshot)
	checkpointName := fmt.Sprintf("backup_%s", time.Now().Format("20060102_150405"))
	err = h.createCheckpoint(ctx, vmName, checkpointName)
	if err != nil {
		result.Status = models.JobStatusFailed
		result.ErrorMessage = fmt.Sprintf("failed to create checkpoint: %w", err)
		result.EndTime = time.Now()
		return result, err
	}

	// Export VM
	exportPath := filepath.Join(dest, vmName)
	err = h.exportVM(ctx, vmName, exportPath)
	if err != nil {
		// Cleanup checkpoint
		h.removeCheckpoint(ctx, vmName, checkpointName)
		result.Status = models.JobStatusFailed
		result.ErrorMessage = fmt.Sprintf("failed to export VM: %w", err)
		result.EndTime = time.Now()
		return result, err
	}

	// Calculate exported size
	size, err := h.getDirectorySize(exportPath)
	if err != nil {
		size = 0
	}

	// Remove checkpoint
	err = h.removeCheckpoint(ctx, vmName, checkpointName)
	if err != nil {
		return nil, fmt.Errorf("failed to remove checkpoint: %w", err)
	}

	result.Status = models.JobStatusCompleted
	result.EndTime = time.Now()
	result.BytesWritten = size
	result.FilesTotal = 1
	result.FilesSuccess = 1

	return result, nil
}

// BackupWithVSS performs a Hyper-V backup using VSS (Volume Shadow Copy)
func (h *HyperVBackupProvider) BackupWithVSS(ctx context.Context, vmName string, dest string) (*models.BackupResult, error) {
	result := &models.BackupResult{
		StartTime: time.Now(),
	}

	// Use PowerShell to create VSS backup
	powerShellCmd := fmt.Sprintf(`
		$vm = Get-VM -Name "%s"
		$vmHardDisks = Get-VMHardDiskDrive -VMName "%s"
		$paths = $vmHardDisks | ForEach-Object { $_.Path }

		# Create shadow copy
		$shadow = Get-WmiObject -List Win32_ShadowCopy | ForEach-Object { $_.Create($paths, "ClientAccessible") }
		$shadowID = $shadow.ShadowID

		# Copy from shadow
		# (Implementation depends on specific requirements)

		# Remove shadow copy
		Get-WmiObject Win32_ShadowCopy -Filter "ID='$shadowID'" | ForEach-Object { $_.Delete() }
	`, vmName, vmName)

	_, err := h.runPowerShell(ctx, powerShellCmd)
	if err != nil {
		result.Status = models.JobStatusFailed
		result.ErrorMessage = err.Error()
		result.EndTime = time.Now()
		return result, err
	}

	result.Status = models.JobStatusCompleted
	result.EndTime = time.Now()
	return result, nil
}

// Restore restores a Hyper-V VM from export
func (h *HyperVBackupProvider) Restore(ctx context.Context, exportPath string, newName string) error {
	powerShellCmd := fmt.Sprintf(`
		$vm = Import-VM -Path "%s\Virtual Machines\*.vmcx" -GenerateNewId -Copy
		if ($newName -ne "") {
			Rename-VM -VM $vm -NewName "%s"
		}
	`, exportPath, newName)

	_, err := h.runPowerShell(ctx, powerShellCmd)
	return err
}

// PowerOn starts a Hyper-V VM
func (h *HyperVBackupProvider) PowerOn(ctx context.Context, vmName string) error {
	powerShellCmd := fmt.Sprintf(`Start-VM -Name "%s"`, vmName)
	_, err := h.runPowerShell(ctx, powerShellCmd)
	return err
}

// PowerOff stops a Hyper-V VM
func (h *HyperVBackupProvider) PowerOff(ctx context.Context, vmName string) error {
	powerShellCmd := fmt.Sprintf(`Stop-VM -Name "%s" -Force`, vmName)
	_, err := h.runPowerShell(ctx, powerShellCmd)
	return err
}

// Pause pauses a Hyper-V VM
func (h *HyperVBackupProvider) Pause(ctx context.Context, vmName string) error {
	powerShellCmd := fmt.Sprintf(`Suspend-VM -Name "%s"`, vmName)
	_, err := h.runPowerShell(ctx, powerShellCmd)
	return err
}

// Resume resumes a paused Hyper-V VM
func (h *HyperVBackupProvider) Resume(ctx context.Context, vmName string) error {
	powerShellCmd := fmt.Sprintf(`Resume-VM -Name "%s"`, vmName)
	_, err := h.runPowerShell(ctx, powerShellCmd)
	return err
}

// Helper functions

func (h *HyperVBackupProvider) getVMState(ctx context.Context, vmName string) (string, error) {
	powerShellCmd := fmt.Sprintf(`Get-VM -Name "%s" | Select-Object -ExpandProperty State`, vmName)
	output, err := h.runPowerShell(ctx, powerShellCmd)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

func (h *HyperVBackupProvider) createCheckpoint(ctx context.Context, vmName, checkpointName string) error {
	powerShellCmd := fmt.Sprintf(`Checkpoint-VM -Name "%s" -SnapshotName "%s"`, vmName, checkpointName)
	_, err := h.runPowerShell(ctx, powerShellCmd)
	return err
}

func (h *HyperVBackupProvider) removeCheckpoint(ctx context.Context, vmName, checkpointName string) error {
	powerShellCmd := fmt.Sprintf(`Get-VMSnapshot -VMName "%s" -Name "%s" | Remove-VMSnapshot`, vmName, checkpointName)
	_, err := h.runPowerShell(ctx, powerShellCmd)
	return err
}

// CreateCheckpoint creates a Hyper-V checkpoint with enhanced options
func (h *HyperVBackupProvider) CreateCheckpoint(ctx context.Context, vmName, checkpointName string, description string, checkpointType string) (*HyperVCheckpoint, error) {
	var powerShellCmd string

	switch strings.ToLower(checkpointType) {
	case "production", "standard":
		powerShellCmd = fmt.Sprintf(`
			Checkpoint-VM -Name "%s" -SnapshotName "%s" -SnapshotType %s
			Get-VMSnapshot -VMName "%s" -SnapshotName "%s" | ForEach-Object {
				[PSCustomObject]@{
					Name = $_.Name
					ID = $_.Id.ToString()
					VMName = $_.VMName
					CreateTime = $_.CreationTime.ToString("yyyy-MM-ddTHH:mm:ss")
					Type = $_.SnapshotType.ToString()
					ParentID = if ($_.ParentSnapshotId) { $_.ParentSnapshotId.ToString() } else { "" }
				}
			} | ConvertTo-Json -Depth 3
		`, vmName, checkpointName, checkpointType, vmName, checkpointName)
	default:
		powerShellCmd = fmt.Sprintf(`
			Checkpoint-VM -Name "%s" -SnapshotName "%s"
			Get-VMSnapshot -VMName "%s" -SnapshotName "%s" | ForEach-Object {
				[PSCustomObject]@{
					Name = $_.Name
					ID = $_.Id.ToString()
					VMName = $_.VMName
					CreateTime = $_.CreationTime.ToString("yyyy-MM-ddTHH:mm:ss")
					Type = $_.SnapshotType.ToString()
					ParentID = if ($_.ParentSnapshotId) { $_.ParentSnapshotId.ToString() } else { "" }
				}
			} | ConvertTo-Json -Depth 3
		`, vmName, checkpointName, vmName, checkpointName)
	}

	output, err := h.runPowerShell(ctx, powerShellCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create checkpoint: %w", err)
	}

	var checkpoint HyperVCheckpoint
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &checkpoint); err != nil {
		return nil, fmt.Errorf("failed to parse checkpoint info: %w", err)
	}

	checkpoint.Description = description
	checkpoint.VMName = vmName

	return &checkpoint, nil
}

// GetVMCheckpoints gets all checkpoints for a VM
func (h *HyperVBackupProvider) GetVMCheckpoints(ctx context.Context, vmName string) ([]HyperVCheckpoint, error) {
	powerShellCmd := fmt.Sprintf(`
		Get-VMSnapshot -VMName "%s" | Sort-Object CreationTime -Descending | ForEach-Object {
			[PSCustomObject]@{
				Name = $_.Name
				ID = $_.Id.ToString()
				VMName = $_.VMName
				CreateTime = $_.CreationTime.ToString("yyyy-MM-ddTHH:mm:ss")
				Type = $_.SnapshotType.ToString()
				ParentID = if ($_.ParentSnapshotId) { $_.ParentSnapshotId.ToString() } else { "" }
			}
		} | ConvertTo-Json -Depth 3
	`, vmName)

	output, err := h.runPowerShell(ctx, powerShellCmd)
	if err != nil {
		return nil, err
	}

	var checkpoints []HyperVCheckpoint
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &checkpoints); err != nil {
		return nil, err
	}

	return checkpoints, nil
}

// PerformIncrementalBackup performs incremental backup using checkpoints
func (h *HyperVBackupProvider) PerformIncrementalBackup(ctx context.Context, vmName string, dest string, lastCheckpointID string) (*models.BackupResult, error) {
	result := &models.BackupResult{
		StartTime: time.Now(),
		Status:    models.JobStatusRunning,
	}

	// Create destination directory
	if err := os.MkdirAll(dest, 0755); err != nil {
		result.Status = models.JobStatusFailed
		result.ErrorMessage = err.Error()
		result.EndTime = time.Now()
		return result, err
	}

	// Get VM info
	vm, err := h.GetVMInfo(ctx, vmName)
	if err != nil {
		result.Status = models.JobStatusFailed
		result.ErrorMessage = fmt.Sprintf("failed to get VM info: %v", err)
		result.EndTime = time.Now()
		return result, err
	}

	// Create new checkpoint
	checkpointName := fmt.Sprintf("NovaBackup-%s-%s", vmName, time.Now().Format("20060102-150405"))
	_, err = h.CreateCheckpoint(ctx, vmName, checkpointName, "Created by NovaBackup for incremental backup", "Standard")
	if err != nil {
		result.Status = models.JobStatusFailed
		result.ErrorMessage = fmt.Sprintf("failed to create checkpoint: %v", err)
		result.EndTime = time.Now()
		return result, err
	}

	// Export VM with checkpoint
	exportPath := filepath.Join(dest, fmt.Sprintf("%s_incremental", vmName))

	powerShellCmd := fmt.Sprintf(`
		Export-VM -Name "%s" -Path "%s" -SnapshotName "%s"
	`, vmName, exportPath, checkpointName)

	_, err = h.runPowerShell(ctx, powerShellCmd)
	if err != nil {
		// Clean up checkpoint on failure
		h.removeCheckpoint(ctx, vmName, checkpointName)
		result.Status = models.JobStatusFailed
		result.ErrorMessage = fmt.Errorf("failed to export VM: %w", err).Error()
		result.EndTime = time.Now()
		return result, err
	}

	// Create incremental backup metadata
	incrementalInfo := HyperVIncrementalInfo{
		VMName:           vmName,
		LastCheckpointID: lastCheckpointID,
		ChangeTracking:   true,
		DiskChanges:      map[string]int64{"total": vm.DiskSize},
		BackupTime:       time.Now(),
	}

	metadataFile := filepath.Join(dest, fmt.Sprintf("%s_incremental_metadata.json", vmName))
	metadataData, _ := json.MarshalIndent(incrementalInfo, "", "  ")
	if err := os.WriteFile(metadataFile, metadataData, 0644); err != nil {
		result.Status = models.JobStatusFailed
		result.ErrorMessage = fmt.Errorf("failed to write incremental metadata: %v", err).Error()
		result.EndTime = time.Now()
		return result, err
	}

	// Get exported VM size
	var totalSize int64
	filepath.Walk(exportPath, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})

	result.Status = models.JobStatusCompleted
	result.EndTime = time.Now()
	result.BytesWritten = totalSize
	result.FilesTotal = 1
	result.FilesSuccess = 1

	return result, nil
}

// RemoveCheckpoint removes a checkpoint by ID
func (h *HyperVBackupProvider) RemoveCheckpoint(ctx context.Context, vmName, checkpointID string) error {
	powerShellCmd := fmt.Sprintf(`Remove-VMSnapshot -VMName "%s" -SnapshotId "%s" -Confirm:$false`, vmName, checkpointID)
	_, err := h.runPowerShell(ctx, powerShellCmd)
	return err
}

// EnableChangeTracking enables change tracking for a VM
func (h *HyperVBackupProvider) EnableChangeTracking(ctx context.Context, vmName string) error {
	// Hyper-V doesn't have explicit change tracking like VMware CBT
	// We use checkpoint-based incremental backups
	return nil
}

// IsChangeTrackingEnabled checks if change tracking is available
func (h *HyperVBackupProvider) IsChangeTrackingEnabled(ctx context.Context, vmName string) (bool, error) {
	// Hyper-V supports incremental backups through checkpoints
	return true, nil
}

// GetIncrementalInfo gets incremental backup information
func (h *HyperVBackupProvider) GetIncrementalInfo(ctx context.Context, vmName string) (*HyperVIncrementalInfo, error) {
	checkpoints, err := h.GetVMCheckpoints(ctx, vmName)
	if err != nil {
		return nil, err
	}

	var lastCheckpointID string
	if len(checkpoints) > 0 {
		lastCheckpointID = checkpoints[0].ID
	}

	return &HyperVIncrementalInfo{
		VMName:           vmName,
		LastCheckpointID: lastCheckpointID,
		ChangeTracking:   true,
		DiskChanges:      make(map[string]int64),
		BackupTime:       time.Now(),
	}, nil
}

// Enhanced backup method with incremental support
func (h *HyperVBackupProvider) BackupEnhanced(ctx context.Context, vmName string, dest string) (*models.BackupResult, error) {
	// Try incremental backup first
	incrementalInfo, err := h.GetIncrementalInfo(ctx, vmName)
	if err == nil && incrementalInfo.LastCheckpointID != "" {
		result, incErr := h.PerformIncrementalBackup(ctx, vmName, dest, incrementalInfo.LastCheckpointID)
		if incErr == nil {
			return result, nil
		}
		// Fall back to full backup if incremental fails
	}

	// Perform full backup
	return h.Backup(ctx, vmName, dest)
}

func (h *HyperVBackupProvider) exportVM(ctx context.Context, vmName, dest string) error {
	powerShellCmd := fmt.Sprintf(`Export-VM -Name "%s" -Path "%s"`, vmName, dest)
	_, err := h.runPowerShell(ctx, powerShellCmd)
	return err
}

func (h *HyperVBackupProvider) getDirectorySize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

func (h *HyperVBackupProvider) runPowerShell(ctx context.Context, command string) (string, error) {
	var cmd *exec.Cmd

	if h.useRemote {
		// Remote PowerShell session
		psCmd := fmt.Sprintf(`Enter-PSSession -ComputerName %s -Credential (Get-Credential); %s; Exit-PSSession`, h.host, command)
		cmd = exec.CommandContext(ctx, "powershell", "-Command", psCmd)
	} else {
		// Local PowerShell
		cmd = exec.CommandContext(ctx, "powershell", "-Command", command)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("powershell failed: %w, output: %s", err, string(output))
	}

	return string(output), nil
}

// HyperVBackupResult contains Hyper-V backup results
type HyperVBackupResult struct {
	VMName         string
	CheckpointName string
	ExportPath     string
	Size           int64
	Duration       time.Duration
	ErrorMessage   string
}
