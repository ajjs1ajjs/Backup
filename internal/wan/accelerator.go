package wan

import (
	"context"
	"fmt"
	"sync"
	"time"

	"novabackup/internal/deduplication"
	"novabackup/internal/multitenancy"
)

// WANAccelerator manages WAN optimization for backup traffic
type WANAccelerator interface {
	// Traffic management
	OptimizeTransfer(ctx context.Context, transfer *TransferRequest) (*TransferResult, error)
	GetOptimizationStats(ctx context.Context) (*OptimizationStats, error)

	// Caching
	CacheData(ctx context.Context, cacheKey string, data []byte, ttl time.Duration) error
	GetCachedData(ctx context.Context, cacheKey string) ([]byte, error)
	InvalidateCache(ctx context.Context, cacheKey string) error

	// Traffic shaping
	ApplyTrafficShaping(ctx context.Context, config *TrafficShapingConfig) error
	GetCurrentBandwidth(ctx context.Context) (*BandwidthInfo, error)

	// Compression
	CompressData(ctx context.Context, data []byte, algorithm CompressionAlgorithm) (*CompressedData, error)
	DecompressData(ctx context.Context, compressed *CompressedData) ([]byte, error)

	// Configuration
	UpdateConfiguration(ctx context.Context, config *WANConfig) error
	GetConfiguration(ctx context.Context) (*WANConfig, error)
}

// TransferRequest represents a data transfer request
type TransferRequest struct {
	ID          string            `json:"id"`
	Source      string            `json:"source"`
	Destination string            `json:"destination"`
	Data        []byte            `json:"data"`
	Size        int64             `json:"size"`
	Priority    Priority          `json:"priority"`
	TenantID    string            `json:"tenant_id"`
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   time.Time         `json:"created_at"`
}

// TransferResult represents the result of an optimized transfer
type TransferResult struct {
	RequestID        string            `json:"request_id"`
	OriginalSize     int64             `json:"original_size"`
	CompressedSize   int64             `json:"compressed_size"`
	TransferTime     time.Duration     `json:"transfer_time"`
	BandwidthUsed    int64             `json:"bandwidth_used"`
	CacheHit         bool              `json:"cache_hit"`
	CompressionRatio float64           `json:"compression_ratio"`
	Optimizations    []string          `json:"optimizations"`
	Metadata         map[string]string `json:"metadata"`
	CompletedAt      time.Time         `json:"completed_at"`
}

// OptimizationStats contains WAN optimization statistics
type OptimizationStats struct {
	TotalTransfers       int64                    `json:"total_transfers"`
	TotalDataTransferred int64                    `json:"total_data_transferred"`
	TotalBandwidthSaved  int64                    `json:"total_bandwidth_saved"`
	AverageCompression   float64                  `json:"average_compression"`
	CacheHitRate         float64                  `json:"cache_hit_rate"`
	ActiveConnections    int                      `json:"active_connections"`
	QueueLength          int                      `json:"queue_length"`
	BandwidthUtilization float64                  `json:"bandwidth_utilization"`
	TransferTimes        map[string]time.Duration `json:"transfer_times"`
	LastUpdated          time.Time                `json:"last_updated"`
}

// TrafficShapingConfig defines traffic shaping parameters
type TrafficShapingConfig struct {
	MaxBandwidth      int64      `json:"max_bandwidth"` // bytes per second
	BurstSize         int64      `json:"burst_size"`    // bytes
	PriorityLevels    []Priority `json:"priority_levels"`
	ThrottlingEnabled bool       `json:"throttling_enabled"`
	QualityOfService  QoSProfile `json:"quality_of_service"`
}

// BandwidthInfo contains current bandwidth information
type BandwidthInfo struct {
	CurrentBandwidth   int64         `json:"current_bandwidth"`
	AvailableBandwidth int64         `json:"available_bandwidth"`
	Utilization        float64       `json:"utilization"`
	QueueDepth         int           `json:"queue_depth"`
	AverageLatency     time.Duration `json:"average_latency"`
	PacketLoss         float64       `json:"packet_loss"`
}

