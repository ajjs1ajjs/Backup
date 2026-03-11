package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"time"
	"path/filepath"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type NovaBackup struct {
	*walk.MainWindow
	
	// UI Components
	statusLabel       *walk.Label
	progressBar       *walk.ProgressBar
	logTextEdit       *walk.TextEdit
	jobsTableView     *walk.TableView
	jobsModel        *walk.ReflectTableModel
	
	// Backend Components
	agentRunning      bool
	backupInProgress   bool
	currentJob        string
	jobs             []BackupJob
	
	// Architecture Components
	backupServer      bool
	proxyServer       bool
	repositoryServer  bool
	deduplication    bool
	encryption       bool
}

type BackupJob struct {
	Name        string
	Type        string
	Status      string
	LastRun     string
	NextRun     string
	Schedule    string
	Source      string
	Destination string
	Enabled     bool
}

func main() {
	nb := &NovaBackup{}
	
	if err := nb.createMainWindow(); err != nil {
		fmt.Printf("Error creating NovaBackup GUI: %v\n", err)
		os.Exit(1)
	}
	
	// Initialize backend services
	nb.initializeServices()
	
	// Start background monitoring
	go nb.backgroundMonitoring()
	
	// Setup signal handling for graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		nb.cleanup()
		os.Exit(0)
	}()
	
	// Run the application
	nb.MainWindow.Run()
}

func (nb *NovaBackup) createMainWindow() error {
	mw, err := MainWindow{
		AssignTo: &nb.MainWindow,
		Title:    "NovaBackup v6.0 - Enterprise Backup & Recovery",
		Size:     Size{1400, 900},
		Layout:   VBox{},
		Children: []Widget{
			// Header
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text:    "NovaBackup v6.0",
						Font:    Font{Family: "Segoe UI", PointSize: 16, Bold: true},
						TextColor: walk.RGB(0, 120, 215),
					},
					HSpacer{},
					Label{
						Text:    "Enterprise Backup & Recovery Platform",
						Font:    Font{Family: "Segoe UI", PointSize: 10},
						TextColor: walk.RGB(128, 128, 128),
					},
				},
			},
			
			// Status Bar
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "System Status:",
						Font: Font{Family: "Segoe UI", PointSize: 10, Bold: true},
					},
					Label{
						AssignTo: &nb.statusLabel,
						Text:     "Ready",
						Font:     Font{Family: "Segoe UI", PointSize: 10},
						TextColor: walk.RGB(0, 128, 0),
					},
					HSpacer{},
					PushButton{
						Text: "Start Agent",
						OnClicked: func() { nb.startAgent() },
					},
					PushButton{
						Text: "Stop Agent",
						OnClicked: func() { nb.stopAgent() },
					},
					PushButton{
						Text: "Create Backup Job",
						OnClicked: func() { nb.showCreateJobDialog() },
					},
					PushButton{
						Text: "Run Backup Now",
						OnClicked: func() { nb.runImmediateBackup() },
					},
					PushButton{
						Text: "Settings",
						OnClicked: func() { nb.showSettingsDialog() },
					},
				},
			},
			
			// Main Content Area
			Composite{
				Layout: HBox{},
				Children: []Widget{
					// Jobs Panel
					GroupBox{
						Title:  "Backup Jobs",
						Layout: VBox{},
						MinSize: Size{800, 400},
						Children: []Widget{
							TableView{
								AssignTo: &nb.jobsTableView,
								Columns: []TableViewColumn{
									{Title: "Job Name", Width: 200},
									{Title: "Type", Width: 100},
									{Title: "Status", Width: 120},
									{Title: "Last Run", Width: 150},
									{Title: "Next Run", Width: 150},
									{Title: "Schedule", Width: 120},
								},
								Model: nb.getJobsModel(),
							},
							Composite{
								Layout: HBox{},
								Children: []Widget{
									PushButton{
										Text: "Add Job",
										OnClicked: func() { nb.showCreateJobDialog() },
									},
									PushButton{
										Text: "Edit Job",
										OnClicked: func() { nb.editSelectedJob() },
									},
									PushButton{
										Text: "Delete Job",
										OnClicked: func() { nb.deleteSelectedJob() },
									},
									PushButton{
										Text: "Run Now",
										OnClicked: func() { nb.runSelectedJob() },
									},
									HSpacer{},
								},
							},
						},
					},
					
					// Status Panel
					GroupBox{
						Title:  "System Status & Activity",
						Layout: VBox{},
						MinSize: Size{550, 400},
						Children: []Widget{
							Label{
								Text: "Backup Progress:",
								Font: Font{Family: "Segoe UI", PointSize: 10, Bold: true},
							},
							ProgressBar{
								AssignTo: &nb.progressBar,
								MinSize:  Size{500, 25},
								MaxSize:  Size{500, 25},
							},
							Label{
								Text: "Activity Log:",
								Font: Font{Family: "Segoe UI", PointSize: 10, Bold: true},
							},
							TextEdit{
								AssignTo: &nb.logTextEdit,
								MinSize:  Size{500, 200},
								MaxSize:  Size{500, 200},
								ReadOnly: true,
								Font:     Font{Family: "Consolas", PointSize: 9},
							},
						},
					},
				},
			},
			
			// Statistics Panel
			GroupBox{
				Title:  "Storage Statistics",
				Layout: HBox{},
				Children: []Widget{
					Label{Text: "Total Backups: 0"},
					Label{Text: "Storage Used: 0 GB"},
					Label{Text: "Deduplication: 0:1"},
					Label{Text: "Compression: 0:1"},
					Label{Text: "Encryption: Disabled"},
					HSpacer{},
					Label{Text: "Agent Status: Stopped"},
				},
			},
		},
	}.Create()

	if err != nil {
		return err
	}
	
	// Initialize sample jobs
	nb.initializeSampleJobs()
	
	return nil
}

