// Package monitoring provides monitoring, analytics and reporting functionality
package monitoring

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// MetricsCollector collects and stores backup metrics
type MetricsCollector struct {
	logger   *zap.Logger
	metrics  map[string]*JobMetrics
	mu       sync.RWMutex
}

// JobMetrics contains metrics for a backup job
type JobMetrics struct {
	JobID           string                 `json:"job_id"`
	JobName         string                 `json:"job_name"`
	StartTime       time.Time              `json:"start_time"`
	EndTime         time.Time              `json:"end_time"`
	Duration        time.Duration          `json:"duration"`
	Status          string                 `json:"status"`
	BytesProcessed  int64                  `json:"bytes_processed"`
	BytesTransferred int64                 `json:"bytes_transferred"`
	FilesProcessed  int                    `json:"files_processed"`
	SuccessCount    int                    `json:"success_count"`
	FailureCount    int                    `json:"failure_count"`
	Warnings        []string               `json:"warnings"`
	ThroughputMBps  float64                `json:"throughput_mbps"`
	DedupeRatio     float64                `json:"dedupe_ratio"`
	CompressionRatio float64               `json:"compression_ratio"`
}

// Dashboard provides real-time dashboard data
type Dashboard struct {
	logger    *zap.Logger
	collector *MetricsCollector
}

// DashboardData contains all dashboard metrics
type DashboardData struct {
	Timestamp       time.Time              `json:"timestamp"`
	ActiveJobs      int                    `json:"active_jobs"`
	CompletedJobs   int                    `json:"completed_jobs_24h"`
	FailedJobs      int                    `json:"failed_jobs_24h"`
	TotalBackupSize int64                  `json:"total_backup_size"`
	StorageSavings  int64                  `json:"storage_savings"`
	JobStatus       map[string]JobMetrics  `json:"job_status"`
	RecentSessions  []BackupSession      `json:"recent_sessions"`
	Alerts          []Alert              `json:"alerts"`
}

