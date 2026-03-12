package replication

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ReplicationEngine manages VM replication operations
type ReplicationEngine interface {
	// StartReplication starts a new replication job
	StartReplication(ctx context.Context, req *ReplicationRequest) (*ReplicationJob, error)

	// StopReplication stops an active replication
	StopReplication(ctx context.Context, jobID string) error

	// GetJob returns a specific replication job
	GetJob(ctx context.Context, jobID string) (*ReplicationJob, error)

	// ListJobs returns all replication jobs
	ListJobs(ctx context.Context) ([]*ReplicationJob, error)

	// GetJobStatus returns the status of a replication job
	GetJobStatus(ctx context.Context, jobID string) (*ReplicationStatus, error)
}

// ReplicationRequest defines parameters for starting a replication job
type ReplicationRequest struct {
	SourceVM             string               `json:"source_vm"`
	SourceVC             string               `json:"source_vc"` // Source vCenter
	DestinationHost      string               `json:"destination_host"`
	DestinationDatastore string               `json:"destination_datastore"`
	DestinationVC        string               `json:"destination_vc"` // Destination vCenter
	ReplicationType      ReplicationType      `json:"replication_type"`
	Schedule             *ReplicationSchedule `json:"schedule"`
	NetworkMap           map[string]string    `json:"network_map"` // source network -> dest network
	StoragePolicy        string               `json:"storage_policy"`
	Priority             JobPriority          `json:"priority"`
	BandwidthLimit       int                  `json:"bandwidth_limit_mbps"` // 0 = unlimited
	EnableRPO            bool                 `json:"enable_rpo"`
	RPOTarget            time.Duration        `json:"rpo_target"` // e.g., 15 minutes
	RetentionPolicy      *RetentionPolicy     `json:"retention_policy"`
}

// ReplicationType defines the type of replication
type ReplicationType string

const (
	ReplicationTypeSync   ReplicationType = "sync"   // Synchronous
	ReplicationTypeAsync  ReplicationType = "async"  // Asynchronous
	ReplicationTypeBackup ReplicationType = "backup" // Backup-based replication
)

// ReplicationSchedule defines when replication should run
type ReplicationSchedule struct {
	Type     string        `json:"type"`     // "continuous", "scheduled", "manual"
	Time     string        `json:"time"`     // e.g., "22:00"
	Days     []int         `json:"days"`     // 0=Sunday, 1=Monday, etc.
	Interval time.Duration `json:"interval"` // for periodic
}

// JobPriority defines replication job priority
type JobPriority struct {
	Level   int  `json:"level"`   // 1-5, 1=highest
	Preempt bool `json:"preempt"` // can preempt other jobs
}

// RetentionPolicy defines how long to keep replicated VMs
type RetentionPolicy struct {
	MaxSnapshots  int           `json:"max_snapshots"`
	RetentionDays int           `json:"retention_days"`
	ArchiveAfter  time.Duration `json:"archive_after"`
}

