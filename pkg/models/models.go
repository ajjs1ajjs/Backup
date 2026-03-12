package models

import (
	"time"

	"github.com/google/uuid"
)

// TimeNow returns the current time (helper for testing)
func TimeNow() time.Time {
	return time.Now()
}

// StorageInfo contains storage backend information
type StorageInfo struct {
	Type        string    `json:"type"`
	TotalSize   int64     `json:"total_size"`
	UsedSize    int64     `json:"used_size"`
	FreeSize    int64     `json:"free_size"`
	ObjectCount int64     `json:"object_count"`
	Endpoint    string    `json:"endpoint"`
	Bucket      string    `json:"bucket"`
	Region      string    `json:"region"`
	LastUpdated time.Time `json:"last_updated"`
}

type JobType string

const (
	JobTypeFile     JobType = "file"
	JobTypeDatabase JobType = "database"
	JobTypeVM       JobType = "vm"
)

type RepositoryType string

const (
	RepositoryTypeLocal  RepositoryType = "local"
	RepositoryTypeS3     RepositoryType = "s3"
	RepositoryTypeSOBR   RepositoryType = "sobr"
)

type Repository struct {
	ID          uuid.UUID      `json:"id"`
	Name        string         `json:"name"`
	Type        RepositoryType `json:"type"`
	Path        string         `json:"path"`      // For local
	Endpoint    string         `json:"endpoint"`  // For S3
	Bucket      string         `json:"bucket"`    // For S3
	Region      string         `json:"region"`    // For S3
	AccessKey   string         `json:"access_key"`
	SecretKey   string         `json:"secret_key"`
	IsSOBR      bool           `json:"is_sobr"`
	ParentSOBR  uuid.UUID      `json:"parent_sobr"`
	Tier        string         `json:"tier"`      // Performance, Capacity
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
)

type Job struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	JobType     JobType   `json:"job_type"`
	Source      string    `json:"source"`
	Destination string    `json:"destination"`
	Schedule    string    `json:"schedule"`
	Enabled       bool      `json:"enabled"`
	RetentionDays int       `json:"retention_days"`
	
	// Guest Processing
	EnableGuestProcessing bool   `json:"enable_guest_processing"`
	GuestCredentialsID    string `json:"guest_credentials_id"`
	
	// GFS Retention Policy
	GFSEnabled    bool `json:"gfs_enabled"`
	GFSDaily      int  `json:"gfs_daily"`    // Daily backups to keep
	GFSWeekly     int  `json:"gfs_weekly"`   // Weekly backups to keep
	GFSMonthly    int  `json:"gfs_monthly"`  // Monthly backups to keep
	GFSYearly     int  `json:"gfs_yearly"`   // Yearly backups to keep
	GFSWeeklyDay  int  `json:"gfs_weekly_day"`  // 0=Sunday
	GFSMonthlyDay int  `json:"gfs_monthly_day"` // 1-31
	
	// Backup Windows
	BackupWindowID string `json:"backup_window_id"`
	
	// Synthetic Full
	SyntheticFullEnabled bool   `json:"synthetic_full_enabled"`
	SyntheticFullDay     int    `json:"synthetic_full_day"` // 0=Sunday
	
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type BackupResult struct {
	ID           uuid.UUID `json:"id"`
	JobID        uuid.UUID `json:"job_id"`
	Status       JobStatus `json:"status"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	BytesRead    int64     `json:"bytes_read"`
	BytesWritten int64     `json:"bytes_written"`
	FilesTotal   int       `json:"files_total"`
	FilesSuccess int       `json:"files_success"`
	FilesFailed  int       `json:"files_failed"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

type BackupConfig struct {
	Source        string
	Destination   string
	Compress      bool
	Encrypt       bool
	EncryptionKey []byte
	ChunkSize     int64
	BufferSize    int
	ParallelJobs  int `json:"parallel_jobs"`
	RetentionDays int `json:"retention_days"`
	
	// Guest Processing
	EnableGuestProcessing bool              `json:"enable_guest_processing"`
	GuestCredentialsID    string            `json:"guest_credentials_id"`
	GuestApplications     []string          `json:"guest_applications"` // "SQL", "Exchange", "AD", "SharePoint"
	EnableQuiesce         bool              `json:"enable_quiesce"`
	TruncateLogs          bool              `json:"truncate_logs"`
	PreFreezeScript       string            `json:"pre_freeze_script"`
	PostThawScript        string            `json:"post_thaw_script"`
}

type RestoreConfig struct {
	Source      string
	Destination string
	Overwrite   bool
}

type RestorePoint struct {
	ID              uuid.UUID `json:"id"`
	JobID           uuid.UUID `json:"job_id"`
	PointTime       time.Time `json:"point_time"`
	Status          JobStatus `json:"status"`
	TotalBytes      int64     `json:"total_bytes"`
	ProcessedBytes  int64     `json:"processed_bytes"`
	DurationSeconds int       `json:"duration_seconds"`
	Metadata        string    `json:"metadata,omitempty"`
}

type FileInfo struct {
	Path     string
	Name     string
	Size     int64
	ModTime  time.Time
	IsDir    bool
	Checksum string
}

// ChunkInfo represents a chunk with metadata
type ChunkInfo struct {
	Hash           string
	SizeBytes      int64
	CompressedSize int64
	StoragePath    string
}

type ChunkMapping struct {
	ChunkHash    string
	Sequence     int
	OriginalPath string
}

// ============================================================================
// Guest Processing Models
// ============================================================================

// GuestApplicationType represents supported application types for guest processing
type GuestApplicationType string

const (
	GuestAppSQLServer  GuestApplicationType = "sql_server"
	GuestAppExchange   GuestApplicationType = "exchange"
	GuestAppActiveDir  GuestApplicationType = "active_directory"
	GuestAppSharePoint GuestApplicationType = "sharepoint"
)

// GuestProcessingSettings contains guest processing configuration for a job
type GuestProcessingSettings struct {
	Enabled             bool                   `json:"enabled"`
	CredentialsID       string                 `json:"credentials_id"`
	Applications        []GuestApplicationType `json:"applications"`
	EnableQuiesce       bool                   `json:"enable_quiesce"`
	TruncateLogs        bool                   `json:"truncate_logs"`
	SkipCrypto          bool                   `json:"skip_crypto"`
	PreFreezeScriptPath string                 `json:"pre_freeze_script_path"`
	PostThawScriptPath  string                 `json:"post_thaw_script_path"`
	Timeout             int                    `json:"timeout_seconds"`
}

// VSSWriterInfo represents VSS writer information
type VSSWriterInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	State    string `json:"state"` // "Stable", "Failed", "Waiting"
	Type     string `json:"type"`
	Volume   string `json:"volume"`
}

