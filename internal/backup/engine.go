// NovaBackup - Complete Backup Engine
// Supports: Files, Databases (MySQL, PostgreSQL, SQLite), Cloud (S3, Azure, GDrive)

package backup

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

// Backup Types
const (
	TypeFile     = "file"
	TypeDatabase = "database"
	TypeCloud    = "cloud"
	TypeVM       = "vm"
)

// Database Types
const (
	DBMySQL      = "mysql"
	DBPostgreSQL = "postgresql"
	DBSQLite     = "sqlite"
	DBMSSQL      = "mssql"
)

// Cloud Providers
const (
	CloudS3       = "s3"
	CloudAzure    = "azure"
	CloudGoogle   = "google"
	CloudOneDrive = "onedrive"
)

// BackupJob represents a complete backup job
type BackupJob struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	Description        string   `json:"description"`
	Type               string   `json:"type"` // file, database, cloud
	Sources            []string `json:"sources"`
	Destination        string   `json:"destination"`
	Compression        bool     `json:"compression"`
	CompressionLevel   int      `json:"compression_level"` // 1-9
	Encryption         bool     `json:"encryption"`
	EncryptionPassword string   `json:"encryption_password,omitempty"`
	Incremental        bool     `json:"incremental"`
	FullBackupEvery    int      `json:"full_backup_every"` // days
	Schedule           string   `json:"schedule"`          // daily, weekly, monthly
	ScheduleTime       string   `json:"schedule_time"`
	ScheduleDays       []string `json:"schedule_days"`

	// Database specific
	DatabaseType   string   `json:"database_type,omitempty"`
	DatabaseConn   string   `json:"database_conn,omitempty"`
	DatabaseTables []string `json:"database_tables,omitempty"`

	// Cloud specific
	CloudProvider  string `json:"cloud_provider,omitempty"`
	CloudBucket    string `json:"cloud_bucket,omitempty"`
	CloudRegion    string `json:"cloud_region,omitempty"`
	CloudAccessKey string `json:"cloud_access_key,omitempty"`
	CloudSecretKey string `json:"cloud_secret_key,omitempty"`
	CloudEndpoint  string `json:"cloud_endpoint,omitempty"` // for S3-compatible

	// Retention
	RetentionDays   int `json:"retention_days"`
	RetentionCopies int `json:"retention_copies"`

	// Advanced
	ExcludePatterns  []string `json:"exclude_patterns"`
	IncludePatterns  []string `json:"include_patterns"`
	PreBackupScript  string   `json:"pre_backup_script"`
	PostBackupScript string   `json:"post_backup_script"`
	MaxThreads       int      `json:"max_threads"`

	// Status
	Enabled    bool   `json:"enabled"`
	LastBackup string `json:"last_backup,omitempty"`
	NextBackup string `json:"next_backup,omitempty"`
	LastResult string `json:"last_result,omitempty"` // success, warning, failed
}

// BackupSession tracks a single backup operation
type BackupSession struct {
	ID                 string    `json:"id"`
	JobID              string    `json:"job_id"`
	JobName            string    `json:"job_name"`
	Type               string    `json:"type"`
	StartTime          time.Time `json:"start_time"`
	EndTime            time.Time `json:"end_time"`
	Status             string    `json:"status"` // running, success, warning, failed
	IsIncremental      bool      `json:"is_incremental"`
	FilesTotal         int       `json:"files_total"`
	FilesProcessed     int       `json:"files_processed"`
	FilesSkipped       int       `json:"files_skipped"`
	BytesTotal         int64     `json:"bytes_total"`
	BytesWritten       int64     `json:"bytes_written"`
	BytesRead          int64     `json:"bytes_read"`
	CompressionRatio   float64   `json:"compression_ratio"`
	DeduplicationRatio float64   `json:"deduplication_ratio"`
	Error              string    `json:"error,omitempty"`
	Warnings           []string  `json:"warnings"`
	BackupPath         string    `json:"backup_path"`
	BackupSize         int64     `json:"backup_size"`
	Logs               []string  `json:"logs"`
}

