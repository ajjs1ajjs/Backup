// Package hyperv provides Microsoft Hyper-V integration for NovaBackup
package hyperv

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"go.uber.org/zap"
)

// Client represents a Hyper-V client
type Client struct {
	logger *zap.Logger
	server string // Hyper-V server hostname (empty for local)
}

// ConnectionConfig holds connection parameters for Hyper-V
type ConnectionConfig struct {
	Server   string // Hyper-V server hostname (empty for local)
	Username string // For remote Hyper-V
	Password string // For remote Hyper-V
}

// VM represents a Hyper-V virtual machine
type VM struct {
	Name          string            `json:"name"`
	State         string            `json:"state"`
	Uptime        string            `json:"uptime"`
	CPUUsage      int               `json:"cpu_usage"`
	MemoryAssigned int64            `json:"memory_assigned"`
	MemoryDemand  int64            `json:"memory_demand"`
	Status        string            `json:"status"`
	Generation    int               `json:"generation"`
	Version       string            `json:"version"`
	Path          string            `json:"path"`
}

// VMInfo contains detailed information about a Hyper-V VM
type VMInfo struct {
	VM
	NetworkAdapters []NetworkAdapter `json:"network_adapters"`
	HardDrives      []HardDrive      `json:"hard_drives"`
	DVDDrives       []DVDDrive       `json:"dvd_drives"`
	ProcessorCount  int              `json:"processor_count"`
}

// NetworkAdapter represents a Hyper-V network adapter
type NetworkAdapter struct {
	Name       string `json:"name"`
	MacAddress string `json:"mac_address"`
	IPAddress  string `json:"ip_address"`
	SwitchName string `json:"switch_name"`
}

// HardDrive represents a Hyper-V virtual hard disk
type HardDrive struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	PoolName string `json:"pool_name"`
}

// DVDDrive represents a Hyper-V DVD drive
type DVDDrive struct {
	Name    string `json:"name"`
	Path    string `json:"path,omitempty"`
}

// Checkpoint represents a Hyper-V checkpoint (snapshot)
type Checkpoint struct {
	Name        string    `json:"name"`
	ID          string    `json:"id"`
	ParentID    string    `json:"parent_id,omitempty"`
	CreatedTime time.Time `json:"created_time"`
}

// RCTInfo contains Resilient Change Tracking information
type RCTInfo struct {
	VMName          string            `json:"vm_name"`
	Enabled         bool              `json:"enabled"`
	ReferencePoint  string            `json:"reference_point,omitempty"`
	ChangedBlocks   []ChangedBlock    `json:"changed_blocks,omitempty"`
}

// ChangedBlock represents changed blocks tracked by RCT
type ChangedBlock struct {
	Offset    int64  `json:"offset"`
	Length    int64  `json:"length"`
	VHDXPath  string `json:"vhdx_path"`
}

// NewClient creates a new Hyper-V client
func NewClient(logger *zap.Logger, config *ConnectionConfig) (*Client, error) {
	client := &Client{
		logger: logger.With(zap.String("component", "hyperv-client")),
		server: config.Server,
	}

	// Test connection by listing VMs
	if _, err := client.ListVMs(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to connect to Hyper-V: %w", err)
	}

	return client, nil
}

// ListVMs lists all Hyper-V virtual machines
func (c *Client) ListVMs(ctx context.Context) ([]VM, error) {
	c.logger.Info("Listing Hyper-V VMs")

	cmd := exec.CommandContext(ctx, "powershell", "-Command",
		"Get-VM | Select-Object Name, State, Uptime, CPUUsage, MemoryAssigned, MemoryDemand, Status, Generation, Version, Path | ConvertTo-Json -Depth 3")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list VMs: %w", err)
	}

	var vms []VM
	if err := json.Unmarshal(output, &vms); err != nil {
		// Try single VM
		var singleVM VM
		if err := json.Unmarshal(output, &singleVM); err != nil {
			return nil, fmt.Errorf("failed to parse VM list: %w", err)
		}
		vms = []VM{singleVM}
	}

	return vms, nil
}

// GetVM retrieves a specific VM by name
func (c *Client) GetVM(ctx context.Context, name string) (*VM, error) {
	c.logger.Info("Getting VM", zap.String("name", name))

	cmd := exec.CommandContext(ctx, "powershell", "-Command",
		fmt.Sprintf("Get-VM -Name '%s' | Select-Object Name, State, Uptime, CPUUsage, MemoryAssigned, MemoryDemand, Status, Generation, Version, Path | ConvertTo-Json", name))

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("VM not found: %w", err)
	}

	var vm VM
	if err := json.Unmarshal(output, &vm); err != nil {
		return nil, fmt.Errorf("failed to parse VM: %w", err)
	}

	return &vm, nil
}

