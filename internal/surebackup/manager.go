package surebackup

import (
	"context"
	"fmt"
	"sync"
	"time"

	"novabackup/internal/multitenancy"
	"novabackup/internal/storage"

	"github.com/google/uuid"
)

// SureBackupManager manages sandbox environments for backup verification
type SureBackupManager interface {
	// Sandbox operations
	CreateSandbox(ctx context.Context, request *SandboxRequest) (*Sandbox, error)
	GetSandbox(ctx context.Context, sandboxID string) (*Sandbox, error)
	ListSandboxes(ctx context.Context, filter *SandboxFilter) ([]*Sandbox, error)
	DeleteSandbox(ctx context.Context, sandboxID string) error
	StartSandbox(ctx context.Context, sandboxID string) error
	StopSandbox(ctx context.Context, sandboxID string) error

	// Verification operations
	StartVerification(ctx context.Context, request *VerificationRequest) (*Verification, error)
	GetVerification(ctx context.Context, verificationID string) (*Verification, error)
	ListVerifications(ctx context.Context, filter *VerificationFilter) ([]*Verification, error)
	StopVerification(ctx context.Context, verificationID string) error

	// Statistics
	GetSureBackupStats(ctx context.Context, tenantID string) (*SureBackupStats, error)
	GetGlobalStats(ctx context.Context) (*GlobalSureBackupStats, error)
}

// Sandbox represents an isolated verification environment
type Sandbox struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	TenantID    string             `json:"tenant_id"`
	Status      SandboxStatus      `json:"status"`
	Type        SandboxType        `json:"type"`
	Environment SandboxEnvironment `json:"environment"`
	Resources   SandboxResources   `json:"resources"`
	Network     SandboxNetwork     `json:"network"`
	Storage     SandboxStorage     `json:"storage"`
	CreatedAt   time.Time          `json:"created_at"`
	StartedAt   *time.Time         `json:"started_at,omitempty"`
	StoppedAt   *time.Time         `json:"stopped_at,omitempty"`
	ExpiresAt   *time.Time         `json:"expires_at,omitempty"`
	Metadata    map[string]string  `json:"metadata"`
}

// SandboxRequest contains parameters for creating a sandbox
type SandboxRequest struct {
	Name        string             `json:"name"`
	TenantID    string             `json:"tenant_id"`
	Type        SandboxType        `json:"type"`
	Environment SandboxEnvironment `json:"environment"`
	Resources   SandboxResources   `json:"resources"`
	Network     SandboxNetwork     `json:"network"`
	Storage     SandboxStorage     `json:"storage"`
	AutoStart   bool               `json:"auto_start"`
	ExpiresIn   time.Duration      `json:"expires_in"`
	Metadata    map[string]string  `json:"metadata"`
}

// SandboxFilter contains filters for listing sandboxes
type SandboxFilter struct {
	TenantID      string        `json:"tenant_id,omitempty"`
	Status        SandboxStatus `json:"status,omitempty"`
	Type          SandboxType   `json:"type,omitempty"`
	CreatedAfter  *time.Time    `json:"created_after,omitempty"`
	CreatedBefore *time.Time    `json:"created_before,omitempty"`
}

// SandboxStatus represents the status of a sandbox
type SandboxStatus string

const (
	SandboxStatusCreating SandboxStatus = "creating"
	SandboxStatusRunning  SandboxStatus = "running"
	SandboxStatusStopped  SandboxStatus = "stopped"
	SandboxStatusError    SandboxStatus = "error"
	SandboxStatusDeleting SandboxStatus = "deleting"
	SandboxStatusExpired  SandboxStatus = "expired"
)

// SandboxType represents the type of sandbox environment
type SandboxType string

const (
	SandboxTypeVM        SandboxType = "vm"
	SandboxTypeContainer SandboxType = "container"
	SandboxTypeIsolated  SandboxType = "isolated"
)

// SandboxEnvironment contains environment configuration
type SandboxEnvironment struct {
	OS           string            `json:"os"`
	Version      string            `json:"version"`
	Architecture string            `json:"architecture"`
	Tools        []string          `json:"tools"`
	Config       map[string]string `json:"config"`
}

