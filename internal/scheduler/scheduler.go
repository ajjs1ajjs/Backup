// Scheduler - Veeam-style job scheduling with Cron support
package scheduler

import (
	"context"
	"fmt"
	"novabackup/internal/backup"
	"novabackup/internal/database"
	"strings"
	"sync"
	"time"
)

// Schedule types
const (
	ScheduleManual  = "manual"
	ScheduleDaily   = "daily"
	ScheduleWeekly  = "weekly"
	ScheduleMonthly = "monthly"
	ScheduleCron    = "cron"
	ScheduleHourly  = "hourly"
)

// ScheduledJob represents a job in the scheduler
type ScheduledJob struct {
	JobID          string
	JobName        string
	ScheduleType   string
	ScheduleTime   string
	ScheduleDays   []string
	CronExpression string
	NextRun        time.Time
	LastRun        *time.Time
	Enabled        bool
	job            *backup.BackupJob
}

// Scheduler manages backup job scheduling
type Scheduler struct {
	db           *database.Database
	backupEngine *backup.BackupEngine
	jobs         map[string]*ScheduledJob
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	ticker       *time.Ticker
}

// NewScheduler creates a new scheduler
func NewScheduler(db *database.Database) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())

	return &Scheduler{
		db:     db,
		jobs:   make(map[string]*ScheduledJob),
		ctx:    ctx,
		cancel: cancel,
		ticker: time.NewTicker(1 * time.Minute), // Check every minute
	}
}

// SetBackupEngine sets the backup engine for executing jobs
func (s *Scheduler) SetBackupEngine(engine *backup.BackupEngine) {
	s.backupEngine = engine
}

// Start starts the scheduler
func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Load jobs from database
	if err := s.loadJobs(); err != nil {
		return fmt.Errorf("помилка завантаження завдань: %v", err)
	}

	// Start scheduler loop
	go s.run()

	return nil
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cancel()
	s.ticker.Stop()
}

// run is the main scheduler loop
func (s *Scheduler) run() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-s.ticker.C:
			s.checkAndRunJobs()
		}
	}
}

// checkAndRunJobs checks for jobs that should run
func (s *Scheduler) checkAndRunJobs() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()

	for _, job := range s.jobs {
		if !job.Enabled {
			continue
		}

		if job.NextRun.IsZero() {
			continue
		}

		// Check if job should run (within 1 minute window)
		if now.After(job.NextRun) || now.Equal(job.NextRun) {
			// Run job in background
			go s.executeJob(job)
		}
	}
}

// executeJob executes a scheduled backup job
func (s *Scheduler) executeJob(job *ScheduledJob) {
	if s.backupEngine == nil {
		return
	}

	fmt.Printf("🕐 Запуск запланованого завдання: %s\n", job.JobName)

	// Execute backup
	session, err := s.backupEngine.ExecuteBackup(job.job)
	if session != nil {
		dbSession := &database.Session{
			ID:             session.ID,
			JobID:          session.JobID,
			JobName:        session.JobName,
			StartTime:      session.StartTime,
			EndTime:        session.EndTime,
			Status:         session.Status,
			FilesProcessed: session.FilesProcessed,
			BytesWritten:   session.BytesWritten,
			Error:          session.Error,
		}
		if err := s.db.CreateSession(dbSession); err != nil {
			fmt.Printf("Warning: failed to persist session %s: %v\n", session.ID, err)
		}
	}

	// Update job in database
	now := time.Now()
	if err == nil {
		// Success
		s.db.UpdateJobLastRun(job.JobID, now, s.calculateNextRun(job))
		fmt.Printf("✅ Завдання %s успішно завершено\n", job.JobName)
	} else {
		// Failed
		fmt.Printf("❌ Завдання %s не виконано: %v\n", job.JobName, err)
	}

	// Update in-memory job
	job.LastRun = &now
	job.NextRun = s.calculateNextRun(job)
}

// calculateNextRun calculates the next run time for a job
func (s *Scheduler) calculateNextRun(job *ScheduledJob) time.Time {
	now := time.Now()

	switch job.ScheduleType {
	case ScheduleHourly:
		return now.Add(1 * time.Hour)

	case ScheduleDaily:
		// Parse time (HH:MM format)
		hour, minute := s.parseTime(job.ScheduleTime)
		next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())

		// If already passed today, schedule for tomorrow
		if now.After(next) {
			next = next.Add(24 * time.Hour)
		}

		return next

	case ScheduleWeekly:
		// Get next scheduled day
		nextDay := s.getNextWeekday(job.ScheduleDays)
		hour, minute := s.parseTime(job.ScheduleTime)

		next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())

		// Add days until we reach the scheduled day
		for int(next.Weekday()) != nextDay {
			next = next.Add(24 * time.Hour)
		}

		// If already passed today, schedule for next week
		if now.After(next) {
			next = next.Add(7 * 24 * time.Hour)
		}

		return next

	case ScheduleMonthly:
		// Schedule for first day of next month at scheduled time
		hour, minute := s.parseTime(job.ScheduleTime)
		next := time.Date(now.Year(), now.Month()+1, 1, hour, minute, 0, 0, now.Location())
		return next

	case ScheduleCron:
		// Parse cron expression and calculate next run
		return s.parseCron(job.CronExpression, now)

	default:
		// Manual schedule - no automatic run
		return time.Time{}
	}
}

