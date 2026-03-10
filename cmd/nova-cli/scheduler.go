package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"novabackup/internal/database"
	"novabackup/internal/scheduler"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var (
	schedulerDBPath string
)

func initSchedulerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scheduler",
		Short: "Manage backup scheduler",
		Long:  "Start, stop, or manage the backup job scheduler",
	}

	cmd.PersistentFlags().StringVarP(&schedulerDBPath, "db", "", "novabackup.db", "Database path")

	cmd.AddCommand(&cobra.Command{
		Use:   "start",
		Short: "Start the scheduler daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("🕐 Starting backup scheduler...")
			fmt.Printf("   Database: %s\n", schedulerDBPath)

			db, err := database.NewSQLiteConnection(schedulerDBPath)
			if err != nil {
				return fmt.Errorf("failed to initialize database: %w", err)
			}
			defer db.Close()

			sched, err := scheduler.NewScheduler(db)
			if err != nil {
				return fmt.Errorf("failed to create scheduler: %w", err)
			}

			if err := sched.Start(); err != nil {
				return fmt.Errorf("failed to start scheduler: %w", err)
			}

			fmt.Printf("   Scheduler started with %d jobs\n", sched.GetJobCount())
			fmt.Println("   Press Ctrl+C to stop")

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
			<-sigChan

			fmt.Println("\n🛑 Stopping scheduler...")
			if err := sched.Stop(); err != nil {
				fmt.Printf("Error stopping scheduler: %v\n", err)
			}
			fmt.Println("Scheduler stopped gracefully")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show scheduler status",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := database.NewSQLiteConnection(schedulerDBPath)
			if err != nil {
				return fmt.Errorf("failed to initialize database: %w", err)
			}
			defer db.Close()

			sched, err := scheduler.NewScheduler(db)
			if err != nil {
				return fmt.Errorf("failed to create scheduler: %w", err)
			}

			fmt.Println("📊 Scheduler Status:")
			fmt.Printf("   Active jobs: %d\n", sched.GetJobCount())
			// TODO: Show next scheduled run
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "run-now <job-id>",
		Short: "Run a job immediately",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobID, err := uuid.Parse(args[0])
			if err != nil {
				return fmt.Errorf("invalid job ID: %w", err)
			}

			db, err := database.NewSQLiteConnection(schedulerDBPath)
			if err != nil {
				return fmt.Errorf("failed to initialize database: %w", err)
			}
			defer db.Close()

			job, err := db.GetJobByID(jobID)
			if err != nil {
				return fmt.Errorf("job not found: %w", err)
			}

			fmt.Printf("🚀 Running job %s (%s)...\n", job.ID, job.Name)
			// TODO: Implement immediate job execution
			fmt.Println("Job execution started (placeholder)")
			return nil
		},
	})

	return cmd
}
