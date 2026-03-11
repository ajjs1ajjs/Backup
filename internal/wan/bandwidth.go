package wan

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"time"
)

// NewBandwidthLimiter creates a new bandwidth limiter
func NewBandwidthLimiter(config *TrafficShapingConfig) *BandwidthLimiter {
	if config == nil {
		config = &TrafficShapingConfig{
			MaxBandwidth: 10000000, // 10MB/s default
		}
	}

	limiter := &BandwidthLimiter{
		config:    config,
		current:   0,
		available: config.MaxBandwidth,
		stopChan:  make(chan struct{}),
	}

	// Start bandwidth monitoring
	go limiter.startMonitoring()

	return limiter
}

// ReserveBandwidth reserves bandwidth for a transfer
func (bl *BandwidthLimiter) ReserveBandwidth(bytes int64, priority Priority) error {
	bl.mutex.Lock()
	defer bl.mutex.Unlock()

	if bl.available < bytes {
		return ErrInsufficientBandwidth
	}

	bl.available -= bytes
	bl.current += bytes

	return nil
}

// ReleaseBandwidth releases reserved bandwidth
func (bl *BandwidthLimiter) ReleaseBandwidth(bytes int64) {
	bl.mutex.Lock()
	defer bl.mutex.Unlock()

	bl.available += bytes
	bl.current -= bytes

	if bl.current < 0 {
		bl.current = 0
	}
}

// UpdateConfig updates the bandwidth limiter configuration
func (bl *BandwidthLimiter) UpdateConfig(config *TrafficShapingConfig) {
	bl.mutex.Lock()
	defer bl.mutex.Unlock()

	bl.config = config
	bl.available = config.MaxBandwidth - bl.current
}

// startMonitoring starts bandwidth monitoring
func (bl *BandwidthLimiter) startMonitoring() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			bl.resetAvailableBandwidth()
		case <-bl.stopChan:
			return
		}
	}
}

// resetAvailableBandwidth resets available bandwidth based on config
func (bl *BandwidthLimiter) resetAvailableBandwidth() {
	bl.mutex.Lock()
	defer bl.mutex.Unlock()

	bl.available = bl.config.MaxBandwidth - bl.current
}

// Stop stops the bandwidth limiter
func (bl *BandwidthLimiter) Stop() {
	close(bl.stopChan)
}

// NewCompressor creates a new compressor
func NewCompressor(profiles map[Priority]CompressionAlgorithm) *Compressor {
	compressor := &Compressor{
		algorithms: make(map[CompressionAlgorithm]CompressionFunc),
		config:     make(map[Priority]CompressionAlgorithm),
	}

	// Register compression algorithms
	compressor.registerAlgorithms()

	// Set default profiles
	if profiles != nil {
		compressor.config = profiles
	} else {
		compressor.config = map[Priority]CompressionAlgorithm{
			PriorityCritical: CompressionLZ4,
			PriorityHigh:     CompressionLZ4,
			PriorityNormal:   CompressionGzip,
			PriorityLow:      CompressionGzip,
		}
	}

	return compressor
}

// Compress compresses data using the specified algorithm
func (c *Compressor) Compress(data []byte, algorithm CompressionAlgorithm) (*CompressedData, error) {
	if algorithm == CompressionNone {
		return &CompressedData{
			Algorithm:      CompressionNone,
			OriginalSize:   int64(len(data)),
			CompressedSize: int64(len(data)),
			Data:           data,
			Checksum:       calculateChecksum(data),
			CompressedAt:   time.Now(),
		}, nil
	}

	compressFunc, exists := c.algorithms[algorithm]
	if !exists {
		return nil, ErrUnsupportedAlgorithm
	}

	compressed, err := compressFunc(data)
	if err != nil {
		return nil, err
	}

	return &CompressedData{
		Algorithm:      algorithm,
		OriginalSize:   int64(len(data)),
		CompressedSize: int64(len(compressed)),
		Data:           compressed,
		Checksum:       calculateChecksum(compressed),
		CompressedAt:   time.Now(),
	}, nil
}

// Decompress decompresses data
func (c *Compressor) Decompress(compressed *CompressedData) ([]byte, error) {
	if compressed.Algorithm == CompressionNone {
		return compressed.Data, nil
	}

	switch compressed.Algorithm {
	case CompressionGzip:
		return c.decompressGzip(compressed.Data)
	case CompressionLZ4:
		return c.decompressLZ4(compressed.Data)
	case CompressionSnappy:
		return c.decompressSnappy(compressed.Data)
	default:
		return nil, ErrUnsupportedAlgorithm
	}
}

// UpdateProfiles updates compression profiles
func (c *Compressor) UpdateProfiles(profiles map[Priority]CompressionAlgorithm) {
	c.config = profiles
}

// registerAlgorithms registers compression algorithms
func (c *Compressor) registerAlgorithms() {
	c.algorithms[CompressionGzip] = c.compressGzip
	// Note: LZ4 and Snappy would require external dependencies
	// For now, we'll implement simple compression only
}

// Compression implementations
func (c *Compressor) compressGzip(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	if _, err := gz.Write(data); err != nil {
		return nil, err
	}

	if err := gz.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Placeholder implementations for other algorithms
func (c *Compressor) compressLZ4(data []byte) ([]byte, error) {
	// Simple placeholder - in real implementation would use LZ4
	return c.compressGzip(data)
}

func (c *Compressor) compressSnappy(data []byte) ([]byte, error) {
	// Simple placeholder - in real implementation would use Snappy
	return c.compressGzip(data)
}

// Decompression implementations
func (c *Compressor) decompressGzip(data []byte) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	return io.ReadAll(gz)
}

func (c *Compressor) decompressLZ4(data []byte) ([]byte, error) {
	// Simple placeholder - in real implementation would use LZ4
	return c.decompressGzip(data)
}

func (c *Compressor) decompressSnappy(data []byte) ([]byte, error) {
	// Simple placeholder - in real implementation would use Snappy
	return c.decompressGzip(data)
}

// calculateChecksum calculates a simple checksum
func calculateChecksum(data []byte) string {
	var sum uint32
	for _, b := range data {
		sum += uint32(b)
	}
	return fmt.Sprintf("%08x", sum)
}

// Error definitions
var (
	ErrInsufficientBandwidth = fmt.Errorf("insufficient bandwidth available")
	ErrUnsupportedAlgorithm  = fmt.Errorf("unsupported compression algorithm")
)
