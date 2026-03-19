// Backup Immutability - Protection against ransomware and deletion
package backup

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"
)

// ImmutabilityType represents the type of immutability
type ImmutabilityType string

const (
	ImmutabilityLinux   ImmutabilityType = "linux_chattr"
	ImmutabilityWindows ImmutabilityType = "windows_acl"
	ImmutabilityS3      ImmutabilityType = "s3_object_lock"
)

// ImmutabilityConfig represents immutability configuration
type ImmutabilityConfig struct {
	Enabled       bool             `json:"enabled"`
	Type          ImmutabilityType `json:"type"`
	RetentionDays int              `json:"retention_days"`
	S3Bucket      string           `json:"s3_bucket,omitempty"`
	S3ObjectLock  bool             `json:"s3_object_lock"`
}

// SetImmutability sets immutable flag on backup
func SetImmutability(backupPath string, config *ImmutabilityConfig) error {
	if !config.Enabled {
		return nil
	}

	switch config.Type {
	case ImmutabilityLinux:
		return setLinuxImmutability(backupPath, config.RetentionDays)
	case ImmutabilityWindows:
		return setWindowsImmutability(backupPath)
	case ImmutabilityS3:
		return setS3Immutability(backupPath, config.S3Bucket, config.RetentionDays)
	default:
		return fmt.Errorf("unsupported immutability type: %s", config.Type)
	}
}

// setLinuxImmutability uses chattr +i to make files immutable
func setLinuxImmutability(path string, days int) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("linux immutability only supported on Linux")
	}

	// Use chattr +i to make immutable
	cmd := exec.Command("chattr", "+i", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("chattr failed: %v - %s", err, string(output))
	}

	// Set retention timestamp
	retentionFile := path + ".retention"
	retention := time.Now().AddDate(0, 0, days)
	os.WriteFile(retentionFile, []byte(retention.Format(time.RFC3339)), 0644)

	return nil
}

// setWindowsImmutability uses ACL to prevent deletion
func setWindowsImmutability(path string) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("windows immutability only supported on Windows")
	}

	// Use icacls to deny delete permissions
	// This prevents deletion even by admins until ACL is removed
	cmd := exec.Command("icacls", path, "/grant", "Administrators:(OI)(CI)RX", "/deny", "Everyone:(OI)(CI)D")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("icacls failed: %v - %s", err, string(output))
	}

	return nil
}

// setS3Immutability uses S3 Object Lock (WORM)
func setS3Immutability(objectKey, bucket string, days int) error {
	// S3 Object Lock requires bucket to have Object Lock enabled
	// Use AWS SDK to set retention
	// Example:
	// s3Client.PutObjectRetention(&s3.PutObjectRetentionInput{
	//     Bucket: aws.String(bucket),
	//     Key: aws.String(objectKey),
	//     Retention: &types.ObjectLockRetention{
	//         Mode: types.ObjectLockRetentionModeCompliance,
	//         RetainUntilDate: aws.Time(time.Now().AddDate(0, 0, days)),
	//     },
	// })

	return fmt.Errorf("S3 Object Lock not yet implemented - use AWS SDK")
}

// RemoveImmutability removes immutable flag (only if retention period expired)
func RemoveImmutability(path string, config *ImmutabilityConfig) error {
	if !config.Enabled {
		return nil
	}

	// Check if retention period has expired
	if config.RetentionDays > 0 {
		retentionFile := path + ".retention"
		data, err := os.ReadFile(retentionFile)
		if err == nil {
			retention, err := time.Parse(time.RFC3339, string(data))
			if err == nil && time.Now().Before(retention) {
				return fmt.Errorf("retention period has not expired yet (until %s)", retention.Format("2006-01-02"))
			}
		}
	}

	switch config.Type {
	case ImmutabilityLinux:
		return removeLinuxImmutability(path)
	case ImmutabilityWindows:
		return removeWindowsImmutability(path)
	default:
		return nil
	}
}

// removeLinuxImmutability removes immutable flag
func removeLinuxImmutability(path string) error {
	cmd := exec.Command("chattr", "-i", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("chattr -i failed: %v - %s", err, string(output))
	}

	// Remove retention file
	retentionFile := path + ".retention"
	os.Remove(retentionFile)

	return nil
}

// removeWindowsImmutability removes ACL restrictions
func removeWindowsImmutability(path string) error {
	cmd := exec.Command("icacls", path, "/remove", "Everyone:D")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("icacls /remove failed: %v - %s", err, string(output))
	}
	return nil
}

// IsImmutable checks if backup is immutable
func IsImmutable(path string, config *ImmutabilityConfig) (bool, error) {
	if !config.Enabled {
		return false, nil
	}

	// Check retention
	if config.RetentionDays > 0 {
		retentionFile := path + ".retention"
		if _, err := os.Stat(retentionFile); err == nil {
			data, _ := os.ReadFile(retentionFile)
			retention, err := time.Parse(time.RFC3339, string(data))
			if err == nil && time.Now().Before(retention) {
				return true, nil
			}
		}
	}

	// For Linux, check chattr flags
	if runtime.GOOS == "linux" && config.Type == ImmutabilityLinux {
		cmd := exec.Command("lsattr", path)
		output, err := cmd.CombinedOutput()
		if err == nil && len(output) > 0 {
			// Check for 'i' flag
			return containsFlag(string(output), 'i'), nil
		}
	}

	return false, nil
}

func containsFlag(s string, flag rune) bool {
	for _, c := range s {
		if c == flag {
			return true
		}
	}
	return false
}

// GetImmutabilityStatus returns immutability status for all backups
func GetImmutabilityStatus(backupPath string, config *ImmutabilityConfig) map[string]interface{} {
	status := make(map[string]interface{})

	immutable, _ := IsImmutable(backupPath, config)
	status["immutable"] = immutable

	if immutable {
		retentionFile := backupPath + ".retention"
		if data, err := os.ReadFile(retentionFile); err == nil {
			if retention, err := time.Parse(time.RFC3339, string(data)); err == nil {
				status["retention_until"] = retention.Format("2006-01-02")
				status["days_remaining"] = int(time.Until(retention).Hours() / 24)
			}
		}
	}

	return status
}
