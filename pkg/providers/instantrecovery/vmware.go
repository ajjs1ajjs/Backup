// Package instantrecovery provides instant VM recovery functionality with VMware integration
package instantrecovery

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
	"go.uber.org/zap"
)

// VMwareInstantRecovery provides VMware-specific instant recovery
type VMwareInstantRecovery struct {
	logger       *zap.Logger
	nfsManager   *ProviderInstantRecoveryManager
	vmwareClient VMwareClientInterface
}

// VMwareClientInterface defines the interface needed from VMware client
type VMwareClientInterface interface {
	GetDatacenter(ctx context.Context, name string) (*object.Datacenter, error)
	GetHost(ctx context.Context, path string) (*object.HostSystem, error)
	GetDatastore(ctx context.Context, name string) (*object.Datastore, error)
	GetFinder() VMwareFinder
}

// VMwareFinder interface for finding objects
type VMwareFinder interface {
	Datacenter(ctx context.Context, path string) (*object.Datacenter, error)
	Datastore(ctx context.Context, path string) (*object.Datastore, error)
	HostSystem(ctx context.Context, path string) (*object.HostSystem, error)
	VirtualMachine(ctx context.Context, path string) (*object.VirtualMachine, error)
}

// NewVMwareInstantRecovery creates a new VMware instant recovery manager
func NewVMwareInstantRecovery(logger *zap.Logger, vmwareClient VMwareClientInterface, nfsConfig *NFSConfig) (*VMwareInstantRecovery, error) {
	// Initialize NFS manager
	nfsManager := NewProviderInstantRecoveryManager(logger)
	if err := nfsManager.InitializeNFS(nfsConfig); err != nil {
		return nil, fmt.Errorf("failed to initialize NFS: %w", err)
	}

	return &VMwareInstantRecovery{
		logger:       logger.With(zap.String("component", "vmware-instant-recovery")),
		nfsManager:   nfsManager,
		vmwareClient: vmwareClient,
	}, nil
}

// InstantRecoveryVMConfig contains configuration for instant recovery VM
type InstantRecoveryVMConfig struct {
	VMName          string            `json:"vm_name"`
	BackupPath      string            `json:"backup_path"`
	TargetHost      string            `json:"target_host"`
	TargetDatastore string            `json:"target_datastore"`
	TargetFolder    string            `json:"target_folder"`
	ResourcePool    string            `json:"resource_pool"`
	PowerOn         bool              `json:"power_on"`
	NetworkMapping  map[string]string `json:"network_mapping"` // Old network -> New network
}

// InstantRecoveryResult contains the result of instant recovery
type InstantRecoveryResult struct {
	SessionID        string `json:"session_id"`
	VMName           string `json:"vm_name"`
	NFSDatastoreName string `json:"nfs_datastore_name"`
	NFSExportPath    string `json:"nfs_export_path"`
	TargetHost       string `json:"target_host"`
	RegisteredVM     string `json:"registered_vm"`
	PowerOnSuccess   bool   `json:"power_on_success"`
	Status           string `json:"status"`
	Error            string `json:"error,omitempty"`
}

// StartInstantRecovery starts instant recovery of a VM
func (v *VMwareInstantRecovery) StartInstantRecovery(ctx context.Context, config InstantRecoveryVMConfig) (*InstantRecoveryResult, error) {
	v.logger.Info("Starting VMware instant recovery",
		zap.String("vm", config.VMName),
		zap.String("host", config.TargetHost))

	result := &InstantRecoveryResult{
		VMName: config.VMName,
		Status: "starting",
	}

	// Step 1: Start NFS session
	session, err := v.nfsManager.StartInstantRecovery(ctx, config.VMName, config.BackupPath)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("NFS start failed: %v", err)
		return result, fmt.Errorf("failed to start NFS: %w", err)
	}

	result.SessionID = session.SessionID
	result.NFSExportPath = session.NFSExport

	// Step 2: Mount NFS as datastore to vSphere
	datastoreName := fmt.Sprintf("NovaBackup_IR_%s", session.SessionID[:8])
	result.NFSDatastoreName = datastoreName

	err = v.mountNFSDatastore(ctx, config.TargetHost, session.NFSExport, datastoreName)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("Datastore mount failed: %v", err)
		v.nfsManager.StopInstantRecovery(session.SessionID)
		return result, fmt.Errorf("failed to mount datastore: %w", err)
	}

	// Step 3: Register VM from NFS datastore
	vmRef, err := v.registerVMFromDatastore(ctx, config, datastoreName)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("VM registration failed: %v", err)
		// Cleanup: Unmount datastore
		v.unmountNFSDatastore(ctx, config.TargetHost, datastoreName)
		v.nfsManager.StopInstantRecovery(session.SessionID)
		return result, fmt.Errorf("failed to register VM: %w", err)
	}

	result.RegisteredVM = vmRef
	result.Status = "registered"

	// Step 4: Power on VM if requested
	if config.PowerOn {
		err = v.powerOnVM(ctx, vmRef)
		if err != nil {
			v.logger.Warn("Failed to power on VM", zap.Error(err))
			result.PowerOnSuccess = false
			result.Error = fmt.Sprintf("Power on failed: %v", err)
		} else {
			result.PowerOnSuccess = true
			result.Status = "running"
		}
	}

	result.TargetHost = config.TargetHost

	v.logger.Info("Instant recovery completed",
		zap.String("session", result.SessionID),
		zap.String("status", result.Status))

	return result, nil
}

