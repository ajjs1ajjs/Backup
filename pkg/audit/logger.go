// Package audit provides comprehensive audit logging functionality
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

// EventType represents the type of audit event
type EventType string

// EventSeverity represents the severity of an audit event
type EventSeverity string

const (
	// Event Types
	EventTypeUserLogin        EventType = "user_login"
	EventTypeUserLogout       EventType = "user_logout"
	EventTypeUserCreate       EventType = "user_create"
	EventTypeUserUpdate       EventType = "user_update"
	EventTypeUserDelete       EventType = "user_delete"
	EventTypeUserFailedLogin  EventType = "user_failed_login"

	EventTypeJobCreate    EventType = "job_create"
	EventTypeJobUpdate    EventType = "job_update"
	EventTypeJobDelete    EventType = "job_delete"
	EventTypeJobStart     EventType = "job_start"
	EventTypeJobComplete  EventType = "job_complete"
	EventTypeJobFailed    EventType = "job_failed"
	EventTypeJobCancelled EventType = "job_cancelled"

	EventTypeBackupCreate   EventType = "backup_create"
	EventTypeBackupDelete   EventType = "backup_delete"
	EventTypeBackupRestore  EventType = "backup_restore"
	EventTypeBackupVerify   EventType = "backup_verify"

	EventTypeVMBackup      EventType = "vm_backup"
	EventTypeVMRestore     EventType = "vm_restore"
	EventTypeVMSnapshot    EventType = "vm_snapshot"
	EventTypeInstantRecovery EventType = "instant_recovery"

	EventTypeConfigChange  EventType = "config_change"
	EventTypeSecurityAlert EventType = "security_alert"
	EventTypePermissionChange EventType = "permission_change"

	// Event Severities
	SeverityInfo     EventSeverity = "info"
	SeverityWarning  EventSeverity = "warning"
	SeverityError    EventSeverity = "error"
	SeverityCritical EventSeverity = "critical"
)

// AuditEvent represents a single audit event
type AuditEvent struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	Type        EventType              `json:"type"`
	Severity    EventSeverity          `json:"severity"`
	UserID      string                 `json:"user_id,omitempty"`
	Username    string                 `json:"username,omitempty"`
	SessionID   string                 `json:"session_id,omitempty"`
	IPAddress   string                 `json:"ip_address,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	Resource    string                 `json:"resource"`
	Action      string                 `json:"action"`
	Status      string                 `json:"status"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Duration    time.Duration          `json:"duration,omitempty"`
}

// AuditLogger manages audit logging
type AuditLogger struct {
	logger     *zap.Logger
	logDir     string
	logFile    *os.File
	mu         sync.Mutex
	events     []*AuditEvent
	maxEvents  int
	writeToFile bool
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(logger *zap.Logger, logDir string, maxEvents int) (*AuditLogger, error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create audit log directory: %w", err)
	}

	logPath := filepath.Join(logDir, "audit.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log file: %w", err)
	}

	return &AuditLogger{
		logger:      logger.With(zap.String("component", "audit")),
		logDir:      logDir,
		logFile:     logFile,
		events:      make([]*AuditEvent, 0),
		maxEvents:   maxEvents,
		writeToFile: true,
	}, nil
}

// Log logs an audit event
func (a *AuditLogger) Log(event *AuditEvent) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Set defaults
	if event.ID == "" {
		event.ID = fmt.Sprintf("evt_%d_%d", time.Now().Unix(), len(a.events))
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	if event.Severity == "" {
		event.Severity = SeverityInfo
	}

	// Add to memory buffer
	a.events = append(a.events, event)

	// Trim buffer if too large
	if len(a.events) > a.maxEvents {
		a.events = a.events[len(a.events)-a.maxEvents:]
	}

	// Write to file
	if a.writeToFile {
		data, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal audit event: %w", err)
		}

		if _, err := a.logFile.Write(append(data, '\n')); err != nil {
			return fmt.Errorf("failed to write audit event: %w", err)
		}
	}

	a.logger.Debug("Audit event logged",
		zap.String("id", event.ID),
		zap.String("type", string(event.Type)),
		zap.String("user", event.Username))

	return nil
}

// LogUserLogin logs a user login event
func (a *AuditLogger) LogUserLogin(userID, username, sessionID, ipAddress, userAgent string, success bool, errorMsg string) error {
	severity := SeverityInfo
	status := "success"
	if !success {
		severity = SeverityWarning
		status = "failed"
	}

	return a.Log(&AuditEvent{
		Type:      EventTypeUserLogin,
		Severity:  severity,
		UserID:    userID,
		Username:  username,
		SessionID: sessionID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Resource:  "authentication",
		Action:    "login",
		Status:    status,
		Error:     errorMsg,
	})
}

