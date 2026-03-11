package network

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"
)

// mountNFS mounts an NFS share
func (e *NetworkStorageEngine) mountNFS(ctx context.Context) error {
	if e.mounted {
		return nil
	}

	// Create mount point if not exists
	if err := os.MkdirAll(e.config.MountPoint, 0755); err != nil {
		return fmt.Errorf("failed to create mount point: %w", err)
	}

	// Build mount command based on OS
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		// Linux: mount -t nfs host:share/path mountpoint
		remotePath := fmt.Sprintf("%s:%s%s", e.config.Host, e.config.Share, e.config.Path)
		cmd = exec.CommandContext(ctx, "mount", "-t", "nfs", remotePath, e.config.MountPoint)

	case "darwin":
		// macOS: mount -t nfs host:share/path mountpoint
		remotePath := fmt.Sprintf("%s:%s%s", e.config.Host, e.config.Share, e.config.Path)
		cmd = exec.CommandContext(ctx, "mount", "-t", "nfs", remotePath, e.config.MountPoint)

	case "windows":
		// Windows: Use NFS client or map network drive
		return fmt.Errorf("NFS mounting not supported on Windows, use SMB instead")

	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	// Execute mount command with retries
	var lastErr error
	for i := 0; i < e.config.RetryCount; i++ {
		output, err := cmd.CombinedOutput()
		if err == nil {
			e.mounted = true
			return nil
		}
		lastErr = fmt.Errorf("mount attempt %d failed: %w, output: %s", i+1, err, string(output))

		// Wait before retry
		if i < e.config.RetryCount-1 {
			time.Sleep(time.Second * time.Duration(i+1))
		}
	}

	return lastErr
}

// unmountNFS unmounts an NFS share
func (e *NetworkStorageEngine) unmountNFS(ctx context.Context) error {
	if !e.mounted {
		return nil
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.CommandContext(ctx, "umount", e.config.MountPoint)

	case "darwin":
		cmd = exec.CommandContext(ctx, "umount", e.config.MountPoint)

	case "windows":
		// Windows doesn't typically use NFS
		e.mounted = false
		return nil

	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to unmount NFS share: %w", err)
	}

	e.mounted = false
	return nil
}

// mountSMB mounts an SMB/CIFS share
func (e *NetworkStorageEngine) mountSMB(ctx context.Context) error {
	if e.mounted {
		return nil
	}

	// Create mount point if not exists
	if err := os.MkdirAll(e.config.MountPoint, 0755); err != nil {
		return fmt.Errorf("failed to create mount point: %w", err)
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		// Linux: mount -t cifs //host/share mountpoint -o username=user,password=pass
		remotePath := fmt.Sprintf("//%s/%s", e.config.Host, e.config.Share)
		options := fmt.Sprintf("username=%s,password=%s", e.config.Username, e.config.Password)
		if e.config.Domain != "" {
			options += fmt.Sprintf(",domain=%s", e.config.Domain)
		}
		options += ",iocharset=utf8,vers=3.0"

		cmd = exec.CommandContext(ctx, "mount", "-t", "cifs", remotePath, e.config.MountPoint, "-o", options)

	case "darwin":
		// macOS: mount_smbfs //user@host/share mountpoint
		userAuth := e.config.Username
		if e.config.Password != "" {
			userAuth = fmt.Sprintf("%s:%s", e.config.Username, e.config.Password)
		}
		remotePath := fmt.Sprintf("//%s@%s/%s", userAuth, e.config.Host, e.config.Share)

		cmd = exec.CommandContext(ctx, "mount_smbfs", remotePath, e.config.MountPoint)

	case "windows":
		// Windows: net use Z: \\host\share /user:user password
		driveLetter := "Z:" // Could be configurable

		// First disconnect if already connected
		exec.CommandContext(ctx, "net", "use", driveLetter, "/delete").Run()

		userArg := ""
		if e.config.Username != "" {
			if e.config.Domain != "" {
				userArg = fmt.Sprintf("/user:%s\\%s", e.config.Domain, e.config.Username)
			} else {
				userArg = fmt.Sprintf("/user:%s", e.config.Username)
			}
		}

		args := []string{"use", driveLetter, fmt.Sprintf("\\\\%s\\%s", e.config.Host, e.config.Share)}
		if userArg != "" {
			args = append(args, userArg)
		}
		if e.config.Password != "" {
			args = append(args, e.config.Password)
		}

		cmd = exec.CommandContext(ctx, "net", args...)

	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	// Execute mount command with retries
	var lastErr error
	for i := 0; i < e.config.RetryCount; i++ {
		output, err := cmd.CombinedOutput()
		if err == nil {
			e.mounted = true
			return nil
		}
		lastErr = fmt.Errorf("mount attempt %d failed: %w, output: %s", i+1, err, string(output))

		// Wait before retry
		if i < e.config.RetryCount-1 {
			time.Sleep(time.Second * time.Duration(i+1))
		}
	}

	return lastErr
}

// unmountSMB unmounts an SMB/CIFS share
func (e *NetworkStorageEngine) unmountSMB(ctx context.Context) error {
	if !e.mounted {
		return nil
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.CommandContext(ctx, "umount", e.config.MountPoint)

	case "darwin":
		cmd = exec.CommandContext(ctx, "umount", e.config.MountPoint)

	case "windows":
		// Windows: net use Z: /delete
		cmd = exec.CommandContext(ctx, "net", "use", "Z:", "/delete")

	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to unmount SMB share: %w", err)
	}

	e.mounted = false
	return nil
}

// CheckMount checks if a share is already mounted
func (e *NetworkStorageEngine) CheckMount() bool {
	// Simple check by trying to access mount point
	if _, err := os.Stat(e.config.MountPoint); err != nil {
		return false
	}

	// Try to read directory
	file, err := os.Open(e.config.MountPoint)
	if err != nil {
		return false
	}
	defer file.Close()

	_, err = file.Readdirnames(1)
	return err == nil
}

// GetMountCommand returns the mount command for debugging
func (e *NetworkStorageEngine) GetMountCommand() string {
	switch e.engineType {
	case StorageNFS:
		remotePath := fmt.Sprintf("%s:%s%s", e.config.Host, e.config.Share, e.config.Path)
		return fmt.Sprintf("mount -t nfs %s %s", remotePath, e.config.MountPoint)

	case StorageSMB:
		switch runtime.GOOS {
		case "linux":
			remotePath := fmt.Sprintf("//%s/%s", e.config.Host, e.config.Share)
			options := fmt.Sprintf("username=%s,password=***", e.config.Username)
			return fmt.Sprintf("mount -t cifs %s %s -o %s", remotePath, e.config.MountPoint, options)

		case "darwin":
			userAuth := e.config.Username
			remotePath := fmt.Sprintf("//%s@%s/%s", userAuth, e.config.Host, e.config.Share)
			return fmt.Sprintf("mount_smbfs %s %s", remotePath, e.config.MountPoint)

		case "windows":
			return fmt.Sprintf("net use Z: \\\\%s\\%s /user:%s ***", e.config.Host, e.config.Share, e.config.Username)
		}
	}

	return "unknown mount command"
}