// mountNFSDatastore mounts NFS export as vSphere datastore
func (v *VMwareInstantRecovery) mountNFSDatastore(ctx context.Context, hostName, nfsPath, datastoreName string) error {
	v.logger.Info("Mounting NFS datastore",
		zap.String("host", hostName),
		zap.String("datastore", datastoreName),
		zap.String("nfs_path", nfsPath))

	host, err := v.vmwareClient.GetHost(ctx, hostName)
	if err != nil {
		return fmt.Errorf("host not found: %w", err)
	}

	datastoreSystem, err := host.ConfigManager().DatastoreSystem(ctx)
	if err != nil {
		return fmt.Errorf("failed to get datastore system: %w", err)
	}

	// For NFS, the remote host is typically the machine running NovaBackup
	// Use helper to detect reachable IP
	localIP, err := v.getReachableLocalIP(hostName)
	if err != nil {
		v.logger.Warn("Failed to detect reachable local IP, falling back to 127.0.0.1 (likely to fail)", zap.Error(err))
		localIP = "127.0.0.1"
	}

	spec := types.HostNasVolumeSpec{
		RemoteHost: localIP,
		RemotePath: nfsPath,
		LocalPath:  datastoreName,
		AccessMode: string(types.HostMountModeReadOnly),
		Type:       "NFS",
	}

	_, err = datastoreSystem.CreateNasDatastore(ctx, spec)
	if err != nil {
		return fmt.Errorf("failed to create NAS datastore: %w", err)
	}

	return nil
}

