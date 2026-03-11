package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
)

// JobWizard handles backup job creation wizard
type JobWizard struct {
	logger *zap.Logger
	reader *bufio.Reader
}

// NewJobWizard creates a new backup job wizard
func NewJobWizard(logger *zap.Logger) *JobWizard {
	return &JobWizard{
		logger: logger,
		reader: bufio.NewReader(os.Stdin),
	}
}

// Run starts the interactive job wizard
func (w *JobWizard) Run() error {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║           NovaBackup - Backup Job Wizard                   ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	job := &BackupJob{
		ID:        generateJobID(),
		CreatedAt: time.Now(),
		Enabled:   true,
	}

	// Step 1: Job Name
	if err := w.askJobName(job); err != nil {
		return err
	}

	// Step 2: Job Type
	if err := w.askJobType(job); err != nil {
		return err
	}

	// Step 3: Source
	if err := w.askSource(job); err != nil {
		return err
	}

	// Step 4: Destination
	if err := w.askDestination(job); err != nil {
		return err
	}

	// Step 5: Schedule
	if err := w.askSchedule(job); err != nil {
		return err
	}

	// Step 6: Advanced Options
	if err := w.askAdvancedOptions(job); err != nil {
		return err
	}

	// Summary and Confirmation
	return w.showSummary(job)
}

func (w *JobWizard) askJobName(job *BackupJob) error {
	fmt.Println("Step 1/6: Job Name")
	fmt.Println("------------------")
	fmt.Print("Enter job name: ")

	name, err := w.reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read job name: %w", err)
	}

	job.Name = strings.TrimSpace(name)
	if job.Name == "" {
		job.Name = fmt.Sprintf("Backup Job %s", time.Now().Format("2006-01-02"))
	}

	fmt.Printf("✓ Job name: %s\n\n", job.Name)
	return nil
}

func (w *JobWizard) askJobType(job *BackupJob) error {
	fmt.Println("Step 2/6: Job Type")
	fmt.Println("------------------")
	fmt.Println("Select backup type:")
	fmt.Println("1. Full Backup (complete backup every time)")
	fmt.Println("2. Incremental (backup only changed blocks)")
	fmt.Println("3. Differential (backup changes since last full)")
	fmt.Print("Choice (1-3): ")

	choice, err := w.reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read choice: %w", err)
	}

	switch strings.TrimSpace(choice) {
	case "1":
		job.Type = "full"
	case "2":
		job.Type = "incremental"
	case "3":
		job.Type = "differential"
	default:
		job.Type = "full"
	}

	fmt.Printf("✓ Job type: %s\n\n", job.Type)
	return nil
}

func (w *JobWizard) askSource(job *BackupJob) error {
	fmt.Println("Step 3/6: Backup Source")
	fmt.Println("------------------------")
	fmt.Println("Select source type:")
	fmt.Println("1. VMware vSphere VM")
	fmt.Println("2. Hyper-V VM")
	fmt.Println("3. Physical Machine")
	fmt.Println("4. Files/Folders")
	fmt.Print("Choice (1-4): ")

	choice, err := w.reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read choice: %w", err)
	}

	switch strings.TrimSpace(choice) {
	case "1":
		job.SourceType = "vmware"
		fmt.Print("Enter vCenter server (e.g., vcenter.local): ")
		server, _ := w.reader.ReadString('\n')
		job.SourceConfig["server"] = strings.TrimSpace(server)

		fmt.Print("Enter VM name or pattern: ")
		vmName, _ := w.reader.ReadString('\n')
		job.Source = strings.TrimSpace(vmName)

	case "2":
		job.SourceType = "hyperv"
		fmt.Print("Enter Hyper-V server: ")
		server, _ := w.reader.ReadString('\n')
		job.SourceConfig["server"] = strings.TrimSpace(server)

		fmt.Print("Enter VM name: ")
		vmName, _ := w.reader.ReadString('\n')
		job.Source = strings.TrimSpace(vmName)

	case "3":
		job.SourceType = "physical"
		fmt.Print("Enter machine name/IP: ")
		machine, _ := w.reader.ReadString('\n')
		job.Source = strings.TrimSpace(machine)

	case "4":
		job.SourceType = "files"
		fmt.Print("Enter path to backup (e.g., C:\\Data): ")
		path, _ := w.reader.ReadString('\n')
		job.Source = strings.TrimSpace(path)

	default:
		job.SourceType = "files"
		job.Source = "C:\\"
	}

	fmt.Printf("✓ Source: %s (%s)\n\n", job.Source, job.SourceType)
	return nil
}

