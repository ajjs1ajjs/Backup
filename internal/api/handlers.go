package api

import (
	"time"

	"novabackup/internal/backup"
	"novabackup/internal/database"
	"novabackup/internal/notifications"
	"novabackup/internal/rbac"
	"novabackup/internal/reports"
	"novabackup/internal/restore"
	"novabackup/internal/scheduler"
	"novabackup/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var (
	DB                 *database.Database
	BackupEngine       *backup.BackupEngine
	RestoreEngine      *restore.RestoreEngine
	Scheduler          *scheduler.Scheduler
	StorageEngine      *storage.StorageEngine
	NotificationEngine *notifications.NotificationEngine
	RBACEngine         *rbac.RBACEngine
	ReportEngine       *reports.ReportEngine
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

	user, err := RBACEngine.Authenticate(req.Username, req.Password)
	if err != nil {
		c.JSON(401, gin.H{"error": err.Error()})
		return
	}

	session, err := RBACEngine.CreateSession(user.ID, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"token":   session.Token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role,
		},
	})
}

func Logout(c *gin.Context) {
	token := c.GetHeader("Authorization")
	RBACEngine.Logout(token)
	c.JSON(200, gin.H{"success": true})
}

// Middleware for auth
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(401, gin.H{"error": "Потрібен токен"})
			c.Abort()
			return
		}

		user, err := RBACEngine.ValidateSession(token)
		if err != nil {
			c.JSON(401, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Next()
	}
}

// Permission check middleware
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userValue, _ := c.Get("user")
		user, ok := userValue.(*rbac.User)
		if !ok {
			c.JSON(403, gin.H{"error": "Недостатньо прав"})
			c.Abort()
			return
		}
		if !RBACEngine.CheckPermission(user, permission) {
			c.JSON(403, gin.H{"error": "Недостатньо прав"})
			c.Abort()
			return
		}
		c.Next()
	}
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

	// Add to scheduler
	backupJob := &backup.BackupJob{
		ID:          job.ID,
		Name:        job.Name,
		Type:        job.Type,
		Sources:     job.Sources,
		Destination: job.Destination,
		Compression: job.Compression,
		Encryption:  job.Encryption,
	}

	Scheduler.AddJob(backupJob, job.Schedule, "", nil)

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

	Scheduler.RemoveJob(id)

	c.JSON(200, gin.H{"success": true})
}

func RunJob(c *gin.Context) {
	id := c.Param("id")

	job, err := DB.GetJob(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Завдання не знайдено"})
		return
	}

	// Send notification
	NotificationEngine.SendBackupStarted(job.Name, job.Type)

	// Convert database.Job to backup.BackupJob
	backupJob := &backup.BackupJob{
		ID:          job.ID,
		Name:        job.Name,
		Type:        job.Type,
		Sources:     job.Sources,
		Destination: job.Destination,
		Compression: job.Compression,
		Encryption:  job.Encryption,
	}

	// Execute backup
	session, err := BackupEngine.ExecuteBackup(backupJob)
	if err != nil {
		NotificationEngine.SendBackupFailed(job.Name, err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Send success notification
	NotificationEngine.SendBackupSuccess(
		job.Name,
		session.EndTime.Sub(session.StartTime),
		session.FilesProcessed,
		session.BytesWritten,
	)

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

	NotificationEngine.SendRestoreStarted("Відновлення файлів", "files")

	session, err := RestoreEngine.ExecuteRestore(restoreReq)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	NotificationEngine.SendRestoreSuccess("Відновлення файлів", session.FilesRestored)

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
	repos := StorageEngine.ListRepos()
	c.JSON(200, gin.H{"repos": repos})
}

func CreateRepo(c *gin.Context) {
	var repo storage.StorageRepo
	if err := c.ShouldBindJSON(&repo); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	repo.ID = uuid.New().String()
	repo.CreatedAt = time.Now()

	if err := StorageEngine.AddRepo(&repo); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"success": true, "repo": repo})
}

func DeleteRepo(c *gin.Context) {
	id := c.Param("id")

	if err := StorageEngine.RemoveRepo(id); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

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

// Reports & Statistics
func GetStatistics(c *gin.Context) {
	stats, err := ReportEngine.GetStatistics()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"statistics": stats})
}

func GetDailyReport(c *gin.Context) {
	dateStr := c.Query("date")
	date := time.Now()
	if dateStr != "" {
		date, _ = time.Parse("2006-01-02", dateStr)
	}

	report, err := ReportEngine.GenerateDailyReport(date)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"report": report})
}

func GetWeeklyReport(c *gin.Context) {
	report, err := ReportEngine.GenerateWeeklyReport(time.Now())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"report": report})
}

func GetMonthlyReport(c *gin.Context) {
	report, err := ReportEngine.GenerateMonthlyReport(time.Now())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"report": report})
}

func ExportReport(c *gin.Context) {
	var req struct {
		ReportID string `json:"report_id"`
		Format   string `json:"format"` // json, html, pdf
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Generate report data
	report, err := ReportEngine.GenerateDailyReport(time.Now())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	data, err := ReportEngine.ExportReport(report, req.Format)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.Data(200, "application/"+req.Format, data)
}

// Users (Admin only)
func ListUsers(c *gin.Context) {
	users := RBACEngine.ListUsers()
	c.JSON(200, gin.H{"users": users})
}

func CreateUser(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
		FullName string `json:"full_name"`
		Role     string `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	user, err := RBACEngine.CreateUser(req.Username, req.Password, req.Email, req.FullName, req.Role)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"success": true, "user": user})
}

func DeleteUser(c *gin.Context) {
	id := c.Param("id")

	if err := RBACEngine.DeleteUser(id); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true})
}

// Audit Logs
func GetAuditLogs(c *gin.Context) {
	c.JSON(200, gin.H{"logs": "TODO"})
}

// Scheduler
func GetSchedule(c *gin.Context) {
	nextRuns := Scheduler.GetNextRuns()
	c.JSON(200, gin.H{"schedule": nextRuns})
}

func GetJobStatus(c *gin.Context) {
	id := c.Param("id")

	status, err := Scheduler.GetJobStatus(id)
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": status})
}
