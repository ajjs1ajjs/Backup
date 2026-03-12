package guest

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
)

// Helper function to create test logger
func createTestLogger() *zap.Logger {
	logger, _ := zap.NewDevelopment()
	return logger
}

// ============================================================================
// GuestInteractionProxy Tests
// ============================================================================

// TestGuestInteractionProxy_RegisterAgent tests agent registration functionality
func TestGuestInteractionProxy_RegisterAgent(t *testing.T) {
	tests := []struct {
		name        string
		agent       *GuestAgent
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful registration",
			agent: &GuestAgent{
				ID:           "agent-001",
				VMID:         "vm-001",
				VMName:       "TestVM",
				Hostname:     "test-host",
				IPAddress:    "192.168.1.100",
				AgentType:    GuestAgentTypeWindows,
				AgentVersion: "7.0.0",
				Capabilities: []string{"backup", "restore"},
			},
			expectError: false,
		},
		{
			name: "empty agent ID",
			agent: &GuestAgent{
				VMID:      "vm-002",
				VMName:    "TestVM2",
				Hostname:  "test-host2",
				IPAddress: "192.168.1.101",
			},
			expectError: true,
			errorMsg:    "agent ID cannot be empty",
		},
		{
			name: "registration with metadata",
			agent: &GuestAgent{
				ID:           "agent-003",
				VMID:         "vm-003",
				VMName:       "TestVM3",
				Hostname:     "test-host3",
				IPAddress:    "192.168.1.102",
				AgentType:    GuestAgentTypeLinux,
				AgentVersion: "7.0.0",
				Metadata:     map[string]string{"os": "ubuntu", "version": "22.04"},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy := NewGuestInteractionProxy(DefaultGuestInteractionProxyConfig())
			defer proxy.Shutdown(context.Background())

			err := proxy.RegisterAgent(tt.agent)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error '%s' but got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Verify agent was registered with correct status
				retrieved, err := proxy.GetAgent(tt.agent.ID)
				if err != nil {
					t.Errorf("Failed to retrieve registered agent: %v", err)
				} else {
					if retrieved.Status != GuestAgentStatusStarting {
						t.Errorf("Expected status 'Starting' but got '%s'", retrieved.Status)
					}
					if retrieved.HealthScore != 50 {
						t.Errorf("Expected health score 50 but got %d", retrieved.HealthScore)
					}
					if retrieved.RegisteredAt.IsZero() {
						t.Error("Expected RegisteredAt to be set")
					}
					if retrieved.LastHeartbeat.IsZero() {
						t.Error("Expected LastHeartbeat to be set")
					}
				}
			}
		})
	}
}

// TestGuestInteractionProxy_ListAgents tests agent listing functionality
func TestGuestInteractionProxy_ListAgents(t *testing.T) {
	proxy := NewGuestInteractionProxy(DefaultGuestInteractionProxyConfig())
	defer proxy.Shutdown(context.Background())

	// Register multiple agents
	agents := []*GuestAgent{
		{ID: "agent-001", VMID: "vm-001", VMName: "VM1", IPAddress: "192.168.1.1", AgentType: GuestAgentTypeWindows},
		{ID: "agent-002", VMID: "vm-002", VMName: "VM2", IPAddress: "192.168.1.2", AgentType: GuestAgentTypeLinux},
		{ID: "agent-003", VMID: "vm-003", VMName: "VM3", IPAddress: "192.168.1.3", AgentType: GuestAgentTypeWindows},
	}

	for _, agent := range agents {
		if err := proxy.RegisterAgent(agent); err != nil {
			t.Fatalf("Failed to register agent: %v", err)
		}
	}

	// Test listing all agents
	allAgents := proxy.ListAgents()
	if len(allAgents) != 3 {
		t.Errorf("Expected 3 agents but got %d", len(allAgents))
	}

	// Update heartbeat to set some agents online
	proxy.UpdateAgentHeartbeat("agent-001", nil)
	proxy.UpdateAgentHeartbeat("agent-002", nil)

	// Test filtering by status
	onlineAgents := proxy.ListAgents(GuestAgentStatusOnline)
	if len(onlineAgents) != 2 {
		t.Errorf("Expected 2 online agents but got %d", len(onlineAgents))
	}

	startingAgents := proxy.ListAgents(GuestAgentStatusStarting)
	if len(startingAgents) != 1 {
		t.Errorf("Expected 1 starting agent but got %d", len(startingAgents))
	}
}

