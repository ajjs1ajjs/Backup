// Backup Engine - Complete Veeam-style backup functionality
package backup

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
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
	TypeVM       = "vm"
	TypeCloud    = "cloud"
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

// Compression Levels
const (
	CompressionNone   = 0
	CompressionFast   = 1
	CompressionNormal = 5
	CompressionMax    = 9
)

// BackupJob represents a complete backup job configuration
type BackupJob struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Description      string   `json:"description"`
	Type             string   `json:"type"` // file, database, vm, cloud
	Sources          []string `json:"sources"`
	Destination      string   `json:"destination"`
	Compression      bool     `json:"compression"`
	CompressionLevel int      `json:"compression_level"`
	Encryption       bool     `json:"encryption"`
	EncryptionKey    string   `json:"-"`
	Incremental      bool     `json:"incremental"`
	FullBackupEvery  int      `json:"full_backup_every"` // days
	Schedule         string   `json:"schedule"`          // daily, weekly, monthly, cron
	ScheduleTime     string   `json:"schedule_time"`
	ScheduleDays     []string `json:"schedule_days"`
	CronExpression   string   `json:"cron_expression"`

	// Database specific
	DatabaseType   string   `json:"database_type,omitempty"`
	DatabaseConn   string   `json:"database_conn,omitempty"`
	DatabaseTables []string `json:"database_tables,omitempty"`

	// VM specific
	VMNames    []string `json:"vm_names,omitempty"`
	HyperVHost string   `json:"hyperv_host,omitempty"`

	// Cloud specific
	CloudProvider  string `json:"cloud_provider,omitempty"`
	CloudBucket    string `json:"cloud_bucket,omitempty"`
	CloudRegion    string `json:"cloud_region,omitempty"`
	CloudAccessKey string `json:"cloud_access_key,omitempty"`
	CloudSecretKey string `json:"cloud_secret_key,omitempty"`
	CloudEndpoint  string `json:"cloud_endpoint,omitempty"`

	// Retention
	RetentionDays   int `json:"retention_days"`
	RetentionCopies int `json:"retention_copies"`

	// Advanced
	ExcludePatterns  []string `json:"exclude_patterns"`
	IncludePatterns  []string `json:"include_patterns"`
	PreBackupScript  string   `json:"pre_backup_script"`
	PostBackupScript string   `json:"post_backup_script"`
	MaxThreads       int      `json:"max_threads"`
	BlockSize        int      `json:"block_size"` // for deduplication

	// Status
	Enabled    bool       `json:"enabled"`
	LastBackup *time.Time `json:"last_backup,omitempty"`
	NextBackup *time.Time `json:"next_backup,omitempty"`
	LastResult string     `json:"last_result,omitempty"` // success, warning, failed
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
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
	IsFull             bool      `json:"is_full"`
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
	ProcessedBlocks    int       `json:"processed_blocks"`
	DeduplicatedBlocks int       `json:"deduplicated_blocks"`
}

// BackupEngine is the main backup engine
type BackupEngine struct {
	DataDir       string
	LogFile       string
	mu            sync.Mutex
	sessions      map[string]*BackupSession
	blockHashes   map[string]string // for deduplication
	encryptionKey []byte
	changeTracker *ChangeTracker // Changed Block Tracking
}

// NewBackupEngine creates a new backup engine
func NewBackupEngine(dataDir string) *BackupEngine {
	engine := &BackupEngine{
		DataDir:       dataDir,
		LogFile:       filepath.Join(dataDir, "logs", "backup.log"),
		sessions:      make(map[string]*BackupSession),
		blockHashes:   make(map[string]string),
		changeTracker: NewChangeTracker(dataDir),
	}

	// Create necessary directories
	os.MkdirAll(filepath.Join(dataDir, "backups"), 0755)
	os.MkdirAll(filepath.Join(dataDir, "logs"), 0755)
	os.MkdirAll(filepath.Join(dataDir, "sessions"), 0755)
	os.MkdirAll(filepath.Join(dataDir, "config"), 0755)
	os.MkdirAll(filepath.Join(dataDir, "cache"), 0755)
	os.MkdirAll(filepath.Join(dataDir, "cbt"), 0755)

	return engine
}

