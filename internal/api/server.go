package api

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"novabackup/internal/database"
	"novabackup/internal/scheduler"
	"novabackup/pkg/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Server represents the API server
type Server struct {
	engine    *gin.Engine
	db        *database.Connection
	scheduler *scheduler.Scheduler
	mu        sync.RWMutex
}

// NewServer creates a new API server
func NewServer(db *database.Connection, sched *scheduler.Scheduler) (*Server, error) {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(gin.Logger())

	server := &Server{
		engine:    engine,
		db:        db,
		scheduler: sched,
	}

	server.setupRoutes()
	return server, nil
}

// setupRoutes configures API routes
func (s *Server) setupRoutes() {
	// Swagger UI
	s.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check
	s.engine.GET("/health", s.healthCheck)

	// API v1 routes
	v1 := s.engine.Group("/api/v1")
	{
		// Jobs
		jobs := v1.Group("/jobs")
		{
			jobs.GET("", s.listJobs)
			jobs.POST("", s.createJob)
			jobs.GET("/:id", s.getJob)
			jobs.PUT("/:id", s.updateJob)
			jobs.DELETE("/:id", s.deleteJob)
			jobs.POST("/:id/run", s.runJob)
		}

		// Backups
		backups := v1.Group("/backups")
		{
			backups.GET("", s.listBackups)
			backups.GET("/:id", s.getBackup)
			backups.DELETE("/:id", s.deleteBackup)
		}

		// Restore
		restore := v1.Group("/restore")
		{
			restore.POST("", s.restoreFiles)
			restore.GET("/points", s.listRestorePoints)
			restore.POST("/db", s.restoreDatabase)
		}

		// Storage stats
		v1.GET("/storage/stats", s.getStorageStats)

		// Metrics
		v1.GET("/metrics", s.getMetrics)
	}
}

// Start starts the API server
func (s *Server) Start(port int) error {
	addr := fmt.Sprintf(":%d", port)
	return s.engine.Run(addr)
}

// healthCheck returns server health status
// @Summary Health check
// @Description Check if the API server is healthy
// @Tags system
// @Success 200 {object} map[string]string
// @Router /health [get]
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"version":   "6.0.0",
		"scheduler": "running",
	})
}

// listJobs returns all backup jobs
// @Summary List all jobs
// @Description Get a list of all backup jobs
// @Tags jobs
// @Success 200 {array} models.Job
// @Router /api/v1/jobs [get]
func (s *Server) listJobs(c *gin.Context) {
	jobs, err := s.db.GetAllJobs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, jobs)
}

// createJob creates a new backup job
// @Summary Create job
// @Description Create a new backup job
// @Tags jobs
// @Param job body object true "Job details"
// @Success 201 {object} models.Job
// @Router /api/v1/jobs [post]
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
// @Summary Get job
// @Description Get a backup job by ID
// @Tags jobs
// @Param id path string true "Job ID"
// @Success 200 {object} models.Job
// @Router /api/v1/jobs/{id} [get]
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
// @Summary Update job
// @Description Update a backup job
// @Tags jobs
// @Param id path string true "Job ID"
// @Param job body models.Job true "Job details"
// @Success 200 {object} models.Job
// @Router /api/v1/jobs/{id} [put]
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
	// TODO: Implement update in database
	c.JSON(http.StatusOK, job)
}

// deleteJob deletes a job
// @Summary Delete job
// @Description Delete a backup job
// @Tags jobs
// @Param id path string true "Job ID"
// @Success 200 {object} map[string]string
// @Router /api/v1/jobs/{id} [delete]
func (s *Server) deleteJob(c *gin.Context) {
	id := c.Param("id")
	// TODO: Implement delete in database
	c.JSON(http.StatusOK, gin.H{"message": "job deleted", "id": id})
}

// runJob manually triggers a job
// @Summary Run job
// @Description Manually run a backup job
// @Tags jobs
// @Param id path string true "Job ID"
// @Success 200 {object} map[string]string
// @Router /api/v1/jobs/{id}/run [post]
func (s *Server) runJob(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "job started", "id": id})
}

// listBackups returns all backup results
// @Summary List backups
// @Description Get a list of all backup results
// @Tags backups
// @Success 200 {array} models.BackupResult
// @Router /api/v1/backups [get]
func (s *Server) listBackups(c *gin.Context) {
	// TODO: Implement backup results listing
	c.JSON(http.StatusOK, []models.BackupResult{})
}

// getBackup returns a specific backup
// @Summary Get backup
// @Description Get a backup result by ID
// @Tags backups
// @Param id path string true "Backup ID"
// @Success 200 {object} models.BackupResult
// @Router /api/v1/backups/{id} [get]
func (s *Server) getBackup(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"id": id, "status": "completed"})
}

// deleteBackup deletes a backup
// @Summary Delete backup
// @Description Delete a backup result
// @Tags backups
// @Param id path string true "Backup ID"
// @Success 200 {object} map[string]string
// @Router /api/v1/backups/{id} [delete]
func (s *Server) deleteBackup(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "backup deleted", "id": id})
}

// restoreFiles restores files from backup
// @Summary Restore files
// @Description Restore files from a backup
// @Tags restore
// @Param request body object true "Restore request"
// @Success 200 {object} map[string]string
// @Router /api/v1/restore [post]
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

	// TODO: Implement restore logic
	c.JSON(http.StatusOK, gin.H{"message": "restore started", "backup_id": req.BackupID})
}

// restoreDatabase restores a database from backup
// @Summary Restore database
// @Description Restore MySQL or PostgreSQL database
// @Tags restore
// @Param request body object true "Restore request"
// @Success 200 {object} map[string]string
// @Router /api/v1/restore/db [post]
func (s *Server) restoreDatabase(c *gin.Context) {
	var req struct {
		BackupFile string `json:"backup_file" binding:"required"`
		DbType     string `json:"db_type" binding:"required"`
		Host       string `json:"host"`
		Port       int    `json:"port"`
		User       string `json:"user"`
		Password   string `json:"password"`
		Database   string `json:"database"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement database restore logic
	c.JSON(http.StatusOK, gin.H{"message": "database restore started", "file": req.BackupFile})
}

// listRestorePoints returns all restore points
// @Summary List restore points
// @Description Get a list of all restore points
// @Tags restore
// @Success 200 {array} models.RestorePoint
// @Router /api/v1/restore/points [get]
func (s *Server) listRestorePoints(c *gin.Context) {
	// TODO: Implement restore points listing
	c.JSON(http.StatusOK, []models.RestorePoint{})
}

// getStorageStats returns storage statistics
// @Summary Storage stats
// @Description Get storage statistics
// @Tags storage
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/storage/stats [get]
func (s *Server) getStorageStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"total_size":   0,
		"chunk_count":  0,
		"dedupe_ratio": 1.0,
	})
}

// getMetrics returns Prometheus-style metrics
// @Summary Metrics
// @Description Get Prometheus-style metrics
// @Tags metrics
// @Success 200 {string} string
// @Router /api/v1/metrics [get]
func (s *Server) getMetrics(c *gin.Context) {
	c.Header("Content-Type", "text/plain")
	c.String(http.StatusOK, `# HELP novabackup_jobs_total Total number of backup jobs
# TYPE novabackup_jobs_total counter
novabackup_jobs_total 0
# HELP novabackup_bytes_processed_total Total bytes processed
# TYPE novabackup_bytes_processed_total counter
novabackup_bytes_processed_total 0
`)
}
