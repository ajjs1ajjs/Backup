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
	ID                string     `json:"id"`
	Name              string     `json:"name"`
	Type              string     `json:"type"`
	Sources           []string   `json:"sources"`
	Destination       string     `json:"destination"`
	Compression       bool       `json:"compression"`
	CompressionLevel  int        `json:"compression_level"` // 0-9
	Encryption        bool       `json:"encryption"`
	Deduplication     bool       `json:"deduplication"`
	BlockSize         int        `json:"block_size"`  // Block size for dedup
	MaxThreads        int        `json:"max_threads"` // Parallel threads
	Incremental       bool       `json:"incremental"`
	FullBackupEvery   int        `json:"full_backup_every"` // Days between full backups
	ExcludePatterns   []string   `json:"exclude_patterns"`
	IncludePatterns   []string   `json:"include_patterns"`
	PreBackupScript   string     `json:"pre_backup_script"`
	PostBackupScript  string     `json:"post_backup_script"`
	Schedule          string     `json:"schedule"`
	ScheduleTime      string     `json:"schedule_time"`
	ScheduleDays      []string   `json:"schedule_days"`
	CronExpression    string     `json:"cron_expression"`
	Enabled           bool       `json:"enabled"`
	RetentionDays     int        `json:"retention_days"`
	RetentionCopies   int        `json:"retention_copies"`
	GFSDaily          int        `json:"gfs_daily"`
	GFSWeekly         int        `json:"gfs_weekly"`
	GFSMonthly        int        `json:"gfs_monthly"`
	GFSQuarterly      int        `json:"gfs_quarterly"`
	GFSYearly         int        `json:"gfs_yearly"`
	BackupCopyEnabled bool       `json:"backup_copy_enabled"`
	BackupCopyDestID  string     `json:"backup_copy_dest_id"`
	BackupCopyDelay   int        `json:"backup_copy_delay"`
	BackupCopyEncrypt bool       `json:"backup_copy_encrypt"`
	CreatedAt         time.Time  `json:"created_at"`
	LastRun           *time.Time `json:"last_run,omitempty"`
	NextRun           *time.Time `json:"next_run,omitempty"`

	// Database specific fields
	DatabaseType string `json:"database_type,omitempty"`
	Server       string `json:"server,omitempty"`
	Port         int    `json:"port,omitempty"`
	AuthType     string `json:"auth_type,omitempty"`
	Login        string `json:"login,omitempty"`
	Password     string `json:"password,omitempty"`
	Service      string `json:"service,omitempty"`

	// VM specific fields
	VMNames    []string `json:"vm_names,omitempty"`
	HyperVHost string   `json:"hyperv_host,omitempty"`
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
		compression_level INTEGER DEFAULT 5,
		encryption BOOLEAN DEFAULT false,
		deduplication BOOLEAN DEFAULT false,
		block_size INTEGER DEFAULT 1048576,
		max_threads INTEGER DEFAULT 4,
		incremental BOOLEAN DEFAULT false,
		full_backup_every INTEGER DEFAULT 7,
		exclude_patterns TEXT,
		include_patterns TEXT,
		pre_backup_script TEXT,
		post_backup_script TEXT,
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
	excludePatterns, _ := json.Marshal(job.ExcludePatterns)
	includePatterns, _ := json.Marshal(job.IncludePatterns)
	databaseType, _ := json.Marshal(job.DatabaseType)
	server, _ := json.Marshal(job.Server)
	authType, _ := json.Marshal(job.AuthType)
	login, _ := json.Marshal(job.Login)
	password, _ := json.Marshal(job.Password)
	service, _ := json.Marshal(job.Service)
	vmNames, _ := json.Marshal(job.VMNames)

	_, err := d.db.Exec(`
		INSERT INTO jobs (id, name, type, sources, destination, compression, compression_level, encryption, deduplication, block_size, max_threads, incremental, full_backup_every, exclude_patterns, include_patterns, pre_backup_script, post_backup_script, schedule, enabled, retention_days, retention_copies, database_type, server, port, auth_type, login, password, service, vm_names, hyperv_host)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, job.ID, job.Name, job.Type, string(sources), job.Destination, job.Compression, job.CompressionLevel, job.Encryption, job.Deduplication, job.BlockSize, job.MaxThreads, job.Incremental, job.FullBackupEvery, string(excludePatterns), string(includePatterns), job.PreBackupScript, job.PostBackupScript, job.Schedule, job.Enabled, job.RetentionDays, job.RetentionCopies, string(databaseType), string(server), job.Port, string(authType), string(login), string(password), string(service), string(vmNames), job.HyperVHost)

	return err
}

func (d *Database) ListJobs() ([]Job, error) {
	rows, err := d.db.Query(`
		SELECT id, name, type, sources, destination, compression, encryption, deduplication, incremental, schedule, enabled, retention_days, retention_copies, created_at, last_run, next_run, database_type, server, port, auth_type, login, password, service, vm_names, hyperv_host
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
		var databaseType, server, authType, login, password, service, vmNames sql.NullString

		err := rows.Scan(&job.ID, &job.Name, &job.Type, &sources, &job.Destination,
			&job.Compression, &job.Encryption, &job.Deduplication, &job.Incremental,
			&job.Schedule, &job.Enabled, &job.RetentionDays, &job.RetentionCopies,
			&job.CreatedAt, &lastRun, &nextRun, &databaseType, &server, &job.Port, &authType, &login, &password, &service, &vmNames, &job.HyperVHost)
		if err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(sources), &job.Sources)

		if databaseType.Valid {
			json.Unmarshal([]byte(databaseType.String), &job.DatabaseType)
		}
		if server.Valid {
			json.Unmarshal([]byte(server.String), &job.Server)
		}
		if authType.Valid {
			json.Unmarshal([]byte(authType.String), &job.AuthType)
		}
		if login.Valid {
			json.Unmarshal([]byte(login.String), &job.Login)
		}
		if password.Valid {
			json.Unmarshal([]byte(password.String), &job.Password)
		}
		if service.Valid {
			json.Unmarshal([]byte(service.String), &job.Service)
		}
		if vmNames.Valid {
			json.Unmarshal([]byte(vmNames.String), &job.VMNames)
		}

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
	var databaseType, server, authType, login, password, service, vmNames sql.NullString

	err := d.db.QueryRow(`
		SELECT id, name, type, sources, destination, compression, encryption, schedule, enabled, created_at, last_run, next_run, database_type, server, port, auth_type, login, password, service, vm_names, hyperv_host
		FROM jobs WHERE id = ?
	`, id).Scan(&job.ID, &job.Name, &job.Type, &sources, &job.Destination,
		&job.Compression, &job.Encryption, &job.Schedule, &job.Enabled,
		&job.CreatedAt, &lastRun, &nextRun, &databaseType, &server, &job.Port, &authType, &login, &password, &service, &vmNames, &job.HyperVHost)

	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(sources), &job.Sources)

	if databaseType.Valid {
		json.Unmarshal([]byte(databaseType.String), &job.DatabaseType)
	}
	if server.Valid {
		json.Unmarshal([]byte(server.String), &job.Server)
	}
	if authType.Valid {
		json.Unmarshal([]byte(authType.String), &job.AuthType)
	}
	if login.Valid {
		json.Unmarshal([]byte(login.String), &job.Login)
	}
	if password.Valid {
		json.Unmarshal([]byte(password.String), &job.Password)
	}
	if service.Valid {
		json.Unmarshal([]byte(service.String), &job.Service)
	}
	if vmNames.Valid {
		json.Unmarshal([]byte(vmNames.String), &job.VMNames)
	}

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
	databaseType, _ := json.Marshal(job.DatabaseType)
	server, _ := json.Marshal(job.Server)
	authType, _ := json.Marshal(job.AuthType)
	login, _ := json.Marshal(job.Login)
	password, _ := json.Marshal(job.Password)
	service, _ := json.Marshal(job.Service)
	vmNames, _ := json.Marshal(job.VMNames)

	_, err := d.db.Exec(`
		UPDATE jobs SET name=?, type=?, sources=?, destination=?, compression=?, encryption=?, deduplication=?, incremental=?, schedule=?, enabled=?, retention_days=?, retention_copies=?, database_type=?, server=?, port=?, auth_type=?, login=?, password=?, service=?, vm_names=?, hyperv_host=?
		WHERE id=?
	`, job.Name, job.Type, string(sources), job.Destination, job.Compression, job.Encryption, job.Deduplication, job.Incremental, job.Schedule, job.Enabled, job.RetentionDays, job.RetentionCopies, string(databaseType), string(server), job.Port, string(authType), string(login), string(password), string(service), string(vmNames), job.HyperVHost, job.ID)

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
		{"compression_level", "5"},
		{"block_size", "1048576"},
		{"max_threads", "4"},
		{"full_backup_every", "7"},
		{"exclude_patterns", "''"},
		{"include_patterns", "''"},
		{"pre_backup_script", "''"},
		{"post_backup_script", "''"},
		{"cron_expression", "''"},
		{"gfs_daily", "7"},
		{"gfs_weekly", "4"},
		{"gfs_monthly", "12"},
		{"gfs_quarterly", "40"},
		{"gfs_yearly", "7"},
		{"backup_copy_enabled", "false"},
		{"backup_copy_dest_id", "''"},
		{"backup_copy_delay", "0"},
		{"backup_copy_encrypt", "false"},
		// Database specific columns
		{"database_type", "''"},
		{"server", "''"},
		{"port", "0"},
		{"auth_type", "''"},
		{"login", "''"},
		{"password", "''"},
		{"service", "''"},
		// VM specific columns
		{"vm_names", "''"},
		{"hyperv_host", "''"},
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
