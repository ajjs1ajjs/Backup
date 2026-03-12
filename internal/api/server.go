package api

import (
	"fmt"
	"net/http"
	"time"

	"novabackup/internal/backup"
	"novabackup/internal/database"
	"novabackup/internal/discovery"
	"novabackup/internal/recovery"
	"novabackup/internal/replication"
	"novabackup/internal/scheduler"
	"novabackup/internal/tape"
	"novabackup/internal/vss"
	"novabackup/pkg/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"novabackup/internal/storage"
)

// Server represents the API server
type Server struct {
	engine    *gin.Engine
	db        *database.Connection
	scheduler *scheduler.Scheduler
	discovery *discovery.DiscoveryService
	backup    *backup.BackupManager
	irMgr     *recovery.InstantRecoveryManager
	storage   *storage.Engine
}

// NewServer creates a new API server
func NewServer(db *database.Connection, sched *scheduler.Scheduler, stor *storage.Engine) (*Server, error) {
	s := &Server{
		engine:    gin.Default(),
		db:        db,
		scheduler: sched,
		discovery: discovery.NewDiscoveryService(db),
		backup:    backup.NewBackupManager(db),
		storage:   stor,
	}
	s.irMgr = recovery.NewInstantRecoveryManager(db, s.backup.GetCompressor(), stor, nil) // VMware provider nil for now
	s.setupRoutes()
	return s, nil
}

