package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
)

// EventLevel represents audit event severity
type EventLevel string

const (
	LevelInfo     EventLevel = "INFO"
	LevelWarning  EventLevel = "WARNING"
	LevelError    EventLevel = "ERROR"
	LevelCritical EventLevel = "CRITICAL"
	LevelAudit    EventLevel = "AUDIT"
)

// EventType represents different types of audit events
type EventType string

const (
	// Authentication events
	EventLogin          EventType = "auth.login"
	EventLogout         EventType = "auth.logout"
	EventLoginFailed    EventType = "auth.login_failed"
	EventPasswordChange EventType = "auth.password_change"
	
	// Authorization events
	EventRoleAssigned   EventType = "auth.role_assigned"
	EventPermissionDeny EventType = "auth.permission_deny"
	
	// Backup job events
	EventJobCreate      EventType = "backup.job_create"
	EventJobDelete      EventType = "backup.job_delete"
	EventJobStart       EventType = "backup.job_start"
	EventJobComplete    EventType = "backup.job_complete"
	EventJobFailed      EventType = "backup.job_failed"
	EventJobModified    EventType = "backup.job_modified"
	
	// Restore events
	EventRestoreStart   EventType = "restore.start"
	EventRestoreComplete EventType = "restore.complete"
	EventRestoreFailed  EventType = "restore.failed"
	
	// Configuration events
	EventConfigChange   EventType = "config.change"
	EventSettingsChange EventType = "settings.change"
	
	// Storage events
	EventStorageAdd     EventType = "storage.add"
	EventStorageRemove  EventType = "storage.remove"
	
	// Security events
	EventSecurityAlert  EventType = "security.alert"
	EventPolicyChange   EventType = "security.policy_change"
)

