//go:build windows

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

func runAsServiceIfNeeded() bool {
	isService, err := svc.IsWindowsService()
	if err != nil || !isService {
		return false
	}

	if err := svc.Run(ServiceName, &novaService{}); err != nil {
		log.Fatalf("Service failed: %v", err)
	}

	return true
}

type novaService struct{}

func (s *novaService) Execute(args []string, r <-chan svc.ChangeRequest, status chan<- svc.Status) (bool, uint32) {
	const acceptedCmds = svc.AcceptStop | svc.AcceptShutdown
	status <- svc.Status{State: svc.StartPending}

	server, err := buildServer()
	if err != nil {
		log.Printf("Server init failed: %v", err)
		status <- svc.Status{State: svc.Stopped}
		return false, 1
	}

	serverErr := make(chan error, 1)
	go func() {
		if err := serveHTTP(server); err != nil {
			serverErr <- err
		}
	}()

	status <- svc.Status{State: svc.Running, Accepts: acceptedCmds}

	for {
		select {
		case change := <-r:
			switch change.Cmd {
			case svc.Interrogate:
				status <- change.CurrentStatus
			case svc.Stop, svc.Shutdown:
				status <- svc.Status{State: svc.StopPending}
				shutdownServer(context.Background(), server)
				status <- svc.Status{State: svc.Stopped}
				return false, 0
			default:
			}
		case err := <-serverErr:
			log.Printf("Server stopped: %v", err)
			status <- svc.Status{State: svc.StopPending}
			shutdownServer(context.Background(), server)
			status <- svc.Status{State: svc.Stopped}
			return false, 1
		}
	}
}

func installService() {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get executable path: %v", err)
	}

	manager, err := mgr.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to service manager: %v", err)
	}
	defer manager.Disconnect()

	service, err := manager.OpenService(ServiceName)
	if err == nil {
		service.Close()
		fmt.Printf("Service %s already installed\n", ServiceName)
		return
	}

	config := mgr.Config{
		DisplayName: ServiceName,
		StartType:   mgr.StartAutomatic,
		Description: "NovaBackup Enterprise v7.0",
	}

	service, err = manager.CreateService(ServiceName, exePath, config, "server")
	if err != nil {
		log.Fatalf("Failed to create service: %v", err)
	}
	defer service.Close()

	if err := eventlog.InstallAsEventCreate(ServiceName, eventlog.Error|eventlog.Warning|eventlog.Info); err != nil {
		fmt.Printf("Event log install failed: %v\n", err)
	}

	fmt.Printf("Service %s installed\n", ServiceName)
}

func removeService() {
	manager, err := mgr.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to service manager: %v", err)
	}
	defer manager.Disconnect()

	service, err := manager.OpenService(ServiceName)
	if err != nil {
		log.Fatalf("Service %s is not installed", ServiceName)
	}
	defer service.Close()

	stopServiceIfRunning(service)

	if err := service.Delete(); err != nil {
		log.Fatalf("Failed to delete service: %v", err)
	}

	_ = eventlog.Remove(ServiceName)
	fmt.Printf("Service %s removed\n", ServiceName)
}

func stopServiceIfRunning(service *mgr.Service) {
	status, err := service.Query()
	if err != nil {
		return
	}
	if status.State != svc.Running {
		return
	}

	status, err = service.Control(svc.Stop)
	if err != nil {
		return
	}

	deadline := time.Now().Add(10 * time.Second)
	for status.State != svc.Stopped {
		if time.Now().After(deadline) {
			return
		}
		time.Sleep(300 * time.Millisecond)
		status, err = service.Query()
		if err != nil {
			return
		}
	}
}