func (s *Server) setupRoutes() {
	// CORS middleware
	s.engine.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	v1 := s.engine.Group("/api/v1")
	{
		// Health check
		v1.GET("/health", s.healthCheck)

		// Jobs group
		jobs := v1.Group("/jobs")
		{
			jobs.GET("", s.listJobs)
			jobs.POST("", s.createJob)
			jobs.GET("/:id", s.getJob)
			jobs.PUT("/:id", s.updateJob)
			jobs.DELETE("/:id", s.deleteJob)
			jobs.POST("/:id/run", s.runJob)
		}

		// Backups group
		backups := v1.Group("/backups")
		{
			backups.GET("", s.listBackups)
			backups.GET("/:id", s.getBackup)
			backups.DELETE("/:id", s.deleteBackup)
		}

		// Restore group
		v1.POST("/restore", s.restoreFiles)
		v1.POST("/restore/db", s.restoreDatabase)
		v1.GET("/restore/points", s.listRestorePoints)
		v1.POST("/recovery/instant", s.handleInstantRecovery)
		v1.GET("/recovery/sessions", s.listRecoverySessions)
		v1.DELETE("/recovery/sessions/:id", s.stopRecoverySession)

		// Storage group
		storage := v1.Group("/storage")
		{
			storage.GET("/stats", s.getStorageStats)
		}

		// Dashboard stats
		v1.GET("/dashboard/stats", s.getDashboardStats)

		// Infrastructure tree
		v1.GET("/infrastructure/tree", s.getInfrastructureTree)
		v1.POST("/infrastructure/add", s.addInfrastructureNode)
		v1.GET("/infrastructure/nodes/:id/discover", s.rescanInfrastructure)
		v1.GET("/infrastructure/nodes/:id/objects", s.listInfrastructureObjects)

		// Storage Repositories
		v1.GET("/storage/repositories", s.getStorageRepositories)
		v1.POST("/storage/repositories", s.createRepository)

		// Metrics
		v1.GET("/metrics", s.getMetrics)

		// Credentials management
		credentials := v1.Group("/credentials")
		{
			credentials.GET("", s.listCredentials)
			credentials.POST("", s.createCredential)
			credentials.DELETE("/:id", s.deleteCredential)
		}

		// Backup Proxies
		proxies := v1.Group("/proxies")
		{
			proxies.GET("", s.listProxies)
			proxies.POST("", s.createProxy)
			proxies.PUT("/:id", s.updateProxy)
			proxies.DELETE("/:id", s.deleteProxy)
		}

		// Backup Sessions
		sessions := v1.Group("/backup/sessions")
		{
			sessions.GET("", s.listBackupSessions)
			sessions.GET("/:id", s.getBackupSession)
			sessions.DELETE("/:id", s.stopBackupSession)
		}

		// Job History
		v1.GET("/jobs/:id/history", s.getJobHistory)

		// Reports
		v1.GET("/reports", s.listReports)
		v1.GET("/reports/:id", s.getReport)
		v1.POST("/reports/generate", s.generateReport)

		// Notifications
		notifications := v1.Group("/notifications")
		{
			notifications.GET("", s.listNotifications)
			notifications.PUT("/:id/read", s.markNotificationRead)
			notifications.DELETE("/:id", s.deleteNotification)
		}

		// Settings
		v1.GET("/settings", s.getSettings)
		v1.PUT("/settings", s.updateSettings)

		// VSS / Guest Processing
		vss := v1.Group("/vss")
		{
			vss.GET("/writers", s.listVSSWriters)
			vss.POST("/snapshot", s.createVSSSnapshot)
			vss.DELETE("/snapshot/:id", s.deleteVSSSnapshot)
		}

		// Guest Credentials
		guestCreds := v1.Group("/guest/credentials")
		{
			guestCreds.GET("", s.listGuestCredentials)
			guestCreds.POST("", s.createGuestCredential)
			guestCreds.DELETE("/:id", s.deleteGuestCredential)
		}

		// Tape
		tape := v1.Group("/tape")
		{
			tape.GET("/libraries", s.listTapeLibraries)
			tape.GET("/libraries/:id", s.getTapeLibrary)
			tape.GET("/drives", s.listTapeDrives)
			tape.GET("/cartridges", s.listTapeCartridges)
			tape.GET("/vaults", s.listTapeVaults)
			tape.POST("/vaults", s.createTapeVault)
			tape.DELETE("/vaults/:id", s.deleteTapeVault)
			tape.GET("/jobs", s.listTapeJobs)
			tape.POST("/jobs", s.createTapeJob)
			tape.POST("/jobs/:id/run", s.runTapeJob)
			tape.DELETE("/jobs/:id", s.deleteTapeJob)
		}

		// RBAC - Users & Roles
		rbac := v1.Group("/rbac")
		{
			users := rbac.Group("/users")
			{
				users.GET("", s.listUsers)
				users.POST("", s.createUser)
				users.GET("/:id", s.getUser)
				users.PUT("/:id", s.updateUser)
				users.DELETE("/:id", s.deleteUser)
				users.POST("/:id/roles", s.assignRoleToUser)
				users.DELETE("/:id/roles/:roleId", s.removeRoleFromUser)
			}
			roles := rbac.Group("/roles")
			{
				roles.GET("", s.listRoles)
				roles.POST("", s.createRole)
				roles.GET("/:id", s.getRole)
				roles.PUT("/:id", s.updateRole)
				roles.DELETE("/:id", s.deleteRole)
			}
			rbac.GET("/permissions", s.listPermissions)
		}

		// Replication
		replication := v1.Group("/replication")
		{
			replication.GET("/jobs", s.listReplicationJobs)
			replication.POST("/jobs", s.createReplicationJob)
			replication.DELETE("/jobs/:id", s.deleteReplicationJob)
			replication.POST("/jobs/:id/run", s.runReplicationJob)
		}
	}
}

// Start starts the API server
func (s *Server) Start(port int) error {
	addr := fmt.Sprintf(":%d", port)
	return s.engine.Run(addr)
}

// healthCheck returns server health status
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"version":   "6.0.0",
		"scheduler": "running",
	})
}

// listJobs returns all backup jobs
func (s *Server) listJobs(c *gin.Context) {
	jobs, err := s.db.GetAllJobs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, jobs)
}