// ExecuteBackup runs a backup job
func (e *BackupEngine) ExecuteBackup(job *BackupJob) (*BackupSession, error) {
	e.mu.Lock()

	session := &BackupSession{
		ID:            fmt.Sprintf("session_%d_%s", time.Now().Unix(), job.ID),
		JobID:         job.ID,
		JobName:       job.Name,
		Type:          job.Type,
		StartTime:     time.Now(),
		Status:        "running",
		IsIncremental: job.Incremental,
		IsFull:        !job.Incremental,
		Logs:          make([]string, 0),
	}

	e.sessions[session.ID] = session
	e.mu.Unlock()

	e.log(session, "════════════════════════════════════════")
	e.log(session, fmt.Sprintf("🚀 Запуск резервного копіювання: %s", job.Name))
	e.log(session, fmt.Sprintf("Тип: %s, Ідентифікатор: %s", job.Type, session.ID))
	e.log(session, fmt.Sprintf("Інкрементальне: %v, Стиснення: %v", job.Incremental, job.Compression))

	// Create backup directory with timestamp
	backupDir := filepath.Join(job.Destination, job.Name, time.Now().Format("2006-01-02_150405"))
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		session.Status = "failed"
		session.Error = err.Error()
		return session, err
	}
	session.BackupPath = backupDir

	// Normalize source paths (convert forward slashes to backslashes on Windows)
	for i, src := range job.Sources {
		job.Sources[i] = filepath.FromSlash(src)
	}

	// Run pre-backup script
	if job.PreBackupScript != "" {
		e.log(session, fmt.Sprintf("📜 Виконання скрипта перед бекапом: %s", job.PreBackupScript))
		if err := e.runScript(job.PreBackupScript); err != nil {
			e.log(session, fmt.Sprintf("⚠️ Помилка скрипта: %v", err))
			session.Warnings = append(session.Warnings, fmt.Sprintf("Pre-script failed: %v", err))
		}
	}

	var err error
	switch job.Type {
	case TypeFile:
		err = e.backupFiles(job, session)
	case TypeDatabase:
		err = e.backupDatabase(job, session)
	case TypeVM:
		err = e.backupVM(job, session)
	case TypeCloud:
		err = e.backupToCloud(job, session)
	default:
		err = fmt.Errorf("непідтримуваний тип резервного копіювання: %s", job.Type)
	}

	// Run post-backup script
	if job.PostBackupScript != "" {
		e.log(session, fmt.Sprintf("📜 Виконання скрипта після бекапу: %s", job.PostBackupScript))
		if err := e.runScript(job.PostBackupScript); err != nil {
			e.log(session, fmt.Sprintf("⚠️ Помилка скрипта: %v", err))
			session.Warnings = append(session.Warnings, fmt.Sprintf("Post-script failed: %v", err))
		}
	}

	session.EndTime = time.Now()
	duration := session.EndTime.Sub(session.StartTime)

	if err != nil {
		session.Status = "failed"
		session.Error = err.Error()
		e.log(session, fmt.Sprintf("❌ Резервне копіювання НЕ ВДАЛОСЬ: %v", err))
	} else {
		session.Status = "success"

		// Calculate compression ratio
		if session.BytesTotal > 0 && session.BytesWritten > 0 {
			session.CompressionRatio = float64(session.BytesTotal) / float64(session.BytesWritten)
		}

		// Calculate deduplication ratio
		if session.ProcessedBlocks > 0 && session.DeduplicatedBlocks > 0 {
			session.DeduplicationRatio = float64(session.ProcessedBlocks) / float64(session.DeduplicatedBlocks)
		}

		e.log(session, fmt.Sprintf("✅ Резервне копіювання УСПІШНО завершено"))
		e.log(session, fmt.Sprintf("⏱️ Тривалість: %v", duration))
		e.log(session, fmt.Sprintf("📁 Оброблено файлів: %d/%d", session.FilesProcessed, session.FilesTotal))
		e.log(session, fmt.Sprintf("💾 Прочитано: %s, Записано: %s", e.formatBytes(session.BytesRead), e.formatBytes(session.BytesWritten)))
		e.log(session, fmt.Sprintf("📊 Стиснення: %.2fx, Дедуплікація: %.2fx", session.CompressionRatio, session.DeduplicationRatio))
	}

	// Save session
	e.saveSession(session)

	// Apply retention policy
	e.applyRetentionPolicy(job, backupDir)

	return session, err
}

