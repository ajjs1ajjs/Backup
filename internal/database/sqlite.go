package database

import (
	"database/sql"
	"fmt"
	"time"

	"novabackup/pkg/models"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
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
			retention_days INTEGER DEFAULT 30,
			enable_guest_processing BOOLEAN DEFAULT 0,
			guest_credentials_id TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		// Repositories table
		`CREATE TABLE IF NOT EXISTS backup_repositories (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			repo_type TEXT NOT NULL,
			path TEXT,
			endpoint TEXT,
			bucket TEXT,
			region TEXT,
			access_key TEXT,
			secret_key TEXT,
			is_sobr BOOLEAN DEFAULT 0,
			parent_sobr TEXT,
			tier TEXT DEFAULT 'performance',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		// Chunks table (updated with tier info)
		`CREATE TABLE IF NOT EXISTS chunks (
			hash TEXT PRIMARY KEY,
			size_bytes INTEGER NOT NULL,
			compressed_size INTEGER,
			storage_path TEXT NOT NULL,
			repo_id TEXT,
			tier TEXT DEFAULT 'performance',
			ref_count INTEGER DEFAULT 1,
			first_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			last_accessed TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (repo_id) REFERENCES backup_repositories(id)
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
		// Infrastructure nodes table
		`CREATE TABLE IF NOT EXISTS infrastructure_nodes (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			ip_address TEXT NOT NULL,
			node_type TEXT NOT NULL,
			username TEXT,
			password_encrypted TEXT,
			status TEXT DEFAULT 'online',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		// Infrastructure objects (VMs, Disks, etc.)
		`CREATE TABLE IF NOT EXISTS infrastructure_objects (
			id TEXT PRIMARY KEY,
			node_id TEXT NOT NULL,
			name TEXT NOT NULL,
			obj_type TEXT NOT NULL,
			external_id TEXT,
			metadata TEXT,
			status TEXT DEFAULT 'discovered',
			last_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (node_id) REFERENCES infrastructure_nodes(id) ON DELETE CASCADE
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
	query := `INSERT INTO backup_jobs (id, name, description, job_type, source, destination, schedule, enabled, retention_days, enable_guest_processing, guest_credentials_id)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := c.db.Exec(query, job.ID.String(), job.Name, job.Description, job.JobType, job.Source, job.Destination, job.Schedule, job.Enabled, job.RetentionDays, job.EnableGuestProcessing, job.GuestCredentialsID)
	return err
}

func (c *Connection) GetJobByID(id uuid.UUID) (*models.Job, error) {
	job := &models.Job{}
	query := `SELECT id, name, description, job_type, source, destination, schedule, enabled, retention_days, enable_guest_processing, guest_credentials_id, created_at, updated_at
	          FROM backup_jobs WHERE id = ?`
	row := c.db.QueryRow(query, id.String())
	err := row.Scan(&job.ID, &job.Name, &job.Description, &job.JobType, &job.Source, &job.Destination, &job.Schedule, &job.Enabled, &job.RetentionDays, &job.EnableGuestProcessing, &job.GuestCredentialsID, &job.CreatedAt, &job.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return job, nil
}

// CreateInfrastructureObject saves a discovered object to DB
func (c *Connection) CreateInfrastructureObject(obj_id uuid.UUID, node_id uuid.UUID, name string, obj_type string, external_id string, metadata string) error {
	query := `INSERT INTO infrastructure_objects (id, node_id, name, obj_type, external_id, metadata)
	          VALUES (?, ?, ?, ?, ?, ?)`
	_, err := c.db.Exec(query, obj_id.String(), node_id.String(), name, obj_type, external_id, metadata)
	return err
}

// GetObjectsByNode returns all objects for a specific host
func (c *Connection) GetObjectsByNode(nodeID uuid.UUID) ([]ginH, error) {
	query := `SELECT id, name, obj_type, external_id, metadata FROM infrastructure_objects WHERE node_id = ?`
	rows, err := c.db.Query(query, nodeID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ginH
	for rows.Next() {
		var id, name, obj_type, ext_id, meta string
		if err := rows.Scan(&id, &name, &obj_type, &ext_id, &meta); err != nil {
			return nil, err
		}
		results = append(results, ginH{
			"id":          id,
			"name":        name,
			"obj_type":    obj_type,
			"external_id": ext_id,
			"metadata":    meta,
		})
	}
	return results, nil
}

type ginH map[string]interface{}

func (c *Connection) GetAllJobs() ([]models.Job, error) {
	query := `SELECT id, name, description, job_type, source, destination, schedule, enabled, retention_days, enable_guest_processing, guest_credentials_id, created_at, updated_at
	          FROM backup_jobs ORDER BY created_at DESC`
	rows, err := c.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []models.Job
	for rows.Next() {
		var job models.Job
		if err := rows.Scan(&job.ID, &job.Name, &job.Description, &job.JobType, &job.Source, &job.Destination, &job.Schedule, &job.Enabled, &job.RetentionDays, &job.EnableGuestProcessing, &job.GuestCredentialsID, &job.CreatedAt, &job.UpdatedAt); err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

// Repository-related methods
func (c *Connection) CreateRepository(repo *models.Repository) error {
	query := `INSERT INTO backup_repositories (id, name, repo_type, path, endpoint, bucket, region, access_key, secret_key, is_sobr, parent_sobr, tier)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := c.db.Exec(query, repo.ID.String(), repo.Name, repo.Type, repo.Path, repo.Endpoint, repo.Bucket, repo.Region, repo.AccessKey, repo.SecretKey, repo.IsSOBR, repo.ParentSOBR.String(), repo.Tier)
	return err
}

func (c *Connection) GetRepositoryByID(id uuid.UUID) (*models.Repository, error) {
	repo := &models.Repository{}
	query := `SELECT id, name, repo_type, path, endpoint, bucket, region, access_key, secret_key, is_sobr, parent_sobr, tier, created_at, updated_at
	          FROM backup_repositories WHERE id = ?`
	row := c.db.QueryRow(query, id.String())
	var parentSOBR string
	err := row.Scan(&repo.ID, &repo.Name, &repo.Type, &repo.Path, &repo.Endpoint, &repo.Bucket, &repo.Region, &repo.AccessKey, &repo.SecretKey, &repo.IsSOBR, &parentSOBR, &repo.Tier, &repo.CreatedAt, &repo.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if parentSOBR != "" {
		repo.ParentSOBR, _ = uuid.Parse(parentSOBR)
	}
	return repo, nil
}

func (c *Connection) GetAllRepositories() ([]models.Repository, error) {
	query := `SELECT id, name, repo_type, path, endpoint, bucket, region, access_key, secret_key, is_sobr, parent_sobr, tier, created_at, updated_at
	          FROM backup_repositories ORDER BY name ASC`
	rows, err := c.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var repos []models.Repository
	for rows.Next() {
		var repo models.Repository
		var parentSOBR string
		err := rows.Scan(&repo.ID, &repo.Name, &repo.Type, &repo.Path, &repo.Endpoint, &repo.Bucket, &repo.Region, &repo.AccessKey, &repo.SecretKey, &repo.IsSOBR, &parentSOBR, &repo.Tier, &repo.CreatedAt, &repo.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if parentSOBR != "" {
			repo.ParentSOBR, _ = uuid.Parse(parentSOBR)
		}
		repos = append(repos, repo)
	}
	return repos, nil
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

func (c *Connection) AddChunk(hash string, size int64, compressedSize int, storagePath string, repoID uuid.UUID, tier string) error {
	query := `INSERT OR IGNORE INTO chunks (hash, size_bytes, compressed_size, storage_path, repo_id, tier) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := c.db.Exec(query, hash, size, compressedSize, storagePath, repoID.String(), tier)
	return err
}

// Restore point methods
func (c *Connection) CreateRestorePoint(rp *models.RestorePoint) error {
	query := `INSERT INTO restore_points (id, job_id, point_time, status, total_bytes, metadata)
	          VALUES (?, ?, ?, ?, ?, ?)`
	_, err := c.db.Exec(query, rp.ID.String(), rp.JobID.String(), rp.PointTime, rp.Status, rp.TotalBytes, rp.Metadata)
	return err
}

// CreateBackupResult saves a new backup result record
func (c *Connection) CreateBackupResult(res *models.BackupResult) error {
	query := `INSERT INTO backup_results (id, job_id, status, start_time, bytes_read, bytes_written, files_total, files_success, files_failed)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := c.db.Exec(query, res.ID.String(), res.JobID.String(), res.Status, res.StartTime, res.BytesRead, res.BytesWritten, res.FilesTotal, res.FilesSuccess, res.FilesFailed)
	return err
}

// UpdateBackupResult updates an existing backup result
func (c *Connection) UpdateBackupResult(res *models.BackupResult) error {
	query := `UPDATE backup_results SET status = ?, end_time = ?, bytes_read = ?, bytes_written = ?, files_total = ?, files_success = ?, files_failed = ?, error_message = ?
	          WHERE id = ?`
	_, err := c.db.Exec(query, res.Status, res.EndTime, res.BytesRead, res.BytesWritten, res.FilesTotal, res.FilesSuccess, res.FilesFailed, res.ErrorMessage, res.ID.String())
	return err
}

// ChunkExists checks if a chunk with given hash exists
func (c *Connection) ChunkExists(hash string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM chunks WHERE hash = ?)`
	err := c.db.QueryRow(query, hash).Scan(&exists)
	return exists, err
}

// IncrementChunkRef increments the reference count for a chunk
func (c *Connection) IncrementChunkRef(hash string) error {
	query := `UPDATE chunks SET ref_count = ref_count + 1, last_accessed = CURRENT_TIMESTAMP WHERE hash = ?`
	_, err := c.db.Exec(query, hash)
	return err
}

// CreateChunk is an alias for AddChunk with basic params
func (c *Connection) CreateChunk(hash string, size int64, path string) error {
	return c.AddChunk(hash, size, 0, path, uuid.Nil, "performance")
}

// SaveRestorePointChunk saves the mapping between a restore point and a chunk
func (c *Connection) SaveRestorePointChunk(rpID uuid.UUID, hash string, sequence int, relPath string) error {
	query := `INSERT INTO restore_point_chunks (restore_point_id, chunk_hash, sequence_order, original_path)
	          VALUES (?, ?, ?, ?)`
	_, err := c.db.Exec(query, rpID.String(), hash, sequence, relPath)
	return err
}

// GetExpiredRestorePoints returns restore points older than limit
func (c *Connection) GetExpiredRestorePoints(jobID uuid.UUID, days int) ([]models.RestorePoint, error) {
	query := `SELECT id, job_id, point_time, status, total_bytes FROM restore_points 
	          WHERE job_id = ? AND point_time < DATETIME('now', ?)`

	rows, err := c.db.Query(query, jobID.String(), fmt.Sprintf("-%d days", days))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []models.RestorePoint
	for rows.Next() {
		var rp models.RestorePoint
		if err := rows.Scan(&rp.ID, &rp.JobID, &rp.PointTime, &rp.Status, &rp.TotalBytes); err != nil {
			return nil, err
		}
		points = append(points, rp)
	}
	return points, nil
}

// GetRestorePointChunks returns all chunk hashes for a restore point
func (c *Connection) GetRestorePointMapping(rpID uuid.UUID) ([]models.ChunkMapping, error) {
	query := `SELECT chunk_hash, sequence_order, original_path FROM restore_point_chunks 
	          WHERE restore_point_id = ? ORDER BY original_path, sequence_order`

	rows, err := c.db.Query(query, rpID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mappings []models.ChunkMapping
	for rows.Next() {
		var m models.ChunkMapping
		if err := rows.Scan(&m.ChunkHash, &m.Sequence, &m.OriginalPath); err != nil {
			return nil, err
		}
		mappings = append(mappings, m)
	}
	return mappings, nil
}

// GetRestorePointChunks returns all chunk hashes for a restore point
func (c *Connection) GetRestorePointChunks(rpID uuid.UUID) ([]string, error) {
	query := `SELECT chunk_hash FROM restore_point_chunks WHERE restore_point_id = ? ORDER BY sequence_order`
	rows, err := c.db.Query(query, rpID.String())
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

// GetChunkPath returns the storage path for a chunk
func (c *Connection) GetChunkPath(hash string) (string, error) {
	var path string
	query := `SELECT storage_path FROM chunks WHERE hash = ?`
	err := c.db.QueryRow(query, hash).Scan(&path)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return path, err
}

// GetRestorePointTotalSize returns the total size of a restore point
func (c *Connection) GetRestorePointTotalSize(rpID uuid.UUID) (int64, error) {
	var size int64
	query := `SELECT total_bytes FROM restore_points WHERE id = ?`
	err := c.db.QueryRow(query, rpID.String()).Scan(&size)
	return size, err
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
		`CREATE TABLE IF NOT EXISTS infrastructure_nodes (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			ip_address TEXT NOT NULL,
			node_type TEXT NOT NULL,
			username TEXT,
			password_encrypted TEXT,
			status TEXT DEFAULT 'online',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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