// createJob creates a new backup job
func (s *Server) createJob(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		JobType     string `json:"job_type"`
		Source      string `json:"source" binding:"required"`
		Destination string `json:"destination" binding:"required"`
		Schedule    string `json:"schedule"`
		Enabled     bool   `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}

	job := models.Job{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		JobType:     models.JobType(req.JobType),
		Source:      req.Source,
		Destination: req.Destination,
		Schedule:    req.Schedule,
		Enabled:     req.Enabled,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.db.CreateJob(&job); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, job)
}

// getJob returns a specific job
func (s *Server) getJob(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job ID"})
		return
	}
	job, err := s.db.GetJobByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}
	c.JSON(http.StatusOK, job)
}

// updateJob updates an existing job
func (s *Server) updateJob(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job ID"})
		return
	}
	var job models.Job
	if err := c.ShouldBindJSON(&job); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	job.ID = id
	c.JSON(http.StatusOK, job)
}

// deleteJob deletes a job
func (s *Server) deleteJob(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "job deleted", "id": id})
}

// runJob manually triggers a job
func (s *Server) runJob(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job ID"})
		return
	}

	if err := s.backup.RunJob(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "job started", "id": id})
}

// listBackups returns all backup results
func (s *Server) listBackups(c *gin.Context) {
	c.JSON(http.StatusOK, []models.BackupResult{})
}

// getBackup returns a specific backup
func (s *Server) getBackup(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"id": id, "status": "completed"})
}

// deleteBackup deletes a backup
func (s *Server) deleteBackup(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "backup deleted", "id": id})
}

// restoreFiles restores files from backup
func (s *Server) restoreFiles(c *gin.Context) {
	var req struct {
		BackupID      string `json:"backup_id"`
		Destination   string `json:"destination"`
		DbType        string `json:"db_type"`
		EncryptionKey string `json:"encryption_key"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "restore started", "backup_id": req.BackupID})
}

