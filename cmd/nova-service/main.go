// NovaBackup Windows Service
// A complete Windows Service implementation for NovaBackup
// Supports: install, remove, start, stop, debug commands

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"novabackup/internal/api"
	"novabackup/internal/database"
	"novabackup/internal/scheduler"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

const (
	serviceName        = "NovaBackup"
	serviceDisplayName = "NovaBackup Service"
	serviceDescription = "Enterprise Backup and Disaster Recovery Platform"
	defaultPort        = 8080
)

var (
	version   = "6.0.0"
	buildTime = "unknown"
	gitCommit = "unknown"
)

type Service struct {
	db        *database.Connection
	scheduler *scheduler.Scheduler
	apiServer *api.Server
	eventLog  debug.Log
	stopChan  chan struct{}
}

func NewService(eventLog debug.Log) *Service {
	return &Service{
		eventLog: eventLog,
		stopChan: make(chan struct{}),
	}
}

// Event log type constants
const (
	eventLogInfo    = 1
	eventLogWarning = 2
	eventLogError   = 4
)

func (s *Service) logEvent(eventType uint32, msg string) {
	if s.eventLog != nil {
		switch eventType {
		case eventLogError:
			s.eventLog.Error(1, msg)
		case eventLogWarning:
			s.eventLog.Warning(1, msg)
		default:
			s.eventLog.Info(1, msg)
		}
	}
	log.Printf("[%s] %s", time.Now().Format(time.RFC3339), msg)
}

func (s *Service) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (bool, uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue

	changes <- svc.Status{State: svc.StartPending}
	s.logEvent(eventLogInfo, "NovaBackup service is starting...")

	if err := s.startComponents(); err != nil {
		s.logEvent(eventLogError, fmt.Sprintf("Failed to start components: %v", err))
		return true, 1
	}

	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
	s.logEvent(eventLogInfo, fmt.Sprintf("NovaBackup service started successfully on port %d", defaultPort))

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
				s.logEvent(eventLogInfo, "NovaBackup service is stopping...")
				changes <- svc.Status{State: svc.StopPending}
				if err := s.stopComponents(); err != nil {
					s.logEvent(eventLogError, fmt.Sprintf("Error during shutdown: %v", err))
				}
				s.logEvent(eventLogInfo, "NovaBackup service stopped successfully")
				break loop
			case svc.Pause:
				changes <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
				s.logEvent(eventLogInfo, "NovaBackup service paused")
			case svc.Continue:
				changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
				s.logEvent(eventLogInfo, "NovaBackup service resumed")
			default:
				s.logEvent(eventLogWarning, fmt.Sprintf("Unexpected control request: %d", c))
			}
		}
	}

	return false, 0
}

func (s *Service) startComponents() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	exeDir := filepath.Dir(exePath)
	dbPath := filepath.Join(exeDir, "novabackup.db")

	s.logEvent(eventLogInfo, fmt.Sprintf("Opening database: %s", dbPath))
	s.db, err = database.NewSQLiteConnection(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	s.logEvent(eventLogInfo, "Initializing scheduler...")
	s.scheduler, err = scheduler.NewScheduler(s.db)
	if err != nil {
		s.db.Close()
		return fmt.Errorf("failed to create scheduler: %w", err)
	}

	if err := s.scheduler.Start(); err != nil {
		s.db.Close()
		return fmt.Errorf("failed to start scheduler: %w", err)
	}
	s.logEvent(eventLogInfo, "Scheduler started successfully")

	s.logEvent(eventLogInfo, fmt.Sprintf("Starting API server on port %d...", defaultPort))
	s.apiServer, err = api.NewServer(s.db, s.scheduler)
	if err != nil {
		s.scheduler.Stop()
		s.db.Close()
		return fmt.Errorf("failed to create API server: %w", err)
	}

	go func() {
		if err := s.apiServer.Start(defaultPort); err != nil && err != http.ErrServerClosed {
			s.logEvent(eventLogError, fmt.Sprintf("API server error: %v", err))
		}
	}()

	time.Sleep(100 * time.Millisecond)
	return nil
}

