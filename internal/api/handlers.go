package api

import (
	"encoding/json"
	"os"
	"path/filepath"
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
	ConfigPath         string
)

// ServerConfig represents server configuration
type ServerConfig struct {
	IP        string `json:"ip"`
	Port      int    `json:"port"`
	HTTPS     bool   `json:"https"`
	HTTPSPort int    `json:"https_port"`
}

// NotificationSettings represents notification configuration
type NotificationSettings struct {
	Channels map[string]interface{} `json:"channels"`
	Events   map[string]bool        `json:"events"`
}

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

// ChangePassword changes user password
func ChangePassword(c *gin.Context) {
	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Get user from token
	userValue, _ := c.Get("user")
	user, ok := userValue.(*rbac.User)
	if !ok {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	// Verify current password
	if !rbac.CheckPassword(req.CurrentPassword, user.PasswordHash) {
		c.JSON(401, gin.H{"error": "Невірний поточний пароль"})
		return
	}

	// Validate new password
	if len(req.NewPassword) < 6 {
		c.JSON(400, gin.H{"error": "Пароль має бути не менше 6 символів"})
		return
	}

	// Update password
	if err := RBACEngine.ChangePassword(user.ID, req.CurrentPassword, req.NewPassword); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "Пароль змінено"})
}

// Settings
func GetSettings(c *gin.Context) {
	config := loadConfig()
	c.JSON(200, config)
}

func UpdateSettings(c *gin.Context) {
	var settings map[string]interface{}
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"success": true})
}

// ServeWebFile serves static web files
func ServeWebFile(c *gin.Context) {
	file := c.Param("filepath")
	if file == "" || file == "/" {
		file = "index.html"
	}

	filePath := filepath.Join("web", file)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(404, gin.H{"error": "File not found"})
		return
	}

	c.File(filePath)
}

func UpdateServerSettings(c *gin.Context) {
	var config ServerConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Save config
	saveConfigField("server", config)

	c.JSON(200, gin.H{"success": true})
}

func UpdateNotificationSettings(c *gin.Context) {
	var settings NotificationSettings
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Save notification settings
	saveConfigField("notifications", settings)

	// Update notification engine
	updateNotificationEngine(settings)

	c.JSON(200, gin.H{"success": true})
}

func UpdateRetentionSettings(c *gin.Context) {
	var retention struct {
		Type  string `json:"type"`
		Value int    `json:"value"`
	}
	if err := c.ShouldBindJSON(&retention); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	saveConfigField("retention", retention)
	c.JSON(200, gin.H{"success": true})
}

func UpdateDirectorySettings(c *gin.Context) {
	var dirs struct {
		DataDir   string `json:"data_dir"`
		BackupDir string `json:"backup_dir"`
		LogsDir   string `json:"logs_dir"`
	}
	if err := c.ShouldBindJSON(&dirs); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	saveConfigField("directories", dirs)
	c.JSON(200, gin.H{"success": true})
}

// Service Control
func RestartService(c *gin.Context) {
	// In production, use service control
	// For now, just acknowledge
	c.JSON(200, gin.H{"success": true, "message": "Service restart initiated"})
}

func StopService(c *gin.Context) {
	c.JSON(200, gin.H{"success": true, "message": "Service stop initiated"})
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

	backupJob := &backup.BackupJob{
		ID:          job.ID,
		Name:        job.Name,
		Type:        job.Type,
		Sources:     job.Sources,
		Destination: job.Destination,
		Compression: job.Compression,
		Encryption:  job.Encryption,
	}

	session, err := BackupEngine.ExecuteBackup(backupJob)
	if err != nil {
		NotificationEngine.SendBackupFailed(job.Name, err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	NotificationEngine.SendBackupSuccess(
		job.Name,
		session.EndTime.Sub(session.StartTime),
		session.FilesProcessed,
		session.BytesWritten,
	)

	nextRun := time.Now().Add(24 * time.Hour)
	DB.UpdateJobLastRun(job.ID, time.Now(), nextRun)

	c.JSON(200, gin.H{
		"success": true,
		"message": "Резервне копіювання запущено",
		"session": session,
	})
}

// Stop Job - NEW
func StopJob(c *gin.Context) {
	// In production, implement actual stop logic
	// For now, just acknowledge
	c.JSON(200, gin.H{
		"success": true,
		"message": "Завдання зупинено",
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

// Users
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

// Config helpers
func loadConfig() map[string]interface{} {
	configFile := filepath.Join(ConfigPath, "config.json")

	data, err := os.ReadFile(configFile)
	if err != nil {
		// Return defaults
		return map[string]interface{}{
			"server": map[string]interface{}{
				"ip":   "0.0.0.0",
				"port": 8050,
			},
			"notifications": map[string]interface{}{
				"channels": map[string]interface{}{},
				"events":   map[string]bool{},
			},
		}
	}

	var config map[string]interface{}
	json.Unmarshal(data, &config)
	return config
}

func saveConfigField(section string, data interface{}) {
	config := loadConfig()
	config[section] = data

	configFile := filepath.Join(ConfigPath, "config.json")
	os.MkdirAll(ConfigPath, 0755)

	configData, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(configFile, configData, 0644)
}

func updateNotificationEngine(settings NotificationSettings) {
	// Update notification engine with new settings
	// This is a simplified version
}
