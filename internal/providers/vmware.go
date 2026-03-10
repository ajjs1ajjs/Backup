package providers

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"novabackup/pkg/models"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
)

// VMwareBackupProvider handles VMware vSphere VM backups
type VMwareBackupProvider struct {
	vcenter    string
	username   string
	password   string
	insecure   bool
	datacenter string
	client     *govmomi.Client
	finder     *find.Finder
}

// VMwareConfig contains VMware connection configuration
type VMwareConfig struct {
	VCenter    string
	Username   string
	Password   string
	Insecure   bool
	Datacenter string
}

// VMwareVMInfo contains VMware VM information
type VMwareVMInfo struct {
	Name       string
	Moid       string
	PowerState types.VirtualMachinePowerState
	GuestOS    string
	NumCPU     int32
	MemoryMB   int32
	Datastores []string
	Networks   []string
	Snapshot   *types.VirtualMachineSnapshotInfo
}

// NewVMwareBackupProvider creates a new VMware backup provider
func NewVMwareBackupProvider(cfg VMwareConfig) *VMwareBackupProvider {
	return &VMwareBackupProvider{
		vcenter:    cfg.VCenter,
		username:   cfg.Username,
		password:   cfg.Password,
		insecure:   cfg.Insecure,
		datacenter: cfg.Datacenter,
	}
}

// Connect establishes connection to vCenter
func (v *VMwareBackupProvider) Connect(ctx context.Context) error {
	if v.client != nil {
		return nil
	}

	// Build vCenter URL
	vcenterURL := fmt.Sprintf("https://%s/sdk", v.vcenter)
	u, err := soap.ParseURL(vcenterURL)
	if err != nil {
		return fmt.Errorf("failed to parse vCenter URL: %w", err)
	}

	// Create SOAP client
	soapClient := soap.NewClient(u, v.insecure)

	// Create vim25 client
	vimClient, err := vim25.NewClient(ctx, soapClient)
	if err != nil {
		return fmt.Errorf("failed to create vim25 client: %w", err)
	}
	vimClient.UserAgent = "NovaBackup-v6.0"

	// Create govmomi client
	v.client = &govmomi.Client{
		Client: vimClient,
	}

	// Login
	err = v.client.SessionManager.Login(ctx, url.UserPassword(v.username, v.password))
	if err != nil {
		return fmt.Errorf("failed to login to vCenter: %w", err)
	}

	// Create finder
	v.finder = find.NewFinder(v.client.Client, false)

	// Set datacenter if specified
	if v.datacenter != "" {
		dc, err := v.finder.Datacenter(ctx, v.datacenter)
		if err != nil {
			return fmt.Errorf("failed to find datacenter %s: %w", v.datacenter, err)
		}
		v.finder.SetDatacenter(dc)
	}

	return nil
}

// Disconnect closes vCenter connection
func (v *VMwareBackupProvider) Disconnect(ctx context.Context) error {
	if v.client != nil {
		return v.client.Logout(ctx)
	}
	return nil
}

// ListVMs lists all VMs in the datacenter
func (v *VMwareBackupProvider) ListVMs(ctx context.Context) ([]VMwareVMInfo, error) {
	if err := v.Connect(ctx); err != nil {
		return nil, err
	}

	// Find all VMs
	vms, err := v.finder.VirtualMachineList(ctx, "*")
	if err != nil {
		return nil, fmt.Errorf("failed to list VMs: %w", err)
	}

	var vmInfos []VMwareVMInfo
	for _, vm := range vms {
		var props mo.VirtualMachine
		if err := vm.Properties(ctx, vm.Reference(), nil, &props); err != nil {
			continue
		}

		vmInfo := VMwareVMInfo{
			Name:       props.Name,
			Moid:       props.Self.Value,
			PowerState: props.Runtime.PowerState,
			GuestOS:    props.Config.GuestFullName,
			NumCPU:     props.Config.Hardware.NumCPU,
			MemoryMB:   props.Config.Hardware.MemoryMB,
		}

		// Get datastores
		for _, ds := range props.Datastore {
			vmInfo.Datastores = append(vmInfo.Datastores, ds.Value)
		}

		// Get networks
		for _, net := range props.Network {
			vmInfo.Networks = append(vmInfo.Networks, net.Value)
		}

		// Get snapshot info
		vmInfo.Snapshot = props.Snapshot

		vmInfos = append(vmInfos, vmInfo)
	}

	return vmInfos, nil
}

// GetVMByName finds a VM by name
func (v *VMwareBackupProvider) GetVMByName(ctx context.Context, name string) (*object.VirtualMachine, error) {
	if err := v.Connect(ctx); err != nil {
		return nil, err
	}

	vm, err := v.finder.VirtualMachine(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to find VM %s: %w", name, err)
	}

	return vm, nil
}

