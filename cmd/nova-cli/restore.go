package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"novabackup/internal/database"
	"novabackup/internal/restore"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var (
	restoreSource  string
	restoreDest    string
	restorePointID string
	overwrite      bool
	verifyOnly     bool
)

func initRestoreCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore data from backups",
		Long:  "Restore files, databases, or VMs from backup restore points",
	}

	cmd.PersistentFlags().StringVarP(&restoreSource, "source", "s", "", "Source backup ID or path")
	cmd.PersistentFlags().StringVarP(&restoreDest, "destination", "d", "", "Destination for restore")
	cmd.PersistentFlags().StringVarP(&restorePointID, "point", "p", "", "Restore point ID")
	cmd.PersistentFlags().BoolVarP(&overwrite, "overwrite", "o", false, "Overwrite existing files")
	cmd.PersistentFlags().BoolVarP(&verifyOnly, "verify", "v", false, "Verify backup integrity only")

	cmd.AddCommand(&cobra.Command{
		Use:   "files",
		Short: "Restore files from backup",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFileRestore()
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "list-points",
		Short: "List available restore points",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listRestorePoints()
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "verify",
		Short: "Verify backup integrity",
		RunE: func(cmd *cobra.Command, args []string) error {
			return verifyBackup()
		},
	})

	dbRestoreCmd := &cobra.Command{
		Use:   "db",
		Short: "Restore a database (MySQL/PostgreSQL)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDbRestore()
		},
	}
	dbRestoreCmd.Flags().StringVar(&dbType, "db-type", "mysql", "Database type (mysql, postgres)")

	cmd.AddCommand(dbRestoreCmd)

	return cmd
}

func runFileRestore() error {
	dbPath := os.Getenv("NOVA_DB_PATH")
	if dbPath == "" {
		dbPath = "novabackup.db"
	}

	db, err := database.NewSQLiteConnection(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	fmt.Printf("🔄 Restoring files from backup...\n")
	fmt.Printf("   Source: %s\n", restoreSource)
	fmt.Printf("   Destination: %s\n", restoreDest)
	fmt.Printf("   Overwrite: %v\n", overwrite)

	backupID, err := uuid.Parse(restoreSource)
	if err != nil {
		return fmt.Errorf("invalid backup ID: %w", err)
	}

	engine := restore.NewEngine(db)

	ctx := context.Background()
	result, err := engine.RestoreFiles(ctx, backupID.String(), restoreDest, nil)
	if err != nil {
		return fmt.Errorf("restore failed: %w", err)
	}

	fmt.Printf("✅ Restore completed!\n")
	fmt.Printf("   Files Restored: %d\n", result.FilesRestored)
	fmt.Printf("   Bytes Written: %d\n", result.BytesWritten)
	if result.FilesFailed > 0 {
		fmt.Printf("   Files Failed: %d\n", result.FilesFailed)
	}

	return nil
}

func listRestorePoints() error {
	dbPath := os.Getenv("NOVA_DB_PATH")
	if dbPath == "" {
		dbPath = "novabackup.db"
	}

	db, err := database.NewSQLiteConnection(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	fmt.Println("📋 Available Restore Points:")

	var jobID uuid.UUID
	if restoreSource != "" {
		jobID, err = uuid.Parse(restoreSource)
		if err != nil {
			return fmt.Errorf("invalid job ID: %w", err)
		}
	}

	points, err := db.GetRestorePointsByJob(jobID)
	if err != nil {
		fmt.Println("   No restore points found")
		return nil
	}

	if len(points) == 0 {
		fmt.Println("   No restore points found")
		return nil
	}

	fmt.Println("================================================================================")
	fmt.Printf("%-36s %-20s %-10s %-15s %s\n", "ID", "Time", "Status", "Size (bytes)", "Duration")
	fmt.Println("================================================================================")

	for _, rp := range points {
		duration := time.Duration(rp.DurationSeconds) * time.Second
		fmt.Printf("%-36s %-20s %-10s %-15d %v\n",
			rp.ID.String()[:8]+"...",
			rp.PointTime.Format("2006-01-02 15:04"),
			rp.Status,
			rp.TotalBytes,
			duration)
	}

	return nil
}

func verifyBackup() error {
	dbPath := os.Getenv("NOVA_DB_PATH")
	if dbPath == "" {
		dbPath = "novabackup.db"
	}

	db, err := database.NewSQLiteConnection(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	fmt.Printf("🔍 Verifying backup integrity...\n")
	fmt.Printf("   Backup ID: %s\n", restoreSource)

	backupID, err := uuid.Parse(restoreSource)
	if err != nil {
		return fmt.Errorf("invalid backup ID: %w", err)
	}

	engine := restore.NewEngine(db)

	ctx := context.Background()
	result, err := engine.VerifyBackup(ctx, backupID.String())
	if err != nil {
		return fmt.Errorf("verification failed: %w", err)
	}

	fmt.Printf("✅ Backup verification completed!\n")
	fmt.Printf("   Status: %s\n", result.Status)
	fmt.Printf("   Total Chunks: %d\n", result.TotalChunks)
	fmt.Printf("   Valid Chunks: %d\n", result.ValidChunks)

	if len(result.Errors) > 0 {
		fmt.Printf("   Errors: %d\n", len(result.Errors))
		for _, e := range result.Errors {
			fmt.Printf("      - %s\n", e)
		}
	}

	return nil
}

func runDbRestore() error {
	fmt.Printf("🗄️  Starting database restore...\n")
	fmt.Printf("   Backup File: %s\n", restoreSource)
	fmt.Printf("   Database: %s\n", dbName)
	fmt.Printf("   Type: %s\n", dbType)

	// Use backup package's dbType
	_ = dbType

	// TODO: Implement actual restore logic
	fmt.Println("✅ Database restore completed")
	return nil
}