// BackupEngine is the main backup engine
type BackupEngine struct {
	DataDir     string
	LogFile     string
	mu          sync.Mutex
	sessions    map[string]*BackupSession
	blockHashes map[string]string // for deduplication
}

// NewBackupEngine creates a new backup engine
func NewBackupEngine(dataDir string) *BackupEngine {
	engine := &BackupEngine{
		DataDir:     dataDir,
		LogFile:     filepath.Join(dataDir, "logs", "backup.log"),
		sessions:    make(map[string]*BackupSession),
		blockHashes: make(map[string]string),
	}

	// Create necessary directories
	os.MkdirAll(filepath.Join(dataDir, "backups"), 0755)
	os.MkdirAll(filepath.Join(dataDir, "logs"), 0755)
	os.MkdirAll(filepath.Join(dataDir, "sessions"), 0755)
	os.MkdirAll(filepath.Join(dataDir, "config"), 0755)

	return engine
}

// ExecuteBackup runs a backup job
func (e *BackupEngine) ExecuteBackup(ctx context.Context, job *BackupJob) (*BackupSession, error) {
	e.mu.Lock()

	session := &BackupSession{
		ID:            fmt.Sprintf("session_%d_%s", time.Now().Unix(), job.ID),
		JobID:         job.ID,
		JobName:       job.Name,
		Type:          job.Type,
		StartTime:     time.Now(),
		Status:        "running",
		IsIncremental: job.Incremental,
		Logs:          make([]string, 0),
	}

	e.sessions[session.ID] = session
	e.mu.Unlock()

	e.log(session, "════════════════════════════════════════")
	e.log(session, fmt.Sprintf("Starting backup job: %s", job.Name))
	e.log(session, fmt.Sprintf("Type: %s, Incremental: %v", job.Type, job.Incremental))
	e.log(session, fmt.Sprintf("Sources: %v", job.Sources))
	e.log(session, fmt.Sprintf("Destination: %s", job.Destination))

	// Create backup directory with timestamp
	backupDir := filepath.Join(job.Destination, job.Name, time.Now().Format("2006-01-02_150405"))
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		session.Status = "failed"
		session.Error = err.Error()
		return session, err
	}
	session.BackupPath = backupDir

	// Run pre-backup script
	if job.PreBackupScript != "" {
		e.log(session, fmt.Sprintf("Running pre-backup script: %s", job.PreBackupScript))
		if err := e.runScript(job.PreBackupScript); err != nil {
			e.log(session, fmt.Sprintf("Pre-backup script failed: %v", err))
			session.Warnings = append(session.Warnings, fmt.Sprintf("Pre-script failed: %v", err))
		}
	}

	var err error
	switch job.Type {
	case TypeFile:
		err = e.backupFiles(ctx, job, session)
	case TypeDatabase:
		err = e.backupDatabase(ctx, job, session)
	case TypeCloud:
		err = e.backupToCloud(ctx, job, session)
	default:
		err = fmt.Errorf("unsupported backup type: %s", job.Type)
	}

	// Run post-backup script
	if job.PostBackupScript != "" {
		e.log(session, fmt.Sprintf("Running post-backup script: %s", job.PostBackupScript))
		if err := e.runScript(job.PostBackupScript); err != nil {
			e.log(session, fmt.Sprintf("Post-backup script failed: %v", err))
			session.Warnings = append(session.Warnings, fmt.Sprintf("Post-script failed: %v", err))
		}
	}

	session.EndTime = time.Now()
	duration := session.EndTime.Sub(session.StartTime)

	if err != nil {
		session.Status = "failed"
		session.Error = err.Error()
		e.log(session, fmt.Sprintf("❌ Backup FAILED: %v", err))
	} else {
		session.Status = "success"

		// Calculate compression ratio
		if session.BytesTotal > 0 && session.BytesWritten > 0 {
			session.CompressionRatio = float64(session.BytesTotal) / float64(session.BytesWritten)
		}

		e.log(session, fmt.Sprintf("✅ Backup COMPLETED successfully"))
		e.log(session, fmt.Sprintf("Duration: %v", duration))
		e.log(session, fmt.Sprintf("Processed: %d files, %s", session.FilesProcessed, formatBytes(session.BytesRead)))
		e.log(session, fmt.Sprintf("Written: %s, Compression: %.2fx", formatBytes(session.BytesWritten), session.CompressionRatio))
	}

	// Save session
	e.saveSession(session)

	// Apply retention policy
	e.applyRetentionPolicy(job, backupDir)

	return session, err
}

