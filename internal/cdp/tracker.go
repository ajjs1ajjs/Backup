package cdp

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// FileChange represents a single file change record
type FileChange struct {
	ID           string            `json:"id"`
	Path         string            `json:"path"`
	Type         EventType         `json:"type"`
	Size         int64             `json:"size"`
	OldSize      int64             `json:"old_size,omitempty"`
	Checksum     string            `json:"checksum"`
	OldChecksum  string            `json:"old_checksum,omitempty"`
	ModTime      time.Time         `json:"mod_time"`
	TenantID     string            `json:"tenant_id"`
	WatchedPath  string            `json:"watched_path"`
	CreatedAt    time.Time         `json:"created_at"`
	Processed    bool              `json:"processed"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	Error        string            `json:"error,omitempty"`
}

// ChangeLog represents a collection of file changes
type ChangeLog struct {
	Changes    []*FileChange `json:"changes"`
	TotalCount int64         `json:"total_count"`
	StartTime  time.Time     `json:"start_time"`
	EndTime    time.Time     `json:"end_time"`
}

// PathConfig holds configuration for a watched path
type PathConfig struct {
	Path             string        `json:"path"`
	Enabled          bool          `json:"enabled"`
	Recursive        bool          `json:"recursive"`
	MaxDepth         int           `json:"max_depth"`
	ExcludePatterns  []string      `json:"exclude_patterns"`
	IncludePatterns  []string      `json:"include_patterns"`
	MaxChangeAge     time.Duration `json:"max_change_age"`
	MaxChangesCount  int           `json:"max_changes_count"`
	CreatedAt        time.Time     `json:"created_at"`
	LastModified     time.Time     `json:"last_modified"`
	TotalChanges     int64         `json:"total_changes"`
}

// TrackerStats contains statistics for the change tracker
type TrackerStats struct {
	TotalChanges      int64         `json:"total_changes"`
	ProcessedChanges  int64         `json:"processed_changes"`
	PendingChanges    int64         `json:"pending_changes"`
	FailedChanges     int64         `json:"failed_changes"`
	WatchedPaths      int           `json:"watched_paths"`
	ActivePaths       int           `json:"active_paths"`
	TotalTrackedSize  int64         `json:"total_tracked_size"`
	LastChangeTime    *time.Time    `json:"last_change_time"`
	TrackerUptime     time.Duration `json:"tracker_uptime"`
	ChangesPerSecond  float64       `json:"changes_per_second"`
	AverageChangeSize int64         `json:"average_change_size"`
}

// ChangeTracker manages file change history and tracking
type ChangeTracker struct {
	config         CDPConfig
	logger         *zap.Logger
	mutex          sync.RWMutex
	changes        []*FileChange
	watchedPaths   map[string]*PathConfig
	stats          *TrackerStats
	startTime      time.Time
	maxHistorySize int
	cleanupEnabled bool
}

// NewChangeTracker creates a new change tracker instance
func NewChangeTracker(config CDPConfig) *ChangeTracker {
	logger, _ := zap.NewDevelopment()

	return &ChangeTracker{
		config:         config,
		logger:         logger,
		changes:        make([]*FileChange, 0),
		watchedPaths:   make(map[string]*PathConfig),
		stats: &TrackerStats{
			LastChangeTime: nil,
		},
		startTime:      time.Now(),
		maxHistorySize: 10000,
		cleanupEnabled: true,
	}
}

// TrackChange records a file change event
func (ct *ChangeTracker) TrackChange(path string, eventType EventType, size int64, checksum string, tenantID string, metadata map[string]string) (*FileChange, error) {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	// Resolve to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check if path is being watched
	watchedPath := ct.findWatchedPath(absPath)
	if watchedPath == "" {
		return nil, fmt.Errorf("path %s is not being watched", absPath)
	}

	pathConfig := ct.watchedPaths[watchedPath]
	if !pathConfig.Enabled {
		return nil, fmt.Errorf("watched path %s is disabled", watchedPath)
	}

	// Get old change info if exists
	var oldSize int64
	var oldChecksum string
	for i := len(ct.changes) - 1; i >= 0; i-- {
		if ct.changes[i].Path == absPath {
			oldSize = ct.changes[i].Size
			oldChecksum = ct.changes[i].Checksum
			break
		}
	}

	// Create the file change record
	change := &FileChange{
		ID:          uuid.New().String(),
		Path:        absPath,
		Type:        eventType,
		Size:        size,
		OldSize:     oldSize,
		Checksum:    checksum,
		OldChecksum: oldChecksum,
		ModTime:     time.Now(),
		TenantID:    tenantID,
		WatchedPath: watchedPath,
		CreatedAt:   time.Now(),
		Processed:   false,
		Metadata:    metadata,
	}

	// Add to changes list
	ct.changes = append(ct.changes, change)

	// Update statistics
	ct.stats.TotalChanges++
	ct.stats.PendingChanges++
	ct.stats.TotalTrackedSize += size

	now := time.Now()
	ct.stats.LastChangeTime = &now

	// Update path config stats
	pathConfig.TotalChanges++
	pathConfig.LastModified = now

	// Calculate average change size
	if ct.stats.TotalChanges > 0 {
		ct.stats.AverageChangeSize = ct.stats.TotalTrackedSize / ct.stats.TotalChanges
	}

	// Calculate changes per second
	uptime := time.Since(ct.startTime).Seconds()
	if uptime > 0 {
		ct.stats.ChangesPerSecond = float64(ct.stats.TotalChanges) / uptime
	}

	ct.logger.Debug("Change tracked",
		zap.String("id", change.ID),
		zap.String("path", change.Path),
		zap.String("type", string(change.Type)),
		zap.Int64("size", change.Size))

	return change, nil
}

// GetChangeHistory returns change history with optional filtering
func (ct *ChangeTracker) GetChangeHistory(limit int, offset int, pathFilter string) (*ChangeLog, error) {
	ct.mutex.RLock()
	defer ct.mutex.RUnlock()

	// Filter changes if path filter is provided
	var filtered []*FileChange
	if pathFilter != "" {
		for _, change := range ct.changes {
			if change.Path == pathFilter || change.WatchedPath == pathFilter {
				filtered = append(filtered, change)
			}
		}
	} else {
		filtered = ct.changes
	}

	// Apply pagination
	totalCount := int64(len(filtered))
	if offset >= len(filtered) {
		return &ChangeLog{
			Changes:    make([]*FileChange, 0),
			TotalCount: totalCount,
		}, nil
	}

	start := offset
	end := start + limit
	if limit <= 0 || end > len(filtered) {
		end = len(filtered)
	}

	result := make([]*FileChange, end-start)
	copy(result, filtered[start:end])

	// Determine time range
	var startTime, endTime time.Time
	if len(result) > 0 {
		startTime = result[0].CreatedAt
		endTime = result[len(result)-1].CreatedAt
	}

	return &ChangeLog{
		Changes:    result,
		TotalCount: totalCount,
		StartTime:  startTime,
		EndTime:    endTime,
	}, nil
}

// GetChangesSince returns all changes since a specific time
func (ct *ChangeTracker) GetChangesSince(since time.Time, pathFilter string) (*ChangeLog, error) {
	ct.mutex.RLock()
	defer ct.mutex.RUnlock()

	var filtered []*FileChange
	for _, change := range ct.changes {
		if change.CreatedAt.After(since) {
			if pathFilter == "" || change.Path == pathFilter || change.WatchedPath == pathFilter {
				filtered = append(filtered, change)
			}
		}
	}

	// Determine time range
	var startTime, endTime time.Time
	if len(filtered) > 0 {
		startTime = filtered[0].CreatedAt
		endTime = filtered[len(filtered)-1].CreatedAt
	}

	return &ChangeLog{
		Changes:    filtered,
		TotalCount: int64(len(filtered)),
		StartTime:  startTime,
		EndTime:    endTime,
	}, nil
}

// AddWatchedPath adds a path to be watched
func (ct *ChangeTracker) AddWatchedPath(path string, recursive bool, excludePatterns []string, includePatterns []string) error {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	// Resolve to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check if already watched
	if _, exists := ct.watchedPaths[absPath]; exists {
		return fmt.Errorf("path %s is already being watched", absPath)
	}

	// Create path configuration
	pathConfig := &PathConfig{
		Path:            absPath,
		Enabled:         true,
		Recursive:       recursive,
		MaxDepth:        -1, // unlimited
		ExcludePatterns: excludePatterns,
		IncludePatterns: includePatterns,
		MaxChangeAge:    24 * time.Hour,
		MaxChangesCount: ct.maxHistorySize,
		CreatedAt:       time.Now(),
		LastModified:    time.Now(),
		TotalChanges:    0,
	}

	ct.watchedPaths[absPath] = pathConfig
	ct.stats.WatchedPaths++
	ct.stats.ActivePaths++

	ct.logger.Info("Watched path added",
		zap.String("path", absPath),
		zap.Bool("recursive", recursive))

	return nil
}

// RemoveWatchedPath removes a path from being watched
func (ct *ChangeTracker) RemoveWatchedPath(path string) error {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	// Resolve to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check if path is watched
	if _, exists := ct.watchedPaths[absPath]; !exists {
		return fmt.Errorf("path %s is not being watched", absPath)
	}

	delete(ct.watchedPaths, absPath)
	ct.stats.WatchedPaths--

	ct.logger.Info("Watched path removed", zap.String("path", absPath))

	return nil
}

// GetStats returns current tracker statistics
func (ct *ChangeTracker) GetStats() *TrackerStats {
	ct.mutex.RLock()
	defer ct.mutex.RUnlock()

	// Calculate pending and failed changes
	var pending, failed int64
	for _, change := range ct.changes {
		if !change.Processed {
			pending++
		}
		if change.Error != "" {
			failed++
		}
	}

	ct.stats.PendingChanges = pending
	ct.stats.FailedChanges = failed
	ct.stats.ProcessedChanges = ct.stats.TotalChanges - pending

	// Update uptime
	if ct.startTime.IsZero() {
		ct.stats.TrackerUptime = 0
	} else {
		ct.stats.TrackerUptime = time.Since(ct.startTime)
	}

	// Recalculate changes per second
	uptime := ct.stats.TrackerUptime.Seconds()
	if uptime > 0 {
		ct.stats.ChangesPerSecond = float64(ct.stats.TotalChanges) / uptime
	}

	// Create a copy to prevent race conditions
	statsCopy := *ct.stats
	return &statsCopy
}

// CleanupOldChanges removes changes older than the configured retention period
func (ct *ChangeTracker) CleanupOldChanges() (int64, error) {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	if !ct.cleanupEnabled {
		return 0, nil
	}

	// Determine cutoff time based on the minimum MaxChangeAge from all path configs
	cutoffTime := time.Now().Add(-24 * time.Hour) // default
	for _, pathConfig := range ct.watchedPaths {
		if pathConfig.MaxChangeAge > 0 && time.Now().Add(-pathConfig.MaxChangeAge).Before(cutoffTime) {
			cutoffTime = time.Now().Add(-pathConfig.MaxChangeAge)
		}
	}

	// Find changes to remove
	var changesToRemove int64
	var retained []*FileChange

	for _, change := range ct.changes {
		if change.CreatedAt.Before(cutoffTime) {
			changesToRemove++
			ct.stats.TotalTrackedSize -= change.Size
		} else {
			retained = append(retained, change)
		}
	}

	// Also enforce max changes count per path
	pathChangeCounts := make(map[string]int)
	for i := len(retained) - 1; i >= 0; i-- {
		change := retained[i]
		pathChangeCounts[change.WatchedPath]++

		pathConfig, exists := ct.watchedPaths[change.WatchedPath]
		maxCount := ct.maxHistorySize
		if exists && pathConfig.MaxChangesCount > 0 {
			maxCount = pathConfig.MaxChangesCount
		}

		if pathChangeCounts[change.WatchedPath] > maxCount {
			changesToRemove++
			ct.stats.TotalTrackedSize -= change.Size
			// Remove from retained by setting to nil
			retained[i] = nil
		}
	}

	// Compact the retained slice
	var compacted []*FileChange
	for _, change := range retained {
		if change != nil {
			compacted = append(compacted, change)
		}
	}

	ct.changes = compacted

	if changesToRemove > 0 {
		ct.logger.Info("Cleanup completed",
			zap.Int64("changes_removed", changesToRemove),
			zap.Int("changes_remaining", len(ct.changes)))
	}

	return changesToRemove, nil
}

// EnablePath enables change tracking for a watched path
func (ct *ChangeTracker) EnablePath(path string) error {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	pathConfig, exists := ct.watchedPaths[absPath]
	if !exists {
		return fmt.Errorf("path %s is not being watched", absPath)
	}

	pathConfig.Enabled = true
	pathConfig.LastModified = time.Now()
	ct.stats.ActivePaths++

	ct.logger.Info("Path enabled", zap.String("path", absPath))
	return nil
}

// DisablePath disables change tracking for a watched path
func (ct *ChangeTracker) DisablePath(path string) error {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	pathConfig, exists := ct.watchedPaths[absPath]
	if !exists {
		return fmt.Errorf("path %s is not being watched", absPath)
	}

	pathConfig.Enabled = false
	pathConfig.LastModified = time.Now()
	ct.stats.ActivePaths--

	ct.logger.Info("Path disabled", zap.String("path", absPath))
	return nil
}

// MarkChangeProcessed marks a change as processed
func (ct *ChangeTracker) MarkChangeProcessed(changeID string) error {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	for _, change := range ct.changes {
		if change.ID == changeID {
			change.Processed = true
			ct.stats.ProcessedChanges++
			ct.stats.PendingChanges--
			return nil
		}
	}

	return fmt.Errorf("change %s not found", changeID)
}

// MarkChangeFailed marks a change as failed with an error message
func (ct *ChangeTracker) MarkChangeFailed(changeID string, errMsg string) error {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	for _, change := range ct.changes {
		if change.ID == changeID {
			change.Processed = true
			change.Error = errMsg
			ct.stats.FailedChanges++
			ct.stats.PendingChanges--
			return nil
		}
	}

	return fmt.Errorf("change %s not found", changeID)
}

// GetWatchedPaths returns all watched paths
func (ct *ChangeTracker) GetWatchedPaths() []*PathConfig {
	ct.mutex.RLock()
	defer ct.mutex.RUnlock()

	paths := make([]*PathConfig, 0, len(ct.watchedPaths))
	for _, pathConfig := range ct.watchedPaths {
		paths = append(paths, pathConfig)
	}

	return paths
}

// GetPathConfig returns configuration for a specific watched path
func (ct *ChangeTracker) GetPathConfig(path string) (*PathConfig, error) {
	ct.mutex.RLock()
	defer ct.mutex.RUnlock()

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path: %w", err)
	}

	pathConfig, exists := ct.watchedPaths[absPath]
	if !exists {
		return nil, fmt.Errorf("path %s is not being watched", absPath)
	}

	return pathConfig, nil
}

// findWatchedPath finds which watched path contains the given path
func (ct *ChangeTracker) findWatchedPath(path string) string {
	for watchedPath := range ct.watchedPaths {
		pathConfig := ct.watchedPaths[watchedPath]
		if pathConfig.Recursive {
			// Check if path is under the watched directory
			if filepath.HasPrefix(path, watchedPath) {
				return watchedPath
			}
		} else {
			// Exact match only
			if path == watchedPath {
				return watchedPath
			}
		}
	}
	return ""
}

// SetMaxHistorySize sets the maximum number of changes to retain in history
func (ct *ChangeTracker) SetMaxHistorySize(size int) {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	ct.maxHistorySize = size
	ct.logger.Debug("Max history size updated", zap.Int("size", size))
}

// EnableCleanup enables or disables automatic cleanup
func (ct *ChangeTracker) EnableCleanup(enabled bool) {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	ct.cleanupEnabled = enabled
	ct.logger.Debug("Cleanup setting updated", zap.Bool("enabled", enabled))
}

// GetChangeByID retrieves a specific change by ID
func (ct *ChangeTracker) GetChangeByID(changeID string) (*FileChange, error) {
	ct.mutex.RLock()
	defer ct.mutex.RUnlock()

	for _, change := range ct.changes {
		if change.ID == changeID {
			return change, nil
		}
	}

	return nil, fmt.Errorf("change %s not found", changeID)
}

// GetUnprocessedChanges returns all changes that haven't been processed yet
func (ct *ChangeTracker) GetUnprocessedChanges() []*FileChange {
	ct.mutex.RLock()
	defer ct.mutex.RUnlock()

	var unprocessed []*FileChange
	for _, change := range ct.changes {
		if !change.Processed {
			unprocessed = append(unprocessed, change)
		}
	}

	return unprocessed
}

// ClearHistory removes all change history
func (ct *ChangeTracker) ClearHistory() {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	ct.changes = make([]*FileChange, 0)
	ct.stats.TotalChanges = 0
	ct.stats.ProcessedChanges = 0
	ct.stats.PendingChanges = 0
	ct.stats.FailedChanges = 0
	ct.stats.TotalTrackedSize = 0
	ct.stats.LastChangeTime = nil

	ct.logger.Info("Change history cleared")
}