func (s *Service) stopComponents() error {
	var errors []string

	if s.apiServer != nil {
		s.logEvent(eventLogInfo, "API server shutdown initiated")
	}

	if s.scheduler != nil {
		if err := s.scheduler.Stop(); err != nil {
			errors = append(errors, fmt.Sprintf("scheduler: %v", err))
		} else {
			s.logEvent(eventLogInfo, "Scheduler stopped")
		}
	}

	if s.db != nil {
		if err := s.db.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("database: %v", err))
		} else {
			s.logEvent(eventLogInfo, "Database connection closed")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors during shutdown: %s", strings.Join(errors, ", "))
	}
	return nil
}

func installService(exePath string) error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err == nil {
		s.Close()
		return fmt.Errorf("service %s already exists", serviceName)
	}

	config := mgr.Config{
		ServiceType:    windows.SERVICE_WIN32_OWN_PROCESS,
		StartType:      mgr.StartAutomatic,
		ErrorControl:   mgr.ErrorNormal,
		BinaryPathName: exePath,
		DisplayName:    serviceDisplayName,
		Description:    serviceDescription,
	}

	s, err = m.CreateService(serviceName, exePath, config)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}
	defer s.Close()

	recoveryActions := []mgr.RecoveryAction{
		{Type: mgr.ServiceRestart, Delay: 60000},
		{Type: mgr.ServiceRestart, Delay: 60000},
		{Type: mgr.ServiceRestart, Delay: 60000},
	}
	err = s.SetRecoveryActions(recoveryActions, 86400)
	if err != nil {
		log.Printf("Warning: Failed to set recovery actions: %v", err)
	}

	err = eventlog.InstallAsEventCreate(serviceName, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		log.Printf("Warning: Failed to install event log: %v", err)
	}

	fmt.Printf("Service %s installed successfully\n", serviceName)
	fmt.Printf("  Display Name: %s\n", serviceDisplayName)
	fmt.Printf("  Binary Path: %s\n", exePath)
	fmt.Printf("  Start Type: Automatic\n\n")
	fmt.Printf("To start the service, run:\n")
	fmt.Printf("  net start %s\n", serviceName)
	fmt.Printf("  or use: sc start %s\n", serviceName)
	return nil
}

func removeService() error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("service %s is not installed: %w", serviceName, err)
	}
	defer s.Close()

	status, err := s.Query()
	if err == nil && status.State != svc.Stopped {
		fmt.Println("Stopping service...")
		_, err = s.Control(svc.Stop)
		if err != nil {
			return fmt.Errorf("failed to stop service: %w", err)
		}
		for i := 0; i < 30; i++ {
			status, _ = s.Query()
			if status.State == svc.Stopped {
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
	}

	err = s.Delete()
	if err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}
	eventlog.Remove(serviceName)
	fmt.Printf("Service %s removed successfully\n", serviceName)
	return nil
}

