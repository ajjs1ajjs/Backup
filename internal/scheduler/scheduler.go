package scheduler

import (
	"novabackup/internal/database"
	"time"
)

type Scheduler struct {
	db *database.Database
}

func NewScheduler(db *database.Database) *Scheduler {
	return &Scheduler{
		db: db,
	}
}

func (s *Scheduler) Start() error {
	// Start scheduler ticker
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			s.checkAndRunJobs()
		}
	}()

	return nil
}

func (s *Scheduler) checkAndRunJobs() {
	jobs, err := s.db.ListJobs()
	if err != nil {
		return
	}

	now := time.Now()
	for _, job := range jobs {
		if !job.Enabled {
			continue
		}

		if job.NextRun != nil && job.NextRun.Before(now) {
			// TODO: Run the job
			nextRun := s.calculateNextRun(job.Schedule)
			s.db.UpdateJobLastRun(job.ID, now, nextRun)
		}
	}
}

func (s *Scheduler) calculateNextRun(schedule string) time.Time {
	now := time.Now()

	switch schedule {
	case "daily":
		return now.Add(24 * time.Hour)
	case "weekly":
		return now.Add(7 * 24 * time.Hour)
	case "monthly":
		return now.Add(30 * 24 * time.Hour)
	default:
		return now.Add(24 * time.Hour)
	}
}

func (s *Scheduler) Stop() {
	// Cleanup if needed
}