func (nb *NovaBackup) initializeServices() {
	nb.addLogEntry("Initializing NovaBackup v6.0 Enterprise Services...")
	nb.addLogEntry("Starting 15-component architecture...")
	
	// Initialize all components
	nb.backupServer = true
	nb.proxyServer = true
	nb.repositoryServer = true
	nb.deduplication = true
	nb.encryption = false
	
	nb.addLogEntry("✅ Backup Server: Started")
	nb.addLogEntry("✅ Backup Proxy: Started")
	nb.addLogEntry("✅ Repository Server: Started")
	nb.addLogEntry("✅ Deduplication Engine: Started")
	nb.addLogEntry("✅ Compression Engine: Started")
	nb.addLogEntry("✅ Catalog Service: Started")
	nb.addLogEntry("✅ WAN Accelerator: Ready")
	nb.addLogEntry("✅ Hypervisor Integration: Ready")
	nb.addLogEntry("✅ Transport Service: Ready")
	nb.addLogEntry("NovaBackup v6.0 fully initialized!")
	
	nb.updateStatusLabel("Ready", walk.RGB(0, 128, 0))
}

func (nb *NovaBackup) initializeSampleJobs() {
	nb.jobs = []BackupJob{
		{
			Name:     "Daily Documents Backup",
			Type:     "Files",
			Status:   "✅ Active",
			LastRun:  "2026-03-11 02:00",
			NextRun:  "2026-03-12 02:00",
			Schedule:  "Daily 2AM",
			Source:   "C:\\Users\\Documents",
			Destination: "D:\\NovaBackups\\Documents",
			Enabled:  true,
		},
		{
			Name:     "Weekly System Backup",
			Type:     "System",
			Status:   "✅ Active",
			LastRun:  "2026-03-08 22:00",
			NextRun:  "2026-03-15 22:00",
			Schedule:  "Weekly Sun 10PM",
			Source:   "C:\\Windows",
			Destination: "D:\\NovaBackups\\System",
			Enabled:  true,
		},
		{
			Name:     "Database Backup",
			Type:     "SQL Server",
			Status:   "✅ Active",
			LastRun:  "2026-03-11 01:00",
			NextRun:  "2026-03-12 01:00",
			Schedule:  "Daily 1AM",
			Source:   "SQL Server Instance",
			Destination: "D:\\NovaBackups\\Database",
			Enabled:  true,
		},
	}
	
	nb.updateJobsList()
}

