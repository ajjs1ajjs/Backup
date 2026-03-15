// NovaBackup Windows Service - Minimal Version
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
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
	serviceName        = "NovaBackup"
	serviceDisplayName = "NovaBackup Service"
	serviceDescription = "Enterprise Backup and Disaster Recovery Platform"
	defaultPort        = 8080
	eventLogInfo       = 1
	eventLogWarning    = 2
	eventLogError      = 4
)

var (
	version   = "6.0.0"
	buildTime = "unknown"
	gitCommit = "unknown"
)

type Service struct {
	eventLog debug.Log
	stopChan chan struct{}
}

func NewService(eventLog debug.Log) *Service {
	return &Service{
		eventLog: eventLog,
		stopChan: make(chan struct{}),
	}
}

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

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down...")
	return nil
}

func runService() error {
	var eventLog debug.Log

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
	fmt.Println("  install   Install NovaBackup as a Windows Service")
	fmt.Println("  remove    Remove NovaBackup Windows Service")
	fmt.Println("  start     Start the NovaBackup service")
	fmt.Println("  stop      Stop the NovaBackup service")
	fmt.Println("  debug     Run in debug mode (console application)")
	fmt.Println("  version   Show version information")
	fmt.Println()
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

	command := strings.ToLower(os.Args[1])

	switch command {
	case "install":
		exePath, err := os.Executable()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to get executable path: %v\n", err)
			os.Exit(1)
		}
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