// SandboxResources contains resource allocation
type SandboxResources struct {
	CPUCount    int `json:"cpu_count"`
	MemoryMB    int `json:"memory_mb"`
	StorageGB   int `json:"storage_gb"`
	NetworkMbps int `json:"network_mbps"`
}

// SandboxNetwork contains network configuration
type SandboxNetwork struct {
	Isolated     bool     `json:"isolated"`
	Subnet       string   `json:"subnet,omitempty"`
	AllowedPorts []int    `json:"allowed_ports,omitempty"`
	DNS          []string `json:"dns,omitempty"`
	ProxyURL     string   `json:"proxy_url,omitempty"`
}

// SandboxStorage contains storage configuration
type SandboxStorage struct {
	BackendType string            `json:"backend_type"`
	Config      map[string]string `json:"config"`
	MountPoints []MountPoint      `json:"mount_points"`
}

// MountPoint represents a storage mount point
type MountPoint struct {
	Path     string `json:"path"`
	SizeGB   int    `json:"size_gb"`
	ReadOnly bool   `json:"read_only"`
}

// Verification represents a backup verification process
type Verification struct {
	ID          string              `json:"id"`
	SandboxID   string              `json:"sandbox_id"`
	BackupID    string              `json:"backup_id"`
	TenantID    string              `json:"tenant_id"`
	Status      VerificationStatus  `json:"status"`
	Type        VerificationType    `json:"type"`
	Config      VerificationConfig  `json:"config"`
	Results     VerificationResults `json:"results,omitempty"`
	CreatedAt   time.Time           `json:"created_at"`
	StartedAt   *time.Time          `json:"started_at,omitempty"`
	CompletedAt *time.Time          `json:"completed_at,omitempty"`
	Error       string              `json:"error,omitempty"`
}

// VerificationRequest contains parameters for starting verification
type VerificationRequest struct {
	SandboxID string             `json:"sandbox_id"`
	BackupID  string             `json:"backup_id"`
	TenantID  string             `json:"tenant_id"`
	Type      VerificationType   `json:"type"`
	Config    VerificationConfig `json:"config"`
}

// VerificationFilter contains filters for listing verifications
type VerificationFilter struct {
	TenantID      string             `json:"tenant_id,omitempty"`
	SandboxID     string             `json:"sandbox_id,omitempty"`
	BackupID      string             `json:"backup_id,omitempty"`
	Status        VerificationStatus `json:"status,omitempty"`
	Type          VerificationType   `json:"type,omitempty"`
	CreatedAfter  *time.Time         `json:"created_after,omitempty"`
	CreatedBefore *time.Time         `json:"created_before,omitempty"`
}

// VerificationStatus represents the status of a verification
type VerificationStatus string

const (
	VerificationStatusPending   VerificationStatus = "pending"
	VerificationStatusRunning   VerificationStatus = "running"
	VerificationStatusPassed    VerificationStatus = "passed"
	VerificationStatusFailed    VerificationStatus = "failed"
	VerificationStatusCancelled VerificationStatus = "cancelled"
	VerificationStatusError     VerificationStatus = "error"
)

// VerificationType represents the type of verification
type VerificationType string

const (
	VerificationTypeIntegrity VerificationType = "integrity"
	VerificationTypeMount     VerificationType = "mount"
	VerificationTypeBoot      VerificationType = "boot"
	VerificationTypeApp       VerificationType = "application"
	VerificationTypeFull      VerificationType = "full"
)

// VerificationConfig contains verification configuration
type VerificationConfig struct {
	Timeout       time.Duration        `json:"timeout"`
	Tests         []VerificationTest   `json:"tests"`
	Scripts       []VerificationScript `json:"scripts"`
	Notifications NotificationConfig   `json:"notifications"`
}

// VerificationTest represents a specific test to run
type VerificationTest struct {
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	Command    string            `json:"command"`
	Expected   string            `json:"expected"`
	Timeout    time.Duration     `json:"timeout"`
	RetryCount int               `json:"retry_count"`
	Parameters map[string]string `json:"parameters"`
}

// VerificationScript represents a script to run
type VerificationScript struct {
	Name        string            `json:"name"`
	Path        string            `json:"path"`
	Args        []string          `json:"args"`
	Environment map[string]string `json:"environment"`
	RunAs       string            `json:"run_as"`
	Timeout     time.Duration     `json:"timeout"`
}

