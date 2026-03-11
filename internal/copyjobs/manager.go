package copyjobs

import (
	"context"
	"fmt"
	"sync"
	"time"

	"novabackup/internal/deduplication"
	"novabackup/internal/multitenancy"
)

// BackupCopyJob represents a backup copy job
type BackupCopyJob struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	SourceRepo string                 `json:"source_repo"`
	TargetRepo string                 `json:"target_repo"`
	BackupType string                 `json:"backup_type"`
	Schedule   string                 `json:"schedule"`
	Enabled    bool                   `json:"enabled"`
	Priority   Priority               `json:"priority"`
	TenantID   string                 `json:"tenant_id"`
	CreatedAt  time.Time              `json:"created_at"`
	LastRunAt  *time.Time             `json:"last_run_at"`
	NextRunAt  *time.Time             `json:"next_run_at"`
	Status     JobStatus              `json:"status"`
	Progress   JobProgress            `json:"progress"`
	Settings   map[string]interface{} `json:"settings"`
	Metadata   map[string]string      `json:"metadata"`
}

// JobStatus represents the status of a backup copy job
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusPaused    JobStatus = "paused"
	JobStatusCancelled JobStatus = "cancelled"
)

// JobProgress represents the progress of a backup copy job
type JobProgress struct {
	Percentage       float64    `json:"percentage"`
	BytesTransferred int64      `json:"bytes_transferred"`
	TotalBytes       int64      `json:"total_bytes"`
	CurrentFile      string     `json:"current_file"`
	EstimatedETA     *time.Time `json:"estimated_eta"`
	StartedAt        time.Time  `json:"started_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// Priority defines job priority levels
type Priority string

const (
	PriorityLow      Priority = "low"
	PriorityNormal   Priority = "normal"
	PriorityHigh     Priority = "high"
	PriorityCritical Priority = "critical"
)

// BackupCopyManager manages backup copy jobs
type BackupCopyManager interface {
	// Job management
	CreateJob(ctx context.Context, job *BackupCopyJob) error
	GetJob(ctx context.Context, jobID string) (*BackupCopyJob, error)
	UpdateJob(ctx context.Context, job *BackupCopyJob) error
	DeleteJob(ctx context.Context, jobID string) error
	ListJobs(ctx context.Context, filter *JobFilter) ([]BackupCopyJob, error)

	// Job execution
	StartJob(ctx context.Context, jobID string) error
	StopJob(ctx context.Context, jobID string) error
	PauseJob(ctx context.Context, jobID string) error
	ResumeJob(ctx context.Context, jobID string) error

	// Scheduling
	ScheduleJob(ctx context.Context, jobID string, schedule string) error
	UnscheduleJob(ctx context.Context, jobID string) error
	GetNextRun(ctx context.Context) (*BackupCopyJob, error)

	// Statistics
	GetJobStats(ctx context.Context, jobID string) (*JobStats, error)
	GetManagerStats(ctx context.Context) (*ManagerStats, error)
}

// JobFilter filters backup copy jobs
type JobFilter struct {
	TenantID   string    `json:"tenant_id"`
	Status     JobStatus `json:"status"`
	Priority   Priority  `json:"priority"`
	SourceRepo string    `json:"source_repo"`
	TargetRepo string    `json:"target_repo"`
	Limit      int       `json:"limit"`
	Offset     int       `json:"offset"`
}

// JobStats contains statistics for a specific job
type JobStats struct {
	JobID            string        `json:"job_id"`
	TotalRuns        int64         `json:"total_runs"`
	SuccessfulRuns   int64         `json:"successful_runs"`
	FailedRuns       int64         `json:"failed_runs"`
	AverageRunTime   time.Duration `json:"average_run_time"`
	TotalBytesCopied int64         `json:"total_bytes_copied"`
	LastRunAt        time.Time     `json:"last_run_at"`
	NextScheduledRun *time.Time    `json:"next_scheduled_run"`
}

// ManagerStats contains backup copy manager statistics
type ManagerStats struct {
	TotalJobs        int64     `json:"total_jobs"`
	ActiveJobs       int       `json:"active_jobs"`
	PendingJobs      int64     `json:"pending_jobs"`
	CompletedJobs    int64     `json:"completed_jobs"`
	FailedJobs       int64     `json:"failed_jobs"`
	TotalBytesCopied int64     `json:"total_bytes_copied"`
	LastActivity     time.Time `json:"last_activity"`
}

// CopyResult represents the result of a backup copy operation
type CopyResult struct {
	JobID       string            `json:"job_id"`
	Success     bool              `json:"success"`
	BytesCopied int64             `json:"bytes_copied"`
	Duration    time.Duration     `json:"duration"`
	Error       string            `json:"error,omitempty"`
	FilesCopied int               `json:"files_copied"`
	StartedAt   time.Time         `json:"started_at"`
	CompletedAt time.Time         `json:"completed_at"`
	Metadata    map[string]string `json:"metadata"`
}

// InMemoryBackupCopyManager implements BackupCopyManager in memory
type InMemoryBackupCopyManager struct {
	jobs      map[string]*BackupCopyJob
	running   map[string]*JobExecution
	scheduler *CopyJobScheduler
	tenantMgr multitenancy.TenantManager
	dedupeMgr deduplication.DeduplicationManager
	mutex     sync.RWMutex
	stats     *ManagerStats
}

// JobExecution represents a running job execution
type JobExecution struct {
	Job        *BackupCopyJob
	StartTime  time.Time
	StopChan   chan struct{}
	Progress   *JobProgress
	Context    context.Context
	CancelFunc context.CancelFunc
}

// JobScheduler manages job scheduling
type JobScheduler struct {
	jobs     map[string]*BackupCopyJob
	queue    []string
	ticker   *time.Ticker
	stopChan chan struct{}
	mutex    sync.RWMutex
}

// NewInMemoryBackupCopyManager creates a new in-memory backup copy manager
func NewInMemoryBackupCopyManager(tenantMgr multitenancy.TenantManager, dedupeMgr deduplication.DeduplicationManager) *InMemoryBackupCopyManager {
	manager := &InMemoryBackupCopyManager{
		jobs:      make(map[string]*BackupCopyJob),
		running:   make(map[string]*JobExecution),
		scheduler: NewCopyJobScheduler(),
		tenantMgr: tenantMgr,
		dedupeMgr: dedupeMgr,
		mutex:     sync.RWMutex{},
		stats:     &ManagerStats{},
	}

	// Start background tasks
	go manager.startBackgroundTasks()

	return manager
}

// CreateJob creates a new backup copy job
func (m *InMemoryBackupCopyManager) CreateJob(ctx context.Context, job *BackupCopyJob) error {
	// Validate tenant context
	tenantID := m.tenantMgr.GetTenantFromContext(ctx)
	if tenantID == "" {
		return fmt.Errorf("tenant ID required for backup copy job creation")
	}

	job.TenantID = tenantID
	job.ID = fmt.Sprintf("copy-job-%d", time.Now().UnixNano())
	job.CreatedAt = time.Now()
	job.Status = JobStatusPending
	job.Progress = JobProgress{
		Percentage: 0.0,
		UpdatedAt:  time.Now(),
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.jobs[job.ID] = job
	m.scheduler.AddJob(job)

	// Update statistics
	m.stats.TotalJobs++
	m.stats.PendingJobs++

	return nil
}

// GetJob retrieves a backup copy job by ID
func (m *InMemoryBackupCopyManager) GetJob(ctx context.Context, jobID string) (*BackupCopyJob, error) {
	// Validate tenant context
	tenantID := m.tenantMgr.GetTenantFromContext(ctx)
	if tenantID == "" {
		return nil, fmt.Errorf("tenant ID required for backup copy job access")
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	job, exists := m.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("backup copy job %s not found", jobID)
	}

	// Check tenant access
	if job.TenantID != tenantID {
		return nil, fmt.Errorf("access denied: tenant %s cannot access job %s", tenantID, jobID)
	}

	return job, nil
}

// UpdateJob updates an existing backup copy job
func (m *InMemoryBackupCopyManager) UpdateJob(ctx context.Context, job *BackupCopyJob) error {
	// Validate tenant context
	tenantID := m.tenantMgr.GetTenantFromContext(ctx)
	if tenantID == "" {
		return fmt.Errorf("tenant ID required for backup copy job update")
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	existingJob, exists := m.jobs[job.ID]
	if !exists {
		return fmt.Errorf("backup copy job %s not found", job.ID)
	}

	// Check tenant access
	if existingJob.TenantID != tenantID {
		return fmt.Errorf("access denied: tenant %s cannot update job %s", tenantID, job.ID)
	}

	// Update job
	existingJob.Name = job.Name
	existingJob.SourceRepo = job.SourceRepo
	existingJob.TargetRepo = job.TargetRepo
	existingJob.BackupType = job.BackupType
	existingJob.Schedule = job.Schedule
	existingJob.Enabled = job.Enabled
	existingJob.Priority = job.Priority
	existingJob.Settings = job.Settings
	existingJob.Metadata = job.Metadata
	existingJob.Progress.UpdatedAt = time.Now()

	return nil
}

// DeleteJob deletes a backup copy job
func (m *InMemoryBackupCopyManager) DeleteJob(ctx context.Context, jobID string) error {
	// Validate tenant context
	tenantID := m.tenantMgr.GetTenantFromContext(ctx)
	if tenantID == "" {
		return fmt.Errorf("tenant ID required for backup copy job deletion")
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	job, exists := m.jobs[jobID]
	if !exists {
		return fmt.Errorf("backup copy job %s not found", jobID)
	}

	// Check tenant access
	if job.TenantID != tenantID {
		return fmt.Errorf("access denied: tenant %s cannot delete job %s", tenantID, jobID)
	}

	// Stop job if running
	if _, running := m.running[jobID]; running {
		m.StopJob(ctx, jobID)
	}

	delete(m.jobs, jobID)
	m.scheduler.RemoveJob(jobID)

	// Update statistics
	m.stats.TotalJobs--
	switch job.Status {
	case JobStatusPending:
		m.stats.PendingJobs--
	case JobStatusRunning:
		m.stats.ActiveJobs--
	case JobStatusCompleted:
		m.stats.CompletedJobs--
	case JobStatusFailed:
		m.stats.FailedJobs--
	}

	return nil
}

// ListJobs lists backup copy jobs with optional filtering
func (m *InMemoryBackupCopyManager) ListJobs(ctx context.Context, filter *JobFilter) ([]BackupCopyJob, error) {
	// Validate tenant context
	tenantID := m.tenantMgr.GetTenantFromContext(ctx)
	if tenantID == "" {
		return nil, fmt.Errorf("tenant ID required for backup copy job listing")
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var jobs []BackupCopyJob
	for _, job := range m.jobs {
		// Apply tenant filter
		if filter.TenantID != "" && job.TenantID != filter.TenantID {
			continue
		}

		// Apply status filter
		if filter.Status != "" && job.Status != filter.Status {
			continue
		}

		// Apply priority filter
		if filter.Priority != "" && job.Priority != filter.Priority {
			continue
		}

		// Apply source repo filter
		if filter.SourceRepo != "" && job.SourceRepo != filter.SourceRepo {
			continue
		}

		// Apply target repo filter
		if filter.TargetRepo != "" && job.TargetRepo != filter.TargetRepo {
			continue
		}

		jobs = append(jobs, *job)
	}

	// Apply pagination
	if filter.Offset > 0 && filter.Offset < len(jobs) {
		end := filter.Offset + filter.Limit
		if end > len(jobs) {
			end = len(jobs)
		}
		jobs = jobs[filter.Offset:end]
	}

	return jobs, nil
}

// StartJob starts a backup copy job
func (m *InMemoryBackupCopyManager) StartJob(ctx context.Context, jobID string) error {
	// Validate tenant context
	tenantID := m.tenantMgr.GetTenantFromContext(ctx)
	if tenantID == "" {
		return fmt.Errorf("tenant ID required for backup copy job start")
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	job, exists := m.jobs[jobID]
	if !exists {
		return fmt.Errorf("backup copy job %s not found", jobID)
	}

	// Check tenant access
	if job.TenantID != tenantID {
		return fmt.Errorf("access denied: tenant %s cannot start job %s", tenantID, jobID)
	}

	if job.Status == JobStatusRunning {
		return fmt.Errorf("backup copy job %s is already running", jobID)
	}

	// Create job execution context
	jobCtx, cancel := context.WithCancel(ctx)
	execution := &JobExecution{
		Job:       job,
		StartTime: time.Now(),
		StopChan:  make(chan struct{}),
		Progress: &JobProgress{
			Percentage: 0.0,
			UpdatedAt:  time.Now(),
		},
		Context:    jobCtx,
		CancelFunc: cancel,
	}

	m.running[jobID] = execution

	// Update job status
	job.Status = JobStatusRunning
	job.LastRunAt = &execution.StartTime
	job.Progress.StartedAt = execution.StartTime

	// Update statistics
	m.stats.ActiveJobs++
	if job.Status == JobStatusPending {
		m.stats.PendingJobs--
	}

	// Start the copy process in background
	go m.executeJob(execution)

	return nil
}

// StopJob stops a running backup copy job
func (m *InMemoryBackupCopyManager) StopJob(ctx context.Context, jobID string) error {
	// Validate tenant context
	tenantID := m.tenantMgr.GetTenantFromContext(ctx)
	if tenantID == "" {
		return fmt.Errorf("tenant ID required for backup copy job stop")
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	execution, running := m.running[jobID]
	if !running {
		return fmt.Errorf("backup copy job %s is not running", jobID)
	}

	// Check tenant access
	if execution.Job.TenantID != tenantID {
		return fmt.Errorf("access denied: tenant %s cannot stop job %s", tenantID, jobID)
	}

	// Signal stop
	close(execution.StopChan)

	// Update job status
	execution.Job.Status = JobStatusPaused
	execution.Progress.UpdatedAt = time.Now()

	return nil
}

// PauseJob pauses a running backup copy job
func (m *InMemoryBackupCopyManager) PauseJob(ctx context.Context, jobID string) error {
	return m.StopJob(ctx, jobID)
}

// ResumeJob resumes a paused backup copy job
func (m *InMemoryBackupCopyManager) ResumeJob(ctx context.Context, jobID string) error {
	// Validate tenant context
	tenantID := m.tenantMgr.GetTenantFromContext(ctx)
	if tenantID == "" {
		return fmt.Errorf("tenant ID required for backup copy job resume")
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	job, exists := m.jobs[jobID]
	if !exists {
		return fmt.Errorf("backup copy job %s not found", jobID)
	}

	// Check tenant access
	if job.TenantID != tenantID {
		return fmt.Errorf("access denied: tenant %s cannot resume job %s", tenantID, jobID)
	}

	if job.Status != JobStatusPaused {
		return fmt.Errorf("backup copy job %s is not paused", jobID)
	}

	// Update job status
	job.Status = JobStatusRunning
	job.Progress.UpdatedAt = time.Now()

	// Update statistics
	m.stats.ActiveJobs++

	// Resume execution
	if execution, exists := m.running[jobID]; exists {
		// Create new execution context
		jobCtx, cancel := context.WithCancel(ctx)
		execution.Context = jobCtx
		execution.CancelFunc = cancel
		execution.StopChan = make(chan struct{})

		m.running[jobID] = execution

		// Start the copy process in background
		go m.executeJob(execution)
	}

	return nil
}

// ScheduleJob schedules a backup copy job
func (m *InMemoryBackupCopyManager) ScheduleJob(ctx context.Context, jobID string, schedule string) error {
	// Validate tenant context
	tenantID := m.tenantMgr.GetTenantFromContext(ctx)
	if tenantID == "" {
		return fmt.Errorf("tenant ID required for backup copy job scheduling")
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	job, exists := m.jobs[jobID]
	if !exists {
		return fmt.Errorf("backup copy job %s not found", jobID)
	}

	// Check tenant access
	if job.TenantID != tenantID {
		return fmt.Errorf("access denied: tenant %s cannot schedule job %s", tenantID, jobID)
	}

	job.Schedule = schedule
	m.scheduler.ScheduleJob(jobID, schedule)

	return nil
}

// UnscheduleJob unschedules a backup copy job
func (m *InMemoryBackupCopyManager) UnscheduleJob(ctx context.Context, jobID string) error {
	// Validate tenant context
	tenantID := m.tenantMgr.GetTenantFromContext(ctx)
	if tenantID == "" {
		return fmt.Errorf("tenant ID required for backup copy job unscheduling")
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	job, exists := m.jobs[jobID]
	if !exists {
		return fmt.Errorf("backup copy job %s not found", jobID)
	}

	// Check tenant access
	if job.TenantID != tenantID {
		return fmt.Errorf("access denied: tenant %s cannot unschedule job %s", tenantID, jobID)
	}

	job.Schedule = ""
	m.scheduler.UnscheduleJob(jobID)

	return nil
}

// GetNextRun retrieves the next scheduled job to run
func (m *InMemoryBackupCopyManager) GetNextRun(ctx context.Context) (*BackupCopyJob, error) {
	// Validate tenant context
	tenantID := m.tenantMgr.GetTenantFromContext(ctx)
	if tenantID == "" {
		return nil, fmt.Errorf("tenant ID required for next run retrieval")
	}

	return m.scheduler.GetNextRun(tenantID, m.jobs)
}

// GetJobStats retrieves statistics for a specific job
func (m *InMemoryBackupCopyManager) GetJobStats(ctx context.Context, jobID string) (*JobStats, error) {
	// Validate tenant context
	tenantID := m.tenantMgr.GetTenantFromContext(ctx)
	if tenantID == "" {
		return nil, fmt.Errorf("tenant ID required for job statistics retrieval")
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	job, exists := m.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("backup copy job %s not found", jobID)
	}

	// Check tenant access
	if job.TenantID != tenantID {
		return nil, fmt.Errorf("access denied: tenant %s cannot access stats for job %s", tenantID, jobID)
	}

	// Calculate job statistics (simplified for in-memory implementation)
	stats := &JobStats{
		JobID:            jobID,
		TotalRuns:        1, // Simplified
		SuccessfulRuns:   0, // Simplified
		FailedRuns:       0, // Simplified
		AverageRunTime:   0, // Simplified
		TotalBytesCopied: 0, // Simplified
		LastRunAt:        time.Now(),
		NextScheduledRun: job.NextRunAt,
	}

	if job.LastRunAt != nil {
		stats.LastRunAt = *job.LastRunAt
		stats.TotalRuns = 1
		if job.Status == JobStatusCompleted {
			stats.SuccessfulRuns = 1
		} else if job.Status == JobStatusFailed {
			stats.FailedRuns = 1
		}
	}

	return stats, nil
}

// GetManagerStats retrieves backup copy manager statistics
func (m *InMemoryBackupCopyManager) GetManagerStats(ctx context.Context) (*ManagerStats, error) {
	// Validate tenant context
	tenantID := m.tenantMgr.GetTenantFromContext(ctx)
	if tenantID == "" {
		return nil, fmt.Errorf("tenant ID required for manager statistics retrieval")
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Filter jobs by tenant
	var tenantJobs int64
	var activeJobs int
	var pendingJobs int64
	var completedJobs int64
	var failedJobs int64

	for _, job := range m.jobs {
		if job.TenantID == tenantID {
			tenantJobs++
			switch job.Status {
			case JobStatusPending:
				pendingJobs++
			case JobStatusRunning:
				activeJobs++
			case JobStatusCompleted:
				completedJobs++
			case JobStatusFailed:
				failedJobs++
			}
		}
	}

	// Update stats snapshot
	m.stats.TotalJobs = tenantJobs
	m.stats.ActiveJobs = activeJobs
	m.stats.PendingJobs = pendingJobs
	m.stats.CompletedJobs = completedJobs
	m.stats.FailedJobs = failedJobs
	m.stats.LastActivity = time.Now()

	return m.stats, nil
}

// executeJob executes a backup copy job
func (m *InMemoryBackupCopyManager) executeJob(execution *JobExecution) {
	defer close(execution.StopChan)

	// Simulate backup copy process
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	startTime := time.Now()
	var totalBytes int64
	var filesCopied int

	for {
		select {
		case <-ticker.C:
			// Update progress
			elapsed := time.Since(startTime)
			totalBytes += 1024 * 1024 // Simulate 1MB per second
			filesCopied++

			percentage := float64(totalBytes) / float64(100*1024*1024) * 100
			if percentage > 100 {
				percentage = 100
			}

			execution.Progress.Percentage = percentage
			execution.Progress.BytesTransferred = totalBytes
			execution.Progress.TotalBytes = 100 * 1024 * 1024 // Simulate 100MB total
			execution.Progress.CurrentFile = fmt.Sprintf("backup_file_%d.txt", filesCopied)
			execution.Progress.UpdatedAt = time.Now()

			// Estimate ETA
			if percentage > 0 && percentage < 100 {
				remaining := float64(execution.Progress.TotalBytes-totalBytes) / (float64(totalBytes) / float64(elapsed.Seconds()))
				eta := time.Now().Add(time.Duration(remaining * float64(time.Second)))
				execution.Progress.EstimatedETA = &eta
			}

			// Check for completion
			if percentage >= 100 {
				execution.Job.Status = JobStatusCompleted
				execution.Job.Progress.UpdatedAt = time.Now()
				completedAt := time.Now()

				// Update statistics
				m.mutex.Lock()
				m.stats.ActiveJobs--
				m.stats.CompletedJobs++
				m.stats.TotalBytesCopied += totalBytes
				m.stats.LastActivity = completedAt
				m.mutex.Unlock()

				// Clean up execution
				delete(m.running, execution.Job.ID)

				return
			}

		case <-execution.StopChan:
			// Job was stopped
			execution.Job.Status = JobStatusPaused
			execution.Job.Progress.UpdatedAt = time.Now()

			// Update statistics
			m.mutex.Lock()
			m.stats.ActiveJobs--
			m.stats.PendingJobs++
			m.stats.LastActivity = time.Now()
			m.mutex.Unlock()

			// Clean up execution
			delete(m.running, execution.Job.ID)

			return

		case <-execution.Context.Done():
			// Context was cancelled
			execution.Job.Status = JobStatusCancelled
			execution.Job.Progress.UpdatedAt = time.Now()

			// Update statistics
			m.mutex.Lock()
			m.stats.ActiveJobs--
			m.stats.FailedJobs++
			m.stats.LastActivity = time.Now()
			m.mutex.Unlock()

			// Clean up execution
			delete(m.running, execution.Job.ID)

			return
		}
	}
}

// startBackgroundTasks starts background tasks for the manager
func (m *InMemoryBackupCopyManager) startBackgroundTasks() {
	// Start scheduler
	go m.scheduler.Start()

	// Cleanup routine
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cleanupCompletedJobs()
		}
	}
}

// cleanupCompletedJobs cleans up completed jobs older than 24 hours
func (m *InMemoryBackupCopyManager) cleanupCompletedJobs() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	cutoff := time.Now().Add(-24 * time.Hour)

	for jobID, job := range m.jobs {
		if job.Status == JobStatusCompleted && job.LastRunAt != nil && job.LastRunAt.Before(cutoff) {
			// Archive or cleanup old completed jobs
			delete(m.jobs, jobID)
			m.scheduler.RemoveJob(jobID)
		}
	}
}
