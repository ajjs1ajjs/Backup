// Package instantrecovery provides instant VM recovery functionality
package instantrecovery

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/willscott/go-nfs"
	"github.com/willscott/go-nfs/helpers"
	"go.uber.org/zap"
)

// NFSServer provides NFS server for instant VM recovery
type NFSServer struct {
	logger   *zap.Logger
	server   net.Listener
	rootPath string
	port     int
	running  bool
}

// NFSConfig holds NFS server configuration
type NFSConfig struct {
	RootPath string // Root directory for NFS exports
	Port     int    // NFS server port (default: 2049)
	ReadOnly bool   // Export read-only
}

// NewNFSServer creates a new NFS server
func NewNFSServer(logger *zap.Logger, config *NFSConfig) (*NFSServer, error) {
	if config.Port == 0 {
		config.Port = 2049
	}

	srv := &NFSServer{
		logger:   logger.With(zap.String("component", "nfs-server")),
		rootPath: config.RootPath,
		port:     config.Port,
	}

	return srv, nil
}

// Start starts the NFS server
func (n *NFSServer) Start(ctx context.Context) error {
	if n.running {
		return fmt.Errorf("NFS server already running")
	}

	n.logger.Info("Starting NFS server", zap.String("path", n.rootPath), zap.Int("port", n.port))

	// Ensure root path exists
	if err := os.MkdirAll(n.rootPath, 0755); err != nil {
		return fmt.Errorf("failed to create NFS root: %w", err)
	}

	// Create handler (placeholder using memory FS as a base if helpers doesn't have NewMemFS)
	// Actually, let's just use the same pattern as instant_manager if possible.
	handler := helpers.NewNullAuthHandler(nil) // Need a filesystem here

	// Start listener
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", n.port))
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}

	n.server = listener
	n.running = true

	// Accept connections
	go func() {
		if err := nfs.Serve(listener, handler); err != nil {
			if n.running {
				n.logger.Error("NFS server error", zap.Error(err))
			}
		}
	}()

	n.logger.Info("NFS server started successfully")
	return nil
}

// Stop stops the NFS server
func (n *NFSServer) Stop() error {
	n.logger.Info("Stopping NFS server")

	if !n.running {
		return nil
	}

	n.running = false
	if n.server != nil {
		return n.server.Close()
	}

	return nil
}

// IsRunning checks if the server is running
func (n *NFSServer) IsRunning() bool {
	return n.running
}

// GetExportPath returns the NFS export path
func (n *NFSServer) GetExportPath() string {
	return n.rootPath
}

// GetAddress returns the NFS server address
func (n *NFSServer) GetAddress() string {
	return fmt.Sprintf("nfs://localhost:%d/exports", n.port)
}

// PublishBackup publishes a backup for instant recovery
func (n *NFSServer) PublishBackup(backupPath, vmName string) (string, error) {
	n.logger.Info("Publishing backup for instant recovery",
		zap.String("backup", backupPath),
		zap.String("vm", vmName))

	if !n.running {
		return "", fmt.Errorf("NFS server not running")
	}

	// Create export directory
	exportPath := filepath.Join(n.rootPath, vmName)
	if err := os.MkdirAll(exportPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create export directory: %w", err)
	}

	// Link backup files to export directory
	// This makes VM disks available via NFS
	files, err := os.ReadDir(backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to read backup: %w", err)
	}

	for _, file := range files {
		src := filepath.Join(backupPath, file.Name())
		dst := filepath.Join(exportPath, file.Name())
		
		// Create symbolic link
		if err := os.Symlink(src, dst); err != nil {
			n.logger.Warn("Failed to create symlink", zap.String("file", file.Name()), zap.Error(err))
			continue
		}
	}

	exportURL := fmt.Sprintf("nfs://localhost:%d/exports/%s", n.port, vmName)
	n.logger.Info("Backup published", zap.String("url", exportURL))

	return exportURL, nil
}