// NotificationConfig contains notification settings
type NotificationConfig struct {
	OnSuccess []string `json:"on_success"`
	OnFailure []string `json:"on_failure"`
	OnError   []string `json:"on_error"`
}

// VerificationResults contains verification results
type VerificationResults struct {
	Passed        bool                   `json:"passed"`
	Duration      time.Duration          `json:"duration"`
	TestResults   []TestResult           `json:"test_results"`
	ScriptResults []ScriptResult         `json:"script_results"`
	Summary       VerificationSummary    `json:"summary"`
	Artifacts     []VerificationArtifact `json:"artifacts"`
}

// TestResult represents result of a single test
type TestResult struct {
	Name     string        `json:"name"`
	Type     string        `json:"type"`
	Passed   bool          `json:"passed"`
	Duration time.Duration `json:"duration"`
	Output   string        `json:"output"`
	Error    string        `json:"error,omitempty"`
	Retries  int           `json:"retries"`
}

// ScriptResult represents result of a script execution
type ScriptResult struct {
	Name     string        `json:"name"`
	ExitCode int           `json:"exit_code"`
	Duration time.Duration `json:"duration"`
	Output   string        `json:"output"`
	Error    string        `json:"error,omitempty"`
}

// VerificationSummary contains verification summary
type VerificationSummary struct {
	TotalTests      int  `json:"total_tests"`
	PassedTests     int  `json:"passed_tests"`
	FailedTests     int  `json:"failed_tests"`
	TotalScripts    int  `json:"total_scripts"`
	PassedScripts   int  `json:"passed_scripts"`
	FailedScripts   int  `json:"failed_scripts"`
	DataIntegrityOK bool `json:"data_integrity_ok"`
	MountSuccess    bool `json:"mount_success"`
	BootSuccess     bool `json:"boot_success"`
}

// VerificationArtifact represents verification artifacts
type VerificationArtifact struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	Checksum string `json:"checksum"`
}

// SureBackupStats contains tenant-specific statistics
type SureBackupStats struct {
	TenantID             string        `json:"tenant_id"`
	TotalSandboxes       int64         `json:"total_sandboxes"`
	RunningSandboxes     int64         `json:"running_sandboxes"`
	TotalVerifications   int64         `json:"total_verifications"`
	RunningVerifications int64         `json:"running_verifications"`
	PassedVerifications  int64         `json:"passed_verifications"`
	FailedVerifications  int64         `json:"failed_verifications"`
	AvgVerificationTime  time.Duration `json:"avg_verification_time"`
	LastUpdated          time.Time     `json:"last_updated"`
}

// GlobalSureBackupStats contains global statistics
type GlobalSureBackupStats struct {
	TotalTenants         int64         `json:"total_tenants"`
	TotalSandboxes       int64         `json:"total_sandboxes"`
	RunningSandboxes     int64         `json:"running_sandboxes"`
	TotalVerifications   int64         `json:"total_verifications"`
	RunningVerifications int64         `json:"running_verifications"`
	PassedVerifications  int64         `json:"passed_verifications"`
	FailedVerifications  int64         `json:"failed_verifications"`
	AvgVerificationTime  time.Duration `json:"avg_verification_time"`
	LastUpdated          time.Time     `json:"last_updated"`
}

// InMemorySureBackupManager implements SureBackupManager interface
type InMemorySureBackupManager struct {
	sandboxes      map[string]*Sandbox
	verifications  map[string]*Verification
	stats          *SureBackupStats
	globalStats    *GlobalSureBackupStats
	tenantManager  multitenancy.TenantManager
	storageManager storage.Engine
	mutex          sync.RWMutex
}

// NewInMemorySureBackupManager creates a new in-memory SureBackup manager
func NewInMemorySureBackupManager(tenantMgr multitenancy.TenantManager, storageMgr storage.Engine) *InMemorySureBackupManager {
	return &InMemorySureBackupManager{
		sandboxes:     make(map[string]*Sandbox),
		verifications: make(map[string]*Verification),
		stats: &SureBackupStats{
			LastUpdated: time.Now(),
		},
		globalStats: &GlobalSureBackupStats{
			LastUpdated: time.Now(),
		},
		tenantManager:  tenantMgr,
		storageManager: storageMgr,
	}
}