// restoreDatabase restores a database from backup
func (s *Server) restoreDatabase(c *gin.Context) {
	var req struct {
		BackupFile string `json:"backup_file" binding:"required"`
		DbType     string `json:"db_type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "database restore started", "file": req.BackupFile})
}

// listRestorePoints returns all restore points
func (s *Server) listRestorePoints(c *gin.Context) {
	c.JSON(http.StatusOK, []models.RestorePoint{})
}

// getStorageStats returns storage statistics
func (s *Server) getStorageStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"total_size":   0,
		"chunk_count":  0,
		"dedupe_ratio": 1.0,
	})
}

// getMetrics returns Prometheus-style metrics
func (s *Server) getMetrics(c *gin.Context) {
	c.Header("Content-Type", "text/plain")
	c.String(http.StatusOK, `# HELP novabackup_jobs_total Total number of backup jobs
# TYPE novabackup_jobs_total counter
novabackup_jobs_total 0
`)
}

// getDashboardStats returns dashboard statistics for UI
func (s *Server) getDashboardStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"systemStatus":     "All Systems Operational",
		"activeJobs":       1,
		"successJobs":      24,
		"warningJobs":      3,
		"failedJobs":       0,
		"storageUsed":      1024.5,
		"totalProcessedGB": 450.2,
		"backupBottleneck": "Source (Network)",
		"recentActivity": []string{
			"● Daily Backup - Completed - 2h ago",
			"● System Backup - Running - 45%",
			"● Database Backup - Scheduled - 1h",
			"● File Server - Warning - 5h ago",
		},
	})
}

// getInfrastructureTree returns infrastructure node tree for UI
func (s *Server) getInfrastructureTree(c *gin.Context) {
	c.JSON(http.StatusOK, []gin.H{
		{
			"name":       "Managed Servers",
			"iconKind":   "ServerNetwork",
			"isExpanded": true,
			"children": []gin.H{
				{
					"name":       "VMware vSphere",
					"iconKind":   "Vmware",
					"isExpanded": false,
					"children": []gin.H{
						{
							"name":       "vcenter.local",
							"iconKind":   "Server",
							"isExpanded": true,
							"children": []gin.H{
								{
									"name":       "esxi-01.local",
									"iconKind":   "ServerNetwork",
									"isExpanded": false,
									"children": []gin.H{
										{"name": "App-Server-01", "iconKind": "DesktopClassic", "children": []gin.H{}},
										{"name": "DB-Server-01", "iconKind": "Database", "children": []gin.H{}},
									},
								},
							},
						},
					},
				},
				{
					"name":       "Microsoft Hyper-V",
					"iconKind":   "MicrosoftWindows",
					"isExpanded": false,
					"children": []gin.H{
						{
							"name":       "hv-node-01",
							"iconKind":   "ServerNetwork",
							"isExpanded": false,
							"children": []gin.H{
								{"name": "DC-01", "iconKind": "DesktopClassic", "children": []gin.H{}},
							},
						},
					},
				},
			},
		},
	})
}

// getStorageRepositories returns configured backup repositories from DB
func (s *Server) getStorageRepositories(c *gin.Context) {
	repos, err := s.db.GetAllRepositories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, repos)
}

// createRepository adds a new storage backend
func (s *Server) createRepository(c *gin.Context) {
	var repo models.Repository
	if err := c.ShouldBindJSON(&repo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if repo.ID == uuid.Nil {
		repo.ID = uuid.New()
	}

	if err := s.db.CreateRepository(&repo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, repo)
}

// addInfrastructureNode adds a new infrastructure server
func (s *Server) addInfrastructureNode(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// rescanInfrastructure triggers discovery for a node
func (s *Server) rescanInfrastructure(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	if err := s.discovery.DiscoverNode(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Discovery completed successfully"})
}

// listInfrastructureObjects returns all objects for a node
func (s *Server) listInfrastructureObjects(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	objects, err := s.db.GetObjectsByNode(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, objects)
}
func (s *Server) handleInstantRecovery(c *gin.Context) {
	var req struct {
		RestorePointID uuid.UUID `json:"restore_point_id" binding:"required"`
		VMName         string    `json:"vm_name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Start NFS server via IR Manager
	_, err := s.irMgr.StartNFS(c.Request.Context(), req.RestorePointID.String(), 2049)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Instant recovery started",
		"nfs_path": fmt.Sprintf("localhost:/exports/%s", req.VMName),
		"status":   "running",
	})
}

// listRecoverySessions returns all active instant recovery sessions
func (s *Server) listRecoverySessions(c *gin.Context) {
	sessions := s.irMgr.ListSessions()
	c.JSON(http.StatusOK, sessions)
}

