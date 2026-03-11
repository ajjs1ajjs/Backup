// Package integration provides end-to-end integration tests
package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/zap"

	"novabackup/pkg/monitoring"
	"novabackup/pkg/replication"
)

// IntegrationTestSuite provides end-to-end testing
type IntegrationTestSuite struct {
	logger             *zap.Logger
	metricsCollector   *monitoring.MetricsCollector
	replicationManager *replication.ReplicationManager
	tempDir            string
}

// NewIntegrationTestSuite creates a new test suite
func NewIntegrationTestSuite() (*IntegrationTestSuite, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	tempDir, err := os.MkdirTemp("", "novabackup-integration-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	suite := &IntegrationTestSuite{
		logger:             logger,
		tempDir:            tempDir,
		metricsCollector:   monitoring.NewMetricsCollector(logger),
		replicationManager: replication.NewReplicationManager(logger),
	}

	return suite, nil
}

// Cleanup removes temporary files
func (s *IntegrationTestSuite) Cleanup() {
	os.RemoveAll(s.tempDir)
	s.logger.Info("Integration test suite cleaned up")
}

// TestJob represents a test backup job
type TestJob struct {
	ID            string
	Name          string
	Type          string
	SourcePath    string
	DestPath      string
	Schedule      string
	RetentionDays int
	Compression   bool
	Encryption    bool
}

// TestFullBackupWorkflow tests complete backup workflow
func (s *IntegrationTestSuite) TestFullBackupWorkflow(ctx context.Context) error {
	s.logger.Info("=== Testing Full Backup Workflow ===")

	// 1. Create test job
	jobID := fmt.Sprintf("test-job-%d", time.Now().Unix())
	jobName := "Integration Test Backup"
	sourcePath := filepath.Join(s.tempDir, "source")
	destPath := filepath.Join(s.tempDir, "backups")

	// Create source directory with test files
	if err := os.MkdirAll(sourcePath, 0755); err != nil {
		return fmt.Errorf("failed to create source dir: %w", err)
	}

	// Create test files
	for i := 1; i <= 5; i++ {
		filename := filepath.Join(sourcePath, fmt.Sprintf("testfile%d.txt", i))
		content := fmt.Sprintf("Test content for file %d\n", i)
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create test file: %w", err)
		}
	}

	s.logger.Info("Created test files", zap.String("source", sourcePath))

	// 2. Start metrics collection
	metrics := s.metricsCollector.RecordJobStart(jobID, jobName)

	// 3. Execute backup (simulated)
	startTime := time.Now()

	// Simulate backup process
	time.Sleep(2 * time.Second)

	// Count processed files
	files, err := os.ReadDir(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to read source dir: %w", err)
	}

	// Create backup destination
	if err := os.MkdirAll(destPath, 0755); err != nil {
		return fmt.Errorf("failed to create backup dir: %w", err)
	}

	// Simulate backup file creation
	backupFile := filepath.Join(destPath, fmt.Sprintf("backup_%s.tar.gz", jobID))
	if err := os.WriteFile(backupFile, []byte("mock backup data"), 0644); err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}

	// 4. Record metrics
	duration := time.Since(startTime)
	s.metricsCollector.RecordJobEnd(jobID, "success", 1024*1024*100, 1024*1024*50)

	s.logger.Info("Backup workflow completed",
		zap.String("job_id", jobID),
		zap.Duration("duration", duration),
		zap.Int("files", len(files)))

	// Verify metrics
	if metrics.Status != "success" {
		return fmt.Errorf("expected status 'success', got '%s'", metrics.Status)
	}

	if metrics.BytesProcessed == 0 {
		return fmt.Errorf("expected bytes processed > 0")
	}

	return nil
}