func startService() error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("service %s is not installed: %w", serviceName, err)
	}
	defer s.Close()

	status, err := s.Query()
	if err == nil && status.State == svc.Running {
		return fmt.Errorf("service %s is already running", serviceName)
	}

	fmt.Println("Starting service...")
	err = s.Start()
	if err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	for i := 0; i < 30; i++ {
		status, _ = s.Query()
		if status.State == svc.Running {
			fmt.Printf("Service %s started successfully\n", serviceName)
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("service start timed out")
}

func stopService() error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("service %s is not installed: %w", serviceName, err)
	}
	defer s.Close()

	status, err := s.Query()
	if err == nil && status.State == svc.Stopped {
		return fmt.Errorf("service %s is already stopped", serviceName)
	}

	fmt.Println("Stopping service...")
	_, err = s.Control(svc.Stop)
	if err != nil {
		return fmt.Errorf("failed to stop service: %w", err)
	}

	for i := 0; i < 30; i++ {
		status, _ = s.Query()
		if status.State == svc.Stopped {
			fmt.Printf("Service %s stopped successfully\n", serviceName)
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("service stop timed out")
}

func runDebug() error {
	fmt.Printf("NovaBackup v%s - Debug Mode\n", version)
	fmt.Println("Running as console application (not as Windows Service)")
	fmt.Println("Press Ctrl+C to stop\n")

	eventLog := debug.New(serviceName)
	service := NewService(eventLog)

	if err := service.startComponents(); err != nil {
		return fmt.Errorf("failed to start: %w", err)
	}

	fmt.Printf("API Server: http://localhost:%d\n", defaultPort)
	fmt.Printf("Swagger UI: http://localhost:%d/swagger/index.html\n", defaultPort)
	fmt.Println("Service is running... Press Ctrl+C to stop")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down...")
	if err := service.stopComponents(); err != nil {
		log.Printf("Shutdown error: %v", err)
	}
	return nil
}

// runService runs as a Windows Service
func runService() error {
	var eventLog debug.Log
	var err error

	el, err := eventlog.Open(serviceName)
	if err != nil {
		eventLog = debug.New(serviceName)
	} else {
		eventLog = el
	}

	service := NewService(eventLog)
	run := svc.Run
	err = run(serviceName, service)
	if err != nil {
		return fmt.Errorf("service failed: %w", err)
	}
	return nil
}

func printUsage() {
	fmt.Printf("NovaBackup v%s - Windows Service Manager\n\n", version)
	fmt.Println("Usage: nova-service <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  install   Install NovaBackup as a Windows Service (runs automatically on boot)")
	fmt.Println("  remove    Remove NovaBackup Windows Service")
	fmt.Println("  start     Start the NovaBackup service")
	fmt.Println("  stop      Stop the NovaBackup service")
	fmt.Println("  debug     Run in debug mode (console application)")
	fmt.Println("  version   Show version information")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Printf("  %s install    # Install as Windows Service\n", os.Args[0])
	fmt.Printf("  %s start      # Start the service\n", os.Args[0])
	fmt.Printf("  %s debug      # Run in console for debugging\n", os.Args[0])
	fmt.Println()
	fmt.Println("After installation, the service can also be managed via:")
	fmt.Println("  - Windows Services Manager (services.msc)")
	fmt.Println("  - net start/stop NovaBackup")
	fmt.Println("  - sc start/stop NovaBackup")
}

func printVersion() {
	fmt.Printf("NovaBackup v%s\n", version)
	fmt.Printf("Build Time: %s\n", buildTime)
	fmt.Printf("Git Commit: %s\n", gitCommit)
}

func main() {
	isWindowsService, err := svc.IsWindowsService()
	if err != nil {
		log.Fatalf("Failed to determine if running as service: %v", err)
	}

	if isWindowsService {
		if err := runService(); err != nil {
			log.Fatalf("Service error: %v", err)
		}
		return
	}

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	command := strings.ToLower(args[0])

	switch command {
	case "install":
		exePath, err := os.Executable()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to get executable path: %v\n", err)
			os.Exit(1)
		}
		if err := checkAdminPrivileges(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Administrator privileges required for installation\n")
			fmt.Fprintf(os.Stderr, "Please run: Right-click -> Run as Administrator\n")
			os.Exit(1)
		}
		if err := installService(exePath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "remove":
		if err := checkAdminPrivileges(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Administrator privileges required for removal\n")
			fmt.Fprintf(os.Stderr, "Please run: Right-click -> Run as Administrator\n")
			os.Exit(1)
		}
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
		if err := runDebug(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "version":
		printVersion()

	case "help", "-h", "--help":
		printUsage()

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func checkAdminPrivileges() error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	m.Disconnect()
	return nil
}
