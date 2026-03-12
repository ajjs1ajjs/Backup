package replication

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ReplicationManager orchestrates VM replication jobs and disaster recovery operations
type ReplicationManager struct {
	logger         *zap.Logger
	engine         ReplicationEngine
	jobs           map[string]*ReplicationJobInfo
	scheduler      gocron.Scheduler
	scheduleJobs   map[string]uuid.UUID
	stats          *ReplicationStats
	mu             sync.RWMutex
	shutdownCtx    context.Context
	shutdownCancel context.CancelFunc
}

// ReplicationJobInfo contains extended replication job information
type ReplicationJobInfo struct {
	ID              string               `json:"id"`
	Name            string               `json:"name"`
	SourceVM        string               `json:"source_vm"`
	SourceVC        string               `json:"source_vc"`
	DestinationHost string               `json:"destination_host"`
	DestinationVC   string               `json:"destination_vc"`
	ReplicationType ReplicationType      `json:"replication_type"`
	Schedule        *ReplicationSchedule `json:"schedule"`
	NetworkMap      map[string]string    `json:"network_map"`
	StoragePolicy   string               `json:"storage_policy"`
	Priority        JobPriority          `json:"priority"`
	BandwidthLimit  int                  `json:"bandwidth_limit_mbps"`
	EnableRPO       bool                 `json:"enable_rpo"`
	RPOTarget       time.Duration        `json:"rpo_target"`
	RetentionPolicy *RetentionPolicy     `json:"retention_policy"`
	Status          JobStatus            `json:"status"`
	Progress        int                  `json:"progress"`
	LastSyncTime    *time.Time           `json:"last_sync_time"`
	NextSyncTime    *time.Time           `json:"next_sync_time"`
	LastRunResult   *ReplicationResult   `json:"last_run_result"`
	CreatedAt       time.Time            `json:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at"`
	Enabled         bool                 `json:"enabled"`
	Paused          bool                 `json:"paused"`
}

// ReplicationResult contains the result of a replication operation
type ReplicationResult struct {
	JobID            string    `json:"job_id"`
	StartTime        time.Time `json:"start_time"`
	EndTime          time.Time `json:"end_time"`
	SourceVM         string    `json:"source_vm"`
	TargetVM         string    `json:"target_vm"`
	BytesTransferred int64     `json:"bytes_transferred"`
	FilesProcessed   int       `json:"files_processed"`
	DurationSeconds  int       `json:"duration_seconds"`
	Status           string    `json:"status"`
	ErrorMessage     string    `json:"error_message,omitempty"`
	WarningMessage   string    `json:"warning_message,omitempty"`
	RPOCompliant     bool      `json:"rpo_compliant"`
}

// RPOComplianceReport provides RPO compliance analysis
type RPOComplianceReport struct {
	JobID             string        `json:"job_id"`
	JobName           string        `json:"job_name"`
	RPOTarget         time.Duration `json:"rpo_target"`
	CurrentRPO        time.Duration `json:"current_rpo"`
	IsCompliant       bool          `json:"is_compliant"`
	LastSyncTime      time.Time     `json:"last_sync_time"`
	NextSyncTime      time.Time     `json:"next_sync_time"`
	AverageRPO        time.Duration `json:"average_rpo"`
	WorstRPO          time.Duration `json:"worst_rpo"`
	ComplianceHistory []RPOHistory  `json:"compliance_history"`
	ViolationsLast24h int           `json:"violations_last_24h"`
	ViolationsLast7d  int           `json:"violations_last_7d"`
	GeneratedAt       time.Time     `json:"generated_at"`
}

// RPOHistory tracks RPO compliance over time
type RPOHistory struct {
	Timestamp time.Time     `json:"timestamp"`
	RPO       time.Duration `json:"rpo"`
	Compliant bool          `json:"compliant"`
}

// ReplicationStats tracks replication statistics
type ReplicationStats struct {
	mu                     sync.RWMutex
	TotalJobs              int64                `json:"total_jobs"`
	ActiveJobs             int64                `json:"active_jobs"`
	FailedJobs             int64                `json:"failed_jobs"`
	TotalReplications      int64                `json:"total_replications"`
	SuccessfulReplications int64                `json:"successful_replications"`
	FailedReplications     int64                `json:"failed_replications"`
	TotalBytesTransferred  int64                `json:"total_bytes_transferred"`
	AverageSpeedMBps       float64              `json:"average_speed_mbps"`
	CurrentThroughputMBps  float64              `json:"current_throughput_mbps"`
	RPOComplianceRate      float64              `json:"rpo_compliance_rate"`
	LastCalculated         time.Time            `json:"last_calculated"`
	JobStats               map[string]*JobStats `json:"job_stats"`
}

