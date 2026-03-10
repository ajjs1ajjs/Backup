package scheduler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"novabackup/internal/backup"
	"novabackup/internal/database"
	"novabackup/pkg/models"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
)

// Scheduler handles scheduled backup jobs
type Scheduler struct {
	scheduler gocron.Scheduler
	db        *database.Connection
	jobs      map[uuid.UUID]uuid.UUID
	mu        sync.RWMutex
}

// NewScheduler creates a new scheduler
func NewScheduler(db *database.Connection) (*Scheduler, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	return &Scheduler{
		scheduler: s,
		db:        db,
		jobs:      make(map[uuid.UUID]uuid.UUID),
	}, nil
}

// Start starts the scheduler
func (s *Scheduler) Start() error {
	// Load all enabled jobs from database
	jobs, err := s.db.GetAllJobs()
	if err != nil {
		return fmt.Errorf("failed to load jobs: %w", err)
	}

	for _, job := range jobs {
		if job.Enabled && job.Schedule != "" {
			if err := s.AddJob(&job); err != nil {
				log.Printf("Failed to add job %s: %v", job.ID, err)
			}
		}
	}

	log.Println("Scheduler started with", len(s.jobs), "jobs")
	return nil
}

// Stop stops the scheduler
func (s *Scheduler) Stop() error {
	return s.scheduler.Shutdown()
}

// AddJob adds a job to the scheduler
func (s *Scheduler) AddJob(job *models.Job) error {
	if job.Schedule == "" {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Create job with cron schedule
	j, err := s.scheduler.NewJob(
		gocron.CronJob(job.Schedule, false),
		gocron.NewTask(s.executeBackup, job),
		gocron.WithEventListeners(
			gocron.AfterJobRuns(func(jobID uuid.UUID, jobName string) {
				log.Printf("Job %s completed successfully", job.ID)
			}),
			gocron.AfterJobRunsWithError(func(jobID uuid.UUID, jobName string, err error) {
				log.Printf("Job %s failed: %v", job.ID, err)
			}),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	s.jobs[job.ID] = j.ID()
	log.Printf("Scheduled job %s (%s) with schedule: %s", job.ID, job.Name, job.Schedule)
	return nil
}

// RemoveJob removes a job from the scheduler
func (s *Scheduler) RemoveJob(jobID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if id, ok := s.jobs[jobID]; ok {
		if err := s.scheduler.RemoveJob(id); err != nil {
			return err
		}
		delete(s.jobs, jobID)
		log.Printf("Removed job %s from scheduler", jobID)
	}
	return nil
}

// executeBackup executes a backup job
func (s *Scheduler) executeBackup(job *models.Job) {
	ctx := context.Background()

	config := &models.BackupConfig{
		Source:      job.Source,
		Destination: job.Destination,
		Compress:    true,
		Encrypt:     false,
		ChunkSize:   4 * 1024 * 1024,
		BufferSize:  256,
	}

	engine := backup.NewBackupEngine(s.db, config)
	result, err := engine.PerformBackup(ctx, job.Source)
	if err != nil {
		log.Printf("Backup job %s failed: %v", job.ID, err)
		return
	}

	log.Printf("Backup job %s completed: %d files, %d bytes read, %d bytes written",
		job.ID, result.FilesTotal, result.BytesRead, result.BytesWritten)
}

// GetJobCount returns the number of scheduled jobs
func (s *Scheduler) GetJobCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.jobs)
}

// GetNextRun returns the next run time for a job
func (s *Scheduler) GetNextRun(jobID uuid.UUID) (time.Time, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if id, ok := s.jobs[jobID]; ok {
		jobs := s.scheduler.Jobs()
		for _, j := range jobs {
			if j.ID() == id {
				nextRun, _ := j.NextRun()
				return nextRun, nil
			}
		}
	}

	return time.Time{}, fmt.Errorf("job not found")
}