// getReachableLocalIP attempts to find a local IP that can reach the target host
func (v *VMwareInstantRecovery) getReachableLocalIP(targetHost string) (string, error) {
	// Try to resolve targetHost to IP if it's a hostname
	addrs, err := net.LookupIP(targetHost)
	if err != nil || len(addrs) == 0 {
		return "", fmt.Errorf("could not resolve host %s: %v", targetHost, err)
	}

	// Use the first resolved IP to establish a "dummy" connection to detect local address
	conn, err := net.DialTimeout("udp", net.JoinHostPort(addrs[0].String(), "1"), 2*time.Second)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

// unmountNFSDatastore removes NFS datastore from host
func (v *VMwareInstantRecovery) unmountNFSDatastore(ctx context.Context, hostName, datastoreName string) error {
	v.logger.Info("Unmounting NFS datastore",
		zap.String("host", hostName),
		zap.String("datastore", datastoreName))

	// Find the datastore to unmount
	ds, err := v.vmwareClient.GetDatastore(ctx, datastoreName)
	if err != nil {
		return err
	}

	// Unmount / Destroy the datastore
	task, err := ds.Destroy(ctx)
	if err != nil {
		return err
	}

	return task.Wait(ctx)
}

// registerVMFromDatastore registers VM from NFS datastore
func (v *VMwareInstantRecovery) registerVMFromDatastore(ctx context.Context, config InstantRecoveryVMConfig, datastoreName string) (string, error) {
	v.logger.Info("Registering VM from datastore",
		zap.String("vm", config.VMName),
		zap.String("datastore", datastoreName))

	vmxPath := fmt.Sprintf("[%s] %s/%s.vmx", datastoreName, config.VMName, config.VMName)

	host, err := v.vmwareClient.GetHost(ctx, config.TargetHost)
	if err != nil {
		return "", err
	}

	// Use the host's datacenter and its VM folder
	dc, err := v.vmwareClient.GetDatacenter(ctx, "") // Get default datacenter
	if err != nil {
		return "", err
	}

	folders, err := dc.Folders(ctx)
	if err != nil {
		return "", err
	}

	// Register the VM
	task, err := folders.VmFolder.RegisterVM(ctx, vmxPath, config.VMName, false, nil, host)
	if err != nil {
		return "", fmt.Errorf("failed to register VM: %w", err)
	}

	err = task.Wait(ctx)
	if err != nil {
		return "", fmt.Errorf("wait for VM registration failed: %w", err)
	}

	return vmxPath, nil
}

// powerOnVM powers on the recovered VM
func (v *VMwareInstantRecovery) powerOnVM(ctx context.Context, vmRef string) error {
	v.logger.Info("Powering on VM", zap.String("vm", vmRef))

	// Find VM by path (vmRef is currently vmxPath from registration)
	finder := v.vmwareClient.GetFinder()
	vm, err := finder.VirtualMachine(ctx, vmRef)
	if err != nil {
		return fmt.Errorf("failed to find VM %s: %w", vmRef, err)
	}

	task, err := vm.PowerOn(ctx)
	if err != nil {
		return fmt.Errorf("failed to trigger power on: %w", err)
	}

	return task.Wait(ctx)
}

// StopInstantRecovery stops instant recovery and cleans up
func (v *VMwareInstantRecovery) StopInstantRecovery(ctx context.Context, sessionID string) error {
	v.logger.Info("Stopping instant recovery", zap.String("session", sessionID))

	// Get session info
	session, err := v.nfsManager.GetSession(sessionID)
	if err != nil {
		return err
	}

	// 1. Power off VM if running
	// 2. Unregister VM
	// 3. Unmount datastore
	// 4. Stop NFS session

	// Cleanup NFS
	if err := v.nfsManager.StopInstantRecovery(sessionID); err != nil {
		v.logger.Warn("Failed to stop NFS session", zap.Error(err))
	}

	v.logger.Info("Instant recovery stopped",
		zap.String("session", sessionID),
		zap.String("vm", session.VMName))

	return nil
}

// GetSessionStatus returns current status of instant recovery session
func (v *VMwareInstantRecovery) GetSessionStatus(sessionID string) (*InstantRecoverySession, error) {
	return v.nfsManager.GetSession(sessionID)
}

// ListActiveSessions returns all active instant recovery sessions
func (v *VMwareInstantRecovery) ListActiveSessions() []*InstantRecoverySession {
	return v.nfsManager.ListSessions()
}

// MigrateToProduction migrates instantly recovered VM to production storage using Storage vMotion
func (v *VMwareInstantRecovery) MigrateToProduction(ctx context.Context, sessionID string, targetDatastore string) error {
	v.logger.Info("Migrating VM to production (Storage vMotion)",
		zap.String("session", sessionID),
		zap.String("target_datastore", targetDatastore))

	// 1. Get session info
	session, err := v.nfsManager.GetSession(sessionID)
	if err != nil {
		return err
	}

	// 2. Find VM
	finder := v.vmwareClient.GetFinder()
	// session.VMName is used as the key in registerVMFromDatastore
	// Assuming VM is reachable by just name if uniquely registered
	vm, err := finder.VirtualMachine(ctx, session.VMName)
	if err != nil {
		// Fallback to searching by path if stored in session metadata (to be added)
		return fmt.Errorf("failed to find VM for migration: %w", err)
	}

	// 3. Find target datastore
	ds, err := finder.Datastore(ctx, targetDatastore)
	if err != nil {
		return fmt.Errorf("target datastore %s not found: %w", targetDatastore, err)
	}

	// 4. Create relocation spec for Storage vMotion
	spec := types.VirtualMachineRelocateSpec{
		Datastore: types.NewReference(ds.Reference()),
	}

	// 5. Initiate RelocateVM_Task
	v.logger.Info("Initiating Storage vMotion task", zap.String("vm", vm.Name()), zap.String("ds", targetDatastore))
	task, err := vm.Relocate(ctx, spec, types.VirtualMachineMovePriorityDefaultPriority)
	if err != nil {
		return fmt.Errorf("failed to initiate Storage vMotion: %w", err)
	}

	err = task.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Storage vMotion task failed: %w", err)
	}

	v.logger.Info("Storage vMotion completed successfully", zap.String("vm", session.VMName))

	// 6. Cleanup NFS (optional: should we do it automatically? Veeam typically asks)
	// For now, satisfy the "Advanced Restore" goal by completing the move.
	
	return nil
}