// LogUserLogout logs a user logout event
func (a *AuditLogger) LogUserLogout(userID, username, sessionID string) error {
	return a.Log(&AuditEvent{
		Type:      EventTypeUserLogout,
		Severity:  SeverityInfo,
		UserID:    userID,
		Username:  username,
		SessionID: sessionID,
		Resource:  "authentication",
		Action:    "logout",
		Status:    "success",
	})
}

// LogJobStart logs a job start event
func (a *AuditLogger) LogJobStart(userID, username, jobID, jobName string, details map[string]interface{}) error {
	return a.Log(&AuditEvent{
		Type:     EventTypeJobStart,
		Severity: SeverityInfo,
		UserID:   userID,
		Username: username,
		Resource: fmt.Sprintf("job:%s", jobID),
		Action:   "start",
		Status:   "started",
		Details:  details,
	})
}

// LogJobComplete logs a job completion event
func (a *AuditLogger) LogJobComplete(userID, username, jobID string, duration time.Duration, details map[string]interface{}) error {
	return a.Log(&AuditEvent{
		Type:     EventTypeJobComplete,
		Severity: SeverityInfo,
		UserID:   userID,
		Username: username,
		Resource: fmt.Sprintf("job:%s", jobID),
		Action:   "complete",
		Status:   "success",
		Duration: duration,
		Details:  details,
	})
}

// LogJobFailed logs a job failure event
func (a *AuditLogger) LogJobFailed(userID, username, jobID string, errorMsg string, details map[string]interface{}) error {
	return a.Log(&AuditEvent{
		Type:     EventTypeJobFailed,
		Severity: SeverityError,
		UserID:   userID,
		Username: username,
		Resource: fmt.Sprintf("job:%s", jobID),
		Action:   "execute",
		Status:   "failed",
		Error:    errorMsg,
		Details:  details,
	})
}

// LogBackupCreate logs a backup creation event
func (a *AuditLogger) LogBackupCreate(userID, username, jobID, backupID string, details map[string]interface{}) error {
	return a.Log(&AuditEvent{
		Type:     EventTypeBackupCreate,
		Severity: SeverityInfo,
		UserID:   userID,
		Username: username,
		Resource: fmt.Sprintf("backup:%s", backupID),
		Action:   "create",
		Status:   "success",
		Details:  details,
	})
}

// LogBackupDelete logs a backup deletion event
func (a *AuditLogger) LogBackupDelete(userID, username, backupID string, details map[string]interface{}) error {
	return a.Log(&AuditEvent{
		Type:     EventTypeBackupDelete,
		Severity: SeverityWarning,
		UserID:   userID,
		Username: username,
		Resource: fmt.Sprintf("backup:%s", backupID),
		Action:   "delete",
		Status:   "success",
		Details:  details,
	})
}

// LogBackupRestore logs a restore event
func (a *AuditLogger) LogBackupRestore(userID, username, backupID, targetVM string, details map[string]interface{}) error {
	return a.Log(&AuditEvent{
		Type:     EventTypeBackupRestore,
		Severity: SeverityInfo,
		UserID:   userID,
		Username: username,
		Resource: fmt.Sprintf("backup:%s", backupID),
		Action:   "restore",
		Status:   "success",
		Details:  details,
	})
}

// LogConfigChange logs a configuration change event
func (a *AuditLogger) LogConfigChange(userID, username, configKey string, oldValue, newValue interface{}) error {
	return a.Log(&AuditEvent{
		Type:     EventTypeConfigChange,
		Severity: SeverityWarning,
		UserID:   userID,
		Username: username,
		Resource: fmt.Sprintf("config:%s", configKey),
		Action:   "update",
		Status:   "success",
		Details: map[string]interface{}{
			"old_value": oldValue,
			"new_value": newValue,
		},
	})
}

// LogSecurityAlert logs a security alert event
func (a *AuditLogger) LogSecurityAlert(alertType, description, ipAddress string, severity EventSeverity) error {
	return a.Log(&AuditEvent{
		Type:      EventTypeSecurityAlert,
		Severity:  severity,
		IPAddress: ipAddress,
		Resource:  "security",
		Action:    alertType,
		Status:    "alert",
		Error:     description,
	})
}

// LogPermissionChange logs a permission change event
func (a *AuditLogger) LogPermissionChange(userID, username, targetUserID, action string, permissions []string) error {
	return a.Log(&AuditEvent{
		Type:     EventTypePermissionChange,
		Severity: SeverityWarning,
		UserID:   userID,
		Username: username,
		Resource: fmt.Sprintf("user:%s", targetUserID),
		Action:   action,
		Status:   "success",
		Details: map[string]interface{}{
			"target_user": targetUserID,
			"permissions": permissions,
		},
	})
}