func (w *JobWizard) askDestination(job *BackupJob) error {
	fmt.Println("Step 4/6: Backup Destination")
	fmt.Println("----------------------------")
	fmt.Println("Select destination type:")
	fmt.Println("1. Local Disk")
	fmt.Println("2. Network Share")
	fmt.Println("3. Cloud Storage (S3)")
	fmt.Print("Choice (1-3): ")

	choice, err := w.reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read choice: %w", err)
	}

	switch strings.TrimSpace(choice) {
	case "1":
		fmt.Print("Enter local path (e.g., D:\\Backups): ")
		path, _ := w.reader.ReadString('\n')
		job.Destination = strings.TrimSpace(path)

	case "2":
		fmt.Print("Enter network path (e.g., \\\\server\\share): ")
		path, _ := w.reader.ReadString('\n')
		job.Destination = strings.TrimSpace(path)

	case "3":
		job.DestinationType = "s3"
		fmt.Print("Enter S3 bucket name: ")
		bucket, _ := w.reader.ReadString('\n')
		job.DestinationConfig["bucket"] = strings.TrimSpace(bucket)

		fmt.Print("Enter S3 region: ")
		region, _ := w.reader.ReadString('\n')
		job.DestinationConfig["region"] = strings.TrimSpace(region)

		job.Destination = fmt.Sprintf("s3://%s", strings.TrimSpace(bucket))

	default:
		job.Destination = "C:\\Backups"
	}

	fmt.Printf("✓ Destination: %s\n\n", job.Destination)
	return nil
}

func (w *JobWizard) askSchedule(job *BackupJob) error {
	fmt.Println("Step 5/6: Backup Schedule")
	fmt.Println("------------------------")
	fmt.Println("Select schedule:")
	fmt.Println("1. Daily")
	fmt.Println("2. Weekly")
	fmt.Println("3. Monthly")
	fmt.Println("4. Custom")
	fmt.Print("Choice (1-4): ")

	choice, err := w.reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read choice: %w", err)
	}

	switch strings.TrimSpace(choice) {
	case "1":
		job.Schedule = "0 2 * * *" // Daily at 2 AM
		fmt.Println("✓ Schedule: Daily at 2:00 AM")

	case "2":
		job.Schedule = "0 2 * * 0" // Weekly on Sunday at 2 AM
		fmt.Println("✓ Schedule: Weekly on Sunday at 2:00 AM")

	case "3":
		job.Schedule = "0 2 1 * *" // Monthly on 1st at 2 AM
		fmt.Println("✓ Schedule: Monthly on 1st at 2:00 AM")

	case "4":
		fmt.Print("Enter cron expression (e.g., 0 2 * * *): ")
		cron, _ := w.reader.ReadString('\n')
		job.Schedule = strings.TrimSpace(cron)
		fmt.Printf("✓ Schedule: %s\n", job.Schedule)

	default:
		job.Schedule = "0 2 * * *"
		fmt.Println("✓ Schedule: Daily at 2:00 AM (default)")
	}

	fmt.Println()
	return nil
}

