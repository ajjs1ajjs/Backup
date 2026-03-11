package datamover

import (
	"context"
	"fmt"
	"io"
	"os"
)

// DiskReader implements the DataMover interface for local disk/file access
type DiskReader struct {
}

func NewDiskReader() *DiskReader {
	return &DiskReader{}
}

func (r *DiskReader) ReadDisk(ctx context.Context, sourceURI string, offset int64, size int64) (io.ReadCloser, error) {
	file, err := os.Open(sourceURI)
	if err != nil {
		return nil, fmt.Errorf("failed to open source: %w", err)
	}

	_, err = file.Seek(offset, io.SeekStart)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to seek to offset %d: %w", offset, err)
	}

	// Use a LimitedReader to ensure we only read the requested size
	return &closeableLimitedReader{
		R:    io.LimitReader(file, size),
		File: file,
	}, nil
}

func (r *DiskReader) WriteChunk(ctx context.Context, chunkID string, data io.Reader) error {
	return fmt.Errorf("DiskReader does not support WriteChunk")
}

func (r *DiskReader) GetSystemInfo(ctx context.Context) (*SystemInfo, error) {
	hostname, _ := os.Hostname()
	return &SystemInfo{
		Hostname: hostname,
		OS:       os.Getenv("OS"), // Simplistic for now
		Arch:     "amd64",
	}, nil
}

func (r *DiskReader) Ping(ctx context.Context) error {
	return nil
}

type closeableLimitedReader struct {
	R    io.Reader
	File *os.File
}

func (c *closeableLimitedReader) Read(p []byte) (n int, err error) {
	return c.R.Read(p)
}

func (c *closeableLimitedReader) Close() error {
	return c.File.Close()
}