// CreateSandbox creates a new sandbox environment
func (m *InMemorySureBackupManager) CreateSandbox(ctx context.Context, request *SandboxRequest) (*Sandbox, error) {
	if err := m.validateTenantAccess(ctx, request.TenantID); err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	sandbox := &Sandbox{
		ID:          generateSandboxID(),
		Name:        request.Name,
		TenantID:    request.TenantID,
		Status:      SandboxStatusCreating,
		Type:        request.Type,
		Environment: request.Environment,
		Resources:   request.Resources,
		Network:     request.Network,
		Storage:     request.Storage,
		CreatedAt:   time.Now(),
		Metadata:    request.Metadata,
	}

	if request.ExpiresIn > 0 {
		expiresAt := time.Now().Add(request.ExpiresIn)
		sandbox.ExpiresAt = &expiresAt
	}

	m.sandboxes[sandbox.ID] = sandbox

	if request.AutoStart {
		if err := m.startSandboxInternal(ctx, sandbox); err != nil {
			sandbox.Status = SandboxStatusError
			return sandbox, err
		}
	}

	return sandbox, nil
}

// GetSandbox retrieves a sandbox by ID
func (m *InMemorySureBackupManager) GetSandbox(ctx context.Context, sandboxID string) (*Sandbox, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	sandbox, exists := m.sandboxes[sandboxID]
	if !exists {
		return nil, fmt.Errorf("sandbox %s not found", sandboxID)
	}

	if err := m.validateTenantAccess(ctx, sandbox.TenantID); err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	return sandbox, nil
}