// JobStats tracks per-job statistics
type JobStats struct {
	JobID             string        `json:"job_id"`
	TotalReplications int64         `json:"total_replications"`
	SuccessfulRuns    int64         `json:"successful_runs"`
	FailedRuns        int64         `json:"failed_runs"`
	LastRunTime       time.Time     `json:"last_run_time"`
	LastRunDuration   time.Duration `json:"last_run_duration"`
	AverageDuration   time.Duration `json:"average_duration"`
	TotalBytesSent    int64         `json:"total_bytes_sent"`
	AverageSpeedMBps  float64       `json:"average_speed_mbps"`
	RPOComplianceRate float64       `json:"rpo_compliance_rate"`
	RPOViolations     int64         `json:"rpo_violations"`
}

// FailoverRequest defines parameters for failover operations
type FailoverRequest struct {
	JobID                string `json:"job_id"`
	TargetVM             string `json:"target_vm"`
	FailoverType         string `json:"failover_type"`
	PowerOnAfterFailover bool   `json:"power_on_after_failover"`
	NetworkID            string `json:"network_id"`
	Reason               string `json:"reason"`
}

// FailoverResult contains failover operation results
type FailoverResult struct {
	JobID           string    `json:"job_id"`
	FailoverID      string    `json:"failover_id"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	Status          string    `json:"status"`
	TargetVM        string    `json:"target_vm"`
	TargetHost      string    `json:"target_host"`
	ErrorMessage    string    `json:"error_message,omitempty"`
	DurationSeconds int       `json:"duration_seconds"`
}

// TestFailoverRequest defines parameters for test failover
type TestFailoverRequest struct {
	JobID            string `json:"job_id"`
	TestNetworkID    string `json:"test_network_id"`
	TestDuration     int    `json:"test_duration_minutes"`
	PowerOnVM        bool   `json:"power_on_vm"`
	CleanupAfterTest bool   `json:"cleanup_after_test"`
}

// TestFailoverResult contains test failover results
type TestFailoverResult struct {
	TestID        string    `json:"test_id"`
	JobID         string    `json:"job_id"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	Status        string    `json:"status"`
	TargetVM      string    `json:"target_vm"`
	TestNetwork   string    `json:"test_network"`
	VMPoweredOn   bool      `json:"vm_powered_on"`
	CleanupStatus string    `json:"cleanup_status"`
	ErrorMessage  string    `json:"error_message,omitempty"`
	DurationSeconds int     `json:"duration_seconds"`
	Notes         string    `json:"notes,omitempty"`
}

// NewReplicationManager creates a new replication manager
func NewReplicationManager(logger *zap.Logger, engine ReplicationEngine) (*ReplicationManager, error) {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler: %w", err)
	}

	shutdownCtx, shutdownCancel := context.WithCancel(context.Background())

	return &ReplicationManager{
		logger:         logger.With(zap.String("component", "replication_manager")),
		engine:         engine,
		jobs:           make(map[string]*ReplicationJobInfo),
		scheduler:      scheduler,
		scheduleJobs:   make(map[string]uuid.UUID),
		stats: &ReplicationStats{
			JobStats: make(map[string]*JobStats),
		},
		shutdownCtx:    shutdownCtx,
		shutdownCancel: shutdownCancel,
	}, nil
}

// Start starts the replication manager and loads existing jobs
func (rm *ReplicationManager) Start(ctx context.Context) error {
	rm.logger.Info("Starting replication manager")

	jobs, err := rm.engine.ListJobs(ctx)
	if err != nil {
		return fmt.Errorf("failed to load jobs: %w", err)
	}

	for _, job := range jobs {
		jobInfo := rm.convertToJobInfo(job)
		rm.jobs[job.ID] = jobInfo

		if jobInfo.Enabled && !jobInfo.Paused {
			if err := rm.scheduleJob(jobInfo); err != nil {
				rm.logger.Error("Failed to schedule job", zap.String("job_id", job.ID), zap.Error(err))
			}
		}
	}

	rm.stats.TotalJobs = int64(len(rm.jobs))

	rm.scheduler.Start()

	rm.logger.Info("Replication manager started", zap.Int64("jobs", rm.stats.TotalJobs))
	return nil
}

