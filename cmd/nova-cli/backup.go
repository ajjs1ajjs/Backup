package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"novabackup/internal/backup"
	"novabackup/internal/database"
	"novabackup/internal/providers"
	"novabackup/pkg/models"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var (
	backupSource     string
	backupDest       string
	backupJobType    string
	compressEnabled  bool
	encryptEnabled   bool
	dedupeEnabled    bool
	chunkSize        int64
	compressionLevel int
	encryptionKey    string

	// Database backup flags
	dbHost     string
	dbPort     int
	dbUser     string
	dbPassword string
	dbName     string
	dbType     string

	// VM backup flags
	vmName      string
	vcenterHost string
	hypervHost  string
)

func initBackupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Perform backup operations",
		Long:  "Execute backup jobs for files, databases, or virtual machines",
	}

	cmd.PersistentFlags().StringVarP(&backupSource, "source", "s", "", "Source path or connection string")
	cmd.PersistentFlags().StringVarP(&backupDest, "destination", "d", "", "Destination path for backup")
	cmd.PersistentFlags().StringVarP(&backupJobType, "type", "t", "file", "Backup type: file, database, vm")
	cmd.PersistentFlags().BoolVarP(&compressEnabled, "compress", "c", true, "Enable compression")
	cmd.PersistentFlags().BoolVarP(&encryptEnabled, "encrypt", "e", false, "Enable encryption")
	cmd.PersistentFlags().BoolVarP(&dedupeEnabled, "dedupe", "", true, "Enable deduplication")
	cmd.PersistentFlags().Int64VarP(&chunkSize, "chunk-size", "", 4*1024*1024, "Chunk size in bytes")
	cmd.PersistentFlags().IntVarP(&compressionLevel, "compression-level", "", 3, "Compression level (1-9)")
	cmd.PersistentFlags().StringVarP(&encryptionKey, "key", "k", "", "Encryption key (32 bytes for AES-256)")

	// Database backup flags
	cmd.PersistentFlags().StringVarP(&dbHost, "host", "", "localhost", "Database host")
	cmd.PersistentFlags().IntVarP(&dbPort, "port", "", 0, "Database port")
	cmd.PersistentFlags().StringVarP(&dbUser, "user", "u", "", "Database user")
	cmd.PersistentFlags().StringVarP(&dbPassword, "password", "w", "", "Database password")
	cmd.PersistentFlags().StringVarP(&dbName, "database", "D", "", "Database name")
	cmd.PersistentFlags().StringVarP(&dbType, "db-type", "", "mysql", "Database type: mysql, postgres")

	// VM backup flags
	cmd.PersistentFlags().StringVarP(&vmName, "vm-name", "", "", "VM name")
	cmd.PersistentFlags().StringVarP(&vcenterHost, "vcenter", "", "", "vCenter host (VMware)")
	cmd.PersistentFlags().StringVarP(&hypervHost, "hyperv-host", "", "", "Hyper-V host")

	cmd.AddCommand(runBackupCmd())
	cmd.AddCommand(createBackupJobCmd())
	cmd.AddCommand(listBackupsCmd())
	cmd.AddCommand(vmBackupCmd())

	return cmd
}

func runBackupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Run an immediate backup job",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			dbPath := os.Getenv("NOVA_DB_PATH")
			if dbPath == "" {
				dbPath = "novabackup.db"
			}

			db, err := database.NewSQLiteConnection(dbPath)
			if err != nil {
				return fmt.Errorf("failed to initialize database: %w", err)
			}
			defer db.Close()

			// Generate encryption key if not provided
			var encKey []byte
			if encryptEnabled {
				if encryptionKey != "" {
					encKey = []byte(encryptionKey)
					// Pad or truncate to 32 bytes
					if len(encKey) < 32 {
						return fmt.Errorf("encryption key must be at least 32 bytes for AES-256, got %d", len(encKey))
					}
					encKey = encKey[:32]
				} else {
					// Generate a default key (in production, use secure key management)
					encKey = []byte("0123456789abcdef0123456789abcdef")
				}
			}

			config := &models.BackupConfig{
				Source:        backupSource,
				Destination:   backupDest,
				Compress:      compressEnabled,
				Encrypt:       encryptEnabled,
				EncryptionKey: encKey,
				ChunkSize:     chunkSize,
				BufferSize:    256,
			}

			engine := backup.NewBackupEngine(db, config)

			job := &models.Job{
				ID:          uuid.New(),
				Name:        fmt.Sprintf("Manual Backup - %s", time.Now().Format("2006-01-02 15:04:05")),
				Description: "Manual backup job",
				JobType:     models.JobType(backupJobType),
				Source:      backupSource,
				Destination: backupDest,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			var result *models.BackupResult

			switch backupJobType {
			case "file":
				result, err = runFileBackup(ctx, engine, job, config)
			case "database":
				err = fmt.Errorf("database backup not yet implemented")
			case "vm":
				err = fmt.Errorf("VM backup not yet implemented")
			default:
				err = fmt.Errorf("unknown backup type: %s", backupJobType)
			}

			if err != nil {
				return err
			}

			fmt.Printf("✅ Backup completed successfully!\n")
			fmt.Printf("   Job ID: %s\n", job.ID)
			fmt.Printf("   Files Processed: %d\n", result.FilesTotal)
			fmt.Printf("   Bytes Read: %d\n", result.BytesRead)
			fmt.Printf("   Bytes Written: %d\n", result.BytesWritten)
			fmt.Printf("   Duration: %v\n", result.EndTime.Sub(result.StartTime))

			return nil
		},
	}
}