// CompressionAlgorithm represents different compression algorithms
type CompressionAlgorithm string

const (
	CompressionNone   CompressionAlgorithm = "none"
	CompressionGzip   CompressionAlgorithm = "gzip"
	CompressionLZ4    CompressionAlgorithm = "lz4"
	CompressionZSTD   CompressionAlgorithm = "zstd"
	CompressionSnappy CompressionAlgorithm = "snappy"
)

// CompressedData represents compressed data with metadata
type CompressedData struct {
	Algorithm      CompressionAlgorithm `json:"algorithm"`
	OriginalSize   int64                `json:"original_size"`
	CompressedSize int64                `json:"compressed_size"`
	Data           []byte               `json:"data"`
	Checksum       string               `json:"checksum"`
	CompressedAt   time.Time            `json:"compressed_at"`
}

// Priority defines transfer priority levels
type Priority string

const (
	PriorityLow      Priority = "low"
	PriorityNormal   Priority = "normal"
	PriorityHigh     Priority = "high"
	PriorityCritical Priority = "critical"
)

// QoSProfile defines quality of service profiles
type QoSProfile struct {
	Name         string        `json:"name"`
	MinBandwidth int64         `json:"min_bandwidth"`
	MaxBandwidth int64         `json:"max_bandwidth"`
	MaxLatency   time.Duration `json:"max_latency"`
	PacketLoss   float64       `json:"packet_loss"`
	Jitter       time.Duration `json:"jitter"`
}

// WANConfig contains WAN accelerator configuration
type WANConfig struct {
	Enabled               bool                              `json:"enabled"`
	CacheEnabled          bool                              `json:"cache_enabled"`
	CompressionEnabled    bool                              `json:"compression_enabled"`
	TrafficShapingEnabled bool                              `json:"traffic_shaping_enabled"`
	DefaultAlgorithm      CompressionAlgorithm              `json:"default_algorithm"`
	CacheSize             int64                             `json:"cache_size"`
	CacheTTL              time.Duration                     `json:"cache_ttl"`
	MaxConnections        int                               `json:"max_connections"`
	TrafficShaping        *TrafficShapingConfig             `json:"traffic_shaping"`
	CompressionProfiles   map[Priority]CompressionAlgorithm `json:"compression_profiles"`
	QoSProfiles           map[string]QoSProfile             `json:"qos_profiles"`
}

// InMemoryWANAccelerator implements WANAccelerator in memory
type InMemoryWANAccelerator struct {
	config           WANConfig
	tenantManager    multitenancy.TenantManager
	dedupeManager    deduplication.DeduplicationManager
	cache            map[string]*CacheEntry
	stats            *OptimizationStats
	mutex            sync.RWMutex
	activeTransfers  map[string]*ActiveTransfer
	bandwidthLimiter *BandwidthLimiter
	compressor       *Compressor
	stopChan         chan struct{}
}

// CacheEntry represents a cache entry
type CacheEntry struct {
	Key         string    `json:"key"`
	Data        []byte    `json:"data"`
	Size        int64     `json:"size"`
	ExpiresAt   time.Time `json:"expires_at"`
	AccessCount int64     `json:"access_count"`
	LastAccess  time.Time `json:"last_access"`
	CreatedAt   time.Time `json:"created_at"`
}

// ActiveTransfer represents an active transfer
type ActiveTransfer struct {
	ID        string        `json:"id"`
	StartTime time.Time     `json:"start_time"`
	Size      int64         `json:"size"`
	Priority  Priority      `json:"priority"`
	Bandwidth int64         `json:"bandwidth"`
	Progress  float64       `json:"progress"`
	Estimated time.Duration `json:"estimated_time"`
}

// BandwidthLimiter manages bandwidth allocation
type BandwidthLimiter struct {
	config    *TrafficShapingConfig
	current   int64
	available int64
	mutex     sync.Mutex
	ticker    *time.Ticker
	stopChan  chan struct{}
}

