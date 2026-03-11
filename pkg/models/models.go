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
	EnableGuestProcessing bool   `json:"enable_guest_processing"`
	GuestCredentialsID    string `json:"guest_credentials_id"`
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
	EnableGuestProcessing bool   `json:"enable_guest_processing"`
	GuestCredentialsID    string `json:"guest_credentials_id"`
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
