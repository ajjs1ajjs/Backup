package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
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
	AuditEngine        *rbac.AuditEngine
	ConfigPath         string
)

type backupSessionResponse struct {
	backup.BackupSession
	TotalSize   int64  `json:"total_size"`
	ArchivePath string `json:"archive_path"`
}

func toBackupSessionResponse(session backup.BackupSession) backupSessionResponse {
	totalSize := session.BytesTotal
	if totalSize == 0 {
		totalSize = session.BytesWritten
	}
	return backupSessionResponse{
		BackupSession: session,
		TotalSize:     totalSize,
		ArchivePath:   session.BackupPath,
	}
}

func loadBackupSessionByID(id string) (*backup.BackupSession, error) {
	if BackupEngine == nil {
		return nil, fmt.Errorf("backup engine not initialized")
	}
	if id == "" {
		return nil, fmt.Errorf("session id is required")
	}
	sessionFile := filepath.Join(BackupEngine.DataDir, "sessions", fmt.Sprintf("%s.json", id))
	data, err := os.ReadFile(sessionFile)
	if err != nil {
		return nil, err
	}
	var session backup.BackupSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

func loadBackupSessionsFromDisk() ([]backup.BackupSession, error) {
	if BackupEngine == nil {
		return nil, fmt.Errorf("backup engine not initialized")
	}
	sessionsDir := filepath.Join(BackupEngine.DataDir, "sessions")
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		return nil, err
	}

	sessions := make([]backup.BackupSession, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(sessionsDir, entry.Name()))
		if err != nil {
			continue
		}
		var session backup.BackupSession
		if err := json.Unmarshal(data, &session); err != nil {
			continue
		}
		sessions = append(sessions, session)
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].StartTime.After(sessions[j].StartTime)
	})

	if len(sessions) > 100 {
		sessions = sessions[:100]
	}

	return sessions, nil
}

func persistBackupSession(session *backup.BackupSession) {
	if DB == nil || session == nil {
		return
	}

	dbSession := &database.Session{
		ID:             session.ID,
		JobID:          session.JobID,
		JobName:        session.JobName,
		StartTime:      session.StartTime,
		EndTime:        session.EndTime,
		Status:         session.Status,
		FilesProcessed: session.FilesProcessed,
		BytesWritten:   session.BytesWritten,
		Error:          session.Error,
	}

	if err := DB.CreateSession(dbSession); err != nil {
		log.Printf("Warning: failed to persist session %s: %v", session.ID, err)
	}
}

func toBackupJob(job database.Job) *backup.BackupJob {
	return &backup.BackupJob{
		ID:               job.ID,
		Name:             job.Name,
		Type:             job.Type,
		Sources:          job.Sources,
		Destination:      job.Destination,
		Compression:      job.Compression,
		CompressionLevel: job.CompressionLevel,
		Encryption:       job.Encryption,
		EncryptionKey:    job.EncryptionKey,
		Incremental:      job.Incremental,
		FullBackupEvery:  job.FullBackupEvery,
		Schedule:         job.Schedule,
		ScheduleTime:     job.ScheduleTime,
		ScheduleDays:     job.ScheduleDays,
		CronExpression:   job.CronExpression,
		DatabaseType:     job.DatabaseType,
		Server:           job.Server,
		Port:             job.Port,
		AuthType:         job.AuthType,
		Login:            job.Login,
		Password:         job.Password,
		Service:          job.Service,
		VMNames:          job.VMNames,
		HyperVHost:       job.HyperVHost,
		RetentionDays:    job.RetentionDays,
		RetentionCopies:  job.RetentionCopies,
		ExcludePatterns:  job.ExcludePatterns,
		IncludePatterns:  job.IncludePatterns,
		PreBackupScript:  job.PreBackupScript,
		PostBackupScript: job.PostBackupScript,
		MaxThreads:       job.MaxThreads,
		BlockSize:        job.BlockSize,
	}
}

