package retention

import (
	"context"
	"fmt"
	"sort"
	"time"
)

// GFSRetention implements Grandfather-Father-Son retention policy
type GFSRetention struct {
	// Daily backups to keep
	Daily int `json:"daily"`
	// Weekly backups to keep (on specific day, usually Sunday)
	Weekly int `json:"weekly"`
	// Monthly backups to keep
	Monthly int `json:"monthly"`
	// Yearly backups to keep
	Yearly int `json:"yearly"`
	// Day of week for weekly (0=Sunday, 1=Monday)
	WeeklyDay int `json:"weekly_day"`
	// Day of month for monthly (1-31)
	MonthlyDay int `json:"monthly_day"`
}

// RetentionPolicy combines all retention options
type RetentionPolicy struct {
	GFS        *GFSRetention    `json:"gfs"`
	Simple     *SimpleRetention `json:"simple"`
	BackupType string           `json:"backup_type"` // "gfs", "simple"
}

// SimpleRetention defines simple retention by count
type SimpleRetention struct {
	KeepLast int `json:"keep_last"` // Keep last N backups
}

// GFSManager manages GFS retention
type GFSManager struct {
	policy RetentionPolicy
}

// NewGFSManager creates a new GFS retention manager
func NewGFSManager(policy RetentionPolicy) *GFSManager {
	if policy.GFS == nil {
		policy.GFS = &GFSRetention{
			Daily:      7,
			Weekly:     4,
			Monthly:    12,
			Yearly:     7,
			WeeklyDay:  0, // Sunday
			MonthlyDay: 1,
		}
	}
	return &GFSManager{policy: policy}
}

// BackupPoint represents a single backup point
type BackupPoint struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	SizeGB    int64     `json:"size_gb"`
	Type      string    `json:"type"` // "full", "incremental", "differential"
	JobID     string    `json:"job_id"`
	Path      string    `json:"path"`
}

// CalculateRetainedPoints determines which points to keep
func (g *GFSManager) CalculateRetainedPoints(ctx context.Context, points []BackupPoint) ([]BackupPoint, error) {
	if len(points) == 0 {
		return []BackupPoint{}, nil
	}

	// Sort by timestamp descending (newest first)
	sorted := make([]BackupPoint, len(points))
	copy(sorted, points)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Timestamp.After(sorted[j].Timestamp)
	})

	var kept []BackupPoint
	now := time.Now()

	// Categorize points
	var daily, weekly, monthly, yearly, others []BackupPoint

	for _, p := range sorted {
		age := now.Sub(p.Timestamp)
		ageHours := age.Hours()

		// Yearly check
		if p.Timestamp.Day() == g.policy.GFS.MonthlyDay &&
			p.Timestamp.Month() == time.January &&
			ageHours <= 24*365*float64(g.policy.GFS.Yearly) {
			yearly = append(yearly, p)
			continue
		}

		// Monthly check
		if p.Timestamp.Day() == g.policy.GFS.MonthlyDay &&
			ageHours <= 24*30*float64(g.policy.GFS.Monthly) {
			monthly = append(monthly, p)
			continue
		}

		// Weekly check
		if int(p.Timestamp.Weekday()) == g.policy.GFS.WeeklyDay &&
			ageHours <= 24*7*float64(g.policy.GFS.Weekly) {
			weekly = append(weekly, p)
			continue
		}

		// Daily check
		if ageHours <= 24*float64(g.policy.GFS.Daily) {
			daily = append(daily, p)
			continue
		}

		others = append(others, p)
	}

	// Apply limits
	if len(yearly) > g.policy.GFS.Yearly {
		yearly = yearly[:g.policy.GFS.Yearly]
	}
	if len(monthly) > g.policy.GFS.Monthly {
		monthly = monthly[:g.policy.GFS.Monthly]
	}
	if len(weekly) > g.policy.GFS.Weekly {
		weekly = weekly[:g.policy.GFS.Weekly]
	}
	if len(daily) > g.policy.GFS.Daily {
		daily = daily[:g.policy.GFS.Daily]
	}

	// Combine kept points
	kept = append(kept, yearly...)
	kept = append(kept, monthly...)
	kept = append(kept, weekly...)
	kept = append(kept, daily...)

	// Sort final result by timestamp descending
	sort.Slice(kept, func(i, j int) bool {
		return kept[i].Timestamp.After(kept[j].Timestamp)
	})

	return kept, nil
}

// GetPointsToDelete returns points that should be deleted
func (g *GFSManager) GetPointsToDelete(ctx context.Context, points []BackupPoint) ([]BackupPoint, error) {
	retained, err := g.CalculateRetainedPoints(ctx, points)
	if err != nil {
		return nil, err
	}

	retainedMap := make(map[string]bool)
	for _, p := range retained {
		retainedMap[p.ID] = true
	}

	var toDelete []BackupPoint
	for _, p := range points {
		if !retainedMap[p.ID] {
			toDelete = append(toDelete, p)
		}
	}

	return toDelete, nil
}

// SyntheticFullManager manages synthetic full backup creation
type SyntheticFullManager struct {
	incrementalBase *BackupPoint
}

// NewSyntheticFullManager creates a new synthetic full manager
func NewSyntheticFullManager() *SyntheticFullManager {
	return &SyntheticFullManager{}
}

// CreateSyntheticFull creates a synthetic full backup from incremental backups
func (s *SyntheticFullManager) CreateSyntheticFull(ctx context.Context, incBackups []BackupPoint) (*BackupPoint, error) {
	if len(incBackups) == 0 {
		return nil, fmt.Errorf("no incremental backups provided")
	}

	// Sort by timestamp
	sorted := make([]BackupPoint, len(incBackups))
	copy(sorted, incBackups)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Timestamp.Before(sorted[j].Timestamp)
	})

	var totalSize int64
	for _, b := range sorted {
		totalSize += b.SizeGB
	}

	return &BackupPoint{
		ID:        fmt.Sprintf("synth-full-%d", time.Now().Unix()),
		Timestamp: time.Now(),
		SizeGB:    totalSize,
		Type:      "full",
		Path:      sorted[len(sorted)-1].Path, // Use latest path
	}, nil
}
