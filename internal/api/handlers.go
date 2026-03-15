package api

import (
	"time"

	"novabackup/internal/database"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var DB *database.Database

// Health check
func GetHealth(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "healthy",
		"version": "7.0.0",
		"time":    time.Now().Format(time.RFC3339),
	})
}

// Auth (simple for now)
func Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	// Simple auth (TODO: implement proper auth)
	if req.Username == "admin" && req.Password == "admin123" {
		c.JSON(200, gin.H{
			"success": true,
			"token":   uuid.New().String(),
			"user":    "admin",
		})
		return
	}

	c.JSON(401, gin.H{"error": "Invalid credentials"})
}

func Logout(c *gin.Context) {
	c.JSON(200, gin.H{"success": true})
}

// Jobs
func ListJobs(c *gin.Context) {
	jobs, err := DB.ListJobs()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"jobs": jobs})
}

func CreateJob(c *gin.Context) {
	var job database.Job
	if err := c.ShouldBindJSON(&job); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	job.ID = uuid.New().String()
	job.CreatedAt = time.Now()

	if err := DB.CreateJob(&job); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"success": true, "job": job})
}

func UpdateJob(c *gin.Context) {
	id := c.Param("id")

	job, err := DB.GetJob(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Job not found"})
		return
	}

	var update database.Job
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	job.Name = update.Name
	job.Type = update.Type
	job.Sources = update.Sources
	job.Destination = update.Destination
	job.Compression = update.Compression
	job.Encryption = update.Encryption
	job.Schedule = update.Schedule
	job.Enabled = update.Enabled

	if err := DB.UpdateJob(job); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "job": job})
}

func DeleteJob(c *gin.Context) {
	id := c.Param("id")

	if err := DB.DeleteJob(id); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true})
}

func RunJob(c *gin.Context) {
	id := c.Param("id")

	job, err := DB.GetJob(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Job not found"})
		return
	}

	// TODO: Implement actual backup execution
	session := &database.Session{
		ID:             uuid.New().String(),
		JobID:          job.ID,
		JobName:        job.Name,
		StartTime:      time.Now(),
		EndTime:        time.Now(),
		Status:         "success",
		FilesProcessed: 0,
		BytesWritten:   0,
	}

	if err := DB.CreateSession(session); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Update job last run
	nextRun := time.Now().Add(24 * time.Hour)
	DB.UpdateJobLastRun(job.ID, time.Now(), nextRun)

	c.JSON(200, gin.H{
		"success": true,
		"message": "Backup job started",
		"session": session,
	})
}

// Backup
func RunBackup(c *gin.Context) {
	var req struct {
		Name        string   `json:"name"`
		Type        string   `json:"type"`
		Sources     []string `json:"sources"`
		Destination string   `json:"destination"`
		Compression bool     `json:"compression"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Create ad-hoc session
	session := &database.Session{
		ID:             uuid.New().String(),
		JobID:          "adhoc",
		JobName:        req.Name,
		StartTime:      time.Now(),
		EndTime:        time.Now(),
		Status:         "success",
		FilesProcessed: 0,
		BytesWritten:   0,
	}

	if err := DB.CreateSession(session); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"session": session,
		"message": "Backup started",
	})
}

func ListSessions(c *gin.Context) {
	sessions, err := DB.ListSessions()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"sessions": sessions})
}

func GetSession(c *gin.Context) {
	id := c.Param("id")

	sessions, err := DB.ListSessions()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	for _, s := range sessions {
		if s.ID == id {
			c.JSON(200, s)
			return
		}
	}

	c.JSON(404, gin.H{"error": "Session not found"})
}

// Restore
func ListRestorePoints(c *gin.Context) {
	sessions, err := DB.ListSessions()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Filter successful sessions
	var points []database.Session
	for _, s := range sessions {
		if s.Status == "success" {
			points = append(points, s)
		}
	}

	c.JSON(200, gin.H{"restore_points": points})
}

func RestoreFiles(c *gin.Context) {
	var req struct {
		BackupPath  string `json:"backup_path"`
		Destination string `json:"destination"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement actual restore
	c.JSON(200, gin.H{
		"success": true,
		"message": "Restore started",
	})
}

func RestoreDatabase(c *gin.Context) {
	var req struct {
		DBType   string `json:"db_type"`
		DumpFile string `json:"dump_file"`
		ConnStr  string `json:"conn_str"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement actual database restore
	c.JSON(200, gin.H{
		"success": true,
		"message": "Database restore started",
	})
}

// Storage
func ListRepos(c *gin.Context) {
	c.JSON(200, gin.H{"repos": []interface{}{}})
}

func CreateRepo(c *gin.Context) {
	c.JSON(201, gin.H{"success": true})
}

func DeleteRepo(c *gin.Context) {
	c.JSON(200, gin.H{"success": true})
}

// Settings
func GetSettings(c *gin.Context) {
	c.JSON(200, gin.H{
		"port":       8050,
		"data_dir":   "/data",
		"backup_dir": "/data/backups",
		"log_level":  "info",
		"notifications": gin.H{
			"email":    false,
			"telegram": false,
		},
	})
}

func UpdateSettings(c *gin.Context) {
	var settings map[string]interface{}
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// TODO: Save settings to database
	c.JSON(200, gin.H{"success": true})
}
