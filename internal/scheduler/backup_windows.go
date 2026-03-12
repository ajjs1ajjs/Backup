package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// BackupWindow defines a time window when backups are allowed
type BackupWindow struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	DayOfWeek   int       `json:"day_of_week"` // 0=Sunday, 6=Saturday, -1=everyday
	StartTime   string    `json:"start_time"`  // "22:00"
	EndTime     string    `json:"end_time"`    // "06:00"
	Enabled     bool      `json:"enabled"`
	AllowBackup bool      `json:"allow_backup"` // true=allow, false=deny
	CreatedAt   time.Time `json:"created_at"`
}

// BackupWindowManager manages backup windows
type BackupWindowManager struct {
	windows map[string]*BackupWindow
	mu      sync.RWMutex
}

// NewBackupWindowManager creates a new backup window manager
func NewBackupWindowManager() *BackupWindowManager {
	return &BackupWindowManager{
		windows: make(map[string]*BackupWindow),
	}
}

// AddWindow adds a new backup window
func (m *BackupWindowManager) AddWindow(ctx context.Context, window *BackupWindow) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if window.ID == "" {
		window.ID = fmt.Sprintf("bw-%d", time.Now().Unix())
	}
	window.CreatedAt = time.Now()
	m.windows[window.ID] = window
	return nil
}

// RemoveWindow removes a backup window
func (m *BackupWindowManager) RemoveWindow(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.windows[id]; !ok {
		return fmt.Errorf("window %s not found", id)
	}
	delete(m.windows, id)
	return nil
}

// IsBackupAllowed checks if backup is allowed at current time
func (m *BackupWindowManager) IsBackupAllowed() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	now := time.Now()
	currentDay := int(now.Weekday())
	currentTime := now.Format("15:04")

	// If no windows defined, allow by default
	if len(m.windows) == 0 {
		return true
	}

	for _, window := range m.windows {
		if !window.Enabled {
			continue
		}

		// Check day
		dayMatch := (window.DayOfWeek == -1) || (window.DayOfWeek == currentDay)
		if !dayMatch {
			continue
		}

		// Check time window (handles overnight windows like 22:00-06:00)
		if window.StartTime <= window.EndTime {
			// Normal window (e.g., 14:00-18:00)
			if currentTime >= window.StartTime && currentTime <= window.EndTime {
				return window.AllowBackup
			}
		} else {
			// Overnight window (e.g., 22:00-06:00)
			if currentTime >= window.StartTime || currentTime <= window.EndTime {
				return window.AllowBackup
			}
		}
	}

	// No matching window found - default to deny if any windows exist
	if len(m.windows) > 0 {
		return false
	}
	return true
}

// GetWindows returns all backup windows
func (m *BackupWindowManager) GetWindows() []*BackupWindow {
	m.mu.RLock()
	defer m.mu.RUnlock()

	windows := make([]*BackupWindow, 0, len(m.windows))
	for _, w := range m.windows {
		windows = append(windows, w)
	}
	return windows
}