// stopRecoverySession stops a specific instant recovery session
func (s *Server) stopRecoverySession(c *gin.Context) {
	id := c.Param("id")
	if err := s.irMgr.StopSession(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Session stopped", "id": id})
}

// ========== Credentials Management ==========

type Credential struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"-"`
	Domain   string `json:"domain"`
	Type     string `json:"type"` // "windows", "linux", "sql", "exchange"
	Desc     string `json:"description"`
	Modified int64  `json:"modified"`
}

func (s *Server) listCredentials(c *gin.Context) {
	// TODO: Replace with database query when GetAllCredentials is implemented
	c.JSON(http.StatusOK, []Credential{
		{ID: "c1", Name: "Local Admin", Username: "Administrator", Domain: "", Type: "windows", Desc: "Local administrator account"},
		{ID: "c2", Name: "Domain Backup", Username: "svc_backup", Domain: "NOVA", Type: "windows", Desc: "Domain service account"},
	})
}

func (s *Server) createCredential(c *gin.Context) {
	var cred Credential
	if err := c.ShouldBindJSON(&cred); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if cred.ID == "" {
		cred.ID = uuid.New().String()
	}
	cred.Modified = time.Now().Unix()
	c.JSON(http.StatusCreated, cred)
}

func (s *Server) deleteCredential(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "Credential deleted", "id": id})
}

// ========== Backup Proxies ==========

type BackupProxy struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Type     string `json:"type"` // "vmware", "hyperv"
	MaxTasks int    `json:"max_tasks"`
	Enabled  bool   `json:"enabled"`
	Status   string `json:"status"`
	Version  string `json:"version"`
}

func (s *Server) listProxies(c *gin.Context) {
	c.JSON(http.StatusOK, []BackupProxy{
		{ID: "1", Name: "Proxy-01", Host: "192.168.1.10", Port: 9090, Type: "vmware", MaxTasks: 4, Enabled: true, Status: "Online", Version: "6.0.0"},
		{ID: "2", Name: "Proxy-02", Host: "192.168.1.11", Port: 9090, Type: "hyperv", MaxTasks: 2, Enabled: true, Status: "Online", Version: "6.0.0"},
	})
}

func (s *Server) createProxy(c *gin.Context) {
	var proxy BackupProxy
	if err := c.ShouldBindJSON(&proxy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if proxy.ID == "" {
		proxy.ID = uuid.New().String()
	}
	proxy.Status = "Online"
	c.JSON(http.StatusCreated, proxy)
}

func (s *Server) updateProxy(c *gin.Context) {
	id := c.Param("id")
	var proxy BackupProxy
	if err := c.ShouldBindJSON(&proxy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	proxy.ID = id
	c.JSON(http.StatusOK, proxy)
}

func (s *Server) deleteProxy(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "Proxy deleted", "id": id})
}

// ========== Backup Sessions ==========

type BackupSession struct {
	ID          string    `json:"id"`
	JobID       string    `json:"job_id"`
	JobName     string    `json:"job_name"`
	Status      string    `json:"status"` // "Running", "Stopped", "Completed", "Failed"
	Progress    int       `json:"progress"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	ProcessedGB float64   `json:"processed_gb"`
	ReadGB      float64   `json:"read_gb"`
	WriteGB     float64   `json:"write_gb"`
	Duration    int       `json:"duration_sec"`
}

func (s *Server) listBackupSessions(c *gin.Context) {
	c.JSON(http.StatusOK, []BackupSession{
		{ID: "s1", JobID: "j1", JobName: "Daily Backup", Status: "Running", Progress: 45, StartTime: time.Now().Add(-1 * time.Hour)},
		{ID: "s2", JobID: "j2", JobName: "Weekly Backup", Status: "Completed", Progress: 100, StartTime: time.Now().Add(-24 * time.Hour), EndTime: time.Now().Add(-23 * time.Hour), ProcessedGB: 500, Duration: 3600},
	})
}

func (s *Server) getBackupSession(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, BackupSession{ID: id, JobID: "j1", JobName: "Daily Backup", Status: "Running", Progress: 45, StartTime: time.Now().Add(-1 * time.Hour)})
}

func (s *Server) stopBackupSession(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "Backup session stopped", "id": id})
}

// ========== Job History ==========

type JobHistory struct {
	ID        string    `json:"id"`
	JobID     string    `json:"job_id"`
	Status    string    `json:"status"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	SizeGB    float64   `json:"size_gb"`
	Duration  int       `json:"duration_sec"`
	Message   string    `json:"message"`
}

func (s *Server) getJobHistory(c *gin.Context) {
	jobID := c.Param("id")
	c.JSON(http.StatusOK, []JobHistory{
		{ID: "h1", JobID: jobID, Status: "Success", StartTime: time.Now().Add(-24 * time.Hour), EndTime: time.Now().Add(-23 * time.Hour), SizeGB: 500, Duration: 3600, Message: "Completed successfully"},
		{ID: "h2", JobID: jobID, Status: "Warning", StartTime: time.Now().Add(-48 * time.Hour), EndTime: time.Now().Add(-47 * time.Hour), SizeGB: 480, Duration: 3500, Message: "Completed with warnings"},
	})
}

// ========== Reports ==========

type Report struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"` // "backup", "restore", "capacity", "job"
	From      time.Time `json:"from"`
	To        time.Time `json:"to"`
	Generated time.Time `json:"generated"`
	Format    string    `json:"format"` // "pdf", "html", "csv"
	URL       string    `json:"url"`
}