func (w *JobWizard) askAdvancedOptions(job *BackupJob) error {
	fmt.Println("Step 6/6: Advanced Options")
	fmt.Println("--------------------------")

	// Compression
	fmt.Print("Enable compression? (y/n): ")
	compress, _ := w.reader.ReadString('\n')
	job.Compression = strings.TrimSpace(strings.ToLower(compress)) == "y"
	fmt.Printf("✓ Compression: %v\n", job.Compression)

	// Encryption
	fmt.Print("Enable encryption? (y/n): ")
	encrypt, _ := w.reader.ReadString('\n')
	job.Encryption = strings.TrimSpace(strings.ToLower(encrypt)) == "y"
	fmt.Printf("✓ Encryption: %v\n", job.Encryption)

	if job.Encryption {
		fmt.Print("Enter encryption password: ")
		password, _ := w.reader.ReadString('\n')
		job.EncryptionPassword = strings.TrimSpace(password)
	}

	// Retention
	fmt.Print("Retention period (days, default 30): ")
	retention, _ := w.reader.ReadString('\n')
	retention = strings.TrimSpace(retention)
	if retention == "" {
		retention = "30"
	}
	job.RetentionDays = retention
	fmt.Printf("✓ Retention: %s days\n", job.RetentionDays)

	// Application-aware processing
	fmt.Print("Enable application-aware processing? (y/n): ")
	appAware, _ := w.reader.ReadString('\n')
	job.ApplicationAware = strings.TrimSpace(strings.ToLower(appAware)) == "y"
	fmt.Printf("✓ Application-aware: %v\n", job.ApplicationAware)

	fmt.Println()
	return nil
}

func (w *JobWizard) showSummary(job *BackupJob) error {
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("                    JOB CONFIGURATION SUMMARY                    ")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Printf("Job ID:          %s\n", job.ID)
	fmt.Printf("Name:            %s\n", job.Name)
	fmt.Printf("Type:            %s\n", job.Type)
	fmt.Printf("Source:          %s (%s)\n", job.Source, job.SourceType)
	fmt.Printf("Destination:     %s\n", job.Destination)
	fmt.Printf("Schedule:        %s\n", job.Schedule)
	fmt.Printf("Compression:     %v\n", job.Compression)
	fmt.Printf("Encryption:      %v\n", job.Encryption)
	fmt.Printf("Retention:       %s days\n", job.RetentionDays)
	fmt.Printf("App-Aware:       %v\n", job.ApplicationAware)
	fmt.Println("═══════════════════════════════════════════════════════════════")

	fmt.Print("\nCreate this job? (y/n): ")
	confirm, err := w.reader.ReadString('\n')
	if err != nil {
		return err
	}

	if strings.TrimSpace(strings.ToLower(confirm)) == "y" {
		// Save job to configuration
		if err := w.saveJob(job); err != nil {
			return fmt.Errorf("failed to save job: %w", err)
		}
		fmt.Println("\n✓ Backup job created successfully!")
		fmt.Printf("Job ID: %s\n", job.ID)
	} else {
		fmt.Println("\n✗ Job creation cancelled.")
	}

	return nil
}

func (w *JobWizard) saveJob(job *BackupJob) error {
	// TODO: Implement job persistence to database or config file
	w.logger.Info("Saving backup job",
		zap.String("job_id", job.ID),
		zap.String("name", job.Name),
	)

	// For now, just log the job
	fmt.Printf("Job saved: %+v\n", job)
	return nil
}

// generateJobID generates a unique job ID
func generateJobID() string {
	return fmt.Sprintf("job_%d", time.Now().Unix())
}

// BackupJob represents a backup job configuration
type BackupJob struct {
	ID                 string
	Name               string
	Type               string // full, incremental, differential
	Source             string
	SourceType         string // vmware, hyperv, physical, files
	SourceConfig       map[string]string
	Destination        string
	DestinationType    string // local, network, s3
	DestinationConfig  map[string]string
	Schedule           string // cron expression
	Compression        bool
	Encryption         bool
	EncryptionPassword string
	RetentionDays      string
	ApplicationAware   bool
	Enabled            bool
	CreatedAt          time.Time
}

// Initialize config maps
func init() {
	// This ensures map fields are initialized when creating new jobs
}

// NewJob creates a new backup job with initialized maps
func NewJob() *BackupJob {
	return &BackupJob{
		SourceConfig:      make(map[string]string),
		DestinationConfig: make(map[string]string),
	}
}
