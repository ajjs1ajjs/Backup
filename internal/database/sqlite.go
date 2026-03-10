package database

import (
	"database/sql"
	"fmt"
	"time"

	"novabackup/pkg/models"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

// DatabaseType defines the type of database
type DatabaseType string

const (
	DatabaseSQLite   DatabaseType = "sqlite"
	DatabasePostgres DatabaseType = "postgres"
)

// Config holds database configuration
type Config struct {
	Type     DatabaseType
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
	DBPath   string // For SQLite
}

// DefaultSQLiteConfig returns default SQLite configuration
func DefaultSQLiteConfig(dbPath string) *Config {
	return &Config{
		Type:   DatabaseSQLite,
		DBPath: dbPath,
	}
}

// DefaultPostgresConfig returns default PostgreSQL configuration
func DefaultPostgresConfig(host, user, password, database string) *Config {
	return &Config{
		Type:     DatabasePostgres,
		Host:     host,
		Port:     5432,
		User:     user,
		Password: password,
		Database: database,
		SSLMode:  "disable",
	}
}

// Connection represents a database connection
type Connection struct {
	db     *sql.DB
	config *Config
}

// NewConnection creates a new database connection based on config
func NewConnection(config *Config) (*Connection, error) {
	switch config.Type {
	case DatabaseSQLite:
		return NewSQLiteConnection(config.DBPath)
	case DatabasePostgres:
		return NewPostgresConnection(config)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}
}

func initSchema(db *sql.DB) error {
	queries := []string{
		// Jobs table
		`CREATE TABLE IF NOT EXISTS backup_jobs (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			job_type TEXT NOT NULL,
			source TEXT NOT NULL,
			destination TEXT NOT NULL,
			schedule TEXT,
			enabled BOOLEAN DEFAULT 1,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		// Chunks table (deduplication index)
		`CREATE TABLE IF NOT EXISTS chunks (
			hash TEXT PRIMARY KEY,
			size_bytes INTEGER NOT NULL,
			compressed_size INTEGER,
			storage_path TEXT NOT NULL,
			ref_count INTEGER DEFAULT 1,
			first_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			last_accessed TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		// Restore points table
		`CREATE TABLE IF NOT EXISTS restore_points (
			id TEXT PRIMARY KEY,
			job_id TEXT NOT NULL,
			point_time TIMESTAMP NOT NULL,
			status TEXT DEFAULT 'pending',
			total_bytes INTEGER DEFAULT 0,
			processed_bytes INTEGER DEFAULT 0,
			duration_seconds INTEGER DEFAULT 0,
			metadata TEXT,
			FOREIGN KEY (job_id) REFERENCES backup_jobs(id) ON DELETE CASCADE
		)`,

		// Restore point chunks mapping
		`CREATE TABLE IF NOT EXISTS restore_point_chunks (
			restore_point_id TEXT NOT NULL,
			chunk_hash TEXT NOT NULL,
			sequence_order INTEGER NOT NULL,
			original_path TEXT,
			PRIMARY KEY (restore_point_id, sequence_order),
			FOREIGN KEY (restore_point_id) REFERENCES restore_points(id) ON DELETE CASCADE,
			FOREIGN KEY (chunk_hash) REFERENCES chunks(hash)
		)`,

		// Backup results table
		`CREATE TABLE IF NOT EXISTS backup_results (
			id TEXT PRIMARY KEY,
			job_id TEXT NOT NULL,
			status TEXT NOT NULL,
			start_time TIMESTAMP NOT NULL,
			end_time TIMESTAMP,
			bytes_read INTEGER DEFAULT 0,
			bytes_written INTEGER DEFAULT 0,
			files_total INTEGER DEFAULT 0,
			files_success INTEGER DEFAULT 0,
			files_failed INTEGER DEFAULT 0,
			error_message TEXT,
			FOREIGN KEY (job_id) REFERENCES backup_jobs(id) ON DELETE CASCADE
		)`,

		// Indexes
		`CREATE INDEX IF NOT EXISTS idx_chunks_hash ON chunks(hash)`,
		`CREATE INDEX IF NOT EXISTS idx_restore_points_job ON restore_points(job_id, point_time DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_rpc_lookup ON restore_point_chunks(chunk_hash)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %w\nQuery: %s", err, query)
		}
	}

	return nil
}

func (c *Connection) Close() error {
	return c.db.Close()
}

func (c *Connection) GetDB() *sql.DB {
	return c.db
}

// Job-related methods
func (c *Connection) CreateJob(job *models.Job) error {
	query := `INSERT INTO backup_jobs (id, name, description, job_type, source, destination, schedule, enabled)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := c.db.Exec(query, job.ID.String(), job.Name, job.Description, job.JobType, job.Source, job.Destination, job.Schedule, job.Enabled)
	return err
}

func (c *Connection) GetJobByID(id uuid.UUID) (*models.Job, error) {
	job := &models.Job{}
	query := `SELECT id, name, description, job_type, source, destination, schedule, enabled, created_at, updated_at
	          FROM backup_jobs WHERE id = ?`
	row := c.db.QueryRow(query, id.String())
	err := row.Scan(&job.ID, &job.Name, &job.Description, &job.JobType, &job.Source, &job.Destination, &job.Schedule, &job.Enabled, &job.CreatedAt, &job.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return job, nil
}

func (c *Connection) GetAllJobs() ([]models.Job, error) {
	query := `SELECT id, name, description, job_type, source, destination, schedule, enabled, created_at, updated_at
	          FROM backup_jobs ORDER BY created_at DESC`
	rows, err := c.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []models.Job
	for rows.Next() {
		var job models.Job
		if err := rows.Scan(&job.ID, &job.Name, &job.Description, &job.JobType, &job.Source, &job.Destination, &job.Schedule, &job.Enabled, &job.CreatedAt, &job.UpdatedAt); err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

// Chunk-related methods
func (c *Connection) GetChunkByHash(hash string) (string, error) {
	query := `SELECT storage_path FROM chunks WHERE hash = ?`
	var path string
	err := c.db.QueryRow(query, hash).Scan(&path)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return path, err
}

func (c *Connection) AddChunk(hash string, sizeBytes int, compressedSize int, storagePath string) error {
	query := `INSERT OR IGNORE INTO chunks (hash, size_bytes, compressed_size, storage_path) VALUES (?, ?, ?, ?)`
	_, err := c.db.Exec(query, hash, sizeBytes, compressedSize, storagePath)
	return err
}

// Restore point methods
func (c *Connection) CreateRestorePoint(rp *models.RestorePoint) error {
	query := `INSERT INTO restore_points (id, job_id, point_time, status, total_bytes, metadata)
	          VALUES (?, ?, ?, ?, ?, ?)`
	_, err := c.db.Exec(query, rp.ID.String(), rp.JobID.String(), rp.PointTime, rp.Status, rp.TotalBytes, rp.Metadata)
	return err
}

// GetRestorePointsByJob returns all restore points for a job
func (c *Connection) GetRestorePointsByJob(jobID uuid.UUID) ([]models.RestorePoint, error) {
	query := `SELECT id, job_id, point_time, status, total_bytes, processed_bytes, duration_seconds, metadata
	          FROM restore_points WHERE job_id = ? ORDER BY point_time DESC`
	rows, err := c.db.Query(query, jobID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []models.RestorePoint
	for rows.Next() {
		var rp models.RestorePoint
		var metadata sql.NullString
		if err := rows.Scan(&rp.ID, &rp.JobID, &rp.PointTime, &rp.Status, &rp.TotalBytes, &rp.ProcessedBytes, &rp.DurationSeconds, &metadata); err != nil {
			return nil, err
		}
		if metadata.Valid {
			rp.Metadata = metadata.String
		}
		points = append(points, rp)
	}
	return points, nil
}

// GetRestorePointByID returns a specific restore point
func (c *Connection) GetRestorePointByID(id uuid.UUID) (*models.RestorePoint, error) {
	query := `SELECT id, job_id, point_time, status, total_bytes, processed_bytes, duration_seconds, metadata
	          FROM restore_points WHERE id = ?`
	row := c.db.QueryRow(query, id.String())

	var rp models.RestorePoint
	var metadata sql.NullString
	if err := row.Scan(&rp.ID, &rp.JobID, &rp.PointTime, &rp.Status, &rp.TotalBytes, &rp.ProcessedBytes, &rp.DurationSeconds, &metadata); err != nil {
		return nil, err
	}
	if metadata.Valid {
		rp.Metadata = metadata.String
	}
	return &rp, nil
}

// GetChunksForRestorePoint returns all chunks for a restore point
func (c *Connection) GetChunksForRestorePoint(rpID uuid.UUID) ([]models.ChunkInfo, error) {
	query := `SELECT c.hash, c.size_bytes, c.compressed_size, c.storage_path
	          FROM chunks c
	          INNER JOIN restore_point_chunks rpc ON c.hash = rpc.chunk_hash
	          WHERE rpc.restore_point_id = ?
	          ORDER BY rpc.sequence_order`
	rows, err := c.db.Query(query, rpID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chunks []models.ChunkInfo
	for rows.Next() {
		var chunk models.ChunkInfo
		if err := rows.Scan(&chunk.Hash, &chunk.SizeBytes, &chunk.CompressedSize, &chunk.StoragePath); err != nil {
			return nil, err
		}
		chunks = append(chunks, chunk)
	}
	return chunks, nil
}

// DeleteRestorePoint deletes a restore point
func (c *Connection) DeleteRestorePoint(id uuid.UUID) error {
	query := `DELETE FROM restore_points WHERE id = ?`
	_, err := c.db.Exec(query, id.String())
	return err
}

// DecrementChunkRef decrements chunk reference count
func (c *Connection) DecrementChunkRef(hash string) error {
	query := `UPDATE chunks SET ref_count = ref_count - 1 WHERE hash = ?`
	_, err := c.db.Exec(query, hash)
	return err
}

// GetOrphanedChunks returns chunks with ref_count = 0
func (c *Connection) GetOrphanedChunks() ([]string, error) {
	query := `SELECT hash FROM chunks WHERE ref_count <= 0`
	rows, err := c.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hashes []string
	for rows.Next() {
		var hash string
		if err := rows.Scan(&hash); err != nil {
			return nil, err
		}
		hashes = append(hashes, hash)
	}
	return hashes, nil
}

// DeleteChunk deletes a chunk from database
func (c *Connection) DeleteChunk(hash string) error {
	query := `DELETE FROM chunks WHERE hash = ?`
	_, err := c.db.Exec(query, hash)
	return err
}

// NewSQLiteConnection creates a new SQLite connection
func NewSQLiteConnection(dbPath string) (*Connection, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Initialize schema
	if err := initSchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return &Connection{db: db, config: DefaultSQLiteConfig(dbPath)}, nil
}

// NewPostgresConnection creates a new PostgreSQL connection
func NewPostgresConnection(config *Config) (*Connection, error) {
	if config.Host == "" || config.Database == "" || config.User == "" {
		return nil, fmt.Errorf("host, database, and user are required for PostgreSQL")
	}

	// Build connection string
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.Database, config.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping PostgreSQL database: %w", err)
	}

	// Initialize schema
	if err := initPostgresSchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return &Connection{db: db, config: config}, nil
}

// initPostgresSchema initializes PostgreSQL database schema
func initPostgresSchema(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS backup_jobs (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			job_type TEXT NOT NULL,
			source TEXT NOT NULL,
			destination TEXT NOT NULL,
			schedule TEXT,
			enabled BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS chunks (
			hash TEXT PRIMARY KEY,
			size_bytes BIGINT NOT NULL,
			compressed_size BIGINT,
			storage_path TEXT NOT NULL,
			ref_count INTEGER DEFAULT 1,
			first_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			last_accessed TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS restore_points (
			id TEXT PRIMARY KEY,
			job_id TEXT NOT NULL REFERENCES backup_jobs(id) ON DELETE CASCADE,
			point_time TIMESTAMP NOT NULL,
			status TEXT DEFAULT 'pending',
			total_bytes BIGINT DEFAULT 0,
			processed_bytes BIGINT DEFAULT 0,
			duration_seconds INTEGER DEFAULT 0,
			metadata JSONB
		)`,
		`CREATE TABLE IF NOT EXISTS restore_point_chunks (
			restore_point_id TEXT NOT NULL REFERENCES restore_points(id) ON DELETE CASCADE,
			chunk_hash TEXT NOT NULL REFERENCES chunks(hash),
			sequence_order INTEGER NOT NULL,
			original_path TEXT,
			PRIMARY KEY (restore_point_id, sequence_order)
		)`,
		`CREATE TABLE IF NOT EXISTS backup_results (
			id TEXT PRIMARY KEY,
			job_id TEXT NOT NULL REFERENCES backup_jobs(id) ON DELETE CASCADE,
			status TEXT NOT NULL,
			start_time TIMESTAMP NOT NULL,
			end_time TIMESTAMP,
			bytes_read BIGINT DEFAULT 0,
			bytes_written BIGINT DEFAULT 0,
			files_total INTEGER DEFAULT 0,
			files_success INTEGER DEFAULT 0,
			files_failed INTEGER DEFAULT 0,
			error_message TEXT
		)`,
		// Create indexes for better performance
		`CREATE INDEX IF NOT EXISTS idx_chunks_ref_count ON chunks(ref_count)`,
		`CREATE INDEX IF NOT EXISTS idx_restore_points_job_id ON restore_points(job_id)`,
		`CREATE INDEX IF NOT EXISTS idx_backup_results_job_id ON backup_results(job_id)`,
		`CREATE INDEX IF NOT EXISTS idx_backup_results_status ON backup_results(status)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	return nil
}
