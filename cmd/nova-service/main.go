// NovaBackup - Minimal Windows Service with Real Backup Functionality
package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

const (
	serviceName = "NovaBackup"
	defaultPort = 8080
)

type Service struct {
	eventLog debug.Log
	stopChan chan struct{}
	dataDir  string
}

func NewService(eventLog debug.Log) *Service {
	exePath, _ := os.Executable()
	dataDir := filepath.Join(filepath.Dir(exePath), "..", "..", "ProgramData", "NovaBackup")

	return &Service{
		eventLog: eventLog,
		stopChan: make(chan struct{}),
		dataDir:  dataDir,
	}
}

func (s *Service) logEvent(msg string) {
	if s.eventLog != nil {
		s.eventLog.Info(1, msg)
	}
	log.Println(msg)
}

func (s *Service) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (bool, uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown

	changes <- svc.Status{State: svc.StartPending}
	s.logEvent("NovaBackup service starting...")

	if err := s.startAPI(); err != nil {
		s.logEvent(fmt.Sprintf("Failed to start API: %v", err))
		return true, 1
	}

	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
	s.logEvent(fmt.Sprintf("NovaBackup service started on port %d", defaultPort))

loop:
	for {
		select {
		case <-s.stopChan:
			break loop
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				s.logEvent("NovaBackup service stopping...")
				changes <- svc.Status{State: svc.StopPending}
				break loop
			}
		}
	}

	return false, 0
}

func (s *Service) startAPI() error {
	// Create directories
	os.MkdirAll(filepath.Join(s.dataDir, "backups"), 0755)
	os.MkdirAll(filepath.Join(s.dataDir, "logs"), 0755)
	os.MkdirAll(filepath.Join(s.dataDir, "config"), 0755)
	os.MkdirAll(filepath.Join(s.dataDir, "sessions"), 0755)

	// Setup HTTP handlers
	http.HandleFunc("/api/backup/run", s.handleBackupRun)
	http.HandleFunc("/api/backup/sessions", s.handleSessions)
	http.HandleFunc("/api/restore/files", s.handleRestoreFiles)
	http.HandleFunc("/api/jobs", s.handleJobs)
	http.HandleFunc("/api/health", s.handleHealth)

	go func() {
		addr := fmt.Sprintf(":%d", defaultPort)
		log.Printf("Starting API server on %s", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Printf("API server error: %v", err)
		}
	}()

	return nil
}

// API Handlers

func (s *Service) handleHealth(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "healthy",
		"version": "6.0.0",
		"time":    time.Now().Format(time.RFC3339),
	})
}