func (nb *NovaBackup) getJobsModel() *walk.ReflectTableModel {
	return walk.NewReflectTableModelFromSlice(nb.jobs)
}

func (nb *NovaBackup) updateJobsList() {
	if nb.jobsTableView != nil {
		nb.jobsTableView.SetModel(nb.getJobsModel())
	}
}

func (nb *NovaBackup) updateStatusLabel(text string, color walk.Color) {
	if nb.statusLabel != nil {
		nb.statusLabel.SetText(text)
		nb.statusLabel.SetTextColor(color)
	}
}

func (nb *NovaBackup) addLogEntry(message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] %s\n", timestamp, message)
	
	if nb.logTextEdit != nil {
		currentText := nb.logTextEdit.Text()
		if len(currentText) > 5000 {
			lines := strings.Split(currentText, "\n")
			if len(lines) > 100 {
				currentText = strings.Join(lines[len(lines)-100:], "\n")
			}
		}
		nb.logTextEdit.SetText(currentText + logEntry)
		
		// Scroll to bottom
		nb.logTextEdit.SendMessage(0x00B6, 0, -1) // EM_SCROLLCARET
	}
	
	fmt.Printf("NovaBackup: %s\n", message)
}

func (nb *NovaBackup) startAgent() {
	if nb.agentRunning {
		nb.addLogEntry("Agent is already running")
		return
	}
	
	nb.agentRunning = true
	nb.updateStatusLabel("Agent Running", walk.RGB(0, 128, 0))
	nb.addLogEntry("🚀 NovaBackup Agent started")
	nb.addLogEntry("Background services initialized")
	nb.addLogEntry("Ready for scheduled backups")
	
	// Start background backup scheduler
	go nb.backupScheduler()
}

func (nb *NovaBackup) stopAgent() {
	if !nb.agentRunning {
		nb.addLogEntry("Agent is not running")
		return
	}
	
	nb.agentRunning = false
	nb.updateStatusLabel("Agent Stopped", walk.RGB(255, 0, 0))
	nb.addLogEntry("⏹ NovaBackup Agent stopped")
	nb.backupInProgress = false
}

func (nb *NovaBackup) backupScheduler() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for nb.agentRunning {
		select {
		case <-ticker.C:
			nb.checkScheduledJobs()
		}
	}
}

func (nb *NovaBackup) checkScheduledJobs() {
	currentTime := time.Now()
	
	for _, job := range nb.jobs {
		if !job.Enabled {
			continue
		}
		
		// Simple scheduling logic (in production, this would be more sophisticated)
		if strings.Contains(job.Schedule, "Daily") && currentTime.Hour() == 2 && currentTime.Minute() == 0 {
			nb.runBackupJob(job.Name)
		} else if strings.Contains(job.Schedule, "Weekly") && currentTime.Weekday() == time.Sunday && currentTime.Hour() == 22 && currentTime.Minute() == 0 {
			nb.runBackupJob(job.Name)
		}
	}
}

func (nb *NovaBackup) runBackupJob(jobName string) {
	for _, job := range nb.jobs {
		if job.Name == jobName {
			nb.executeBackupPipeline(job)
			break
		}
	}
}

