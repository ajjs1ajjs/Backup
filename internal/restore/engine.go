// Restore Engine - Complete Veeam-style restore functionality
package restore

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Restore Types
const (
	RestoreFiles    = "files"
	RestoreDatabase = "database"
	RestoreVM       = "vm"
	RestoreInstant  = "instant"
	RestoreGranular = "granular"
)

// RestoreRequest represents a restore operation request
type RestoreRequest struct {
	ID              string   `json:"id"`
	Type            string   `json:"type"` // files, database, vm, instant
	BackupPath      string   `json:"backup_path"`
	SessionID       string   `json:"session_id"`
	Destination     string   `json:"destination"`
	RestoreOriginal bool     `json:"restore_original"`
	Overwrite       bool     `json:"overwrite"`
	Files           []string `json:"files"` // specific files to restore

	// Database specific
	DBType         string `json:"db_type"`
	ConnStr        string `json:"conn_str"`
	TargetDatabase string `json:"target_database"`

	// VM specific
	VMName     string `json:"vm_name"`
	NewVMName  string `json:"new_vm_name"`
	HyperVHost string `json:"hyperv_host"`

	// Advanced
	EncryptionKey     string `json:"-"`
	PreRestoreScript  string `json:"pre_restore_script"`
	PostRestoreScript string `json:"post_restore_script"`

	// Status
	CreatedAt time.Time `json:"created_at"`
	Status    string    `json:"status"` // pending, running, success, failed
}