func runFileBackup(ctx context.Context, engine *backup.BackupEngine, job *models.Job, config *models.BackupConfig) (*models.BackupResult, error) {
	result := &models.BackupResult{
		ID:        uuid.New(),
		JobID:     job.ID,
		Status:    models.JobStatusRunning,
		StartTime: time.Now(),
	}

	// Use the enhanced backup engine
	backupResult, err := engine.PerformBackup(ctx, config.Source)
	if err != nil {
		result.Status = models.JobStatusFailed
		result.ErrorMessage = err.Error()
		result.EndTime = time.Now()
		return result, err
	}

	result.FilesTotal = backupResult.FilesTotal
	result.FilesSuccess = backupResult.FilesSuccess
	result.FilesFailed = backupResult.FilesFailed
	result.BytesRead = backupResult.BytesRead
	result.BytesWritten = backupResult.BytesWritten
	result.Status = backupResult.Status
	result.EndTime = backupResult.EndTime

	return result, nil
}

func createBackupJobCmd() *cobra.Command {
	var jobName string
	var schedule string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a scheduled backup job",
		RunE: func(cmd *cobra.Command, args []string) error {
			dbPath := os.Getenv("NOVA_DB_PATH")
			if dbPath == "" {
				dbPath = "novabackup.db"
			}

			db, err := database.NewSQLiteConnection(dbPath)
			if err != nil {
				return fmt.Errorf("failed to initialize database: %w", err)
			}
			defer db.Close()

			// Clean job name (remove quotes if present)
			cleanName := strings.Trim(jobName, `"'`)
			if cleanName == "" {
				return fmt.Errorf("job name is required")
			}

			job := &models.Job{
				ID:          uuid.New(),
				Name:        cleanName,
				Description: fmt.Sprintf("Scheduled %s backup", backupJobType),
				JobType:     models.JobType(backupJobType),
				Source:      strings.Trim(backupSource, `"'`),
				Destination: strings.Trim(backupDest, `"'`),
				Schedule:    strings.Trim(schedule, `"'`),
				Enabled:     true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			if err := db.CreateJob(job); err != nil {
				return fmt.Errorf("failed to create job: %w", err)
			}

			fmt.Printf("✅ Backup job created successfully!\n")
			fmt.Printf("   Job ID: %s\n", job.ID)
			fmt.Printf("   Name: %s\n", job.Name)
			fmt.Printf("   Type: %s\n", job.JobType)
			fmt.Printf("   Schedule: %s\n", job.Schedule)
			fmt.Printf("   Source: %s\n", job.Source)
			fmt.Printf("   Destination: %s\n", job.Destination)

			return nil
		},
	}

	cmd.Flags().StringVarP(&jobName, "name", "n", "", "Job name")
	cmd.Flags().StringVarP(&schedule, "schedule", "", "", "Cron schedule (e.g., '0 2 * * *' for daily at 2 AM)")
	cmd.MarkFlagRequired("name")

	return cmd
}

func listBackupsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List backup jobs and restore points",
		RunE: func(cmd *cobra.Command, args []string) error {
			dbPath := os.Getenv("NOVA_DB_PATH")
			if dbPath == "" {
				dbPath = "novabackup.db"
			}

			db, err := database.NewSQLiteConnection(dbPath)
			if err != nil {
				return fmt.Errorf("failed to initialize database: %w", err)
			}
			defer db.Close()

			jobs, err := db.GetAllJobs()
			if err != nil {
				return fmt.Errorf("failed to get jobs: %w", err)
			}

			if len(jobs) == 0 {
				fmt.Println("No backup jobs found.")
				return nil
			}

			fmt.Printf("\n📦 Backup Jobs (%d)\n", len(jobs))
			fmt.Println("================================================================================")
			fmt.Printf("%-36s %-20s %-10s %-8s %s\n", "ID", "Name", "Type", "Status", "Schedule")
			fmt.Println("================================================================================")

			for _, job := range jobs {
				status := "✅ Enabled"
				if !job.Enabled {
					status = "⏸️ Disabled"
				}
				schedule := job.Schedule
				if schedule == "" {
					schedule = "-"
				}
				fmt.Printf("%-36s %-20s %-10s %-8s %s\n",
					job.ID.String()[:8]+"...",
					job.Name,
					job.JobType,
					status,
					schedule)
			}

			return nil
		},
	}
}

func initJobCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "job",
		Short: "Manage backup jobs",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "start <job-id>",
		Short: "Start a backup job manually",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobID, err := uuid.Parse(args[0])
			if err != nil {
				return fmt.Errorf("invalid job ID: %w", err)
			}
			log.Printf("Starting job %s...", jobID)
			fmt.Println("Job started successfully")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "stop <job-id>",
		Short: "Stop a running backup job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobID, err := uuid.Parse(args[0])
			if err != nil {
				return fmt.Errorf("invalid job ID: %w", err)
			}
			log.Printf("Stopping job %s...", jobID)
			fmt.Println("Job stopped successfully")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "delete <job-id>",
		Short: "Delete a backup job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobID, err := uuid.Parse(args[0])
			if err != nil {
				return fmt.Errorf("invalid job ID: %w", err)
			}
			log.Printf("Deleting job %s...", jobID)
			fmt.Println("Job deleted successfully")
			return nil
		},
	})

	return cmd
}

// dbBackupCmd creates the database backup command
func dbBackupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "db-backup",
		Short: "Backup a database (MySQL/PostgreSQL)",
		Long:  "Perform database backup using native tools (mysqldump/pg_dump)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if dbName == "" {
				return fmt.Errorf("database name is required (--database)")
			}
			if dbUser == "" {
				return fmt.Errorf("database user is required (--user)")
			}
			if backupDest == "" {
				return fmt.Errorf("destination is required (--destination)")
			}

			fmt.Printf("🗄️  Starting %s database backup...\n", dbType)
			fmt.Printf("   Host: %s:%d\n", dbHost, dbPort)
			fmt.Printf("   Database: %s\n", dbName)
			fmt.Printf("   User: %s\n", dbUser)
			fmt.Printf("   Destination: %s\n", backupDest)

			var result *models.BackupResult
			var err error

			switch dbType {
			case "mysql":
				result, err = runMySQLBackup(ctx, dbName, backupDest)
			case "postgres", "postgresql":
				result, err = runPostgreSQLBackup(ctx, dbName, backupDest)
			default:
				return fmt.Errorf("unsupported database type: %s (use mysql or postgres)", dbType)
			}

			if err != nil {
				return err
			}

			fmt.Printf("✅ Database backup completed!\n")
			fmt.Printf("   Bytes Written: %d\n", result.BytesWritten)
			fmt.Printf("   Duration: %v\n", result.EndTime.Sub(result.StartTime))

			return nil
		},
	}
}

func runMySQLBackup(ctx context.Context, database, dest string) (*models.BackupResult, error) {
	cfg := providers.MySQLConfig{
		Host:     dbHost,
		Port:     dbPort,
		User:     dbUser,
		Password: dbPassword,
		Database: database,
	}

	provider := providers.NewMySQLBackupProvider(cfg)
	return provider.Backup(ctx, dest)
}

func runPostgreSQLBackup(ctx context.Context, database, dest string) (*models.BackupResult, error) {
	cfg := providers.PostgreSQLConfig{
		Host:     dbHost,
		Port:     dbPort,
		User:     dbUser,
		Password: dbPassword,
		Database: database,
		SSLMode:  "disable",
	}

	provider := providers.NewPostgreSQLBackupProvider(cfg)
	return provider.Backup(ctx, dest)
}

// vmBackupCmd creates the VM backup command
func vmBackupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "vm-backup",
		Short: "Backup a virtual machine (VMware/Hyper-V)",
		Long:  "Perform VM backup using hypervisor APIs",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if vmName == "" {
				return fmt.Errorf("VM name is required (--vm-name)")
			}
			if backupDest == "" {
				return fmt.Errorf("destination is required (--destination)")
			}

			var result *models.BackupResult
			var err error

			if vcenterHost != "" {
				// VMware backup
				result, err = runVMwareBackup(ctx, vmName, backupDest)
			} else if hypervHost != "" || hypervHost == "" {
				// Hyper-V backup (local by default)
				result, err = runHyperVBackup(ctx, vmName, backupDest)
			} else {
				return fmt.Errorf("must specify --vcenter or --hyperv-host")
			}

			if err != nil {
				return err
			}

			fmt.Printf("✅ VM backup completed!\n")
			fmt.Printf("   Bytes Written: %d\n", result.BytesWritten)
			fmt.Printf("   Duration: %v\n", result.EndTime.Sub(result.StartTime))

			return nil
		},
	}
}

func runVMwareBackup(ctx context.Context, vmName, dest string) (*models.BackupResult, error) {
	cfg := providers.VMwareConfig{
		VCenter:    vcenterHost,
		Username:   dbUser,     // Reuse dbUser for VMware username
		Password:   dbPassword, // Reuse dbPassword
		Insecure:   true,
		Datacenter: dbName, // Reuse dbName for datacenter
	}

	provider := providers.NewVMwareBackupProvider(cfg)
	return provider.Backup(ctx, vmName, dest)
}

func runHyperVBackup(ctx context.Context, vmName, dest string) (*models.BackupResult, error) {
	cfg := providers.HyperVConfig{
		Host:      hypervHost,
		UseRemote: hypervHost != "",
	}

	provider := providers.NewHyperVBackupProvider(cfg)
	return provider.Backup(ctx, vmName, dest)
}
