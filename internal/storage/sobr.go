package storage

import (
	"context"
	"fmt"
	"io"
	"sync"
)

// RepositoryPool implements the Scale-Out Backup Repository (SOBR)
type RepositoryPool struct {
	performanceTier []Provider
	capacityTier    []Provider
	mu             sync.RWMutex
}

func NewRepositoryPool() *RepositoryPool {
	return &RepositoryPool{
		performanceTier: make([]Provider, 0),
		capacityTier:    make([]Provider, 0),
	}
}

func (p *RepositoryPool) AddPerformanceExtent(provider Provider) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.performanceTier = append(p.performanceTier, provider)
}

func (p *RepositoryPool) AddCapacityExtent(provider Provider) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.capacityTier = append(p.capacityTier, provider)
}

// Store implementation for SOBR (always stores to Performance Tier first)
func (p *RepositoryPool) Store(ctx context.Context, key string, data io.Reader, size int64) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.performanceTier) == 0 {
		return fmt.Errorf("no performance extents available in pool")
	}

	// Simple round-robin or first-available for now
	// In production, this would choose the extent with the most free space
	return p.performanceTier[0].Store(ctx, key, data, size)
}

// Retrieve implementation for SOBR (checks Performance then Capacity)
func (p *RepositoryPool) Retrieve(ctx context.Context, key string) (io.ReadCloser, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Check performance tier
	for _, provider := range p.performanceTier {
		exists, _ := provider.Exists(ctx, key)
		if exists {
			return provider.Retrieve(ctx, key)
		}
	}

	// Check capacity tier
	for _, provider := range p.capacityTier {
		exists, _ := provider.Exists(ctx, key)
		if exists {
			return provider.Retrieve(ctx, key)
		}
	}

	return nil, fmt.Errorf("object %s not found in pool", key)
}

func (p *RepositoryPool) Delete(ctx context.Context, key string) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, provider := range p.performanceTier {
		provider.Delete(ctx, key)
	}
	for _, provider := range p.capacityTier {
		provider.Delete(ctx, key)
	}
	return nil
}

func (p *RepositoryPool) Exists(ctx context.Context, key string) (bool, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, provider := range p.performanceTier {
		if ok, _ := provider.Exists(ctx, key); ok {
			return true, nil
		}
	}
	for _, provider := range p.capacityTier {
		if ok, _ := provider.Exists(ctx, key); ok {
			return true, nil
		}
	}
	return false, nil
}

func (p *RepositoryPool) GetStats(ctx context.Context) (*Stats, error) {
	totalStats := &Stats{Provider: "sobr"}
	
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, provider := range p.performanceTier {
		s, err := provider.GetStats(ctx)
		if err == nil {
			totalStats.TotalSize += s.TotalSize
			totalStats.UsedSize += s.UsedSize
			totalStats.ObjectCount += s.ObjectCount
		}
	}
	return totalStats, nil
}

func (p *RepositoryPool) Close() error {
	return nil
}

// TieringJob moves old data from Performance to Capacity tier
func (p *RepositoryPool) TieringJob(ctx context.Context, key string) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.capacityTier) == 0 {
		return nil // No capacity tier defined
	}

	// 1. Retrieve from performance
	reader, err := p.Retrieve(ctx, key)
	if err != nil {
		return err
	}
	defer reader.Close()

	// 2. Store in capacity
	err = p.capacityTier[0].Store(ctx, key, reader, 0) // Size 0 is fine for S3 uploader
	if err != nil {
		return err
	}

	// 3. Delete from performance
	return p.Delete(ctx, key)
}
