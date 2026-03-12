package guest

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// GuestProcessingStatus represents the status of guest processing
type GuestProcessingStatus string

const (
	GuestProcessingStatusPending    GuestProcessingStatus = "pending"
	GuestProcessingStatusInProgress GuestProcessingStatus = "in_progress"
	GuestProcessingStatusSuccess    GuestProcessingStatus = "success"
	GuestProcessingStatusFailed     GuestProcessingStatus = "failed"
	GuestProcessingStatusSkipped    GuestProcessingStatus = "skipped"
)

// GuestProcessingType represents the type of guest processing
type GuestProcessingType string

const (
	GuestProcessingTypeVSSFull     GuestProcessingType = "vss_full"
	GuestProcessingTypeVSSCopy     GuestProcessingType = "vss_copy"
	GuestProcessingTypeVSSIncremental GuestProcessingType = "vss_incremental"
	GuestProcessingTypeScript      GuestProcessingType = "script"
	GuestProcessingTypeCommand     GuestProcessingType = "command"
)

// VSSWriterType represents the type of VSS writer
type VSSWriterType string

const (
	VSSWriterSystem      VSSWriterType = "system"
	VSSWriterSQLServer   VSSWriterType = "sql_server"
	VSSWriterExchange    VSSWriterType = "exchange"
	VSSWriterActiveDir   VSSWriterType = "active_directory"
	VSSWriterSharePoint  VSSWriterType = "sharepoint"
	VSSWriterHyperV      VSSWriterType = "hyperv"
	VSSWriterRegistry    VSSWriterType = "registry"
	VSSWriterCOMPlus     VSSWriterType = "com_plus"
	VSSWriterCertificate VSSWriterType = "certificate"
	VSSWriterIIS         VSSWriterType = "iis"
	VSSWriterWINS        VSSWriterType = "wins"
	VSSWriterDHCP        VSSWriterType = "dhcp"
)

// VSSWriterStatus represents the status of a VSS writer
type VSSWriterStatus struct {
	WriterType VSSWriterType `json:"writer_type"`
	Name       string        `json:"name"`
	State      string        `json:"state"` // "Stable", "Failed", "WaitingForCompletion", etc.
	LastError  string        `json:"last_error,omitempty"`
	LastSeen   time.Time     `json:"last_seen"`
	IsHealthy  bool          `json:"is_healthy"`
}

