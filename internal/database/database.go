package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type Database struct {
	db *sql.DB
}

type Job struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Type            string     `json:"type"`
	Sources         []string   `json:"sources"`
	Destination     string     `json:"destination"`
	Compression     bool       `json:"compression"`
	Encryption      bool       `json:"encryption"`
	Deduplication   bool       `json:"deduplication"` // Дедуплікація даних
	Incremental     bool       `json:"incremental"`   // Інкрементальне бекапування
	Schedule        string     `json:"schedule"`
	ScheduleTime    string     `json:"schedule_time"`
	ScheduleDays    []string   `json:"schedule_days"`
	Enabled         bool       `json:"enabled"`
	RetentionDays   int        `json:"retention_days"`   // Індивідуальна політика зберігання
	RetentionCopies int        `json:"retention_copies"` // Індивідуальна політика
	CreatedAt       time.Time  `json:"created_at"`
	LastRun         *time.Time `json:"last_run,omitempty"`
	NextRun         *time.Time `json:"next_run,omitempty"`
}

type Session struct {
	ID             string    `json:"id"`
	JobID          string    `json:"job_id"`
	JobName        string    `json:"job_name"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	Status         string    `json:"status"`
	FilesProcessed int       `json:"files_processed"`
	BytesWritten   int64     `json:"bytes_written"`
	Error          string    `json:"error,omitempty"`
}

type UserSession struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	LastUsed  time.Time `json:"last_used"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
}

func NewDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	d := &Database{db: db}
	if err := d.init(); err != nil {
		return nil, err
	}

	return d, nil
}

func (d *Database) init() error {
	schema := `
	CREATE TABLE IF NOT EXISTS jobs (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		type TEXT DEFAULT 'file',
		sources TEXT,
		destination TEXT,
		compression BOOLEAN DEFAULT false,
		encryption BOOLEAN DEFAULT false,
		deduplication BOOLEAN DEFAULT false,
		incremental BOOLEAN DEFAULT false,
		schedule TEXT DEFAULT 'manual',
		enabled BOOLEAN DEFAULT true,
		retention_days INTEGER DEFAULT 30,
		retention_copies INTEGER DEFAULT 10,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_run DATETIME,
		next_run DATETIME
	);

	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		job_id TEXT,
		job_name TEXT,
		start_time DATETIME,
		end_time DATETIME,
		status TEXT,
		files_processed INTEGER DEFAULT 0,
		bytes_written INTEGER DEFAULT 0,
		error TEXT,
		FOREIGN KEY (job_id) REFERENCES jobs(id)
	);

	CREATE TABLE IF NOT EXISTS user_sessions (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		token TEXT UNIQUE NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		expires_at DATETIME NOT NULL,
		last_used DATETIME DEFAULT CURRENT_TIMESTAMP,
		ip_address TEXT,
		user_agent TEXT,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT
	);
	`

	_, err := d.db.Exec(schema)
	return err
}