// TestGuestInteractionProxy_UnregisterAgent tests agent unregistration
func TestGuestInteractionProxy_UnregisterAgent(t *testing.T) {
	proxy := NewGuestInteractionProxy(DefaultGuestInteractionProxyConfig())
	defer proxy.Shutdown(context.Background())

	// Register an agent
	agent := &GuestAgent{
		ID:        "agent-001",
		VMID:      "vm-001",
		VMName:    "TestVM",
		IPAddress: "192.168.1.1",
	}
	if err := proxy.RegisterAgent(agent); err != nil {
		t.Fatalf("Failed to register agent: %v", err)
	}

	// Update heartbeat to bring agent online
	proxy.UpdateAgentHeartbeat("agent-001", nil)

	// Create a session for this agent
	_, err := proxy.StartSession(context.Background(), "agent-001", "testuser", nil)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Verify session exists before unregister
	sessions := proxy.ListSessions()
	if len(sessions) != 1 {
		t.Fatalf("Expected 1 session before unregister, got %d", len(sessions))
	}

	// Unregister the agent
	err = proxy.UnregisterAgent("agent-001")
	if err != nil {
		t.Errorf("Unexpected error during unregistration: %v", err)
	}

	// Verify agent is removed
	_, err = proxy.GetAgent("agent-001")
	if err == nil {
		t.Error("Expected error when getting unregistered agent")
	}

	// Verify session was removed from active sessions (UnregisterAgent deletes sessions)
	sessions = proxy.ListSessions()
	if len(sessions) != 0 {
		t.Errorf("Expected 0 active sessions after unregister, got %d", len(sessions))
	}

	// Test unregistering non-existent agent
	err = proxy.UnregisterAgent("non-existent")
	if err == nil {
		t.Error("Expected error when unregistering non-existent agent")
	}
}

// TestGuestInteractionProxy_StartSession tests session creation
func TestGuestInteractionProxy_StartSession(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func() (*GuestInteractionProxy, func())
		agentID     string
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful session creation",
			setupFunc: func() (*GuestInteractionProxy, func()) {
				proxy := NewGuestInteractionProxy(DefaultGuestInteractionProxyConfig())
				agent := &GuestAgent{
					ID:        "agent-001",
					VMID:      "vm-001",
					VMName:    "TestVM",
					IPAddress: "192.168.1.1",
				}
				proxy.RegisterAgent(agent)
				proxy.UpdateAgentHeartbeat("agent-001", nil)
				return proxy, func() { proxy.Shutdown(context.Background()) }
			},
			agentID:     "agent-001",
			expectError: false,
		},
		{
			name: "non-existent agent",
			setupFunc: func() (*GuestInteractionProxy, func()) {
				proxy := NewGuestInteractionProxy(DefaultGuestInteractionProxyConfig())
				return proxy, func() { proxy.Shutdown(context.Background()) }
			},
			agentID:     "non-existent",
			expectError: true,
			errorMsg:    "agent non-existent not found",
		},
		{
			name: "agent not online",
			setupFunc: func() (*GuestInteractionProxy, func()) {
				proxy := NewGuestInteractionProxy(DefaultGuestInteractionProxyConfig())
				agent := &GuestAgent{
					ID:        "agent-001",
					VMID:      "vm-001",
					VMName:    "TestVM",
					IPAddress: "192.168.1.1",
				}
				proxy.RegisterAgent(agent)
				// Don't update heartbeat - agent stays in Starting status
				return proxy, func() { proxy.Shutdown(context.Background()) }
			},
			agentID:     "agent-001",
			expectError: true,
			errorMsg:    "agent agent-001 is not online (status: Starting)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy, cleanup := tt.setupFunc()
			defer cleanup()

			session, err := proxy.StartSession(context.Background(), tt.agentID, "testuser", map[string]string{
				"ENV": "production",
			})

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error '%s' but got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else {
					if session.AgentID != tt.agentID {
						t.Errorf("Expected agent ID '%s' but got '%s'", tt.agentID, session.AgentID)
					}
					if session.Status != "active" {
						t.Errorf("Expected session status 'active' but got '%s'", session.Status)
					}
					if session.Username != "testuser" {
						t.Errorf("Expected username 'testuser' but got '%s'", session.Username)
					}
				}
			}
		})
	}
}

// TestGuestInteractionProxy_UpdateSessionProgress tests session activity tracking
func TestGuestInteractionProxy_UpdateSessionProgress(t *testing.T) {
	proxy := NewGuestInteractionProxy(DefaultGuestInteractionProxyConfig())
	defer proxy.Shutdown(context.Background())

	// Register and setup agent
	agent := &GuestAgent{
		ID:        "agent-001",
		VMID:      "vm-001",
		VMName:    "TestVM",
		IPAddress: "192.168.1.1",
	}
	proxy.RegisterAgent(agent)
	proxy.UpdateAgentHeartbeat("agent-001", nil)

	// Create session
	session, _ := proxy.StartSession(context.Background(), "agent-001", "testuser", nil)

	// Record initial activity time
	initialActivity := session.LastActivity

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Execute an operation to update session activity
	_, err := proxy.ExecuteOperation(context.Background(), session.ID, "command", "echo test", 30*time.Second)
	if err != nil {
		t.Fatalf("Failed to execute operation: %v", err)
	}

	// Get updated session
	updatedSession, err := proxy.GetSession(session.ID)
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	if !updatedSession.LastActivity.After(initialActivity) {
		t.Error("Expected LastActivity to be updated after operation")
	}
}