// BackupSession represents a backup session
type BackupSession struct {
	SessionID   string    `json:"session_id"`
	JobName     string    `json:"job_name"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Status      string    `json:"status"`
	VMCount     int       `json:"vm_count"`
	DataSizeGB  float64   `json:"data_size_gb"`
}

// Alert represents a monitoring alert
type Alert struct {
	ID          string    `json:"id"`
	Severity    string    `json:"severity"` // info, warning, error, critical
	Type        string    `json:"type"`     // job_failed, storage_low, rpo_missed
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
	Acknowledged bool     `json:"acknowledged"`
}

// ReportGenerator generates backup reports
type ReportGenerator struct {
	logger    *zap.Logger
	collector *MetricsCollector
}

// Report contains backup report data
type Report struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Type            string                 `json:"type"` // daily, weekly, monthly
	PeriodStart     time.Time              `json:"period_start"`
	PeriodEnd       time.Time              `json:"period_end"`
	GeneratedAt     time.Time              `json:"generated_at"`
	Summary         ReportSummary          `json:"summary"`
	JobDetails      []JobMetrics           `json:"job_details"`
	StorageDetails  StorageReport          `json:"storage_details"`
}

// ReportSummary contains report summary
type ReportSummary struct {
	TotalJobs       int     `json:"total_jobs"`
	SuccessfulJobs  int     `json:"successful_jobs"`
	FailedJobs      int     `json:"failed_jobs"`
	SuccessRate     float64 `json:"success_rate"`
	TotalDataGB     float64 `json:"total_data_gb"`
	TransferredGB   float64 `json:"transferred_gb"`
	StorageSavedGB  float64 `json:"storage_saved_gb"`
	AvgDuration     string  `json:"avg_duration"`
}

// StorageReport contains storage metrics
type StorageReport struct {
	TotalCapacityGB float64                `json:"total_capacity_gb"`
	UsedSpaceGB     float64                `json:"used_space_gb"`
	FreeSpaceGB     float64                `json:"free_space_gb"`
	UsagePercent    float64                `json:"usage_percent"`
	RepoDetails     []RepositoryMetrics    `json:"repository_details"`
}

// RepositoryMetrics contains repository metrics
type RepositoryMetrics struct {
	Name         string  `json:"name"`
	Type         string  `json:"type"`
	TotalGB      float64 `json:"total_gb"`
	UsedGB       float64 `json:"used_gb"`
	FreeGB       float64 `json:"free_gb"`
	BackupCount  int     `json:"backup_count"`
	GrowthRate   float64 `json:"growth_rate_gb_per_day"`
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(logger *zap.Logger) *MetricsCollector {
	return &MetricsCollector{
		logger:  logger.With(zap.String("component", "metrics")),
		metrics: make(map[string]*JobMetrics),
	}
}

// RecordJobStart records the start of a backup job
func (m *MetricsCollector) RecordJobStart(jobID, jobName string) *JobMetrics {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	metric := &JobMetrics{
		JobID:     jobID,
		JobName:   jobName,
		StartTime: time.Now(),
		Status:    "running",
	}
	
	m.metrics[jobID] = metric
	m.logger.Info("Job started", zap.String("job", jobName))
	
	return metric
}

// RecordJobEnd records the end of a backup job
func (m *MetricsCollector) RecordJobEnd(jobID string, status string, bytesProcessed, bytesTransferred int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	metric, exists := m.metrics[jobID]
	if !exists {
		return
	}
	
	metric.EndTime = time.Now()
	metric.Duration = metric.EndTime.Sub(metric.StartTime)
	metric.Status = status
	metric.BytesProcessed = bytesProcessed
	metric.BytesTransferred = bytesTransferred
	
	if metric.Duration > 0 {
		metric.ThroughputMBps = float64(bytesTransferred) / (1024 * 1024) / metric.Duration.Seconds()
	}
	
	if bytesProcessed > 0 {
		metric.CompressionRatio = float64(bytesProcessed) / float64(bytesTransferred)
	}
	
	if status == "success" {
		metric.SuccessCount++
	} else {
		metric.FailureCount++
	}
	
	m.logger.Info("Job completed",
		zap.String("job", metric.JobName),
		zap.String("status", status),
		zap.Duration("duration", metric.Duration),
		zap.Float64("throughput_mbps", metric.ThroughputMBps))
}

// GetJobMetrics returns metrics for a specific job
func (m *MetricsCollector) GetJobMetrics(jobID string) (*JobMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	metric, exists := m.metrics[jobID]
	if !exists {
		return nil, fmt.Errorf("job metrics not found: %s", jobID)
	}
	
	return metric, nil
}

// GetAllMetrics returns all collected metrics
func (m *MetricsCollector) GetAllMetrics() []*JobMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	metrics := make([]*JobMetrics, 0, len(m.metrics))
	for _, m := range m.metrics {
		metrics = append(metrics, m)
	}
	return metrics
}

// NewDashboard creates a new dashboard
func NewDashboard(logger *zap.Logger, collector *MetricsCollector) *Dashboard {
	return &Dashboard{
		logger:    logger.With(zap.String("component", "dashboard")),
		collector: collector,
	}
}

// GetDashboardData returns current dashboard data
func (d *Dashboard) GetDashboardData() *DashboardData {
	allMetrics := d.collector.GetAllMetrics()
	
	data := &DashboardData{
		Timestamp:       time.Now(),
		JobStatus:       make(map[string]JobMetrics),
		RecentSessions:  []BackupSession{},
		Alerts:          []Alert{},
	}
	
	now := time.Now()
	dayAgo := now.Add(-24 * time.Hour)
	
	for _, metric := range allMetrics {
		// Count active jobs
		if metric.Status == "running" {
			data.ActiveJobs++
		}
		
		// Count jobs in last 24h
		if metric.EndTime.After(dayAgo) || metric.StartTime.After(dayAgo) {
			if metric.Status == "success" {
				data.CompletedJobs++
			} else if metric.Status == "failed" {
				data.FailedJobs++
			}
		}
		
		data.TotalBackupSize += metric.BytesTransferred
		data.JobStatus[metric.JobID] = *metric
		
		// Add to recent sessions
		session := BackupSession{
			SessionID:  metric.JobID,
			JobName:    metric.JobName,
			StartTime:  metric.StartTime,
			EndTime:    metric.EndTime,
			Status:     metric.Status,
			DataSizeGB: float64(metric.BytesTransferred) / (1024 * 1024 * 1024),
		}
		data.RecentSessions = append(data.RecentSessions, session)
	}
	
	// Calculate storage savings
	for _, metric := range allMetrics {
		if metric.CompressionRatio > 0 {
			saved := int64(float64(metric.BytesProcessed) * (1 - 1/metric.CompressionRatio))
			data.StorageSavings += saved
		}
	}
	
	// Generate sample alerts
	if data.FailedJobs > 0 {
		data.Alerts = append(data.Alerts, Alert{
			ID:        fmt.Sprintf("alert_%d", time.Now().Unix()),
			Severity:  "warning",
			Type:      "job_failed",
			Message:   fmt.Sprintf("%d jobs failed in the last 24 hours", data.FailedJobs),
			Timestamp: time.Now(),
		})
	}
	
	return data
}

// NewReportGenerator creates a new report generator
func NewReportGenerator(logger *zap.Logger, collector *MetricsCollector) *ReportGenerator {
	return &ReportGenerator{
		logger:    logger.With(zap.String("component", "reports")),
		collector: collector,
	}
}

// GenerateReport generates a backup report
func (r *ReportGenerator) GenerateReport(reportType string, periodStart, periodEnd time.Time) (*Report, error) {
	r.logger.Info("Generating report",
		zap.String("type", reportType),
		zap.Time("start", periodStart),
		zap.Time("end", periodEnd))
	
	allMetrics := r.collector.GetAllMetrics()
	
	report := &Report{
		ID:          fmt.Sprintf("report_%d", time.Now().Unix()),
		Name:        fmt.Sprintf("%s Report", reportType),
		Type:        reportType,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		GeneratedAt: time.Now(),
		JobDetails:  []JobMetrics{},
	}
	
	var totalDuration time.Duration
	successCount := 0
	var totalBytes, transferredBytes int64
	
	for _, metric := range allMetrics {
		// Filter by period
		if metric.StartTime.Before(periodEnd) && (metric.EndTime.IsZero() || metric.EndTime.After(periodStart)) {
			report.JobDetails = append(report.JobDetails, *metric)
			
			totalDuration += metric.Duration
			totalBytes += metric.BytesProcessed
			transferredBytes += metric.BytesTransferred
			
			if metric.Status == "success" {
				successCount++
			}
		}
	}
	
	totalJobs := len(report.JobDetails)
	successRate := 0.0
	if totalJobs > 0 {
		successRate = float64(successCount) / float64(totalJobs) * 100
	}
	
	avgDuration := "0s"
	if totalJobs > 0 {
		avgDuration = (totalDuration / time.Duration(totalJobs)).String()
	}
	
	savedGB := 0.0
	if totalBytes > 0 && transferredBytes > 0 {
		savedGB = float64(totalBytes-transferredBytes) / (1024 * 1024 * 1024)
	}
	
	report.Summary = ReportSummary{
		TotalJobs:      totalJobs,
		SuccessfulJobs: successCount,
		FailedJobs:     totalJobs - successCount,
		SuccessRate:    successRate,
		TotalDataGB:    float64(totalBytes) / (1024 * 1024 * 1024),
		TransferredGB:  float64(transferredBytes) / (1024 * 1024 * 1024),
		StorageSavedGB: savedGB,
		AvgDuration:    avgDuration,
	}
	
	report.StorageDetails = StorageReport{
		TotalCapacityGB: 1000,
		UsedSpaceGB:     float64(transferredBytes) / (1024 * 1024 * 1024),
		FreeSpaceGB:     1000 - float64(transferredBytes)/(1024*1024*1024),
		UsagePercent:    float64(transferredBytes) / (1024 * 1024 * 1024) / 10,
		RepoDetails: []RepositoryMetrics{
			{
				Name:        "Default Repository",
				Type:        "local",
				TotalGB:     500,
				UsedGB:      float64(transferredBytes) / (1024 * 1024 * 1024),
				FreeGB:      500 - float64(transferredBytes)/(1024*1024*1024),
				BackupCount: totalJobs,
				GrowthRate:  10.5,
			},
		},
	}
	
	r.logger.Info("Report generated",
		zap.String("id", report.ID),
		zap.Int("jobs", totalJobs),
		zap.Float64("success_rate", successRate))
	
	return report, nil
}

// CreateAlert creates a monitoring alert
func (d *Dashboard) CreateAlert(severity, alertType, message string) *Alert {
	alert := &Alert{
		ID:           fmt.Sprintf("alert_%d", time.Now().Unix()),
		Severity:     severity,
		Type:         alertType,
		Message:      message,
		Timestamp:    time.Now(),
		Acknowledged: false,
	}
	
	d.logger.Info("Alert created",
		zap.String("id", alert.ID),
		zap.String("severity", severity),
		zap.String("type", alertType))
	
	return alert
}

// GetJobPerformanceTrend returns performance trend for a job
func (m *MetricsCollector) GetJobPerformanceTrend(jobID string, days int) ([]JobMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// In production, query historical data from database
	// For now, return current metrics
	metric, exists := m.metrics[jobID]
	if !exists {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}
	
	return []JobMetrics{*metric}, nil
}
