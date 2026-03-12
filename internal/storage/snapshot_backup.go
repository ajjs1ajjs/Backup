package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"novabackup/pkg/providers/san"
)

// SnapshotSource identifies a SAN volume to back up via snapshot
type SnapshotSource struct {
	ProviderName string // e.g. "netapp", "dell"
	VolumeID     string
	MountPath    string // Where the snapshot will be mounted for reading
}

// SnapshotBackupJob performs a backup by reading data from a transient SAN snapshot.
// This offloads I/O from the hypervisor entirely — data is read directly from the storage array.
type SnapshotBackupJob struct {
	sanManager *san.SANManager
	target     Provider // Where to write the backup data
}

// NewSnapshotBackupJob creates a new SAN-offloaded backup job
func NewSnapshotBackupJob(mgr *san.SANManager, target Provider) *SnapshotBackupJob {
	return &SnapshotBackupJob{
		sanManager: mgr,
		target:     target,
	}
}

// Run executes the snapshot-based backup lifecycle:
// CreateSnapshot → MountSnapshot → ReadData → UnmountSnapshot → DeleteSnapshot
func (j *SnapshotBackupJob) Run(ctx context.Context, src SnapshotSource, reader func(mountPath string) (io.Reader, int64, error)) error {
	provider, err := j.sanManager.GetProvider(src.ProviderName)
	if err != nil {
		return fmt.Errorf("SAN provider %q not found: %w", src.ProviderName, err)
	}

	// 1. Create SAN snapshot
	log.Printf("[SnapshotBackup] Creating snapshot for volume %s on %s", src.VolumeID, src.ProviderName)
	snapshot, err := provider.CreateSnapshot(ctx, src.VolumeID)
	if err != nil {
		return fmt.Errorf("failed to create SAN snapshot: %w", err)
	}

	defer func() {
		// 5. Cleanup: always delete snapshot when done
		log.Printf("[SnapshotBackup] Deleting snapshot %s", snapshot.ID)
		if delErr := provider.DeleteSnapshot(ctx, snapshot.ID); delErr != nil {
			log.Printf("[SnapshotBackup] WARNING: failed to delete snapshot %s: %v", snapshot.ID, delErr)
		}
	}()

	// 2. Mount snapshot (if provider supports it)
	mountedPath := src.MountPath
	if mounter, ok := provider.(san.SnapshotMounter); ok {
		log.Printf("[SnapshotBackup] Mounting snapshot %s at %s", snapshot.ID, src.MountPath)
		mountedPath, err = mounter.MountSnapshot(ctx, snapshot.ID, src.MountPath)
		if err != nil {
			return fmt.Errorf("failed to mount snapshot: %w", err)
		}
		defer func() {
			// 4. Unmount after reading
			log.Printf("[SnapshotBackup] Unmounting snapshot %s", snapshot.ID)
			if unErr := mounter.UnmountSnapshot(ctx, snapshot.ID); unErr != nil {
				log.Printf("[SnapshotBackup] WARNING: failed to unmount snapshot %s: %v", snapshot.ID, unErr)
			}
		}()
	}

	// 3. Read data and store in target
	log.Printf("[SnapshotBackup] Reading data from snapshot mount %s", mountedPath)
	data, size, err := reader(mountedPath)
	if err != nil {
		return fmt.Errorf("failed to read snapshot data: %w", err)
	}

	key := fmt.Sprintf("snap/%s/%s_%d", src.ProviderName, src.VolumeID, time.Now().Unix())
	if storeErr := j.target.Store(ctx, key, data, size); storeErr != nil {
		return fmt.Errorf("failed to store snapshot backup: %w", storeErr)
	}

	log.Printf("[SnapshotBackup] Snapshot backup complete: %s", key)
	return nil
}