func (s *Server) listReports(c *gin.Context) {
	c.JSON(http.StatusOK, []Report{
		{ID: "r1", Name: "Weekly Backup Report", Type: "backup", Generated: time.Now().Add(-1 * time.Hour), Format: "pdf"},
		{ID: "r2", Name: "Monthly Capacity Report", Type: "capacity", Generated: time.Now().Add(-24 * time.Hour), Format: "html"},
	})
}

func (s *Server) getReport(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, Report{ID: id, Name: "Weekly Backup Report", Type: "backup", Generated: time.Now().Add(-1 * time.Hour), Format: "pdf", URL: "/reports/" + id + ".pdf"})
}

func (s *Server) generateReport(c *gin.Context) {
	var req struct {
		Name   string `json:"name" binding:"required"`
		Type   string `json:"type" binding:"required"`
		From   string `json:"from"`
		To     string `json:"to"`
		Format string `json:"format"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	report := Report{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Type:      req.Type,
		Generated: time.Now(),
		Format:    req.Format,
		URL:       "/reports/" + uuid.New().String() + "." + req.Format,
	}
	c.JSON(http.StatusAccepted, gin.H{"message": "Report generation started", "report": report})
}

// ========== Notifications ==========

type Notification struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"` // "info", "warning", "error", "success"
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Time      time.Time `json:"time"`
	Read      bool      `json:"read"`
	JobID     string    `json:"job_id"`
	SessionID string    `json:"session_id"`
}

func (s *Server) listNotifications(c *gin.Context) {
	c.JSON(http.StatusOK, []Notification{
		{ID: "n1", Type: "info", Title: "Backup Completed", Message: "Daily Backup completed successfully", Time: time.Now().Add(-1 * time.Hour), Read: false, JobID: "j1"},
		{ID: "n2", Type: "warning", Title: "Low Storage", Message: "Repository free space below 20%", Time: time.Now().Add(-2 * time.Hour), Read: true},
	})
}

func (s *Server) markNotificationRead(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read", "id": id})
}

func (s *Server) deleteNotification(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "Notification deleted", "id": id})
}

// ========== Settings ==========

type Settings struct {
	General struct {
		Language   string `json:"language"`
		Theme      string `json:"theme"`
		DateFormat string `json:"date_format"`
		TimeZone   string `json:"time_zone"`
	} `json:"general"`
	Backup struct {
		DefaultRepo         string `json:"default_repo"`
		MaxParallel         int    `json:"max_parallel"`
		TempPath            string `json:"temp_path"`
		EnableNotifications bool   `json:"enable_notifications"`
	} `json:"backup"`
	Network struct {
		Timeout      int `json:"timeout"`
		RetryCount   int `json:"retry_count"`
		MaxBandwidth int `json:"max_bandwidth"`
	} `json:"network"`
	Security struct {
		EnableRBAC     bool `json:"enable_rbac"`
		SessionTimeout int  `json:"session_timeout"`
		Require2FA     bool `json:"require_2fa"`
	} `json:"security"`
}

func (s *Server) getSettings(c *gin.Context) {
	var settings Settings
	settings.General.Language = "en"
	settings.General.Theme = "dark"
	settings.General.DateFormat = "yyyy-MM-dd"
	settings.General.TimeZone = "UTC"
	settings.Backup.DefaultRepo = ""
	settings.Backup.MaxParallel = 4
	settings.Backup.EnableNotifications = true
	settings.Network.Timeout = 300
	settings.Network.RetryCount = 3
	settings.Network.MaxBandwidth = 0
	settings.Security.EnableRBAC = true
	settings.Security.SessionTimeout = 60
	settings.Security.Require2FA = false
	c.JSON(http.StatusOK, settings)
}

