package copyjobs

import (
	"sync"
	"time"
)

// CopyJobScheduler manages backup copy job scheduling
type CopyJobScheduler struct {
	jobs     map[string]*BackupCopyJob
	queue    []string
	ticker   *time.Ticker
	stopChan chan struct{}
	mutex    sync.RWMutex
}

// NewCopyJobScheduler creates a new job scheduler
func NewCopyJobScheduler() *CopyJobScheduler {
	return &CopyJobScheduler{
		jobs:     make(map[string]*BackupCopyJob),
		queue:    make([]string, 0),
		ticker:   time.NewTicker(1 * time.Minute),
		stopChan: make(chan struct{}),
	}
}

// Start starts the job scheduler
func (s *CopyJobScheduler) Start() {
	go s.run()
}

// AddJob adds a job to the scheduler
func (s *JobScheduler) AddJob(job *BackupCopyJob) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.jobs[job.ID] = job

	// Add to queue if scheduled
	if job.Schedule != "" {
		s.queue = append(s.queue, job.ID)
	}
}

// RemoveJob removes a job from the scheduler
func (s *JobScheduler) RemoveJob(jobID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.jobs, jobID)

	// Remove from queue
	for i, id := range s.queue {
		if id == jobID {
			s.queue = append(s.queue[:i], s.queue[i+1:]...)
			break
		}
	}
}

// ScheduleJob schedules a job with the given schedule
func (s *JobScheduler) ScheduleJob(jobID string, schedule string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if job, exists := s.jobs[jobID]; exists {
		job.Schedule = schedule

		// Add to queue if scheduled
		if schedule != "" {
			// Remove from queue first (to avoid duplicates)
			s.RemoveJob(jobID)
			s.queue = append(s.queue, jobID)
		}
	}
}

// UnscheduleJob unschedules a job
func (s *JobScheduler) UnscheduleJob(jobID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if job, exists := s.jobs[jobID]; exists {
		job.Schedule = ""
		s.RemoveJob(jobID)
	}
}

// GetNextRun retrieves the next job to run
func (s *JobScheduler) GetNextRun(tenantID string, jobs map[string]*BackupCopyJob) (*BackupCopyJob, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	now := time.Now()

	// Find the next scheduled job
	for _, jobID := range s.queue {
		if job, exists := jobs[jobID]; exists &&
			job.TenantID == tenantID &&
			job.Status == JobStatusPending &&
			job.Schedule != "" {

			// Check if it's time to run
			if s.shouldRunJob(job, now) {
				return job, nil
			}
		}
	}

	return nil, nil
}

// shouldRunJob checks if a job should run based on its schedule
func (s *JobScheduler) shouldRunJob(job *BackupCopyJob, now time.Time) bool {
	// Simple scheduling - in a real implementation, this would parse cron expressions
	// For now, we'll use simple time-based scheduling

	// For demo purposes, run jobs every 5 minutes
	if job.LastRunAt == nil || now.Sub(*job.LastRunAt) >= 5*time.Minute {
		return true
	}

	return false
}

// run runs the scheduler main loop
func (s *CopyJobScheduler) run() {
	for {
		select {
		case <-s.ticker.C:
			// Check for scheduled jobs to run
			// This is a simplified implementation
			// In a real system, this would check actual schedules

		case <-s.stopChan:
			return
		}
	}
}

// Stop stops the job scheduler
func (s *JobScheduler) Stop() {
	close(s.stopChan)
}