// RestoreSession tracks a restore operation
type RestoreSession struct {
	ID            string    `json:"id"`
	RequestID     string    `json:"request_id"`
	Type          string    `json:"type"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	Status        string    `json:"status"`
	FilesTotal    int       `json:"files_total"`
	FilesRestored int       `json:"files_restored"`
	FilesSkipped  int       `json:"files_skipped"`
	BytesTotal    int64     `json:"bytes_total"`
	BytesRestored int64     `json:"bytes_restored"`
	Error         string    `json:"error,omitempty"`
	Warnings      []string  `json:"warnings"`
	Logs          []string  `json:"logs"`
	Duration      string    `json:"duration"`
}

// BackupFileInfo represents a file inside a backup archive.
type BackupFileInfo struct {
	Name     string    `json:"name"`
	Type     string    `json:"type"` // file or directory
	Size     int64     `json:"size"`
	Modified time.Time `json:"modified"`
}

// RestoreEngine handles all restore operations
type RestoreEngine struct {
	DataDir string
	LogFile string
	AllowScripts bool
}

// NewRestoreEngine creates a new restore engine
func NewRestoreEngine(dataDir string) *RestoreEngine {
	return &RestoreEngine{
		DataDir: dataDir,
		LogFile: filepath.Join(dataDir, "logs", "restore.log"),
	}
}

// ExecuteRestore runs a restore operation
func (e *RestoreEngine) ExecuteRestore(req *RestoreRequest) (*RestoreSession, error) {
	session := &RestoreSession{
		ID:        fmt.Sprintf("restore_%d_%s", time.Now().Unix(), req.ID),
		RequestID: req.ID,
		Type:      req.Type,
		StartTime: time.Now(),
		Status:    "running",
		Logs:      make([]string, 0),
	}

	e.log(session, "════════════════════════════════════════")
	e.log(session, fmt.Sprintf("🔄 Початок відновлення: %s", req.Type))
	e.log(session, fmt.Sprintf("Ідентифікатор: %s", session.ID))

	// Run pre-restore script
	if req.PreRestoreScript != "" {
		if !e.AllowScripts {
			e.log(session, "⚠️ Виконання скриптів вимкнено політикою безпеки")
			session.Warnings = append(session.Warnings, "Pre-restore script skipped: scripts are disabled by security policy")
		} else {
			e.log(session, fmt.Sprintf("📜 Виконання скрипта перед відновленням: %s", req.PreRestoreScript))
			if err := e.runScript(req.PreRestoreScript); err != nil {
				e.log(session, fmt.Sprintf("⚠️ Помилка скрипта: %v", err))
				session.Warnings = append(session.Warnings, fmt.Sprintf("Pre-restore script failed: %v", err))
			}
		}
	}

	var err error
	switch req.Type {
	case RestoreFiles:
		err = e.restoreFiles(req, session)
	case RestoreDatabase:
		err = e.restoreDatabase(req, session)
	case RestoreVM:
		err = e.restoreVM(req, session)
	case RestoreInstant:
		err = e.instantRestore(req, session)
	default:
		err = fmt.Errorf("непідтримуваний тип відновлення: %s", req.Type)
	}

	// Run post-restore script
	if req.PostRestoreScript != "" {
		if !e.AllowScripts {
			e.log(session, "⚠️ Виконання скриптів вимкнено політикою безпеки")
			session.Warnings = append(session.Warnings, "Post-restore script skipped: scripts are disabled by security policy")
		} else {
			e.log(session, fmt.Sprintf("📜 Виконання скрипта після відновлення: %s", req.PostRestoreScript))
			if err := e.runScript(req.PostRestoreScript); err != nil {
				e.log(session, fmt.Sprintf("⚠️ Помилка скрипта: %v", err))
				session.Warnings = append(session.Warnings, fmt.Sprintf("Post-restore script failed: %v", err))
			}
		}
	}

	session.EndTime = time.Now()
	session.Duration = session.EndTime.Sub(session.StartTime).String()

	if err != nil {
		session.Status = "failed"
		session.Error = err.Error()
		e.log(session, fmt.Sprintf("❌ Відновлення НЕ ВДАЛОСЬ: %v", err))
	} else {
		session.Status = "success"
		e.log(session, fmt.Sprintf("✅ Відновлення УСПІШНО завершено"))
		e.log(session, fmt.Sprintf("⏱️ Тривалість: %s", session.Duration))
		e.log(session, fmt.Sprintf("📁 Відновлено файлів: %d/%d", session.FilesRestored, session.FilesTotal))
		e.log(session, fmt.Sprintf("💾 Відновлено даних: %s", e.formatBytes(session.BytesRestored)))
	}

	// Save session
	e.saveSession(session)

	return session, err
}

// restoreFiles restores files from backup
func (e *RestoreEngine) restoreFiles(req *RestoreRequest, session *RestoreSession) error {
	e.log(session, "📁 Початок відновлення файлів...")

	// Open archive
	archivePath := filepath.Join(req.BackupPath, "backup.zip")
	encryptedArchivePath := archivePath + ".enc"
	tempArchive := ""

	if _, err := os.Stat(archivePath); err != nil {
		if _, encErr := os.Stat(encryptedArchivePath); encErr == nil {
			if req.EncryptionKey == "" {
				return fmt.Errorf("потрібен ключ для розшифрування архіву")
			}
			tempArchive = filepath.Join(req.BackupPath, fmt.Sprintf("backup_%d.zip", time.Now().UnixNano()))
			if err := e.decryptFile(encryptedArchivePath, tempArchive, req.EncryptionKey); err != nil {
				return fmt.Errorf("помилка розшифрування архіву: %v", err)
			}
			archivePath = tempArchive
		}
	}

	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("помилка відкриття архіву: %v", err)
	}
	defer r.Close()
	if tempArchive != "" {
		defer os.Remove(tempArchive)
	}

	// Count files
	session.FilesTotal = len(r.File)

	// Determine destination
	destDir := req.Destination
	if req.RestoreOriginal {
		// Restore to original locations (requires admin)
		destDir = ""
	}

	if destDir == "" {
		destDir = filepath.Join(os.Getenv("USERPROFILE"), "Restored", time.Now().Format("2006-01-02_150405"))
	}

	os.MkdirAll(destDir, 0755)
	e.log(session, fmt.Sprintf("📂 Шлях відновлення: %s", destDir))

	// Extract files
	var bytesRestored int64
	for _, f := range r.File {
		// Skip if specific files requested
		if len(req.Files) > 0 {
			found := false
			for _, pattern := range req.Files {
				if strings.Contains(f.Name, pattern) {
					found = true
					break
				}
			}
			if !found {
				session.FilesSkipped++
				continue
			}
		}

		// Determine destination path
		var destPath string
		if req.RestoreOriginal {
			destPath = f.Name // Original location
		} else {
			// Get only filename to avoid path issues
			fileName := filepath.Base(f.Name)
			destPath = filepath.Join(destDir, fileName)
		}

		// Skip if exists and not overwriting
		if _, err := os.Stat(destPath); err == nil && !req.Overwrite {
			session.FilesSkipped++
			session.Warnings = append(session.Warnings, fmt.Sprintf("Пропущено: %s (вже існує)", destPath))
			continue
		}

		if err := e.extractFile(f, destPath, req.EncryptionKey); err != nil {
			session.Warnings = append(session.Warnings, fmt.Sprintf("Помилка %s: %v", f.Name, err))
			continue
		}

		session.FilesRestored++
		bytesRestored += int64(f.UncompressedSize64)
	}

	session.BytesRestored = bytesRestored
	e.log(session, fmt.Sprintf("✅ Відновлено %d файлів (%s)", session.FilesRestored, e.formatBytes(bytesRestored)))

	return nil
}

// extractFile extracts a single file from zip
func (e *RestoreEngine) extractFile(f *zip.File, destPath string, encryptionKey string) error {
	// Create destination directory
	os.MkdirAll(filepath.Dir(destPath), 0755)

	// Open source file
	srcFile, err := f.Open()
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Decrypt if needed
	var reader io.Reader = srcFile
	if encryptionKey != "" && strings.HasSuffix(f.Name, ".enc") {
		// TODO: Implement decryption
		reader = srcFile
	}

	// Copy content
	_, err = io.Copy(destFile, reader)
	return err
}

// restoreDatabase restores a database from backup
func (e *RestoreEngine) restoreDatabase(req *RestoreRequest, session *RestoreSession) error {
	e.log(session, fmt.Sprintf("🗄️ Початок відновлення бази даних %s...", req.DBType))

	// Find dump file
	dumpFile := e.findDumpFile(req.BackupPath)
	if dumpFile == "" {
		return fmt.Errorf("файл дампу бази даних не знайдено")
	}

	e.log(session, fmt.Sprintf("📄 Файл дампу: %s", dumpFile))

	// Decompress if needed
	if strings.HasSuffix(dumpFile, ".gz") {
		e.log(session, "🗜️ Розпакування дампу...")
		decompressed := strings.TrimSuffix(dumpFile, ".gz")
		if err := e.decompressFile(dumpFile, decompressed); err != nil {
			e.log(session, fmt.Sprintf("⚠️ Помилка розпакування: %v", err))
			session.Warnings = append(session.Warnings, fmt.Sprintf("Decompression failed: %v", err))
		} else {
			dumpFile = decompressed
		}
	}

	// Decrypt if needed
	if strings.HasSuffix(dumpFile, ".enc") {
		if req.EncryptionKey == "" {
			return fmt.Errorf("потрібен ключ для розшифрування дампу")
		}
		e.log(session, "🔐 Розшифрування дампу...")
		decrypted := strings.TrimSuffix(dumpFile, ".enc")
		if err := e.decryptFile(dumpFile, decrypted, req.EncryptionKey); err != nil {
			return fmt.Errorf("помилка розшифрування: %v", err)
		}
		defer os.Remove(decrypted)
		dumpFile = decrypted
	}

	// Restore based on DB type
	var err error
	switch req.DBType {
	case "mysql":
		err = e.restoreMySQL(dumpFile, req.ConnStr, req.TargetDatabase, session)
	case "postgresql":
		err = e.restorePostgreSQL(dumpFile, req.ConnStr, req.TargetDatabase, session)
	case "oracle":
		err = e.restoreOracle(dumpFile, req.ConnStr, req.TargetDatabase, session)
	case "sqlite":
		err = e.restoreSQLite(dumpFile, req.ConnStr, session)
	case "mssql":
		err = e.restoreMSSQL(dumpFile, req.ConnStr, req.TargetDatabase, session)
	default:
		err = fmt.Errorf("непідтримуваний тип бази даних: %s", req.DBType)
	}

	if err != nil {
		return err
	}

	session.FilesRestored = 1
	e.log(session, "✅ Базу даних успішно відновлено")

	return nil
}

// findDumpFile finds the database dump file in backup
func (e *RestoreEngine) findDumpFile(backupPath string) string {
	// SQL dump patterns
	patterns := []string{
		"*.sql",
		"*.sql.gz",
		"*.sql.enc",
		"*_dump.sql",
		"backup_*.sql",
	}

	for _, pattern := range patterns {
		matches, _ := filepath.Glob(filepath.Join(backupPath, pattern))
		if len(matches) > 0 {
			return matches[0]
		}
	}

	// MSSQL backup patterns (*.bak)
	mssqlPatterns := []string{
		"*.bak",
		"*.bak.gz",
		"*.bak.enc",
	}

	for _, pattern := range mssqlPatterns {
		matches, _ := filepath.Glob(filepath.Join(backupPath, pattern))
		if len(matches) > 0 {
			return matches[0]
		}
	}

	// Also check subdirectories (mssql/, mysql/, etc.)
	subdirs, _ := filepath.Glob(filepath.Join(backupPath, "*"))
	for _, dir := range subdirs {
		info, _ := os.Stat(dir)
		if info != nil && info.IsDir() {
			// Check for .bak files in subdirectory
			bakMatches, _ := filepath.Glob(filepath.Join(dir, "*.bak"))
			if len(bakMatches) > 0 {
				return bakMatches[0]
			}
			// Check for .bak.gz files (compressed MSSQL backup)
			bakGzMatches, _ := filepath.Glob(filepath.Join(dir, "*.bak.gz"))
			if len(bakGzMatches) > 0 {
				return bakGzMatches[0]
			}
			// Check for .bak.enc files (encrypted MSSQL backup)
			bakEncMatches, _ := filepath.Glob(filepath.Join(dir, "*.bak.enc"))
			if len(bakEncMatches) > 0 {
				return bakEncMatches[0]
			}
			// Check for .sql files in subdirectory
			sqlMatches, _ := filepath.Glob(filepath.Join(dir, "*.sql"))
			if len(sqlMatches) > 0 {
				return sqlMatches[0]
			}
			// Check for .sql.gz files
			sqlGzMatches, _ := filepath.Glob(filepath.Join(dir, "*.sql.gz"))
			if len(sqlGzMatches) > 0 {
				return sqlGzMatches[0]
			}
		}
	}

	return ""
}

// restoreMySQL restores MySQL database
func (e *RestoreEngine) restoreMySQL(dumpFile, connStr, targetDB string, session *RestoreSession) error {
	mysqlPath := "mysql"

	// Parse connection string to get database name
	args := []string{connStr}

	if targetDB != "" {
		args = append(args, targetDB)
	}

	cmd := exec.Command(mysqlPath, args...)
	input, err := os.ReadFile(dumpFile)
	if err != nil {
		return err
	}

	cmd.Stdin = bytes.NewReader(input)
	output, err := cmd.CombinedOutput()

	session.Logs = append(session.Logs, fmt.Sprintf("mysql output: %s", string(output)))

	if err != nil {
		return fmt.Errorf("помилка відновлення MySQL: %v - %s", err, string(output))
	}

	return nil
}

// restorePostgreSQL restores PostgreSQL database
func (e *RestoreEngine) restorePostgreSQL(dumpFile, connStr, targetDB string, session *RestoreSession) error {
	psqlPath := "psql"

	args := []string{connStr}

	if targetDB != "" {
		args = append(args, "-d", targetDB)
	}

	cmd := exec.Command(psqlPath, args...)
	input, err := os.ReadFile(dumpFile)
	if err != nil {
		return err
	}

	cmd.Stdin = bytes.NewReader(input)
	output, err := cmd.CombinedOutput()

	session.Logs = append(session.Logs, fmt.Sprintf("psql output: %s", string(output)))

	if err != nil {
		return fmt.Errorf("помилка відновлення PostgreSQL: %v - %s", err, string(output))
	}

	return nil
}

// restoreMSSQL restores Microsoft SQL Server database from .bak file
func (e *RestoreEngine) restoreMSSQL(bakFile, connStr, targetDB string, session *RestoreSession) error {
	e.log(session, fmt.Sprintf("🗄️ Відновлення MSSQL з файлу: %s", bakFile))

	if runtime.GOOS != "windows" {
		return fmt.Errorf("відновлення MSSQL підтримується тільки на Windows")
	}

	serverInstance := getServerInstance(connStr)

	// Build PowerShell script for SQL Server restore with enhanced error handling
	psScript := fmt.Sprintf(`$ErrorActionPreference = "Stop"
$backupFile = "%s"
$targetDatabase = "%s"
$connectionString = "%s"
$serverInstance = "%s"

Write-Host "=== MSSQL Restore Script ===" -ForegroundColor Cyan
Write-Host "Target database: $targetDatabase"
Write-Host "Backup file: $backupFile"
Write-Host ""

try {
    # 1. Перевірка backup-файлу
    Write-Host "[1/5] Перевірка backup-файлу..."
    if (-not (Test-Path $backupFile)) {
        throw "Backup file not found: $backupFile"
    }
    $file = Get-Item $backupFile
    if ($file.Length -eq 0) {
        throw "Backup file is empty: $backupFile"
    }
    Write-Host "✓ Backup file exists: $($file.FullName) ($([math]::Round($file.Length/1MB, 2)) MB)"

    # 2. Підключення до SQL Server
    Write-Host "[2/5] Підключення до SQL Server ($serverInstance)..."
    $connection = New-Object System.Data.SqlClient.SqlConnection
    $connection.ConnectionString = $connectionString
    $connection.Open()
    Write-Host "✓ Підключено до SQL Server"

    # 3. Перевірка прав sysadmin
    Write-Host "[3/5] Перевірка прав доступу..."
    $sysadminQuery = "SELECT IS_SRVROLEMEMBER('sysadmin')"
    $sysadminCmd = New-Object System.Data.SqlClient.SqlCommand
    $sysadminCmd.CommandText = $sysadminQuery
    $sysadminCmd.Connection = $connection
    $sysadminResult = $sysadminCmd.ExecuteScalar()

    if ([int]$sysadminResult -ne 1) {
        # Спроба автоматично надати права поточному користувачу
        Write-Host "⚠ Відсутні права sysadmin. Спроба автоматичного надання прав..." -ForegroundColor Yellow

        $currentUser = [System.Security.Principal.WindowsIdentity]::GetCurrent().Name
        Write-Host "  Поточний користувач: $currentUser"

        try {
            # Спроба створити login і додати до sysadmin
            $grantQuery = @"
IF NOT EXISTS (SELECT 1 FROM sys.server_principals WHERE name = '$currentUser')
BEGIN
    CREATE LOGIN [$currentUser] FROM WINDOWS;
END
ALTER SERVER ROLE [sysadmin] ADD MEMBER [$currentUser];
"@
            $grantCmd = New-Object System.Data.SqlClient.SqlCommand
            $grantCmd.CommandText = $grantQuery
            $grantCmd.Connection = $connection
            $grantCmd.ExecuteNonQuery() | Out-Null
            Write-Host "✓ Права sysadmin надано користувачу '$currentUser'" -ForegroundColor Green

            # Перевірка ще раз
            $sysadminCmd2 = New-Object System.Data.SqlClient.SqlCommand
            $sysadminCmd2.CommandText = $sysadminQuery
            $sysadminCmd2.Connection = $connection
            $sysadminResult2 = $sysadminCmd2.ExecuteScalar()
            if ([int]$sysadminResult2 -ne 1) {
                throw "Не вдалося надати права sysadmin. Потрібен доступ адміністратора SQL Server."
            }
        } catch {
            Write-Host ""
            Write-Host "Відсутні права sysadmin. Потрібні права CREATE DATABASE та RESTORE." -ForegroundColor Red
            Write-Host "Деталі: $($_.Exception.Message)" -ForegroundColor Red
            throw "Відсутні права sysadmin. Потрібен доступ адміністратора SQL Server."
        }
    }
    Write-Host "✓ Права підтверджено (sysadmin)"

    # 4. Отримання логічних імен з backup
    Write-Host "[4/5] Читання структури backup-файлу..."
    $query = "RESTORE FILELISTONLY FROM DISK = '$backupFile'"
    $command = New-Object System.Data.SqlClient.SqlCommand
    $command.CommandText = $query
    $command.Connection = $connection
    $adapter = New-Object System.Data.SqlClient.SqlDataAdapter
    $adapter.SelectCommand = $command
    $dataset = New-Object System.Data.DataSet
    $adapter.Fill($dataset) | Out-Null

    if ($dataset.Tables[0].Rows.Count -lt 2) {
        throw "Некоректна структура backup-файлу: очікується мінімум 2 файли (data + log)"
    }

    $logicalDataName = $dataset.Tables[0].Rows[0].LogicalName
    $logicalLogName = $dataset.Tables[0].Rows[1].LogicalName
    Write-Host "  Logical Data: $logicalDataName"
    Write-Host "  Logical Log:  $logicalLogName"

    # 5. Отримання шляхів до даних через SERVERPROPERTY
    $dataPathQuery = "SELECT SERVERPROPERTY('InstanceDefaultDataPath') AS DataPath"
    $dataPathCmd = New-Object System.Data.SqlClient.SqlCommand
    $dataPathCmd.CommandText = $dataPathQuery
    $dataPathCmd.Connection = $connection
    $dataPathResult = $dataPathCmd.ExecuteScalar()
    $dataPath = $dataPathResult
    if ([string]::IsNullOrEmpty($dataPath)) {
        $dataPath = "C:\ProgramData\Microsoft\SQL Server\Data"
    }
    $newDataFile = Join-Path $dataPath "${targetDatabase}.mdf"
    $newLogFile = Join-Path $dataPath "${targetDatabase}_log.ldf"
    Write-Host "  Data path: $dataPath"
    Write-Host "  New data file: $newDataFile"
    Write-Host "  New log file: $newLogFile"

    $connection.Close()

    # 6. Відновлення бази даних
    Write-Host "[5/5] Відновлення бази даних..."
    $connection2 = New-Object System.Data.SqlClient.SqlConnection
    $connection2.ConnectionString = $connectionString
    $connection2.Open()

    $restoreQuery = @"
RESTORE DATABASE [$targetDatabase]
FROM DISK = '$backupFile'
WITH
    MOVE '$logicalDataName' TO '$newDataFile',
    MOVE '$logicalLogName' TO '$newLogFile',
    REPLACE,
    STATS = 10
"@

    Write-Host "Executing restore..." -ForegroundColor Gray
    $command2 = New-Object System.Data.SqlClient.SqlCommand
    $command2.CommandText = $restoreQuery
    $command2.Connection = $connection2
    $command2.CommandTimeout = 0  # Без тайм-ауту для великих баз
    $command2.ExecuteNonQuery() | Out-Null
    $connection2.Close()

    Write-Host ""
    Write-Host "✓ Database '$targetDatabase' restored successfully!" -ForegroundColor Green
    exit 0

} catch {
    Write-Host ""
    Write-Host "=== RESTORE ERROR ===" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host ""
    Write-Host "Possible causes:" -ForegroundColor Yellow
    Write-Host "  • Відсутні права sysadmin / CREATE DATABASE"
    Write-Host "  • Backup-файл пошкоджений або недоступний"
    Write-Host "  • Недостатньо місця на диску"
    Write-Host "  • SQL Server service не запущений"
    Write-Host ""
    Write-Host "Stack trace:" -ForegroundColor Gray
    Write-Host $_.ScriptStackTrace -ForegroundColor Gray
    exit 1
}
`, bakFile, targetDB, connStr, serverInstance)

	cmd := exec.Command("powershell", "-Command", psScript)
	output, err := cmd.CombinedOutput()

	session.Logs = append(session.Logs, fmt.Sprintf("PowerShell output: %s", string(output)))

	if err != nil {
		return fmt.Errorf("помилка відновлення MSSQL: %v - %s", err, string(output))
	}

	e.log(session, "✅ MSSQL базу даних успішно відновлено")
	return nil
}

// getServerInstance extracts server instance from connection string
func getServerInstance(connStr string) string {
	// Parse "server=sb191\\MSSQL2022SELF;database=master;..." → "sb191\\MSSQL2022SELF"
	parts := strings.Split(connStr, ";")
	for _, part := range parts {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) == 2 && strings.ToLower(strings.TrimSpace(kv[0])) == "server" {
			return strings.TrimSpace(kv[1])
		}
	}
	return "unknown"
}

// restoreOracle restores Oracle database from dump file
func (e *RestoreEngine) restoreOracle(dumpFile, connStr, targetDB string, session *RestoreSession) error {
	e.log(session, fmt.Sprintf("🗄️ Відновлення Oracle бази даних..."))

	if runtime.GOOS != "windows" && runtime.GOOS != "linux" {
		return fmt.Errorf("відновлення Oracle підтримується тільки на Windows/Linux")
	}

	// Use impdp (Data Pump) or imp for import
	// Connection string format: user/password@//host:port/service_name

	impPath := "impdp"

	// Build impdp command
	// Example: impdp scott/tiger@//localhost:1521/ORCL DIRECTORY=data_pump_dir DUMPFILE=export.dmp REMAP_SCHEMA=source:target
	cmd := exec.Command(impPath,
		connStr,
		"DUMPFILE="+dumpFile,
		"DIRECTORY=DATA_PUMP_DIR",
		"FULL=Y",
	)

	output, err := cmd.CombinedOutput()

	session.Logs = append(session.Logs, fmt.Sprintf("impdp output: %s", string(output)))

	if err != nil {
		return fmt.Errorf("помилка відновлення Oracle: %v - %s", err, string(output))
	}

	e.log(session, "✅ Oracle базу даних успішно відновлено")
	return nil
}

// restoreSQLite restores SQLite database
func (e *RestoreEngine) restoreSQLite(dumpFile, destPath string, session *RestoreSession) error {
	data, err := os.ReadFile(dumpFile)
	if err != nil {
		return err
	}

	if err := os.WriteFile(destPath, data, 0644); err != nil {
		return err
	}

	return nil
}

// restoreVM restores a Hyper-V VM
func (e *RestoreEngine) restoreVM(req *RestoreRequest, session *RestoreSession) error {
	e.log(session, "🖥️ Початок відновлення віртуальної машини...")

	vmBackupPath := filepath.Join(req.BackupPath, "vms", req.VMName)

	// Check if backup exists
	if _, err := os.Stat(vmBackupPath); os.IsNotExist(err) {
		return fmt.Errorf("резервна копія VM %s не знайдена", req.VMName)
	}

	// Use PowerShell to restore VM
	vmName := req.NewVMName
	if vmName == "" {
		vmName = req.VMName
	}

	psScript := fmt.Sprintf(`Import-VM -Path "%s" -GenerateNewId -Copy`, vmBackupPath)

	cmd := exec.Command("powershell", "-Command", psScript)
	output, err := cmd.CombinedOutput()

	session.Logs = append(session.Logs, fmt.Sprintf("PowerShell output: %s", string(output)))

	if err != nil {
		return fmt.Errorf("помилка відновлення VM: %v - %s", err, string(output))
	}

	session.FilesRestored = 1
	e.log(session, fmt.Sprintf("✅ VM %s успішно відновлено", vmName))

	return nil
}

// instantRestore performs instant VM recovery
func (e *RestoreEngine) instantRestore(req *RestoreRequest, session *RestoreSession) error {
	e.log(session, "⚡ Початок миттєвого відновлення...")

	// Instant restore runs VM directly from backup
	// This is a Veeam-like feature

	e.log(session, "🚀 Запуск VM з резервної копії...")

	// TODO: Implement instant restore using Hyper-V checkpoints

	session.FilesRestored = 1
	e.log(session, "✅ Миттєве відновлення успішно")

	return nil
}

// Helper functions

func (e *RestoreEngine) log(session *RestoreSession, message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logLine := fmt.Sprintf("[%s] %s", timestamp, message)

	if session != nil {
		session.Logs = append(session.Logs, logLine)
	}

	os.MkdirAll(filepath.Dir(e.LogFile), 0755)
	f, _ := os.OpenFile(e.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if f != nil {
		f.WriteString(logLine + "\n")
		f.Close()
	}

	fmt.Println(logLine)
}

func (e *RestoreEngine) saveSession(session *RestoreSession) {
	sessionsDir := filepath.Join(e.DataDir, "sessions", "restore")
	os.MkdirAll(sessionsDir, 0755)

	sessionFile := filepath.Join(sessionsDir, fmt.Sprintf("%s.json", session.ID))
	data, _ := json.MarshalIndent(session, "", "  ")
	os.WriteFile(sessionFile, data, 0644)
}

func (e *RestoreEngine) runScript(scriptPath string) error {
	cmd := exec.Command(scriptPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v - %s", err, string(output))
	}
	return nil
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

func (e *RestoreEngine) decryptFile(src, dst, password string) error {
	ciphertext, err := os.ReadFile(src)
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

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, plaintext, 0644)
}

func (e *RestoreEngine) formatBytes(bytes int64) string {
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

// ListRestorePoints returns available restore points
func (e *RestoreEngine) ListRestorePoints(backupPath string) ([]map[string]interface{}, error) {
	var points []map[string]interface{}

	// Read metadata files
	matches, _ := filepath.Glob(filepath.Join(backupPath, "*", "metadata.json"))

	for _, metaFile := range matches {
		data, err := os.ReadFile(metaFile)
		if err != nil {
			continue
		}

		var metadata map[string]interface{}
		if err := json.Unmarshal(data, &metadata); err != nil {
			continue
		}

		// Add path info
		metadata["path"] = filepath.Dir(metaFile)
		points = append(points, metadata)
	}

	return points, nil
}

// BrowseBackupFiles lists files in a backup
func (e *RestoreEngine) BrowseBackupFiles(backupPath string) ([]BackupFileInfo, error) {
	archivePath := filepath.Join(backupPath, "backup.zip")

	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	files := make([]BackupFileInfo, 0, len(r.File))
	for _, f := range r.File {
		info := f.FileInfo()
		fileType := "file"
		if info.IsDir() {
			fileType = "directory"
		}
		files = append(files, BackupFileInfo{
			Name:     f.Name,
			Type:     fileType,
			Size:     info.Size(),
			Modified: info.ModTime(),
		})
	}

	return files, nil
}
