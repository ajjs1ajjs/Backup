package database

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Helper function to create test database
func newTestDatabase(t *testing.T) (*Database, func()) {
	t.Helper()

	// Create temp directory
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.Remove(dbPath)
	}

	return db, cleanup
}

func TestNewDatabase(t *testing.T) {
	db, cleanup := newTestDatabase(t)
	defer cleanup()

	if db == nil {
		t.Fatal("Expected non-nil database")
	}

	if db.db == nil {
		t.Fatal("Expected non-nil underlying database connection")
	}
}

func TestDatabase_CreateJob(t *testing.T) {
	db, cleanup := newTestDatabase(t)
	defer cleanup()

	job := &Job{
		ID:          "job-1",
		Name:        "Test Job",
		Type:        "file",
		Sources:     []string{"/path/to/source"},
		Destination: "/path/to/dest",
		Compression: true,
		Encryption:  false,
		Schedule:    "daily",
		Enabled:     true,
		CreatedAt:   time.Now(),
	}

	err := db.CreateJob(job)
	if err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	// Verify job was created
	retrieved, err := db.GetJob("job-1")
	if err != nil {
		t.Fatalf("Failed to get job: %v", err)
	}

	if retrieved.Name != job.Name {
		t.Errorf("Expected name '%s', got '%s'", job.Name, retrieved.Name)
	}

	if retrieved.Type != job.Type {
		t.Errorf("Expected type '%s', got '%s'", job.Type, retrieved.Type)
	}

	if !retrieved.Compression {
		t.Error("Expected compression to be true")
	}

	if retrieved.Encryption {
		t.Error("Expected encryption to be false")
	}
}

func TestDatabase_ListJobs(t *testing.T) {
	db, cleanup := newTestDatabase(t)
	defer cleanup()

	// Create multiple jobs with different timestamps
	jobs := []*Job{
		{ID: "job-1", Name: "Job 1", Type: "file", Sources: []string{"/src1"}, Destination: "/dst1", CreatedAt: time.Now().Add(-2 * time.Second)},
		{ID: "job-2", Name: "Job 2", Type: "database", Sources: []string{"/src2"}, Destination: "/dst2", CreatedAt: time.Now().Add(-1 * time.Second)},
		{ID: "job-3", Name: "Job 3", Type: "file", Sources: []string{"/src3"}, Destination: "/dst3", CreatedAt: time.Now()},
	}

	for _, job := range jobs {
		if err := db.CreateJob(job); err != nil {
			t.Fatalf("Failed to create job: %v", err)
		}
	}

	// List jobs
	retrieved, err := db.ListJobs()
	if err != nil {
		t.Fatalf("Failed to list jobs: %v", err)
	}

	if len(retrieved) != 3 {
		t.Errorf("Expected 3 jobs, got %d", len(retrieved))
	}
}

func TestDatabase_UpdateJob(t *testing.T) {
	db, cleanup := newTestDatabase(t)
	defer cleanup()

	// Create initial job
	job := &Job{
		ID:          "job-update",
		Name:        "Original Name",
		Type:        "file",
		Sources:     []string{"/original"},
		Destination: "/original/dest",
		Enabled:     true,
	}
	if err := db.CreateJob(job); err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	// Update job
	job.Name = "Updated Name"
	job.Sources = []string{"/updated1", "/updated2"}
	job.Enabled = false

	if err := db.UpdateJob(job); err != nil {
		t.Fatalf("Failed to update job: %v", err)
	}

	// Verify update
	retrieved, err := db.GetJob("job-update")
	if err != nil {
		t.Fatalf("Failed to get job: %v", err)
	}

	if retrieved.Name != "Updated Name" {
		t.Errorf("Expected updated name, got '%s'", retrieved.Name)
	}

	if retrieved.Enabled {
		t.Error("Expected job to be disabled")
	}

	if len(retrieved.Sources) != 2 {
		t.Errorf("Expected 2 sources, got %d", len(retrieved.Sources))
	}
}

func TestDatabase_DeleteJob(t *testing.T) {
	db, cleanup := newTestDatabase(t)
	defer cleanup()

	// Create job
	job := &Job{
		ID:   "job-delete",
		Name: "To Delete",
		Type: "file",
	}
	if err := db.CreateJob(job); err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	// Delete job
	if err := db.DeleteJob("job-delete"); err != nil {
		t.Fatalf("Failed to delete job: %v", err)
	}

	// Verify deletion
	_, err := db.GetJob("job-delete")
	if err == nil {
		t.Error("Expected error when getting deleted job")
	}
}

func TestDatabase_CreateSession(t *testing.T) {
	db, cleanup := newTestDatabase(t)
	defer cleanup()

	now := time.Now()
	session := &Session{
		ID:             "session-1",
		JobID:          "job-1",
		JobName:        "Test Job",
		StartTime:      now,
		EndTime:        now.Add(5 * time.Minute),
		Status:         "completed",
		FilesProcessed: 150,
		BytesWritten:   1024 * 1024 * 100, // 100 MB
		Error:          "",
	}

	err := db.CreateSession(session)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Verify session
	sessions, err := db.ListSessions()
	if err != nil {
		t.Fatalf("Failed to list sessions: %v", err)
	}

	if len(sessions) != 1 {
		t.Errorf("Expected 1 session, got %d", len(sessions))
	}

	retrieved := sessions[0]
	if retrieved.JobID != session.JobID {
		t.Errorf("Expected job_id '%s', got '%s'", session.JobID, retrieved.JobID)
	}

	if retrieved.FilesProcessed != 150 {
		t.Errorf("Expected 150 files processed, got %d", retrieved.FilesProcessed)
	}

	if retrieved.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", retrieved.Status)
	}
}

