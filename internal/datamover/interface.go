package datamover

import (
	"context"
	"io"
)

// DataMover defines the interface for distributed data transport
type DataMover interface {
	// ReadDisk reads blocks from a source (Hyper-V/VMware)
	ReadDisk(ctx context.Context, sourceURI string, offset int64, size int64) (io.ReadCloser, error)
	
	// WriteChunk writes a deduplicated chunk to storage
	WriteChunk(ctx context.Context, chunkID string, data io.Reader) error
	
	// GetSystemInfo returns hardware and OS info for optimization
	GetSystemInfo(ctx context.Context) (*SystemInfo, error)
}

type SystemInfo struct {
	Hostname string
	OS       string
	Arch     string
	CPUCount int
	Memory   int64 // bytes
	Network  []NetworkInfo
}

type NetworkInfo struct {
	Interface string
	IP        string
	Speed     int64 // Mbps
}