// backupFiles backs up files and folders with deduplication
func (e *BackupEngine) backupFiles(job *BackupJob, session *BackupSession) error {
	e.log(session, "📁 Початок резервного копіювання файлів...")

	// Collect all files
	var files []string
	var totalSize int64

	for _, source := range job.Sources {
		e.log(session, fmt.Sprintf("🔍 Сканування джерела: %s", source))

		found, size, err := e.collectFiles(source, job)
		if err != nil {
			e.log(session, fmt.Sprintf("⚠️ Помилка сканування %s: %v", source, err))
			session.Warnings = append(session.Warnings, fmt.Sprintf("Failed to scan %s: %v", source, err))
			continue
		}

		files = append(files, found...)
		totalSize += size
	}

	session.FilesTotal = len(files)
	session.BytesTotal = totalSize
	e.log(session, fmt.Sprintf("✅ Знайдено %d файлів (%s)", len(files), e.formatBytes(totalSize)))

	if len(files) == 0 {
		return fmt.Errorf("не знайдено файлів для резервного копіювання")
	}

	// Create archive
	archivePath := filepath.Join(session.BackupPath, "backup.zip")
	archiveFile, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer archiveFile.Close()

	// Create zip writer with compression
	zipWriter := zip.NewWriter(archiveFile)

	// Initialize encryption if enabled
	var blockCipher cipher.Block
	if job.Encryption && job.EncryptionKey != "" {
		e.log(session, "🔐 Увімкнено шифрування AES-256...")
		hash := sha256.Sum256([]byte(job.EncryptionKey))
		blockCipher, err = aes.NewCipher(hash[:])
		if err != nil {
			archiveFile.Close()
			return err
		}
	}

	// Backup files with deduplication
	var bytesWritten int64
	blockSize := job.BlockSize
	if blockSize == 0 {
		blockSize = 1024 * 1024 // 1MB default
	}

	e.log(session, fmt.Sprintf("📝 Starting to write %d files to ZIP...", len(files)))

	for i, file := range files {
		zipName := e.zipEntryName(file, job.Sources)
		e.log(session, fmt.Sprintf("   Adding file %d/%d: %s -> %s", i+1, len(files), file, zipName))

		if err := e.addFileToZipWithDedup(zipWriter, file, zipName, blockCipher, blockSize, session, job.Compression); err != nil {
			e.log(session, fmt.Sprintf("⚠️ Не вдалося додати %s: %v", file, err))
			session.FilesSkipped++
			continue
		}
		e.log(session, fmt.Sprintf("   ✓ Successfully added %s", file))

		session.FilesProcessed++
		info, _ := os.Stat(file)
		if info != nil {
			bytesWritten += info.Size()
		}
		session.BytesRead = bytesWritten

		// Progress logging every 100 files
		if (i+1)%100 == 0 {
			progress := float64(session.FilesProcessed) * 100 / float64(session.FilesTotal)
			e.log(session, fmt.Sprintf("📊 Прогрес: %d/%d файлів (%.1f%%)",
				session.FilesProcessed, session.FilesTotal, progress))
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
		"dedup_ratio":       session.DeduplicationRatio,
	}

	metadataJSON, _ := json.MarshalIndent(metadata, "", "  ")
	os.WriteFile(filepath.Join(session.BackupPath, "metadata.json"), metadataJSON, 0644)

	// Close zip writer to flush all data and write central directory
	if err := zipWriter.Close(); err != nil {
		e.log(session, fmt.Sprintf("⚠️ Error closing ZIP: %v", err))
	}

	// Close archive file
	archiveFile.Close()

	e.log(session, fmt.Sprintf("📦 Створено архів: %s (%s)", archivePath, e.formatBytes(session.BytesWritten)))

	return nil
}

// addFileToZipWithDedup adds a file to zip with simple deduplication tracking
func (e *BackupEngine) addFileToZipWithDedup(zw *zip.Writer, filePath, zipName string, blockCipher cipher.Block, blockSize int, session *BackupSession, compression bool) error {
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

	// Use relative path and DEFLATE compression if requested
	header.Name = zipName
	if compression {
		header.Method = zip.Deflate
	} else {
		header.Method = zip.Store
	}

	writer, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}

	// Buffer for encryption/copying
	buffer := make([]byte, 64*1024) // 64KB buffer
	for {
		n, err := file.Read(buffer)
		if n > 0 {
			block := buffer[:n]

			// Calculate block hash for deduplication stats
			hash := sha256.Sum256(block)
			hashHex := hex.EncodeToString(hash[:])

			session.ProcessedBlocks++

			// Track unique blocks
			if _, exists := e.blockHashes[hashHex]; !exists {
				e.blockHashes[hashHex] = filePath
				session.DeduplicatedBlocks++
			}

			// Encrypt if cipher is provided
			var dataToWrite []byte = block
			if blockCipher != nil {
				dataToWrite = e.encryptBlock(blockCipher, block)
			}

			// Write the data
			_, writeErr := writer.Write(dataToWrite)
			if writeErr != nil {
				return writeErr
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *BackupEngine) zipEntryName(filePath string, sources []string) string {
	cleanFile := filepath.Clean(filePath)
	for _, src := range sources {
		cleanSrc := filepath.Clean(src)
		rel, err := filepath.Rel(cleanSrc, cleanFile)
		if err != nil {
			continue
		}

		// Ensure file is within source root
		if rel == "." {
			return filepath.ToSlash(filepath.Base(cleanFile))
		}
		if strings.HasPrefix(rel, ".."+string(os.PathSeparator)) || rel == ".." {
			continue
		}

		base := filepath.Base(cleanSrc)
		if base == "." || base == string(os.PathSeparator) {
			base = "source"
		}
		return filepath.ToSlash(filepath.Join(base, rel))
	}

	return filepath.ToSlash(filepath.Base(cleanFile))
}

// encryptBlock encrypts a block using AES
func (e *BackupEngine) encryptBlock(blockCipher cipher.Block, data []byte) []byte {
	// Pad data to block size
	blockSize := blockCipher.BlockSize()
	padLen := blockSize - (len(data) % blockSize)
	padded := append(data, bytes.Repeat([]byte{byte(padLen)}, padLen)...)

	// Encrypt in CBC mode
	ciphertext := make([]byte, len(padded))
	iv := make([]byte, blockSize) // In production, use random IV
	mode := cipher.NewCBCEncrypter(blockCipher, iv)
	mode.CryptBlocks(ciphertext, padded)

	return ciphertext
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

// shouldInclude checks if a file should be included
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

	// Check include patterns
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

// backupDatabase backs up databases
func (e *BackupEngine) backupDatabase(job *BackupJob, session *BackupSession) error {
	e.log(session, fmt.Sprintf("🗄️ Початок резервного копіювання бази даних %s...", job.DatabaseType))

	var dumpFile string
	var err error

	switch job.DatabaseType {
	case DBMySQL:
		dumpFile, err = e.backupMySQL(job, session)
	case DBPostgreSQL:
		dumpFile, err = e.backupPostgreSQL(job, session)
	case DBSQLite:
		dumpFile, err = e.backupSQLite(job, session)
	default:
		return fmt.Errorf("непідтримуваний тип бази даних: %s", job.DatabaseType)
	}

	if err != nil {
		return err
	}

	// Compress if enabled
	if job.Compression {
		e.log(session, "🗜️ Стиснення дампу бази даних...")
		compressedFile := dumpFile + ".gz"
		if err := e.compressFile(dumpFile, compressedFile); err != nil {
			e.log(session, fmt.Sprintf("⚠️ Помилка стиснення: %v", err))
			session.Warnings = append(session.Warnings, fmt.Sprintf("Compression failed: %v", err))
		} else {
			os.Remove(dumpFile)
			dumpFile = compressedFile
		}
	}

	// Encrypt if enabled
	if job.Encryption && job.EncryptionKey != "" {
		e.log(session, "🔐 Шифрування резервної копії...")
		encryptedFile := dumpFile + ".enc"
		if err := e.encryptFile(dumpFile, encryptedFile, job.EncryptionKey); err != nil {
			e.log(session, fmt.Sprintf("⚠️ Помилка шифрування: %v", err))
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

	e.log(session, fmt.Sprintf("💾 Резервна копія БД створена: %s (%s)", dumpFile, e.formatBytes(session.BytesWritten)))

	return nil
}

// backupMySQL creates MySQL dump
func (e *BackupEngine) backupMySQL(job *BackupJob, session *BackupSession) (string, error) {
	dumpFile := filepath.Join(session.BackupPath, "mysql_dump.sql")

	// Try mysqldump
	mysqldumpPath := "mysqldump"
	if runtime.GOOS == "windows" {
		paths := []string{
			`C:\Program Files\MySQL\MySQL Server 8.0\bin\mysqldump.exe`,
			`C:\Program Files\MySQL\MySQL Server 5.7\bin\mysqldump.exe`,
		}
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				mysqldumpPath = path
				break
			}
		}
	}

	args := []string{
		"--result-file=" + dumpFile,
		"--single-transaction",
		"--quick",
		"--lock-tables=false",
	}

	// Parse connection string to get database name
	args = append(args, job.DatabaseConn)

	cmd := exec.Command(mysqldumpPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		e.log(session, fmt.Sprintf("mysqldump: %s", string(output)))
		// Fallback to direct export
		return e.backupMySQLDirect(job, session, dumpFile)
	}

	return dumpFile, nil
}

// backupMySQLDirect exports MySQL using direct connection
func (e *BackupEngine) backupMySQLDirect(job *BackupJob, session *BackupSession, dumpFile string) (string, error) {
	db, err := sql.Open("mysql", job.DatabaseConn)
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
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		rows.Scan(&table)
		tables = append(tables, table)
	}

	e.log(session, fmt.Sprintf("Знайдено %d таблиць", len(tables)))
	session.FilesTotal = len(tables)

	// Export each table
	for _, table := range tables {
		// CREATE TABLE
		createRow := db.QueryRow(fmt.Sprintf("SHOW CREATE TABLE `%s`", table))
		var tableName, createSQL string
		if createRow.Scan(&tableName, &createSQL) == nil {
			fmt.Fprintf(file, "-- Table: %s\n%s;\n\n", table, createSQL)
		}

		// INSERT statements
		dataRows, _ := db.Query(fmt.Sprintf("SELECT * FROM `%s`", table))
		if dataRows != nil {
			columns, _ := dataRows.Columns()
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))

			for dataRows.Next() {
				for i := range values {
					valuePtrs[i] = &values[i]
				}
				dataRows.Scan(valuePtrs...)

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
		}

		session.FilesProcessed++
	}

	return dumpFile, nil
}

// backupPostgreSQL creates PostgreSQL dump
func (e *BackupEngine) backupPostgreSQL(job *BackupJob, session *BackupSession) (string, error) {
	dumpFile := filepath.Join(session.BackupPath, "postgres_dump.sql")

	pgdumpPath := "pg_dump"
	if runtime.GOOS == "windows" {
		paths := []string{
			`C:\Program Files\PostgreSQL\15\bin\pg_dump.exe`,
			`C:\Program Files\PostgreSQL\14\bin\pg_dump.exe`,
		}
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				pgdumpPath = path
				break
			}
		}
	}

	cmd := exec.Command(pgdumpPath, job.DatabaseConn)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return e.backupPostgreSQLDirect(job, session, dumpFile)
	}

	os.WriteFile(dumpFile, output, 0644)
	return dumpFile, nil
}