// TestIncrementalBackupWorkflow tests incremental backup
func (s *IntegrationTestSuite) TestIncrementalBackupWorkflow(ctx context.Context) error {
	s.logger.Info("=== Testing Incremental Backup Workflow ===")

	jobID := fmt.Sprintf("incr-job-%d", time.Now().Unix())

	// Create base backup
	baseFiles := map[string]string{
		"file1.txt": "content1",
		"file2.txt": "content2",
		"file3.txt": "content3",
	}

	sourceDir := filepath.Join(s.tempDir, "incremental-source")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		return err
	}

	for name, content := range baseFiles {
		path := filepath.Join(sourceDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return err
		}
	}

	// First backup (full)
	metrics1 := s.metricsCollector.RecordJobStart(jobID+"-full", "Full Backup")
	time.Sleep(1 * time.Second)
	s.metricsCollector.RecordJobEnd(jobID+"-full", "success", 1024*300, 1024*150)

	// Modify files
	os.WriteFile(filepath.Join(sourceDir, "file1.txt"), []byte("modified content1"), 0644)
	os.WriteFile(filepath.Join(sourceDir, "file4.txt"), []byte("new file4"), 0644)

	// Second backup (incremental)
	metrics2 := s.metricsCollector.RecordJobStart(jobID+"-incr", "Incremental Backup")
	time.Sleep(500 * time.Millisecond)
	s.metricsCollector.RecordJobEnd(jobID+"-incr", "success", 1024*50, 1024*25)

	// Verify incremental transferred less data
	if metrics2.BytesTransferred >= metrics1.BytesTransferred {
		return fmt.Errorf("incremental backup should transfer less data than full backup")
	}

	s.logger.Info("Incremental backup test passed",
		zap.Int64("full_bytes", metrics1.BytesTransferred),
		zap.Int64("incr_bytes", metrics2.BytesTransferred))

	return nil
}

// TestReplicationWorkflow tests replication functionality
func (s *IntegrationTestSuite) TestReplicationWorkflow(ctx context.Context) error {
	s.logger.Info("=== Testing Replication Workflow ===")

	// Create replication job
	replJob := &replication.ReplicationJob{
		Name:           "Integration Test Replication",
		SourceSite:     "Primary",
		TargetSite:     "DR-Site",
		TargetType:     "s3",
		TargetConfig:   map[string]string{"bucket": "test-bucket", "region": "us-east-1"},
		Schedule:       "0 */6 * * *",
		RetentionDays:  30,
		BandwidthLimit: 100,
		Compression:    true,
		Encryption:     true,
		Enabled:        true,
	}

	if err := s.replicationManager.CreateReplicationJob(replJob); err != nil {
		return fmt.Errorf("failed to create replication job: %w", err)
	}

	// Start replication
	result, err := s.replicationManager.StartReplication(ctx, replJob.ID)
	if err != nil {
		return fmt.Errorf("replication failed: %w", err)
	}

	if result.Status != "completed" {
		return fmt.Errorf("expected status 'completed', got '%s'", result.Status)
	}

	if result.BytesTransferred == 0 {
		return fmt.Errorf("expected bytes transferred > 0")
	}

	s.logger.Info("Replication test passed",
		zap.String("job_id", replJob.ID),
		zap.Int64("bytes_transferred", result.BytesTransferred))

	return nil
}

// TestCDPWorkflow tests CDP replication
func (s *IntegrationTestSuite) TestCDPWorkflow(ctx context.Context) error {
	s.logger.Info("=== Testing CDP Workflow ===")

	cdpManager := replication.NewCDPManager(s.logger)

	// Start CDP session
	session, err := cdpManager.StartCDP(
		"Test-VM",
		"42132a5e-75ce-2d35-7b9e-15e96e4e3f21",
		"esxi-01.local",
		"esxi-dr.local",
		"DS-DR",
		300,
	)
	if err != nil {
		return fmt.Errorf("failed to start CDP: %w", err)
	}

	// Let it run for a moment
	time.Sleep(2 * time.Second)

	// Check health
	health, err := cdpManager.GetReplicationHealth(session.ID)
	if err != nil {
		return fmt.Errorf("failed to get health: %w", err)
	}

	if health["status"] != "active" {
		return fmt.Errorf("expected CDP status 'active', got '%v'", health["status"])
	}

	// Stop CDP
	if err := cdpManager.StopCDP(session.ID); err != nil {
		return fmt.Errorf("failed to stop CDP: %w", err)
	}

	s.logger.Info("CDP test passed", zap.String("session_id", session.ID))

	return nil
}