func TestDatabase_ListSessionsLimit(t *testing.T) {
	db, cleanup := newTestDatabase(t)
	defer cleanup()

	// Create 150 sessions with unique IDs
	for i := 0; i < 150; i++ {
		session := &Session{
			ID:             filepath.Join("session", fmt.Sprintf("%d", i)),
			JobID:          "job-1",
			JobName:        "Test Job",
			StartTime:      time.Now().Add(time.Duration(-i) * time.Hour),
			EndTime:        time.Now(),
			Status:         "completed",
			FilesProcessed: i,
			BytesWritten:   int64(i * 1000),
		}
		if err := db.CreateSession(session); err != nil {
			t.Fatalf("Failed to create session %d: %v", i, err)
		}
	}

	// List sessions - should be limited to 100
	sessions, err := db.ListSessions()
	if err != nil {
		t.Fatalf("Failed to list sessions: %v", err)
	}

	if len(sessions) != 100 {
		t.Errorf("Expected 100 sessions (limit), got %d", len(sessions))
	}
}

func TestDatabase_UpdateJobLastRun(t *testing.T) {
	db, cleanup := newTestDatabase(t)
	defer cleanup()

	// Create job
	job := &Job{
		ID:   "job-lastrun",
		Name: "Test Job",
		Type: "file",
	}
	if err := db.CreateJob(job); err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	// Update last run
	lastRun := time.Now()
	nextRun := lastRun.Add(24 * time.Hour)

	if err := db.UpdateJobLastRun("job-lastrun", lastRun, nextRun); err != nil {
		t.Fatalf("Failed to update last run: %v", err)
	}

	// Verify update
	retrieved, err := db.GetJob("job-lastrun")
	if err != nil {
		t.Fatalf("Failed to get job: %v", err)
	}

	if retrieved.LastRun == nil {
		t.Fatal("Expected LastRun to be set")
	}

	if retrieved.NextRun == nil {
		t.Fatal("Expected NextRun to be set")
	}

	// Check times are close (within 1 second due to SQLite precision)
	if retrieved.LastRun.Sub(lastRun).Abs() > time.Second {
		t.Errorf("LastRun time mismatch: expected %v, got %v", lastRun, retrieved.LastRun)
	}
}

func TestDatabase_GetJobNotFound(t *testing.T) {
	db, cleanup := newTestDatabase(t)
	defer cleanup()

	_, err := db.GetJob("nonexistent-job")
	if err == nil {
		t.Error("Expected error when getting non-existent job")
	}
}

func TestDatabase_JobWithNullTimes(t *testing.T) {
	db, cleanup := newTestDatabase(t)
	defer cleanup()

	// Create job without setting times
	job := &Job{
		ID:          "job-null-times",
		Name:        "Test Job",
		Type:        "file",
		Destination: "/dest",
	}
	if err := db.CreateJob(job); err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	// Get job and verify null times are handled
	retrieved, err := db.GetJob("job-null-times")
	if err != nil {
		t.Fatalf("Failed to get job: %v", err)
	}

	if retrieved.LastRun != nil {
		t.Error("Expected LastRun to be nil")
	}

	if retrieved.NextRun != nil {
		t.Error("Expected NextRun to be nil")
	}
}

func TestDatabase_MultipleSources(t *testing.T) {
	db, cleanup := newTestDatabase(t)
	defer cleanup()

	// Create job with multiple sources
	sources := []string{
		"/path/to/source1",
		"/path/to/source2",
		"/path/to/source3",
	}

	job := &Job{
		ID:          "job-multi-src",
		Name:        "Multi Source Job",
		Type:        "file",
		Sources:     sources,
		Destination: "/dest",
	}

	if err := db.CreateJob(job); err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	// Retrieve and verify sources
	retrieved, err := db.GetJob("job-multi-src")
	if err != nil {
		t.Fatalf("Failed to get job: %v", err)
	}

	if len(retrieved.Sources) != len(sources) {
		t.Errorf("Expected %d sources, got %d", len(sources), len(retrieved.Sources))
	}

	for i, src := range sources {
		if retrieved.Sources[i] != src {
			t.Errorf("Source %d mismatch: expected %s, got %s", i, src, retrieved.Sources[i])
		}
	}
}

func BenchmarkDatabase_CreateJob(b *testing.B) {
	tempDir := b.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := NewDatabase(dbPath)
	if err != nil {
		b.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		job := &Job{
			ID:          filepath.Join("job", string(rune(i))),
			Name:        filepath.Join("Job", string(rune(i))),
			Type:        "file",
			Sources:     []string{"/source"},
			Destination: "/dest",
		}
		db.CreateJob(job)
	}
}

func BenchmarkDatabase_GetJob(b *testing.B) {
	tempDir := b.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := NewDatabase(dbPath)
	if err != nil {
		b.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Create a job first
	job := &Job{
		ID:   "bench-job",
		Name: "Bench Job",
		Type: "file",
	}
	db.CreateJob(job)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.GetJob("bench-job")
	}
}