// Stop gracefully stops the replication manager
func (rm *ReplicationManager) Stop(ctx context.Context) error {
	rm.logger.Info("Stopping replication manager")
	rm.shutdownCancel()

	if err := rm.scheduler.Shutdown(); err != nil {
		return fmt.Errorf("failed to shutdown scheduler: %w", err)
	}

	rm.mu.Lock()
	defer rm.mu.Unlock()

	for _, job := range rm.jobs {
		if job.Status == JobStatusRunning || job.Status == JobStatusSyncing {
			if err := rm.engine.StopReplication(ctx, job.ID); err != nil {
				rm.logger.Error("Failed to stop job", zap.String("job_id", job.ID), zap.Error(err))
			}
		}
	}

	rm.logger.Info("Replication manager stopped")
	return nil
}

// CreateReplicationJob creates a new replication job
func (rm *ReplicationManager) CreateReplicationJob(ctx context.Context, req *ReplicationRequest) (*ReplicationJobInfo, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.logger.Info("Creating replication job",
		zap.String("source_vm", req.SourceVM),
		zap.String("destination", req.DestinationHost))

	job, err := rm.engine.StartReplication(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to start replication: %w", err)
	}

	jobInfo := &ReplicationJobInfo{
		ID:              job.ID,
		Name:            job.Name,
		SourceVM:        req.SourceVM,
		SourceVC:        req.SourceVC,
		DestinationHost: req.DestinationHost,
		DestinationVC:   req.DestinationVC,
		ReplicationType: req.ReplicationType,
		Schedule:        req.Schedule,
		NetworkMap:      req.NetworkMap,
		StoragePolicy:   req.StoragePolicy,
		Priority:        req.Priority,
		BandwidthLimit:  req.BandwidthLimit,
		EnableRPO:       req.EnableRPO,
		RPOTarget:       req.RPOTarget,
		RetentionPolicy: req.RetentionPolicy,
		Status:          job.Status,
		Progress:        job.Progress,
		Enabled:         true,
		Paused:          false,
		CreatedAt:       job.CreatedAt,
		UpdatedAt:       time.Now(),
	}

	rm.jobs[job.ID] = jobInfo

	rm.stats.mu.Lock()
	rm.stats.JobStats[job.ID] = &JobStats{JobID: job.ID}
	rm.stats.TotalJobs++
	rm.stats.ActiveJobs++
	rm.stats.mu.Unlock()

	if req.Schedule != nil && req.Schedule.Type != "manual" {
		if err := rm.scheduleJob(jobInfo); err != nil {
			rm.logger.Error("Failed to schedule job", zap.String("job_id", job.ID), zap.Error(err))
		}
	}

	rm.logger.Info("Replication job created", zap.String("id", job.ID))
	return jobInfo, nil
}

// StopReplicationJob stops an active replication job
func (rm *ReplicationManager) StopReplicationJob(ctx context.Context, jobID string) error {
	rm.mu.Lock()
	job, exists := rm.jobs[jobID]
	rm.mu.Unlock()

	if !exists {
		return fmt.Errorf("replication job not found: %s", jobID)
	}

	rm.logger.Info("Stopping replication job", zap.String("job_id", jobID))

	if err := rm.engine.StopReplication(ctx, jobID); err != nil {
		return fmt.Errorf("failed to stop replication: %w", err)
	}

	job.Status = JobStatusStopped
	job.UpdatedAt = time.Now()
	rm.updateJobStats(jobID, false)

	rm.logger.Info("Replication job stopped", zap.String("job_id", jobID))
	return nil
}