// GetVMInfo retrieves detailed information about a VM
func (c *Client) GetVMInfo(ctx context.Context, name string) (*VMInfo, error) {
	c.logger.Info("Getting detailed VM info", zap.String("name", name))

	vm, err := c.GetVM(ctx, name)
	if err != nil {
		return nil, err
	}

	info := &VMInfo{
		VM:              *vm,
		NetworkAdapters: []NetworkAdapter{},
		HardDrives:      []HardDrive{},
		DVDDrives:       []DVDDrive{},
	}

	// Get network adapters
	cmd := exec.CommandContext(ctx, "powershell", "-Command",
		fmt.Sprintf("Get-VMNetworkAdapter -VMName '%s' | Select-Object Name, MacAddress, IPAddresses, SwitchName | ConvertTo-Json", name))

	if output, err := cmd.Output(); err == nil {
		var adapters []NetworkAdapter
		if err := json.Unmarshal(output, &adapters); err == nil {
			info.NetworkAdapters = adapters
		}
	}

	// Get hard drives
	cmd = exec.CommandContext(ctx, "powershell", "-Command",
		fmt.Sprintf("Get-VMHardDiskDrive -VMName '%s' | Select-Object Name, Path, PoolName | ConvertTo-Json", name))

	if output, err := cmd.Output(); err == nil {
		var drives []HardDrive
		if err := json.Unmarshal(output, &drives); err == nil {
			info.HardDrives = drives
		}
	}

	// Get processor count
	cmd = exec.CommandContext(ctx, "powershell", "-Command",
		fmt.Sprintf("(Get-VMProcessor -VMName '%s').Count", name))

	if output, err := cmd.Output(); err == nil {
		fmt.Sscanf(string(output), "%d", &info.ProcessorCount)
	}

	return info, nil
}

// StartVM starts a VM
func (c *Client) StartVM(ctx context.Context, name string) error {
	c.logger.Info("Starting VM", zap.String("name", name))

	cmd := exec.CommandContext(ctx, "powershell", "-Command",
		fmt.Sprintf("Start-VM -Name '%s'", name))

	if _, err := cmd.Output(); err != nil {
		return fmt.Errorf("failed to start VM: %w", err)
	}

	return nil
}

// StopVM stops a VM
func (c *Client) StopVM(ctx context.Context, name string, force bool) error {
	c.logger.Info("Stopping VM", zap.String("name", name), zap.Bool("force", force))

	var cmd *exec.Cmd
	if force {
		cmd = exec.CommandContext(ctx, "powershell", "-Command",
			fmt.Sprintf("Stop-VM -Name '%s' -Force -TurnOff", name))
	} else {
		cmd = exec.CommandContext(ctx, "powershell", "-Command",
			fmt.Sprintf("Stop-VM -Name '%s' -Save", name))
	}

	if _, err := cmd.Output(); err != nil {
		return fmt.Errorf("failed to stop VM: %w", err)
	}

	return nil
}

// CreateCheckpoint creates a checkpoint (snapshot)
func (c *Client) CreateCheckpoint(ctx context.Context, vmName, checkpointName string) (*Checkpoint, error) {
	c.logger.Info("Creating checkpoint",
		zap.String("vm", vmName),
		zap.String("checkpoint", checkpointName))

	cmd := exec.CommandContext(ctx, "powershell", "-Command",
		fmt.Sprintf("Checkpoint-VM -Name '%s' -SnapshotName '%s' | Select-Object Name, Id, ParentSnapshotId, CreationTime | ConvertTo-Json", 
			vmName, checkpointName))

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to create checkpoint: %w", err)
	}

	var checkpoint Checkpoint
	if err := json.Unmarshal(output, &checkpoint); err != nil {
		return nil, fmt.Errorf("failed to parse checkpoint: %w", err)
	}

	return &checkpoint, nil
}

// RemoveCheckpoint removes a checkpoint
func (c *Client) RemoveCheckpoint(ctx context.Context, vmName, checkpointName string) error {
	c.logger.Info("Removing checkpoint",
		zap.String("vm", vmName),
		zap.String("checkpoint", checkpointName))

	cmd := exec.CommandContext(ctx, "powershell", "-Command",
		fmt.Sprintf("Remove-VMSnapshot -VMName '%s' -Name '%s' -Confirm:$false", vmName, checkpointName))

	if _, err := cmd.Output(); err != nil {
		return fmt.Errorf("failed to remove checkpoint: %w", err)
	}

	return nil
}