// Event represents a single audit event
type Event struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	Level       EventLevel             `json:"level"`
	Type        EventType              `json:"type"`
	Category    string                 `json:"category"`
	UserID      string                 `json:"user_id,omitempty"`
	Username    string                 `json:"username,omitempty"`
	SourceIP    string                 `json:"source_ip,omitempty"`
	Resource    string                 `json:"resource,omitempty"`
	ResourceID  string                 `json:"resource_id,omitempty"`
	Action      string                 `json:"action,omitempty"`
	Status      string                 `json:"status,omitempty"`
	Message     string                 `json:"message"`
	Details     map[string]interface{} `json:"details,omitempty"`
	SessionID   string                 `json:"session_id,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	Duration    int64                  `json:"duration_ms,omitempty"`
}

// AuditLogger manages audit logging
type AuditLogger struct {
	mu           sync.RWMutex
	logger       *zap.Logger
	fileLogger   *os.File
	events       []*Event
	maxEvents    int
	enabled      bool
	logFilePath  string
	retentionDays int
}

// AuditConfig holds audit logger configuration
type AuditConfig struct {
	Enabled       bool   `json:"enabled"`
	LogFilePath   string `json:"log_file_path"`
	MaxEvents     int    `json:"max_events"`
	RetentionDays int    `json:"retention_days"`
	ConsoleOutput bool   `json:"console_output"`
}

// DefaultAuditConfig returns default audit configuration
func DefaultAuditConfig() AuditConfig {
	return AuditConfig{
		Enabled:       true,
		LogFilePath:   "audit.log",
		MaxEvents:     10000,
		RetentionDays: 90,
		ConsoleOutput: true,
	}
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(config AuditConfig) (*AuditLogger, error) {
	al := &AuditLogger{
		events:        make([]*Event, 0),
		maxEvents:     config.MaxEvents,
		enabled:       config.Enabled,
		logFilePath:   config.LogFilePath,
		retentionDays: config.RetentionDays,
	}

	// Initialize zap logger
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.OutputPaths = []string{config.LogFilePath}
	logger, err := loggerConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to create audit logger: %w", err)
	}
	al.logger = logger

	// Open file logger
	if err := al.openFileLogger(); err != nil {
		return nil, err
	}

	return al, nil
}

// openFileLogger opens the audit log file
func (al *AuditLogger) openFileLogger() error {
	if al.logFilePath == "" {
		return nil
	}

	dir := filepath.Dir(al.logFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.OpenFile(al.logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	al.fileLogger = file
	return nil
}

// Log creates and logs a new audit event
func (al *AuditLogger) Log(ctx context.Context, event *Event) {
	al.mu.Lock()
	defer al.mu.Unlock()

	if !al.enabled {
		return
	}

	// Set timestamp if not set
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Generate ID if not set
	if event.ID == "" {
		event.ID = fmt.Sprintf("audit-%d", time.Now().UnixNano())
	}

	// Add to events list
	al.events = append(al.events, event)

	// Limit events
	if len(al.events) > al.maxEvents {
		al.events = al.events[1:]
	}

	// Log to zap
	al.logToZap(event)

	// Log to file
	al.logToFile(event)
}

// logToZap logs event to zap logger
func (al *AuditLogger) logToZap(event *Event) {
	fields := []zap.Field{
		zap.String("event_id", event.ID),
		zap.String("type", string(event.Type)),
		zap.String("level", string(event.Level)),
		zap.String("user", event.Username),
		zap.String("resource", event.Resource),
	}

	switch event.Level {
	case LevelCritical, LevelError:
		al.logger.Error(event.Message, fields...)
	case LevelWarning:
		al.logger.Warn(event.Message, fields...)
	default:
		al.logger.Info(event.Message, fields...)
	}
}

// logToFile logs event to file in JSON format
func (al *AuditLogger) logToFile(event *Event) {
	if al.fileLogger == nil {
		return
	}

	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	al.fileLogger.Write(data)
	al.fileLogger.Write([]byte("\n"))
}

// LogAuthEvent logs authentication events
func (al *AuditLogger) LogAuthEvent(ctx context.Context, eventType EventType, username, sourceIP, status string) {
	event := &Event{
		Level:    LevelAudit,
		Type:     eventType,
		Category: "authentication",
		Username: username,
		SourceIP: sourceIP,
		Status:   status,
		Message:  fmt.Sprintf("Authentication event: %s", eventType),
	}
	al.Log(ctx, event)
}

// LogBackupEvent logs backup job events
func (al *AuditLogger) LogBackupEvent(ctx context.Context, eventType EventType, userID, username, jobID, jobName, status string, details map[string]interface{}) {
	event := &Event{
		Level:      LevelInfo,
		Type:       eventType,
		Category:   "backup",
		UserID:     userID,
		Username:   username,
		Resource:   "backup_job",
		ResourceID: jobID,
		Status:     status,
		Message:    fmt.Sprintf("Backup job %s: %s", eventType, jobName),
		Details:    details,
	}
	al.Log(ctx, event)
}

// LogRestoreEvent logs restore events
func (al *AuditLogger) LogRestoreEvent(ctx context.Context, eventType EventType, userID, username, resourceID, status string, duration int64) {
	event := &Event{
		Level:      LevelInfo,
		Type:       eventType,
		Category:   "restore",
		UserID:     userID,
		Username:   username,
		Resource:   "restore_operation",
		ResourceID: resourceID,
		Status:     status,
		Duration:   duration,
		Message:    fmt.Sprintf("Restore operation %s", eventType),
	}
	al.Log(ctx, event)
}

// LogSecurityEvent logs security events
func (al *AuditLogger) LogSecurityEvent(ctx context.Context, eventType EventType, userID, username, message string, details map[string]interface{}) {
	level := LevelWarning
	if eventType == EventSecurityAlert {
		level = LevelCritical
	}

	event := &Event{
		Level:    level,
		Type:     eventType,
		Category: "security",
		UserID:   userID,
		Username: username,
		Message:  message,
		Details:  details,
	}
	al.Log(ctx, event)
}

// LogConfigChangeEvent logs configuration changes
func (al *AuditLogger) LogConfigChangeEvent(ctx context.Context, userID, username, resource, action, details string) {
	event := &Event{
		Level:    LevelAudit,
		Type:     EventConfigChange,
		Category: "configuration",
		UserID:   userID,
		Username: username,
		Resource: resource,
		Action:   action,
		Message:  fmt.Sprintf("Configuration change: %s %s", resource, action),
	}
	if details != "" {
		event.Details = map[string]interface{}{"details": details}
	}
	al.Log(ctx, event)
}

// GetEvents returns audit events with optional filtering
func (al *AuditLogger) GetEvents(ctx context.Context, filter *EventFilter) []*Event {
	al.mu.RLock()
	defer al.mu.RUnlock()

	if filter == nil {
		return al.events
	}

	var filtered []*Event
	for _, event := range al.events {
		if filter.Matches(event) {
			filtered = append(filtered, event)
		}
	}

	return filtered
}

// EventFilter filters audit events
type EventFilter struct {
	StartTime  time.Time
	EndTime    time.Time
	Level      EventLevel
	Type       EventType
	Category   string
	UserID     string
	ResourceID string
	Limit      int
}

// Matches checks if an event matches the filter
func (f *EventFilter) Matches(event *Event) bool {
	if !f.StartTime.IsZero() && event.Timestamp.Before(f.StartTime) {
		return false
	}
	if !f.EndTime.IsZero() && event.Timestamp.After(f.EndTime) {
		return false
	}
	if f.Level != "" && event.Level != f.Level {
		return false
	}
	if f.Type != "" && event.Type != f.Type {
		return false
	}
	if f.Category != "" && event.Category != f.Category {
		return false
	}
	if f.UserID != "" && event.UserID != f.UserID {
		return false
	}
	if f.ResourceID != "" && event.ResourceID != f.ResourceID {
		return false
	}
	return true
}

// GetStats returns audit statistics
func (al *AuditLogger) GetStats(ctx context.Context) *AuditStats {
	al.mu.RLock()
	defer al.mu.RUnlock()

	stats := &AuditStats{
		TotalEvents: len(al.events),
		ByLevel:     make(map[EventLevel]int),
		ByCategory:  make(map[string]int),
		ByType:      make(map[EventType]int),
	}

	for _, event := range al.events {
		stats.ByLevel[event.Level]++
		stats.ByCategory[event.Category]++
		stats.ByType[event.Type]++
	}

	return stats
}

// AuditStats contains audit statistics
type AuditStats struct {
	TotalEvents int                `json:"total_events"`
	ByLevel     map[EventLevel]int `json:"by_level"`
	ByCategory  map[string]int     `json:"by_category"`
	ByType      map[EventType]int  `json:"by_type"`
}

// CleanupOldEvents removes events older than retention period
func (al *AuditLogger) CleanupOldEvents() int {
	al.mu.Lock()
	defer al.mu.Unlock()

	cutoff := time.Now().AddDate(0, 0, -al.retentionDays)
	cleaned := 0

	var kept []*Event
	for _, event := range al.events {
		if event.Timestamp.After(cutoff) {
			kept = append(kept, event)
		} else {
			cleaned++
		}
	}

	al.events = kept
	return cleaned
}

// Close closes the audit logger
func (al *AuditLogger) Close() error {
	al.mu.Lock()
	defer al.mu.Unlock()

	if al.fileLogger != nil {
		al.fileLogger.Close()
	}

	if al.logger != nil {
		al.logger.Sync()
	}

	return nil
}

// ExportEvents exports events to JSON file
func (al *AuditLogger) ExportEvents(ctx context.Context, filter *EventFilter, filePath string) error {
	events := al.GetEvents(ctx, filter)

	data, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

// ImportEvents imports events from JSON file
func (al *AuditLogger) ImportEvents(ctx context.Context, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var events []*Event
	if err := json.Unmarshal(data, &events); err != nil {
		return err
	}

	al.mu.Lock()
	defer al.mu.Unlock()

	al.events = append(al.events, events...)
	return nil
}