// Compressor handles data compression
type Compressor struct {
	algorithms map[CompressionAlgorithm]CompressionFunc
	config     map[Priority]CompressionAlgorithm
}

// CompressionFunc represents a compression function
type CompressionFunc func(data []byte) ([]byte, error)

// NewInMemoryWANAccelerator creates a new in-memory WAN accelerator
func NewInMemoryWANAccelerator(config WANConfig, tenantMgr multitenancy.TenantManager, dedupeMgr deduplication.DeduplicationManager) *InMemoryWANAccelerator {
	accelerator := &InMemoryWANAccelerator{
		config:           config,
		tenantManager:    tenantMgr,
		dedupeManager:    dedupeMgr,
		cache:            make(map[string]*CacheEntry),
		stats:            &OptimizationStats{},
		activeTransfers:  make(map[string]*ActiveTransfer),
		bandwidthLimiter: NewBandwidthLimiter(config.TrafficShaping),
		compressor:       NewCompressor(config.CompressionProfiles),
		stopChan:         make(chan struct{}),
	}

	// Start background tasks
	go accelerator.startBackgroundTasks()

	return accelerator
}

// OptimizeTransfer optimizes a data transfer
func (w *InMemoryWANAccelerator) OptimizeTransfer(ctx context.Context, transfer *TransferRequest) (*TransferResult, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// Validate tenant context
	tenantID := w.tenantManager.GetTenantFromContext(ctx)
	if tenantID == "" {
		return nil, fmt.Errorf("tenant ID required for WAN operations")
	}

	transfer.TenantID = tenantID
	startTime := time.Now()

	// Check cache first
	if w.config.CacheEnabled {
		cacheKey := w.generateCacheKey(transfer)
		if cached, err := w.GetCachedData(ctx, cacheKey); err == nil && cached != nil {
			w.stats.CacheHitRate = float64(w.stats.TotalTransfers+1) / float64(w.stats.TotalTransfers+2)
			return &TransferResult{
				RequestID:        transfer.ID,
				OriginalSize:     transfer.Size,
				CompressedSize:   int64(len(cached)),
				TransferTime:     time.Since(startTime),
				CacheHit:         true,
				CompressionRatio: float64(transfer.Size) / float64(len(cached)),
				Optimizations:    []string{"cache_hit"},
				CompletedAt:      time.Now(),
			}, nil
		}
	}

	// Apply compression if enabled
	var compressedData []byte
	var compressionRatio float64
	var optimizations []string

	if w.config.CompressionEnabled {
		algorithm := w.getCompressionAlgorithm(transfer.Priority)
		compressed, err := w.CompressData(ctx, transfer.Data, algorithm)
		if err != nil {
			return nil, fmt.Errorf("compression failed: %w", err)
		}
		compressedData = compressed.Data
		compressionRatio = float64(transfer.Size) / float64(compressed.CompressedSize)
		optimizations = append(optimizations, fmt.Sprintf("compression_%s", algorithm))
	} else {
		compressedData = transfer.Data
		compressionRatio = 1.0
	}

	// Apply traffic shaping
	if w.config.TrafficShapingEnabled {
		err := w.bandwidthLimiter.ReserveBandwidth(int64(len(compressedData)), transfer.Priority)
		if err != nil {
			return nil, fmt.Errorf("bandwidth reservation failed: %w", err)
		}
		defer w.bandwidthLimiter.ReleaseBandwidth(int64(len(compressedData)))
		optimizations = append(optimizations, "traffic_shaping")
	}

	// Simulate transfer time
	transferTime := w.simulateTransferTime(len(compressedData))

	// Update statistics
	w.stats.TotalTransfers++
	w.stats.TotalDataTransferred += transfer.Size
	w.stats.TotalBandwidthSaved += transfer.Size - int64(len(compressedData))
	w.stats.AverageCompression = (w.stats.AverageCompression*float64(w.stats.TotalTransfers-1) + compressionRatio) / float64(w.stats.TotalTransfers)

	// Cache the result
	if w.config.CacheEnabled {
		cacheKey := w.generateCacheKey(transfer)
		w.CacheData(ctx, cacheKey, compressedData, w.config.CacheTTL)
	}

	return &TransferResult{
		RequestID:        transfer.ID,
		OriginalSize:     transfer.Size,
		CompressedSize:   int64(len(compressedData)),
		TransferTime:     transferTime,
		BandwidthUsed:    int64(len(compressedData)),
		CacheHit:         false,
		CompressionRatio: compressionRatio,
		Optimizations:    optimizations,
		CompletedAt:      time.Now(),
	}, nil
}