// CreateSnapshot creates a VM snapshot
func (v *VMwareBackupProvider) CreateSnapshot(ctx context.Context, vmName, snapshotName string, description string, memory bool, quiesce bool) (*types.ManagedObjectReference, error) {
	vm, err := v.GetVMByName(ctx, vmName)
	if err != nil {
		return nil, err
	}

	task, err := vm.CreateSnapshot(ctx, snapshotName, description, memory, quiesce)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot: %w", err)
	}

	// Wait for task completion
	_, err = task.WaitForResultEx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("snapshot task failed: %w", err)
	}

	// Get snapshot reference
	var props mo.VirtualMachine
	if err := vm.Properties(ctx, vm.Reference(), []string{"snapshot"}, &props); err != nil {
		return nil, err
	}

	if props.Snapshot == nil || props.Snapshot.CurrentSnapshot == nil {
		return nil, fmt.Errorf("no snapshot created")
	}

	return props.Snapshot.CurrentSnapshot, nil
}

// RemoveSnapshot removes a VM snapshot
func (v *VMwareBackupProvider) RemoveSnapshot(ctx context.Context, vmName string, snapshotRef *types.ManagedObjectReference) error {
	vm, err := v.GetVMByName(ctx, vmName)
	if err != nil {
		return err
	}

	// Remove all snapshots (consolidate)
	consolidate := true
	task, err := vm.RemoveAllSnapshot(ctx, &consolidate)
	if err != nil {
		return fmt.Errorf("failed to remove snapshot: %w", err)
	}

	_, err = task.WaitForResultEx(ctx, nil)
	return err
}

// Backup performs VMware VM backup using govmomi
func (v *VMwareBackupProvider) Backup(ctx context.Context, vmName string, dest string) (*models.BackupResult, error) {
	result := &models.BackupResult{
		StartTime: time.Now(),
		Status:    models.JobStatusRunning,
	}

	// Connect to vCenter
	if err := v.Connect(ctx); err != nil {
		result.Status = models.JobStatusFailed
		result.ErrorMessage = err.Error()
		result.EndTime = time.Now()
		return result, err
	}
	defer v.Disconnect(ctx)

	// Create destination directory
	if err := os.MkdirAll(dest, 0755); err != nil {
		result.Status = models.JobStatusFailed
		result.ErrorMessage = err.Error()
		result.EndTime = time.Now()
		return result, err
	}

	// Get VM reference
	vm, err := v.GetVMByName(ctx, vmName)
	if err != nil {
		result.Status = models.JobStatusFailed
		result.ErrorMessage = err.Error()
		result.EndTime = time.Now()
		return result, err
	}

	// Get VM properties
	var props mo.VirtualMachine
	if err := vm.Properties(ctx, vm.Reference(), nil, &props); err != nil {
		result.Status = models.JobStatusFailed
		result.ErrorMessage = err.Error()
		result.EndTime = time.Now()
		return result, err
	}

	// Create snapshot for consistent backup
	snapshotName := fmt.Sprintf("NovaBackup-%s-%s", vmName, time.Now().Format("20060102-150405"))
	snapshotRef, err := v.CreateSnapshot(ctx, vmName, snapshotName, "Created by NovaBackup", false, false)
	if err != nil {
		result.Status = models.JobStatusFailed
		result.ErrorMessage = fmt.Sprintf("Failed to create snapshot: %v", err)
		result.EndTime = time.Now()
		return result, err
	}

	// Ensure snapshot is removed on exit
	defer v.RemoveSnapshot(ctx, vmName, snapshotRef)

	// Export VM to OVF/OVA
	ovaFile := filepath.Join(dest, fmt.Sprintf("%s.ova", vmName))

	// Use ovftool if available
	ovftoolPath, err := exec.LookPath("ovftool")
	if err == nil {
		// Use ovftool
		vcenterURL := fmt.Sprintf("vi://%s:%s@%s/%s/vm/%s", v.username, v.password, v.vcenter, v.datacenter, vmName)
		args := []string{vcenterURL, ovaFile, "--acceptAllEulas", "--noSSLVerify"}

		cmd := exec.CommandContext(ctx, ovftoolPath, args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			result.Status = models.JobStatusFailed
			result.ErrorMessage = fmt.Sprintf("ovftool failed: %v, output: %s", err, string(output))
			result.EndTime = time.Now()
			return result, err
		}
	} else {
		// Use govmomi export (basic implementation)
		// Note: Full export requires more complex implementation
		result.Status = models.JobStatusFailed
		result.ErrorMessage = "ovftool not found. Please install ovftool for full VM export."
		result.EndTime = time.Now()
		return result, fmt.Errorf("ovftool not found")
	}

	// Get backup file size
	fileInfo, err := os.Stat(ovaFile)
	if err != nil {
		result.Status = models.JobStatusFailed
		result.ErrorMessage = err.Error()
		result.EndTime = time.Now()
		return result, err
	}

	result.Status = models.JobStatusCompleted
	result.EndTime = time.Now()
	result.BytesWritten = fileInfo.Size()
	result.FilesTotal = 1
	result.FilesSuccess = 1

	return result, nil
}

