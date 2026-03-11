// Package vss provides Windows Volume Shadow Copy Service (VSS) integration
package vss

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/denisenkom/go-mssqldb" // SQL Server driver
)

// SQLServerManager provides SQL Server-specific backup functionality
type SQLServerManager struct {
	connectionString string
}

// SQLServerInfo contains information about a SQL Server instance
type SQLServerInfo struct {
	ServerName       string            `json:"server_name"`
	InstanceName     string            `json:"instance_name"`
	Version          string            `json:"version"`
	Edition          string            `json:"edition"`
	Databases        []DatabaseInfo    `json:"databases"`
	IsClustered      bool              `json:"is_clustered"`
	IsAlwaysOn       bool              `json:"is_always_on"`
}

// DatabaseInfo contains information about a SQL Server database
type DatabaseInfo struct {
	Name           string `json:"name"`
	DatabaseID     int    `json:"database_id"`
	State          string `json:"state"`
	RecoveryModel  string `json:"recovery_model"`
	SizeMB         float64 `json:"size_mb"`
	LastBackupTime string `json:"last_backup_time,omitempty"`
	IsSystemDB     bool   `json:"is_system_db"`
}

// SQLBackupResult contains SQL Server backup results
type SQLBackupResult struct {
	DatabaseName     string  `json:"database_name"`
	BackupType       string  `json:"backup_type"`
	StartTime        string  `json:"start_time"`
	EndTime          string  `json:"end_time"`
	BackupFile       string  `json:"backup_file"`
	BackupSizeMB     float64 `json:"backup_size_mb"`
	Compressed       bool    `json:"compressed"`
	Checksum         bool    `json:"checksum"`
}

// NewSQLServerManager creates a new SQL Server manager
func NewSQLServerManager(connectionString string) *SQLServerManager {
	return &SQLServerManager{
		connectionString: connectionString,
	}
}

// Connect establishes connection to SQL Server
func (s *SQLServerManager) Connect() (*sql.DB, error) {
	db, err := sql.Open("sqlserver", s.connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQL Server connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to SQL Server: %w", err)
	}

	return db, nil
}

// GetServerInfo retrieves SQL Server instance information
func (s *SQLServerManager) GetServerInfo() (*SQLServerInfo, error) {
	db, err := s.Connect()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	info := &SQLServerInfo{
		Databases: []DatabaseInfo{},
	}

	// Get server properties
	query := `
		SELECT 
			@@SERVERNAME as server_name,
			SERVERPROPERTY('InstanceName') as instance_name,
			@@VERSION as version,
			SERVERPROPERTY('Edition') as edition,
			SERVERPROPERTY('IsClustered') as is_clustered,
			SERVERPROPERTY('IsHadrEnabled') as is_always_on
	`

	row := db.QueryRow(query)
	var isClustered, isAlwaysOn sql.NullInt64
	err = row.Scan(&info.ServerName, &info.InstanceName, &info.Version, 
		&info.Edition, &isClustered, &isAlwaysOn)
	if err != nil {
		return nil, fmt.Errorf("failed to get server info: %w", err)
	}

	if isClustered.Valid {
		info.IsClustered = isClustered.Int64 == 1
	}
	if isAlwaysOn.Valid {
		info.IsAlwaysOn = isAlwaysOn.Int64 == 1
	}

	// Get database list
	dbQuery := `
		SELECT 
			d.name,
			d.database_id,
			d.state_desc,
			d.recovery_model_desc,
			CAST(SUM(size) * 8.0 / 1024 AS DECIMAL(10,2)) as size_mb,
			ISNULL(MAX(b.backup_finish_date), '1900-01-01') as last_backup,
			CASE WHEN d.name IN ('master', 'model', 'msdb', 'tempdb') THEN 1 ELSE 0 END as is_system
		FROM sys.databases d
		LEFT JOIN msdb.dbo.backupset b ON d.name = b.database_name AND b.type = 'D'
		LEFT JOIN sys.master_files mf ON d.database_id = mf.database_id
		GROUP BY d.name, d.database_id, d.state_desc, d.recovery_model_desc
		ORDER BY d.name
	`

	rows, err := db.Query(dbQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get database list: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var dbInfo DatabaseInfo
		var lastBackup sql.NullTime
		err := rows.Scan(&dbInfo.Name, &dbInfo.DatabaseID, &dbInfo.State, 
			&dbInfo.RecoveryModel, &dbInfo.SizeMB, &lastBackup, &dbInfo.IsSystemDB)
		if err != nil {
			continue
		}
		if lastBackup.Valid {
			dbInfo.LastBackupTime = lastBackup.Time.Format("2006-01-02 15:04:05")
		}
		info.Databases = append(info.Databases, dbInfo)
	}

	return info, nil
}