// VSSSnapshot represents a VSS snapshot
type VSSSnapshot struct {
	ID            string            `json:"id"`
	Volume        string            `json:"volume"`
	WriterTypes   []VSSWriterType   `json:"writer_types"`
	BackupType    string            `json:"backup_type"` // "Full", "Copy", "Incremental", "Differential", "Log"
	CreatedAt     time.Time         `json:"created_at"`
	ExpiresAt     time.Time         `json:"expires_at,omitempty"`
	Size          int64             `json:"size"`
	Status        string            `json:"status"`
	SnapshotSetID string            `json:"snapshot_set_id"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// GuestProcessingTask represents a task for guest processing
type GuestProcessingTask struct {
	ID                    string                `json:"id"`
	VMID                  string                `json:"vm_id"`
	VMName                string                `json:"vm_name"`
	JobID                 string                `json:"job_id"`
	ProcessingType        GuestProcessingType   `json:"processing_type"`
	CredentialID          string                `json:"credential_id"`
	EnableVSS             bool                  `json:"enable_vss"`
	VSSWriterTypes        []VSSWriterType       `json:"vss_writer_types"`
	SkipFailedWriters     bool                  `json:"skip_failed_writers"`
	TruncateLogs          bool                  `json:"truncate_logs"`
	PreBackupScript       string                `json:"pre_backup_script,omitempty"`
	PostBackupScript      string                `json:"post_backup_script,omitempty"`
	Status                GuestProcessingStatus `json:"status"`
	Progress              int                   `json:"progress"`
	ErrorMessage          string                `json:"error_message,omitempty"`
	VSSWriterStatuses     []*VSSWriterStatus    `json:"vss_writer_statuses,omitempty"`
	VSSSnapshot           *VSSSnapshot          `json:"vss_snapshot,omitempty"`
	StartedAt             time.Time             `json:"started_at,omitempty"`
	CompletedAt           time.Time             `json:"completed_at,omitempty"`
	CreatedAt             time.Time             `json:"created_at"`
	UpdatedAt             time.Time             `json:"updated_at"`
	ApplicationProcessing map[string]string     `json:"application_processing,omitempty"`
}

// GuestProcessingConfig holds configuration for guest processing
type GuestProcessingConfig struct {
	DefaultVSSBackupType    string        `json:"default_vss_backup_type"`
	VSSSnapshotTimeout      time.Duration `json:"vss_snapshot_timeout"`
	VSSSnapshotExpiry       time.Duration `json:"vss_snapshot_expiry"`
	MaxConcurrentProcessing int           `json:"max_concurrent_processing"`
	Logger                  *zap.Logger   `json:"-"`
}

// DefaultGuestProcessingConfig returns default configuration
func DefaultGuestProcessingConfig() GuestProcessingConfig {
	return GuestProcessingConfig{
		DefaultVSSBackupType:    "Full",
		VSSSnapshotTimeout:      30 * time.Minute,
		VSSSnapshotExpiry:       1 * time.Hour,
		MaxConcurrentProcessing: 5,
		Logger:                  zap.NewNop(),
	}
}

// GuestProcessor handles guest processing operations including VSS
type GuestProcessor struct {
	mu              sync.RWMutex
	tasks           map[string]*GuestProcessingTask
	vssSnapshots    map[string]*VSSSnapshot
	config          GuestProcessingConfig
	credentialStore CredentialStore
	logger          *zap.Logger
	onTaskUpdate    func(task *GuestProcessingTask)
	semaphore       chan struct{}
}

// GuestProcessorConfig holds configuration for the processor
type GuestProcessorConfig struct {
	Config          GuestProcessingConfig
	CredentialStore CredentialStore
	Logger          *zap.Logger
	OnTaskUpdate    func(task *GuestProcessingTask)
}

// NewGuestProcessor creates a new guest processor
func NewGuestProcessor(config GuestProcessorConfig) *GuestProcessor {
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}
	if config.Config.Logger == nil {
		config.Config.Logger = config.Logger
	}

	processor := &GuestProcessor{
		tasks:           make(map[string]*GuestProcessingTask),
		vssSnapshots:    make(map[string]*VSSSnapshot),
		config:          config.Config,
		credentialStore: config.CredentialStore,
		logger:          config.Logger,
		onTaskUpdate:    config.OnTaskUpdate,
		semaphore:       make(chan struct{}, config.Config.MaxConcurrentProcessing),
	}

	processor.logger.Info("GuestProcessor initialized",
		zap.String("default_vss_backup_type", config.Config.DefaultVSSBackupType),
		zap.Duration("vss_snapshot_timeout", config.Config.VSSSnapshotTimeout),
		zap.Int("max_concurrent_processing", config.Config.MaxConcurrentProcessing))

	return processor
}

// CreateProcessingTask creates a new guest processing task
func (p *GuestProcessor) CreateProcessingTask(vmID, vmName, jobID string, credentialID string, processingType GuestProcessingType) (*GuestProcessingTask, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	taskID := fmt.Sprintf("process-%s-%d", vmID, time.Now().UnixNano())

	task := &GuestProcessingTask{
		ID:                    taskID,
		VMID:                  vmID,
		VMName:                vmName,
		JobID:                 jobID,
		ProcessingType:        processingType,
		CredentialID:          credentialID,
		EnableVSS:             processingType == GuestProcessingTypeVSSFull || processingType == GuestProcessingTypeVSSCopy,
		VSSWriterTypes:        []VSSWriterType{VSSWriterSystem},
		SkipFailedWriters:     false,
		TruncateLogs:          true,
		Status:                GuestProcessingStatusPending,
		Progress:              0,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
		ApplicationProcessing: make(map[string]string),
	}

	// Add default VSS writers based on processing type
	switch processingType {
	case GuestProcessingTypeVSSFull:
		task.VSSWriterTypes = []VSSWriterType{
			VSSWriterSystem,
			VSSWriterSQLServer,
			VSSWriterExchange,
			VSSWriterActiveDir,
			VSSWriterHyperV,
		}
	case GuestProcessingTypeVSSCopy:
		task.VSSWriterTypes = []VSSWriterType{
			VSSWriterSystem,
		}
	}

	p.tasks[taskID] = task

	p.logger.Info("Guest processing task created",
		zap.String("task_id", taskID),
		zap.String("vm_id", vmID),
		zap.String("vm_name", vmName),
		zap.String("type", string(processingType)))

	return task, nil
}

// GetProcessingTask retrieves a processing task by ID
func (p *GuestProcessor) GetProcessingTask(taskID string) (*GuestProcessingTask, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	task, exists := p.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("processing task %s not found", taskID)
	}

	taskCopy := *task
	if task.VSSWriterTypes != nil {
		taskCopy.VSSWriterTypes = make([]VSSWriterType, len(task.VSSWriterTypes))
		copy(taskCopy.VSSWriterTypes, task.VSSWriterTypes)
	}
	if task.VSSWriterStatuses != nil {
		taskCopy.VSSWriterStatuses = make([]*VSSWriterStatus, len(task.VSSWriterStatuses))
		for i, s := range task.VSSWriterStatuses {
			sCopy := *s
			taskCopy.VSSWriterStatuses[i] = &sCopy
		}
	}
	if task.ApplicationProcessing != nil {
		taskCopy.ApplicationProcessing = make(map[string]string)
		for k, v := range task.ApplicationProcessing {
			taskCopy.ApplicationProcessing[k] = v
		}
	}

	return &taskCopy, nil
}

// ListProcessingTasks returns all processing tasks
func (p *GuestProcessor) ListProcessingTasks(statusFilter ...GuestProcessingStatus) []*GuestProcessingTask {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var tasks []*GuestProcessingTask

	for _, task := range p.tasks {
		if len(statusFilter) == 0 {
			taskCopy := *task
			tasks = append(tasks, &taskCopy)
			continue
		}

		for _, status := range statusFilter {
			if task.Status == status {
				taskCopy := *task
				tasks = append(tasks, &taskCopy)
				break
			}
		}
	}

	return tasks
}

// ExecuteProcessing executes a guest processing task
func (p *GuestProcessor) ExecuteProcessing(ctx context.Context, taskID string) error {
	p.mu.Lock()
	task, exists := p.tasks[taskID]
	if !exists {
		p.mu.Unlock()
		return fmt.Errorf("processing task %s not found", taskID)
	}

	if task.Status != GuestProcessingStatusPending {
		p.mu.Unlock()
		return fmt.Errorf("task %s is not in pending status (current: %s)", taskID, task.Status)
	}

	task.Status = GuestProcessingStatusInProgress
	task.StartedAt = time.Now()
	task.UpdatedAt = time.Now()
	p.mu.Unlock()

	p.logger.Info("Starting guest processing",
		zap.String("task_id", taskID),
		zap.String("vm_id", task.VMID),
		zap.String("type", string(task.ProcessingType)))

	// Acquire semaphore for concurrent processing limit
	select {
	case p.semaphore <- struct{}{}:
		defer func() { <-p.semaphore }()
	case <-ctx.Done():
		return ctx.Err()
	}

	go p.executeProcessingAsync(ctx, task)

	return nil
}

// executeProcessingAsync executes processing asynchronously
func (p *GuestProcessor) executeProcessingAsync(ctx context.Context, task *GuestProcessingTask) {
	defer func() {
		if r := recover(); r != nil {
			p.mu.Lock()
			task.Status = GuestProcessingStatusFailed
			task.ErrorMessage = fmt.Sprintf("Panic during processing: %v", r)
			task.CompletedAt = time.Now()
			task.UpdatedAt = time.Now()
			p.mu.Unlock()
			p.logger.Error("Processing panic",
				zap.String("task_id", task.ID),
				zap.Any("panic", r))
			p.notifyTaskUpdate(task)
		}
	}()

	var err error

	switch task.ProcessingType {
	case GuestProcessingTypeVSSFull, GuestProcessingTypeVSSCopy, GuestProcessingTypeVSSIncremental:
		err = p.processWithVSS(ctx, task)
	case GuestProcessingTypeScript:
		err = p.processWithScript(ctx, task)
	case GuestProcessingTypeCommand:
		err = p.processWithCommand(ctx, task)
	default:
		err = fmt.Errorf("unsupported processing type: %s", task.ProcessingType)
	}

	p.mu.Lock()
	task.UpdatedAt = time.Now()
	if err != nil {
		task.Status = GuestProcessingStatusFailed
		task.ErrorMessage = err.Error()
		task.Progress = 0
		p.logger.Error("Guest processing failed",
			zap.String("task_id", task.ID),
			zap.String("vm_id", task.VMID),
			zap.Error(err))
	} else {
		task.Status = GuestProcessingStatusSuccess
		task.Progress = 100
		p.logger.Info("Guest processing successful",
			zap.String("task_id", task.ID),
			zap.String("vm_id", task.VMID))
	}

	task.CompletedAt = time.Now()
	p.mu.Unlock()
	p.notifyTaskUpdate(task)
}

// processWithVSS performs VSS-based guest processing
func (p *GuestProcessor) processWithVSS(ctx context.Context, task *GuestProcessingTask) error {
	p.updateProgress(task, 10, "Initializing VSS processing...")

	// Get credentials
	cred, err := p.getCredentials(task.CredentialID)
	if err != nil {
		return fmt.Errorf("failed to get credentials: %w", err)
	}

	p.updateProgress(task, 20, "Enumerating VSS writers...")

	// Enumerate and check VSS writers
	writerStatuses, err := p.enumerateVSSWriters(ctx, task.VMID, task.VSSWriterTypes)
	if err != nil {
		if task.SkipFailedWriters {
			p.logger.Warn("VSS writer enumeration failed, continuing with skip",
				zap.String("task_id", task.ID),
				zap.Error(err))
			task.ApplicationProcessing["vss_warning"] = err.Error()
		} else {
			return fmt.Errorf("VSS writer enumeration failed: %w", err)
		}
	}
	task.VSSWriterStatuses = writerStatuses

	p.updateProgress(task, 40, "Creating VSS snapshot...")

	// Create VSS snapshot
	snapshot, err := p.createVSSSnapshot(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to create VSS snapshot: %w", err)
	}
	task.VSSSnapshot = snapshot

	// Store snapshot for cleanup
	p.mu.Lock()
	p.vssSnapshots[snapshot.ID] = snapshot
	p.mu.Unlock()

	p.updateProgress(task, 60, "Freezing VSS writers...")

	// Simulate VSS freeze
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(1 * time.Second):
	}

	p.updateProgress(task, 70, "Performing backup (VSS-consistent)...")

	// Simulate backup operation
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(3 * time.Second):
	}

	p.updateProgress(task, 85, "Thawing VSS writers...")

	// Simulate VSS thaw
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(1 * time.Second):
	}

	if task.TruncateLogs {
		p.updateProgress(task, 90, "Truncating application logs...")
		err = p.truncateApplicationLogs(ctx, task)
		if err != nil {
			p.logger.Warn("Failed to truncate logs",
				zap.String("task_id", task.ID),
				zap.Error(err))
			task.ApplicationProcessing["log_truncation_warning"] = err.Error()
		}
	}

	p.updateProgress(task, 100, "VSS processing complete")

	p.logger.Info("VSS processing completed",
		zap.String("task_id", task.ID),
		zap.String("vm_id", task.VMID),
		zap.String("username", cred.Username),
		zap.Int("writers_processed", len(writerStatuses)))

	return nil
}

// processWithScript performs script-based guest processing
func (p *GuestProcessor) processWithScript(ctx context.Context, task *GuestProcessingTask) error {
	p.updateProgress(task, 10, "Preparing script execution...")

	if task.PreBackupScript == "" && task.PostBackupScript == "" {
		return fmt.Errorf("no scripts configured for script processing")
	}

	// Execute pre-backup script
	if task.PreBackupScript != "" {
		p.updateProgress(task, 20, "Executing pre-backup script...")
		err := p.executeGuestScript(ctx, task, task.PreBackupScript)
		if err != nil {
			return fmt.Errorf("pre-backup script failed: %w", err)
		}
	}

	p.updateProgress(task, 50, "Backup window (script processing complete)...")

	// Simulate backup
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(2 * time.Second):
	}

	// Execute post-backup script
	if task.PostBackupScript != "" {
		p.updateProgress(task, 80, "Executing post-backup script...")
		err := p.executeGuestScript(ctx, task, task.PostBackupScript)
		if err != nil {
			return fmt.Errorf("post-backup script failed: %w", err)
		}
	}

	p.updateProgress(task, 100, "Script processing complete")

	p.logger.Info("Script processing completed",
		zap.String("task_id", task.ID),
		zap.String("vm_id", task.VMID))

	return nil
}

// processWithCommand performs command-based guest processing
func (p *GuestProcessor) processWithCommand(ctx context.Context, task *GuestProcessingTask) error {
	p.updateProgress(task, 10, "Preparing command execution...")

	// Execute command on guest
	p.updateProgress(task, 30, "Executing guest command...")

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(2 * time.Second):
	}

	p.updateProgress(task, 70, "Processing command output...")

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(1 * time.Second):
	}

	p.updateProgress(task, 100, "Command processing complete")

	p.logger.Info("Command processing completed",
		zap.String("task_id", task.ID),
		zap.String("vm_id", task.VMID))

	return nil
}

// enumerateVSSWriters enumerates VSS writers on the guest
func (p *GuestProcessor) enumerateVSSWriters(ctx context.Context, vmID string, writerTypes []VSSWriterType) ([]*VSSWriterStatus, error) {
	var statuses []*VSSWriterStatus

	for _, wt := range writerTypes {
		status := &VSSWriterStatus{
			WriterType: wt,
			Name:       string(wt),
			State:      "Stable",
			LastSeen:   time.Now(),
			IsHealthy:  true,
		}

		// Simulate writer status check
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}

		statuses = append(statuses, status)
	}

	return statuses, nil
}

// createVSSSnapshot creates a VSS snapshot
func (p *GuestProcessor) createVSSSnapshot(ctx context.Context, task *GuestProcessingTask) (*VSSSnapshot, error) {
	snapshotID := fmt.Sprintf("snap-%s-%d", task.VMID, time.Now().UnixNano())

	snapshot := &VSSSnapshot{
		ID:            snapshotID,
		Volume:        "C:",
		WriterTypes:   task.VSSWriterTypes,
		BackupType:    p.config.DefaultVSSBackupType,
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(p.config.VSSSnapshotExpiry),
		Size:          0,
		Status:        "active",
		SnapshotSetID: fmt.Sprintf("set-%d", time.Now().UnixNano()),
		Metadata: map[string]string{
			"vm_id":    task.VMID,
			"task_id":  task.ID,
			"job_id":   task.JobID,
		},
	}

	// Simulate snapshot creation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(2 * time.Second):
	}

	p.logger.Info("VSS snapshot created",
		zap.String("snapshot_id", snapshotID),
		zap.String("vm_id", task.VMID),
		zap.String("backup_type", snapshot.BackupType))

	return snapshot, nil
}

// truncateApplicationLogs truncates application transaction logs
func (p *GuestProcessor) truncateApplicationLogs(ctx context.Context, task *GuestProcessingTask) error {
	// Simulate log truncation for various applications
	for _, writerType := range task.VSSWriterTypes {
		switch writerType {
		case VSSWriterSQLServer:
			task.ApplicationProcessing["sql_log_truncation"] = "success"
		case VSSWriterExchange:
			task.ApplicationProcessing["exchange_log_truncation"] = "success"
		}
	}

	return nil
}

// executeGuestScript executes a script on the guest
func (p *GuestProcessor) executeGuestScript(ctx context.Context, task *GuestProcessingTask, script string) error {
	// Simulate script execution
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(1 * time.Second):
	}

	return nil
}

// DeleteVSSSnapshot deletes a VSS snapshot
func (p *GuestProcessor) DeleteVSSSnapshot(snapshotID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	snapshot, exists := p.vssSnapshots[snapshotID]
	if !exists {
		return fmt.Errorf("snapshot %s not found", snapshotID)
	}

	snapshot.Status = "deleted"
	delete(p.vssSnapshots, snapshotID)

	p.logger.Info("VSS snapshot deleted", zap.String("snapshot_id", snapshotID))
	return nil
}

// CleanupExpiredSnapshots cleans up expired VSS snapshots
func (p *GuestProcessor) CleanupExpiredSnapshots() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	cleaned := 0

	for id, snapshot := range p.vssSnapshots {
		if !snapshot.ExpiresAt.IsZero() && now.After(snapshot.ExpiresAt) {
			snapshot.Status = "expired"
			delete(p.vssSnapshots, id)
			cleaned++
			p.logger.Info("Expired snapshot cleaned up",
				zap.String("snapshot_id", id),
				zap.Time("expired_at", snapshot.ExpiresAt))
		}
	}

	return cleaned
}

// CancelProcessing cancels a processing task
func (p *GuestProcessor) CancelProcessing(taskID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	task, exists := p.tasks[taskID]
	if !exists {
		return fmt.Errorf("processing task %s not found", taskID)
	}

	if task.Status == GuestProcessingStatusSuccess ||
		task.Status == GuestProcessingStatusFailed ||
		task.Status == GuestProcessingStatusSkipped {
		return fmt.Errorf("cannot cancel task in status %s", task.Status)
	}

	task.Status = GuestProcessingStatusFailed
	task.ErrorMessage = "Cancelled by user"
	task.UpdatedAt = time.Now()
	task.CompletedAt = time.Now()

	p.logger.Info("Processing cancelled", zap.String("task_id", taskID))
	p.notifyTaskUpdate(task)

	return nil
}

// GetProcessingStats returns statistics about processing tasks
func (p *GuestProcessor) GetProcessingStats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	statusCounts := make(map[GuestProcessingStatus]int)
	typeCounts := make(map[GuestProcessingType]int)

	for _, task := range p.tasks {
		statusCounts[task.Status]++
		typeCounts[task.ProcessingType]++
	}

	successRate := 0.0
	if len(p.tasks) > 0 {
		successRate = float64(statusCounts[GuestProcessingStatusSuccess]) / float64(len(p.tasks)) * 100
	}

	return map[string]interface{}{
		"total_tasks":     len(p.tasks),
		"status_counts":   statusCounts,
		"type_counts":     typeCounts,
		"success_rate":    successRate,
		"active_snapshots": len(p.vssSnapshots),
	}
}

// SetTaskUpdateCallback sets a callback for task updates
func (p *GuestProcessor) SetTaskUpdateCallback(callback func(task *GuestProcessingTask)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.onTaskUpdate = callback
}

// updateProgress updates the progress of a task
func (p *GuestProcessor) updateProgress(task *GuestProcessingTask, progress int, message string) {
	p.mu.Lock()
	task.Progress = progress
	task.UpdatedAt = time.Now()
	p.mu.Unlock()

	p.logger.Debug("Processing progress",
		zap.String("task_id", task.ID),
		zap.Int("progress", progress),
		zap.String("message", message))

	p.notifyTaskUpdate(task)
}

// notifyTaskUpdate notifies listeners of task updates
func (p *GuestProcessor) notifyTaskUpdate(task *GuestProcessingTask) {
	p.mu.RLock()
	callback := p.onTaskUpdate
	p.mu.RUnlock()

	if callback != nil {
		go callback(task)
	}
}

// getCredentials retrieves credentials from the store
func (p *GuestProcessor) getCredentials(credentialID string) (*GuestCredential, error) {
	if p.credentialStore == nil {
		return &GuestCredential{
			Username: "admin",
			Password: "password",
		}, nil
	}

	return p.credentialStore.Get(credentialID)
}

// GetVSSWriterTypes returns available VSS writer types
func (p *GuestProcessor) GetVSSWriterTypes() []VSSWriterType {
	return []VSSWriterType{
		VSSWriterSystem,
		VSSWriterSQLServer,
		VSSWriterExchange,
		VSSWriterActiveDir,
		VSSWriterSharePoint,
		VSSWriterHyperV,
		VSSWriterRegistry,
		VSSWriterCOMPlus,
		VSSWriterCertificate,
		VSSWriterIIS,
		VSSWriterWINS,
		VSSWriterDHCP,
	}
}

// GetVSSWriterName returns a human-readable name for a VSS writer type
func (p *GuestProcessor) GetVSSWriterName(writerType VSSWriterType) string {
	names := map[VSSWriterType]string{
		VSSWriterSystem:      "System State",
		VSSWriterSQLServer:   "SQL Server",
		VSSWriterExchange:    "Microsoft Exchange",
		VSSWriterActiveDir:   "Active Directory",
		VSSWriterSharePoint:  "SharePoint",
		VSSWriterHyperV:      "Hyper-V",
		VSSWriterRegistry:    "Registry",
		VSSWriterCOMPlus:     "COM+ Class Registration",
		VSSWriterCertificate: "Certificate Services",
		VSSWriterIIS:         "IIS Metabase",
		VSSWriterWINS:        "WINS",
		VSSWriterDHCP:        "DHCP",
	}

	if name, ok := names[writerType]; ok {
		return name
	}
	return string(writerType)
}

// Shutdown gracefully shuts down the processor
func (p *GuestProcessor) Shutdown(ctx context.Context) error {
	p.logger.Info("Shutting down GuestProcessor")

	// Clean up expired snapshots
	p.CleanupExpiredSnapshots()

	// Wait for active tasks to complete or timeout
	p.mu.RLock()
	activeCount := 0
	for _, task := range p.tasks {
		if task.Status == GuestProcessingStatusInProgress {
			activeCount++
		}
	}
	p.mu.RUnlock()

	if activeCount > 0 {
		p.logger.Info("Waiting for active tasks to complete",
			zap.Int("active_count", activeCount))

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Minute):
			p.logger.Warn("Timeout waiting for active tasks")
		}
	}

	p.logger.Info("GuestProcessor shutdown complete")
	return nil
}
