package providers

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"novabackup/pkg/models"
)

// PostgreSQLBackupProvider handles PostgreSQL database backups
type PostgreSQLBackupProvider struct {
	host       string
	port       int
	user       string
	password   string
	database   string
	sslMode    string
	schemaOnly bool
}

// PostgreSQLConfig contains PostgreSQL connection configuration
type PostgreSQLConfig struct {
	Host       string
	Port       int
	User       string
	Password   string
	Database   string
	SSLMode    string
	SchemaOnly bool
}

// NewPostgreSQLBackupProvider creates a new PostgreSQL backup provider
func NewPostgreSQLBackupProvider(cfg PostgreSQLConfig) *PostgreSQLBackupProvider {
	if cfg.Port == 0 {
		cfg.Port = 5432
	}
	if cfg.SSLMode == "" {
		cfg.SSLMode = "disable"
	}
	return &PostgreSQLBackupProvider{
		host:       cfg.Host,
		port:       cfg.Port,
		user:       cfg.User,
		password:   cfg.Password,
		database:   cfg.Database,
		sslMode:    cfg.SSLMode,
		schemaOnly: cfg.SchemaOnly,
	}
}

// Backup performs a PostgreSQL database backup using pg_dump
func (p *PostgreSQLBackupProvider) Backup(ctx context.Context, dest string) (*models.BackupResult, error) {
	result := &models.BackupResult{
		StartTime: time.Now(),
	}

	// Check if pg_dump is available
	pgdumpPath, err := exec.LookPath("pg_dump")
	if err != nil {
		return nil, fmt.Errorf("pg_dump not found in PATH: %w", err)
	}

	// Create destination directory
	if err := os.MkdirAll(dest, 0755); err != nil {
		return nil, fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Generate backup filename
	timestamp := time.Now().Format("20060102_150405")
	backupFile := filepath.Join(dest, fmt.Sprintf("%s_%s.sql", p.database, timestamp))

	// Build pg_dump command
	args := []string{
		"-h", p.host,
		"-p", fmt.Sprintf("%d", p.port),
		"-U", p.user,
		"-d", p.database,
		"-f", backupFile,
	}

	// Add format option (custom format for better compression)
	args = append(args, "-F", "c")

	// Schema only option
	if p.schemaOnly {
		args = append(args, "--schema-only")
	}

	// Execute pg_dump with environment variable for password
	cmd := exec.CommandContext(ctx, pgdumpPath, args...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", p.password))
	cmd.Env = append(cmd.Env, fmt.Sprintf("PGSSLMODE=%s", p.sslMode))

	// Redirect stderr to stdout for error capture
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		result.Status = models.JobStatusFailed
		result.ErrorMessage = err.Error()
		result.EndTime = time.Now()
		return result, fmt.Errorf("pg_dump failed: %w", err)
	}

	// Get file info
	fileInfo, err := os.Stat(backupFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get backup file info: %w", err)
	}

	result.Status = models.JobStatusCompleted
	result.EndTime = time.Now()
	result.BytesWritten = fileInfo.Size()
	result.FilesTotal = 1
	result.FilesSuccess = 1

	return result, nil
}

// BackupSQL performs a PostgreSQL backup in plain SQL format
func (p *PostgreSQLBackupProvider) BackupSQL(ctx context.Context, dest string) (*models.BackupResult, error) {
	result := &models.BackupResult{
		StartTime: time.Now(),
	}

	pgdumpPath, err := exec.LookPath("pg_dump")
	if err != nil {
		return nil, fmt.Errorf("pg_dump not found in PATH: %w", err)
	}

	if err := os.MkdirAll(dest, 0755); err != nil {
		return nil, fmt.Errorf("failed to create destination directory: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	backupFile := filepath.Join(dest, fmt.Sprintf("%s_%s.sql", p.database, timestamp))

	args := []string{
		"-h", p.host,
		"-p", fmt.Sprintf("%d", p.port),
		"-U", p.user,
		"-d", p.database,
		"-F", "p", // Plain SQL format
	}

	if p.schemaOnly {
		args = append(args, "--schema-only")
	}

	outFile, err := os.Create(backupFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup file: %w", err)
	}
	defer outFile.Close()

	cmd := exec.CommandContext(ctx, pgdumpPath, args...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", p.password))
	cmd.Stdout = outFile
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		result.Status = models.JobStatusFailed
		result.ErrorMessage = err.Error()
		result.EndTime = time.Now()
		return result, fmt.Errorf("pg_dump failed: %w", err)
	}

	fileInfo, err := os.Stat(backupFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get backup file info: %w", err)
	}

	result.Status = models.JobStatusCompleted
	result.EndTime = time.Now()
	result.BytesWritten = fileInfo.Size()
	result.FilesTotal = 1
	result.FilesSuccess = 1

	return result, nil
}

// TestConnection tests the PostgreSQL connection
func (p *PostgreSQLBackupProvider) TestConnection(ctx context.Context) error {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.host, p.port, p.user, p.password, p.database, p.sslMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

	return nil
}

// GetDatabaseSize returns the size of the database in bytes
func (p *PostgreSQLBackupProvider) GetDatabaseSize(ctx context.Context) (int64, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.host, p.port, p.user, p.password, p.database, p.sslMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	query := `SELECT pg_database_size(current_database())`

	var size int64
	err = db.QueryRowContext(ctx, query).Scan(&size)
	if err != nil {
		return 0, err
	}

	return size, nil
}

// ListDatabases lists all databases accessible by the user
func (p *PostgreSQLBackupProvider) ListDatabases(ctx context.Context) ([]string, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=%s",
		p.host, p.port, p.user, p.password, p.sslMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.QueryContext(ctx, "SELECT datname FROM pg_database WHERE datistemplate = false")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var db string
		if err := rows.Scan(&db); err != nil {
			return nil, err
		}
		databases = append(databases, db)
	}

	return databases, rows.Err()
}

// ListSchemas lists all schemas in the database
func (p *PostgreSQLBackupProvider) ListSchemas(ctx context.Context) ([]string, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.host, p.port, p.user, p.password, p.database, p.sslMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.QueryContext(ctx, "SELECT schema_name FROM information_schema.schemata")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schemas []string
	for rows.Next() {
		var schema string
		if err := rows.Scan(&schema); err != nil {
			return nil, err
		}
		schemas = append(schemas, schema)
	}

	return schemas, rows.Err()
}

// Restore restores a PostgreSQL database from a backup file
func (p *PostgreSQLBackupProvider) Restore(ctx context.Context, backupFile string) error {
	pgrestorePath, err := exec.LookPath("pg_restore")
	if err != nil {
		return fmt.Errorf("pg_restore not found in PATH: %w", err)
	}

	args := []string{
		"-h", p.host,
		"-p", fmt.Sprintf("%d", p.port),
		"-U", p.user,
		"-d", p.database,
		"--clean",
		"--if-exists",
		backupFile,
	}

	cmd := exec.CommandContext(ctx, pgrestorePath, args...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", p.password))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// PostgreSQLBackupResult contains PostgreSQL backup results
type PostgreSQLBackupResult struct {
	BackupFile   string
	Size         int64
	Format       string
	Duration     time.Duration
	TablesBacked int
	ErrorMessage string
}