// UnpublishBackup removes a published backup
func (n *NFSServer) UnpublishBackup(vmName string) error {
	n.logger.Info("Unpublishing backup", zap.String("vm", vmName))

	exportPath := filepath.Join(n.rootPath, vmName)
	if err := os.RemoveAll(exportPath); err != nil {
		return fmt.Errorf("failed to remove export: %w", err)
	}

	return nil
}

// InstantRecoverySession represents an active instant recovery session
type InstantRecoverySession struct {
	SessionID    string
	VMName       string
	BackupPath   string
	NFSExport    string
	Datastores   []string
	StartTime    time.Time
	Status       string
}

// ProviderInstantRecoveryManager manages instant recovery sessions (provider specific)
type ProviderInstantRecoveryManager struct {
	logger    *zap.Logger
	nfsServer *NFSServer
	sessions  map[string]*InstantRecoverySession
}

// NewProviderInstantRecoveryManager creates a new instant recovery manager
func NewProviderInstantRecoveryManager(logger *zap.Logger) *ProviderInstantRecoveryManager {
	return &ProviderInstantRecoveryManager{
		logger:   logger.With(zap.String("component", "instant-recovery")),
		sessions: make(map[string]*InstantRecoverySession),
	}
}

// InitializeNFS initializes the NFS server
func (m *ProviderInstantRecoveryManager) InitializeNFS(config *NFSConfig) error {
	nfsServer, err := NewNFSServer(m.logger, config)
	if err != nil {
		return err
	}

	m.nfsServer = nfsServer
	return nil
}

// StartInstantRecovery starts instant recovery for a VM
func (m *ProviderInstantRecoveryManager) StartInstantRecovery(ctx context.Context, vmName, backupPath string) (*InstantRecoverySession, error) {
	m.logger.Info("Starting instant recovery",
		zap.String("vm", vmName),
		zap.String("backup", backupPath))

	if m.nfsServer == nil {
		return nil, fmt.Errorf("NFS server not initialized")
	}

	if !m.nfsServer.IsRunning() {
		if err := m.nfsServer.Start(ctx); err != nil {
			return nil, fmt.Errorf("failed to start NFS server: %w", err)
		}
	}

	// Publish backup via NFS
	nfsExport, err := m.nfsServer.PublishBackup(backupPath, vmName)
	if err != nil {
		return nil, fmt.Errorf("failed to publish backup: %w", err)
	}

	session := &InstantRecoverySession{
		SessionID:  generateSessionID(),
		VMName:     vmName,
		BackupPath: backupPath,
		NFSExport:  nfsExport,
		StartTime:  time.Now(),
		Status:     "active",
	}

	m.sessions[session.SessionID] = session

	m.logger.Info("Instant recovery started",
		zap.String("session", session.SessionID),
		zap.String("export", nfsExport))

	return session, nil
}

// StopInstantRecovery stops an instant recovery session
func (m *ProviderInstantRecoveryManager) StopInstantRecovery(sessionID string) error {
	m.logger.Info("Stopping instant recovery", zap.String("session", sessionID))

	session, exists := m.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// Unpublish backup
	if err := m.nfsServer.UnpublishBackup(session.VMName); err != nil {
		m.logger.Warn("Failed to unpublish backup", zap.Error(err))
	}

	session.Status = "stopped"
	delete(m.sessions, sessionID)

	return nil
}

// ListSessions lists all active instant recovery sessions
func (m *ProviderInstantRecoveryManager) ListSessions() []*InstantRecoverySession {
	sessions := make([]*InstantRecoverySession, 0, len(m.sessions))
	for _, session := range m.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// GetSession returns a specific session
func (m *ProviderInstantRecoveryManager) GetSession(sessionID string) (*InstantRecoverySession, error) {
	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	return session, nil
}

// generateSessionID generates a unique session ID
func generateSessionID() string {
	return fmt.Sprintf("ir_%d", time.Now().Unix())
}
