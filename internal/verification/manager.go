package verification

import (
	"context"
	"fmt"
	"sync"
	"time"

	"novabackup/internal/multitenancy"
	"novabackup/internal/storage"
	"novabackup/internal/surebackup"

	"github.com/google/uuid"
)

// Re-export types from surebackup package for convenience
type (
	VerificationConfig  = surebackup.VerificationConfig
	VerificationResults = surebackup.VerificationResults
	NotificationConfig  = surebackup.NotificationConfig
)

// AutoVerificationManager manages automated backup verification
type AutoVerificationManager interface {
	// Schedule operations
	CreateSchedule(ctx context.Context, request *ScheduleRequest) (*Schedule, error)
	GetSchedule(ctx context.Context, scheduleID string) (*Schedule, error)
	ListSchedules(ctx context.Context, filter *ScheduleFilter) ([]*Schedule, error)
	UpdateSchedule(ctx context.Context, scheduleID string, request *UpdateScheduleRequest) (*Schedule, error)
	DeleteSchedule(ctx context.Context, scheduleID string) error
	EnableSchedule(ctx context.Context, scheduleID string) error
	DisableSchedule(ctx context.Context, scheduleID string) error

	// Verification job operations
	GetVerificationJob(ctx context.Context, jobID string) (*VerificationJob, error)
	ListVerificationJobs(ctx context.Context, filter *JobFilter) ([]*VerificationJob, error)
	CancelVerificationJob(ctx context.Context, jobID string) error
	RerunVerificationJob(ctx context.Context, jobID string) error

	// Statistics and monitoring
	GetVerificationStats(ctx context.Context, tenantID string, timeRange TimeRange) (*VerificationStats, error)
	GetScheduleStats(ctx context.Context, scheduleID string, timeRange TimeRange) (*ScheduleStats, error)
	GetGlobalStats(ctx context.Context, timeRange TimeRange) (*GlobalVerificationStats, error)

	// Health and status
	GetSystemHealth(ctx context.Context) (*SystemHealth, error)
	GetPendingJobs(ctx context.Context) ([]*VerificationJob, error)
}