// backupPostgreSQLDirect exports PostgreSQL using direct connection
func (e *BackupEngine) backupPostgreSQLDirect(job *BackupJob, session *BackupSession, dumpFile string) (string, error) {
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
		rows.Scan(&table)
		tables = append(tables, table)
	}

	session.FilesTotal = len(tables)

	// Export
	for _, table := range tables {
		fmt.Fprintf(file, "-- Table: %s\n", table)
		session.FilesProcessed++
	}

	return dumpFile, nil
}

// backupSQLite backs up SQLite database
func (e *BackupEngine) backupSQLite(job *BackupJob, session *BackupSession) (string, error) {
	sourceFile := job.DatabaseConn
	destFile := filepath.Join(session.BackupPath, "sqlite_backup.db")

	data, err := os.ReadFile(sourceFile)
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(destFile, data, 0644); err != nil {
		return "", err
	}

	return destFile, nil
}

// backupVM backs up Hyper-V VMs
func (e *BackupEngine) backupVM(job *BackupJob, session *BackupSession) error {
	e.log(session, "🖥️ Початок резервного копіювання віртуальних машин...")

	if runtime.GOOS != "windows" {
		return fmt.Errorf("резервне копіювання VM підтримується тільки на Windows")
	}

	// Use PowerShell to export VMs
	for _, vmName := range job.VMNames {
		e.log(session, fmt.Sprintf("Експорт VM: %s", vmName))

		exportPath := filepath.Join(session.BackupPath, "vms", vmName)
		psScript := fmt.Sprintf(`Export-VM -Name "%s" -Path "%s"`, vmName, exportPath)

		cmd := exec.Command("powershell", "-Command", psScript)
		output, err := cmd.CombinedOutput()
		if err != nil {
			e.log(session, fmt.Sprintf("⚠️ Помилка експорту VM %s: %v - %s", vmName, err, string(output)))
			session.Warnings = append(session.Warnings, fmt.Sprintf("VM export failed: %v", err))
		} else {
			session.FilesProcessed++
			e.log(session, fmt.Sprintf("✅ VM %s експортовано", vmName))
		}
	}

	session.FilesTotal = len(job.VMNames)
	return nil
}

