# NovaBackup v6.0 - API Documentation

## Table of Contents

1. [VMware Provider API](#vmware-provider-api)
2. [Hyper-V Provider API](#hyper-v-provider-api)
3. [CBT (Changed Block Tracking) API](#cbt-api)
4. [Instant Recovery API](#instant-recovery-api)
5. [Cloud Storage API](#cloud-storage-api)
6. [CLI Commands](#cli-commands)

---

## VMware Provider API

### Connection

```go
// Create VMware client
config := &vmware.ConnectionConfig{
    Host:     "vcenter.local",
    Username: "admin",
    Password: "password",
    Port:     443,
    Insecure: true,
}

client, err := vmware.NewClient(logger, config)
if err != nil {
    log.Fatal(err)
}
defer client.Close()
```

### VM Inventory

```go
inventory := vmware.NewInventory(client)

// List all VMs
vms, err := inventory.ListVirtualMachines(ctx)

// Get specific VM
vm, err := inventory.GetVirtualMachine(ctx, "VM-Name")

// Get VM info
info, err := vm.GetInfo()
```

### Backup Operations

```go
// Full backup
engine := vmware.NewBackupEngine(client)
result, err := engine.FullBackup(ctx, vm, &vmware.BackupConfig{
    Name:        "backup-job",
    Destination: "/backups",
    Compression: true,
    Encryption:  false,
}, nil)

// Incremental backup
incEngine := vmware.NewIncrementalBackupEngine(client, "/state")
result, err := incEngine.PerformIncrementalBackup(ctx, vm, config, callback)
```

### Snapshot Management

```go
// Create snapshot
task, err := vm.CreateSnapshot("snapshot-name", "description", true, true)
err = task.Wait(ctx)

// Remove snapshot
task, err = vm.RemoveSnapshot("snapshot-name", true)

// Revert to snapshot
task, err = vm.RevertToSnapshot("snapshot-name", false)
```

### CBT Operations

```go
cbtMgr := vmware.NewCBTManager(client)

// Enable CBT
err := cbtMgr.EnableCBTForVM(ctx, vm)

// Get CBT status
status, err := cbtMgr.GetCBTStatus(ctx, vm)

// Query changes
changes, err := cbtMgr.QueryAllDiskChanges(ctx, vm, previousState.DiskChangeIDs)
```

---

## Hyper-V Provider API

### Connection

```go
config := &hyperv.ConnectionConfig{
    Server: "hyperv-server", // empty for local
}

client, err := hyperv.NewClient(logger, config)
```

### VM Operations

```go
// List VMs
vms, err := client.ListVMs(ctx)

// Get VM info
info, err := client.GetVMInfo(ctx, "VM-Name")

// Start/Stop VM
err = client.StartVM(ctx, "VM-Name")
err = client.StopVM(ctx, "VM-Name", false) // true for force
```

### Checkpoint (Snapshot) Management

```go
// Create checkpoint
checkpoint, err := client.CreateCheckpoint(ctx, "VM-Name", "checkpoint-name")

// Get checkpoints
checkpoints, err := client.GetCheckpoints(ctx, "VM-Name")

// Remove checkpoint
err = client.RemoveCheckpoint(ctx, "VM-Name", "checkpoint-name")
```

### Backup/Restore

```go
// Export VM
err = client.ExportVM(ctx, "VM-Name", "/export/path")

// Import VM
vm, err := client.ImportVM(ctx, "/import/path")

// Backup VM
err = client.BackupVM(ctx, "VM-Name", "/backup/path", false)
```

---

## CBT API

### Enable CBT

```go
cbtMgr := vmware.NewCBTManager(client)

// Enable on VM
err := cbtMgr.EnableCBTForVM(ctx, vm)

// Check status
status, err := cbtMgr.GetCBTStatus(ctx, vm)
```

### Query Changes

```go
// Query all disks
changes, err := cbtMgr.QueryAllDiskChanges(ctx, vm, map[int32]string{
    2000: "*", // Query all changed blocks
})

// Query specific disk
changeInfo, err := cbtMgr.QueryDiskChanges(ctx, vm, diskKey, changeID)
```

### Reset CBT

```go
err := cbtMgr.ResetCBT(ctx, vm)
```

---

## Instant Recovery API

### NFS Server

```go
// Create NFS server
nfsServer, err := instantrecovery.NewNFSServer(logger, &instantrecovery.NFSConfig{
    RootPath: "/exports",
    Port:     2049,
})

// Start server
err = nfsServer.Start(ctx)

// Publish backup
exportURL, err := nfsServer.PublishBackup("/backup/path", "VM-Name")
```

### Instant Recovery Manager

```go
manager := instantrecovery.NewInstantRecoveryManager(logger)

// Initialize NFS
err := manager.InitializeNFS(&instantrecovery.NFSConfig{
    RootPath: "/exports",
})

// Start instant recovery
session, err := manager.StartInstantRecovery(ctx, "VM-Name", "/backup/path")

// Stop instant recovery
err = manager.StopInstantRecovery(session.SessionID)

// List sessions
sessions := manager.ListSessions()
```

---

## Cloud Storage API

### S3 Provider

```go
// Create S3 provider
provider, err := cloud.NewS3Provider(logger, &cloud.S3Config{
    Endpoint:        "s3.amazonaws.com",
    Region:          "us-east-1",
    Bucket:          "novabackup",
    AccessKeyID:     "AKIA...",
    SecretAccessKey: "secret...",
})

// Upload
err = provider.Upload(ctx, "backup/key", reader, size)

// Download
err = provider.Download(ctx, "backup/key", writer)

// List objects
objects, err := provider.List(ctx, "prefix/")

// Delete
err = provider.Delete(ctx, "backup/key")

// Archive to Glacier
err = provider.ArchiveTier(ctx, "backup/key")

// Restore from Glacier
err = provider.RestoreFromArchive(ctx, "backup/key", 7)
```

---

## CLI Commands

### VMware Commands

```bash
# Connect to vCenter
nova vmware connect vcenter.local --username admin --password pass

# List VMs
nova vmware list

# Get VM info
nova vmware info VM-Name
nova vmware info VM-Name --json

# Create snapshot
nova vmware snapshot create VM-Name snapshot-name --quiesce --memory

# Backup VM
nova vmware backup VM-Name --destination /backups --compression
nova vmware backup VM-Name --incremental

# CBT operations
nova vmware cbt enable VM-Name
nova vmware cbt status VM-Name
```

### Hyper-V Commands

```bash
# List VMs
nova hyperv list

# Backup VM
nova hyperv backup VM-Name --destination /backups
```

---

## Configuration

### Environment Variables

```bash
export NOVA_VCENTER_HOST="vcenter.local"
export NOVA_VCENTER_USER="admin"
export NOVA_VCENTER_PASS="password"
export NOVA_BACKUP_PATH="/backups"
export NOVA_LOG_LEVEL="info"
```

### Configuration File

```yaml
# config.yaml
vmware:
  host: vcenter.local
  username: admin
  password: password
  insecure: true

backup:
  default_path: /backups
  compression: true
  encryption: false

logging:
  level: info
  format: json
```

---

## Error Handling

All APIs return errors that can be checked:

```go
result, err := client.SomeOperation(ctx)
if err != nil {
    // Handle error
    log.Printf("Operation failed: %v", err)
    return
}
```

Common error types:
- `ConnectionError` - Failed to connect to hypervisor
- `NotFoundError` - VM or resource not found
- `PermissionError` - Insufficient permissions
- `TimeoutError` - Operation timed out

---

## Best Practices

1. **Always close connections**: Use `defer client.Close()`
2. **Use contexts**: Pass contexts with timeouts for operations
3. **Handle errors**: Check all errors appropriately
4. **Enable CBT**: For efficient incremental backups
5. **Use snapshots**: For consistent backups of running VMs
6. **Monitor progress**: Use callbacks for long operations

---

## Examples

See `examples/` directory for complete working examples:
- `vmware_backup.go` - Full VM backup example
- `incremental_backup.go` - Incremental backup with CBT
- `instant_recovery.go` - Instant VM recovery
- `sql_backup.go` - SQL Server backup

---

## Support

For support and questions:
- Documentation: https://docs.novabackup.io
- Issues: https://github.com/novabackup/nova/issues
- Email: support@novabackup.io