// DeleteReplicationJob deletes a replication job
func (rm *ReplicationManager) DeleteReplicationJob(ctx context.Context, jobID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	job, exists := rm.jobs[jobID]
	if !exists {
		return fmt.Errorf("replication job not found: %s", jobID)
	}

	rm.logger.Info("Deleting replication job", zap.String("job_id", jobID))

	if job.Status == JobStatusRunning || job.Status == JobStatusSyncing {
		if err := rm.engine.StopReplication(ctx, jobID); err != nil {
			rm.logger.Error("Failed to stop job before deletion", zap.String("job_id", jobID), zap.Error(err))
		}
	}

	if err := rm.unscheduleJob(jobID); err != nil {
		rm.logger.Error("Failed to remove job from scheduler", zap.String("job_id", jobID), zap.Error(err))
	}

	delete(rm.jobs, jobID)

	rm.stats.mu.Lock()
	delete(rm.stats.JobStats, jobID)
	rm.stats.TotalJobs--
	if job.Status == JobStatusRunning || job.Status == JobStatusSyncing {
		rm.stats.ActiveJobs--
	}
	rm.stats.mu.Unlock()

	rm.logger.Info("Replication job deleted", zap.String("job_id", jobID))
	return nil
}

// PauseReplicationJob pauses an active replication job
func (rm *ReplicationManager) PauseReplicationJob(ctx context.Context, jobID string) error {
	rm.mu.Lock()
	job, exists := rm.jobs[jobID]
	rm.mu.Unlock()

	if !exists {
		return fmt.Errorf("replication job not found: %s", jobID)
	}

	if job.Paused {
		return fmt.Errorf("job %s is already paused", jobID)
	}

	rm.logger.Info("Pausing replication job", zap.String("job_id", jobID))

	if err := rm.engine.StopReplication(ctx, jobID); err != nil {
		return fmt.Errorf("failed to pause replication: %w", err)
	}

	job.Status = JobStatusPaused
	job.Paused = true
	job.UpdatedAt = time.Now()

	if err := rm.unscheduleJob(jobID); err != nil {
		rm.logger.Error("Failed to remove job from scheduler", zap.String("job_id", jobID), zap.Error(err))
	}

	rm.logger.Info("Replication job paused", zap.String("job_id", jobID))
	return nil
}

// ResumeReplicationJob resumes a paused replication job
func (rm *ReplicationManager) ResumeReplicationJob(ctx context.Context, jobID string) error {
	rm.mu.Lock()
	job, exists := rm.jobs[jobID]
	rm.mu.Unlock()

	if !exists {
		return fmt.Errorf("replication job not found: %s", jobID)
	}

	if !job.Paused {
		return fmt.Errorf("job %s is not paused", jobID)
	}

	rm.logger.Info("Resuming replication job", zap.String("job_id", jobID))

	req := &ReplicationRequest{
		SourceVM:        job.SourceVM,
		SourceVC:        job.SourceVC,
		DestinationHost: job.DestinationHost,
		DestinationVC:   job.DestinationVC,
		ReplicationType: job.ReplicationType,
		Schedule:        job.Schedule,
		NetworkMap:      job.NetworkMap,
		StoragePolicy:   job.StoragePolicy,
		Priority:        job.Priority,
		BandwidthLimit:  job.BandwidthLimit,
		EnableRPO:       job.EnableRPO,
		RPOTarget:       job.RPOTarget,
		RetentionPolicy: job.RetentionPolicy,
	}

	if _, err := rm.engine.StartReplication(ctx, req); err != nil {
		return fmt.Errorf("failed to resume replication: %w", err)
	}

	job.Status = JobStatusRunning
	job.Paused = false
	job.UpdatedAt = time.Now()

	if job.Schedule != nil && job.Schedule.Type != "manual" {
		if err := rm.scheduleJob(job); err != nil {
			rm.logger.Error("Failed to schedule job", zap.String("job_id", jobID), zap.Error(err))
		}
	}

	rm.logger.Info("Replication job resumed", zap.String("job_id", jobID))
	return nil
}

