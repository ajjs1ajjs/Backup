package synthetic

import (
	"context"
	"fmt"
	"sync"
	"time"

	"novabackup/internal/deduplication"
	"novabackup/internal/multitenancy"

	"github.com/google/uuid"
)

// SyntheticBackupManager manages synthetic full backup operations
type SyntheticBackupManager interface {
	// Synthetic backup operations
	CreateSyntheticBackup(ctx context.Context, request *SyntheticBackupRequest) (*SyntheticBackupResult, error)
	GetSyntheticBackup(ctx context.Context, backupID string) (*SyntheticBackup, error)
	ListSyntheticBackups(ctx context.Context, filter *SyntheticBackupFilter) ([]SyntheticBackup, error)
	DeleteSyntheticBackup(ctx context.Context, backupID string) error

	// Incremental merge operations
	MergeIncrementals(ctx context.Context, request *MergeRequest) (*MergeResult, error)
	GetBackupChain(ctx context.Context, filter *ChainFilter) (*BackupChain, error)

	// Statistics and monitoring
	GetSyntheticStats(ctx context.Context) (*SyntheticStats, error)
	GetBackupChainStats(ctx context.Context) (*ChainStats, error)
}

// SyntheticBackupRequest represents a synthetic backup creation request
type SyntheticBackupRequest struct {
	SourceRepo       string                 `json:"source_repo"`
	TargetRepo       string                 `json:"target_repo"`
	BackupType       string                 `json:"backup_type"`
	IncrementalSince *time.Time             `json:"incremental_since"`
	Compression      bool                   `json:"compression"`
	RetentionDays    int                    `json:"retention_days"`
	TenantID         string                 `json:"tenant_id"`
	Settings         map[string]interface{} `json:"settings"`
	Metadata         map[string]string      `json:"metadata"`
}

// SyntheticBackupResult represents a synthetic backup operation result
type SyntheticBackupResult struct {
	BackupID         string            `json:"backup_id"`
	Success          bool              `json:"success"`
	BytesProcessed   int64             `json:"bytes_processed"`
	BytesOriginal    int64             `json:"bytes_original"`
	CompressionRatio float64           `json:"compression_ratio"`
	Duration         time.Duration     `json:"duration"`
	CreatedAt        time.Time         `json:"created_at"`
	CompletedAt      *time.Time        `json:"completed_at,omitempty"`
	Error            string            `json:"error,omitempty"`
	Metadata         map[string]string `json:"metadata"`
}

// SyntheticBackup represents a synthetic backup
type SyntheticBackup struct {
	ID               string                 `json:"id"`
	SourceRepo       string                 `json:"source_repo"`
	TargetRepo       string                 `json:"target_repo"`
	BackupType       string                 `json:"backup_type"`
	IncrementalSince *time.Time             `json:"incremental_since"`
	Compression      bool                   `json:"compression"`
	RetentionDays    int                    `json:"retention_days"`
	Size             int64                  `json:"size"`
	BytesOriginal    int64                  `json:"bytes_original"`
	BytesCompressed  int64                  `json:"bytes_compressed"`
	CompressionRatio float64                `json:"compression_ratio"`
	Status           BackupStatus           `json:"status"`
	CreatedAt        time.Time              `json:"created_at"`
	CompletedAt      *time.Time             `json:"completed_at,omitempty"`
	LastRunAt        *time.Time             `json:"last_run_at,omitempty"`
	TenantID         string                 `json:"tenant_id"`
	Settings         map[string]interface{} `json:"settings"`
	Metadata         map[string]string      `json:"metadata"`
	ChainID          string                 `json:"chain_id,omitempty"`
	Backups          []string               `json:"backups,omitempty"`
}

// BackupStatus represents synthetic backup status
type BackupStatus string

const (
	BackupStatusPending   BackupStatus = "pending"
	BackupStatusRunning   BackupStatus = "running"
	BackupStatusCompleted BackupStatus = "completed"
	BackupStatusFailed    BackupStatus = "failed"
	BackupStatusCancelled BackupStatus = "cancelled"
)

// SyntheticBackupFilter filters synthetic backups
type SyntheticBackupFilter struct {
	TenantID   string       `json:"tenant_id"`
	Status     BackupStatus `json:"status"`
	SourceRepo string       `json:"source_repo"`
	TargetRepo string       `json:"target_repo"`
	BackupType string       `json:"backup_type"`
	ChainID    string       `json:"chain_id"`
	Limit      int          `json:"limit"`
	Offset     int          `json:"offset"`
}