// Schedule represents an automated verification schedule
type Schedule struct {
	ID            string              `json:"id"`
	Name          string              `json:"name"`
	TenantID      string              `json:"tenant_id"`
	Description   string              `json:"description"`
	Enabled       bool                `json:"enabled"`
	Type          ScheduleType        `json:"type"`
	Trigger       ScheduleTrigger     `json:"trigger"`
	BackupFilter  BackupFilter        `json:"backup_filter"`
	Verification  VerificationConfig  `json:"verification"`
	Notifications NotificationConfig  `json:"notifications"`
	Retention     RetentionPolicy     `json:"retention"`
	Concurrency   ConcurrencySettings `json:"concurrency"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
	LastRunAt     *time.Time          `json:"last_run_at,omitempty"`
	NextRunAt     *time.Time          `json:"next_run_at,omitempty"`
	Metadata      map[string]string   `json:"metadata"`
}

// ScheduleRequest contains parameters for creating a schedule
type ScheduleRequest struct {
	Name          string              `json:"name"`
	TenantID      string              `json:"tenant_id"`
	Description   string              `json:"description"`
	Type          ScheduleType        `json:"type"`
	Trigger       ScheduleTrigger     `json:"trigger"`
	BackupFilter  BackupFilter        `json:"backup_filter"`
	Verification  VerificationConfig  `json:"verification"`
	Notifications NotificationConfig  `json:"notifications"`
	Retention     RetentionPolicy     `json:"retention"`
	Concurrency   ConcurrencySettings `json:"concurrency"`
	Enabled       bool                `json:"enabled"`
	Metadata      map[string]string   `json:"metadata"`
}

// UpdateScheduleRequest contains parameters for updating a schedule
type UpdateScheduleRequest struct {
	Name          *string              `json:"name,omitempty"`
	Description   *string              `json:"description,omitempty"`
	Enabled       *bool                `json:"enabled,omitempty"`
	Trigger       *ScheduleTrigger     `json:"trigger,omitempty"`
	BackupFilter  *BackupFilter        `json:"backup_filter,omitempty"`
	Verification  *VerificationConfig  `json:"verification,omitempty"`
	Notifications *NotificationConfig  `json:"notifications,omitempty"`
	Retention     *RetentionPolicy     `json:"retention,omitempty"`
	Concurrency   *ConcurrencySettings `json:"concurrency,omitempty"`
	Metadata      map[string]string    `json:"metadata,omitempty"`
}

// ScheduleFilter contains filters for listing schedules
type ScheduleFilter struct {
	TenantID      string       `json:"tenant_id,omitempty"`
	Enabled       *bool        `json:"enabled,omitempty"`
	Type          ScheduleType `json:"type,omitempty"`
	TriggerType   TriggerType  `json:"trigger_type,omitempty"`
	CreatedAfter  *time.Time   `json:"created_after,omitempty"`
	CreatedBefore *time.Time   `json:"created_before,omitempty"`
}

// ScheduleType represents the type of schedule
type ScheduleType string

const (
	ScheduleTypeRecurring  ScheduleType = "recurring"
	ScheduleTypeOnDemand   ScheduleType = "on_demand"
	ScheduleTypeEventBased ScheduleType = "event_based"
)

// ScheduleTrigger defines when verification should run
type ScheduleTrigger struct {
	Type     TriggerType   `json:"type"`
	Config   TriggerConfig `json:"config"`
	Timezone string        `json:"timezone"`
}

// TriggerType represents the type of trigger
type TriggerType string

const (
	TriggerTypeCron     TriggerType = "cron"
	TriggerTypeInterval TriggerType = "interval"
	TriggerTypeEvent    TriggerType = "event"
	TriggerTypeManual   TriggerType = "manual"
)

// TriggerConfig contains trigger-specific configuration
type TriggerConfig struct {
	// Cron trigger
	CronExpression string `json:"cron_expression,omitempty"`

	// Interval trigger
	Interval time.Duration `json:"interval,omitempty"`

	// Event trigger
	Events []string `json:"events,omitempty"`

	// Common settings
	MaxRuns    int       `json:"max_runs,omitempty"`
	RunCount   int       `json:"run_count,omitempty"`
	StartAfter time.Time `json:"start_after,omitempty"`
	EndBefore  time.Time `json:"end_before,omitempty"`
}

// BackupFilter defines which backups to verify
type BackupFilter struct {
	RepositoryIDs  []string      `json:"repository_ids,omitempty"`
	JobIDs         []string      `json:"job_ids,omitempty"`
	BackupTypes    []string      `json:"backup_types,omitempty"`
	BackupMethods  []string      `json:"backup_methods,omitempty"`
	Tags           []string      `json:"tags,omitempty"`
	MinAge         time.Duration `json:"min_age,omitempty"`
	MaxAge         time.Duration `json:"max_age,omitempty"`
	MinSize        int64         `json:"min_size,omitempty"`
	MaxSize        int64         `json:"max_size,omitempty"`
	IncludeFailed  bool          `json:"include_failed,omitempty"`
	ExcludeRunning bool          `json:"exclude_running,omitempty"`
}

// RetentionPolicy defines how long to keep verification results
type RetentionPolicy struct {
	ResultsDays   int `json:"results_days"`
	ReportsDays   int `json:"reports_days"`
	LogsDays      int `json:"logs_days"`
	ArtifactsDays int `json:"artifacts_days"`
	MaxResults    int `json:"max_results"`
	MaxReports    int `json:"max_reports"`
	MaxLogs       int `json:"max_logs"`
	MaxArtifacts  int `json:"max_artifacts"`
}

// ConcurrencySettings defines execution concurrency
type ConcurrencySettings struct {
	MaxConcurrentJobs      int            `json:"max_concurrent_jobs"`
	MaxConcurrentPerBackup int            `json:"max_concurrent_per_backup"`
	QueueSize              int            `json:"queue_size"`
	Priority               JobPriority    `json:"priority"`
	ResourceLimits         ResourceLimits `json:"resource_limits"`
}

// JobPriority represents job priority
type JobPriority string

const (
	JobPriorityLow      JobPriority = "low"
	JobPriorityNormal   JobPriority = "normal"
	JobPriorityHigh     JobPriority = "high"
	JobPriorityCritical JobPriority = "critical"
)

// ResourceLimits defines resource constraints
type ResourceLimits struct {
	MaxCPU     int `json:"max_cpu"`
	MaxMemory  int `json:"max_memory"`
	MaxStorage int `json:"max_storage"`
	MaxNetwork int `json:"max_network"`
}

// VerificationJob represents a verification job execution
type VerificationJob struct {
	ID          string               `json:"id"`
	ScheduleID  string               `json:"schedule_id"`
	TenantID    string               `json:"tenant_id"`
	BackupID    string               `json:"backup_id"`
	Status      JobStatus            `json:"status"`
	Priority    JobPriority          `json:"priority"`
	Config      VerificationConfig   `json:"config"`
	Results     *VerificationResults `json:"results,omitempty"`
	Error       string               `json:"error,omitempty"`
	Progress    JobProgress          `json:"progress"`
	Timing      JobTiming            `json:"timing"`
	Resources   JobResources         `json:"resources"`
	RetryInfo   RetryInfo            `json:"retry_info"`
	CreatedAt   time.Time            `json:"created_at"`
	QueuedAt    *time.Time           `json:"queued_at,omitempty"`
	StartedAt   *time.Time           `json:"started_at,omitempty"`
	CompletedAt *time.Time           `json:"completed_at,omitempty"`
	ExpiresAt   *time.Time           `json:"expires_at,omitempty"`
	Metadata    map[string]string    `json:"metadata"`
}

// JobStatus represents the status of a verification job
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusQueued    JobStatus = "queued"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
	JobStatusExpired   JobStatus = "expired"
	JobStatusRetrying  JobStatus = "retrying"
)

// JobProgress contains progress information
type JobProgress struct {
	Percentage         int           `json:"percentage"`
	CurrentStep        string        `json:"current_step"`
	TotalSteps         int           `json:"total_steps"`
	CompletedSteps     int           `json:"completed_steps"`
	EstimatedRemaining time.Duration `json:"estimated_remaining"`
}

// JobTiming contains timing information
type JobTiming struct {
	QueuedDuration    time.Duration `json:"queued_duration"`
	ExecutionDuration time.Duration `json:"execution_duration"`
	TotalDuration     time.Duration `json:"total_duration"`
	TimeoutDuration   time.Duration `json:"timeout_duration"`
}

// JobResources contains resource usage information
type JobResources struct {
	CPUUsage     float64 `json:"cpu_usage"`
	MemoryUsage  int64   `json:"memory_usage"`
	StorageUsage int64   `json:"storage_usage"`
	NetworkUsage int64   `json:"network_usage"`
	SandboxID    string  `json:"sandbox_id,omitempty"`
}

// RetryInfo contains retry information
type RetryInfo struct {
	Attempt     int           `json:"attempt"`
	MaxAttempts int           `json:"max_attempts"`
	RetryDelay  time.Duration `json:"retry_delay"`
	LastError   string        `json:"last_error,omitempty"`
	NextRetryAt *time.Time    `json:"next_retry_at,omitempty"`
}

// JobFilter contains filters for listing verification jobs
type JobFilter struct {
	TenantID        string      `json:"tenant_id,omitempty"`
	ScheduleID      string      `json:"schedule_id,omitempty"`
	BackupID        string      `json:"backup_id,omitempty"`
	Status          JobStatus   `json:"status,omitempty"`
	Priority        JobPriority `json:"priority,omitempty"`
	CreatedAfter    *time.Time  `json:"created_after,omitempty"`
	CreatedBefore   *time.Time  `json:"created_before,omitempty"`
	CompletedAfter  *time.Time  `json:"completed_after,omitempty"`
	CompletedBefore *time.Time  `json:"completed_before,omitempty"`
}

// TimeRange defines a time range for statistics
type TimeRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// VerificationStats contains verification statistics
type VerificationStats struct {
	TenantID             string        `json:"tenant_id"`
	TimeRange            TimeRange     `json:"time_range"`
	TotalJobs            int64         `json:"total_jobs"`
	CompletedJobs        int64         `json:"completed_jobs"`
	FailedJobs           int64         `json:"failed_jobs"`
	CancelledJobs        int64         `json:"cancelled_jobs"`
	RunningJobs          int64         `json:"running_jobs"`
	QueuedJobs           int64         `json:"queued_jobs"`
	SuccessRate          float64       `json:"success_rate"`
	AverageExecutionTime time.Duration `json:"average_execution_time"`
	MinExecutionTime     time.Duration `json:"min_execution_time"`
	MaxExecutionTime     time.Duration `json:"max_execution_time"`
	TotalExecutionTime   time.Duration `json:"total_execution_time"`
	LastUpdated          time.Time     `json:"last_updated"`
}

// ScheduleStats contains schedule-specific statistics
type ScheduleStats struct {
	ScheduleID     string        `json:"schedule_id"`
	TimeRange      TimeRange     `json:"time_range"`
	TotalRuns      int64         `json:"total_runs"`
	SuccessfulRuns int64         `json:"successful_runs"`
	FailedRuns     int64         `json:"failed_runs"`
	SkippedRuns    int64         `json:"skipped_runs"`
	AverageRunTime time.Duration `json:"average_run_time"`
	LastRunAt      *time.Time    `json:"last_run_at,omitempty"`
	NextRunAt      *time.Time    `json:"next_run_at,omitempty"`
	SuccessRate    float64       `json:"success_rate"`
	LastUpdated    time.Time     `json:"last_updated"`
}

// GlobalVerificationStats contains global verification statistics
type GlobalVerificationStats struct {
	TimeRange            TimeRange     `json:"time_range"`
	TotalTenants         int64         `json:"total_tenants"`
	TotalSchedules       int64         `json:"total_schedules"`
	ActiveSchedules      int64         `json:"active_schedules"`
	TotalJobs            int64         `json:"total_jobs"`
	CompletedJobs        int64         `json:"completed_jobs"`
	FailedJobs           int64         `json:"failed_jobs"`
	RunningJobs          int64         `json:"running_jobs"`
	QueuedJobs           int64         `json:"queued_jobs"`
	SuccessRate          float64       `json:"success_rate"`
	AverageExecutionTime time.Duration `json:"average_execution_time"`
	LastUpdated          time.Time     `json:"last_updated"`
}

// SystemHealth contains system health information
type SystemHealth struct {
	Status          HealthStatus  `json:"status"`
	WorkerNodes     []NodeHealth  `json:"worker_nodes"`
	QueueStatus     QueueHealth   `json:"queue_status"`
	ResourceUsage   ResourceUsage `json:"resource_usage"`
	ErrorRate       float64       `json:"error_rate"`
	ResponseTime    time.Duration `json:"response_time"`
	LastHealthCheck time.Time     `json:"last_health_check"`
	Issues          []HealthIssue `json:"issues"`
}

// HealthStatus represents system health status
type HealthStatus string

const (
	HealthStatusHealthy  HealthStatus = "healthy"
	HealthStatusWarning  HealthStatus = "warning"
	HealthStatusCritical HealthStatus = "critical"
	HealthStatusDown     HealthStatus = "down"
)

// NodeHealth contains health information for a worker node
type NodeHealth struct {
	NodeID       string       `json:"node_id"`
	Status       HealthStatus `json:"status"`
	CPUUsage     float64      `json:"cpu_usage"`
	MemoryUsage  float64      `json:"memory_usage"`
	StorageUsage float64      `json:"storage_usage"`
	ActiveJobs   int          `json:"active_jobs"`
	MaxJobs      int          `json:"max_jobs"`
	LastSeen     time.Time    `json:"last_seen"`
	Version      string       `json:"version"`
}

// QueueHealth contains queue health information
type QueueHealth struct {
	Status          HealthStatus  `json:"status"`
	QueueSize       int           `json:"queue_size"`
	MaxSize         int           `json:"max_size"`
	ProcessingRate  float64       `json:"processing_rate"`
	AverageWaitTime time.Duration `json:"average_wait_time"`
}

// ResourceUsage contains system resource usage
type ResourceUsage struct {
	CPUUsage     float64 `json:"cpu_usage"`
	MemoryUsage  float64 `json:"memory_usage"`
	StorageUsage float64 `json:"storage_usage"`
	NetworkUsage float64 `json:"network_usage"`
}

// HealthIssue represents a system health issue
type HealthIssue struct {
	Severity    string     `json:"severity"`
	Type        string     `json:"type"`
	Description string     `json:"description"`
	NodeID      string     `json:"node_id,omitempty"`
	DetectedAt  time.Time  `json:"detected_at"`
	Resolved    bool       `json:"resolved"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
}