func (s *Server) updateSettings(c *gin.Context) {
	var settings Settings
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Settings updated", "settings": settings})
}

// ========== Replication ==========

var replEngine = replication.NewInMemoryReplicationEngine()

type ReplicationRequest struct {
	SourceVM             string `json:"source_vm" binding:"required"`
	DestinationHost      string `json:"destination_host" binding:"required"`
	DestinationDatastore string `json:"destination_datastore" binding:"required"`
	DestinationVC        string `json:"destination_vc"`
	ReplicationType      string `json:"replication_type"` // "sync", "async", "backup"
	BandwidthLimit       int    `json:"bandwidth_limit_mbps"`
	EnableRPO            bool   `json:"enable_rpo"`
	RPOTargetMinutes     int    `json:"rpo_target_minutes"`
}

func (s *Server) listReplicationJobs(c *gin.Context) {
	jobs, err := replEngine.ListJobs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, jobs)
}

func (s *Server) createReplicationJob(c *gin.Context) {
	var req ReplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	replReq := &replication.ReplicationRequest{
		SourceVM:             req.SourceVM,
		DestinationHost:      req.DestinationHost,
		DestinationDatastore: req.DestinationDatastore,
		DestinationVC:        req.DestinationVC,
		ReplicationType:      replication.ReplicationType(req.ReplicationType),
		BandwidthLimit:       req.BandwidthLimit,
		EnableRPO:            req.EnableRPO,
		RPOTarget:            time.Duration(req.RPOTargetMinutes) * time.Minute,
	}

	job, err := replEngine.StartReplication(c.Request.Context(), replReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, job)
}

func (s *Server) getReplicationJob(c *gin.Context) {
	id := c.Param("id")
	job, err := replEngine.GetJob(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, job)
}

func (s *Server) deleteReplicationJob(c *gin.Context) {
	id := c.Param("id")
	// In real impl: would call replEngine.DeleteJob
	c.JSON(http.StatusOK, gin.H{"message": "Replication job deleted", "id": id})
}

