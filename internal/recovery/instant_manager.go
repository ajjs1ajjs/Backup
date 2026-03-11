package recovery

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/uuid"

	"sync"
	"time"

	"novabackup/internal/database"
	"novabackup/internal/compression"
	"novabackup/internal/storage"

	"github.com/willscott/go-nfs"
	"github.com/willscott/go-nfs/helpers"
)

// RecoverySession represents an active instant recovery session
type RecoverySession struct {
	ID             string
	VMName         string
	RestorePointID string
	NFSExport      string
	NFSPort        int
	StartTime      time.Time
	VFS            *ChunkVFS
	Listener       net.Listener
}

// InstantRecoveryManager handles the lifecycle of instant VM restores
type InstantRecoveryManager struct {
	db       *database.Connection
	comp     compression.Compressor
	storage  *storage.Engine
	sessions map[string]*RecoverySession
	mu       sync.RWMutex
}

func NewInstantRecoveryManager(db *database.Connection, comp compression.Compressor, stor *storage.Engine) *InstantRecoveryManager {
	return &InstantRecoveryManager{
		db:       db,
		comp:     comp,
		storage:  stor,
		sessions: make(map[string]*RecoverySession),
	}
}

// StartNFS starts an NFS server presenting a specific Restore Point
func (m *InstantRecoveryManager) StartNFS(ctx context.Context, rpID string, port int) error {
	id, err := uuid.Parse(rpID)
	if err != nil {
		return err
	}

	// 1. Load RP data from DB
	chunkInfos, err := m.db.GetChunksForRestorePoint(id)
	if err != nil {
		return err
	}
	
	size, err := m.db.GetRestorePointTotalSize(id)
	if err != nil {
		return err
	}

	hashes := make([]string, len(chunkInfos))
	for i, c := range chunkInfos {
		hashes[i] = c.Hash
	}

	// 2. Setup VFS for this session
	vfs := NewChunkVFS(m.db, m.comp, m.storage)
	vname := "disk1.vhdx"
	vfs.AddFile(vname, size, hashes)

	// 3. Create NFS handler
	handler := helpers.NewNullAuthHandler(vfs)

	// 4. Start listener
	log.Printf("[InstantRecovery] Starting virtual NFS server for RP: %s on port %d", rpID, port)
	listener, err := net.Listen("tcp", net.JoinHostPort("0.0.0.0", fmt.Sprintf("%d", port)))
	if err != nil {
		return err
	}

	sessionID := uuid.New().String()
	session := &RecoverySession{
		ID:             sessionID,
		VMName:         "unnamed", // Should probably be passed in
		RestorePointID: rpID,
		NFSExport:      fmt.Sprintf("/exports/%s", sessionID),
		NFSPort:        port,
		StartTime:      time.Now(),
		VFS:            vfs,
		Listener:       listener,
	}

	m.mu.Lock()
	m.sessions[sessionID] = session
	m.mu.Unlock()

	go func() {
		if err := nfs.Serve(listener, handler); err != nil {
			log.Printf("[InstantRecovery] NFS server error: %v", err)
			m.mu.Lock()
			delete(m.sessions, sessionID)
			m.mu.Unlock()
		}
	}()

	return nil
}

// ListSessions returns all active recovery sessions
func (m *InstantRecoveryManager) ListSessions() []*RecoverySession {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessions := make([]*RecoverySession, 0, len(m.sessions))
	for _, s := range m.sessions {
		sessions = append(sessions, s)
	}
	return sessions
}

// StopSession stops a specific recovery session
func (m *InstantRecoveryManager) StopSession(id string) error {
	m.mu.Lock()
	session, ok := m.sessions[id]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("session not found: %s", id)
	}
	delete(m.sessions, id)
	m.mu.Unlock()

	log.Printf("[InstantRecovery] Stopping virtual NFS server for session: %s", id)
	return session.Listener.Close()
}

// StopNFS stops the running virtual NFS server
func (m *InstantRecoveryManager) StopNFS() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, s := range m.sessions {
		log.Printf("[InstantRecovery] Stopping virtual NFS server for session: %s", id)
		s.Listener.Close()
	}
	m.sessions = make(map[string]*RecoverySession)
	return nil
}

// Support for Hyper-V: Mount NFS and Start VM
func (m *InstantRecoveryManager) RecoverToHyperV(ctx context.Context, vmName string, nfsPath string) error {
	// 1. Create a differencing disk on local storage to point to NFS base disk
	// This ensures that all writes go to local storage, keeping the backup read-only
	tempDir := filepath.Join(os.Getenv("TEMP"), "NovaBackup_IR")
	os.MkdirAll(tempDir, 0755)
	
	vhdPath := filepath.Join(tempDir, fmt.Sprintf("%s_diff.vhdx", vmName))
	
	log.Printf("[InstantRecovery] Creating differencing VHD: %s -> %s", vhdPath, nfsPath)
	
	// PowerShell: New-VHD -ParentPath $nfsPath -Path $vhdPath -Differencing
	cmd := exec.Command("powershell", "-Command", fmt.Sprintf("New-VHD -ParentPath '%s' -Path '%s' -Differencing", nfsPath, vhdPath))
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create differencing disk: %v (%s)", err, string(output))
	}

	// 2. Create and start VM
	log.Printf("[InstantRecovery] Creating Hyper-V VM: %s", vmName)
	cmd = exec.Command("powershell", "-Command", fmt.Sprintf("New-VM -Name '%s' -MemoryStartupBytes 2GB -VHDPath '%s'", vmName, vhdPath))
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create VM: %v (%s)", err, string(output))
	}

	log.Printf("[InstantRecovery] Starting Hyper-V VM: %s", vmName)
	cmd = exec.Command("powershell", "-Command", fmt.Sprintf("Start-VM -Name '%s'", vmName))
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to start VM: %v (%s)", err, string(output))
	}

	return nil
}