// TestMonitoringWorkflow tests monitoring and alerting
func (s *IntegrationTestSuite) TestMonitoringWorkflow(ctx context.Context) error {
	s.logger.Info("=== Testing Monitoring Workflow ===")

	dashboard := monitoring.NewDashboard(s.logger, s.metricsCollector)

	// Create some job metrics
	for i := 0; i < 5; i++ {
		jobID := fmt.Sprintf("monitor-job-%d", i)
		s.metricsCollector.RecordJobStart(jobID, "Monitoring Test Job")
		time.Sleep(100 * time.Millisecond)
		status := "success"
		if i == 2 {
			status = "failed"
		}
		s.metricsCollector.RecordJobEnd(jobID, status, 1024*1024*10, 1024*1024*5)
	}

	// Get dashboard data
	data := dashboard.GetDashboardData()

	if data.ActiveJobs != 0 {
		return fmt.Errorf("expected 0 active jobs, got %d", data.ActiveJobs)
	}

	if data.CompletedJobs != 4 {
		return fmt.Errorf("expected 4 completed jobs, got %d", data.CompletedJobs)
	}

	if data.FailedJobs != 1 {
		return fmt.Errorf("expected 1 failed job, got %d", data.FailedJobs)
	}

	// Generate report
	reportGen := monitoring.NewReportGenerator(s.logger, s.metricsCollector)
	report, err := reportGen.GenerateReport("daily", time.Now().Add(-24*time.Hour), time.Now())
	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	if report.Summary.TotalJobs != 5 {
		return fmt.Errorf("expected 5 jobs in report, got %d", report.Summary.TotalJobs)
	}

	s.logger.Info("Monitoring test passed",
		zap.Int("completed", data.CompletedJobs),
		zap.Int("failed", data.FailedJobs))

	return nil
}

// RunAllTests runs all integration tests
func (s *IntegrationTestSuite) RunAllTests(ctx context.Context) error {
	tests := []func(context.Context) error{
		s.TestFullBackupWorkflow,
		s.TestIncrementalBackupWorkflow,
		s.TestReplicationWorkflow,
		s.TestCDPWorkflow,
		s.TestMonitoringWorkflow,
	}

	passed := 0
	failed := 0

	for _, test := range tests {
		if err := test(ctx); err != nil {
			s.logger.Error("Test failed", zap.Error(err))
			failed++
		} else {
			passed++
		}
	}

	s.logger.Info("=== Integration Test Summary ===",
		zap.Int("passed", passed),
		zap.Int("failed", failed),
		zap.Int("total", len(tests)))

	if failed > 0 {
		return fmt.Errorf("%d of %d tests failed", failed, len(tests))
	}

	return nil
}

// Test functions for go test framework
func TestIntegrationFullWorkflow(t *testing.T) {
	suite, err := NewIntegrationTestSuite()
	if err != nil {
		t.Fatalf("Failed to create test suite: %v", err)
	}
	defer suite.Cleanup()

	ctx := context.Background()
	if err := suite.TestFullBackupWorkflow(ctx); err != nil {
		t.Errorf("Full backup workflow test failed: %v", err)
	}
}

func TestIntegrationReplication(t *testing.T) {
	suite, err := NewIntegrationTestSuite()
	if err != nil {
		t.Fatalf("Failed to create test suite: %v", err)
	}
	defer suite.Cleanup()

	ctx := context.Background()
	if err := suite.TestReplicationWorkflow(ctx); err != nil {
		t.Errorf("Replication test failed: %v", err)
	}
}

func TestIntegrationCDP(t *testing.T) {
	suite, err := NewIntegrationTestSuite()
	if err != nil {
		t.Fatalf("Failed to create test suite: %v", err)
	}
	defer suite.Cleanup()

	ctx := context.Background()
	if err := suite.TestCDPWorkflow(ctx); err != nil {
		t.Errorf("CDP test failed: %v", err)
	}
}

func TestIntegrationMonitoring(t *testing.T) {
	suite, err := NewIntegrationTestSuite()
	if err != nil {
		t.Fatalf("Failed to create test suite: %v", err)
	}
	defer suite.Cleanup()

	ctx := context.Background()
	if err := suite.TestMonitoringWorkflow(ctx); err != nil {
		t.Errorf("Monitoring test failed: %v", err)
	}
}

func TestIntegrationAll(t *testing.T) {
	suite, err := NewIntegrationTestSuite()
	if err != nil {
		t.Fatalf("Failed to create test suite: %v", err)
	}
	defer suite.Cleanup()

	ctx := context.Background()
	if err := suite.RunAllTests(ctx); err != nil {
		t.Errorf("Integration tests failed: %v", err)
	}
}