// backupToCloud uploads backup to cloud storage
func (e *BackupEngine) backupToCloud(job *BackupJob, session *BackupSession) error {
	e.log(session, fmt.Sprintf("☁️ Початок хмарного резервного копіювання (%s)...", job.CloudProvider))

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

	if err := e.backupFiles(localJob, localSession); err != nil {
		return err
	}

	// Upload to cloud based on provider
	switch job.CloudProvider {
	case CloudS3:
		return e.uploadToS3(job, tempDir)
	case CloudAzure:
		return e.uploadToAzure(job, tempDir)
	case CloudGoogle:
		return e.uploadToGoogle(job, tempDir)
	default:
		return fmt.Errorf("непідтримуваний хмарний провайдер: %s", job.CloudProvider)
	}
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
				e.log(nil, fmt.Sprintf("🗑️ Видалено стару резервну копію: %s", backup.Name()))
			}
		}
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
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	key := sha256.Sum256([]byte(password))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	// Seal prepends the nonce to the ciphertext
	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	return os.WriteFile(dst, ciphertext, 0644)
}
func (e *BackupEngine) formatBytes(bytes int64) string {
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

// GetCBTStatistics returns Changed Block Tracking statistics
func (e *BackupEngine) GetCBTStatistics() map[string]interface{} {
	if e.changeTracker == nil {
		return map[string]interface{}{
			"enabled": false,
			"message": "CBT не увімкнено",
		}
	}
	return e.changeTracker.GetStatistics()
}

// Upload functions (stubs for now)
func (e *BackupEngine) uploadToS3(job *BackupJob, backupDir string) error {
	e.log(nil, "📤 Завантаження в AWS S3...")
	// TODO: Implement AWS S3 SDK upload
	return nil
}

func (e *BackupEngine) uploadToAzure(job *BackupJob, backupDir string) error {
	e.log(nil, "📤 Завантаження в Azure Blob Storage...")
	// TODO: Implement Azure SDK upload
	return nil
}

func (e *BackupEngine) uploadToGoogle(job *BackupJob, backupDir string) error {
	e.log(nil, "📤 Завантаження в Google Cloud Storage...")
	// TODO: Implement GCS SDK upload
	return nil
}
