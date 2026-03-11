package replication

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewReplicationManager(t *testing.T) {
	logger := zap.NewNop()
	manager := NewReplicationManager(logger)

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.replications)
	assert.NotNil(t, manager.logger)
}

func TestCreateReplicationJob(t *testing.T) {
	logger := zap.NewNop()
	manager := NewReplicationManager(logger)

	job := &ReplicationJob{
		Name:           "Test Replication",
		SourceSite:     "Primary",
		TargetSite:     "DR-Site",
		TargetType:     "s3",
		TargetConfig:   map[string]string{"bucket": "test-bucket"},
		Schedule:       "0 */6 * * *",
		RetentionDays:  30,
		BandwidthLimit: 100,
		Compression:    true,
		Encryption:     true,
		Enabled:        true,
	}

	err := manager.CreateReplicationJob(job)
	require.NoError(t, err)
	assert.NotEmpty(t, job.ID)
	assert.Equal(t, "created", job.LastStatus)
}

func TestGetReplicationJob(t *testing.T) {
	logger := zap.NewNop()
	manager := NewReplicationManager(logger)

	job := &ReplicationJob{
		Name:       "Test Job",
		TargetType: "s3",
	}

	err := manager.CreateReplicationJob(job)
	require.NoError(t, err)

	retrieved, err := manager.GetReplicationJob(job.ID)
	require.NoError(t, err)
	assert.Equal(t, job.Name, retrieved.Name)
	assert.Equal(t, job.TargetType, retrieved.TargetType)
}

func TestGetReplicationJob_NotFound(t *testing.T) {
	logger := zap.NewNop()
	manager := NewReplicationManager(logger)

	_, err := manager.GetReplicationJob("non-existent-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestListReplicationJobs(t *testing.T) {
	logger := zap.NewNop()
	manager := NewReplicationManager(logger)

	// Create multiple jobs
	for i := 0; i < 3; i++ {
		job := &ReplicationJob{
			Name:       "Test Job",
			TargetType: "s3",
		}
		err := manager.CreateReplicationJob(job)
		require.NoError(t, err)
	}

	jobs := manager.ListReplicationJobs()
	assert.Len(t, jobs, 3)
}

func TestDeleteReplicationJob(t *testing.T) {
	logger := zap.NewNop()
	manager := NewReplicationManager(logger)

	job := &ReplicationJob{
		Name:       "Test Job",
		TargetType: "s3",
	}

	err := manager.CreateReplicationJob(job)
	require.NoError(t, err)

	err = manager.DeleteReplicationJob(job.ID)
	require.NoError(t, err)

	_, err = manager.GetReplicationJob(job.ID)
	assert.Error(t, err)
}

func TestDeleteReplicationJob_NotFound(t *testing.T) {
	logger := zap.NewNop()
	manager := NewReplicationManager(logger)

	err := manager.DeleteReplicationJob("non-existent-id")
	assert.Error(t, err)
}

func TestStartReplication(t *testing.T) {
	logger := zap.NewNop()
	manager := NewReplicationManager(logger)

	job := &ReplicationJob{
		Name:       "Test Job",
		TargetType: "s3",
		TargetConfig: map[string]string{
			"bucket": "test-bucket",
			"region": "us-east-1",
		},
	}

	err := manager.CreateReplicationJob(job)
	require.NoError(t, err)

	ctx := context.Background()
	result, err := manager.StartReplication(ctx, job.ID)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, job.ID, result.JobID)
	assert.Equal(t, "completed", result.Status)
	assert.NotZero(t, result.BytesTransferred)
	assert.NotZero(t, result.FilesProcessed)
}

func TestStartReplication_NotFound(t *testing.T) {
	logger := zap.NewNop()
	manager := NewReplicationManager(logger)

	ctx := context.Background()
	_, err := manager.StartReplication(ctx, "non-existent-id")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// CDP Tests

func TestNewCDPManager(t *testing.T) {
	logger := zap.NewNop()
	manager := NewCDPManager(logger)

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.sessions)
	assert.NotNil(t, manager.logger)
}

func TestStartCDP(t *testing.T) {
	logger := zap.NewNop()
	manager := NewCDPManager(logger)

	session, err := manager.StartCDP(
		"Test-VM",
		"42132a5e-75ce-2d35-7b9e-15e96e4e3f21",
		"esxi-01.local",
		"esxi-dr.local",
		"DS-DR",
		300,
	)

	require.NoError(t, err)
	assert.NotNil(t, session)
	assert.NotEmpty(t, session.ID)
	assert.Equal(t, "Test-VM", session.VMName)
	assert.Equal(t, "esxi-01.local", session.SourceHost)
	assert.Equal(t, "esxi-dr.local", session.TargetHost)
	assert.Equal(t, 300, session.RPOSeconds)
	assert.Equal(t, "active", session.Status)
}