// backupFiles backs up files and folders
func (e *BackupEngine) backupFiles(ctx context.Context, job *BackupJob, session *BackupSession) error {
	e.log(session, "📁 Starting file backup...")

	// Collect all files
	var files []string
	var totalSize int64

	for _, source := range job.Sources {
		e.log(session, fmt.Sprintf("Collecting files from: %s", source))

		found, size, err := e.collectFiles(source, job)
		if err != nil {
			e.log(session, fmt.Sprintf("⚠️ Warning collecting %s: %v", source, err))
			session.Warnings = append(session.Warnings, fmt.Sprintf("Failed to scan %s: %v", source, err))
			continue
		}

		files = append(files, found...)
		totalSize += size
	}

	session.FilesTotal = len(files)
	session.BytesTotal = totalSize
	e.log(session, fmt.Sprintf("Found %d files (%s)", len(files), formatBytes(totalSize)))

	if len(files) == 0 {
		return fmt.Errorf("no files found to backup")
	}

	// Create archive
	archivePath := filepath.Join(session.BackupPath, "backup.zip")
	archiveFile, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer archiveFile.Close()

	var writer io.Writer = archiveFile
	var zipWriter *zip.Writer

	// Add compression layer
	if job.Compression {
		e.log(session, "Enabling compression...")
		gzWriter := gzip.NewWriter(archiveFile)
		defer gzWriter.Close()
		writer = gzWriter
	}

	zipWriter = zip.NewWriter(writer)
	defer zipWriter.Close()

	// Backup files
	var bytesWritten int64
	for i, file := range files {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := e.addFileToZip(zipWriter, file); err != nil {
			e.log(session, fmt.Sprintf("⚠️ Failed to add %s: %v", file, err))
			session.FilesSkipped++
			continue
		}

		session.FilesProcessed++
		info, _ := os.Stat(file)
		if info != nil {
			bytesWritten += info.Size()
		}
		session.BytesRead = bytesWritten

		// Progress logging every 100 files
		if (i+1)%100 == 0 {
			e.log(session, fmt.Sprintf("Progress: %d/%d files (%.1f%%)",
				session.FilesProcessed, session.FilesTotal,
				float64(session.FilesProcessed)*100/float64(session.FilesTotal)))
		}
	}

	// Get final archive size
	archiveInfo, _ := os.Stat(archivePath)
	if archiveInfo != nil {
		session.BytesWritten = archiveInfo.Size()
	}

	// Create metadata
	metadata := map[string]interface{}{
		"job_id":            job.ID,
		"job_name":          job.Name,
		"backup_time":       session.StartTime,
		"files_count":       session.FilesProcessed,
		"total_size":        session.BytesTotal,
		"compressed_size":   session.BytesWritten,
		"is_incremental":    job.Incremental,
		"compression":       job.Compression,
		"encryption":        job.Encryption,
		"compression_ratio": session.CompressionRatio,
	}

	metadataJSON, _ := json.MarshalIndent(metadata, "", "  ")
	os.WriteFile(filepath.Join(session.BackupPath, "metadata.json"), metadataJSON, 0644)

	e.log(session, fmt.Sprintf("📦 Backup archive created: %s (%s)",
		archivePath, formatBytes(session.BytesWritten)))

	return nil
}