func (d *Database) CreateJob(job *Job) error {
	sources, _ := json.Marshal(job.Sources)

	_, err := d.db.Exec(`
		INSERT INTO jobs (id, name, type, sources, destination, compression, encryption, deduplication, incremental, schedule, enabled, retention_days, retention_copies)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, job.ID, job.Name, job.Type, string(sources), job.Destination, job.Compression, job.Encryption, job.Deduplication, job.Incremental, job.Schedule, job.Enabled, job.RetentionDays, job.RetentionCopies)

	return err
}

func (d *Database) ListJobs() ([]Job, error) {
	rows, err := d.db.Query(`
		SELECT id, name, type, sources, destination, compression, encryption, deduplication, incremental, schedule, enabled, retention_days, retention_copies, created_at, last_run, next_run
		FROM jobs ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []Job
	for rows.Next() {
		var job Job
		var sources string
		var lastRun, nextRun sql.NullTime

		err := rows.Scan(&job.ID, &job.Name, &job.Type, &sources, &job.Destination,
			&job.Compression, &job.Encryption, &job.Deduplication, &job.Incremental,
			&job.Schedule, &job.Enabled, &job.RetentionDays, &job.RetentionCopies,
			&job.CreatedAt, &lastRun, &nextRun)
		if err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(sources), &job.Sources)

		if lastRun.Valid {
			job.LastRun = &lastRun.Time
		}
		if nextRun.Valid {
			job.NextRun = &nextRun.Time
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (d *Database) GetJob(id string) (*Job, error) {
	var job Job
	var sources string
	var lastRun, nextRun sql.NullTime

	err := d.db.QueryRow(`
		SELECT id, name, type, sources, destination, compression, encryption, schedule, enabled, created_at, last_run, next_run
		FROM jobs WHERE id = ?
	`, id).Scan(&job.ID, &job.Name, &job.Type, &sources, &job.Destination,
		&job.Compression, &job.Encryption, &job.Schedule, &job.Enabled,
		&job.CreatedAt, &lastRun, &nextRun)

	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(sources), &job.Sources)

	if lastRun.Valid {
		job.LastRun = &lastRun.Time
	}
	if nextRun.Valid {
		job.NextRun = &nextRun.Time
	}

	return &job, nil
}

func (d *Database) UpdateJob(job *Job) error {
	sources, _ := json.Marshal(job.Sources)

	_, err := d.db.Exec(`
		UPDATE jobs SET name=?, type=?, sources=?, destination=?, compression=?, encryption=?, deduplication=?, incremental=?, schedule=?, enabled=?, retention_days=?, retention_copies=?
		WHERE id=?
	`, job.Name, job.Type, string(sources), job.Destination, job.Compression, job.Encryption, job.Deduplication, job.Incremental, job.Schedule, job.Enabled, job.RetentionDays, job.RetentionCopies, job.ID)

	return err
}

func (d *Database) DeleteJob(id string) error {
	// First delete associated sessions (foreign key constraint)
	_, err := d.db.Exec("DELETE FROM sessions WHERE job_id=?", id)
	if err != nil {
		return fmt.Errorf("failed to delete sessions: %w", err)
	}

	// Then delete the job
	_, err = d.db.Exec("DELETE FROM jobs WHERE id=?", id)
	if err != nil {
		return fmt.Errorf("failed to delete job: %w", err)
	}

	return nil
}

func (d *Database) CreateSession(session *Session) error {
	_, err := d.db.Exec(`
		INSERT INTO sessions (id, job_id, job_name, start_time, end_time, status, files_processed, bytes_written, error)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, session.ID, session.JobID, session.JobName, session.StartTime, session.EndTime, session.Status, session.FilesProcessed, session.BytesWritten, session.Error)

	return err
}

func (d *Database) ListSessions() ([]Session, error) {
	rows, err := d.db.Query(`
		SELECT id, job_id, job_name, start_time, end_time, status, files_processed, bytes_written, error
		FROM sessions ORDER BY start_time DESC LIMIT 100
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var s Session
		err := rows.Scan(&s.ID, &s.JobID, &s.JobName, &s.StartTime, &s.EndTime, &s.Status, &s.FilesProcessed, &s.BytesWritten, &s.Error)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}

	return sessions, nil
}

func (d *Database) UpdateJobLastRun(id string, lastRun time.Time, nextRun time.Time) error {
	_, err := d.db.Exec("UPDATE jobs SET last_run=?, next_run=? WHERE id=?", lastRun, nextRun, id)
	return err
}

// UserSession methods
func (d *Database) CreateUserSession(session *UserSession) error {
	_, err := d.db.Exec(`
		INSERT INTO user_sessions (id, user_id, token, created_at, expires_at, last_used, ip_address, user_agent)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, session.ID, session.UserID, session.Token, session.CreatedAt, session.ExpiresAt, session.LastUsed, session.IPAddress, session.UserAgent)
	return err
}

func (d *Database) GetUserSessionByToken(token string) (*UserSession, error) {
	var session UserSession
	err := d.db.QueryRow(`
		SELECT id, user_id, token, created_at, expires_at, last_used, ip_address, user_agent
		FROM user_sessions WHERE token = ?
	`, token).Scan(&session.ID, &session.UserID, &session.Token, &session.CreatedAt, &session.ExpiresAt, &session.LastUsed, &session.IPAddress, &session.UserAgent)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (d *Database) UpdateUserSessionLastUsed(token string, lastUsed time.Time) error {
	_, err := d.db.Exec("UPDATE user_sessions SET last_used=? WHERE token=?", lastUsed, token)
	return err
}

func (d *Database) DeleteUserSession(token string) error {
	_, err := d.db.Exec("DELETE FROM user_sessions WHERE token=?", token)
	return err
}

// MigrateCleanPaths removes Unicode characters from all existing job paths
func (d *Database) MigrateCleanPaths() error {
	// Add new columns if they don't exist (for existing databases)
	d.addMissingColumns()

	jobs, err := d.ListJobs()
	if err != nil {
		return err
	}

	for _, job := range jobs {
		needsUpdate := false

		// Clean source paths
		for i, src := range job.Sources {
			cleaned := cleanPath(src)
			if cleaned != src {
				job.Sources[i] = cleaned
				needsUpdate = true
			}
		}

		// Clean destination path
		cleanedDest := cleanPath(job.Destination)
		if cleanedDest != job.Destination {
			job.Destination = cleanedDest
			needsUpdate = true
		}

		if needsUpdate {
			if err := d.UpdateJob(&job); err != nil {
				log.Printf("Failed to update job %s during path migration: %v", job.ID, err)
			} else {
				log.Printf("Cleaned paths for job: %s", job.Name)
			}
		}
	}

	return nil
}

// cleanPath removes Unicode characters and normalizes slashes
func cleanPath(path string) string {
	if path == "" {
		return path
	}
	// Remove BOM and other invisible Unicode characters
	path = strings.TrimSpace(path)
	path = strings.ReplaceAll(path, "\uFEFF", "") // BOM
	path = strings.ReplaceAll(path, "\u200B", "") // Zero-width space
	path = strings.ReplaceAll(path, "\u200C", "") // Zero-width non-joiner
	path = strings.ReplaceAll(path, "\u200D", "") // Zero-width joiner
	path = strings.ReplaceAll(path, "\u2060", "") // Word joiner
	// Replace forward slashes with backslashes for Windows
	path = filepath.FromSlash(path)
	return path
}

// addMissingColumns adds new columns to existing database schema
func (d *Database) addMissingColumns() error {
	columns := []struct {
		name         string
		defaultValue string
	}{
		{"deduplication", "false"},
		{"incremental", "false"},
		{"retention_days", "30"},
		{"retention_copies", "10"},
	}

	for _, col := range columns {
		// Check if column exists
		var count int
		err := d.db.QueryRow(`
			SELECT COUNT(*) FROM pragma_table_info('jobs') WHERE name=?
		`, col.name).Scan(&count)

		if err == nil && count == 0 {
			// Column doesn't exist, add it
			_, err = d.db.Exec(fmt.Sprintf(
				"ALTER TABLE jobs ADD COLUMN %s DEFAULT %s",
				col.name, col.defaultValue,
			))
			if err != nil {
				log.Printf("Warning: Failed to add column %s: %v", col.name, err)
			} else {
				log.Printf("✓ Added column '%s' to jobs table", col.name)
			}
		}
	}

	return nil
}

func (d *Database) Close() error {
	return d.db.Close()
}