func sanitizeJobForResponse(job database.Job) database.Job {
	job.Password = ""
	job.EncryptionKey = ""
	return job
}

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
		c.JSON(400, gin.H{"error": "Невірний формат запиту"})
		return
	}

	user, err := RBACEngine.Authenticate(req.Username, req.Password)
	if err != nil {
		log.Printf("Failed login attempt for user '%s': %v", req.Username, err)
		c.JSON(401, gin.H{"error": err.Error()})
		return
	}

	session, err := RBACEngine.CreateSession(user.ID, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		log.Printf("Failed to create session for user '%s': %v", req.Username, err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	log.Printf("User '%s' logged in successfully from %s", req.Username, c.ClientIP())
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
	userValue, _ := c.Get("user")
	if user, ok := userValue.(*rbac.User); ok {
		log.Printf("User '%s' logged out", user.Username)
	}
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
		c.JSON(400, gin.H{"error": "Невірний формат запиту"})
		return
	}

	// Get user from token (set by AuthMiddleware)
	userValue, exists := c.Get("user")
	if !exists {
		c.JSON(401, gin.H{"error": "Користувач не автентифікований"})
		return
	}

	user, ok := userValue.(*rbac.User)
	if !ok {
		c.JSON(401, gin.H{"error": "Недійсні дані користувача"})
		return
	}

	// Verify current password
	if !rbac.CheckPassword(req.CurrentPassword, user.PasswordHash) {
		log.Printf("Failed password change attempt for user '%s': wrong current password", user.Username)
		c.JSON(401, gin.H{"error": "Невірний поточний пароль"})
		return
	}

	// Validate new password with policy
	if err := rbac.PasswordPolicy(req.NewPassword); err != nil {
		log.Printf("Failed password change for user '%s': %v", user.Username, err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Update password
	if err := RBACEngine.ChangePassword(user.ID, req.CurrentPassword, req.NewPassword); err != nil {
		log.Printf("Error changing password for user '%s': %v", user.Username, err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Password changed successfully for user '%s'", user.Username)
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

	// Update each section
	for section, data := range settings {
		if err := saveConfigField(section, data); err != nil {
			log.Printf("Error saving %s settings: %v", section, err)
			c.JSON(500, gin.H{"error": fmt.Sprintf("Не вдалося зберегти налаштування: %v", err)})
			return
		}

		if section == "security" {
			applySecuritySettings(data)
		}
	}

	c.JSON(200, gin.H{"success": true, "message": "Налаштування збережено"})
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
	if err := saveConfigField("server", config); err != nil {
		log.Printf("Error saving server settings: %v", err)
		c.JSON(500, gin.H{"error": "Не вдалося зберегти налаштування сервера"})
		return
	}

	c.JSON(200, gin.H{"success": true})
}

func UpdateNotificationSettings(c *gin.Context) {
	var settings NotificationSettings
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Save notification settings
	if err := saveConfigField("notifications", settings); err != nil {
		log.Printf("Error saving notification settings: %v", err)
		c.JSON(500, gin.H{"error": "Не вдалося зберегти налаштування сповіщень"})
		return
	}

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

	if err := saveConfigField("retention", retention); err != nil {
		log.Printf("Error saving retention settings: %v", err)
		c.JSON(500, gin.H{"error": "Не вдалося зберегти налаштування зберігання"})
		return
	}

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

	if err := saveConfigField("directories", dirs); err != nil {
		log.Printf("Error saving directory settings: %v", err)
		c.JSON(500, gin.H{"error": "Не вдалося зберегти налаштування директорій"})
		return
	}

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

	safeJobs := make([]database.Job, 0, len(jobs))
	for _, job := range jobs {
		safeJobs = append(safeJobs, sanitizeJobForResponse(job))
	}

	c.JSON(200, gin.H{"jobs": safeJobs})
}

func CreateJob(c *gin.Context) {
	var job database.Job
	if err := c.ShouldBindJSON(&job); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Normalize paths (remove Unicode characters and convert slashes)
	for i, src := range job.Sources {
		job.Sources[i] = normalizePath(src)
	}
	job.Destination = normalizePath(job.Destination)

	if job.Encryption && job.EncryptionKey == "" {
		c.JSON(400, gin.H{"error": "Потрібен ключ шифрування"})
		return
	}
	if !job.Encryption {
		job.EncryptionKey = ""
	}

	job.ID = uuid.New().String()
	job.CreatedAt = time.Now()

	if err := DB.CreateJob(&job); err != nil {
		log.Printf("Failed to create job: %v", err)
		if errors.Is(err, database.ErrMasterKeyMissing) {
			c.JSON(400, gin.H{"error": "Потрібен NOVABACKUP_MASTER_KEY для збереження ключа шифрування"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if Scheduler != nil {
		_ = Scheduler.AddJob(toBackupJob(job), job.Schedule, job.ScheduleTime, job.ScheduleDays, job.CronExpression)
	}

	log.Printf("Job '%s' (%s) created successfully", job.Name, job.ID)
	c.JSON(201, gin.H{"success": true, "job": sanitizeJobForResponse(job)})
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

	// Normalize paths (remove Unicode characters and convert slashes)
	for i, src := range update.Sources {
		update.Sources[i] = normalizePath(src)
	}
	update.Destination = normalizePath(update.Destination)

	job.Name = update.Name
	job.Type = update.Type
	job.Sources = update.Sources
	job.Destination = update.Destination
	job.Compression = update.Compression
	job.CompressionLevel = update.CompressionLevel
	job.Encryption = update.Encryption
	if !update.Encryption {
		job.EncryptionKey = ""
	} else if update.EncryptionKey != "" {
		job.EncryptionKey = update.EncryptionKey
	}
	if job.Encryption && job.EncryptionKey == "" {
		c.JSON(400, gin.H{"error": "Потрібен ключ шифрування"})
		return
	}
	job.Deduplication = update.Deduplication
	job.BlockSize = update.BlockSize
	job.MaxThreads = update.MaxThreads
	job.Incremental = update.Incremental
	job.FullBackupEvery = update.FullBackupEvery
	job.ExcludePatterns = update.ExcludePatterns
	job.IncludePatterns = update.IncludePatterns
	job.PreBackupScript = update.PreBackupScript
	job.PostBackupScript = update.PostBackupScript
	job.Schedule = update.Schedule
	job.ScheduleTime = update.ScheduleTime
	job.ScheduleDays = update.ScheduleDays
	job.CronExpression = update.CronExpression
	job.Enabled = update.Enabled
	job.RetentionDays = update.RetentionDays
	job.RetentionCopies = update.RetentionCopies
	job.GFSDaily = update.GFSDaily
	job.GFSWeekly = update.GFSWeekly
	job.GFSMonthly = update.GFSMonthly
	job.GFSQuarterly = update.GFSQuarterly
	job.GFSYearly = update.GFSYearly
	job.BackupCopyEnabled = update.BackupCopyEnabled
	job.BackupCopyDestID = update.BackupCopyDestID
	job.BackupCopyDelay = update.BackupCopyDelay
	job.BackupCopyEncrypt = update.BackupCopyEncrypt
	job.DatabaseType = update.DatabaseType
	job.Server = update.Server
	job.Port = update.Port
	job.AuthType = update.AuthType
	if update.AuthType != "" && update.AuthType != "sql" {
		job.Login = ""
		job.Password = ""
	} else {
		if update.Login != "" {
			job.Login = update.Login
		}
		if update.Password != "" {
			job.Password = update.Password
		}
	}
	job.Service = update.Service
	job.VMNames = update.VMNames
	job.HyperVHost = update.HyperVHost

	if err := DB.UpdateJob(job); err != nil {
		if errors.Is(err, database.ErrMasterKeyMissing) {
			c.JSON(400, gin.H{"error": "Потрібен NOVABACKUP_MASTER_KEY для збереження ключа шифрування"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if Scheduler != nil {
		_ = Scheduler.RemoveJob(job.ID)
		if job.Enabled {
			_ = Scheduler.AddJob(toBackupJob(*job), job.Schedule, job.ScheduleTime, job.ScheduleDays, job.CronExpression)
		}
	}

	c.JSON(200, gin.H{"success": true, "job": sanitizeJobForResponse(*job)})
}

func DeleteJob(c *gin.Context) {
	id := c.Param("id")

	// Check if job exists first
	_, err := DB.GetJob(id)
	if err != nil {
		log.Printf("Job '%s' not found: %v", id, err)
		c.JSON(404, gin.H{"error": "Завдання не знайдено"})
		return
	}

	// Remove from scheduler first
	Scheduler.RemoveJob(id)

	// Then delete from database
	if err := DB.DeleteJob(id); err != nil {
		log.Printf("Failed to delete job '%s': %v", id, err)
		c.JSON(500, gin.H{"error": "Помилка видалення: " + err.Error()})
		return
	}

	log.Printf("Job '%s' deleted successfully", id)
	c.JSON(200, gin.H{"success": true})
}

func RunJob(c *gin.Context) {
	id := c.Param("id")

	job, err := DB.GetJob(id)
	if err != nil {
		log.Printf("Job '%s' not found: %v", id, err)
		c.JSON(404, gin.H{"error": "Завдання не знайдено"})
		return
	}

	// Send notification
	NotificationEngine.SendBackupStarted(job.Name, job.Type)

	backupJob := &backup.BackupJob{
		ID:            job.ID,
		Name:          job.Name,
		Type:          job.Type,
		Sources:       job.Sources,
		Destination:   job.Destination,
		Compression:   job.Compression,
		Encryption:    job.Encryption,
		EncryptionKey: job.EncryptionKey,
		// Database specific fields
		DatabaseType: job.DatabaseType,
		Server:       job.Server,
		Port:         job.Port,
		AuthType:     job.AuthType,
		Login:        job.Login,
		Password:     job.Password,
		Service:      job.Service,
		// VM specific fields
		VMNames:    job.VMNames,
		HyperVHost: job.HyperVHost,
	}

	log.Printf("Starting backup job '%s' (%s): type=%s, sources=%v, dest=%s", job.Name, job.ID, job.Type, job.Sources, job.Destination)
	session, err := BackupEngine.ExecuteBackup(backupJob)
	persistBackupSession(session)
	if err != nil {
		log.Printf("Backup job '%s' failed: %v", job.Name, err)
		NotificationEngine.SendBackupFailed(job.Name, err)
		c.JSON(500, gin.H{"error": "Помилка виконання резервного копіювання: " + err.Error()})
		return
	}

	NotificationEngine.SendBackupSuccess(
		job.Name,
		session.EndTime.Sub(session.StartTime),
		session.FilesProcessed,
		session.BytesWritten,
	)

	nextRun := time.Now().Add(24 * time.Hour)
	if err := DB.UpdateJobLastRun(job.ID, time.Now(), nextRun); err != nil {
		log.Printf("Warning: Failed to update job last_run: %v", err)
	}

	log.Printf("Backup job '%s' completed successfully", job.Name)
	c.JSON(200, gin.H{
		"success": true,
		"message": "Резервне копіювання запущено",
		"session": session,
	})
}

// Stop Job - NEW
func StopJob(c *gin.Context) {
	id := c.Param("id")

	job, err := DB.GetJob(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Завдання не знайдено"})
		return
	}

	// Try to stop via scheduler if available
	if Scheduler != nil {
		Scheduler.RemoveJob(id)
		log.Printf("Job '%s' stopped by user", job.Name)
	}

	canceled := false
	if BackupEngine != nil {
		canceled = BackupEngine.CancelJob(id)
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": func() string {
			if canceled {
				return "Завдання скасовано"
			}
			return "Завдання зупинено"
		}(),
	})
}

// Backup
func RunBackup(c *gin.Context) {
	var req struct {
		Name          string   `json:"name"`
		Type          string   `json:"type"`
		Sources       []string `json:"sources"`
		Destination   string   `json:"destination"`
		Compression   bool     `json:"compression"`
		EncryptionKey string   `json:"encryption_key"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	backupJob := &backup.BackupJob{
		ID:            uuid.New().String(),
		Name:          req.Name,
		Type:          req.Type,
		Sources:       req.Sources,
		Destination:   req.Destination,
		Compression:   req.Compression,
		Encryption:    req.EncryptionKey != "",
		EncryptionKey: req.EncryptionKey,
	}

	log.Printf("Starting manual backup '%s': sources=%v, dest=%s", req.Name, req.Sources, req.Destination)
	session, err := BackupEngine.ExecuteBackup(backupJob)
	persistBackupSession(session)
	if err != nil {
		log.Printf("Manual backup '%s' failed: %v", req.Name, err)
		c.JSON(500, gin.H{"error": "Помилка резервного копіювання: " + err.Error()})
		return
	}

	log.Printf("Manual backup '%s' completed successfully", req.Name)
	c.JSON(200, gin.H{
		"success": true,
		"session": session,
		"message": "Резервне копіювання запущено",
	})
}

func ListSessions(c *gin.Context) {
	if diskSessions, err := loadBackupSessionsFromDisk(); err == nil && len(diskSessions) > 0 {
		response := make([]backupSessionResponse, 0, len(diskSessions))
		for _, session := range diskSessions {
			response = append(response, toBackupSessionResponse(session))
		}
		c.JSON(200, gin.H{"sessions": response})
		return
	}

	sessions, err := DB.ListSessions()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"sessions": sessions})
}

// ListSessionsPublic - public version without auth requirement
func ListSessionsPublic(c *gin.Context) {
	ListSessions(c)
}

func GetSession(c *gin.Context) {
	id := c.Param("id")

	if session, err := loadBackupSessionByID(id); err == nil {
		c.JSON(200, toBackupSessionResponse(*session))
		return
	}

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
		sessionID := c.Param("id")
		if sessionID != "" {
			if session, err := loadBackupSessionByID(sessionID); err == nil {
				backupPath = session.BackupPath
			}
		}
	}

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
		EncryptionKey   string   `json:"encryption_key"`
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
		EncryptionKey:   req.EncryptionKey,
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

// InstantRestore performs instant VM/file recovery
func InstantRestore(c *gin.Context) {
	var req struct {
		BackupPath  string   `json:"backup_path"`
		Type        string   `json:"type"` // vm, files
		VMName      string   `json:"vm_name"`
		Destination string   `json:"destination"`
		Files       []string `json:"files"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Невірний формат запиту"})
		return
	}

	restoreType := restore.RestoreInstant
	if req.Type == "files" {
		restoreType = restore.RestoreGranular
	}

	restoreReq := &restore.RestoreRequest{
		ID:          uuid.New().String(),
		Type:        restoreType,
		BackupPath:  req.BackupPath,
		VMName:      req.VMName,
		Destination: req.Destination,
		Files:       req.Files,
	}

	session, err := RestoreEngine.ExecuteRestore(restoreReq)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"session": session,
		"message": "Миттєве відновлення запущено",
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

func UpdateRepo(c *gin.Context) {
	id := c.Param("id")

	var repo storage.StorageRepo
	if err := c.ShouldBindJSON(&repo); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Get existing repo
	existingRepo, err := StorageEngine.GetRepo(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Сховище не знайдено"})
		return
	}

	// Update fields
	existingRepo.Name = repo.Name
	existingRepo.Description = repo.Description
	existingRepo.Enabled = repo.Enabled
	existingRepo.MaxThreads = repo.MaxThreads

	// Update credentials (only if provided)
	if repo.AccessKey != "" {
		existingRepo.AccessKey = repo.AccessKey
	}
	if repo.SecretKey != "" {
		existingRepo.SecretKey = repo.SecretKey
	}
	if repo.Username != "" {
		existingRepo.Username = repo.Username
	}
	if repo.Password != "" {
		existingRepo.Password = repo.Password
	}

	// Remove and re-add to update provider
	StorageEngine.RemoveRepo(id)
	if err := StorageEngine.AddRepo(existingRepo); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "repo": existingRepo})
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

func GetUser(c *gin.Context) {
	id := c.Param("id")

	user, err := RBACEngine.GetUser(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Користувача не знайдено"})
		return
	}

	c.JSON(200, gin.H{"user": user})
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

func UpdateUser(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Email    string `json:"email"`
		FullName string `json:"full_name"`
		Role     string `json:"role"`
		Enabled  *bool  `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Невірний формат запиту"})
		return
	}

	// Get current user
	user, err := RBACEngine.GetUser(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Користувача не знайдено"})
		return
	}

	// Update fields
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.Role != "" {
		// Validate role
		if _, exists := rbac.RolePermissions[req.Role]; exists {
			user.Role = req.Role
		} else {
			c.JSON(400, gin.H{"error": "Невідома роль"})
			return
		}
	}

	// Enable/disable user
	if req.Enabled != nil {
		if *req.Enabled {
			err = RBACEngine.EnableUser(id)
		} else {
			err = RBACEngine.DisableUser(id)
		}
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(200, gin.H{"success": true, "user": user})
}

func DeleteUser(c *gin.Context) {
	id := c.Param("id")

	if err := RBACEngine.DeleteUser(id); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true})
}

func EnableUser(c *gin.Context) {
	id := c.Param("id")

	if err := RBACEngine.EnableUser(id); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "Користувача увімкнено"})
}

func DisableUser(c *gin.Context) {
	id := c.Param("id")

	if err := RBACEngine.DisableUser(id); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "Користувача вимкнено"})
}

// Backup Verification (SureBackup-style)
func VerifyBackup(c *gin.Context) {
	var req struct {
		BackupPath string `json:"backup_path"`
		Type       string `json:"type"` // integrity, mountable, bootable, data
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Невірний формат запиту"})
		return
	}

	verificationType := backup.VerificationType(req.Type)
	if verificationType == "" {
		verificationType = backup.VerificationIntegrity
	}

	result, err := BackupEngine.VerifyBackup(req.BackupPath, verificationType)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"result":  result,
	})
}

func GetVerificationHistory(c *gin.Context) {
	backupPath := c.Query("backup_path")
	limit := 10

	results, err := BackupEngine.GetVerificationHistory(backupPath, limit)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"verifications": results})
}

// CBT Statistics
func GetCBTStatistics(c *gin.Context) {
	stats := BackupEngine.GetCBTStatistics()
	c.JSON(200, gin.H{"statistics": stats})
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
		var err error
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			c.JSON(400, gin.H{"error": "Невірний формат дати. Використовуйте YYYY-MM-DD"})
			return
		}
	}

	report, err := ReportEngine.GenerateDailyReport(date)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"report": report})
}

// Audit Logs
func GetAuditLogs(c *gin.Context) {
	limitStr := c.Query("limit")
	limit := 100 // Default limit

	if limitStr != "" {
		fmt.Sscanf(limitStr, "%d", &limit)
		if limit <= 0 || limit > 1000 {
			limit = 100
		}
	}

	logs := AuditEngine.GetLogs(limit)
	c.JSON(200, gin.H{"logs": logs})
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
			"security": map[string]interface{}{
				"allow_scripts": false,
			},
		}
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		log.Printf("Warning: Failed to parse config file: %v", err)
		return map[string]interface{}{}
	}

	return config
}

func saveConfigField(section string, data interface{}) error {
	config := loadConfig()
	config[section] = data

	configFile := filepath.Join(ConfigPath, "config.json")
	if err := os.MkdirAll(ConfigPath, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configFile, configData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func updateNotificationEngine(settings NotificationSettings) {
	// Update notification engine with new settings
	// This is a simplified version
}

type SecuritySettings struct {
	AllowScripts bool `json:"allow_scripts"`
}

func applySecuritySettings(data interface{}) {
	payload, err := json.Marshal(data)
	if err != nil {
		return
	}

	var settings SecuritySettings
	if err := json.Unmarshal(payload, &settings); err != nil {
		return
	}

	if BackupEngine != nil {
		BackupEngine.AllowScripts = settings.AllowScripts
	}
	if RestoreEngine != nil {
		RestoreEngine.AllowScripts = settings.AllowScripts
	}
}

// normalizePath removes Unicode characters and converts slashes for Windows paths
func normalizePath(path string) string {
	if path == "" {
		return path
	}
	// Remove BOM and other invisible Unicode characters
	path = strings.TrimSpace(path)
	path = strings.ReplaceAll(path, "\uFEFF", "") // BOM
	path = strings.ReplaceAll(path, "\u200B", "") // Zero-width space
	path = strings.ReplaceAll(path, "\u200C", "") // Zero-width non-joiner
	path = strings.ReplaceAll(path, "\u200D", "") // Zero-width joiner
	path = strings.ReplaceAll(path, "\u2060", "") // Word joiner
	// Replace forward slashes with backslashes for Windows
	path = filepath.FromSlash(path)
	return path
}

// Database Handlers

// ListDatabases lists all databases on a database server instance
func ListDatabases(c *gin.Context) {
	var req struct {
		Type             string `json:"type"`
		Server           string `json:"server"`
		Port             int    `json:"port"`
		Database         string `json:"database"`
		Login            string `json:"login"`
		Password         string `json:"password"`
		AuthType         string `json:"auth_type"`
		ConnectionString string `json:"connection_string"`
		SSL              string `json:"ssl"`
		Service          string `json:"service"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Невірний формат запиту"})
		return
	}

	log.Printf("Database list request: type=%s, server=%s, port=%d, auth=%s", req.Type, req.Server, req.Port, req.AuthType)

	var databases []gin.H

	if req.Type == "mssql" {
		// Use PowerShell to list SQL Server databases
		server := req.Server
		// Unescape JSON string - converts \\ to \
		if unquoted, err := strconv.Unquote(`"` + server + `"`); err == nil {
			server = unquoted
		}
		log.Printf("Database list request: type=%s, server=%s (normalized: %s), port=%d, auth=%s", req.Type, req.Server, server, req.Port, req.AuthType)
		if server == "" {
			server = "localhost"
		}

		port := req.Port
		if port == 0 {
			port = 1433
		}

		authScript := ""
		if req.AuthType == "sql" {
			// For named instances (SQLEXPRESS), don't specify port - use named pipes
			if strings.Contains(server, "\\") {
				authScript = fmt.Sprintf(`$connectionString = "Server=%s;Database=master;User Id=%s;Password=%s;TrustServerCertificate=true;"`,
					server, req.Login, req.Password)
			} else {
				authScript = fmt.Sprintf(`$connectionString = "Server=%s,%d;Database=master;User Id=%s;Password=%s;TrustServerCertificate=true;"`,
					server, port, req.Login, req.Password)
			}
		} else {
			// Windows Authentication
			if strings.Contains(server, "\\") {
				authScript = fmt.Sprintf(`$connectionString = "Server=%s;Database=master;Integrated Security=true;TrustServerCertificate=true;"`, server)
			} else {
				authScript = fmt.Sprintf(`$connectionString = "Server=%s,%d;Database=master;Integrated Security=true;TrustServerCertificate=true;"`,
					server, port)
			}
		}

		psScript := fmt.Sprintf(`
$ErrorActionPreference = "Stop"
%s

try {
    $connection = New-Object System.Data.SqlClient.SqlConnection
    $connection.ConnectionString = $connectionString
    $connection.Open()

    $query = "SELECT name FROM sys.databases WHERE database_id > 4 ORDER BY name"
    $command = New-Object System.Data.SqlClient.SqlCommand
    $command.CommandText = $query
    $command.Connection = $connection

    $reader = $command.ExecuteReader()
    $databases = @()
    while ($reader.Read()) {
        $databases += @{ name = $reader.GetString(0) }
    }
    $reader.Close()
    $connection.Close()

    $databases | ConvertTo-Json -Depth 1
} catch {
    Write-Error $_.Exception.Message
    exit 1
}
`, authScript)

		cmd := exec.Command("powershell", "-Command", psScript)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()

		log.Printf("SQL Server PowerShell stdout: %s", stdout.String())
		log.Printf("SQL Server PowerShell stderr: %s", stderr.String())

		if err != nil {
			log.Printf("Failed to list SQL Server databases: %v", err)
			c.JSON(500, gin.H{"error": "Не вдалося отримати список БД. Деталі: " + stderr.String()})
			return
		}

		// Parse PowerShell output
		var dbList []map[string]interface{}
		rawJSON := stdout.Bytes()

		// Try to parse as array first
		if err := json.Unmarshal(rawJSON, &dbList); err != nil {
			// If it fails, try parsing as single object and wrap in array
			var singleDB map[string]interface{}
			if err2 := json.Unmarshal(rawJSON, &singleDB); err2 == nil {
				dbList = append(dbList, singleDB)
			} else {
				log.Printf("Failed to parse database list: %v", err)
				c.JSON(500, gin.H{"error": "Помилка обробки даних БД: " + err.Error()})
				return
			}
		}

		for _, db := range dbList {
			databases = append(databases, gin.H{
				"name": db["name"],
				"size": "Unknown",
			})
		}
	} else if req.Type == "postgresql" {
		// Use psql to list databases
		dbName := req.Database
		if dbName == "" {
			dbName = "postgres"
		}
		cmd := exec.Command("psql", "-h", req.Server, "-p", fmt.Sprintf("%d", req.Port), "-U", req.Login, "-d", dbName, "-c", "\\l", "-t")
		if req.Password != "" {
			cmd.Env = append(os.Environ(), "PGPASSWORD="+req.Password)
		}
		output, err := cmd.Output()
		if err != nil {
			log.Printf("Failed to list PostgreSQL databases: %v", err)
			c.JSON(500, gin.H{"error": "Не вдалося отримати список PostgreSQL БД"})
			return
		}

		// Parse psql output
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.Contains(line, "|") {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) > 0 {
				dbName := fields[0]
				if dbName != "Name" && dbName != "List" {
					databases = append(databases, gin.H{
						"name": dbName,
					})
				}
			}
		}
	} else if req.Type == "oracle" {
		// Use sqlplus to list databases
		databases = []gin.H{
			{"name": req.Service},
		}
	} else if req.Type == "mysql" {
		// Use mysql to list databases
		args := []string{"-h", req.Server, "-P", fmt.Sprintf("%d", req.Port), "-u", req.Login, "-e", "SHOW DATABASES"}
		cmd := exec.Command("mysql", args...)
		if req.Password != "" {
			cmd.Env = append(os.Environ(), "MYSQL_PWD="+req.Password)
		}
		output, err := cmd.Output()
		if err != nil {
			log.Printf("Failed to list MySQL databases: %v", err)
			c.JSON(500, gin.H{"error": "Не вдалося отримати список MySQL БД"})
			return
		}

		// Parse mysql output
		lines := strings.Split(string(output), "\n")
		for _, line := range lines[1:] { // Skip header
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			databases = append(databases, gin.H{
				"name": line,
			})
		}
	} else {
		c.JSON(400, gin.H{"error": "Невідомий тип бази даних: " + req.Type})
		return
	}

	c.JSON(200, gin.H{"databases": databases})
}

// BackupDatabase creates an immediate backup of selected databases
func BackupDatabase(c *gin.Context) {
	var req struct {
		Databases        []string `json:"databases"`
		DatabaseType     string   `json:"database_type"`
		Server           string   `json:"server"`
		Port             int      `json:"port"`
		Destination      string   `json:"destination"`
		Login            string   `json:"login"`
		Password         string   `json:"password"`
		AuthType         string   `json:"auth_type"`
		Database         string   `json:"database"`
		SSL              string   `json:"ssl"`
		Service          string   `json:"service"`
		ConnectionString string   `json:"connection_string"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Невірний формат запиту"})
		return
	}

	if len(req.Databases) == 0 {
		c.JSON(400, gin.H{"error": "Виберіть хоча б одну базу даних"})
		return
	}

	sessionID := uuid.New().String()

	log.Printf("Starting %s database backup for %v on server %s", req.DatabaseType, req.Databases, req.Server)

	if BackupEngine == nil {
		c.JSON(500, gin.H{"error": "Backup engine not initialized"})
		return
	}

	job := &backup.BackupJob{
		ID:           sessionID,
		Name:         fmt.Sprintf("DB-%s-%s", strings.ToUpper(req.DatabaseType), time.Now().Format("20060102-150405")),
		Type:         backup.TypeDatabase,
		Sources:      req.Databases,
		Destination:  normalizePath(req.Destination),
		DatabaseType: req.DatabaseType,
		Server:       req.Server,
		Port:         req.Port,
		AuthType:     req.AuthType,
		Login:        req.Login,
		Password:     req.Password,
		Service:      req.Service,
		DatabaseConn: req.ConnectionString,
	}

	session, err := BackupEngine.ExecuteBackup(job)
	persistBackupSession(session)
	if err != nil {
		c.JSON(500, gin.H{"error": "Помилка виконання резервного копіювання: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success":    true,
		"session_id": session.ID,
		"message":    "Бекап завершено",
		"db_type":    req.DatabaseType,
	})
}

// BackupVM creates an immediate backup of selected virtual machines
func BackupVM(c *gin.Context) {
	var req struct {
		VMs         []string `json:"vms"`
		VMType      string   `json:"vm_type"`
		Host        string   `json:"host"`
		Destination string   `json:"destination"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Невірний формат запиту"})
		return
	}

	if len(req.VMs) == 0 {
		c.JSON(400, gin.H{"error": "Виберіть хоча б одну ВМ"})
		return
	}

	sessionID := uuid.New().String()

	log.Printf("Starting %s VM backup for %v on host %s", req.VMType, req.VMs, req.Host)

	// TODO: Implement actual VM backup logic
	// - Hyper-V: Export-VM, checkpoint-based backup
	// - KVM: virsh snapshot-create-as, qemu-img backup

	c.JSON(200, gin.H{
		"success":    true,
		"session_id": sessionID,
		"message":    "Бекап ВМ розпочато",
		"vm_type":    req.VMType,
	})
}

// ListVMs lists virtual machines on a hypervisor
func ListVMs(c *gin.Context) {
	var req struct {
		VMType   string `json:"type"`
		Host     string `json:"host"`
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Невірний формат запиту"})
		return
	}

	var vms []gin.H

	if req.VMType == "hyper-v" {
		// Use PowerShell to list Hyper-V VMs
		psScript := `
$ErrorActionPreference = "Continue"
try {
    # Check if Hyper-V module is available
    $module = Get-Module -ListAvailable -Name Hyper-V
    if ($module -eq $null) {
        Write-Host "Hyper-V module not found"
        echo "[]"
        exit 0
    }

    # Try to get VMs
    $vms = Get-VM -ErrorAction SilentlyContinue
    if ($vms -eq $null) {
        Write-Host "No VMs found or not connected to Hyper-V host"
        echo "[]"
        exit 0
    }

    if ($vms.Count -eq 0) {
        Write-Host "VM list is empty"
        echo "[]"
        exit 0
    }

    # Convert State enum to string
    $stateMap = @{
        0 = "Other"
        1 = "Running"
        2 = "Paused"
        3 = "Off"
        4 = "Saved"
        5 = "Stopping"
        6 = "ShuttingDown"
        7 = "Starting"
        8 = "Reset"
        9 = "Suspending"
        10 = "FastSaved"
        11 = "FastSavedPausing"
        12 = "Pausing"
        13 = "Resuming"
        14 = "Saving"
        15 = "Critical"
    }

    $result = $vms | ForEach-Object {
        $stateText = if ($stateMap.ContainsKey($_.State)) { $stateMap[$_.State] } else { "Unknown" }
        [PSCustomObject]@{
            Name = $_.Name
            State = $stateText.ToLower()
            Memory = [math]::Round($_.MemoryAssigned/1MB, 0)
            Uptime = $_.Uptime.ToString()
            OS = if ($_.GuestOSInDetail) { $_.GuestOSInDetail } else { "Unknown" }
        }
    } | ConvertTo-Json -Depth 3

    if ($result -eq $null -or $result -eq "") {
        Write-Host "ConvertTo-Json returned null"
        echo "[]"
    } else {
        echo $result
    }
} catch {
    Write-Host "Error: $($_.Exception.Message)"
    Write-Host "StackTrace: $($_.ScriptStackTrace)"
    echo "[]"
}
`
		cmd := exec.Command("powershell", "-Command", psScript)
		output, err := cmd.Output()
		if err != nil {
			log.Printf("Failed to list Hyper-V VMs: %v, output: %s", err, string(output))
		} else {
			log.Printf("Hyper-V VMs output: %s", string(output))
		}

		// Parse PowerShell output
		var vmList []map[string]interface{}
		if err := json.Unmarshal(output, &vmList); err != nil {
			log.Printf("Failed to parse VM list: %v, output: %s", err, string(output))
			// Return empty list instead of error
			vms = []gin.H{}
		} else {
			for _, vm := range vmList {
				state := "unknown"
				if vmState, ok := vm["State"].(string); ok {
					state = vmState
				}

				vms = append(vms, gin.H{
					"name":   vm["Name"],
					"state":  strings.ToLower(state),
					"memory": vm["MemoryAssigned"],
					"uptime": vm["Uptime"],
					"os":     vm["OS"],
				})
			}
		}
	} else if req.VMType == "kvm" {
		// Use SSH to list KVM VMs via virsh
		sshCmd := fmt.Sprintf(`sshpass -p '%s' ssh -o StrictHostKeyChecking=no %s@%s "virsh list --all"`,
			req.Password, req.Login, req.Host)

		cmd := exec.Command("bash", "-c", sshCmd)
		output, err := cmd.Output()
		if err != nil {
			log.Printf("Failed to list KVM VMs: %v", err)
			c.JSON(500, gin.H{"error": "Не вдалося отримати список KVM ВМ"})
			return
		}

		// Parse virsh output
		lines := strings.Split(string(output), "\n")
		for _, line := range lines[2:] { // Skip header lines
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				vmName := fields[1]
				vmState := strings.ToLower(fields[2])
				if vmState == "running" {
					vmState = "running"
				} else {
					vmState = "stopped"
				}
				vms = append(vms, gin.H{
					"name":  vmName,
					"state": vmState,
				})
			}
		}
	} else {
		c.JSON(400, gin.H{"error": "Невідомий тип віртуалізації: " + req.VMType})
		return
	}

	c.JSON(200, gin.H{"vms": vms})
}
