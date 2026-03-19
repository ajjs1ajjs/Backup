package api

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
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

	c.JSON(200, gin.H{"jobs": jobs})
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

	job.ID = uuid.New().String()
	job.CreatedAt = time.Now()

	if err := DB.CreateJob(&job); err != nil {
		log.Printf("Failed to create job: %v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Job '%s' (%s) created successfully", job.Name, job.ID)
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
		ID:          job.ID,
		Name:        job.Name,
		Type:        job.Type,
		Sources:     job.Sources,
		Destination: job.Destination,
		Compression: job.Compression,
		Encryption:  job.Encryption,
	}

	log.Printf("Starting backup job '%s' (%s): sources=%v, dest=%s", job.Name, job.ID, job.Sources, job.Destination)
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