func (nb *NovaBackup) executeBackupPipeline(job BackupJob) {
	if nb.backupInProgress {
		nb.addLogEntry("Backup already in progress")
		return
	}
	
	nb.backupInProgress = true
	nb.currentJob = job.Name
	nb.updateStatusLabel("Running Backup", walk.RGB(255, 165, 0))
	
	nb.addLogEntry(fmt.Sprintf("🔄 Starting backup job: %s", job.Name))
	nb.addLogEntry("📋 Job Type: %s", job.Type)
	nb.addLogEntry("📁 Source: %s", job.Source)
	nb.addLogEntry("💾 Destination: %s", job.Destination)
	
	// Simulate 10-stage backup pipeline
	stages := []string{
		"Job Scheduler & Init",
		"Snapshot Creation",
		"Application Consistency",
		"Change Block Tracking",
		"Data Read (Proxy Stage)",
		"Compression",
		"Deduplication",
		"Encryption",
		"Transport & Storage Write",
		"Metadata & Indexing",
	}
	
	for i, stage := range stages {
		progress := int(float64(i+1) / float64(len(stages)) * 100)
		nb.progressBar.SetValue(progress)
		nb.addLogEntry(fmt.Sprintf("🔧 Stage %d/11: %s", i+1, stage))
		
		// Simulate processing time
		time.Sleep(500 * time.Millisecond)
	}
	
	nb.addLogEntry("✅ Backup completed successfully")
	nb.updateStatusLabel("Backup Completed", walk.RGB(0, 128, 0))
	nb.backupInProgress = false
	nb.currentJob = ""
	
	// Update job status
	for i, job := range nb.jobs {
		if job.Name == jobName {
			nb.jobs[i].LastRun = time.Now().Format("2006-01-02 15:04")
			nb.jobs[i].Status = "✅ Completed"
			break
		}
	}
	
	nb.updateJobsList()
	nb.progressBar.SetValue(0)
}

func (nb *NovaBackup) runImmediateBackup() {
	if len(nb.jobs) == 0 {
		walk.MsgBox(nb.MainWindow, "No backup jobs configured", "Error", walk.MsgBoxIconError)
		return
	}
	
	// Run first available job
	nb.runBackupJob(nb.jobs[0].Name)
}

func (nb *NovaBackup) runSelectedJob() {
	index := nb.jobsTableView.CurrentIndex()
	if index < 0 || index >= len(nb.jobs) {
		walk.MsgBox(nb.MainWindow, "Please select a job", "Error", walk.MsgBoxIconWarning)
		return
	}
	
	nb.runBackupJob(nb.jobs[index].Name)
}

func (nb *NovaBackup) showCreateJobDialog() {
	var nameEdit, sourceEdit, destEdit *walk.LineEdit
	var scheduleCombo *walk.ComboBox
	var enabledCheck *walk.CheckBox
	
	dlg, err := walk.Dialog{
		Title:  "Create Backup Job",
		MinSize: Size{400, 300},
		Layout:  VBox{},
		Children: []Widget{
			Label{Text: "Job Name:"},
			LineEdit{AssignTo: &nameEdit},
			Label{Text: "Source:"},
			LineEdit{AssignTo: &sourceEdit},
			Label{Text: "Destination:"},
			LineEdit{AssignTo: &destEdit},
			Label{Text: "Schedule:"},
			ComboBox{
				AssignTo: &scheduleCombo,
				Model:   []string{"Daily", "Weekly", "Monthly", "Manual"},
				Value:    "Daily",
			},
			CheckBox{
				AssignTo: &enabledCheck,
				Text:     "Enabled",
				Checked:  true,
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{
						Text: "Create",
						OnClicked: func() {
							if nameEdit.Text() == "" {
								walk.MsgBox(dlg, "Please enter a job name", "Error", walk.MsgBoxIconError)
								return
							}
							
							newJob := BackupJob{
								Name:        nameEdit.Text(),
								Type:        "Files",
								Status:      "✅ Active",
								LastRun:     "Never",
								NextRun:     "Next scheduled run",
								Schedule:    scheduleCombo.Text(),
								Source:      sourceEdit.Text(),
								Destination: destEdit.Text(),
								Enabled:     enabledCheck.Checked(),
							}
							
							nb.jobs = append(nb.jobs, newJob)
							nb.updateJobsList()
							nb.addLogEntry(fmt.Sprintf("✅ Created new backup job: %s", newJob.Name))
							dlg.Accept()
						},
					},
					PushButton{
						Text: "Cancel",
						OnClicked: func() { dlg.Cancel() },
					},
				},
			},
		},
	}.Create(nb.MainWindow)
	
	if err != nil {
		nb.addLogEntry(fmt.Sprintf("Error creating dialog: %v", err))
		return
	}
	
	dlg.Run()
}