// InMemoryAutoVerificationManager implements AutoVerificationManager interface
type InMemoryAutoVerificationManager struct {
	schedules         map[string]*Schedule
	jobs              map[string]*VerificationJob
	stats             map[string]*VerificationStats
	scheduleStats     map[string]*ScheduleStats
	globalStats       *GlobalVerificationStats
	tenantManager     multitenancy.TenantManager
	storageManager    storage.Engine
	sureBackupManager surebackup.SureBackupManager
	mutex             sync.RWMutex
}

// NewInMemoryAutoVerificationManager creates a new in-memory auto verification manager
func NewInMemoryAutoVerificationManager(
	tenantMgr multitenancy.TenantManager,
	storageMgr storage.Engine,
	sureBackupMgr surebackup.SureBackupManager,
) *InMemoryAutoVerificationManager {
	return &InMemoryAutoVerificationManager{
		schedules:         make(map[string]*Schedule),
		jobs:              make(map[string]*VerificationJob),
		stats:             make(map[string]*VerificationStats),
		scheduleStats:     make(map[string]*ScheduleStats),
		globalStats:       &GlobalVerificationStats{LastUpdated: time.Now()},
		tenantManager:     tenantMgr,
		storageManager:    storageMgr,
		sureBackupManager: sureBackupMgr,
	}
}

