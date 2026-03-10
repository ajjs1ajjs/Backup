package models

import (
	"time"

	"github.com/google/uuid"
)

type JobType string

const (
	JobTypeFile     JobType = "file"
	JobTypeDatabase JobType = "database"
	JobTypeVM       JobType = "vm"
)

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
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
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
}

type RestoreConfig struct {
	Source      string
	Destination string
	Overwrite   bool
}

type FileInfo struct {
	Path     string
	Name     string
	Size     int64
	ModTime  time.Time
	IsDir    bool
	Checksum string
}