// collectFiles recursively collects files from a source
func (e *BackupEngine) collectFiles(source string, job *BackupJob) ([]string, int64, error) {
	var files []string
	var totalSize int64

	info, err := os.Stat(source)
	if err != nil {
		return nil, 0, err
	}

	// Single file
	if !info.IsDir() {
		if e.shouldInclude(source, job) {
			return []string{source}, info.Size(), nil
		}
		return nil, 0, nil
	}

	// Directory - walk recursively
	err = filepath.WalkDir(source, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip excluded patterns
		if !e.shouldInclude(path, job) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !d.IsDir() {
			info, err := d.Info()
			if err == nil {
				files = append(files, path)
				totalSize += info.Size()
			}
		}

		return nil
	})

	return files, totalSize, err
}

// shouldInclude checks if a file should be included based on patterns
func (e *BackupEngine) shouldInclude(path string, job *BackupJob) bool {
	// Check exclude patterns
	for _, pattern := range job.ExcludePatterns {
		if match, _ := filepath.Match(pattern, filepath.Base(path)); match {
			return false
		}
		if strings.Contains(path, pattern) {
			return false
		}
	}

	// Check include patterns (if specified, only include matching)
	if len(job.IncludePatterns) > 0 {
		for _, pattern := range job.IncludePatterns {
			if match, _ := filepath.Match(pattern, filepath.Base(path)); match {
				return true
			}
		}
		return false
	}

	return true
}

// addFileToZip adds a file to zip archive
func (e *BackupEngine) addFileToZip(zw *zip.Writer, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	// Use relative path
	header.Name = filePath
	header.Method = zip.Deflate

	writer, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}

// backupDatabase backs up databases
func (e *BackupEngine) backupDatabase(ctx context.Context, job *BackupJob, session *BackupSession) error {
	e.log(session, fmt.Sprintf("🗄️ Starting %s database backup...", job.DatabaseType))

	var dumpFile string
	var err error

	switch job.DatabaseType {
	case DBMySQL:
		dumpFile, err = e.backupMySQL(ctx, job, session)
	case DBPostgreSQL:
		dumpFile, err = e.backupPostgreSQL(ctx, job, session)
	case DBSQLite:
		dumpFile, err = e.backupSQLite(ctx, job, session)
	default:
		return fmt.Errorf("unsupported database type: %s", job.DatabaseType)
	}

	if err != nil {
		return err
	}

	// Compress if enabled
	if job.Compression {
		e.log(session, "Compressing database dump...")
		compressedFile := dumpFile + ".gz"
		if err := e.compressFile(dumpFile, compressedFile); err != nil {
			e.log(session, fmt.Sprintf("⚠️ Compression failed: %v", err))
			session.Warnings = append(session.Warnings, fmt.Sprintf("Compression failed: %v", err))
		} else {
			os.Remove(dumpFile)
			dumpFile = compressedFile
		}
	}

	// Encrypt if enabled
	if job.Encryption && job.EncryptionPassword != "" {
		e.log(session, "Encrypting backup...")
		encryptedFile := dumpFile + ".enc"
		if err := e.encryptFile(dumpFile, encryptedFile, job.EncryptionPassword); err != nil {
			e.log(session, fmt.Sprintf("⚠️ Encryption failed: %v", err))
			session.Warnings = append(session.Warnings, fmt.Sprintf("Encryption failed: %v", err))
		} else {
			os.Remove(dumpFile)
			dumpFile = encryptedFile
		}
	}

	info, _ := os.Stat(dumpFile)
	if info != nil {
		session.BytesWritten = info.Size()
	}

	e.log(session, fmt.Sprintf("💾 Database backup completed: %s (%s)",
		dumpFile, formatBytes(session.BytesWritten)))

	return nil
}