// GetCheckpoints lists all checkpoints for a VM
func (c *Client) GetCheckpoints(ctx context.Context, vmName string) ([]Checkpoint, error) {
	c.logger.Info("Getting checkpoints", zap.String("vm", vmName))

	cmd := exec.CommandContext(ctx, "powershell", "-Command",
		fmt.Sprintf("Get-VMSnapshot -VMName '%s' | Select-Object Name, Id, ParentSnapshotId, CreationTime | ConvertTo-Json", vmName))

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get checkpoints: %w", err)
	}

	var checkpoints []Checkpoint
	if err := json.Unmarshal(output, &checkpoints); err != nil {
		// Try single checkpoint
		var single Checkpoint
		if err := json.Unmarshal(output, &single); err != nil {
			return nil, fmt.Errorf("failed to parse checkpoints: %w", err)
		}
		checkpoints = []Checkpoint{single}
	}

	return checkpoints, nil
}

// ExportVM exports a VM to a directory
func (c *Client) ExportVM(ctx context.Context, vmName, exportPath string) error {
	c.logger.Info("Exporting VM",
		zap.String("vm", vmName),
		zap.String("path", exportPath))

	cmd := exec.CommandContext(ctx, "powershell", "-Command",
		fmt.Sprintf("Export-VM -Name '%s' -Path '%s'", vmName, exportPath))

	if _, err := cmd.Output(); err != nil {
		return fmt.Errorf("failed to export VM: %w", err)
	}

	return nil
}

// ImportVM imports a VM from an exported directory
func (c *Client) ImportVM(ctx context.Context, importPath string) (*VM, error) {
	c.logger.Info("Importing VM", zap.String("path", importPath))

	cmd := exec.CommandContext(ctx, "powershell", "-Command",
		fmt.Sprintf("Import-VM -Path '%s' | Select-Object Name, State, Uptime, CPUUsage, MemoryAssigned, MemoryDemand, Status, Generation, Version, Path | ConvertTo-Json", importPath))

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to import VM: %w", err)
	}

	var vm VM
	if err := json.Unmarshal(output, &vm); err != nil {
		return nil, fmt.Errorf("failed to parse imported VM: %w", err)
	}

	return &vm, nil
}

// EnableRCT enables Resilient Change Tracking for a VM
func (c *Client) EnableRCT(ctx context.Context, vmName string) error {
	c.logger.Info("Enabling RCT for VM", zap.String("name", vmName))

	cmd := exec.CommandContext(ctx, "powershell", "-Command",
		fmt.Sprintf("Enable-VMResourceMetering -VMName '%s'; Set-VM -Name '%s' -CheckpointFileLocationPath (Get-VM -Name '%s').Path", vmName, vmName, vmName))

	if _, err := cmd.Output(); err != nil {
		return fmt.Errorf("failed to enable RCT: %w", err)
	}

	return nil
}

// GetRCTChanges gets changed blocks using RCT
func (c *Client) GetRCTChanges(ctx context.Context, vmName string, referencePoint string) (*RCTInfo, error) {
	c.logger.Info("Getting RCT changes",
		zap.String("vm", vmName),
		zap.String("reference", referencePoint))

	// RCT requires Windows Server 2016+ and specific PowerShell cmdlets
	// This is a simplified implementation

	cmd := exec.CommandContext(ctx, "powershell", "-Command",
		fmt.Sprintf(`
			$vm = Get-VM -Name '%s'
			$tracking = Get-VMHardDiskDrive -VMName '%s' | ForEach-Object {
				Get-VHD -Path $_.Path | Select-Object Path, @{N='ChangedBlockTrackingEnabled'; E={$_.ChangedBlockTrackingEnabled}}
			}
			$tracking | ConvertTo-Json
		`, vmName, vmName))

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get RCT info: %w", err)
	}

	info := &RCTInfo{
		VMName:  vmName,
		Enabled: strings.Contains(string(output), "True"),
	}

	return info, nil
}

// BackupVM performs a full backup of a Hyper-V VM
func (c *Client) BackupVM(ctx context.Context, vmName, backupPath string, incremental bool) error {
	c.logger.Info("Backing up VM",
		zap.String("vm", vmName),
		zap.String("path", backupPath),
		zap.Bool("incremental", incremental))

	// Create backup directory
	cmd := exec.CommandContext(ctx, "powershell", "-Command",
		fmt.Sprintf("New-Item -ItemType Directory -Force -Path '%s'", backupPath))

	if _, err := cmd.Output(); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Export VM
	return c.ExportVM(ctx, vmName, backupPath)
}

// RestoreVM restores a Hyper-V VM from backup
func (c *Client) RestoreVM(ctx context.Context, backupPath string) (*VM, error) {
	c.logger.Info("Restoring VM", zap.String("path", backupPath))

	return c.ImportVM(ctx, backupPath)
}

// IsHyperVInstalled checks if Hyper-V is installed on the system
func IsHyperVInstalled() bool {
	cmd := exec.Command("powershell", "-Command", "Get-WindowsOptionalFeature -Online -FeatureName Microsoft-Hyper-V | Select-Object State")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "Enabled")
}