// LogInstantRecovery logs an instant recovery event
func (a *AuditLogger) LogInstantRecovery(userID, username, backupID, vmName string, details map[string]interface{}) error {
	return a.Log(&AuditEvent{
		Type:     EventTypeInstantRecovery,
		Severity: SeverityInfo,
		UserID:   userID,
		Username: username,
		Resource: fmt.Sprintf("backup:%s", backupID),
		Action:   "instant_recovery",
		Status:   "success",
		Details:  details,
	})
}

// GetRecentEvents returns recent audit events
func (a *AuditLogger) GetRecentEvents(count int) []*AuditEvent {
	a.mu.Lock()
	defer a.mu.Unlock()

	if count > len(a.events) {
		count = len(a.events)
	}

	// Return a copy of the most recent events
	result := make([]*AuditEvent, count)
	for i := 0; i < count; i++ {
		idx := len(a.events) - count + i
		result[i] = a.events[idx]
	}

	return result
}

// GetEventsByType returns events filtered by type
func (a *AuditLogger) GetEventsByType(eventType EventType, count int) []*AuditEvent {
	a.mu.Lock()
	defer a.mu.Unlock()

	var result []*AuditEvent
	for i := len(a.events) - 1; i >= 0 && len(result) < count; i-- {
		if a.events[i].Type == eventType {
			result = append(result, a.events[i])
		}
	}

	return result
}

// GetEventsByUser returns events filtered by user
func (a *AuditLogger) GetEventsByUser(userID string, count int) []*AuditEvent {
	a.mu.Lock()
	defer a.mu.Unlock()

	var result []*AuditEvent
	for i := len(a.events) - 1; i >= 0 && len(result) < count; i-- {
		if a.events[i].UserID == userID {
			result = append(result, a.events[i])
		}
	}

	return result
}

// GetEventsBySeverity returns events filtered by severity
func (a *AuditLogger) GetEventsBySeverity(severity EventSeverity, count int) []*AuditEvent {
	a.mu.Lock()
	defer a.mu.Unlock()

	var result []*AuditEvent
	for i := len(a.events) - 1; i >= 0 && len(result) < count; i-- {
		if a.events[i].Severity == severity {
			result = append(result, a.events[i])
		}
	}

	return result
}

// SearchEvents searches events by various criteria
func (a *AuditLogger) SearchEvents(startTime, endTime time.Time, eventType EventType, severity EventSeverity, userID string) []*AuditEvent {
	a.mu.Lock()
	defer a.mu.Unlock()

	var result []*AuditEvent
	for _, event := range a.events {
		// Check time range
		if event.Timestamp.Before(startTime) || event.Timestamp.After(endTime) {
			continue
		}

		// Check type
		if eventType != "" && event.Type != eventType {
			continue
		}

		// Check severity
		if severity != "" && event.Severity != severity {
			continue
		}

		// Check user
		if userID != "" && event.UserID != userID {
			continue
		}

		result = append(result, event)
	}

	return result
}

// ExportEvents exports events to JSON file
func (a *AuditLogger) ExportEvents(filename string, events []*AuditEvent) error {
	data, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal events: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write events file: %w", err)
	}

	return nil
}

// RotateLog rotates the audit log file
func (a *AuditLogger) RotateLog() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Close current log file
	if err := a.logFile.Close(); err != nil {
		return fmt.Errorf("failed to close log file: %w", err)
	}

	// Rename current log file with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	oldPath := filepath.Join(a.logDir, "audit.log")
	newPath := filepath.Join(a.logDir, fmt.Sprintf("audit_%s.log", timestamp))

	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to rotate log file: %w", err)
	}

	// Open new log file
	logPath := filepath.Join(a.logDir, "audit.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open new log file: %w", err)
	}

	a.logFile = logFile
	a.logger.Info("Audit log rotated", zap.String("new_file", newPath))

	return nil
}

// Close closes the audit logger
func (a *AuditLogger) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if err := a.logFile.Close(); err != nil {
		return fmt.Errorf("failed to close audit log file: %w", err)
	}

	return nil
}

// AuditMiddleware creates a middleware for automatic audit logging
func (a *AuditLogger) AuditMiddleware(eventType EventType, action string) func(userID, username string, details map[string]interface{}) {
	return func(userID, username string, details map[string]interface{}) {
		a.Log(&AuditEvent{
			Type:     eventType,
			Severity: SeverityInfo,
			UserID:   userID,
			Username: username,
			Resource: "api",
			Action:   action,
			Status:   "success",
			Details:  details,
		})
	}
}