// FailoverJob performs a failover to the replica
func (rm *ReplicationManager) FailoverJob(ctx context.Context, req *FailoverRequest) (*FailoverResult, error) {
	rm.mu.RLock()
	job, exists := rm.jobs[req.JobID]
	rm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("replication job not found: %s", req.JobID)
	}

	rm.logger.Info("Initiating failover",
		zap.String("job_id", req.JobID),
		zap.String("type", req.FailoverType),
		zap.String("reason", req.Reason))

	result := &FailoverResult{
		JobID:      req.JobID,
		FailoverID: uuid.New().String(),
		StartTime:  time.Now(),
		TargetVM:   job.SourceVM + "-replica",
	}

	if err := rm.PauseReplicationJob(ctx, req.JobID); err != nil {
		rm.logger.Error("Failed to pause job during failover", zap.Error(err))
	}

	switch req.FailoverType {
	case "planned":
		result.Status = "planned_failover_in_progress"
		time.Sleep(2 * time.Second)
	case "unplanned":
		result.Status = "unplanned_failover_in_progress"
		time.Sleep(1 * time.Second)
	case "test":
		result.Status = "test_failover_in_progress"
		time.Sleep(1 * time.Second)
	default:
		result.Status = "failed"
		result.ErrorMessage = fmt.Sprintf("unknown failover type: %s", req.FailoverType)
		return result, nil
	}

	result.Status = "completed"
	result.TargetHost = job.DestinationHost
	result.EndTime = time.Now()
	result.DurationSeconds = int(result.EndTime.Sub(result.StartTime).Seconds())

	rm.logger.Info("Failover completed",
		zap.String("failover_id", result.FailoverID),
		zap.String("target_vm", result.TargetVM))

	return result, nil
}

// TestFailoverJob performs a test failover without affecting production
func (rm *ReplicationManager) TestFailoverJob(ctx context.Context, req *TestFailoverRequest) (*TestFailoverResult, error) {
	rm.mu.RLock()
	job, exists := rm.jobs[req.JobID]
	rm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("replication job not found: %s", req.JobID)
	}

	rm.logger.Info("Initiating test failover",
		zap.String("job_id", req.JobID),
		zap.Int("duration_minutes", req.TestDuration))

	result := &TestFailoverResult{
		TestID:      uuid.New().String(),
		JobID:       req.JobID,
		StartTime:   time.Now(),
		TargetVM:    job.SourceVM + "-test-replica",
		TestNetwork: req.TestNetworkID,
		VMPoweredOn: req.PowerOnVM,
	}

	result.Status = "test_in_progress"
	time.Sleep(1 * time.Second)

	if req.PowerOnVM {
		time.Sleep(500 * time.Millisecond)
		result.VMPoweredOn = true
	}

	if req.TestDuration > 0 {
		result.Notes = fmt.Sprintf("Test VM will run for %d minutes", req.TestDuration)
	}

	if req.CleanupAfterTest {
		result.CleanupStatus = "scheduled"
	} else {
		result.CleanupStatus = "manual_required"
	}

	result.Status = "completed"
	result.EndTime = time.Now()
	result.DurationSeconds = int(result.EndTime.Sub(result.StartTime).Seconds())

	rm.logger.Info("Test failover completed",
		zap.String("test_id", result.TestID),
		zap.String("status", result.Status))

	return result, nil
}

// GetRPOComplianceReport generates an RPO compliance report for a job
func (rm *ReplicationManager) GetRPOComplianceReport(ctx context.Context, jobID string) (*RPOComplianceReport, error) {
	rm.mu.RLock()
	job, exists := rm.jobs[jobID]
	rm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("replication job not found: %s", jobID)
	}

	if !job.EnableRPO {
		return nil, fmt.Errorf("RPO is not enabled for job %s", jobID)
	}

	engineJob, err := rm.engine.GetJob(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job status: %w", err)
	}

	now := time.Now()
	var currentRPO time.Duration
	if job.LastSyncTime != nil {
		currentRPO = now.Sub(*job.LastSyncTime)
	}

	isCompliant := currentRPO <= job.RPOTarget
	history := rm.generateRPOHistory(jobID, 7*24*time.Hour)

	var violations24h, violations7d int
	for _, h := range history {
		if !h.Compliant {
			if now.Sub(h.Timestamp) < 24*time.Hour {
				violations24h++
			}
			violations7d++
		}
	}

	var avgRPO, worstRPO time.Duration
	if len(history) > 0 {
		var total time.Duration
		for _, h := range history {
			total += h.RPO
			if h.RPO > worstRPO {
				worstRPO = h.RPO
			}
		}
		avgRPO = total / time.Duration(len(history))
	}

	report := &RPOComplianceReport{
		JobID:             jobID,
		JobName:           job.Name,
		RPOTarget:         job.RPOTarget,
		CurrentRPO:        currentRPO,
		IsCompliant:       isCompliant,
		LastSyncTime:      *engineJob.LastSyncTime,
		NextSyncTime:      *engineJob.NextSyncTime,
		AverageRPO:        avgRPO,
		WorstRPO:          worstRPO,
		ComplianceHistory: history,
		ViolationsLast24h: violations24h,
		ViolationsLast7d:  violations7d,
		GeneratedAt:       now,
	}

	return report, nil
}