// CreateSchedule creates a new verification schedule
func (m *InMemoryAutoVerificationManager) CreateSchedule(ctx context.Context, request *ScheduleRequest) (*Schedule, error) {
	if err := m.validateTenantAccess(ctx, request.TenantID); err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	schedule := &Schedule{
		ID:            generateScheduleID(),
		Name:          request.Name,
		TenantID:      request.TenantID,
		Description:   request.Description,
		Enabled:       request.Enabled,
		Type:          request.Type,
		Trigger:       request.Trigger,
		BackupFilter:  request.BackupFilter,
		Verification:  request.Verification,
		Notifications: request.Notifications,
		Retention:     request.Retention,
		Concurrency:   request.Concurrency,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Metadata:      request.Metadata,
	}

	// Calculate next run time for enabled schedules
	if request.Enabled {
		nextRun := m.calculateNextRun(schedule.Trigger, time.Now())
		schedule.NextRunAt = &nextRun
	}

	m.schedules[schedule.ID] = schedule

	// Initialize schedule statistics
	m.scheduleStats[schedule.ID] = &ScheduleStats{
		ScheduleID:  schedule.ID,
		TimeRange:   TimeRange{From: time.Now(), To: time.Now().Add(24 * time.Hour)},
		LastUpdated: time.Now(),
	}

	return schedule, nil
}