// ReplicationJob represents an active or completed replication job
type ReplicationJob struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	SourceVM        string          `json:"source_vm"`
	Status          JobStatus       `json:"status"`
	Progress        int             `json:"progress"` // 0-100
	ReplicationType ReplicationType `json:"replication_type"`
	SourceSize      int64           `json:"source_size_gb"`
	TransferredSize int64           `json:"transferred_size_gb"`
	SpeedMBps       float64         `json:"speed_mbps"`
	StartTime       time.Time       `json:"start_time"`
	EndTime         *time.Time      `json:"end_time"`
	LastSyncTime    *time.Time      `json:"last_sync_time"`
	NextSyncTime    *time.Time      `json:"next_sync_time"`
	Error           string          `json:"error,omitempty"`
	TargetVM        string          `json:"target_vm"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// JobStatus represents the status of a replication job
type JobStatus string

const (
	JobStatusPending   JobStatus = "Pending"
	JobStatusRunning   JobStatus = "Running"
	JobStatusPaused    JobStatus = "Paused"
	JobStatusStopped   JobStatus = "Stopped"
	JobStatusCompleted JobStatus = "Completed"
	JobStatusFailed    JobStatus = "Failed"
	JobStatusSyncing   JobStatus = "Syncing"
)

// ReplicationStatus provides detailed status of a replication
type ReplicationStatus struct {
	JobID            string            `json:"job_id"`
	Status           JobStatus         `json:"status"`
	Progress         int               `json:"progress"`
	SourceVM         string            `json:"source_vm"`
	TargetVM         string            `json:"target_vm"`
	SourceSizeGB     int64             `json:"source_size_gb"`
	TransferredGB    int64             `json:"transferred_gb"`
	RemainingGB      int64             `json:"remaining_gb"`
	SpeedMBps        float64           `json:"speed_mbps"`
	AverageSpeedMBps float64           `json:"average_speed_mbps"`
	EtaSeconds       int               `json:"eta_seconds"`
	CurrentSnapshot  string            `json:"current_snapshot"`
	TotalSnapshots   int               `json:"total_snapshots"`
	RPOCompliance    bool              `json:"rpo_compliant"`
	LastSyncTime     time.Time         `json:"last_sync_time"`
	NextSyncTime     time.Time         `json:"next_sync_time"`
	NetworkMap       map[string]string `json:"network_map"`
	Warning          string            `json:"warning,omitempty"`
	Error            string            `json:"error,omitempty"`
}

// InMemoryReplicationEngine provides in-memory implementation
type InMemoryReplicationEngine struct {
	jobs map[string]*ReplicationJob
	mu   sync.RWMutex
}

// NewInMemoryReplicationEngine creates a new in-memory replication engine
func NewInMemoryReplicationEngine() *InMemoryReplicationEngine {
	return &InMemoryReplicationEngine{
		jobs: make(map[string]*ReplicationJob),
	}
}

// StartReplication starts a new replication job
func (e *InMemoryReplicationEngine) StartReplication(ctx context.Context, req *ReplicationRequest) (*ReplicationJob, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	job := &ReplicationJob{
		ID:              uuid.New().String(),
		Name:            fmt.Sprintf("Repl-%s-to-%s", req.SourceVM, req.DestinationHost),
		SourceVM:        req.SourceVM,
		Status:          JobStatusRunning,
		Progress:        0,
		ReplicationType: req.ReplicationType,
		SourceSize:      100, // Mock: 100GB
		TargetVM:        req.SourceVM + "-replica",
		StartTime:       time.Now(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	e.jobs[job.ID] = job

	// In real implementation, would start async worker here
	go e.simulateReplication(job.ID)

	return job, nil
}

// simulateReplication simulates replication progress (for demo)
func (e *InMemoryReplicationEngine) simulateReplication(jobID string) {
	time.Sleep(2 * time.Second)

	e.mu.Lock()
	defer e.mu.Unlock()

	if job, ok := e.jobs[jobID]; ok {
		job.Progress = 100
		job.Status = JobStatusCompleted
		job.TransferredSize = job.SourceSize
		now := time.Now()
		job.EndTime = &now
		job.LastSyncTime = &now
	}
}

// StopReplication stops an active replication job
func (e *InMemoryReplicationEngine) StopReplication(ctx context.Context, jobID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	job, ok := e.jobs[jobID]
	if !ok {
		return fmt.Errorf("job %s not found", jobID)
	}

	if job.Status != JobStatusRunning && job.Status != JobStatusSyncing {
		return fmt.Errorf("job %s is not running", jobID)
	}

	job.Status = JobStatusStopped
	now := time.Now()
	job.EndTime = &now
	job.UpdatedAt = time.Now()

	return nil
}

// GetJob returns a specific replication job
func (e *InMemoryReplicationEngine) GetJob(ctx context.Context, jobID string) (*ReplicationJob, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	job, ok := e.jobs[jobID]
	if !ok {
		return nil, fmt.Errorf("job %s not found", jobID)
	}

	return job, nil
}

// ListJobs returns all replication jobs
func (e *InMemoryReplicationEngine) ListJobs(ctx context.Context) ([]*ReplicationJob, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	jobs := make([]*ReplicationJob, 0, len(e.jobs))
	for _, job := range e.jobs {
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// GetJobStatus returns detailed status of a replication job
func (e *InMemoryReplicationEngine) GetJobStatus(ctx context.Context, jobID string) (*ReplicationStatus, error) {
	e.mu.RLock()
	job, ok := e.jobs[jobID]
	e.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("job %s not found", jobID)
	}

	remaining := job.SourceSize - job.TransferredSize
	status := &ReplicationStatus{
		JobID:            job.ID,
		Status:           job.Status,
		Progress:         job.Progress,
		SourceVM:         job.SourceVM,
		TargetVM:         job.TargetVM,
		SourceSizeGB:     job.SourceSize,
		TransferredGB:    job.TransferredSize,
		RemainingGB:      remaining,
		SpeedMBps:        job.SpeedMBps,
		AverageSpeedMBps: job.SpeedMBps,
	}

	return status, nil
}