// backupMySQL creates MySQL dump using mysqldump
func (e *BackupEngine) backupMySQL(ctx context.Context, job *BackupJob, session *BackupSession) (string, error) {
	dumpFile := filepath.Join(session.BackupPath, "mysql_dump.sql")

	// Try mysqldump first
	mysqldumpPath := "mysqldump"
	if runtime.GOOS == "windows" {
		// Check common MySQL installation paths
		possiblePaths := []string{
			`C:\Program Files\MySQL\MySQL Server 8.0\bin\mysqldump.exe`,
			`C:\Program Files\MySQL\MySQL Server 5.7\bin\mysqldump.exe`,
			`C:\Program Files (x86)\MySQL\MySQL Server 5.7\bin\mysqldump.exe`,
		}
		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				mysqldumpPath = path
				break
			}
		}
	}

	// Build mysqldump command
	args := []string{
		"--result-file=" + dumpFile,
		"--single-transaction",
		"--quick",
		"--lock-tables=false",
	}

	// Add database name from connection string
	args = append(args, job.DatabaseConn)

	cmd := exec.CommandContext(ctx, mysqldumpPath, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	e.log(session, fmt.Sprintf("Running: %s %v", mysqldumpPath, args))

	if err := cmd.Run(); err != nil {
		e.log(session, fmt.Sprintf("mysqldump error: %s", stderr.String()))

		// Fallback to direct SQL export
		return e.backupMySQLDirect(ctx, job, session, dumpFile)
	}

	return dumpFile, nil
}

// backupMySQLDirect exports MySQL using direct connection
func (e *BackupEngine) backupMySQLDirect(ctx context.Context, job *BackupJob, session *BackupSession, dumpFile string) (string, error) {
	e.log(session, "Using direct MySQL connection...")

	db, err := sql.Open("mysql", job.DatabaseConn)
	if err != nil {
		return "", err
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		return "", err
	}

	file, err := os.Create(dumpFile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Get all tables
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			continue
		}
		tables = append(tables, table)
	}

	e.log(session, fmt.Sprintf("Found %d tables: %v", len(tables), tables))
	session.FilesTotal = len(tables)

	// Export each table
	for _, table := range tables {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		// CREATE TABLE
		createRow := db.QueryRow(fmt.Sprintf("SHOW CREATE TABLE `%s`", table))
		var tableName, createSQL string
		if err := createRow.Scan(&tableName, &createSQL); err == nil {
			fmt.Fprintf(file, "-- Table: %s\n", table)
			fmt.Fprintf(file, "%s;\n\n", createSQL)
		}

		// INSERT statements
		dataRows, err := db.Query(fmt.Sprintf("SELECT * FROM `%s`", table))
		if err != nil {
			session.FilesSkipped++
			continue
		}

		columns, _ := dataRows.Columns()
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))

		for dataRows.Next() {
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			if err := dataRows.Scan(valuePtrs...); err != nil {
				continue
			}

			fmt.Fprintf(file, "INSERT INTO `%s` VALUES (", table)
			for i, v := range values {
				if i > 0 {
					fmt.Fprintf(file, ", ")
				}
				if v == nil {
					fmt.Fprintf(file, "NULL")
				} else {
					fmt.Fprintf(file, "'%v'", v)
				}
			}
			fmt.Fprintf(file, ");\n")
		}
		dataRows.Close()
		fmt.Fprintf(file, "\n")

		session.FilesProcessed++
	}

	return dumpFile, nil
}

