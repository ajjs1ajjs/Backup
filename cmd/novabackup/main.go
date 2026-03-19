// NovaBackup - Modern Web-Based Backup Platform
// Usage: novabackup [command]
// Commands: server, install, remove, start, stop

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"novabackup/internal/api"
	"novabackup/internal/backup"
	"novabackup/internal/database"
	"novabackup/internal/notifications"
	"novabackup/internal/rbac"
	"novabackup/internal/reports"
	"novabackup/internal/restore"
	"novabackup/internal/scheduler"
	"novabackup/internal/storage"

	"github.com/gin-gonic/gin"
)

const (
	Version     = "7.0.0"
	DefaultPort = 8050
	ServiceName = "NovaBackup"
)

var (
	configPath         string
	dataDir            string
	webDir             string
	db                 *database.Database
	backupEngine       *backup.BackupEngine
	restoreEngine      *restore.RestoreEngine
	jobScheduler       *scheduler.Scheduler
	storageEngine      *storage.StorageEngine
	notificationEngine *notifications.NotificationEngine
	rbacEngine         *rbac.RBACEngine
	reportEngine       *reports.ReportEngine
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := strings.ToLower(os.Args[1])

	switch command {
	case "server", "start":
		runServer()
	case "install":
		installService()
	case "remove", "uninstall":
		removeService()
	case "version":
		fmt.Printf("NovaBackup v%s\n", Version)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func runServer() {
	if runAsServiceIfNeeded() {
		return
	}

	server, err := buildServer()
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	if err := serveHTTP(server); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func buildServer() (*http.Server, error) {
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║         NovaBackup Enterprise v7.0                        ║")
	fmt.Println("║         Modern Web-Based Backup Platform                  ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Initialize directories
	initDirectories()

	// Initialize database
	var err error
	dbPath := filepath.Join(dataDir, "novabackup.db")
	db, err = database.NewDatabase(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	fmt.Println("✓ Database initialized")

	// Migrate existing jobs to clean corrupted paths (Unicode characters)
	if err := db.MigrateCleanPaths(); err != nil {
		log.Printf("Warning: Path migration failed: %v", err)
	} else {
		fmt.Println("✓ Path migration completed (cleaned Unicode characters)")
	}

	// Set global DB for API - DON'T CLOSE, keep open for entire app lifetime
	api.DB = db

	// Set config path for API
	api.ConfigPath = configPath

	// Load configuration
	_ = loadConfig()
	fmt.Println("✓ Configuration loaded")

	// Initialize RBAC engine
	rbacEngine = rbac.NewRBACEngine()
	// Set database for session persistence
	rbacEngine.DB = db
	fmt.Println("✓ RBAC engine initialized")

	// Initialize audit engine
	api.AuditEngine = rbac.NewAuditEngine()
	fmt.Println("✓ Audit engine initialized")

	// Initialize notification engine
	notificationEngine = notifications.NewNotificationEngine()
	fmt.Println("✓ Notification engine initialized")

	// Initialize backup engine
	backupEngine = backup.NewBackupEngine(dataDir)
	fmt.Println("✓ Backup engine initialized")

	// Initialize restore engine
	restoreEngine = restore.NewRestoreEngine(dataDir)
	fmt.Println("✓ Restore engine initialized")

	// Initialize storage engine
	storageEngine = storage.NewStorageEngine(dataDir)
	fmt.Println("✓ Storage engine initialized")

	// Initialize report engine
	reportEngine = reports.NewReportEngine(db)
	fmt.Println("✓ Report engine initialized")

	// Set engines for API
	api.BackupEngine = backupEngine
	api.RestoreEngine = restoreEngine
	api.StorageEngine = storageEngine
	api.NotificationEngine = notificationEngine
	api.RBACEngine = rbacEngine
	api.ReportEngine = reportEngine

	// Initialize scheduler
	jobScheduler = scheduler.NewScheduler(db)
	jobScheduler.SetBackupEngine(backupEngine)
	if err := jobScheduler.Start(); err != nil {
		log.Printf("Warning: Failed to start scheduler: %v", err)
	}
	fmt.Println("✓ Scheduler started")

	// Set scheduler for API (after initialization)
	api.Scheduler = jobScheduler

	// Setup Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// Enable CORS
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Serve web UI from disk (must be before API routes)
	router.GET("/", func(c *gin.Context) {
		serveWebFile(c, "index.html")
	})

	router.GET("/assets/:file", func(c *gin.Context) {
		serveWebFile(c, filepath.Join("assets", c.Param("file")))
	})

	// Serve other web pages
	router.GET("/verification", func(c *gin.Context) {
		serveWebFile(c, "verification.html")
	})

	router.GET("/:filepath", func(c *gin.Context) {
		serveWebFile(c, c.Param("filepath"))
	})

	// API Routes (after web routes)
	apiGroup := router.Group("/api")
	{
		// Health (no auth required)
		apiGroup.GET("/health", api.GetHealth)

		// Auth (no auth required)
		auth := apiGroup.Group("/auth")
		auth.Use(api.AuditMiddleware()) // Log auth attempts
		{
			auth.POST("/login", api.Login)
			auth.POST("/logout", api.Logout)
			auth.POST("/change-password", api.ChangePassword)
		}

		// Protected routes (require authentication)
		protected := apiGroup.Group("")
		protected.Use(api.AuthMiddleware(), api.AuditMiddleware())
		{
			// Jobs
			protected.GET("/jobs", api.ListJobs)
			protected.POST("/jobs", api.CreateJob)
			protected.PUT("/jobs/:id", api.UpdateJob)
			protected.DELETE("/jobs/:id", api.DeleteJob)
			protected.POST("/jobs/:id/run", api.RunJob)
			protected.POST("/jobs/:id/stop", api.StopJob)

			// Backup
			protected.POST("/backup/run", api.RunBackup)
			protected.GET("/backup/sessions", api.ListSessions)
			protected.GET("/backup/sessions/:id", api.GetSession)
			protected.GET("/backup/sessions/:id/files", api.BrowseBackupFiles)
			protected.POST("/backup/verify", api.VerifyBackup)
			protected.GET("/backup/verifications", api.GetVerificationHistory)
			protected.GET("/backup/cbt-stats", api.GetCBTStatistics)

			// Restore
			protected.GET("/restore/points", api.ListRestorePoints)
			protected.POST("/restore/files", api.RestoreFiles)
			protected.POST("/restore/database", api.RestoreDatabase)
			protected.POST("/restore/instant", api.InstantRestore)

			// Storage
			protected.GET("/storage/repos", api.ListRepos)
			protected.POST("/storage/repos", api.CreateRepo)
			protected.PUT("/storage/repos/:id", api.UpdateRepo)
			protected.DELETE("/storage/repos/:id", api.DeleteRepo)

			// Settings
			protected.GET("/settings", api.GetSettings)
			protected.PUT("/settings", api.UpdateSettings)

			// Users
			protected.GET("/users", api.ListUsers)
			protected.GET("/users/:id", api.GetUser)
			protected.POST("/users", api.CreateUser)
			protected.PUT("/users/:id", api.UpdateUser)
			protected.DELETE("/users/:id", api.DeleteUser)
			protected.POST("/users/:id/enable", api.EnableUser)
			protected.POST("/users/:id/disable", api.DisableUser)

			// Reports & Statistics
			protected.GET("/reports/statistics", api.GetStatistics)
			protected.GET("/reports/daily", api.GetDailyReport)

			// Database Management
			protected.POST("/database/list", api.ListDatabases)
			protected.POST("/database/backup", api.BackupDatabase)

			// Audit Logs
			protected.GET("/audit/logs", api.GetAuditLogs)
		}
	}

	// Get server IP
	serverIP := getLocalIP()
	port := DefaultPort

	fmt.Println()
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║              ✅ Server Ready!                             ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("  🌐 Local:   http://localhost:%d\n", port)
	fmt.Printf("  🌐 Network: http://%s:%d\n", serverIP, port)
	fmt.Println()
	fmt.Println("  Default Login:")
	fmt.Println("    Username: admin")
	fmt.Println("    Password: admin123")
	fmt.Println()
	fmt.Println("  Press Ctrl+C to stop")
	fmt.Println()

	// Start server
	addr := fmt.Sprintf(":%d", port)
	return &http.Server{
		Addr:    addr,
		Handler: router,
	}, nil
}

func serveHTTP(server *http.Server) error {
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func shutdownServer(server *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_ = server.Shutdown(ctx)

	if jobScheduler != nil {
		jobScheduler.Stop()
	}

	if db != nil {
		_ = db.Close()
	}
}

func initDirectories() {
	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)

	// Check if running from Program Files
	if strings.Contains(exeDir, "Program Files") {
		dataDir = filepath.Join("C:\\ProgramData", "NovaBackup")
		configPath = filepath.Join(dataDir, "Config")
		webDir = filepath.Join(exeDir, "web")
	} else {
		dataDir = filepath.Join(exeDir, "data")
		configPath = filepath.Join(exeDir, "config")
		webDir = filepath.Join(exeDir, "web")
	}

	os.MkdirAll(dataDir, 0755)
	os.MkdirAll(configPath, 0755)
	os.MkdirAll(filepath.Join(dataDir, "backups"), 0755)
	os.MkdirAll(filepath.Join(dataDir, "logs"), 0755)
	os.MkdirAll(filepath.Join(dataDir, "sessions"), 0755)
	os.MkdirAll(webDir, 0755)
}

// loadConfig loads configuration from file
func loadConfig() map[string]interface{} {
	configFile := filepath.Join(configPath, "config.json")

	data, err := os.ReadFile(configFile)
	if err != nil {
		// Create default config
		defaultConfig := map[string]interface{}{
			"server": map[string]interface{}{
				"ip":         "0.0.0.0",
				"port":       8050,
				"https":      false,
				"https_port": 8443,
			},
			"notifications": map[string]interface{}{
				"channels": map[string]interface{}{},
				"events":   map[string]bool{},
			},
			"retention": map[string]interface{}{
				"type":  "days",
				"value": 30,
			},
			"directories": map[string]interface{}{
				"data_dir":   dataDir,
				"backup_dir": filepath.Join(dataDir, "backups"),
				"logs_dir":   filepath.Join(dataDir, "logs"),
			},
		}

		// Save default config
		configData, _ := json.MarshalIndent(defaultConfig, "", "  ")
		os.WriteFile(configFile, configData, 0644)

		return defaultConfig
	}

	var config map[string]interface{}
	json.Unmarshal(data, &config)
	return config
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}

func printUsage() {
	fmt.Println("NovaBackup v" + Version)
	fmt.Println()
	fmt.Println("Usage: novabackup <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  server, start   Start the web server (default)")
	fmt.Println("  install         Install as system service")
	fmt.Println("  remove          Remove system service")
	fmt.Println("  version         Show version")
	fmt.Println("  help            Show this help")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  novabackup server              # Start web server")
	fmt.Println("  novabackup install             # Install as service")
	fmt.Println("  novabackup version             # Show version")
	fmt.Println()
	fmt.Println("Web UI: http://localhost:8050")
	fmt.Println("Default: admin / admin123")
}