// TestGuestInteractionProxy_CompleteSession tests session closure
func TestGuestInteractionProxy_CompleteSession(t *testing.T) {
	proxy := NewGuestInteractionProxy(DefaultGuestInteractionProxyConfig())
	defer proxy.Shutdown(context.Background())

	// Register and setup agent
	agent := &GuestAgent{
		ID:        "agent-001",
		VMID:      "vm-001",
		VMName:    "TestVM",
		IPAddress: "192.168.1.1",
	}
	proxy.RegisterAgent(agent)
	proxy.UpdateAgentHeartbeat("agent-001", nil)

	// Create session
	session, _ := proxy.StartSession(context.Background(), "agent-001", "testuser", nil)

	// Close session
	err := proxy.CloseSession(session.ID)
	if err != nil {
		t.Errorf("Unexpected error closing session: %v", err)
	}

	// Verify session is removed from active sessions
	sessions := proxy.ListSessions()
	if len(sessions) != 0 {
		t.Errorf("Expected 0 active sessions but got %d", len(sessions))
	}

	// Test closing non-existent session
	err = proxy.CloseSession("non-existent")
	if err == nil {
		t.Error("Expected error when closing non-existent session")
	}
}

// ============================================================================
// GuestCredentialManager Tests
// ============================================================================

// TestGuestCredentialManager_Add tests credential addition
func TestGuestCredentialManager_Add(t *testing.T) {
	store := NewInMemoryCredentialStore(InMemoryCredentialStoreConfig{
		Logger: createTestLogger(),
	})
	defer store.Close()

	tests := []struct {
		name        string
		credential  *GuestCredential
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful credential addition",
			credential: &GuestCredential{
				ID:       "cred-001",
				Name:     "Test Credential",
				Username: "admin",
				Password: "password123",
				Type:     "windows",
				Domain:   "EXAMPLE",
			},
			expectError: false,
		},
		{
			name: "empty ID",
			credential: &GuestCredential{
				Name:     "Test Credential 2",
				Username: "user",
				Type:     "linux",
			},
			expectError: true,
			errorMsg:    "credential ID cannot be empty",
		},
		{
			name: "empty username",
			credential: &GuestCredential{
				ID:   "cred-003",
				Name: "Test Credential 3",
				Type: "linux",
			},
			expectError: true,
			errorMsg:    "credential username cannot be empty",
		},
		{
			name: "credential with metadata",
			credential: &GuestCredential{
				ID:       "cred-004",
				Name:     "Test Credential 4",
				Username: "service_account",
				Password: "secret",
				Type:     "service",
				Metadata: map[string]string{"rotation_policy": "30days"},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.Store(tt.credential)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error '%s' but got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Verify credential was stored
				retrieved, err := store.Get(tt.credential.ID)
				if err != nil {
					t.Errorf("Failed to retrieve stored credential: %v", err)
				} else {
					if retrieved.ID != tt.credential.ID {
						t.Errorf("ID mismatch: expected %s, got %s", tt.credential.ID, retrieved.ID)
					}
					if retrieved.Username != tt.credential.Username {
						t.Errorf("Username mismatch: expected %s, got %s", tt.credential.Username, retrieved.Username)
					}
					if retrieved.CreatedAt.IsZero() {
						t.Error("Expected CreatedAt to be set")
					}
					if retrieved.UpdatedAt.IsZero() {
						t.Error("Expected UpdatedAt to be set")
					}
				}
			}
		})
	}
}