// ListSandboxes lists sandboxes with optional filtering
func (m *InMemorySureBackupManager) ListSandboxes(ctx context.Context, filter *SandboxFilter) ([]*Sandbox, error) {
	if filter != nil && filter.TenantID != "" {
		if err := m.validateTenantAccess(ctx, filter.TenantID); err != nil {
			return nil, fmt.Errorf("access denied: %w", err)
		}
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var results []*Sandbox
	for _, sandbox := range m.sandboxes {
		if filter != nil {
			if filter.TenantID != "" && sandbox.TenantID != filter.TenantID {
				continue
			}
			if filter.Status != "" && sandbox.Status != filter.Status {
				continue
			}
			if filter.Type != "" && sandbox.Type != filter.Type {
				continue
			}
			if filter.CreatedAfter != nil && sandbox.CreatedAt.Before(*filter.CreatedAfter) {
				continue
			}
			if filter.CreatedBefore != nil && sandbox.CreatedAt.After(*filter.CreatedBefore) {
				continue
			}
		}

		// Check tenant access
		if err := m.validateTenantAccess(ctx, sandbox.TenantID); err != nil {
			continue
		}

		results = append(results, sandbox)
	}

	return results, nil
}

// DeleteSandbox deletes a sandbox
func (m *InMemorySureBackupManager) DeleteSandbox(ctx context.Context, sandboxID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	sandbox, exists := m.sandboxes[sandboxID]
	if !exists {
		return fmt.Errorf("sandbox %s not found", sandboxID)
	}

	if err := m.validateTenantAccess(ctx, sandbox.TenantID); err != nil {
		return fmt.Errorf("access denied: %w", err)
	}

	// Stop sandbox if running
	if sandbox.Status == SandboxStatusRunning {
		m.stopSandboxInternal(ctx, sandbox)
	}

	sandbox.Status = SandboxStatusDeleting
	delete(m.sandboxes, sandboxID)

	return nil
}

// StartSandbox starts a sandbox
func (m *InMemorySureBackupManager) StartSandbox(ctx context.Context, sandboxID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	sandbox, exists := m.sandboxes[sandboxID]
	if !exists {
		return fmt.Errorf("sandbox %s not found", sandboxID)
	}

	if err := m.validateTenantAccess(ctx, sandbox.TenantID); err != nil {
		return fmt.Errorf("access denied: %w", err)
	}

	return m.startSandboxInternal(ctx, sandbox)
}

// StopSandbox stops a sandbox
func (m *InMemorySureBackupManager) StopSandbox(ctx context.Context, sandboxID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	sandbox, exists := m.sandboxes[sandboxID]
	if !exists {
		return fmt.Errorf("sandbox %s not found", sandboxID)
	}

	if err := m.validateTenantAccess(ctx, sandbox.TenantID); err != nil {
		return fmt.Errorf("access denied: %w", err)
	}

	m.stopSandboxInternal(ctx, sandbox)
	return nil
}

// StartVerification starts a backup verification
func (m *InMemorySureBackupManager) StartVerification(ctx context.Context, request *VerificationRequest) (*Verification, error) {
	if err := m.validateTenantAccess(ctx, request.TenantID); err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if sandbox exists and is running
	sandbox, exists := m.sandboxes[request.SandboxID]
	if !exists {
		return nil, fmt.Errorf("sandbox %s not found", request.SandboxID)
	}

	if sandbox.Status != SandboxStatusRunning {
		return nil, fmt.Errorf("sandbox %s is not running", request.SandboxID)
	}

	verification := &Verification{
		ID:        generateVerificationID(),
		SandboxID: request.SandboxID,
		BackupID:  request.BackupID,
		TenantID:  request.TenantID,
		Status:    VerificationStatusPending,
		Type:      request.Type,
		Config:    request.Config,
		CreatedAt: time.Now(),
	}

	m.verifications[verification.ID] = verification

	// Start verification in background
	go m.runVerification(ctx, verification)

	return verification, nil
}

// GetVerification retrieves a verification by ID
func (m *InMemorySureBackupManager) GetVerification(ctx context.Context, verificationID string) (*Verification, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	verification, exists := m.verifications[verificationID]
	if !exists {
		return nil, fmt.Errorf("verification %s not found", verificationID)
	}

	if err := m.validateTenantAccess(ctx, verification.TenantID); err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	return verification, nil
}

// ListVerifications lists verifications with optional filtering
func (m *InMemorySureBackupManager) ListVerifications(ctx context.Context, filter *VerificationFilter) ([]*Verification, error) {
	if filter != nil && filter.TenantID != "" {
		if err := m.validateTenantAccess(ctx, filter.TenantID); err != nil {
			return nil, fmt.Errorf("access denied: %w", err)
		}
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var results []*Verification
	for _, verification := range m.verifications {
		if filter != nil {
			if filter.TenantID != "" && verification.TenantID != filter.TenantID {
				continue
			}
			if filter.SandboxID != "" && verification.SandboxID != filter.SandboxID {
				continue
			}
			if filter.BackupID != "" && verification.BackupID != filter.BackupID {
				continue
			}
			if filter.Status != "" && verification.Status != filter.Status {
				continue
			}
			if filter.Type != "" && verification.Type != filter.Type {
				continue
			}
			if filter.CreatedAfter != nil && verification.CreatedAt.Before(*filter.CreatedAfter) {
				continue
			}
			if filter.CreatedBefore != nil && verification.CreatedAt.After(*filter.CreatedBefore) {
				continue
			}
		}

		// Check tenant access
		if err := m.validateTenantAccess(ctx, verification.TenantID); err != nil {
			continue
		}

		results = append(results, verification)
	}

	return results, nil
}

// StopVerification stops a verification
func (m *InMemorySureBackupManager) StopVerification(ctx context.Context, verificationID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	verification, exists := m.verifications[verificationID]
	if !exists {
		return fmt.Errorf("verification %s not found", verificationID)
	}

	if err := m.validateTenantAccess(ctx, verification.TenantID); err != nil {
		return fmt.Errorf("access denied: %w", err)
	}

	if verification.Status == VerificationStatusRunning {
		verification.Status = VerificationStatusCancelled
		now := time.Now()
		verification.CompletedAt = &now
	}

	return nil
}

// GetSureBackupStats retrieves tenant-specific statistics
func (m *InMemorySureBackupManager) GetSureBackupStats(ctx context.Context, tenantID string) (*SureBackupStats, error) {
	if err := m.validateTenantAccess(ctx, tenantID); err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := &SureBackupStats{
		TenantID:             tenantID,
		TotalSandboxes:       0,
		RunningSandboxes:     0,
		TotalVerifications:   0,
		RunningVerifications: 0,
		PassedVerifications:  0,
		FailedVerifications:  0,
		LastUpdated:          time.Now(),
	}

	for _, sandbox := range m.sandboxes {
		if sandbox.TenantID == tenantID {
			stats.TotalSandboxes++
			if sandbox.Status == SandboxStatusRunning {
				stats.RunningSandboxes++
			}
		}
	}

	for _, verification := range m.verifications {
		if verification.TenantID == tenantID {
			stats.TotalVerifications++
			if verification.Status == VerificationStatusRunning {
				stats.RunningVerifications++
			}
			if verification.Status == VerificationStatusPassed {
				stats.PassedVerifications++
			}
			if verification.Status == VerificationStatusFailed {
				stats.FailedVerifications++
			}
		}
	}

	return stats, nil
}

// GetGlobalStats retrieves global statistics
func (m *InMemorySureBackupManager) GetGlobalStats(ctx context.Context) (*GlobalSureBackupStats, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := &GlobalSureBackupStats{
		TotalTenants:         0,
		TotalSandboxes:       0,
		RunningSandboxes:     0,
		TotalVerifications:   0,
		RunningVerifications: 0,
		PassedVerifications:  0,
		FailedVerifications:  0,
		LastUpdated:          time.Now(),
	}

	tenants := make(map[string]bool)
	for _, sandbox := range m.sandboxes {
		tenants[sandbox.TenantID] = true
		stats.TotalSandboxes++
		if sandbox.Status == SandboxStatusRunning {
			stats.RunningSandboxes++
		}
	}

	for _, verification := range m.verifications {
		tenants[verification.TenantID] = true
		stats.TotalVerifications++
		if verification.Status == VerificationStatusRunning {
			stats.RunningVerifications++
		}
		if verification.Status == VerificationStatusPassed {
			stats.PassedVerifications++
		}
		if verification.Status == VerificationStatusFailed {
			stats.FailedVerifications++
		}
	}

	stats.TotalTenants = int64(len(tenants))

	return stats, nil
}

// Helper methods

func (m *InMemorySureBackupManager) startSandboxInternal(ctx context.Context, sandbox *Sandbox) error {
	sandbox.Status = SandboxStatusRunning
	now := time.Now()
	sandbox.StartedAt = &now
	return nil
}

func (m *InMemorySureBackupManager) stopSandboxInternal(ctx context.Context, sandbox *Sandbox) {
	sandbox.Status = SandboxStatusStopped
	now := time.Now()
	sandbox.StoppedAt = &now
}

func (m *InMemorySureBackupManager) runVerification(ctx context.Context, verification *Verification) {
	m.mutex.Lock()
	verification.Status = VerificationStatusRunning
	now := time.Now()
	verification.StartedAt = &now
	m.mutex.Unlock()

	// Simulate verification process
	time.Sleep(5 * time.Second)

	m.mutex.Lock()
	verification.Status = VerificationStatusPassed
	completedAt := time.Now()
	verification.CompletedAt = &completedAt

	// Set mock results
	verification.Results = VerificationResults{
		Passed:   true,
		Duration: completedAt.Sub(*verification.StartedAt),
		TestResults: []TestResult{
			{
				Name:     "integrity_check",
				Type:     "integrity",
				Passed:   true,
				Duration: 2 * time.Second,
				Output:   "All files verified successfully",
			},
		},
		Summary: VerificationSummary{
			TotalTests:      1,
			PassedTests:     1,
			FailedTests:     0,
			DataIntegrityOK: true,
			MountSuccess:    true,
			BootSuccess:     true,
		},
	}
	m.mutex.Unlock()
}

func (m *InMemorySureBackupManager) validateTenantAccess(ctx context.Context, tenantID string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant ID cannot be empty")
	}

	// Get tenant from context
	contextTenantID := m.tenantManager.GetTenantFromContext(ctx)
	if contextTenantID == "" {
		return fmt.Errorf("tenant context not found")
	}

	if contextTenantID != tenantID {
		return fmt.Errorf("tenant access denied")
	}

	return nil
}

func generateSandboxID() string {
	return fmt.Sprintf("sandbox-%s", uuid.New().String()[:8])
}

func generateVerificationID() string {
	return fmt.Sprintf("verif-%s", uuid.New().String()[:8])
}
