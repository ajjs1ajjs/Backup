package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"
	"runtime"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// ImmutableMode defines the backup immutability enforcement mechanism
type ImmutableMode string

const (
	ImmutableModeS3ObjectLock    ImmutableMode = "s3_object_lock"
	ImmutableModeLinuxChattr     ImmutableMode = "linux_chattr"
	ImmutableModeWindowsACL      ImmutableMode = "windows_acl"
	ImmutableModeS3LegalHold     ImmutableMode = "s3_legal_hold"
)

// ImmutableConfig holds configuration for an immutable repository
type ImmutableConfig struct {
	Mode           ImmutableMode
	RetentionDays  int
	Bucket         string // For S3 mode
	S3Client       *s3.Client
}

// ImmutableProvider wraps a Provider with immutability enforcement
type ImmutableProvider struct {
	inner         Provider
	config        ImmutableConfig
	lockedKeys    map[string]time.Time // key -> locked until
}

// NewImmutableProvider creates an Immutable wrapper around any storage Provider
func NewImmutableProvider(inner Provider, config ImmutableConfig) *ImmutableProvider {
	return &ImmutableProvider{
		inner:      inner,
		config:     config,
		lockedKeys: make(map[string]time.Time),
	}
}

// Store stores data and immediately applies immutability lock
func (p *ImmutableProvider) Store(ctx context.Context, key string, data io.Reader, size int64) error {
	if err := p.inner.Store(ctx, key, data, size); err != nil {
		return err
	}

	// Apply immutability after storing
	if err := p.applyLock(ctx, key); err != nil {
		log.Printf("[Immutable] WARNING: stored %s but failed to apply lock: %v", key, err)
		// Intentionally do not fail the backup — log and alert
	} else {
		p.lockedKeys[key] = time.Now().Add(time.Duration(p.config.RetentionDays) * 24 * time.Hour)
	}

	return nil
}

// Delete is blocked by immutability if the retention window has not expired
func (p *ImmutableProvider) Delete(ctx context.Context, key string) error {
	if until, ok := p.lockedKeys[key]; ok {
		if time.Now().Before(until) {
			return fmt.Errorf("[Immutable] cannot delete %s: retention period expires at %s", key, until.Format(time.RFC3339))
		}
	}
	return p.inner.Delete(ctx, key)
}

// Retrieve delegates to the inner provider
func (p *ImmutableProvider) Retrieve(ctx context.Context, key string) (io.ReadCloser, error) {
	return p.inner.Retrieve(ctx, key)
}

// Exists delegates to the inner provider
func (p *ImmutableProvider) Exists(ctx context.Context, key string) (bool, error) {
	return p.inner.Exists(ctx, key)
}

// GetStats delegates to the inner provider
func (p *ImmutableProvider) GetStats(ctx context.Context) (*Stats, error) {
	return p.inner.GetStats(ctx)
}

// Close closes the inner provider
func (p *ImmutableProvider) Close() error {
	return p.inner.Close()
}

// GetImmutableStatus returns the lock expiry for a key, or zero time if not locked
func (p *ImmutableProvider) GetImmutableStatus(key string) (locked bool, until time.Time) {
	if exp, ok := p.lockedKeys[key]; ok && time.Now().Before(exp) {
		return true, exp
	}
	return false, time.Time{}
}

// applyLock applies the configured immutability mechanism to a stored object
func (p *ImmutableProvider) applyLock(ctx context.Context, key string) error {
	switch p.config.Mode {
	case ImmutableModeS3ObjectLock:
		return p.applyS3ObjectLock(ctx, key)
	case ImmutableModeS3LegalHold:
		return p.applyS3LegalHold(ctx, key)
	case ImmutableModeLinuxChattr:
		return p.applyLinuxChattr(key)
	case ImmutableModeWindowsACL:
		return p.applyWindowsACL(key)
	default:
		return fmt.Errorf("unknown immutable mode: %s", p.config.Mode)
	}
}

// applyS3ObjectLock applies S3 Object Lock COMPLIANCE mode
func (p *ImmutableProvider) applyS3ObjectLock(ctx context.Context, key string) error {
	if p.config.S3Client == nil {
		return fmt.Errorf("S3 client not configured for ImmutableProvider")
	}

	retainUntil := time.Now().Add(time.Duration(p.config.RetentionDays) * 24 * time.Hour)

	_, err := p.config.S3Client.PutObjectRetention(ctx, &s3.PutObjectRetentionInput{
		Bucket: aws.String(p.config.Bucket),
		Key:    aws.String(key),
		Retention: &s3types.ObjectLockRetention{
			Mode:            s3types.ObjectLockRetentionModeCompliance,
			RetainUntilDate: aws.Time(retainUntil),
		},
	})
	if err != nil {
		return fmt.Errorf("S3 ObjectLock PutRetention failed: %w", err)
	}

	log.Printf("[Immutable] S3 Object Lock COMPLIANCE applied to %s (until %s)", key, retainUntil.Format(time.RFC3339))
	return nil
}

// applyS3LegalHold applies S3 Legal Hold (indefinite)
func (p *ImmutableProvider) applyS3LegalHold(ctx context.Context, key string) error {
	if p.config.S3Client == nil {
		return fmt.Errorf("S3 client not configured for ImmutableProvider")
	}

	_, err := p.config.S3Client.PutObjectLegalHold(ctx, &s3.PutObjectLegalHoldInput{
		Bucket:    aws.String(p.config.Bucket),
		Key:       aws.String(key),
		LegalHold: &s3types.ObjectLockLegalHold{Status: s3types.ObjectLockLegalHoldStatusOn},
	})
	if err != nil {
		return fmt.Errorf("S3 LegalHold failed: %w", err)
	}

	log.Printf("[Immutable] S3 Legal Hold applied to %s", key)
	return nil
}

// applyLinuxChattr makes a file immutable using chattr +i (requires root)
func (p *ImmutableProvider) applyLinuxChattr(key string) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("chattr is only available on Linux")
	}

	cmd := exec.Command("chattr", "+i", key)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("chattr +i failed for %s: %v (%s)", key, err, string(output))
	}

	log.Printf("[Immutable] chattr +i applied to %s", key)
	return nil
}

// applyWindowsACL removes write permissions via icacls (Windows fallback)
func (p *ImmutableProvider) applyWindowsACL(key string) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("icacls is only available on Windows")
	}

	cmd := exec.Command("icacls", key, "/deny", "Everyone:(W,D,DC,AD)")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("icacls write-deny failed for %s: %v (%s)", key, err, string(output))
	}

	log.Printf("[Immutable] Windows ACL write-deny applied to %s", key)
	return nil
}