// TestGuestCredentialManager_Get tests credential retrieval
func TestGuestCredentialManager_Get(t *testing.T) {
	store := NewInMemoryCredentialStore(InMemoryCredentialStoreConfig{})
	defer store.Close()

	// Store a credential
	original := &GuestCredential{
		ID:          "cred-001",
		Name:        "Test Credential",
		Username:    "admin",
		Password:    "password123",
		Type:        "windows",
		Description: "Test description",
		Metadata:    map[string]string{"key": "value"},
	}
	store.Store(original)

	// Retrieve the credential
	retrieved, err := store.Get("cred-001")
	if err != nil {
		t.Fatalf("Failed to get credential: %v", err)
	}

	if retrieved.ID != original.ID {
		t.Errorf("ID mismatch: expected %s, got %s", original.ID, retrieved.ID)
	}
	if retrieved.Name != original.Name {
		t.Errorf("Name mismatch: expected %s, got %s", original.Name, retrieved.Name)
	}
	if retrieved.Username != original.Username {
		t.Errorf("Username mismatch: expected %s, got %s", original.Username, retrieved.Username)
	}
	if retrieved.Password != original.Password {
		t.Errorf("Password mismatch: expected %s, got %s", original.Password, retrieved.Password)
	}
	if retrieved.Metadata["key"] != "value" {
		t.Errorf("Metadata mismatch: expected value 'value', got '%s'", retrieved.Metadata["key"])
	}

	// Test getting non-existent credential
	_, err = store.Get("non-existent")
	if err == nil {
		t.Error("Expected error when getting non-existent credential")
	}
}

// TestGuestCredentialManager_Delete tests credential deletion
func TestGuestCredentialManager_Delete(t *testing.T) {
	store := NewInMemoryCredentialStore(InMemoryCredentialStoreConfig{})
	defer store.Close()

	// Store a credential
	cred := &GuestCredential{
		ID:       "cred-001",
		Name:     "Test Credential",
		Username: "admin",
		Type:     "windows",
	}
	store.Store(cred)

	// Delete the credential
	err := store.Delete("cred-001")
	if err != nil {
		t.Errorf("Unexpected error deleting credential: %v", err)
	}

	// Verify credential is deleted
	_, err = store.Get("cred-001")
	if err == nil {
		t.Error("Expected error when getting deleted credential")
	}

	// Test deleting non-existent credential
	err = store.Delete("non-existent")
	if err == nil {
		t.Error("Expected error when deleting non-existent credential")
	}
}

// TestGuestCredentialManager_List tests credential listing
func TestGuestCredentialManager_List(t *testing.T) {
	store := NewInMemoryCredentialStore(InMemoryCredentialStoreConfig{})
	defer store.Close()

	// Store multiple credentials
	credentials := []*GuestCredential{
		{ID: "cred-001", Name: "Credential 1", Username: "user1", Type: "windows"},
		{ID: "cred-002", Name: "Credential 2", Username: "user2", Type: "linux"},
		{ID: "cred-003", Name: "Credential 3", Username: "user3", Type: "vmware"},
	}

	for _, cred := range credentials {
		store.Store(cred)
	}

	// List all credentials
	list := store.List()
	if len(list) != 3 {
		t.Errorf("Expected 3 credentials but got %d", len(list))
	}

	// Verify all credentials are returned
	idMap := make(map[string]bool)
	for _, cred := range list {
		idMap[cred.ID] = true
	}

	for _, cred := range credentials {
		if !idMap[cred.ID] {
			t.Errorf("Missing credential %s in list", cred.ID)
		}
	}
}

// TestEncryptedCredentialStore tests encrypted credential storage
func TestEncryptedCredentialStore(t *testing.T) {
	// Create encryption key
	key := []byte("0123456789abcdef0123456789abcdef")

	// Create base store
	baseStore := NewInMemoryCredentialStore(InMemoryCredentialStoreConfig{})
	defer baseStore.Close()

	// Create encrypted store
	encryptedStore, err := NewEncryptedCredentialStore(EncryptedCredentialStoreConfig{
		EncryptionKey: key,
		Store:         baseStore,
	})
	if err != nil {
		t.Fatalf("Failed to create encrypted store: %v", err)
	}

	// Store credential with password
	cred := &GuestCredential{
		ID:       "cred-001",
		Name:     "Encrypted Credential",
		Username: "admin",
		Password: "secret_password",
		Type:     "windows",
	}

	err = encryptedStore.Store(cred)
	if err != nil {
		t.Fatalf("Failed to store encrypted credential: %v", err)
	}

	// Retrieve and verify password is decrypted
	retrieved, err := encryptedStore.Get("cred-001")
	if err != nil {
		t.Fatalf("Failed to get encrypted credential: %v", err)
	}

	if retrieved.Password != "secret_password" {
		t.Errorf("Password decryption failed: expected 'secret_password', got '%s'", retrieved.Password)
	}

	if !retrieved.Encrypted {
		t.Error("Expected credential to be marked as encrypted")
	}
}