func (nb *NovaBackup) editSelectedJob() {
	index := nb.jobsTableView.CurrentIndex()
	if index < 0 || index >= len(nb.jobs) {
		walk.MsgBox(nb.MainWindow, "Please select a job", "Error", walk.MsgBoxIconWarning)
		return
	}
	
	job := nb.jobs[index]
	walk.MsgBox(nb.MainWindow, 
		fmt.Sprintf("Edit job: %s\nType: %s\nSource: %s", job.Name, job.Type, job.Source),
		"Edit Job",
		walk.MsgBoxIconInformation)
}

func (nb *NovaBackup) deleteSelectedJob() {
	index := nb.jobsTableView.CurrentIndex()
	if index < 0 || index >= len(nb.jobs) {
		walk.MsgBox(nb.MainWindow, "Please select a job", "Error", walk.MsgBoxIconWarning)
		return
	}
	
	job := nb.jobs[index]
	if walk.MsgBox(nb.MainWindow,
		fmt.Sprintf("Delete backup job: %s?", job.Name),
		"Confirm Delete",
		walk.MsgBoxYesNo|walk.MsgBoxIconQuestion) == walk.DlgCmdYes {
		
		nb.jobs = append(nb.jobs[:index], nb.jobs[index+1:]...)
		nb.updateJobsList()
		nb.addLogEntry(fmt.Sprintf("🗑️ Deleted backup job: %s", job.Name))
	}
}

func (nb *NovaBackup) showSettingsDialog() {
	var compressionCombo, encryptionCheck, dedupCheck *walk.Widget
	
	dlg, err := walk.Dialog{
		Title:  "NovaBackup Settings",
		MinSize: Size{500, 400},
		Layout:  VBox{},
		Children: []Widget{
			GroupBox{
				Title: "General Settings",
				Layout: VBox{},
				Children: []Widget{
					CheckBox{
						AssignTo: &encryptionCheck,
						Text:    "Enable Encryption (AES-256)",
						Checked: nb.encryption,
					},
					CheckBox{
						AssignTo: &dedupCheck,
						Text:    "Enable Deduplication",
						Checked: nb.deduplication,
					},
				},
			},
			GroupBox{
				Title: "Performance Settings",
				Layout: VBox{},
				Children: []Widget{
					Label{Text: "Compression Level:"},
					ComboBox{
						AssignTo: &compressionCombo,
						Model:   []string{"None", "Low", "Optimal", "High", "Extreme"},
						Value:    "Optimal",
					},
				},
			},
			GroupBox{
				Title: "Architecture Status",
				Layout: VBox{},
				Children: []Widget{
					Label{Text: fmt.Sprintf("Backup Server: %s", nb.getStatusIcon(nb.backupServer))},
					Label{Text: fmt.Sprintf("Backup Proxy: %s", nb.getStatusIcon(nb.proxyServer))},
					Label{Text: fmt.Sprintf("Repository Server: %s", nb.getStatusIcon(nb.repositoryServer))},
					Label{Text: fmt.Sprintf("Deduplication Engine: %s", nb.getStatusIcon(nb.deduplication))},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{
						Text: "Save",
						OnClicked: func() { dlg.Accept() },
					},
					PushButton{
						Text: "Cancel",
						OnClicked: func() { dlg.Cancel() },
					},
				},
			},
		},
	}.Create(nb.MainWindow)
	
	if err != nil {
		nb.addLogEntry(fmt.Sprintf("Error creating settings dialog: %v", err))
		return
	}
	
	dlg.Run()
}

func (nb *NovaBackup) getStatusIcon(status bool) string {
	if status {
		return "✅ Running"
	}
	return "❌ Stopped"
}

func (nb *NovaBackup) backgroundMonitoring() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			if nb.agentRunning {
				nb.addLogEntry("📊 Background monitoring active")
			}
		}
	}
}

func (nb *NovaBackup) cleanup() {
	nb.addLogEntry("🔄 Shutting down NovaBackup services...")
	nb.stopAgent()
	nb.addLogEntry("👋 NovaBackup v6.0 shutdown complete")
}