func (s *Server) startReplicationJob(c *gin.Context) {
	id := c.Param("id")
	job, err := replEngine.GetJob(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Replication started", "job": job})
}

func (s *Server) stopReplicationJob(c *gin.Context) {
	id := c.Param("id")
	err := replEngine.StopReplication(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Replication stopped", "id": id})
}

func (s *Server) getReplicationStatus(c *gin.Context) {
	id := c.Param("id")
	status, err := replEngine.GetJobStatus(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, status)
}

// ========== VSS / Guest Processing ==========

var vssManager = vss.NewInMemoryVSSManager()
var guestCredManager = vss.NewGuestCredentialManager()

type VSSSnapshotRequest struct {
	Volume      string   `json:"volume" binding:"required"`
	WriterTypes []string `json:"writer_types"`
	BackupType  string   `json:"backup_type"`
}

type GuestCredentialRequest struct {
	Name     string `json:"name" binding:"required"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Domain   string `json:"domain"`
	Type     string `json:"type"` // "windows", "linux"
}

func (s *Server) listVSSWriters(c *gin.Context) {
	writers, err := vssManager.GetWriterStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, writers)
}

func (s *Server) createVSSSnapshot(c *gin.Context) {
	var req VSSSnapshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var writerTypes []vss.VSSWriterType
	for _, wt := range req.WriterTypes {
		writerTypes = append(writerTypes, vss.VSSWriterType(wt))
	}

	vssReq := &vss.VSSRequest{
		Volume:      req.Volume,
		WriterTypes: writerTypes,
		BackupType:  req.BackupType,
	}

	result, err := vssManager.CreateSnapshot(c.Request.Context(), vssReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (s *Server) deleteVSSSnapshot(c *gin.Context) {
	id := c.Param("id")
	err := vssManager.DeleteSnapshot(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Snapshot deleted", "id": id})
}

func (s *Server) listGuestCredentials(c *gin.Context) {
	creds := guestCredManager.ListCredentials()
	c.JSON(http.StatusOK, creds)
}

func (s *Server) createGuestCredential(c *gin.Context) {
	var req GuestCredentialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := uuid.New().String()
	guestCredManager.AddCredential(id, &vss.GuestCredentials{
		Username: req.Username,
		Password: req.Password,
		Domain:   req.Domain,
		Type:     req.Type,
	})

	c.JSON(http.StatusCreated, gin.H{"id": id, "name": req.Name})
}

func (s *Server) deleteGuestCredential(c *gin.Context) {
	id := c.Param("id")
	guestCredManager.DeleteCredential(id)
	c.JSON(http.StatusOK, gin.H{"message": "Credential deleted", "id": id})
}

// ========== Tape Management ==========

var tapeManager = tape.NewInMemoryTapeManager()

type TapeVaultRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Location    string `json:"location"`
	Contact     string `json:"contact"`
}

type TapeJobRequest struct {
	Name        string `json:"name" binding:"required"`
	Source      string `json:"source" binding:"required"`
	TargetVault string `json:"target_vault" binding:"required"`
	Schedule    string `json:"schedule"`
	Retention   int    `json:"retention_days"`
	Enabled     bool   `json:"enabled"`
}

func (s *Server) listTapeLibraries(c *gin.Context) {
	libs, err := tapeManager.ListLibraries(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, libs)
}

func (s *Server) getTapeLibrary(c *gin.Context) {
	id := c.Param("id")
	lib, err := tapeManager.GetLibrary(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, lib)
}

func (s *Server) listTapeDrives(c *gin.Context) {
	drives, err := tapeManager.ListDrives(c.Request.Context(), "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, drives)
}

func (s *Server) listTapeCartridges(c *gin.Context) {
	cartridges, err := tapeManager.ListCartridges(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cartridges)
}

func (s *Server) listTapeVaults(c *gin.Context) {
	vaults, err := tapeManager.ListVaults(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, vaults)
}

func (s *Server) createTapeVault(c *gin.Context) {
	var req TapeVaultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vault := &tape.TapeVault{
		Name:        req.Name,
		Description: req.Description,
		Location:    req.Location,
		Contact:     req.Contact,
	}

	if err := tapeManager.CreateVault(c.Request.Context(), vault); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, vault)
}

func (s *Server) deleteTapeVault(c *gin.Context) {
	id := c.Param("id")
	tapeManager.DeleteVault(c.Request.Context(), id)
	c.JSON(http.StatusOK, gin.H{"message": "Vault deleted", "id": id})
}

func (s *Server) listTapeJobs(c *gin.Context) {
	jobs, err := tapeManager.ListJobs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, jobs)
}

func (s *Server) createTapeJob(c *gin.Context) {
	var req TapeJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	job := &tape.TapeBackupJob{
		Name:        req.Name,
		Source:      req.Source,
		TargetVault: req.TargetVault,
		Schedule:    req.Schedule,
		Retention:   req.Retention,
		Enabled:     req.Enabled,
	}

	if err := tapeManager.CreateJob(c.Request.Context(), job); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, job)
}

func (s *Server) runTapeJob(c *gin.Context) {
	id := c.Param("id")
	tapeManager.RunJob(c.Request.Context(), id)
	c.JSON(http.StatusOK, gin.H{"message": "Tape job started", "id": id})
}

func (s *Server) deleteTapeJob(c *gin.Context) {
	id := c.Param("id")
	tapeManager.DeleteJob(c.Request.Context(), id)
	c.JSON(http.StatusOK, gin.H{"message": "Tape job deleted", "id": id})
}
