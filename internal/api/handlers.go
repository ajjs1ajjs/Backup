package api

import (
	"time"

	"novabackup/internal/backup"
	"novabackup/internal/database"
	"novabackup/internal/restore"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var (
	DB            *database.Database
	BackupEngine  *backup.BackupEngine
	RestoreEngine *restore.RestoreEngine
)

// Health check
func GetHealth(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "healthy",
		"version": "7.0.0",
		"time":    time.Now().Format(time.RFC3339),
	})
}

// Auth
func Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Невірний запит"})
		return
	}

	if req.Username == "admin" && req.Password == "admin123" {
		c.JSON(200, gin.H{
			"success": true,
			"token":   uuid.New().String(),
			"user":    "admin",
		})
		return
	}

	c.JSON(401, gin.H{"error": "Невірні облікові дані"})
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
		c.JSON(404, gin.H{"error": "Завдання не знайдено"})
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
		c.JSON(404, gin.H{"error": "Завдання не знайдено"})
		return
	}

	// Convert database.Job to backup.BackupJob
	backupJob := &backup.BackupJob{
		ID:          job.ID,
		Name:        job.Name,
		Type:        job.Type,
		Sources:     job.Sources,
		Destination: job.Destination,
		Compression: job.Compression,
		Encryption:  job.Encryption,
		Schedule:    job.Schedule,
		Enabled:     job.Enabled,
	}

	// Execute backup
	session, err := BackupEngine.ExecuteBackup(backupJob)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Update job last run
	nextRun := time.Now().Add(24 * time.Hour)
	DB.UpdateJobLastRun(job.ID, time.Now(), nextRun)

	c.JSON(200, gin.H{
		"success": true,
		"message": "Резервне копіювання запущено",
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

	backupJob := &backup.BackupJob{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Type:        req.Type,
		Sources:     req.Sources,
		Destination: req.Destination,
		Compression: req.Compression,
	}

	session, err := BackupEngine.ExecuteBackup(backupJob)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"session": session,
		"message": "Резервне копіювання запущено",
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

	c.JSON(404, gin.H{"error": "Сесію не знайдено"})
}

// Restore
func ListRestorePoints(c *gin.Context) {
	backupPath := c.Query("backup_path")
	if backupPath == "" {
		c.JSON(400, gin.H{"error": "Потрібно вказати backup_path"})
		return
	}

	points, err := RestoreEngine.ListRestorePoints(backupPath)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"restore_points": points})
}

func BrowseBackupFiles(c *gin.Context) {
	backupPath := c.Query("backup_path")
	if backupPath == "" {
		c.JSON(400, gin.H{"error": "Потрібно вказати backup_path"})
		return
	}

	files, err := RestoreEngine.BrowseBackupFiles(backupPath)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"files": files})
}

func RestoreFiles(c *gin.Context) {
	var req struct {
		BackupPath      string   `json:"backup_path"`
		Destination     string   `json:"destination"`
		Files           []string `json:"files"`
		RestoreOriginal bool     `json:"restore_original"`
		Overwrite       bool     `json:"overwrite"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	restoreReq := &restore.RestoreRequest{
		ID:              uuid.New().String(),
		Type:            restore.RestoreFiles,
		BackupPath:      req.BackupPath,
		Destination:     req.Destination,
		Files:           req.Files,
		RestoreOriginal: req.RestoreOriginal,
		Overwrite:       req.Overwrite,
	}

	session, err := RestoreEngine.ExecuteRestore(restoreReq)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"session": session,
		"message": "Відновлення запущено",
	})
}

func RestoreDatabase(c *gin.Context) {
	var req struct {
		BackupPath     string `json:"backup_path"`
		DBType         string `json:"db_type"`
		ConnStr        string `json:"conn_str"`
		TargetDatabase string `json:"target_database"`
		EncryptionKey  string `json:"encryption_key"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	restoreReq := &restore.RestoreRequest{
		ID:             uuid.New().String(),
		Type:           restore.RestoreDatabase,
		BackupPath:     req.BackupPath,
		DBType:         req.DBType,
		ConnStr:        req.ConnStr,
		TargetDatabase: req.TargetDatabase,
		EncryptionKey:  req.EncryptionKey,
	}

	session, err := RestoreEngine.ExecuteRestore(restoreReq)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"session": session,
		"message": "Відновлення бази даних запущено",
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
		"language":   "uk",
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

	c.JSON(200, gin.H{"success": true})
}