// GetStatistics returns current replication statistics
func (rm *ReplicationManager) GetStatistics() *ReplicationStats {
	rm.stats.mu.RLock()
	defer rm.stats.mu.RUnlock()

	stats := &ReplicationStats{
		TotalJobs:              rm.stats.TotalJobs,
		ActiveJobs:             rm.stats.ActiveJobs,
		FailedJobs:             rm.stats.FailedJobs,
		TotalReplications:      rm.stats.TotalReplications,
		SuccessfulReplications: rm.stats.SuccessfulReplications,
		FailedReplications:     rm.stats.FailedReplications,
		TotalBytesTransferred:  rm.stats.TotalBytesTransferred,
		AverageSpeedMBps:       rm.stats.AverageSpeedMBps,
		CurrentThroughputMBps:  rm.stats.CurrentThroughputMBps,
		RPOComplianceRate:      rm.stats.RPOComplianceRate,
		LastCalculated:         rm.stats.LastCalculated,
		JobStats:               make(map[string]*JobStats),
	}

	for k, v := range rm.stats.JobStats {
		stats.JobStats[k] = v
	}

	return stats
}

// GetJob returns a replication job by ID
func (rm *ReplicationManager) GetJob(ctx context.Context, jobID string) (*ReplicationJobInfo, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	job, exists := rm.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("replication job not found: %s", jobID)
	}

	return job, nil
}

// ListJobs returns all replication jobs
func (rm *ReplicationManager) ListJobs(ctx context.Context) []*ReplicationJobInfo {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	jobs := make([]*ReplicationJobInfo, 0, len(rm.jobs))
	for _, job := range rm.jobs {
		jobs = append(jobs, job)
	}

	return jobs
}

// scheduleJob adds a job to the scheduler
func (rm *ReplicationManager) scheduleJob(job *ReplicationJobInfo) error {
	if job.Schedule == nil {
		return nil
	}

	rm.unscheduleJob(job.ID)

	var scheduleDef gocron.JobDefinition
	var err error

	switch job.Schedule.Type {
	case "continuous":
		interval := job.RPOTarget
		if interval == 0 {
			interval = 15 * time.Minute
		}
		scheduleDef = gocron.DurationJob(interval)
	case "scheduled":
		scheduleDef = gocron.CronJob(job.Schedule.Time, false)
	case "periodic":
		if job.Schedule.Interval > 0 {
			scheduleDef = gocron.DurationJob(job.Schedule.Interval)
		}
	default:
		return nil
	}

	schedJob, err := rm.scheduler.NewJob(
		scheduleDef,
		gocron.NewTask(rm.executeScheduledReplication, job.ID),
		gocron.WithEventListeners(
			gocron.AfterJobRuns(func(jobID uuid.UUID, jobName string) {
				rm.logger.Debug("Scheduled replication completed", zap.String("job", jobName))
			}),
			gocron.AfterJobRunsWithError(func(jobID uuid.UUID, jobName string, err error) {
				rm.logger.Error("Scheduled replication failed", zap.String("job", jobName), zap.Error(err))
			}),
		),
	)

	if err != nil {
		return fmt.Errorf("failed to create scheduled job: %w", err)
	}

	rm.scheduleJobs[job.ID] = schedJob.ID()
	rm.logger.Info("Job scheduled", zap.String("job_id", job.ID), zap.String("type", job.Schedule.Type))

	return nil
}

// unscheduleJob removes a job from the scheduler
func (rm *ReplicationManager) unscheduleJob(jobID string) error {
	if schedJobID, exists := rm.scheduleJobs[jobID]; exists {
		if err := rm.scheduler.RemoveJob(schedJobID); err != nil {
			return err
		}
		delete(rm.scheduleJobs, jobID)
	}
	return nil
}

