package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"time"
)

// TieringPolicy defines when data should be moved from Performance to Capacity tier
type TieringPolicy struct {
	MaxAgeDays      int     // Move objects older than N days (requires catalog)
	MaxSizePercent  float64 // Move when Performance tier is X% full
	Enabled         bool
}

// ExtentStats holds real-time stats for an extent used in smart selection
type ExtentStats struct {
	FreeBytes int64
}

// RepositoryPool implements the Scale-Out Backup Repository (SOBR)
type RepositoryPool struct {
	performanceTier []Provider
	capacityTier    []Provider
	policy          TieringPolicy
	mu              sync.RWMutex
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

// SetTieringPolicy sets the automated tiering rules
func (p *RepositoryPool) SetTieringPolicy(policy TieringPolicy) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.policy = policy
}

// selectBestPerformanceExtent returns the Performance extent with most free space
func (p *RepositoryPool) selectBestPerformanceExtent(ctx context.Context) Provider {
	var best Provider
	var bestFree int64 = -1

	for _, ext := range p.performanceTier {
		s, err := ext.GetStats(ctx)
		if err != nil {
			continue
		}
		free := s.TotalSize - s.UsedSize
		if free > bestFree {
			bestFree = free
			best = ext
		}
	}

	if best == nil && len(p.performanceTier) > 0 {
		best = p.performanceTier[0] // fallback
	}
	return best
}

// Store implementation for SOBR (always stores to best Performance Tier extent)
func (p *RepositoryPool) Store(ctx context.Context, key string, data io.Reader, size int64) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.performanceTier) == 0 {
		return fmt.Errorf("no performance extents available in pool")
	}

	ext := p.selectBestPerformanceExtent(ctx)
	if ext == nil {
		return fmt.Errorf("no available performance extent")
	}
	return ext.Store(ctx, key, data, size)
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

// TieringJob moves a specific object from Performance to Capacity tier
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

	// 3. Delete from performance tier only
	for _, provider := range p.performanceTier {
		provider.Delete(ctx, key)
	}
	return nil
}

// RunTieringPolicy evaluates the policy and tiers eligible objects.
// keys is the full list of object keys currently known (from catalog/DB).
// createdAt maps key → creation time to implement MaxAgeDays.
func (p *RepositoryPool) RunTieringPolicy(ctx context.Context, keys []string, createdAt map[string]time.Time) (moved int, err error) {
	p.mu.RLock()
	pol := p.policy
	p.mu.RUnlock()

	if !pol.Enabled || len(p.capacityTier) == 0 {
		return 0, nil
	}

	// Check overall performance tier fullness
	perfStats, _ := p.GetStats(ctx)
	var tierFull bool
	if pol.MaxSizePercent > 0 && perfStats != nil && perfStats.TotalSize > 0 {
		usedPct := float64(perfStats.UsedSize) / float64(perfStats.TotalSize) * 100
		tierFull = usedPct >= pol.MaxSizePercent
	}

	for _, key := range keys {
		shouldTier := tierFull

		// Age-based rule
		if pol.MaxAgeDays > 0 {
			if t, ok := createdAt[key]; ok {
				if time.Since(t) >= time.Duration(pol.MaxAgeDays)*24*time.Hour {
					shouldTier = true
				}
			}
		}

		if !shouldTier {
			continue
		}

		if e := p.TieringJob(ctx, key); e != nil {
			log.Printf("[SOBR] Tiering failed for key %s: %v", key, e)
		} else {
			moved++
			log.Printf("[SOBR] Tiered key %s to Capacity tier", key)
		}
	}

	return moved, nil
}