// Restore restores a VM from backup
func (v *VMwareBackupProvider) Restore(ctx context.Context, backupPath string, vmName string) error {
	if err := v.Connect(ctx); err != nil {
		return err
	}
	defer v.Disconnect(ctx)

	// Check if ovftool is available
	ovftoolPath, err := exec.LookPath("ovftool")
	if err != nil {
		return fmt.Errorf("ovftool not found in PATH")
	}

	// Deploy OVA/OVF
	args := []string{backupPath, fmt.Sprintf("vi://%s:%s@%s/%s/host", v.username, v.password, v.vcenter, v.datacenter), "--acceptAllEulas", "--noSSLVerify", "--name=" + vmName}

	cmd := exec.CommandContext(ctx, ovftoolPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ovftool failed: %v, output: %s", err, string(output))
	}

	return nil
}

// PowerOn powers on a VM
func (v *VMwareBackupProvider) PowerOn(ctx context.Context, vmName string) error {
	vm, err := v.GetVMByName(ctx, vmName)
	if err != nil {
		return err
	}

	task, err := vm.PowerOn(ctx)
	if err != nil {
		return err
	}

	_, err = task.WaitForResultEx(ctx, nil)
	return err
}

// PowerOff powers off a VM
func (v *VMwareBackupProvider) PowerOff(ctx context.Context, vmName string) error {
	vm, err := v.GetVMByName(ctx, vmName)
	if err != nil {
		return err
	}

	task, err := vm.PowerOff(ctx)
	if err != nil {
		return err
	}

	_, err = task.WaitForResultEx(ctx, nil)
	return err
}

// GetCBTChangedBlocks gets changed blocks since last snapshot (for incremental backup)
func (v *VMwareBackupProvider) GetCBTChangedBlocks(ctx context.Context, vmName string, lastSnapshotRef *types.ManagedObjectReference) (*types.DiskChangeInfo, error) {
	// Simple implementation - return nil for now
	// Full CBT implementation requires proper snapshot chain tracking
	return nil, fmt.Errorf("CBT requires full snapshot chain implementation")
}

// KVM Backup Provider

// KVMBackupProvider handles KVM/QEMU VM backups
type KVMBackupProvider struct {
	uri string
}

// KVMConfig contains KVM connection configuration
type KVMConfig struct {
	URI string
}

// KVMVMInfo contains KVM VM information
type KVMVMInfo struct {
	Name      string
	UUID      string
	State     string
	CPU       int
	Memory    int64
	DiskPaths []string
}

// NewKVMBackupProvider creates a new KVM backup provider
func NewKVMBackupProvider(cfg KVMConfig) *KVMBackupProvider {
	if cfg.URI == "" {
		cfg.URI = "qemu:///system"
	}
	return &KVMBackupProvider{uri: cfg.URI}
}

// ListVMs lists all VMs
func (k *KVMBackupProvider) ListVMs(ctx context.Context) ([]KVMVMInfo, error) {
	output, err := k.runVirsh(ctx, "list", "--all", "--name")
	if err != nil {
		return nil, err
	}
	var vms []KVMVMInfo
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			if info, err := k.GetVMInfo(ctx, line); err == nil {
				vms = append(vms, *info)
			}
		}
	}
	return vms, nil
}

// GetVMInfo gets VM info
func (k *KVMBackupProvider) GetVMInfo(ctx context.Context, vmName string) (*KVMVMInfo, error) {
	state, _ := k.runVirsh(ctx, "domstate", vmName)
	return &KVMVMInfo{Name: vmName, State: strings.TrimSpace(state)}, nil
}

// Backup performs VM backup
func (k *KVMBackupProvider) Backup(ctx context.Context, vmName, dest string) (*models.BackupResult, error) {
	result := &models.BackupResult{StartTime: time.Now(), Status: models.JobStatusRunning}
	os.MkdirAll(dest, 0755)

	// Export VM config
	configPath := filepath.Join(dest, vmName+".xml")
	if _, err := k.runVirsh(ctx, "dumpxml", vmName, ">", configPath); err != nil {
		result.Status = models.JobStatusFailed
		result.ErrorMessage = err.Error()
		result.EndTime = time.Now()
		return result, err
	}

	result.Status = models.JobStatusCompleted
	result.EndTime = time.Now()
	return result, nil
}

// PowerOn starts VM
func (k *KVMBackupProvider) PowerOn(ctx context.Context, vmName string) error {
	_, err := k.runVirsh(ctx, "start", vmName)
	return err
}

// PowerOff stops VM
func (k *KVMBackupProvider) PowerOff(ctx context.Context, vmName string) error {
	_, err := k.runVirsh(ctx, "shutdown", vmName)
	return err
}

// runVirsh runs virsh command
func (k *KVMBackupProvider) runVirsh(ctx context.Context, args ...string) (string, error) {
	allArgs := append([]string{"-c", k.uri}, args...)
	cmd := exec.CommandContext(ctx, "virsh", allArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("virsh: %w, output: %s", err, output)
	}
	return string(output), nil
}