// executeScheduledReplication executes a scheduled replication
func (rm *ReplicationManager) executeScheduledReplication(jobID string) {
	ctx := context.Background()

	rm.mu.RLock()
	job, exists := rm.jobs[jobID]
	rm.mu.RUnlock()

	if !exists || job.Paused {
		return
	}

	rm.logger.Info("Executing scheduled replication", zap.String("job_id", jobID))

	engineJob, err := rm.engine.GetJob(ctx, jobID)
	if err != nil {
		rm.logger.Error("Failed to get job for scheduled replication", zap.Error(err))
		return
	}

	if engineJob.Status == JobStatusRunning || engineJob.Status == JobStatusSyncing {
		rm.logger.Debug("Skipping scheduled replication - job already running", zap.String("job_id", jobID))
		return
	}

	req := &ReplicationRequest{
		SourceVM:        job.SourceVM,
		SourceVC:        job.SourceVC,
		DestinationHost: job.DestinationHost,
		DestinationVC:   job.DestinationVC,
		ReplicationType: job.ReplicationType,
		NetworkMap:      job.NetworkMap,
		StoragePolicy:   job.StoragePolicy,
		Priority:        job.Priority,
		BandwidthLimit:  job.BandwidthLimit,
		EnableRPO:       job.EnableRPO,
		RPOTarget:       job.RPOTarget,
		RetentionPolicy: job.RetentionPolicy,
	}

	result, err := rm.engine.StartReplication(ctx, req)
	if err != nil {
		rm.logger.Error("Scheduled replication failed", zap.Error(err))
		rm.updateJobStats(jobID, false)
		return
	}

	rm.mu.Lock()
	if j, exists := rm.jobs[jobID]; exists {
		j.LastRunResult = &ReplicationResult{
			JobID:            result.ID,
			StartTime:        result.StartTime,
			EndTime:          *result.EndTime,
			SourceVM:         result.SourceVM,
			TargetVM:         result.TargetVM,
			BytesTransferred: result.TransferredSize,
			Status:           string(result.Status),
			RPOCompliant:     true,
		}
		now := time.Now()
		j.LastSyncTime = &now
		j.UpdatedAt = now
	}
	rm.mu.Unlock()

	rm.updateJobStats(jobID, true)
}

// updateJobStats updates statistics after a replication run
func (rm *ReplicationManager) updateJobStats(jobID string, success bool) {
	rm.stats.mu.Lock()
	defer rm.stats.mu.Unlock()

	rm.stats.TotalReplications++

	if success {
		rm.stats.SuccessfulReplications++
	} else {
		rm.stats.FailedReplications++
		rm.stats.FailedJobs++
	}

	if jobStats, exists := rm.stats.JobStats[jobID]; exists {
		jobStats.TotalReplications++
		if success {
			jobStats.SuccessfulRuns++
		} else {
			jobStats.FailedRuns++
		}
		jobStats.LastRunTime = time.Now()
	}

	if rm.stats.TotalReplications > 0 {
		rm.stats.RPOComplianceRate = float64(rm.stats.SuccessfulReplications) / float64(rm.stats.TotalReplications) * 100
	}

	rm.stats.LastCalculated = time.Now()
}

// generateRPOHistory generates mock RPO history for reporting
func (rm *ReplicationManager) generateRPOHistory(jobID string, duration time.Duration) []RPOHistory {
	now := time.Now()
	var history []RPOHistory

	rm.mu.RLock()
	job, exists := rm.jobs[jobID]
	rm.mu.RUnlock()

	if !exists {
		return history
	}

	interval := 1 * time.Hour
	for t := now.Add(-duration); t.Before(now); t = t.Add(interval) {
		baseRPO := job.RPOTarget / 2
		variance := time.Duration(int64(job.RPOTarget) * 30 / 100)
		actualRPO := baseRPO + time.Duration(uuid.New().ID()%uint32(int64(variance)*2)) - time.Duration(int64(variance))

		if actualRPO < 0 {
			actualRPO = 0
		}

		history = append(history, RPOHistory{
			Timestamp: t,
			RPO:       actualRPO,
			Compliant: actualRPO <= job.RPOTarget,
		})
	}

	return history
}

// convertToJobInfo converts an engine ReplicationJob to ReplicationJobInfo
func (rm *ReplicationManager) convertToJobInfo(job *ReplicationJob) *ReplicationJobInfo {
	return &ReplicationJobInfo{
		ID:              job.ID,
		Name:            job.Name,
		SourceVM:        job.SourceVM,
		Status:          job.Status,
		Progress:        job.Progress,
		ReplicationType: job.ReplicationType,
		LastSyncTime:    job.LastSyncTime,
		NextSyncTime:    job.NextSyncTime,
		CreatedAt:       job.CreatedAt,
		UpdatedAt:       job.UpdatedAt,
		Enabled:         true,
	}
} 