// GetSchedule retrieves a schedule by ID
func (m *InMemoryAutoVerificationManager) GetSchedule(ctx context.Context, scheduleID string) (*Schedule, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	schedule, exists := m.schedules[scheduleID]
	if !exists {
		return nil, fmt.Errorf("schedule %s not found", scheduleID)
	}

	if err := m.validateTenantAccess(ctx, schedule.TenantID); err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	return schedule, nil
}

// ListSchedules lists schedules with optional filtering
func (m *InMemoryAutoVerificationManager) ListSchedules(ctx context.Context, filter *ScheduleFilter) ([]*Schedule, error) {
	if filter != nil && filter.TenantID != "" {
		if err := m.validateTenantAccess(ctx, filter.TenantID); err != nil {
			return nil, fmt.Errorf("access denied: %w", err)
		}
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var results []*Schedule
	for _, schedule := range m.schedules {
		if filter != nil {
			if filter.TenantID != "" && schedule.TenantID != filter.TenantID {
				continue
			}
			if filter.Enabled != nil && schedule.Enabled != *filter.Enabled {
				continue
			}
			if filter.Type != "" && schedule.Type != filter.Type {
				continue
			}
			if filter.TriggerType != "" && schedule.Trigger.Type != filter.TriggerType {
				continue
			}
			if filter.CreatedAfter != nil && schedule.CreatedAt.Before(*filter.CreatedAfter) {
				continue
			}
			if filter.CreatedBefore != nil && schedule.CreatedAt.After(*filter.CreatedBefore) {
				continue
			}
		}

		// Check tenant access
		if err := m.validateTenantAccess(ctx, schedule.TenantID); err != nil {
			continue
		}

		results = append(results, schedule)
	}

	return results, nil
}

// UpdateSchedule updates an existing schedule
func (m *InMemoryAutoVerificationManager) UpdateSchedule(ctx context.Context, scheduleID string, request *UpdateScheduleRequest) (*Schedule, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	schedule, exists := m.schedules[scheduleID]
	if !exists {
		return nil, fmt.Errorf("schedule %s not found", scheduleID)
	}

	if err := m.validateTenantAccess(ctx, schedule.TenantID); err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	// Update fields
	if request.Name != nil {
		schedule.Name = *request.Name
	}
	if request.Description != nil {
		schedule.Description = *request.Description
	}
	if request.Enabled != nil {
		schedule.Enabled = *request.Enabled
	}
	if request.Trigger != nil {
		schedule.Trigger = *request.Trigger
	}
	if request.BackupFilter != nil {
		schedule.BackupFilter = *request.BackupFilter
	}
	if request.Verification != nil {
		schedule.Verification = *request.Verification
	}
	if request.Notifications != nil {
		schedule.Notifications = *request.Notifications
	}
	if request.Retention != nil {
		schedule.Retention = *request.Retention
	}
	if request.Concurrency != nil {
		schedule.Concurrency = *request.Concurrency
	}
	if request.Metadata != nil {
		schedule.Metadata = request.Metadata
	}

	schedule.UpdatedAt = time.Now()

	// Recalculate next run time if enabled
	if schedule.Enabled {
		nextRun := m.calculateNextRun(schedule.Trigger, time.Now())
		schedule.NextRunAt = &nextRun
	} else {
		schedule.NextRunAt = nil
	}

	return schedule, nil
}

// DeleteSchedule deletes a schedule
func (m *InMemoryAutoVerificationManager) DeleteSchedule(ctx context.Context, scheduleID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	schedule, exists := m.schedules[scheduleID]
	if !exists {
		return fmt.Errorf("schedule %s not found", scheduleID)
	}

	if err := m.validateTenantAccess(ctx, schedule.TenantID); err != nil {
		return fmt.Errorf("access denied: %w", err)
	}

	delete(m.schedules, scheduleID)
	delete(m.scheduleStats, scheduleID)

	return nil
}

// EnableSchedule enables a schedule
func (m *InMemoryAutoVerificationManager) EnableSchedule(ctx context.Context, scheduleID string) error {
	_, err := m.UpdateSchedule(ctx, scheduleID, &UpdateScheduleRequest{
		Enabled: &[]bool{true}[0],
	})
	return err
}

// DisableSchedule disables a schedule
func (m *InMemoryAutoVerificationManager) DisableSchedule(ctx context.Context, scheduleID string) error {
	_, err := m.UpdateSchedule(ctx, scheduleID, &UpdateScheduleRequest{
		Enabled: &[]bool{false}[0],
	})
	return err
}

// GetVerificationJob retrieves a verification job by ID
func (m *InMemoryAutoVerificationManager) GetVerificationJob(ctx context.Context, jobID string) (*VerificationJob, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	job, exists := m.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("verification job %s not found", jobID)
	}

	if err := m.validateTenantAccess(ctx, job.TenantID); err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	return job, nil
}