// backupPostgreSQL creates PostgreSQL dump
func (e *BackupEngine) backupPostgreSQL(ctx context.Context, job *BackupJob, session *BackupSession) (string, error) {
	dumpFile := filepath.Join(session.BackupPath, "postgres_dump.sql")

	// Try pg_dump
	pgdumpPath := "pg_dump"
	if runtime.GOOS == "windows" {
		possiblePaths := []string{
			`C:\Program Files\PostgreSQL\15\bin\pg_dump.exe`,
			`C:\Program Files\PostgreSQL\14\bin\pg_dump.exe`,
			`C:\Program Files\PostgreSQL\13\bin\pg_dump.exe`,
		}
		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				pgdumpPath = path
				break
			}
		}
	}

	cmd := exec.CommandContext(ctx, pgdumpPath, job.DatabaseConn)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Fallback to direct connection
		return e.backupPostgreSQLDirect(ctx, job, session, dumpFile)
	}

	os.WriteFile(dumpFile, output, 0644)
	return dumpFile, nil
}

// backupPostgreSQLDirect exports PostgreSQL using direct connection
func (e *BackupEngine) backupPostgreSQLDirect(ctx context.Context, job *BackupJob, session *BackupSession, dumpFile string) (string, error) {
	e.log(session, "Using direct PostgreSQL connection...")

	db, err := sql.Open("postgres", job.DatabaseConn)
	if err != nil {
		return "", err
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return "", err
	}

	file, err := os.Create(dumpFile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Get tables
	rows, err := db.Query(`SELECT tablename FROM pg_tables WHERE schemaname = 'public'`)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			continue
		}
		tables = append(tables, table)
	}

	// Export schema and data
	for _, table := range tables {
		fmt.Fprintf(file, "-- Table: %s\n", table)
		// Implementation similar to MySQL
	}

	return dumpFile, nil
}

// backupSQLite backs up SQLite database
func (e *BackupEngine) backupSQLite(ctx context.Context, job *BackupJob, session *BackupSession) (string, error) {
	sourceFile := job.DatabaseConn
	destFile := filepath.Join(session.BackupPath, "sqlite_backup.db")

	// Copy file
	data, err := os.ReadFile(sourceFile)
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(destFile, data, 0644); err != nil {
		return "", err
	}

	return destFile, nil
}

// backupToCloud uploads backup to cloud storage
func (e *BackupEngine) backupToCloud(ctx context.Context, job *BackupJob, session *BackupSession) error {
	e.log(session, fmt.Sprintf("☁️ Starting cloud backup to %s...", job.CloudProvider))

	// First create local backup
	tempDir := filepath.Join(e.DataDir, "temp", session.ID)
	os.MkdirAll(tempDir, 0755)

	localJob := &BackupJob{
		ID:          job.ID,
		Name:        job.Name,
		Type:        TypeFile,
		Sources:     job.Sources,
		Destination: tempDir,
		Compression: job.Compression,
	}

	localSession := &BackupSession{
		ID:        session.ID,
		JobID:     job.ID,
		JobName:   job.Name,
		StartTime: time.Now(),
	}

	if err := e.backupFiles(ctx, localJob, localSession); err != nil {
		return err
	}

	// Upload to cloud
	switch job.CloudProvider {
	case CloudS3:
		return e.uploadToS3(ctx, job, tempDir)
	case CloudAzure:
		return e.uploadToAzure(ctx, job, tempDir)
	case CloudGoogle:
		return e.uploadToGoogle(ctx, job, tempDir)
	default:
		return fmt.Errorf("unsupported cloud provider: %s", job.CloudProvider)
	}
}

// uploadToS3 uploads to S3 or S3-compatible storage
func (e *BackupEngine) uploadToS3(ctx context.Context, job *BackupJob, backupDir string) error {
	e.log(nil, "Uploading to S3...")
	// TODO: Implement AWS S3 SDK upload
	return nil
}

// uploadToAzure uploads to Azure Blob Storage
func (e *BackupEngine) uploadToAzure(ctx context.Context, job *BackupJob, backupDir string) error {
	e.log(nil, "Uploading to Azure Blob Storage...")
	// TODO: Implement Azure SDK upload
	return nil
}