// MergeRequest represents an incremental merge request
type MergeRequest struct {
	ChainID          string                 `json:"chain_id"`
	IncrementalSince *time.Time             `json:"incremental_since"`
	TargetSize       int64                  `json:"target_size"`
	Compression      bool                   `json:"compression"`
	TenantID         string                 `json:"tenant_id"`
	Settings         map[string]interface{} `json:"settings"`
}

// MergeResult represents a merge operation result
type MergeResult struct {
	Success        bool              `json:"success"`
	MergedBackups  []string          `json:"merged_backups"`
	BackupID       string            `json:"backup_id"`
	BytesProcessed int64             `json:"bytes_processed"`
	BytesReduced   int64             `json:"bytes_reduced"`
	Duration       time.Duration     `json:"duration"`
	CreatedAt      time.Time         `json:"created_at"`
	CompletedAt    *time.Time        `json:"completed_at,omitempty"`
	Error          string            `json:"error,omitempty"`
	Metadata       map[string]string `json:"metadata"`
}

// ChainFilter filters backup chains
type ChainFilter struct {
	TenantID   string       `json:"tenant_id"`
	Status     BackupStatus `json:"status"`
	SourceRepo string       `json:"source_repo"`
	TargetRepo string       `json:"target_repo"`
	ChainID    string       `json:"chain_id"`
	Limit      int          `json:"limit"`
	Offset     int          `json:"offset"`
}

// BackupChain represents a backup chain
type BackupChain struct {
	ID         string       `json:"id"`
	SourceRepo string       `json:"source_repo"`
	TargetRepo string       `json:"target_repo"`
	BackupType string       `json:"backup_type"`
	Status     BackupStatus `json:"status"`
	Size       int64        `json:"size"`
	Backups    []string     `json:"backups"`
	CreatedAt  time.Time    `json:"created_at"`
	LastRunAt  *time.Time   `json:"last_run_at,omitempty"`
	TenantID   string       `json:"tenant_id"`
}

// SyntheticStats represents synthetic backup statistics
type SyntheticStats struct {
	TotalBackups            int64     `json:"total_backups"`
	TotalChains             int64     `json:"total_chains"`
	ActiveBackups           int       `json:"active_backups"`
	TotalSize               int64     `json:"total_size"`
	TotalOriginalSize       int64     `json:"total_original_size"`
	TotalCompressedSize     int64     `json:"total_compressed_size"`
	AverageCompressionRatio float64   `json:"average_compression_ratio"`
	LastActivity            time.Time `json:"last_activity"`
}

// ChainStats represents backup chain statistics
type ChainStats struct {
	TotalChains        int64     `json:"total_chains"`
	ActiveChains       int       `json:"active_chains"`
	TotalBackups       int64     `json:"total_backups"`
	AverageChainLength int64     `json:"average_chain_length"`
	LastActivity       time.Time `json:"last_activity"`
}

// InMemorySyntheticBackupManager implements SyntheticBackupManager in memory
type InMemorySyntheticBackupManager struct {
	backups   map[string]*SyntheticBackup
	chains    map[string]*BackupChain
	tenantMgr multitenancy.TenantManager
	dedupeMgr deduplication.DeduplicationManager
	mutex     sync.RWMutex
	stats     *SyntheticStats
}

// NewInMemorySyntheticBackupManager creates a new in-memory synthetic backup manager
func NewInMemorySyntheticBackupManager(tenantMgr multitenancy.TenantManager, dedupeMgr deduplication.DeduplicationManager) *InMemorySyntheticBackupManager {
	manager := &InMemorySyntheticBackupManager{
		backups:   make(map[string]*SyntheticBackup),
		chains:    make(map[string]*BackupChain),
		tenantMgr: tenantMgr,
		dedupeMgr: dedupeMgr,
		mutex:     sync.RWMutex{},
		stats:     &SyntheticStats{},
	}

	return manager
}