func TestGetCDPSession(t *testing.T) {
	logger := zap.NewNop()
	manager := NewCDPManager(logger)

	session, err := manager.StartCDP(
		"Test-VM",
		"42132a5e-75ce-2d35-7b9e-15e96e4e3f21",
		"esxi-01.local",
		"esxi-dr.local",
		"DS-DR",
		300,
	)
	require.NoError(t, err)

	retrieved, err := manager.GetCDPSession(session.ID)
	require.NoError(t, err)
	assert.Equal(t, session.VMName, retrieved.VMName)
	assert.Equal(t, session.SourceHost, retrieved.SourceHost)
}

func TestGetCDPSession_NotFound(t *testing.T) {
	logger := zap.NewNop()
	manager := NewCDPManager(logger)

	_, err := manager.GetCDPSession("non-existent-id")
	assert.Error(t, err)
}

func TestListCDPSessions(t *testing.T) {
	logger := zap.NewNop()
	manager := NewCDPManager(logger)

	// Create multiple sessions
	for i := 0; i < 3; i++ {
		_, err := manager.StartCDP(
			"Test-VM",
			"42132a5e-75ce-2d35-7b9e-15e96e4e3f21",
			"esxi-01.local",
			"esxi-dr.local",
			"DS-DR",
			300,
		)
		require.NoError(t, err)
	}

	sessions := manager.ListCDPSessions()
	assert.Len(t, sessions, 3)
}

func TestStopCDP(t *testing.T) {
	logger := zap.NewNop()
	manager := NewCDPManager(logger)

	session, err := manager.StartCDP(
		"Test-VM",
		"42132a5e-75ce-2d35-7b9e-15e96e4e3f21",
		"esxi-01.local",
		"esxi-dr.local",
		"DS-DR",
		300,
	)
	require.NoError(t, err)

	err = manager.StopCDP(session.ID)
	require.NoError(t, err)

	retrieved, err := manager.GetCDPSession(session.ID)
	require.NoError(t, err)
	assert.Equal(t, "stopped", retrieved.Status)
}

func TestFailoverToReplica(t *testing.T) {
	logger := zap.NewNop()
	manager := NewCDPManager(logger)

	session, err := manager.StartCDP(
		"Critical-VM",
		"42132a5e-75ce-2d35-7b9e-15e96e4e3f21",
		"esxi-01.local",
		"esxi-dr.local",
		"DS-DR",
		300,
	)
	require.NoError(t, err)

	err = manager.FailoverToReplica(session.ID)
	require.NoError(t, err)

	retrieved, err := manager.GetCDPSession(session.ID)
	require.NoError(t, err)
	assert.Equal(t, "failover_complete", retrieved.Status)
}

func TestGetReplicationHealth(t *testing.T) {
	logger := zap.NewNop()
	manager := NewCDPManager(logger)

	session, err := manager.StartCDP(
		"Test-VM",
		"42132a5e-75ce-2d35-7b9e-15e96e4e3f21",
		"esxi-01.local",
		"esxi-dr.local",
		"DS-DR",
		300,
	)
	require.NoError(t, err)

	// Wait a moment for replication to run
	time.Sleep(100 * time.Millisecond)

	health, err := manager.GetReplicationHealth(session.ID)
	require.NoError(t, err)

	assert.Equal(t, session.ID, health["session_id"])
	assert.Equal(t, "active", health["status"])
	assert.Contains(t, health, "last_sync")
	assert.Contains(t, health, "replication_lag")
	assert.Contains(t, health, "healthy")
}

// Helper functions for test setup

func setupTestReplicationJob() *ReplicationJob {
	return &ReplicationJob{
		ID:             "test-job-id",
		Name:           "Test Replication Job",
		SourceSite:     "Primary",
		TargetSite:     "DR-Site",
		TargetType:     "s3",
		TargetConfig:   map[string]string{"bucket": "test-bucket"},
		Schedule:       "0 */6 * * *",
		RetentionDays:  30,
		BandwidthLimit: 100,
		Compression:    true,
		Encryption:     true,
		Enabled:        true,
		CreatedAt:      time.Now(),
	}
}

func setupTestCDPSession() *CDPSession {
	return &CDPSession{
		ID:              "test-cdp-id",
		VMName:          "Test-VM",
		VMUUID:          "42132a5e-75ce-2d35-7b9e-15e96e4e3f21",
		SourceHost:      "esxi-01.local",
		TargetHost:      "esxi-dr.local",
		TargetDatastore: "DS-DR",
		RPOSeconds:      300,
		Status:          "active",
		LastSync:        time.Now(),
		ReplicationLag:  50,
	}
}
