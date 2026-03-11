package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// LocalProvider implements the Provider interface for local disk
type LocalProvider struct {
	basePath string
}

func NewLocalProvider(basePath string) (*LocalProvider, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base path: %w", err)
	}
	return &LocalProvider{basePath: basePath}, nil
}

func (p *LocalProvider) Store(ctx context.Context, key string, data io.Reader, size int64) error {
	path := filepath.Join(p.basePath, key)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, data)
	return err
}

func (p *LocalProvider) Retrieve(ctx context.Context, key string) (io.ReadCloser, error) {
	path := filepath.Join(p.basePath, key)
	return os.Open(path)
}

func (p *LocalProvider) Delete(ctx context.Context, key string) error {
	path := filepath.Join(p.basePath, key)
	return os.Remove(path)
}

func (p *LocalProvider) Exists(ctx context.Context, key string) (bool, error) {
	path := filepath.Join(p.basePath, key)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

func (p *LocalProvider) GetStats(ctx context.Context) (*Stats, error) {
	var size int64
	var count int64
	err := filepath.Walk(p.basePath, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
			count++
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &Stats{
		TotalSize:   0,
		UsedSize:    size,
		ObjectCount: count,
		Provider:    "local",
	}, nil
}

func (p *LocalProvider) Close() error {
	return nil
}
