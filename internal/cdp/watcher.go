package cdp

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// PollingFileWatcher implements file system monitoring using polling
type PollingFileWatcher struct {
	config       CDPConfig
	paths        []string
	eventQueue   chan<- *FileEvent
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	isWatching   bool
	logger       *zap.Logger
	interval     time.Duration
	excludeGlobs []string
	includeGlobs []string
}

// NewPollingFileWatcher creates a new file watcher
func NewPollingFileWatcher(config CDPConfig) *PollingFileWatcher {
	logger, _ := zap.NewDevelopment()
	
	return &PollingFileWatcher{
		config:       config,
		logger:       logger,
		interval:     5 * time.Second,
		excludeGlobs: config.ExcludePatterns,
		includeGlobs: config.IncludePatterns,
	}
}

// Start begins watching specified paths
func (fw *PollingFileWatcher) Start(ctx context.Context, paths []string, eventQueue chan<- *FileEvent) error {
	if fw.isWatching {
		return fmt.Errorf("file watcher is already running")
	}

	fw.ctx, fw.cancel = context.WithCancel(ctx)
	fw.paths = paths
	fw.eventQueue = eventQueue
	fw.isWatching = true

	for _, path := range paths {
		fw.wg.Add(1)
		go fw.watchPath(path)
	}

	fw.logger.Info("File watcher started", 
		zap.Strings("paths", paths),
		zap.Duration("interval", fw.interval))

	return nil
}

// Stop stops watching
func (fw *PollingFileWatcher) Stop() {
	if !fw.isWatching {
		return
	}

	fw.cancel()
	fw.wg.Wait()
	fw.isWatching = false

	fw.logger.Info("File watcher stopped")
}

// IsWatching returns true if watcher is active
func (fw *PollingFileWatcher) IsWatching() bool {
	return fw.isWatching
}

// GetWatchedPaths returns list of watched paths
func (fw *PollingFileWatcher) GetWatchedPaths() []string {
	return fw.paths
}

// watchPath monitors a single path for changes
func (fw *PollingFileWatcher) watchPath(path string) {
	defer fw.wg.Done()

	fileStates := make(map[string]fileState)
	ticker := time.NewTicker(fw.interval)
	defer ticker.Stop()

	for {
		select {
		case <-fw.ctx.Done():
			return
		case <-ticker.C:
			fw.scanPath(path, fileStates)
		}
	}
}

// fileState holds information about a file's state
type fileState struct {
	size     int64
	modTime  time.Time
	checksum string
}

// scanPath scans a path for changes
func (fw *PollingFileWatcher) scanPath(path string, states map[string]fileState) {
	if !fw.shouldInclude(path) {
		return
	}

	err := filepath.WalkDir(path, func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			fw.logger.Warn("Error accessing path", zap.String("path", filePath), zap.Error(err))
			return nil
		}

		if d.IsDir() {
			return nil
		}

		if !fw.shouldInclude(filePath) {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		checksum, err := fw.calculateChecksum(filePath)
		if err != nil {
			fw.logger.Warn("Failed to calculate checksum", zap.String("path", filePath), zap.Error(err))
			return nil
		}

		oldState, exists := states[filePath]
		if !exists {
			fw.emitEvent(EventCreate, filePath, info.Size(), info.ModTime(), checksum)
			states[filePath] = fileState{size: info.Size(), modTime: info.ModTime(), checksum: checksum}
		} else if oldState.modTime != info.ModTime() || oldState.checksum != checksum {
			fw.emitEvent(EventModify, filePath, info.Size(), info.ModTime(), checksum)
			states[filePath] = fileState{size: info.Size(), modTime: info.ModTime(), checksum: checksum}
		}

		return nil
	})

	if err != nil {
		fw.logger.Warn("Error walking path", zap.String("path", path), zap.Error(err))
	}

	for filePath := range states {
		if !strings.HasPrefix(filePath, path) {
			continue
		}
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			fw.emitEvent(EventDelete, filePath, 0, time.Now(), "")
			delete(states, filePath)
		}
	}
}

// shouldInclude checks if a path should be included based on filters
func (fw *PollingFileWatcher) shouldInclude(path string) bool {
	for _, pattern := range fw.excludeGlobs {
		if match, _ := filepath.Match(pattern, filepath.Base(path)); match {
			return false
		}
	}

	if len(fw.includeGlobs) > 0 {
		for _, pattern := range fw.includeGlobs {
			if match, _ := filepath.Match(pattern, filepath.Base(path)); match {
				return true
			}
		}
		return false
	}

	return true
}

// calculateChecksum calculates SHA256 checksum of a file
func (fw *PollingFileWatcher) calculateChecksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// emitEvent creates and sends a file event
func (fw *PollingFileWatcher) emitEvent(eventType EventType, path string, size int64, modTime time.Time, checksum string) {
	event := &FileEvent{
		ID:        uuid.New().String(),
		Type:      eventType,
		Path:      path,
		Size:      size,
		ModTime:   modTime,
		Checksum:  checksum,
		CreatedAt: time.Now(),
		Metadata:  make(map[string]string),
	}

	select {
	case fw.eventQueue <- event:
		fw.logger.Debug("Event emitted", zap.String("type", string(eventType)), zap.String("path", path))
	default:
		fw.logger.Warn("Event queue full, dropping event", zap.String("path", path))
	}
}