// CreateSyntheticBackup creates a new synthetic backup
func (m *InMemorySyntheticBackupManager) CreateSyntheticBackup(ctx context.Context, request *SyntheticBackupRequest) (*SyntheticBackupResult, error) {
	// Validate tenant context
	tenantID := m.tenantMgr.GetTenantFromContext(ctx)
	if tenantID == "" {
		return nil, fmt.Errorf("tenant ID required for synthetic backup creation")
	}

	// Parse size from metadata if provided
	var bytesOriginal, bytesCompressed int64
	if request.Metadata != nil {
		if val, ok := request.Metadata["bytes_original"]; ok {
			bytesOriginal = parseInt64FromString(val)
		}
		if val, ok := request.Metadata["bytes_compressed"]; ok {
			bytesCompressed = parseInt64FromString(val)
		}
	}

	// Create synthetic backup
	backup := &SyntheticBackup{
		ID:               fmt.Sprintf("synthetic-backup-%s", uuid.New().String()[:8]),
		SourceRepo:       request.SourceRepo,
		TargetRepo:       request.TargetRepo,
		BackupType:       request.BackupType,
		IncrementalSince: request.IncrementalSince,
		Compression:      request.Compression,
		RetentionDays:    request.RetentionDays,
		TenantID:         tenantID,
		Settings:         request.Settings,
		Metadata:         request.Metadata,
		Status:           BackupStatusPending,
		CreatedAt:        time.Now(),
		Size:             bytesOriginal,
		BytesOriginal:    bytesOriginal,
		BytesCompressed:  bytesCompressed,
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.backups[backup.ID] = backup

	// Update statistics
	m.stats.TotalBackups++
	m.stats.LastActivity = time.Now()

	return &SyntheticBackupResult{
		BackupID:         backup.ID,
		Success:          true,
		BytesProcessed:   bytesOriginal,
		BytesOriginal:    bytesOriginal,
		CompressionRatio: compressionRatio(bytesCompressed, bytesOriginal),
		Duration:         0,
		CreatedAt:        backup.CreatedAt,
		Metadata: map[string]string{
			"created_by": "synthetic-backup-manager",
		},
	}, nil
}

// GetSyntheticBackup retrieves a synthetic backup by ID
func (m *InMemorySyntheticBackupManager) GetSyntheticBackup(ctx context.Context, backupID string) (*SyntheticBackup, error) {
	// Validate tenant context
	tenantID := m.tenantMgr.GetTenantFromContext(ctx)
	if tenantID == "" {
		return nil, fmt.Errorf("tenant ID required for synthetic backup access")
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	backup, exists := m.backups[backupID]
	if !exists {
		return nil, fmt.Errorf("synthetic backup %s not found", backupID)
	}

	// Check tenant access
	if backup.TenantID != tenantID {
		return nil, fmt.Errorf("access denied: tenant %s cannot access synthetic backup %s", tenantID, backupID)
	}

	return backup, nil
}

// ListSyntheticBackups lists synthetic backups with filtering
func (m *InMemorySyntheticBackupManager) ListSyntheticBackups(ctx context.Context, filter *SyntheticBackupFilter) ([]SyntheticBackup, error) {
	// Validate tenant context
	tenantID := m.tenantMgr.GetTenantFromContext(ctx)
	if tenantID == "" {
		return nil, fmt.Errorf("tenant ID required for synthetic backup listing")
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var backups []SyntheticBackup
	for _, backup := range m.backups {
		// Always filter by tenant from context
		if backup.TenantID != tenantID {
			continue
		}

		// Apply additional filters if provided
		if filter != nil {
			if filter.TenantID != "" && backup.TenantID != filter.TenantID {
				continue
			}
			if filter.Status != "" && backup.Status != filter.Status {
				continue
			}
			if filter.SourceRepo != "" && backup.SourceRepo != filter.SourceRepo {
				continue
			}
			if filter.TargetRepo != "" && backup.TargetRepo != filter.TargetRepo {
				continue
			}
			if filter.BackupType != "" && backup.BackupType != filter.BackupType {
				continue
			}
			if filter.ChainID != "" && backup.ChainID != filter.ChainID {
				continue
			}
		}

		backups = append(backups, *backup)
	}

	// Apply pagination
	if filter != nil && filter.Offset > 0 && filter.Offset < len(backups) {
		end := filter.Offset + filter.Limit
		if end > len(backups) {
			end = len(backups)
		}
		backups = backups[filter.Offset:end]
	}

	return backups, nil
}

// DeleteSyntheticBackup deletes a synthetic backup
func (m *InMemorySyntheticBackupManager) DeleteSyntheticBackup(ctx context.Context, backupID string) error {
	// Validate tenant context
	tenantID := m.tenantMgr.GetTenantFromContext(ctx)
	if tenantID == "" {
		return fmt.Errorf("tenant ID required for synthetic backup deletion")
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	backup, exists := m.backups[backupID]
	if !exists {
		return fmt.Errorf("synthetic backup %s not found", backupID)
	}

	// Check tenant access
	if backup.TenantID != tenantID {
		return fmt.Errorf("access denied: tenant %s cannot delete synthetic backup %s", tenantID, backupID)
	}

	delete(m.backups, backupID)

	// Update statistics
	m.stats.TotalBackups--
	m.stats.LastActivity = time.Now()

	return nil
}

// MergeIncrementals merges incremental backups into a synthetic backup
func (m *InMemorySyntheticBackupManager) MergeIncrementals(ctx context.Context, request *MergeRequest) (*MergeResult, error) {
	// Validate tenant context
	tenantID := m.tenantMgr.GetTenantFromContext(ctx)
	if tenantID == "" {
		return nil, fmt.Errorf("tenant ID required for incremental merge")
	}

	// Get or create chain
	chain, exists := m.chains[request.ChainID]
	if !exists {
		chain = &BackupChain{
			ID:         request.ChainID,
			SourceRepo: "synthetic-source",
			TargetRepo: "synthetic-target",
			BackupType: "synthetic",
			Status:     BackupStatusPending,
			CreatedAt:  time.Now(),
			Backups:    []string{},
			TenantID:   tenantID,
		}
		m.chains[request.ChainID] = chain
	}

	// Find incremental backups to merge
	var backupsToMerge []string
	for _, backup := range m.backups {
		if backup.TenantID == tenantID &&
			backup.BackupType == "incremental" &&
			(request.IncrementalSince == nil || backup.CreatedAt.After(*request.IncrementalSince)) {
			backupsToMerge = append(backupsToMerge, backup.ID)
		}
	}

	if len(backupsToMerge) == 0 {
		return nil, fmt.Errorf("no incremental backups found to merge")
	}

	// Create synthetic backup
	syntheticBackup := &SyntheticBackup{
		ID:            fmt.Sprintf("synthetic-backup-%d", time.Now().UnixNano()),
		SourceRepo:    "synthetic-source",
		TargetRepo:    "synthetic-target",
		BackupType:    "synthetic",
		Compression:   request.Compression,
		RetentionDays: 30, // Default retention
		TenantID:      tenantID,
		Status:        BackupStatusPending,
		CreatedAt:     time.Now(),
		ChainID:       request.ChainID,
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Process merge
	var totalOriginalSize int64
	var totalCompressedSize int64
	var mergedBackupIDs []string

	for _, backupID := range backupsToMerge {
		backup := m.backups[backupID]
		if backup != nil {
			totalOriginalSize += backup.BytesOriginal
			totalCompressedSize += backup.BytesCompressed
			mergedBackupIDs = append(mergedBackupIDs, backupID)

			// Update backup status to merged
			backup.Status = BackupStatusCompleted
			completedAt := time.Now()
			backup.CompletedAt = &completedAt
		}
	}

	syntheticBackup.Backups = mergedBackupIDs
	syntheticBackup.Size = totalOriginalSize
	syntheticBackup.BytesOriginal = totalOriginalSize
	syntheticBackup.BytesCompressed = totalCompressedSize

	if totalOriginalSize > 0 {
		syntheticBackup.CompressionRatio = float64(totalCompressedSize) / float64(totalOriginalSize)
	}

	m.backups[syntheticBackup.ID] = syntheticBackup

	// Update chain
	chain.Backups = mergedBackupIDs
	chain.Status = BackupStatusCompleted
	completedAt := time.Now()
	chain.LastRunAt = &completedAt
	m.chains[request.ChainID] = chain

	// Update statistics
	m.stats.TotalBackups++
	m.stats.TotalChains++
	m.stats.TotalSize += totalOriginalSize
	m.stats.TotalOriginalSize += totalOriginalSize
	m.stats.TotalCompressedSize += totalCompressedSize
	m.stats.LastActivity = time.Now()

	return &MergeResult{
		Success:        true,
		MergedBackups:  mergedBackupIDs,
		BackupID:       syntheticBackup.ID,
		BytesProcessed: totalOriginalSize,
		BytesReduced:   totalOriginalSize - totalCompressedSize,
		Duration:       0, // Would be calculated in real implementation
		CreatedAt:      syntheticBackup.CreatedAt,
		CompletedAt:    &syntheticBackup.CreatedAt,
		Metadata: map[string]string{
			"merged_backups": fmt.Sprintf("%d", len(backupsToMerge)),
		},
	}, nil
}

// GetBackupChain retrieves a backup chain by ID
func (m *InMemorySyntheticBackupManager) GetBackupChain(ctx context.Context, filter *ChainFilter) (*BackupChain, error) {
	// Validate tenant context
	tenantID := m.tenantMgr.GetTenantFromContext(ctx)
	if tenantID == "" {
		return nil, fmt.Errorf("tenant ID required for backup chain access")
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	chain, exists := m.chains[filter.ChainID]
	if !exists {
		return nil, fmt.Errorf("backup chain %s not found", filter.ChainID)
	}

	// Check tenant access
	if chain.TenantID != tenantID {
		return nil, fmt.Errorf("access denied: tenant %s cannot access backup chain %s", tenantID, filter.ChainID)
	}

	return chain, nil
}

// GetSyntheticStats retrieves synthetic backup statistics
func (m *InMemorySyntheticBackupManager) GetSyntheticStats(ctx context.Context) (*SyntheticStats, error) {
	// Validate tenant context
	tenantID := m.tenantMgr.GetTenantFromContext(ctx)
	if tenantID == "" {
		return nil, fmt.Errorf("tenant ID required for synthetic statistics retrieval")
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Calculate statistics for tenant
	var totalBackups int64
	var totalChains int64
	var activeBackups int
	var totalSize int64
	var totalOriginalSize int64
	var totalCompressedSize int64
	var totalChainLength int64

	for _, backup := range m.backups {
		if backup.TenantID == tenantID {
			totalBackups++
			totalSize += backup.Size
			totalOriginalSize += backup.BytesOriginal
			totalCompressedSize += backup.BytesCompressed

			if backup.Status == BackupStatusPending || backup.Status == BackupStatusRunning {
				activeBackups++
			}
		}
	}

	for _, chain := range m.chains {
		if chain.TenantID == tenantID {
			totalChains++
			totalChainLength += int64(len(chain.Backups))
		}
	}

	// Calculate average compression ratio
	var averageCompressionRatio float64
	if totalOriginalSize > 0 {
		averageCompressionRatio = float64(totalCompressedSize) / float64(totalOriginalSize)
	}

	// Update stats
	m.stats.TotalBackups = totalBackups
	m.stats.TotalChains = totalChains
	m.stats.ActiveBackups = activeBackups
	m.stats.TotalSize = totalSize
	m.stats.TotalOriginalSize = totalOriginalSize
	m.stats.TotalCompressedSize = totalCompressedSize
	m.stats.AverageCompressionRatio = averageCompressionRatio
	m.stats.LastActivity = time.Now()

	return &SyntheticStats{
		TotalBackups:            totalBackups,
		TotalChains:             totalChains,
		ActiveBackups:           activeBackups,
		TotalSize:               totalSize,
		TotalOriginalSize:       totalOriginalSize,
		TotalCompressedSize:     totalCompressedSize,
		AverageCompressionRatio: averageCompressionRatio,
		LastActivity:            time.Now(),
	}, nil
}

// GetBackupChainStats retrieves backup chain statistics
func (m *InMemorySyntheticBackupManager) GetBackupChainStats(ctx context.Context) (*ChainStats, error) {
	// Validate tenant context
	tenantID := m.tenantMgr.GetTenantFromContext(ctx)
	if tenantID == "" {
		return nil, fmt.Errorf("tenant ID required for backup chain statistics retrieval")
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Calculate statistics for tenant
	var totalChains int64
	var activeChains int
	var totalBackups int64
	var averageChainLength int64

	for _, chain := range m.chains {
		if chain.TenantID == tenantID {
			totalChains++
			totalBackups += int64(len(chain.Backups))

			if chain.Status == BackupStatusPending || chain.Status == BackupStatusRunning {
				activeChains++
			}

			averageChainLength += int64(len(chain.Backups))
		}
	}

	// Update stats
	m.stats.LastActivity = time.Now()

	return &ChainStats{
		TotalChains:        totalChains,
		ActiveChains:       activeChains,
		TotalBackups:       totalBackups,
		AverageChainLength: averageChainLength,
		LastActivity:       time.Now(),
	}, nil
}

// Helper functions
func parseInt64FromString(s string) int64 {
	var result int64
	fmt.Sscanf(s, "%d", &result)
	return result
}

func compressionRatio(compressed, original int64) float64 {
	if original == 0 {
		return 0.0
	}
	return float64(compressed) / float64(original)
}
