package database

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Setup test database
	testDBPath := "test_novabackup.db"
	os.Remove(testDBPath) // Clean up any existing test db

	code := m.Run()

	// Cleanup
	os.Remove(testDBPath)
	os.Exit(code)
}

func TestNewDatabase(t *testing.T) {
	dbPath := "test_new_database.db"
	defer os.Remove(dbPath)

	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	if db == nil {
		t.Fatal("Database should not be nil")
	}
}

func TestCreateJob(t *testing.T) {
	dbPath := "test_create_job.db"
	defer os.Remove(dbPath)

	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	job := &Job{
		Name:            "Test Job",
		Type:            "files",
		Sources:         []string{"C:\\test"},
		Destination:     "D:\\backup",
		Schedule:        "0 2 * * *",
		Enabled:         true,
		Deduplication:   true,
		Incremental:     true,
		RetentionDays:   30,
		RetentionCopies: 10,
		Encryption:      true,
		Compression:     "lz4",
	}

	id, err := db.CreateJob(job)
	if err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	if id == 0 {
		t.Fatal("Job ID should not be 0")
	}

	// Verify job was created
	jobs, err := db.ListJobs()
	if err != nil {
		t.Fatalf("Failed to list jobs: %v", err)
	}

	if len(jobs) != 1 {
		t.Fatalf("Expected 1 job, got %d", len(jobs))
	}

	if jobs[0].Name != "Test Job" {
		t.Errorf("Expected job name 'Test Job', got '%s'", jobs[0].Name)
	}
}

func TestListJobs(t *testing.T) {
	dbPath := "test_list_jobs.db"
	defer os.Remove(dbPath)

	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create multiple jobs
	for i := 1; i <= 3; i++ {
		job := &Job{
			Name:        "Job " + string(rune('0'+i)),
			Type:        "files",
			Sources:     []string{"C:\\test"},
			Destination: "D:\\backup",
			Enabled:     true,
		}
		_, err := db.CreateJob(job)
		if err != nil {
			t.Fatalf("Failed to create job %d: %v", i, err)
		}
	}

	jobs, err := db.ListJobs()
	if err != nil {
		t.Fatalf("Failed to list jobs: %v", err)
	}

	if len(jobs) != 3 {
		t.Errorf("Expected 3 jobs, got %d", len(jobs))
	}
}

func TestUpdateJob(t *testing.T) {
	dbPath := "test_update_job.db"
	defer os.Remove(dbPath)

	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	job := &Job{
		Name:        "Original Job",
		Type:        "files",
		Sources:     []string{"C:\\test"},
		Destination: "D:\\backup",
		Enabled:     true,
	}

	id, err := db.CreateJob(job)
	if err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	// Update job
	job.Name = "Updated Job"
	job.Enabled = false
	job.Deduplication = true

	err = db.UpdateJob(id, job)
	if err != nil {
		t.Fatalf("Failed to update job: %v", err)
	}

	// Verify update
	jobs, err := db.ListJobs()
	if err != nil {
		t.Fatalf("Failed to list jobs: %v", err)
	}

	if len(jobs) != 1 {
		t.Fatalf("Expected 1 job, got %d", len(jobs))
	}

	if jobs[0].Name != "Updated Job" {
		t.Errorf("Expected job name 'Updated Job', got '%s'", jobs[0].Name)
	}

	if jobs[0].Enabled != false {
		t.Errorf("Expected job to be disabled")
	}
}

func TestDeleteJob(t *testing.T) {
	dbPath := "test_delete_job.db"
	defer os.Remove(dbPath)

	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	job := &Job{
		Name:        "To Delete",
		Type:        "files",
		Sources:     []string{"C:\\test"},
		Destination: "D:\\backup",
		Enabled:     true,
	}

	id, err := db.CreateJob(job)
	if err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	err = db.DeleteJob(id)
	if err != nil {
		t.Fatalf("Failed to delete job: %v", err)
	}

	jobs, err := db.ListJobs()
	if err != nil {
		t.Fatalf("Failed to list jobs: %v", err)
	}

	if len(jobs) != 0 {
		t.Errorf("Expected 0 jobs after deletion, got %d", len(jobs))
	}
}

func TestGetJob(t *testing.T) {
	dbPath := "test_get_job.db"
	defer os.Remove(dbPath)

	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	job := &Job{
		Name:        "Get Test Job",
		Type:        "files",
		Sources:     []string{"C:\\test"},
		Destination: "D:\\backup",
		Enabled:     true,
	}

	id, err := db.CreateJob(job)
	if err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	retrievedJob, err := db.GetJob(id)
	if err != nil {
		t.Fatalf("Failed to get job: %v", err)
	}

	if retrievedJob.Name != "Get Test Job" {
		t.Errorf("Expected job name 'Get Test Job', got '%s'", retrievedJob.Name)
	}
}

func TestToggleJob(t *testing.T) {
	dbPath := "test_toggle_job.db"
	defer os.Remove(dbPath)

	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	job := &Job{
		Name:        "Toggle Test Job",
		Type:        "files",
		Sources:     []string{"C:\\test"},
		Destination: "D:\\backup",
		Enabled:     true,
	}

	id, err := db.CreateJob(job)
	if err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	// Toggle off
	err = db.ToggleJob(id)
	if err != nil {
		t.Fatalf("Failed to toggle job: %v", err)
	}

	retrievedJob, err := db.GetJob(id)
	if err != nil {
		t.Fatalf("Failed to get job: %v", err)
	}

	if retrievedJob.Enabled != false {
		t.Errorf("Expected job to be disabled after toggle")
	}

	// Toggle on
	err = db.ToggleJob(id)
	if err != nil {
		t.Fatalf("Failed to toggle job: %v", err)
	}

	retrievedJob, err = db.GetJob(id)
	if err != nil {
		t.Fatalf("Failed to get job: %v", err)
	}

	if retrievedJob.Enabled != true {
		t.Errorf("Expected job to be enabled after toggle")
	}
}

func TestJobWithAllFields(t *testing.T) {
	dbPath := "test_job_all_fields.db"
	defer os.Remove(dbPath)

	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	job := &Job{
		Name:            "Complete Job",
		Type:            "vm",
		Sources:         []string{"VM1", "VM2"},
		Destination:     "D:\\vm-backup",
		Schedule:        "0 */6 * * *",
		Enabled:         true,
		Deduplication:   true,
		Incremental:     true,
		RetentionDays:   60,
		RetentionCopies: 20,
		Encryption:      true,
		Compression:     "zstd",
		Description:     "Test job with all fields",
	}

	id, err := db.CreateJob(job)
	if err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	retrievedJob, err := db.GetJob(id)
	if err != nil {
		t.Fatalf("Failed to get job: %v", err)
	}

	if retrievedJob.Type != "vm" {
		t.Errorf("Expected job type 'vm', got '%s'", retrievedJob.Type)
	}

	if retrievedJob.Deduplication != true {
		t.Errorf("Expected deduplication to be true")
	}

	if retrievedJob.Incremental != true {
		t.Errorf("Expected incremental to be true")
	}

	if retrievedJob.RetentionDays != 60 {
		t.Errorf("Expected retention days 60, got %d", retrievedJob.RetentionDays)
	}

	if retrievedJob.RetentionCopies != 20 {
		t.Errorf("Expected retention copies 20, got %d", retrievedJob.RetentionCopies)
	}
}

func TestDatabaseClose(t *testing.T) {
	dbPath := "test_db_close.db"
	defer os.Remove(dbPath)

	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	err = db.Close()
	if err != nil {
		t.Errorf("Failed to close database: %v", err)
	}
}