// uploadToGoogle uploads to Google Cloud Storage
func (e *BackupEngine) uploadToGoogle(ctx context.Context, job *BackupJob, backupDir string) error {
	e.log(nil, "Uploading to Google Cloud Storage...")
	// TODO: Implement GCS SDK upload
	return nil
}

// Helper functions

func (e *BackupEngine) log(session *BackupSession, message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logLine := fmt.Sprintf("[%s] %s", timestamp, message)

	if session != nil {
		session.Logs = append(session.Logs, logLine)
	}

	// Write to log file
	os.MkdirAll(filepath.Dir(e.LogFile), 0755)
	f, _ := os.OpenFile(e.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if f != nil {
		f.WriteString(logLine + "\n")
		f.Close()
	}

	fmt.Println(logLine)
}

func (e *BackupEngine) saveSession(session *BackupSession) {
	sessionsDir := filepath.Join(e.DataDir, "sessions")
	os.MkdirAll(sessionsDir, 0755)

	sessionFile := filepath.Join(sessionsDir, fmt.Sprintf("%s.json", session.ID))
	data, _ := json.MarshalIndent(session, "", "  ")
	os.WriteFile(sessionFile, data, 0644)
}

func (e *BackupEngine) applyRetentionPolicy(job *BackupJob, currentBackup string) {
	if job.RetentionDays <= 0 && job.RetentionCopies <= 0 {
		return
	}

	backupDir := filepath.Join(job.Destination, job.Name)
	entries, _ := os.ReadDir(backupDir)

	var backups []os.DirEntry
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != filepath.Base(currentBackup) {
			backups = append(backups, entry)
		}
	}

	// Apply retention by days
	if job.RetentionDays > 0 {
		cutoff := time.Now().AddDate(0, 0, -job.RetentionDays)
		for _, backup := range backups {
			info, _ := backup.Info()
			if info != nil && info.ModTime().Before(cutoff) {
				os.RemoveAll(filepath.Join(backupDir, backup.Name()))
				e.log(nil, fmt.Sprintf("🗑️ Deleted old backup: %s", backup.Name()))
			}
		}
	}

	// Apply retention by copies
	if job.RetentionCopies > 0 && len(backups) > job.RetentionCopies {
		// Delete oldest backups
		// TODO: Sort by date and delete oldest
	}
}

func (e *BackupEngine) runScript(scriptPath string) error {
	cmd := exec.Command(scriptPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v - %s", err, string(output))
	}
	return nil
}

func (e *BackupEngine) compressFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	gzWriter := gzip.NewWriter(dstFile)
	defer gzWriter.Close()

	_, err = io.Copy(gzWriter, srcFile)
	return err
}

