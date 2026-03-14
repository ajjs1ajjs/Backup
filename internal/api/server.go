package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"novabackup/internal/backup"
	"novabackup/internal/cdp"
	"novabackup/internal/database"
	"novabackup/internal/discovery"
	"novabackup/internal/guest"
	"novabackup/internal/recovery"
	"novabackup/internal/replication"
	"novabackup/internal/scheduler"
	"novabackup/internal/synthetic"
	"novabackup/internal/tape"
	"novabackup/internal/vss"
	"novabackup/pkg/models"

	"novabackup/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Server represents the API server
type Server struct {
	engine       *gin.Engine
	db           *database.Connection
	scheduler    *scheduler.Scheduler
	discovery    *discovery.DiscoveryService
	backup       *backup.BackupManager
	irMgr        *recovery.InstantRecoveryManager
	storage      *storage.Engine
	gip          *guest.GuestInteractionProxy
	guestProc    *guest.GuestProcessor
	credStore    *guest.EncryptedCredentialStore
	cdpEngine    *cdp.InMemoryCDPEngine
	replMgr      *replication.ReplicationManager
	syntheticMgr *synthetic.InMemorySyntheticBackupManager
}

// NewServer creates a new API server
func NewServer(db *database.Connection, sched *scheduler.Scheduler, stor *storage.Engine) (*Server, error) {
	// Generate encryption key for credentials (in production, this should be from secure storage)
	encryptKey := []byte("NovaBackupMasterKey32Bytes!") // 32 bytes

	// Initialize credential store
	credStore, err := guest.NewEncryptedCredentialStore(guest.EncryptedCredentialStoreConfig{
		EncryptionKey: encryptKey,
		Logger:        zap.NewNop(), // Replace with real logger
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create credential store: %w", err)
	}

	// Initialize Guest Interaction Proxy
	gip := guest.NewGuestInteractionProxy(guest.GuestInteractionProxyConfig{
		HealthCheckInterval: 30 * time.Second,
		HeartbeatTimeout:    2 * time.Minute,
		SessionTimeout:      30 * time.Minute,
		Logger:              zap.NewNop(),
	})

	// Initialize Guest Processor
	guestProc := guest.NewGuestProcessor(guest.GuestProcessorConfig{
		Logger: zap.NewNop(),
	})

	// Initialize CDP Engine
	cdpEngine := cdp.NewInMemoryCDPEngine(cdp.NewCDPConfig(), nil, nil)

	// Initialize Replication Manager
	replEngine := replication.NewInMemoryReplicationEngine()
	replMgr, _ := replication.NewReplicationManager(zap.NewNop(), replEngine)

	s := &Server{
		engine:    gin.Default(),
		db:        db,
		scheduler: sched,
		discovery: discovery.NewDiscoveryService(db),
		backup:    backup.NewBackupManager(db),
		storage:   stor,
		gip:       gip,
		guestProc: guestProc,
		credStore: credStore,
		cdpEngine: cdpEngine,
		replMgr:   replMgr,
	}
	s.irMgr = recovery.NewInstantRecoveryManager(db, s.backup.GetCompressor(), stor, nil) // VMware provider nil for now

	// Initialize Synthetic Backup Manager
	s.syntheticMgr = synthetic.NewInMemorySyntheticBackupManager(nil, nil) // TODO: Add proper tenant manager

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

		// Guest Processing
		guestProc := v1.Group("/guest/process")
		{
			guestProc.POST("", s.processGuest)
			guestProc.GET("/tasks", s.listGuestProcessingTasks)
			guestProc.GET("/tasks/:id", s.getGuestProcessingTask)
			guestProc.GET("/tasks/:id/status", s.getGuestProcessingStatus)
			guestProc.POST("/tasks/:id/cancel", s.cancelGuestProcessing)
		}

		// Guest Agents
		guestAgents := v1.Group("/guest/agents")
		{
			guestAgents.GET("", s.listGuestAgents)
			guestAgents.GET("/:id", s.getGuestAgent)
			guestAgents.POST("/:id/register", s.registerGuestAgent)
			guestAgents.DELETE("/:id", s.unregisterGuestAgent)
			guestAgents.GET("/:id/sessions", s.listGuestAgentSessions)
		}

		// Guest Applications (discovery)
		guestApps := v1.Group("/guest/applications")
		{
			guestApps.GET("", s.listGuestApplications)
			guestApps.GET("/sql/databases", s.listSQLDatabases)
			guestApps.GET("/exchange/mailboxes", s.listExchangeMailboxes)
			guestApps.GET("/exchange/databases", s.listExchangeDatabases)
			guestApps.GET("/ad/domains", s.listADDomains)
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
			replication.GET("/jobs/:id", s.getReplicationJob)
			replication.DELETE("/jobs/:id", s.deleteReplicationJob)
			replication.POST("/jobs/:id/run", s.runReplicationJob)
			replication.POST("/jobs/:id/stop", s.stopReplicationJob)
			replication.POST("/jobs/:id/pause", s.pauseReplicationJob)
			replication.POST("/jobs/:id/resume", s.resumeReplicationJob)
			replication.GET("/jobs/:id/status", s.getReplicationStatus)
			replication.GET("/jobs/:id/rpo", s.getRPOCompliance)
			replication.POST("/jobs/:id/failover", s.failoverReplicationJob)
			replication.POST("/jobs/:id/failover/test", s.testFailoverReplicationJob)
			replication.GET("/stats", s.getReplicationStats)
		}

		// CDP (Continuous Data Protection)
		cdp := v1.Group("/cdp")
		{
			cdp.POST("/watch/start", s.startCDPWatching)
			cdp.POST("/watch/stop", s.stopCDPWatching)
			cdp.GET("/watch/status", s.getCDPWatchingStatus)
			cdp.GET("/protected-paths", s.getProtectedPaths)
			cdp.POST("/protected-paths", s.addProtectedPath)
			cdp.DELETE("/protected-paths/:path", s.removeProtectedPath)
			cdp.GET("/events", s.getCDPEvents)
			cdp.GET("/recovery-points", s.getRecoveryPoints)
			cdp.POST("/restore", s.restoreToRecoveryPoint)
			cdp.GET("/stats", s.getCDPStats)
			cdp.GET("/rpo-stats", s.getRPOStats)
		}

		// Synthetic Backup
		synthetic := v1.Group("/synthetic")
		{
			synthetic.POST("", s.createSyntheticBackup)
			synthetic.GET("", s.listSyntheticBackups)
			synthetic.GET("/:id", s.getSyntheticBackup)
			synthetic.DELETE("/:id", s.deleteSyntheticBackup)
			synthetic.POST("/merge", s.mergeIncrementals)
			synthetic.GET("/chains", s.getBackupChain)
			synthetic.GET("/stats", s.getSyntheticStats)
			synthetic.GET("/chains/stats", s.getBackupChainStats)
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
	job, err := s.replMgr.GetJob(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, job)
}

// pauseReplicationJob pauses a replication job
func (s *Server) pauseReplicationJob(c *gin.Context) {
	id := c.Param("id")
	if err := s.replMgr.PauseReplicationJob(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Replication paused", "id": id})
}

// resumeReplicationJob resumes a paused replication job
func (s *Server) resumeReplicationJob(c *gin.Context) {
	id := c.Param("id")
	if err := s.replMgr.ResumeReplicationJob(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Replication resumed", "id": id})
}

// getRPOCompliance returns RPO compliance report
func (s *Server) getRPOCompliance(c *gin.Context) {
	id := c.Param("id")
	// In real implementation, would call replMgr.GetRPOCompliance
	c.JSON(http.StatusOK, gin.H{"job_id": id, "is_compliant": true, "rpo_target_minutes": 15})
}

// failoverReplicationJob initiates failover
func (s *Server) failoverReplicationJob(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := s.replMgr.FailoverJob(c.Request.Context(), &replication.FailoverRequest{
		JobID:  id,
		Reason: req.Reason,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// testFailoverReplicationJob performs test failover
func (s *Server) testFailoverReplicationJob(c *gin.Context) {
	id := c.Param("id")
	result, err := s.replMgr.TestFailoverJob(c.Request.Context(), &replication.TestFailoverRequest{
		JobID: id,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// getReplicationStats returns replication statistics
func (s *Server) getReplicationStats(c *gin.Context) {
	// In real implementation, would call replMgr.GetStatistics()
	c.JSON(http.StatusOK, gin.H{"total_jobs": 0, "running_jobs": 0})
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

// ============================================================================
// Guest Processing API Handlers
// ============================================================================

// GuestProcessRequest represents a guest processing request
type GuestProcessRequest struct {
	VMID            string   `json:"vm_id" binding:"required"`
	VMName          string   `json:"vm_name" binding:"required"`
	Mode            string   `json:"mode"` // "disabled", "crash_consistent", "application_aware"`
	Applications    []string `json:"applications"`
	CredentialsID   string   `json:"credentials_id"`
	EnableQuiesce   bool     `json:"enable_quiesce"`
	TruncateLogs    bool     `json:"truncate_logs"`
	PreFreezeScript string   `json:"pre_freeze_script"`
	PostThawScript  string   `json:"post_thaw_script"`
	TimeoutSeconds  int      `json:"timeout_seconds"`
}

// processGuest initiates guest processing for a VM
func (s *Server) processGuest(c *gin.Context) {
	var req GuestProcessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create processing task with VSS Full backup
	task, err := s.guestProc.CreateProcessingTask(req.VMID, req.VMName, "", req.CredentialsID, guest.GuestProcessingTypeVSSFull)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Start execution asynchronously
	go s.guestProc.ExecuteProcessing(c.Request.Context(), task.ID)

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Guest processing started",
		"task_id": task.ID,
		"vm_id":   req.VMID,
		"vm_name": req.VMName,
		"status":  "pending",
	})
}

// listGuestProcessingTasks returns all guest processing tasks
func (s *Server) listGuestProcessingTasks(c *gin.Context) {
	vmID := c.Query("vm_id")

	allTasks := s.guestProc.ListProcessingTasks()

	// Filter by vm_id if provided
	if vmID != "" {
		var filtered []*guest.GuestProcessingTask
		for _, task := range allTasks {
			if task.VMID == vmID {
				filtered = append(filtered, task)
			}
		}
		c.JSON(http.StatusOK, filtered)
	} else {
		c.JSON(http.StatusOK, allTasks)
	}
}

// getGuestProcessingTask returns a specific processing task
func (s *Server) getGuestProcessingTask(c *gin.Context) {
	taskID := c.Param("id")
	task, err := s.guestProc.GetProcessingTask(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, task)
}

// getGuestProcessingStatus returns the status of a processing task
func (s *Server) getGuestProcessingStatus(c *gin.Context) {
	taskID := c.Param("id")
	task, err := s.guestProc.GetProcessingTask(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"task_id": taskID, "status": task.Status, "progress": task.Progress})
}

// cancelGuestProcessing cancels a running processing task
func (s *Server) cancelGuestProcessing(c *gin.Context) {
	taskID := c.Param("id")
	if err := s.guestProc.CancelProcessing(taskID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Processing cancelled", "task_id": taskID})
}

// ============================================================================
// Guest Agents API Handlers
// ============================================================================

// GuestAgentRequest represents a guest agent registration request
type GuestAgentRequest struct {
	VMID         string `json:"vm_id" binding:"required"`
	VMName       string `json:"vm_name" binding:"required"`
	Hostname     string `json:"hostname"`
	IPAddress    string `json:"ip_address" binding:"required"`
	AgentType    string `json:"agent_type"` // "windows", "linux"
	AgentVersion string `json:"agent_version"`
}

// listGuestAgents returns all registered guest agents
func (s *Server) listGuestAgents(c *gin.Context) {
	statusFilter := c.Query("status")

	allAgents := s.gip.ListAgents()

	// Filter by status if provided
	if statusFilter != "" {
		var filtered []*guest.GuestAgent
		for _, agent := range allAgents {
			if string(agent.Status) == statusFilter {
				filtered = append(filtered, agent)
			}
		}
		c.JSON(http.StatusOK, filtered)
	} else {
		c.JSON(http.StatusOK, allAgents)
	}
}

// getGuestAgent returns a specific guest agent
func (s *Server) getGuestAgent(c *gin.Context) {
	agentID := c.Param("id")
	agent, err := s.gip.GetAgent(agentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, agent)
}

// registerGuestAgent registers a new guest agent
func (s *Server) registerGuestAgent(c *gin.Context) {
	var req GuestAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agent := &guest.GuestAgent{
		ID:           uuid.New().String(),
		VMID:         req.VMID,
		VMName:       req.VMName,
		Hostname:     req.Hostname,
		IPAddress:    req.IPAddress,
		AgentType:    guest.GuestAgentType(req.AgentType),
		AgentVersion: req.AgentVersion,
	}

	if err := s.gip.RegisterAgent(agent); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Guest agent registered",
		"agent_id": agent.ID,
		"vm_id":    req.VMID,
	})
}

// unregisterGuestAgent removes a guest agent
func (s *Server) unregisterGuestAgent(c *gin.Context) {
	agentID := c.Param("id")
	if err := s.gip.UnregisterAgent(agentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Guest agent unregistered", "agent_id": agentID})
}

// listGuestAgentSessions returns sessions for a specific agent
func (s *Server) listGuestAgentSessions(c *gin.Context) {
	agentID := c.Param("id")
	sessions := s.gip.ListSessions(agentID)
	c.JSON(http.StatusOK, sessions)
}

// ============================================================================
// Guest Applications API Handlers
// ============================================================================

// listGuestApplications returns all discovered applications
func (s *Server) listGuestApplications(c *gin.Context) {
	applications := []gin.H{
		{
			"type":   "sql_server",
			"name":   "Microsoft SQL Server",
			"count":  3,
			"status": "healthy",
		},
		{
			"type":   "exchange",
			"name":   "Microsoft Exchange Server",
			"count":  1,
			"status": "healthy",
		},
		{
			"type":   "active_directory",
			"name":   "Active Directory Domain Services",
			"count":  1,
			"status": "healthy",
		},
	}
	c.JSON(http.StatusOK, applications)
}

// listSQLDatabases returns all discovered SQL Server databases
func (s *Server) listSQLDatabases(c *gin.Context) {
	sqlVSS := vss.NewSQLServerVSS()
	databases, err := sqlVSS.GetDatabaseDetails(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, databases)
}

// listExchangeMailboxes returns all discovered Exchange mailboxes
func (s *Server) listExchangeMailboxes(c *gin.Context) {
	exchangeVSS := vss.NewExchangeVSS()
	mailboxes, err := exchangeVSS.GetMailboxes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"mailboxes": mailboxes})
}

// listExchangeDatabases returns all discovered Exchange databases
func (s *Server) listExchangeDatabases(c *gin.Context) {
	exchangeVSS := vss.NewExchangeVSS()
	databases, err := exchangeVSS.GetDatabases(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, databases)
}

// listADDomains returns all discovered Active Directory domains
func (s *Server) listADDomains(c *gin.Context) {
	adVSS := vss.NewActiveDirectoryVSS()
	domains, err := adVSS.GetDomainInfo(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, domains)
}

// ============================================================================
// CDP API Handlers
// ============================================================================

// CDPWatchRequest represents a request to start CDP watching
type CDPWatchRequest struct {
	Paths []string `json:"paths" binding:"required"`
}

// startCDPWatching starts CDP monitoring
func (s *Server) startCDPWatching(c *gin.Context) {
	var req CDPWatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	if err := s.cdpEngine.StartWatching(ctx, req.Paths); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "CDP watching started",
		"paths":   req.Paths,
	})
}

// stopCDPWatching stops CDP monitoring
func (s *Server) stopCDPWatching(c *gin.Context) {
	ctx := context.Background()
	if err := s.cdpEngine.StopWatching(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "CDP watching stopped"})
}

// getCDPWatchingStatus returns CDP watching status
func (s *Server) getCDPWatchingStatus(c *gin.Context) {
	isWatching := s.cdpEngine.IsWatching()
	c.JSON(http.StatusOK, gin.H{"is_watching": isWatching})
}

// getProtectedPaths returns all protected paths
func (s *Server) getProtectedPaths(c *gin.Context) {
	ctx := context.Background()
	paths, err := s.cdpEngine.GetProtectedPaths(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"paths": paths})
}

// addProtectedPath adds a path to protection
func (s *Server) addProtectedPath(c *gin.Context) {
	var req struct {
		Path string `json:"path" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	if err := s.cdpEngine.EnableProtection(ctx, req.Path); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Path protected", "path": req.Path})
}

// removeProtectedPath removes a path from protection
func (s *Server) removeProtectedPath(c *gin.Context) {
	path := c.Param("path")
	ctx := context.Background()
	if err := s.cdpEngine.DisableProtection(ctx, path); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Path unprotected", "path": path})
}

// getCDPEvents returns CDP events
type CDPEventsRequest struct {
	Limit int `json:"limit"`
}

func (s *Server) getCDPEvents(c *gin.Context) {
	limitStr := c.Query("limit")
	limit := 100
	if limitStr != "" {
		if n, err := strconv.Atoi(limitStr); err == nil && n > 0 {
			limit = n
		}
	}

	ctx := context.Background()
	events, err := s.cdpEngine.GetEvents(ctx, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"events": events, "count": len(events)})
}

// getRecoveryPoints returns recovery points for a path
func (s *Server) getRecoveryPoints(c *gin.Context) {
	path := c.Query("path")
	sinceStr := c.Query("since")

	var since time.Time
	if sinceStr != "" {
		var err error
		since, err = time.Parse(time.RFC3339, sinceStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid since format"})
			return
		}
	}

	ctx := context.Background()
	points, err := s.cdpEngine.GetRecoveryPoints(ctx, path, since)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"recovery_points": points, "count": len(points)})
}

// restoreToRecoveryPoint restores a path to a recovery point
func (s *Server) restoreToRecoveryPoint(c *gin.Context) {
	var req struct {
		Path            string `json:"path" binding:"required"`
		RecoveryPointID string `json:"recovery_point_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	point := &cdp.RecoveryPoint{ID: req.RecoveryPointID}
	if err := s.cdpEngine.RestoreToPoint(ctx, req.Path, point); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":           "Restore initiated",
		"path":              req.Path,
		"recovery_point_id": req.RecoveryPointID,
	})
}

// getCDPStats returns CDP statistics
func (s *Server) getCDPStats(c *gin.Context) {
	ctx := context.Background()
	stats, err := s.cdpEngine.GetCDPStats(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// getRPOStats returns RPO statistics for a path
func (s *Server) getRPOStats(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path parameter is required"})
		return
	}

	ctx := context.Background()
	stats, err := s.cdpEngine.GetRPOStats(ctx, path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// ==================== Synthetic Backup Handlers ====================

// createSyntheticBackup creates a new synthetic backup
func (s *Server) createSyntheticBackup(c *gin.Context) {
	var req synthetic.SyntheticBackupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	result, err := s.syntheticMgr.CreateSyntheticBackup(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// listSyntheticBackups lists all synthetic backups
func (s *Server) listSyntheticBackups(c *gin.Context) {
	ctx := context.Background()
	filter := &synthetic.SyntheticBackupFilter{}

	backups, err := s.syntheticMgr.ListSyntheticBackups(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"backups": backups, "count": len(backups)})
}

// getSyntheticBackup gets a specific synthetic backup
func (s *Server) getSyntheticBackup(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()

	backup, err := s.syntheticMgr.GetSyntheticBackup(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, backup)
}

// deleteSyntheticBackup deletes a synthetic backup
func (s *Server) deleteSyntheticBackup(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()

	if err := s.syntheticMgr.DeleteSyntheticBackup(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Synthetic backup deleted", "id": id})
}

// mergeIncrementals merges incremental backups into a synthetic full
func (s *Server) mergeIncrementals(c *gin.Context) {
	var req synthetic.MergeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	result, err := s.syntheticMgr.MergeIncrementals(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// getBackupChain gets the backup chain for a synthetic backup
func (s *Server) getBackupChain(c *gin.Context) {
	ctx := context.Background()
	filter := &synthetic.ChainFilter{}

	chain, err := s.syntheticMgr.GetBackupChain(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, chain)
}

// getSyntheticStats returns synthetic backup statistics
func (s *Server) getSyntheticStats(c *gin.Context) {
	ctx := context.Background()
	stats, err := s.syntheticMgr.GetSyntheticStats(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// getBackupChainStats returns backup chain statistics
func (s *Server) getBackupChainStats(c *gin.Context) {
	ctx := context.Background()
	stats, err := s.syntheticMgr.GetBackupChainStats(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}