// ListVerificationJobs lists verification jobs with optional filtering
func (m *InMemoryAutoVerificationManager) ListVerificationJobs(ctx context.Context, filter *JobFilter) ([]*VerificationJob, error) {
	if filter != nil && filter.TenantID != "" {
		if err := m.validateTenantAccess(ctx, filter.TenantID); err != nil {
			return nil, fmt.Errorf("access denied: %w", err)
		}
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var results []*VerificationJob
	for _, job := range m.jobs {
		if filter != nil {
			if filter.TenantID != "" && job.TenantID != filter.TenantID {
				continue
			}
			if filter.ScheduleID != "" && job.ScheduleID != filter.ScheduleID {
				continue
			}
			if filter.BackupID != "" && job.BackupID != filter.BackupID {
				continue
			}
			if filter.Status != "" && job.Status != filter.Status {
				continue
			}
			if filter.Priority != "" && job.Priority != filter.Priority {
				continue
			}
			if filter.CreatedAfter != nil && job.CreatedAt.Before(*filter.CreatedAfter) {
				continue
			}
			if filter.CreatedBefore != nil && job.CreatedAt.After(*filter.CreatedBefore) {
				continue
			}
			if filter.CompletedAfter != nil && (job.CompletedAt == nil || job.CompletedAt.Before(*filter.CompletedAfter)) {
				continue
			}
			if filter.CompletedBefore != nil && (job.CompletedAt == nil || job.CompletedAt.After(*filter.CompletedBefore)) {
				continue
			}
		}

		// Check tenant access
		if err := m.validateTenantAccess(ctx, job.TenantID); err != nil {
			continue
		}

		results = append(results, job)
	}

	return results, nil
}

// CancelVerificationJob cancels a verification job
func (m *InMemoryAutoVerificationManager) CancelVerificationJob(ctx context.Context, jobID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	job, exists := m.jobs[jobID]
	if !exists {
		return fmt.Errorf("verification job %s not found", jobID)
	}

	if err := m.validateTenantAccess(ctx, job.TenantID); err != nil {
		return fmt.Errorf("access denied: %w", err)
	}

	if job.Status == JobStatusPending || job.Status == JobStatusQueued || job.Status == JobStatusRunning {
		job.Status = JobStatusCancelled
		now := time.Now()
		job.CompletedAt = &now
	}

	return nil
}

// RerunVerificationJob reruns a verification job
func (m *InMemoryAutoVerificationManager) RerunVerificationJob(ctx context.Context, jobID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	job, exists := m.jobs[jobID]
	if !exists {
		return fmt.Errorf("verification job %s not found", jobID)
	}

	if err := m.validateTenantAccess(ctx, job.TenantID); err != nil {
		return fmt.Errorf("access denied: %w", err)
	}

	// Create new job with same configuration
	newJob := &VerificationJob{
		ID:         generateJobID(),
		ScheduleID: job.ScheduleID,
		TenantID:   job.TenantID,
		BackupID:   job.BackupID,
		Status:     JobStatusPending,
		Priority:   job.Priority,
		Config:     job.Config,
		Progress: JobProgress{
			Percentage:     0,
			CurrentStep:    "queued",
			TotalSteps:     1,
			CompletedSteps: 0,
		},
		Timing: JobTiming{
			TimeoutDuration: job.Timing.TimeoutDuration,
		},
		RetryInfo: RetryInfo{
			Attempt:     1,
			MaxAttempts: job.RetryInfo.MaxAttempts,
			RetryDelay:  job.RetryInfo.RetryDelay,
		},
		CreatedAt: time.Now(),
		Metadata:  job.Metadata,
	}

	m.jobs[newJob.ID] = newJob

	// Start job execution in background
	go m.runVerificationJob(ctx, newJob)

	return nil
}

// GetVerificationStats retrieves verification statistics for a tenant
func (m *InMemoryAutoVerificationManager) GetVerificationStats(ctx context.Context, tenantID string, timeRange TimeRange) (*VerificationStats, error) {
	if err := m.validateTenantAccess(ctx, tenantID); err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := &VerificationStats{
		TenantID:    tenantID,
		TimeRange:   timeRange,
		LastUpdated: time.Now(),
	}

	var totalExecutionTime time.Duration
	var executionTimes []time.Duration

	for _, job := range m.jobs {
		if job.TenantID != tenantID {
			continue
		}

		// Filter by time range
		if job.CreatedAt.Before(timeRange.From) || job.CreatedAt.After(timeRange.To) {
			continue
		}

		stats.TotalJobs++

		switch job.Status {
		case JobStatusCompleted:
			stats.CompletedJobs++
		case JobStatusFailed:
			stats.FailedJobs++
		case JobStatusCancelled:
			stats.CancelledJobs++
		case JobStatusRunning:
			stats.RunningJobs++
		case JobStatusQueued, JobStatusPending:
			stats.QueuedJobs++
		}

		// Calculate execution time for completed jobs
		if job.Timing.ExecutionDuration > 0 {
			totalExecutionTime += job.Timing.ExecutionDuration
			executionTimes = append(executionTimes, job.Timing.ExecutionDuration)
		}
	}

	// Calculate success rate
	if stats.TotalJobs > 0 {
		stats.SuccessRate = float64(stats.CompletedJobs) / float64(stats.TotalJobs)
	}

	// Calculate average, min, max execution times
	if len(executionTimes) > 0 {
		stats.AverageExecutionTime = totalExecutionTime / time.Duration(len(executionTimes))
		stats.MinExecutionTime = executionTimes[0]
		stats.MaxExecutionTime = executionTimes[0]

		for _, et := range executionTimes {
			if et < stats.MinExecutionTime {
				stats.MinExecutionTime = et
			}
			if et > stats.MaxExecutionTime {
				stats.MaxExecutionTime = et
			}
		}
	}

	stats.TotalExecutionTime = totalExecutionTime

	return stats, nil
}

// GetScheduleStats retrieves statistics for a specific schedule
func (m *InMemoryAutoVerificationManager) GetScheduleStats(ctx context.Context, scheduleID string, timeRange TimeRange) (*ScheduleStats, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	schedule, exists := m.schedules[scheduleID]
	if !exists {
		return nil, fmt.Errorf("schedule %s not found", scheduleID)
	}

	if err := m.validateTenantAccess(ctx, schedule.TenantID); err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	stats := &ScheduleStats{
		ScheduleID:  scheduleID,
		TimeRange:   timeRange,
		LastUpdated: time.Now(),
		LastRunAt:   schedule.LastRunAt,
		NextRunAt:   schedule.NextRunAt,
	}

	var totalRunTime time.Duration
	var runTimes []time.Duration

	for _, job := range m.jobs {
		if job.ScheduleID != scheduleID {
			continue
		}

		// Filter by time range
		if job.CreatedAt.Before(timeRange.From) || job.CreatedAt.After(timeRange.To) {
			continue
		}

		stats.TotalRuns++

		switch job.Status {
		case JobStatusCompleted:
			stats.SuccessfulRuns++
		case JobStatusFailed:
			stats.FailedRuns++
		default:
			stats.SkippedRuns++
		}

		// Calculate run time
		if job.Timing.ExecutionDuration > 0 {
			totalRunTime += job.Timing.ExecutionDuration
			runTimes = append(runTimes, job.Timing.ExecutionDuration)
		}
	}

	// Calculate success rate
	if stats.TotalRuns > 0 {
		stats.SuccessRate = float64(stats.SuccessfulRuns) / float64(stats.TotalRuns)
	}

	// Calculate average run time
	if len(runTimes) > 0 {
		stats.AverageRunTime = totalRunTime / time.Duration(len(runTimes))
	}

	return stats, nil
}

// GetGlobalStats retrieves global verification statistics
func (m *InMemoryAutoVerificationManager) GetGlobalStats(ctx context.Context, timeRange TimeRange) (*GlobalVerificationStats, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := &GlobalVerificationStats{
		TimeRange:   timeRange,
		LastUpdated: time.Now(),
	}

	tenants := make(map[string]bool)

	for _, schedule := range m.schedules {
		tenants[schedule.TenantID] = true
		if schedule.Enabled {
			stats.ActiveSchedules++
		}
	}

	stats.TotalTenants = int64(len(tenants))
	stats.TotalSchedules = int64(len(m.schedules))

	var totalExecutionTime time.Duration
	var executionTimes []time.Duration

	for _, job := range m.jobs {
		// Filter by time range
		if job.CreatedAt.Before(timeRange.From) || job.CreatedAt.After(timeRange.To) {
			continue
		}

		stats.TotalJobs++
		tenants[job.TenantID] = true

		switch job.Status {
		case JobStatusCompleted:
			stats.CompletedJobs++
		case JobStatusFailed:
			stats.FailedJobs++
		case JobStatusRunning:
			stats.RunningJobs++
		case JobStatusQueued, JobStatusPending:
			stats.QueuedJobs++
		}

		// Calculate execution time for completed jobs
		if job.Timing.ExecutionDuration > 0 {
			totalExecutionTime += job.Timing.ExecutionDuration
			executionTimes = append(executionTimes, job.Timing.ExecutionDuration)
		}
	}

	// Calculate success rate
	if stats.TotalJobs > 0 {
		stats.SuccessRate = float64(stats.CompletedJobs) / float64(stats.TotalJobs)
	}

	// Calculate average execution time
	if len(executionTimes) > 0 {
		stats.AverageExecutionTime = totalExecutionTime / time.Duration(len(executionTimes))
	}

	return stats, nil
}

// GetSystemHealth retrieves system health information
func (m *InMemoryAutoVerificationManager) GetSystemHealth(ctx context.Context) (*SystemHealth, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	health := &SystemHealth{
		Status:          HealthStatusHealthy,
		WorkerNodes:     []NodeHealth{},
		QueueStatus:     QueueHealth{Status: HealthStatusHealthy, QueueSize: 0, MaxSize: 1000},
		ResourceUsage:   ResourceUsage{CPUUsage: 10.5, MemoryUsage: 45.2, StorageUsage: 23.8, NetworkUsage: 5.1},
		ErrorRate:       0.02,
		ResponseTime:    150 * time.Millisecond,
		LastHealthCheck: time.Now(),
		Issues:          []HealthIssue{},
	}

	// Count jobs by status for queue status
	runningJobs := 0
	pendingJobs := 0
	for _, job := range m.jobs {
		if job.Status == JobStatusRunning {
			runningJobs++
		}
		if job.Status == JobStatusPending || job.Status == JobStatusQueued {
			pendingJobs++
		}
	}

	health.QueueStatus.QueueSize = pendingJobs
	health.QueueStatus.ProcessingRate = float64(runningJobs) / 100.0 // jobs per minute

	// Add mock worker node
	health.WorkerNodes = append(health.WorkerNodes, NodeHealth{
		NodeID:       "worker-1",
		Status:       HealthStatusHealthy,
		CPUUsage:     15.2,
		MemoryUsage:  42.8,
		StorageUsage: 28.1,
		ActiveJobs:   runningJobs,
		MaxJobs:      10,
		LastSeen:     time.Now().Add(-5 * time.Minute),
		Version:      "1.0.0",
	})

	return health, nil
}

// GetPendingJobs retrieves pending verification jobs
func (m *InMemoryAutoVerificationManager) GetPendingJobs(ctx context.Context) ([]*VerificationJob, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var pendingJobs []*VerificationJob
	for _, job := range m.jobs {
		if job.Status == JobStatusPending || job.Status == JobStatusQueued {
			// Check tenant access
			if err := m.validateTenantAccess(ctx, job.TenantID); err != nil {
				continue
			}
			pendingJobs = append(pendingJobs, job)
		}
	}

	return pendingJobs, nil
}

// Helper methods

func (m *InMemoryAutoVerificationManager) calculateNextRun(trigger ScheduleTrigger, from time.Time) time.Time {
	switch trigger.Type {
	case TriggerTypeCron:
		// Simple mock implementation - add 1 hour for cron schedules
		return from.Add(1 * time.Hour)
	case TriggerTypeInterval:
		return from.Add(trigger.Config.Interval)
	case TriggerTypeEvent:
		// Event triggers don't have predictable next run times
		return time.Time{}
	case TriggerTypeManual:
		// Manual triggers don't have next run times
		return time.Time{}
	default:
		return from.Add(24 * time.Hour)
	}
}

func (m *InMemoryAutoVerificationManager) runVerificationJob(ctx context.Context, job *VerificationJob) {
	m.mutex.Lock()
	job.Status = JobStatusRunning
	now := time.Now()
	job.StartedAt = &now
	job.Progress.CurrentStep = "starting_verification"
	m.mutex.Unlock()

	// Simulate verification execution
	time.Sleep(2 * time.Second)

	m.mutex.Lock()
	job.Status = JobStatusCompleted
	completedAt := time.Now()
	job.CompletedAt = &completedAt
	job.Progress.Percentage = 100
	job.Progress.CurrentStep = "completed"
	job.Progress.CompletedSteps = 1

	// Set timing
	job.Timing.ExecutionDuration = completedAt.Sub(*job.StartedAt)
	job.Timing.TotalDuration = completedAt.Sub(job.CreatedAt)

	// Set mock results
	job.Results = &surebackup.VerificationResults{
		Passed:   true,
		Duration: job.Timing.ExecutionDuration,
		TestResults: []surebackup.TestResult{
			{
				Name:     "backup_integrity",
				Type:     "integrity",
				Passed:   true,
				Duration: 1 * time.Second,
				Output:   "Backup integrity verified successfully",
			},
		},
		Summary: surebackup.VerificationSummary{
			TotalTests:      1,
			PassedTests:     1,
			FailedTests:     0,
			DataIntegrityOK: true,
			MountSuccess:    true,
			BootSuccess:     true,
		},
	}
	m.mutex.Unlock()
}

func (m *InMemoryAutoVerificationManager) validateTenantAccess(ctx context.Context, tenantID string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant ID cannot be empty")
	}

	// Get tenant from context
	contextTenantID := m.tenantManager.GetTenantFromContext(ctx)
	if contextTenantID == "" {
		return fmt.Errorf("tenant context not found")
	}

	if contextTenantID != tenantID {
		return fmt.Errorf("tenant access denied")
	}

	return nil
}

func generateScheduleID() string {
	return fmt.Sprintf("schedule-%s", uuid.New().String()[:8])
}

func generateJobID() string {
	return fmt.Sprintf("job-%s", uuid.New().String()[:8])
}