// TestValidateCredential tests credential validation
func TestValidateCredential(t *testing.T) {
	tests := []struct {
		name        string
		credential  *GuestCredential
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid credential",
			credential: &GuestCredential{
				ID:       "cred-001",
				Username: "admin",
				Type:     "windows",
			},
			expectError: false,
		},
		{
			name: "missing ID",
			credential: &GuestCredential{
				Username: "admin",
				Type:     "windows",
			},
			expectError: true,
			errorMsg:    "credential ID is required",
		},
		{
			name: "missing username",
			credential: &GuestCredential{
				ID:   "cred-002",
				Type: "windows",
			},
			expectError: true,
			errorMsg:    "username is required",
		},
		{
			name: "missing type",
			credential: &GuestCredential{
				ID:       "cred-003",
				Username: "admin",
			},
			expectError: true,
			errorMsg:    "credential type is required",
		},
		{
			name: "invalid type",
			credential: &GuestCredential{
				ID:       "cred-004",
				Username: "admin",
				Type:     "invalid_type",
			},
			expectError: true,
			errorMsg:    "invalid credential type: invalid_type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCredential(tt.credential)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error '%s' but got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// ============================================================================
// Agent Health Monitoring Tests
// ============================================================================

// TestAgentHealthMonitoring_CheckHealth tests health check functionality
func TestAgentHealthMonitoring_CheckHealth(t *testing.T) {
	proxy := NewGuestInteractionProxy(GuestInteractionProxyConfig{
		HealthCheckInterval: 30 * time.Second,
		HeartbeatTimeout:    2 * time.Minute,
		SessionTimeout:      30 * time.Minute,
		Logger:              createTestLogger(),
	})
	defer proxy.Shutdown(context.Background())

	// Register an agent
	agent := &GuestAgent{
		ID:           "agent-001",
		VMID:         "vm-001",
		VMName:       "TestVM",
		IPAddress:    "192.168.1.1",
		AgentType:    GuestAgentTypeWindows,
		Capabilities: []string{"backup", "restore"},
	}
	proxy.RegisterAgent(agent)
	proxy.UpdateAgentHeartbeat("agent-001", nil)

	// Perform health check
	result, err := proxy.CheckHealth(context.Background(), "agent-001")
	if err != nil {
		t.Fatalf("Failed to perform health check: %v", err)
	}

	if !result.Healthy {
		t.Error("Expected agent to be healthy")
	}

	if result.HealthScore < 70 {
		t.Errorf("Expected health score >= 70, got %d", result.HealthScore)
	}

	// Verify checks
	if !result.Checks["heartbeat_recent"] {
		t.Error("Expected heartbeat_recent check to pass")
	}
	if !result.Checks["status_online"] {
		t.Error("Expected status_online check to pass")
	}
	if !result.Checks["valid_ip"] {
		t.Error("Expected valid_ip check to pass")
	}
	if !result.Checks["has_capabilities"] {
		t.Error("Expected has_capabilities check to pass")
	}

	// Test health check for non-existent agent
	_, err = proxy.CheckHealth(context.Background(), "non-existent")
	if err == nil {
		t.Error("Expected error when checking health of non-existent agent")
	}
}

// TestAgentHealthMonitoring_UnhealthyAgent tests unhealthy agent detection
func TestAgentHealthMonitoring_UnhealthyAgent(t *testing.T) {
	proxy := NewGuestInteractionProxy(GuestInteractionProxyConfig{
		HealthCheckInterval: 30 * time.Second,
		HeartbeatTimeout:    100 * time.Millisecond, // Very short for testing
		SessionTimeout:      30 * time.Minute,
		Logger:              createTestLogger(),
	})
	defer proxy.Shutdown(context.Background())

	// Register an agent
	agent := &GuestAgent{
		ID:        "agent-001",
		VMID:      "vm-001",
		VMName:    "TestVM",
		IPAddress: "192.168.1.1",
	}
	proxy.RegisterAgent(agent)

	// Wait for heartbeat to timeout
	time.Sleep(150 * time.Millisecond)

	// Perform health check
	result, err := proxy.CheckHealth(context.Background(), "agent-001")
	if err != nil {
		t.Fatalf("Failed to perform health check: %v", err)
	}

	if result.Healthy {
		t.Error("Expected agent to be unhealthy due to heartbeat timeout")
	}

	if result.HealthScore >= 70 {
		t.Errorf("Expected health score < 70 for unhealthy agent, got %d", result.HealthScore)
	}

	if !result.Checks["heartbeat_recent"] == false {
		// This is expected - heartbeat should not be recent
	}
}

// TestAgentHealthMonitoring_ListHealthyAgents tests listing healthy agents
func TestAgentHealthMonitoring_ListHealthyAgents(t *testing.T) {
	proxy := NewGuestInteractionProxy(DefaultGuestInteractionProxyConfig())
	defer proxy.Shutdown(context.Background())

	// Register multiple agents
	agents := []*GuestAgent{
		{ID: "agent-001", VMID: "vm-001", VMName: "VM1", IPAddress: "192.168.1.1"},
		{ID: "agent-002", VMID: "vm-002", VMName: "VM2", IPAddress: "192.168.1.2"},
		{ID: "agent-003", VMID: "vm-003", VMName: "VM3", IPAddress: "192.168.1.3"},
	}

	for _, agent := range agents {
		proxy.RegisterAgent(agent)
	}

	// Make some agents healthy
	proxy.UpdateAgentHeartbeat("agent-001", nil)
	proxy.UpdateAgentHeartbeat("agent-002", nil)
	// agent-003 remains in Starting status with health score 50

	// List healthy agents
	healthyAgents := proxy.ListHealthyAgents()

	// Should have 2 healthy agents (agent-001 and agent-002)
	if len(healthyAgents) != 2 {
		t.Errorf("Expected 2 healthy agents but got %d", len(healthyAgents))
	}

	// Verify healthy agents have correct status and score
	for _, agent := range healthyAgents {
		if agent.Status != GuestAgentStatusOnline {
			t.Errorf("Expected healthy agent status to be Online, got %s", agent.Status)
		}
		if agent.HealthScore < 70 {
			t.Errorf("Expected healthy agent score >= 70, got %d", agent.HealthScore)
		}
	}
}

// TestAgentHealthMonitoring_StatusChangeCallback tests status change callbacks
func TestAgentHealthMonitoring_StatusChangeCallback(t *testing.T) {
	proxy := NewGuestInteractionProxy(DefaultGuestInteractionProxyConfig())
	defer proxy.Shutdown(context.Background())

	// Track status changes
	statusChanges := make(map[string][]GuestAgentStatus)
	var mu sync.Mutex

	proxy.SetAgentStatusChangeCallback(func(agentID string, oldStatus, newStatus GuestAgentStatus) {
		mu.Lock()
		defer mu.Unlock()
		statusChanges[agentID] = append(statusChanges[agentID], oldStatus, newStatus)
	})

	// Register an agent
	agent := &GuestAgent{
		ID:        "agent-001",
		VMID:      "vm-001",
		VMName:    "TestVM",
		IPAddress: "192.168.1.1",
	}
	proxy.RegisterAgent(agent)

	// Wait for callback to be invoked
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	changes := statusChanges["agent-001"]
	mu.Unlock()

	if len(changes) == 0 {
		t.Error("Expected status change callback to be invoked")
	}
}

// ============================================================================
// Proxy Statistics Tests
// ============================================================================

// TestProxyStatistics_GetStats tests statistics retrieval
func TestProxyStatistics_GetStats(t *testing.T) {
	proxy := NewGuestInteractionProxy(DefaultGuestInteractionProxyConfig())
	defer proxy.Shutdown(context.Background())

	// Initial stats should show empty
	stats := proxy.GetStats()
	if stats["total_agents"] != 0 {
		t.Errorf("Expected 0 total agents, got %v", stats["total_agents"])
	}
	if stats["total_sessions"] != 0 {
		t.Errorf("Expected 0 total sessions, got %v", stats["total_sessions"])
	}
	if stats["total_operations"] != 0 {
		t.Errorf("Expected 0 total operations, got %v", stats["total_operations"])
	}

	// Register some agents
	agents := []*GuestAgent{
		{ID: "agent-001", VMID: "vm-001", VMName: "VM1", IPAddress: "192.168.1.1"},
		{ID: "agent-002", VMID: "vm-002", VMName: "VM2", IPAddress: "192.168.1.2"},
		{ID: "agent-003", VMID: "vm-003", VMName: "VM3", IPAddress: "192.168.1.3"},
	}

	for _, agent := range agents {
		proxy.RegisterAgent(agent)
	}

	// Make some agents online
	proxy.UpdateAgentHeartbeat("agent-001", nil)
	proxy.UpdateAgentHeartbeat("agent-002", nil)

	// Create some sessions
	proxy.StartSession(context.Background(), "agent-001", "user1", nil)
	proxy.StartSession(context.Background(), "agent-002", "user2", nil)

	// Execute some operations
	session1, _ := proxy.GetSession("session-agent-001-1")
	if session1 != nil {
		proxy.ExecuteOperation(context.Background(), session1.ID, "command", "echo test", 30*time.Second)
	}

	// Get updated stats
	stats = proxy.GetStats()

	if stats["total_agents"] != 3 {
		t.Errorf("Expected 3 total agents, got %v", stats["total_agents"])
	}

	if stats["healthy_agents"] != 2 {
		t.Errorf("Expected 2 healthy agents, got %v", stats["healthy_agents"])
	}

	agentStatusCounts := stats["agent_status"].(map[GuestAgentStatus]int)
	if agentStatusCounts[GuestAgentStatusOnline] != 2 {
		t.Errorf("Expected 2 online agents, got %d", agentStatusCounts[GuestAgentStatusOnline])
	}
	if agentStatusCounts[GuestAgentStatusStarting] != 1 {
		t.Errorf("Expected 1 starting agent, got %d", agentStatusCounts[GuestAgentStatusStarting])
	}
}

// TestProxyStatistics_SessionStats tests session statistics
func TestProxyStatistics_SessionStats(t *testing.T) {
	// Use a very long health check interval to prevent interference
	proxy := NewGuestInteractionProxy(GuestInteractionProxyConfig{
		HealthCheckInterval: 1 * time.Hour,
		HeartbeatTimeout:    2 * time.Minute,
		SessionTimeout:      30 * time.Minute,
		Logger:              zap.NewNop(),
	})
	defer proxy.Shutdown(context.Background())

	// Register and setup agent
	agent := &GuestAgent{
		ID:        "agent-001",
		VMID:      "vm-001",
		VMName:    "TestVM",
		IPAddress: "192.168.1.1",
	}
	proxy.RegisterAgent(agent)
	proxy.UpdateAgentHeartbeat("agent-001", nil)

	// Create multiple sessions with small delays to ensure unique IDs
	_, _ = proxy.StartSession(context.Background(), "agent-001", "user1", nil)
	time.Sleep(1 * time.Millisecond)
	_, _ = proxy.StartSession(context.Background(), "agent-001", "user2", nil)
	time.Sleep(1 * time.Millisecond)
	_, _ = proxy.StartSession(context.Background(), "agent-001", "user3", nil)

	// Get stats before closing any sessions
	stats := proxy.GetStats()
	if stats["total_sessions"] != 3 {
		t.Errorf("Expected 3 total sessions, got %v (stats: %+v)", stats["total_sessions"], stats)
	}

	sessionStatusCounts := stats["session_status"].(map[string]int)
	if sessionStatusCounts["active"] != 3 {
		t.Errorf("Expected 3 active sessions, got %d", sessionStatusCounts["active"])
	}

	// Close one session (note: CloseSession deletes from map)
	sessionList := proxy.ListSessions()
	if len(sessionList) > 0 {
		_ = proxy.CloseSession(sessionList[0].ID)
	}

	// Get stats after closing one session
	stats = proxy.GetStats()
	if stats["total_sessions"] != 2 {
		t.Errorf("Expected 2 total sessions after close, got %v", stats["total_sessions"])
	}
}

// TestProxyStatistics_OperationStats tests operation statistics
func TestProxyStatistics_OperationStats(t *testing.T) {
	// Use a very long health check interval to prevent interference
	proxy := NewGuestInteractionProxy(GuestInteractionProxyConfig{
		HealthCheckInterval: 1 * time.Hour,
		HeartbeatTimeout:    2 * time.Minute,
		SessionTimeout:      30 * time.Minute,
		Logger:              zap.NewNop(),
	})
	defer proxy.Shutdown(context.Background())

	// Register and setup agent
	agent := &GuestAgent{
		ID:        "agent-001",
		VMID:      "vm-001",
		VMName:    "TestVM",
		IPAddress: "192.168.1.1",
	}
	proxy.RegisterAgent(agent)
	proxy.UpdateAgentHeartbeat("agent-001", nil)

	// Create session
	session, _ := proxy.StartSession(context.Background(), "agent-001", "user1", nil)

	// Execute multiple operations with small delays to ensure unique IDs
	op1, _ := proxy.ExecuteOperation(context.Background(), session.ID, "command", "echo 1", 30*time.Second)
	time.Sleep(1 * time.Millisecond)
	op2, _ := proxy.ExecuteOperation(context.Background(), session.ID, "command", "echo 2", 30*time.Second)

	// Update operation statuses
	_ = proxy.UpdateOperationStatus(op1.ID, "completed", 0, "output1", "", "")
	_ = proxy.UpdateOperationStatus(op2.ID, "running", 0, "", "", "")

	// Get stats
	stats := proxy.GetStats()
	if stats["total_operations"] != 2 {
		t.Errorf("Expected 2 total operations, got %v (stats: %+v)", stats["total_operations"], stats)
	}

	opStatusCounts := stats["operation_status"].(map[string]int)
	if opStatusCounts["completed"] != 1 {
		t.Errorf("Expected 1 completed operation, got %d", opStatusCounts["completed"])
	}
	if opStatusCounts["running"] != 1 {
		t.Errorf("Expected 1 running operation, got %d", opStatusCounts["running"])
	}
	if opStatusCounts["pending"] != 0 {
		t.Errorf("Expected 0 pending operations, got %d", opStatusCounts["pending"])
	}
}

// ============================================================================
// Integration Tests
// ============================================================================

// TestGuestInteractionProxy_Integration tests full workflow
func TestGuestInteractionProxy_Integration(t *testing.T) {
	proxy := NewGuestInteractionProxy(GuestInteractionProxyConfig{
		HealthCheckInterval: 30 * time.Second,
		HeartbeatTimeout:    2 * time.Minute,
		SessionTimeout:      30 * time.Minute,
		Logger:              createTestLogger(),
	})
	defer proxy.Shutdown(context.Background())

	// 1. Register agent
	agent := &GuestAgent{
		ID:           "agent-001",
		VMID:         "vm-001",
		VMName:       "ProductionVM",
		Hostname:     "prod-server",
		IPAddress:    "192.168.1.100",
		AgentType:    GuestAgentTypeWindows,
		AgentVersion: "7.0.0",
		Capabilities: []string{"backup", "restore", "quiesce"},
	}

	err := proxy.RegisterAgent(agent)
	if err != nil {
		t.Fatalf("Failed to register agent: %v", err)
	}

	// 2. Update heartbeat to bring agent online
	err = proxy.UpdateAgentHeartbeat("agent-001", map[string]string{"cpu_usage": "45%"})
	if err != nil {
		t.Fatalf("Failed to update heartbeat: %v", err)
	}

	// 3. Start session
	session, err := proxy.StartSession(context.Background(), "agent-001", "administrator", map[string]string{
		"SESSION_TYPE": "backup",
	})
	if err != nil {
		t.Fatalf("Failed to start session: %v", err)
	}

	// 4. Execute operation
	operation, err := proxy.ExecuteOperation(context.Background(), session.ID, "command", "vssadmin list writers", 60*time.Second)
	if err != nil {
		t.Fatalf("Failed to execute operation: %v", err)
	}

	// 5. Update operation status
	err = proxy.UpdateOperationStatus(operation.ID, "completed", 0, "All writers stable", "", "")
	if err != nil {
		t.Fatalf("Failed to update operation status: %v", err)
	}

	// 6. Verify operation completed
	op, err := proxy.GetOperation(operation.ID)
	if err != nil {
		t.Fatalf("Failed to get operation: %v", err)
	}
	if op.Status != "completed" {
		t.Errorf("Expected operation status 'completed', got '%s'", op.Status)
	}
	if op.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", op.ExitCode)
	}

	// 7. Close session
	err = proxy.CloseSession(session.ID)
	if err != nil {
		t.Fatalf("Failed to close session: %v", err)
	}

	// 8. Verify final stats
	stats := proxy.GetStats()
	if stats["total_agents"] != 1 {
		t.Errorf("Expected 1 total agent, got %v", stats["total_agents"])
	}
	if stats["total_sessions"] != 0 {
		t.Errorf("Expected 0 active sessions, got %v", stats["total_sessions"])
	}
	if stats["total_operations"] != 1 {
		t.Errorf("Expected 1 total operation, got %v", stats["total_operations"])
	}
}

// TestConcurrentAccess tests thread safety
func TestConcurrentAccess(t *testing.T) {
	proxy := NewGuestInteractionProxy(DefaultGuestInteractionProxyConfig())
	defer proxy.Shutdown(context.Background())

	// Register initial agent
	proxy.RegisterAgent(&GuestAgent{
		ID:        "agent-001",
		VMID:      "vm-001",
		VMName:    "TestVM",
		IPAddress: "192.168.1.1",
	})

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Concurrent agent registrations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			agent := &GuestAgent{
				ID:        fmt.Sprintf("agent-%03d", i),
				VMID:      fmt.Sprintf("vm-%03d", i),
				VMName:    fmt.Sprintf("VM%d", i),
				IPAddress: fmt.Sprintf("192.168.1.%d", i),
			}
			if err := proxy.RegisterAgent(agent); err != nil {
				errors <- err
			}
		}(i)
	}

	// Concurrent heartbeat updates
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := proxy.UpdateAgentHeartbeat("agent-001", nil); err != nil {
				errors <- err
			}
		}()
	}

	// Concurrent agent listings
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = proxy.ListAgents()
		}()
	}

	// Concurrent stats retrieval
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = proxy.GetStats()
		}()
	}

	wg.Wait()
	close(errors)

	errorCount := 0
	for err := range errors {
		t.Errorf("Concurrent access error: %v", err)
		errorCount++
	}

	if errorCount > 0 {
		t.Errorf("Total concurrent access errors: %d", errorCount)
	}

	// Verify all agents were registered
	agents := proxy.ListAgents()
	if len(agents) < 10 {
		t.Errorf("Expected at least 10 agents, got %d", len(agents))
	}
}
