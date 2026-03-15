package database

import (
	"database/sql"
	"encoding/json"
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
		schedule TEXT DEFAULT 'manual',
		enabled BOOLEAN DEFAULT true,
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
		INSERT INTO jobs (id, name, type, sources, destination, compression, encryption, schedule, enabled)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, job.ID, job.Name, job.Type, string(sources), job.Destination, job.Compression, job.Encryption, job.Schedule, job.Enabled)

	return err
}

func (d *Database) ListJobs() ([]Job, error) {
	rows, err := d.db.Query(`
		SELECT id, name, type, sources, destination, compression, encryption, schedule, enabled, created_at, last_run, next_run
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
		UPDATE jobs SET name=?, type=?, sources=?, destination=?, compression=?, encryption=?, schedule=?, enabled=?
		WHERE id=?
	`, job.Name, job.Type, string(sources), job.Destination, job.Compression, job.Encryption, job.Schedule, job.Enabled, job.ID)

	return err
}

func (d *Database) DeleteJob(id string) error {
	_, err := d.db.Exec("DELETE FROM jobs WHERE id=?", id)
	return err
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

func (d *Database) Close() error {
	return d.db.Close()
}