// CacheData caches data with TTL
func (w *InMemoryWANAccelerator) CacheData(ctx context.Context, cacheKey string, data []byte, ttl time.Duration) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// Check cache size limit
	if int64(len(w.cache)) >= w.config.CacheSize {
		w.evictOldestCacheEntry()
	}

	entry := &CacheEntry{
		Key:         cacheKey,
		Data:        data,
		Size:        int64(len(data)),
		ExpiresAt:   time.Now().Add(ttl),
		AccessCount: 0,
		LastAccess:  time.Now(),
		CreatedAt:   time.Now(),
	}

	w.cache[cacheKey] = entry
	return nil
}

// GetCachedData retrieves cached data
func (w *InMemoryWANAccelerator) GetCachedData(ctx context.Context, cacheKey string) ([]byte, error) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	entry, exists := w.cache[cacheKey]
	if !exists {
		return nil, fmt.Errorf("cache miss")
	}

	// Check expiration
	if time.Now().After(entry.ExpiresAt) {
		delete(w.cache, cacheKey)
		return nil, fmt.Errorf("cache expired")
	}

	// Update access statistics
	entry.AccessCount++
	entry.LastAccess = time.Now()

	return entry.Data, nil
}

// InvalidateCache removes data from cache
func (w *InMemoryWANAccelerator) InvalidateCache(ctx context.Context, cacheKey string) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	delete(w.cache, cacheKey)
	return nil
}

// ApplyTrafficShaping applies traffic shaping configuration
func (w *InMemoryWANAccelerator) ApplyTrafficShaping(ctx context.Context, config *TrafficShapingConfig) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.config.TrafficShaping = config
	w.bandwidthLimiter.UpdateConfig(config)
	return nil
}

// GetCurrentBandwidth returns current bandwidth information
func (w *InMemoryWANAccelerator) GetCurrentBandwidth(ctx context.Context) (*BandwidthInfo, error) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	return &BandwidthInfo{
		CurrentBandwidth:   w.bandwidthLimiter.current,
		AvailableBandwidth: w.bandwidthLimiter.available,
		Utilization:        float64(w.bandwidthLimiter.current) / float64(w.bandwidthLimiter.config.MaxBandwidth),
		QueueDepth:         len(w.activeTransfers),
		AverageLatency:     time.Millisecond * 50, // Simulated
		PacketLoss:         0.01,                  // Simulated
	}, nil
}

// CompressData compresses data using specified algorithm
func (w *InMemoryWANAccelerator) CompressData(ctx context.Context, data []byte, algorithm CompressionAlgorithm) (*CompressedData, error) {
	return w.compressor.Compress(data, algorithm)
}

// DecompressData decompresses data
func (w *InMemoryWANAccelerator) DecompressData(ctx context.Context, compressed *CompressedData) ([]byte, error) {
	return w.compressor.Decompress(compressed)
}

// UpdateConfiguration updates WAN accelerator configuration
func (w *InMemoryWANAccelerator) UpdateConfiguration(ctx context.Context, config *WANConfig) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.config = *config
	if config.TrafficShaping != nil {
		w.bandwidthLimiter.UpdateConfig(config.TrafficShaping)
	}
	w.compressor.UpdateProfiles(config.CompressionProfiles)
	return nil
}

// GetConfiguration returns current configuration
func (w *InMemoryWANAccelerator) GetConfiguration(ctx context.Context) (*WANConfig, error) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	return &w.config, nil
}