// parseTime parses HH:MM format
func (s *Scheduler) parseTime(timeStr string) (hour, minute int) {
	parts := strings.Split(timeStr, ":")
	if len(parts) >= 2 {
		fmt.Sscanf(parts[0], "%d", &hour)
		fmt.Sscanf(parts[1], "%d", &minute)
	}
	return
}

// getNextWeekday returns the next weekday to run
func (s *Scheduler) getNextWeekday(days []string) int {
	if len(days) == 0 {
		return int(time.Monday) // Default to Monday
	}

	dayMap := map[string]int{
		"monday":    int(time.Monday),
		"tuesday":   int(time.Tuesday),
		"wednesday": int(time.Wednesday),
		"thursday":  int(time.Thursday),
		"friday":    int(time.Friday),
		"saturday":  int(time.Saturday),
		"sunday":    int(time.Sunday),
	}

	// Return first valid day
	for _, day := range days {
		if weekday, exists := dayMap[strings.ToLower(day)]; exists {
			return weekday
		}
	}

	return int(time.Monday)
}

// parseCron parses cron expression and returns next run time
// Supports: minute hour day month weekday
func (s *Scheduler) parseCron(expression string, now time.Time) time.Time {
	// Simple cron parser (production should use robust library like robfig/cron)
	parts := strings.Fields(expression)
	if len(parts) < 5 {
		return now.Add(24 * time.Hour) // Default to daily
	}

	minuteStr := parts[0]
	hourStr := parts[1]
	dayStr := parts[2]
	monthStr := parts[3]
	weekdayStr := parts[4]

	// Parse minute
	minute := 0
	if minuteStr != "*" {
		fmt.Sscanf(minuteStr, "%d", &minute)
		if minute < 0 || minute > 59 {
			minute = 0
		}
	}

	// Parse hour
	hour := 0
	if hourStr != "*" {
		fmt.Sscanf(hourStr, "%d", &hour)
		if hour < 0 || hour > 23 {
			hour = 0
		}
	}

	// Start with next minute
	next := now.Add(1 * time.Minute)
	next = time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), next.Minute(), 0, 0, next.Location())

	// Iterate to find next matching time (max 1 year)
	for i := 0; i < 525600; i++ { // 365 * 24 * 60 minutes
		// Check month
		if monthStr != "*" {
			month := int(next.Month())
			if fmt.Sprintf("%d", month) != monthStr {
				next = next.Add(1 * time.Minute)
				continue
			}
		}

		// Check day of month
		if dayStr != "*" {
			day := next.Day()
			if fmt.Sprintf("%d", day) != dayStr {
				next = next.Add(1 * time.Minute)
				continue
			}
		}

		// Check weekday
		if weekdayStr != "*" {
			weekday := int(next.Weekday())
			if fmt.Sprintf("%d", weekday) != weekdayStr {
				next = next.Add(1 * time.Minute)
				continue
			}
		}

		// Check hour
		if hourStr != "*" {
			h := next.Hour()
			if fmt.Sprintf("%d", h) != hourStr {
				next = next.Add(1 * time.Minute)
				continue
			}
		}

		// Check minute
		if minuteStr != "*" {
			m := next.Minute()
			if fmt.Sprintf("%d", m) != minuteStr {
				next = next.Add(1 * time.Minute)
				continue
			}
		}

		// All conditions matched
		return next
	}

	// Fallback to daily
	return now.Add(24 * time.Hour)
}

