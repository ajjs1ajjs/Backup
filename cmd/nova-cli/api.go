package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"novabackup/internal/api"
	"novabackup/internal/database"
	"novabackup/internal/scheduler"

	"github.com/spf13/cobra"
)

var (
	apiPort int
	apiHost string
)

func initAPICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api",
		Short: "Start REST API server",
		Long:  "Start the NovaBackup REST API server with Swagger documentation",
	}

	cmd.Flags().IntVarP(&apiPort, "port", "p", 8080, "API server port")
	cmd.Flags().StringVarP(&apiHost, "host", "", "0.0.0.0", "API server host")

	cmd.AddCommand(&cobra.Command{
		Use:   "start",
		Short: "Start the API server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return startAPIServer()
		},
	})

	return cmd
}

func startAPIServer() error {
	dbPath := os.Getenv("NOVA_DB_PATH")
	if dbPath == "" {
		dbPath = "novabackup.db"
	}

	fmt.Println("🚀 Starting NovaBackup API Server...")
	fmt.Printf("   Database: %s\n", dbPath)
	fmt.Printf("   Host: %s\n", apiHost)
	fmt.Printf("   Port: %d\n", apiPort)
	fmt.Println("   Swagger UI: http://localhost:%d/swagger/index.html", apiPort)

	// Initialize database
	db, err := database.NewSQLiteConnection(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	// Initialize scheduler
	sched, err := scheduler.NewScheduler(db)
	if err != nil {
		return fmt.Errorf("failed to create scheduler: %w", err)
	}

	if err := sched.Start(); err != nil {
		log.Printf("Warning: failed to start scheduler: %v", err)
	}

	// Create API server
	server, err := api.NewServer(db, sched)
	if err != nil {
		return fmt.Errorf("failed to create API server: %w", err)
	}

	// Start server in goroutine
	go func() {
		if err := server.Start(apiPort); err != nil {
			log.Printf("API server error: %v", err)
		}
	}()

	fmt.Println("✅ API server started")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n🛑 Shutting down API server...")
	if err := sched.Stop(); err != nil {
		log.Printf("Error stopping scheduler: %v", err)
	}

	fmt.Println("API server stopped")
	return nil
}