// GuestProcessingResult contains the result of guest processing
type GuestProcessingResult struct {
	Success           bool              `json:"success"`
	SnapshotID        string            `json:"snapshot_id,omitempty"`
	Applications      []string          `json:"applications"`
	ApplicationStatus map[string]string `json:"application_status"` // app -> status
	ErrorMessage      string            `json:"error_message,omitempty"`
	Duration          int64             `json:"duration_ms"`
	LogsTruncated     bool              `json:"logs_truncated"`
}

// ApplicationInfo represents discovered application on a guest
type ApplicationInfo struct {
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Version     string            `json:"version"`
	Status      string            `json:"status"`
	Databases   []string          `json:"databases,omitempty"`
	Mailboxes   []string          `json:"mailboxes,omitempty"`
	Properties  map[string]string `json:"properties,omitempty"`
}

// ============================================================================
// RBAC Models
// ============================================================================

// RBACRole represents a role with permissions
type RBACRole struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"` // "jobs:read", "backups:create", etc.
	Builtin     bool     `json:"builtin"`     // true for system roles
	CreatedAt   time.Time `json:"created_at"`
}

// RBACUser represents a user with role assignments
type RBACUser struct {
	ID        string   `json:"id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	RoleIDs   []string `json:"role_ids"`
	Active    bool     `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	LastLogin time.Time `json:"last_login,omitempty"`
}

// RBACPermission represents a single permission
type RBACPermission struct {
	Resource string `json:"resource"` // "jobs", "backups", "storage"
	Action   string `json:"action"`   // "create", "read", "update", "delete"
	Scope    string `json:"scope"`    // "global", "own", "team"
}

// ============================================================================
// Audit Models
// ============================================================================

// AuditEvent represents an audit log entry
type AuditEvent struct {
	ID         string                 `json:"id"`
	Timestamp  time.Time              `json:"timestamp"`
	Level      string                 `json:"level"` // "INFO", "WARNING", "ERROR", "CRITICAL", "AUDIT"
	Type       string                 `json:"type"`  // "auth.login", "backup.job_create", etc.
	Category   string                 `json:"category"`
	UserID     string                 `json:"user_id,omitempty"`
	Username   string                 `json:"username,omitempty"`
	SourceIP   string                 `json:"source_ip,omitempty"`
	Resource   string                 `json:"resource,omitempty"`
	ResourceID string                 `json:"resource_id,omitempty"`
	Action     string                 `json:"action,omitempty"`
	Status     string                 `json:"status,omitempty"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Duration   int64                  `json:"duration_ms,omitempty"`
}

// AuditStats represents audit statistics
type AuditStats struct {
	TotalEvents int            `json:"total_events"`
	ByLevel     map[string]int `json:"by_level"`
	ByCategory  map[string]int `json:"by_category"`
	ByType      map[string]int `json:"by_type"`
}

// BackupWindow represents a backup time window
type BackupWindow struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	DayOfWeek int    `json:"day_of_week"` // 0=Sunday, 6=Saturday, -1=everyday
	StartTime string `json:"start_time"`  // "22:00"
	EndTime   string `json:"end_time"`    // "06:00"
	Enabled   bool   `json:"enabled"`
}
