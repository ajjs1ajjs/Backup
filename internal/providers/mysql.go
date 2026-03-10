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

// MySQLBackupProvider handles MySQL database backups
type MySQLBackupProvider struct {
	host     string
	port     int
	user     string
	password string
	database string
	useSSL   bool
}

// MySQLConfig contains MySQL connection configuration
type MySQLConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	UseSSL   bool
}

// NewMySQLBackupProvider creates a new MySQL backup provider
func NewMySQLBackupProvider(cfg MySQLConfig) *MySQLBackupProvider {
	if cfg.Port == 0 {
		cfg.Port = 3306
	}
	return &MySQLBackupProvider{
		host:     cfg.Host,
		port:     cfg.Port,
		user:     cfg.User,
		password: cfg.Password,
		database: cfg.Database,
		useSSL:   cfg.UseSSL,
	}
}

// Backup performs a MySQL database backup using mysqldump
func (m *MySQLBackupProvider) Backup(ctx context.Context, dest string) (*models.BackupResult, error) {
	result := &models.BackupResult{
		StartTime: time.Now(),
	}

	// Check if mysqldump is available
	mysqldumpPath, err := exec.LookPath("mysqldump")
	if err != nil {
		return nil, fmt.Errorf("mysqldump not found in PATH: %w", err)
	}

	// Create destination directory
	if err := os.MkdirAll(dest, 0755); err != nil {
		return nil, fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Generate backup filename
	timestamp := time.Now().Format("20060102_150405")
	backupFile := filepath.Join(dest, fmt.Sprintf("%s_%s.sql", m.database, timestamp))

	// Build mysqldump command
	args := []string{
		"-h", m.host,
		"-P", fmt.Sprintf("%d", m.port),
		"-u", m.user,
	}

	if m.password != "" {
		args = append(args, fmt.Sprintf("-p%s", m.password))
	}

	if !m.useSSL {
		args = append(args, "--ssl-mode=DISABLED")
	}

	// Add database name
	args = append(args, m.database)

	// Execute mysqldump
	cmd := exec.CommandContext(ctx, mysqldumpPath, args...)

	// Redirect output to file
	outFile, err := os.Create(backupFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup file: %w", err)
	}
	defer outFile.Close()

	cmd.Stdout = outFile
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		result.Status = models.JobStatusFailed
		result.ErrorMessage = err.Error()
		result.EndTime = time.Now()
		return result, fmt.Errorf("mysqldump failed: %w", err)
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

// TestConnection tests the MySQL connection
func (m *MySQLBackupProvider) TestConnection(ctx context.Context) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		m.user, m.password, m.host, m.port, m.database)

	if !m.useSSL {
		dsn += "&tls=false"
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
	}
	defer db.Close()

	// Set connection timeout
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

	return nil
}

// GetDatabaseSize returns the size of the database in bytes
func (m *MySQLBackupProvider) GetDatabaseSize(ctx context.Context) (int64, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/information_schema?parseTime=true",
		m.user, m.password, m.host, m.port)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	query := `SELECT SUM(data_length + index_length)
			  FROM tables
			  WHERE table_schema = ?`

	var size sql.NullInt64
	err = db.QueryRowContext(ctx, query, m.database).Scan(&size)
	if err != nil {
		return 0, err
	}

	if !size.Valid {
		return 0, nil
	}

	return size.Int64, nil
}

// ListDatabases lists all databases accessible by the user
func (m *MySQLBackupProvider) ListDatabases(ctx context.Context) ([]string, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?parseTime=true",
		m.user, m.password, m.host, m.port)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.QueryContext(ctx, "SHOW DATABASES")
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

// BackupResult contains MySQL backup results
type MySQLBackupResult struct {
	BackupFile   string
	Size         int64
	Duration     time.Duration
	TablesBacked int
	TablesFailed int
	ErrorMessage string
}