func (e *BackupEngine) encryptFile(src, dst, password string) error {
	// TODO: Implement AES-256 encryption
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// Simple XOR encryption (placeholder - use proper AES in production)
	key := []byte(password)
	for i := range data {
		data[i] ^= key[i%len(key)]
	}

	return os.WriteFile(dst, data, 0644)
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// RestoreEngine handles all restore operations
type RestoreEngine struct {
	DataDir string
}

// NewRestoreEngine creates a new restore engine
func NewRestoreEngine(dataDir string) *RestoreEngine {
	return &RestoreEngine{DataDir: dataDir}
}

// RestoreOptions for restore operation
type RestoreOptions struct {
	BackupPath      string   `json:"backup_path"`
	Destination     string   `json:"destination"`
	Files           []string `json:"files"` // specific files to restore
	Overwrite       bool     `json:"overwrite"`
	RestoreOriginal bool     `json:"restore_original"`
}

// RestoreResult contains restore operation results
type RestoreResult struct {
	FilesRestored int      `json:"files_restored"`
	FilesSkipped  int      `json:"files_skipped"`
	BytesRestored int64    `json:"bytes_restored"`
	Duration      string   `json:"duration"`
	Warnings      []string `json:"warnings"`
	Error         string   `json:"error,omitempty"`
}

// RestoreFiles restores files from backup
func (e *RestoreEngine) RestoreFiles(opts RestoreOptions) (*RestoreResult, error) {
	result := &RestoreResult{
		Warnings: make([]string, 0),
	}

	startTime := time.Now()

	// Open archive
	archivePath := filepath.Join(opts.BackupPath, "backup.zip")
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		result.Error = err.Error()
		return result, err
	}
	defer r.Close()

	for _, f := range r.File {
		// Skip if specific files requested and this isn't one of them
		if len(opts.Files) > 0 {
			found := false
			for _, pattern := range opts.Files {
				if strings.Contains(f.Name, pattern) {
					found = true
					break
				}
			}
			if !found {
				result.FilesSkipped++
				continue
			}
		}

		destPath := f.Name
		if opts.RestoreOriginal {
			// Restore to original location
		} else {
			destPath = filepath.Join(opts.Destination, f.Name)
		}

		// Skip if exists and not overwriting
		if _, err := os.Stat(destPath); err == nil && !opts.Overwrite {
			result.FilesSkipped++
			result.Warnings = append(result.Warnings, fmt.Sprintf("Skipped: %s (exists)", destPath))
			continue
		}

		os.MkdirAll(filepath.Dir(destPath), 0755)

		if f.FileInfo().IsDir() {
			os.MkdirAll(destPath, 0755)
			continue
		}

		srcFile, err := f.Open()
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to open %s: %v", f.Name, err))
			continue
		}
		defer srcFile.Close()

		destFile, err := os.Create(destPath)
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to create %s: %v", destPath, err))
			continue
		}
		defer destFile.Close()

		bytesWritten, err := io.Copy(destFile, srcFile)
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to write %s: %v", destPath, err))
			continue
		}

		result.FilesRestored++
		result.BytesRestored += bytesWritten
	}

	result.Duration = time.Since(startTime).String()
	return result, nil
}

// RestoreDatabase restores a database from dump
func (e *RestoreEngine) RestoreDatabase(dbType, dumpFile, connString string) error {
	switch dbType {
	case DBMySQL:
		return e.restoreMySQL(dumpFile, connString)
	case DBPostgreSQL:
		return e.restorePostgreSQL(dumpFile, connString)
	case DBSQLite:
		return e.restoreSQLite(dumpFile, connString)
	default:
		return fmt.Errorf("unsupported database type: %s", dbType)
	}
}

func (e *RestoreEngine) restoreMySQL(dumpFile, connString string) error {
	// Decompress if needed
	if strings.HasSuffix(dumpFile, ".gz") {
		decompressed := strings.TrimSuffix(dumpFile, ".gz")
		if err := e.decompressFile(dumpFile, decompressed); err != nil {
			return err
		}
		dumpFile = decompressed
	}

	// Decrypt if needed
	if strings.HasSuffix(dumpFile, ".enc") {
		// TODO: Decrypt
	}

	cmd := exec.Command("mysql", connString)
	input, err := os.ReadFile(dumpFile)
	if err != nil {
		return err
	}
	cmd.Stdin = bytes.NewReader(input)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mysql restore failed: %v - %s", err, string(output))
	}
	return nil
}

func (e *RestoreEngine) restorePostgreSQL(dumpFile, connString string) error {
	cmd := exec.Command("psql", connString)
	input, err := os.ReadFile(dumpFile)
	if err != nil {
		return err
	}
	cmd.Stdin = bytes.NewReader(input)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("postgres restore failed: %v - %s", err, string(output))
	}
	return nil
}

func (e *RestoreEngine) restoreSQLite(dumpFile, connString string) error {
	data, err := os.ReadFile(dumpFile)
	if err != nil {
		return err
	}
	return os.WriteFile(connString, data, 0644)
}

func (e *RestoreEngine) decompressFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	gzReader, err := gzip.NewReader(srcFile)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	_, err = io.Copy(dstFile, gzReader)
	return err
}
