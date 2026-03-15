// NovaBackup - Modern Web-Based Backup Platform
// Usage: novabackup [command]
// Commands: server, install, remove, start, stop

package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"novabackup/internal/api"
	"novabackup/internal/backup"
	"novabackup/internal/database"
	"novabackup/internal/restore"
	"novabackup/internal/scheduler"

	"github.com/gin-gonic/gin"
)

const (
	Version     = "7.0.0"
	DefaultPort = 8050
	ServiceName = "NovaBackup"
)

var (
	configPath    string
	dataDir       string
	webDir        string
	db            *database.Database
	backupEngine  *backup.BackupEngine
	restoreEngine *restore.RestoreEngine
	jobScheduler  *scheduler.Scheduler
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
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	fmt.Println("✓ Database initialized")

	// Set global DB for API
	api.DB = db

	// Initialize backup engine
	backupEngine = backup.NewBackupEngine(dataDir)
	fmt.Println("✓ Backup engine initialized")

	// Initialize restore engine
	restoreEngine = restore.NewRestoreEngine(dataDir)
	fmt.Println("✓ Restore engine initialized")

	// Set engines for API
	api.BackupEngine = backupEngine
	api.RestoreEngine = restoreEngine

	// Initialize scheduler
	jobScheduler = scheduler.NewScheduler(db)
	if err := jobScheduler.Start(); err != nil {
		log.Printf("Warning: Failed to start scheduler: %v", err)
	}
	fmt.Println("✓ Scheduler started")

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

	// API Routes
	apiGroup := router.Group("/api")
	{
		// Health
		apiGroup.GET("/health", api.GetHealth)

		// Auth
		apiGroup.POST("/auth/login", api.Login)
		apiGroup.POST("/auth/logout", api.Logout)

		// Jobs
		apiGroup.GET("/jobs", api.ListJobs)
		apiGroup.POST("/jobs", api.CreateJob)
		apiGroup.PUT("/jobs/:id", api.UpdateJob)
		apiGroup.DELETE("/jobs/:id", api.DeleteJob)
		apiGroup.POST("/jobs/:id/run", api.RunJob)

		// Backup
		apiGroup.POST("/backup/run", api.RunBackup)
		apiGroup.GET("/backup/sessions", api.ListSessions)
		apiGroup.GET("/backup/sessions/:id", api.GetSession)

		// Restore
		apiGroup.GET("/restore/points", api.ListRestorePoints)
		apiGroup.POST("/restore/files", api.RestoreFiles)
		apiGroup.POST("/restore/database", api.RestoreDatabase)

		// Storage
		apiGroup.GET("/storage/repos", api.ListRepos)
		apiGroup.POST("/storage/repos", api.CreateRepo)
		apiGroup.DELETE("/storage/repos/:id", api.DeleteRepo)

		// Settings
		apiGroup.GET("/settings", api.GetSettings)
		apiGroup.PUT("/settings", api.UpdateSettings)
	}

	// Serve web UI from disk
	router.GET("/", func(c *gin.Context) {
		indexFile := filepath.Join(webDir, "index.html")
		c.File(indexFile)
	})

	router.GET("/assets/:file", func(c *gin.Context) {
		file := filepath.Join(webDir, "assets", c.Param("file"))
		c.File(file)
	})

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
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func initDirectories() {
	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)

	// Check if running from Program Files
	if strings.Contains(exeDir, "Program Files") {
		dataDir = filepath.Join("C:\\ProgramData", "NovaBackup")
		configPath = filepath.Join(dataDir, "config")
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
	os.MkdirAll(webDir, 0755)
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

func installService() {
	fmt.Println("Service installation not yet implemented for this platform")
	fmt.Println("Please run novabackup server to start manually")
}

func removeService() {
	fmt.Println("Service removal not yet implemented for this platform")
}