// GetOptimizationStats returns optimization statistics
func (w *InMemoryWANAccelerator) GetOptimizationStats(ctx context.Context) (*OptimizationStats, error) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	w.stats.LastUpdated = time.Now()
	w.stats.ActiveConnections = len(w.activeTransfers)
	w.stats.QueueLength = len(w.activeTransfers)

	if w.config.TrafficShaping != nil {
		w.stats.BandwidthUtilization = float64(w.bandwidthLimiter.current) / float64(w.config.TrafficShaping.MaxBandwidth)
	}

	return w.stats, nil
}

// Utility methods
func (w *InMemoryWANAccelerator) generateCacheKey(transfer *TransferRequest) string {
	return fmt.Sprintf("%s:%s:%s", transfer.Source, transfer.Destination, transfer.ID)
}

func (w *InMemoryWANAccelerator) getCompressionAlgorithm(priority Priority) CompressionAlgorithm {
	if algorithm, exists := w.config.CompressionProfiles[priority]; exists {
		return algorithm
	}
	return w.config.DefaultAlgorithm
}

func (w *InMemoryWANAccelerator) simulateTransferTime(dataSize int) time.Duration {
	// Simulate transfer time based on data size and bandwidth
	bandwidth := w.bandwidthLimiter.available
	if bandwidth <= 0 {
		bandwidth = 1000000 // 1MB/s default
	}

	seconds := float64(dataSize) / float64(bandwidth)
	return time.Duration(seconds * float64(time.Second))
}

func (w *InMemoryWANAccelerator) evictOldestCacheEntry() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range w.cache {
		if oldestKey == "" || entry.LastAccess.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.LastAccess
		}
	}

	if oldestKey != "" {
		delete(w.cache, oldestKey)
	}
}

func (w *InMemoryWANAccelerator) startBackgroundTasks() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.cleanupExpiredCache()
		case <-w.stopChan:
			return
		}
	}
}

func (w *InMemoryWANAccelerator) cleanupExpiredCache() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	now := time.Now()
	for key, entry := range w.cache {
		if now.After(entry.ExpiresAt) {
			delete(w.cache, key)
		}
	}
}

// NewWANConfig creates a default WAN configuration
func NewWANConfig() WANConfig {
	return WANConfig{
		Enabled:               true,
		CacheEnabled:          true,
		CompressionEnabled:    true,
		TrafficShapingEnabled: true,
		DefaultAlgorithm:      CompressionLZ4,
		CacheSize:             1000, // 1000 entries
		CacheTTL:              1 * time.Hour,
		MaxConnections:        100,
		TrafficShaping: &TrafficShapingConfig{
			MaxBandwidth:      10000000, // 10MB/s
			BurstSize:         1000000,  // 1MB
			PriorityLevels:    []Priority{PriorityCritical, PriorityHigh, PriorityNormal, PriorityLow},
			ThrottlingEnabled: true,
			QualityOfService: QoSProfile{
				Name:         "default",
				MinBandwidth: 1000000,  // 1MB/s
				MaxBandwidth: 10000000, // 10MB/s
				MaxLatency:   100 * time.Millisecond,
				PacketLoss:   0.01,
				Jitter:       10 * time.Millisecond,
			},
		},
		CompressionProfiles: map[Priority]CompressionAlgorithm{
			PriorityCritical: CompressionLZ4,
			PriorityHigh:     CompressionLZ4,
			PriorityNormal:   CompressionGzip,
			PriorityLow:      CompressionGzip,
		},
		QoSProfiles: map[string]QoSProfile{
			"realtime": {
				Name:         "realtime",
				MinBandwidth: 5000000,  // 5MB/s
				MaxBandwidth: 10000000, // 10MB/s
				MaxLatency:   50 * time.Millisecond,
				PacketLoss:   0.001,
				Jitter:       5 * time.Millisecond,
			},
			"bulk": {
				Name:         "bulk",
				MinBandwidth: 1000000, // 1MB/s
				MaxBandwidth: 5000000, // 5MB/s
				MaxLatency:   500 * time.Millisecond,
				PacketLoss:   0.05,
				Jitter:       50 * time.Millisecond,
			},
		},
	}
}