// BackupDatabase performs a native SQL Server backup
func (s *SQLServerManager) BackupDatabase(databaseName, backupPath string, options SQLServerBackupOptions) (*SQLBackupResult, error) {
	db, err := s.Connect()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	result := &SQLBackupResult{
		DatabaseName: databaseName,
		BackupType:   options.BackupType,
		Compressed:   options.Compression,
		Checksum:     options.VerifyBackup,
	}

	// Build BACKUP command
	var backupCmd strings.Builder
	backupCmd.WriteString(fmt.Sprintf("BACKUP DATABASE [%s] TO DISK = '%s' ", databaseName, backupPath))

	// Add options
	if options.Compression {
		backupCmd.WriteString("WITH COMPRESSION ")
	}
	if options.VerifyBackup {
		if options.Compression {
			backupCmd.WriteString(", CHECKSUM ")
		} else {
			backupCmd.WriteString("WITH CHECKSUM ")
		}
	}

	// Execute backup
	_, err = db.Exec(backupCmd.String())
	if err != nil {
		return nil, fmt.Errorf("backup failed: %w", err)
	}

	return result, nil
}

// BackupTransactionLog backs up the transaction log
func (s *SQLServerManager) BackupTransactionLog(databaseName, backupPath string, truncate bool) error {
	db, err := s.Connect()
	if err != nil {
		return err
	}
	defer db.Close()

	var cmd string
	if truncate {
		cmd = fmt.Sprintf("BACKUP LOG [%s] TO DISK = '%s' WITH TRUNCATE_ONLY", databaseName, backupPath)
	} else {
		cmd = fmt.Sprintf("BACKUP LOG [%s] TO DISK = '%s'", databaseName, backupPath)
	}

	_, err = db.Exec(cmd)
	if err != nil {
		return fmt.Errorf("transaction log backup failed: %w", err)
	}

	return nil
}

// RestoreDatabase restores a SQL Server database
func (s *SQLServerManager) RestoreDatabase(databaseName, backupPath string, replace bool) error {
	db, err := s.Connect()
	if err != nil {
		return err
	}
	defer db.Close()

	cmd := fmt.Sprintf("RESTORE DATABASE [%s] FROM DISK = '%s'", databaseName, backupPath)
	if replace {
		cmd += " WITH REPLACE"
	}

	_, err = db.Exec(cmd)
	if err != nil {
		return fmt.Errorf("restore failed: %w", err)
	}

	return nil
}

// GetRecoveryModel gets the recovery model for a database
func (s *SQLServerManager) GetRecoveryModel(databaseName string) (string, error) {
	db, err := s.Connect()
	if err != nil {
		return "", err
	}
	defer db.Close()

	query := "SELECT recovery_model_desc FROM sys.databases WHERE name = @p1"
	row := db.QueryRow(query, databaseName)

	var recoveryModel string
	err = row.Scan(&recoveryModel)
	if err != nil {
		return "", fmt.Errorf("failed to get recovery model: %w", err)
	}

	return recoveryModel, nil
}

// SetRecoveryModel sets the recovery model for a database
func (s *SQLServerManager) SetRecoveryModel(databaseName, model string) error {
	db, err := s.Connect()
	if err != nil {
		return err
	}
	defer db.Close()

	validModels := map[string]bool{"FULL": true, "BULK_LOGGED": true, "SIMPLE": true}
	if !validModels[strings.ToUpper(model)] {
		return fmt.Errorf("invalid recovery model: %s", model)
	}

	cmd := fmt.Sprintf("ALTER DATABASE [%s] SET RECOVERY %s", databaseName, model)
	_, err = db.Exec(cmd)
	if err != nil {
		return fmt.Errorf("failed to set recovery model: %w", err)
	}

	return nil
}

// CheckDB performs database consistency check
func (s *SQLServerManager) CheckDB(databaseName string) error {
	db, err := s.Connect()
	if err != nil {
		return err
	}
	defer db.Close()

	cmd := fmt.Sprintf("DBCC CHECKDB ([%s]) WITH NO_INFOMSGS", databaseName)
	_, err = db.Exec(cmd)
	if err != nil {
		return fmt.Errorf("consistency check failed: %w", err)
	}

	return nil
}

// ShrinkDatabase shrinks database files
func (s *SQLServerManager) ShrinkDatabase(databaseName string) error {
	db, err := s.Connect()
	if err != nil {
		return err
	}
	defer db.Close()

	cmd := fmt.Sprintf("DBCC SHRINKDATABASE ([%s])", databaseName)
	_, err = db.Exec(cmd)
	if err != nil {
		return fmt.Errorf("shrink failed: %w", err)
	}

	return nil
}

// GetConnectionStringFromParams builds connection string from parameters
func GetConnectionStringFromParams(server, instance, user, password string, port int) string {
	var parts []string
	
	if instance != "" {
		parts = append(parts, fmt.Sprintf("Server=%s\\%s", server, instance))
	} else if port > 0 {
		parts = append(parts, fmt.Sprintf("Server=%s,%d", server, port))
	} else {
		parts = append(parts, fmt.Sprintf("Server=%s", server))
	}
	
	if user != "" {
		parts = append(parts, fmt.Sprintf("User Id=%s", user))
		parts = append(parts, fmt.Sprintf("Password=%s", password))
	} else {
		parts = append(parts, "Integrated Security=true")
	}
	
	return strings.Join(parts, ";")
}