func (s *Service) handleBackupRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name         string   `json:"name"`
		Type         string   `json:"type"` // file, database
		Sources      []string `json:"sources"`
		Destination  string   `json:"destination"`
		Compression  bool     `json:"compression"`
		DatabaseType string   `json:"database_type,omitempty"`
		DatabaseConn string   `json:"database_conn,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create session
	sessionID := fmt.Sprintf("session_%d", time.Now().Unix())
	backupDir := filepath.Join(req.Destination, req.Name, time.Now().Format("2006-01-02_150405"))
	os.MkdirAll(backupDir, 0755)

	session := map[string]interface{}{
		"id":          sessionID,
		"job_name":    req.Name,
		"start_time":  time.Now().Format(time.RFC3339),
		"status":      "running",
		"backup_path": backupDir,
	}

	// Run backup in background
	go func() {
		var err error
		var filesProcessed int
		var bytesWritten int64

		if req.Type == "database" {
			// Database backup
			err = s.backupDatabase(req.DatabaseType, req.DatabaseConn, backupDir, req.Compression)
		} else {
			// File backup
			filesProcessed, bytesWritten, err = s.backupFiles(req.Sources, backupDir, req.Compression)
		}

		session["end_time"] = time.Now().Format(time.RFC3339)
		session["files_processed"] = filesProcessed
		session["bytes_written"] = bytesWritten

		if err != nil {
			session["status"] = "failed"
			session["error"] = err.Error()
			s.saveSession(sessionID, session)
			log.Printf("Backup failed: %v", err)
			return
		}

		session["status"] = "success"
		s.saveSession(sessionID, session)
		log.Printf("Backup completed: %s", sessionID)
	}()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"session": session,
		"message": "Backup started",
	})
}

func (s *Service) backupFiles(sources []string, backupDir string, compression bool) (int, int64, error) {
	log.Printf("Starting file backup to %s", backupDir)

	var files []string
	var totalSize int64

	// Collect files
	for _, source := range sources {
		filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				files = append(files, path)
				totalSize += info.Size()
			}
			return nil
		})
	}

	log.Printf("Found %d files (%d bytes)", len(files), totalSize)

	// Create zip archive
	archivePath := filepath.Join(backupDir, "backup.zip")
	archiveFile, err := os.Create(archivePath)
	if err != nil {
		return 0, 0, err
	}
	defer archiveFile.Close()

	zipWriter := zip.NewWriter(archiveFile)
	defer zipWriter.Close()

	// Add files
	for _, file := range files {
		if err := s.addFileToZip(zipWriter, file); err != nil {
			log.Printf("Warning: Failed to add %s: %v", file, err)
		}
	}

	archiveInfo, _ := archiveFile.Stat()

	// Save metadata
	metadata := map[string]interface{}{
		"backup_time":  time.Now().Format(time.RFC3339),
		"files_count":  len(files),
		"total_size":   totalSize,
		"archive_size": archiveInfo.Size(),
		"compression":  compression,
	}
	metadataJSON, _ := json.MarshalIndent(metadata, "", "  ")
	os.WriteFile(filepath.Join(backupDir, "metadata.json"), metadataJSON, 0644)

	log.Printf("Backup completed: %d files, %d bytes written", len(files), archiveInfo.Size())

	return len(files), archiveInfo.Size(), nil
}

func (s *Service) addFileToZip(zw *zip.Writer, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	header.Name = filePath
	header.Method = zip.Deflate

	writer, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}

func (s *Service) backupDatabase(dbType, connStr, backupDir string, compression bool) error {
	log.Printf("Starting %s database backup", dbType)

	var dumpFile string
	var cmd *exec.Cmd

	switch dbType {
	case "mysql":
		dumpFile = filepath.Join(backupDir, "mysql_dump.sql")
		cmd = exec.Command("mysqldump", "--result-file="+dumpFile, connStr)
	case "postgresql":
		dumpFile = filepath.Join(backupDir, "postgres_dump.sql")
		cmd = exec.Command("pg_dump", connStr)
	default:
		return fmt.Errorf("unsupported database type: %s", dbType)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v - %s", err, string(output))
	}

	log.Printf("Database backup completed: %s", dumpFile)
	return nil
}

func (s *Service) handleRestoreFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		BackupPath  string `json:"backup_path"`
		Destination string `json:"destination"`
	}

	json.NewDecoder(r.Body).Decode(&req)

	// Extract zip
	archivePath := filepath.Join(req.BackupPath, "backup.zip")

	// Simple extraction
	err := s.extractZip(archivePath, req.Destination)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Files restored successfully",
	})
}

func (s *Service) extractZip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		destPath := filepath.Join(destDir, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(destPath, 0755)
			continue
		}

		os.MkdirAll(filepath.Dir(destPath), 0755)

		srcFile, err := f.Open()
		if err != nil {
			return err
		}
		defer srcFile.Close()

		destFile, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer destFile.Close()

		io.Copy(destFile, srcFile)
	}

	return nil
}

func (s *Service) handleSessions(w http.ResponseWriter, r *http.Request) {
	sessionsDir := filepath.Join(s.dataDir, "sessions")

	var sessions []map[string]interface{}

	entries, _ := os.ReadDir(sessionsDir)
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".json") {
			data, _ := os.ReadFile(filepath.Join(sessionsDir, entry.Name()))
			var session map[string]interface{}
			json.Unmarshal(data, &session)
			sessions = append(sessions, session)
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessions": sessions,
	})
}

func (s *Service) handleJobs(w http.ResponseWriter, r *http.Request) {
	jobsFile := filepath.Join(s.dataDir, "config", "jobs.json")

	var jobs []map[string]interface{}

	if data, err := os.ReadFile(jobsFile); err == nil {
		json.Unmarshal(data, &jobs)
	}

	if r.Method == http.MethodPost {
		var job map[string]interface{}
		json.NewDecoder(r.Body).Decode(&job)
		jobs = append(jobs, job)
		os.MkdirAll(filepath.Dir(jobsFile), 0755)
		jobsJSON, _ := json.MarshalIndent(jobs, "", "  ")
		os.WriteFile(jobsFile, jobsJSON, 0644)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"jobs": jobs,
	})
}

func (s *Service) saveSession(sessionID string, session map[string]interface{}) {
	sessionsDir := filepath.Join(s.dataDir, "sessions")
	os.MkdirAll(sessionsDir, 0755)

	sessionFile := filepath.Join(sessionsDir, fmt.Sprintf("%s.json", sessionID))
	data, _ := json.MarshalIndent(session, "", "  ")
	os.WriteFile(sessionFile, data, 0644)
}

// Service management functions

func installService(exePath string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err == nil {
		s.Close()
		return fmt.Errorf("service already exists")
	}

	config := mgr.Config{
		ServiceType:    windows.SERVICE_WIN32_OWN_PROCESS,
		StartType:      mgr.StartAutomatic,
		BinaryPathName: exePath,
		DisplayName:    "NovaBackup Service",
		Description:    "Enterprise Backup and Recovery Platform",
	}

	s, err = m.CreateService(serviceName, exePath, config)
	if err != nil {
		return err
	}
	defer s.Close()

	eventlog.InstallAsEventCreate(serviceName, eventlog.Error|eventlog.Warning|eventlog.Info)
	fmt.Printf("Service %s installed successfully\n", serviceName)
	return nil
}

func removeService() error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("service not installed")
	}
	defer s.Close()

	s.Control(svc.Stop)
	time.Sleep(2 * time.Second)
	s.Delete()
	eventlog.Remove(serviceName)
	fmt.Printf("Service %s removed successfully\n", serviceName)
	return nil
}

func startService() error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("service not installed")
	}
	defer s.Close()

	s.Start()
	for i := 0; i < 30; i++ {
		status, _ := s.Query()
		if status.State == svc.Running {
			fmt.Printf("Service %s started\n", serviceName)
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("timeout")
}

func stopService() error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("service not installed")
	}
	defer s.Close()

	s.Control(svc.Stop)
	for i := 0; i < 30; i++ {
		status, _ := s.Query()
		if status.State == svc.Stopped {
			fmt.Printf("Service %s stopped\n", serviceName)
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("timeout")
}

func runDebug() error {
	fmt.Printf("NovaBackup v6.0.0 - Debug Mode\n")
	fmt.Println("API: http://localhost:8080")
	fmt.Println("Press Ctrl+C to stop\n")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down...")
	return nil
}

func printUsage() {
	fmt.Println("NovaBackup v6.0.0")
	fmt.Println("\nUsage: nova-service <command>")
	fmt.Println("\nCommands:")
	fmt.Println("  install   Install as Windows Service")
	fmt.Println("  remove    Remove service")
	fmt.Println("  start     Start service")
	fmt.Println("  stop      Stop service")
	fmt.Println("  debug     Run in console mode")
	fmt.Println("  version   Show version")
}

func main() {
	isService, _ := svc.IsWindowsService()
	if isService {
		el, _ := eventlog.Open(serviceName)
		svc.Run(serviceName, NewService(el))
		return
	}

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := strings.ToLower(os.Args[1])

	switch cmd {
	case "install":
		exePath, _ := os.Executable()
		if err := installService(exePath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "remove":
		if err := removeService(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "start":
		if err := startService(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "stop":
		if err := stopService(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "debug":
		runDebug()
	case "version":
		fmt.Println("NovaBackup v6.0.0")
	case "help", "-h", "--help":
		printUsage()
	default:
		printUsage()
		os.Exit(1)
	}
}