// loadJobs loads jobs from database
func (s *Scheduler) loadJobs() error {
	jobs, err := s.db.ListJobs()
	if err != nil {
		return err
	}

	for _, job := range jobs {
		if !job.Enabled {
			continue
		}

		scheduledJob := &ScheduledJob{
			JobID:        job.ID,
			JobName:      job.Name,
			ScheduleType: job.Schedule,
			Enabled:      job.Enabled,
			job: &backup.BackupJob{
				ID:          job.ID,
				Name:        job.Name,
				Type:        job.Type,
				Sources:     job.Sources,
				Destination: job.Destination,
				Compression: job.Compression,
				Encryption:  job.Encryption,
			},
		}

		// Calculate next run time
		if job.NextRun != nil {
			scheduledJob.NextRun = *job.NextRun
		} else {
			scheduledJob.NextRun = s.calculateNextRun(scheduledJob)
		}

		if job.LastRun != nil {
			scheduledJob.LastRun = job.LastRun
		}

		s.jobs[job.ID] = scheduledJob
	}

	return nil
}

// AddJob adds a job to the scheduler
func (s *Scheduler) AddJob(job *backup.BackupJob, scheduleType, scheduleTime string, scheduleDays []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	scheduledJob := &ScheduledJob{
		JobID:        job.ID,
		JobName:      job.Name,
		ScheduleType: scheduleType,
		ScheduleTime: scheduleTime,
		ScheduleDays: scheduleDays,
		Enabled:      true,
		job:          job,
	}

	scheduledJob.NextRun = s.calculateNextRun(scheduledJob)
	s.jobs[job.ID] = scheduledJob

	return nil
}

// RemoveJob removes a job from the scheduler
func (s *Scheduler) RemoveJob(jobID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.jobs, jobID)
	return nil
}

// EnableJob enables a scheduled job
func (s *Scheduler) EnableJob(jobID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if job, exists := s.jobs[jobID]; exists {
		job.Enabled = true
		job.NextRun = s.calculateNextRun(job)
	}
	return nil
}

// DisableJob disables a scheduled job
func (s *Scheduler) DisableJob(jobID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if job, exists := s.jobs[jobID]; exists {
		job.Enabled = false
	}
	return nil
}

// GetJob returns a scheduled job
func (s *Scheduler) GetJob(jobID string) (*ScheduledJob, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, exists := s.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("завдання не знайдено")
	}
	return job, nil
}

// ListJobs returns all scheduled jobs
func (s *Scheduler) ListJobs() []*ScheduledJob {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var jobs []*ScheduledJob
	for _, job := range s.jobs {
		jobs = append(jobs, job)
	}
	return jobs
}

// GetJobStatus returns the status of a scheduled job
func (s *Scheduler) GetJobStatus(jobID string) (string, error) {
	job, err := s.GetJob(jobID)
	if err != nil {
		return "", err
	}

	if !job.Enabled {
		return "Вимкнено", nil
	}

	now := time.Now()
	timeUntilRun := job.NextRun.Sub(now)

	if timeUntilRun <= 0 {
		return "Виконується", nil
	}

	if timeUntilRun < 5*time.Minute {
		return "Запуск скоро", nil
	}

	return fmt.Sprintf("Заплановано на %s", job.NextRun.Format("15:04 02.01.2006")), nil
}

// GetNextRuns returns next run times for all jobs
func (s *Scheduler) GetNextRuns() map[string]time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()

	nextRuns := make(map[string]time.Time)
	for id, job := range s.jobs {
		if job.Enabled {
			nextRuns[id] = job.NextRun
		}
	}
	return nextRuns
}

// FormatSchedule returns human-readable schedule
func FormatSchedule(scheduleType, scheduleTime string, scheduleDays []string) string {
	switch scheduleType {
	case ScheduleManual:
		return "Тільки вручну"
	case ScheduleHourly:
		return "Щогодини"
	case ScheduleDaily:
		return fmt.Sprintf("Щодня о %s", scheduleTime)
	case ScheduleWeekly:
		days := strings.Join(scheduleDays, ", ")
		return fmt.Sprintf("Щотижня (%s) о %s", days, scheduleTime)
	case ScheduleMonthly:
		return fmt.Sprintf("Щомісяця (1 число) о %s", scheduleTime)
	case ScheduleCron:
		return fmt.Sprintf("Cron: %s", scheduleTime)
	default:
		return scheduleType
	}
}

// FormatTimeUntil formats time until next run
func FormatTimeUntil(nextRun time.Time) string {
	if nextRun.IsZero() {
		return "—"
	}

	now := time.Now()
	duration := nextRun.Sub(now)

	if duration <= 0 {
		return "Зараз"
	}

	if duration < time.Minute {
		return "Менше хвилини"
	}

	if duration < time.Hour {
		minutes := int(duration.Minutes())
		return fmt.Sprintf("через %d хв", minutes)
	}

	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		minutes := int(duration.Minutes()) % 60
		return fmt.Sprintf("через %dг %dхв", hours, minutes)
	}

	days := int(duration.Hours() / 24)
	return fmt.Sprintf("через %d дн", days)
}
