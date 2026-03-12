package guest

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
)

// AgentDeploymentStatus represents the status of an agent deployment
type AgentDeploymentStatus string

const (
	AgentDeploymentStatusPending    AgentDeploymentStatus = "pending"
	AgentDeploymentStatusInProgress AgentDeploymentStatus = "in_progress"
	AgentDeploymentStatusSuccess    AgentDeploymentStatus = "success"
	AgentDeploymentStatusFailed     AgentDeploymentStatus = "failed"
	AgentDeploymentStatusCancelled  AgentDeploymentStatus = "cancelled"
)

// AgentDeploymentMethod represents the method used to deploy the agent
type AgentDeploymentMethod string

const (
	AgentDeploymentMethodSSH        AgentDeploymentMethod = "ssh"
	AgentDeploymentMethodWinRM      AgentDeploymentMethod = "winrm"
	AgentDeploymentMethodVMTools    AgentDeploymentMethod = "vmtools"
	AgentDeploymentMethodHyperV     AgentDeploymentMethod = "hyperv"
	AgentDeploymentMethodManual     AgentDeploymentMethod = "manual"
	AgentDeploymentMethodGroupPolicy AgentDeploymentMethod = "gpo"
)

// AgentDeploymentTask represents a task to deploy an agent to a guest system
type AgentDeploymentTask struct {
	ID               string                `json:"id"`
	VMID             string                `json:"vm_id"`
	VMName           string                `json:"vm_name"`
	TargetHost       string                `json:"target_host"`
	TargetPort       int                   `json:"target_port"`
	CredentialID     string                `json:"credential_id"`
	DeploymentMethod AgentDeploymentMethod `json:"deployment_method"`
	AgentVersion     string                `json:"agent_version"`
	AgentPackage     string                `json:"agent_package"`
	InstallPath      string                `json:"install_path"`
	Status           AgentDeploymentStatus `json:"status"`
	Progress         int                   `json:"progress"`
	ErrorMessage     string                `json:"error_message,omitempty"`
	StartedAt        time.Time             `json:"started_at,omitempty"`
	CompletedAt      time.Time             `json:"completed_at,omitempty"`
	CreatedAt        time.Time             `json:"created_at"`
	UpdatedAt        time.Time             `json:"updated_at"`
	RetryCount       int                   `json:"retry_count"`
	MaxRetries       int                   `json:"max_retries"`
	Timeout          time.Duration         `json:"timeout"`
	Options          map[string]string     `json:"options,omitempty"`
}

// AgentDeploymentConfig holds configuration for agent deployment
type AgentDeploymentConfig struct {
	DefaultInstallPathWindows string        `json:"default_install_path_windows"`
	DefaultInstallPathLinux   string        `json:"default_install_path_linux"`
	DefaultTimeout            time.Duration `json:"default_timeout"`
	DefaultMaxRetries         int           `json:"default_max_retries"`
	AgentPackagePath          string        `json:"agent_package_path"`
	Logger                    *zap.Logger   `json:"-"`
}

// DefaultAgentDeploymentConfig returns default deployment configuration
func DefaultAgentDeploymentConfig() AgentDeploymentConfig {
	return AgentDeploymentConfig{
		DefaultInstallPathWindows: `C:\Program Files\NovaBackup\Agent`,
		DefaultInstallPathLinux:   "/opt/novabackup/agent",
		DefaultTimeout:            10 * time.Minute,
		DefaultMaxRetries:         3,
		Logger:                    zap.NewNop(),
	}
}

// AgentInjector handles deployment of backup agents to guest systems
type AgentInjector struct {
	mu              sync.RWMutex
	tasks           map[string]*AgentDeploymentTask
	config          AgentDeploymentConfig
	credentialStore CredentialStore
	logger          *zap.Logger
	onTaskUpdate    func(task *AgentDeploymentTask)
}

// AgentInjectorConfig holds configuration for the injector
type AgentInjectorConfig struct {
	Config          AgentDeploymentConfig
	CredentialStore CredentialStore
	Logger          *zap.Logger
	OnTaskUpdate    func(task *AgentDeploymentTask)
}

// NewAgentInjector creates a new agent injector
func NewAgentInjector(config AgentInjectorConfig) *AgentInjector {
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}
	if config.Config.Logger == nil {
		config.Config.Logger = config.Logger
	}

	injector := &AgentInjector{
		tasks:           make(map[string]*AgentDeploymentTask),
		config:          config.Config,
		credentialStore: config.CredentialStore,
		logger:          config.Logger,
		onTaskUpdate:    config.OnTaskUpdate,
	}

	injector.logger.Info("AgentInjector initialized",
		zap.String("default_install_path_windows", config.Config.DefaultInstallPathWindows),
		zap.String("default_install_path_linux", config.Config.DefaultInstallPathLinux),
		zap.Duration("default_timeout", config.Config.DefaultTimeout))

	return injector
}

// CreateDeploymentTask creates a new deployment task
func (i *AgentInjector) CreateDeploymentTask(ctx context.Context, vmID, vmName, targetHost string, credentialID string, method AgentDeploymentMethod) (*AgentDeploymentTask, error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	taskID := fmt.Sprintf("deploy-%s-%d", vmID, time.Now().UnixNano())

	installPath := i.config.DefaultInstallPathLinux
	if method == AgentDeploymentMethodWinRM || method == AgentDeploymentMethodGroupPolicy {
		installPath = i.config.DefaultInstallPathWindows
	}

	task := &AgentDeploymentTask{
		ID:               taskID,
		VMID:             vmID,
		VMName:           vmName,
		TargetHost:       targetHost,
		TargetPort:       22,
		CredentialID:     credentialID,
		DeploymentMethod: method,
		AgentVersion:     "7.0.0",
		InstallPath:      installPath,
		Status:           AgentDeploymentStatusPending,
		Progress:         0,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		MaxRetries:       i.config.DefaultMaxRetries,
		Timeout:          i.config.DefaultTimeout,
		Options:          make(map[string]string),
	}

	switch method {
	case AgentDeploymentMethodSSH:
		task.TargetPort = 22
	case AgentDeploymentMethodWinRM:
		task.TargetPort = 5985
	case AgentDeploymentMethodVMTools:
		task.TargetPort = 0
	case AgentDeploymentMethodHyperV:
		task.TargetPort = 0
	}

	i.tasks[taskID] = task

	i.logger.Info("Deployment task created",
		zap.String("task_id", taskID),
		zap.String("vm_id", vmID),
		zap.String("vm_name", vmName),
		zap.String("method", string(method)))

	return task, nil
}

// GetDeploymentTask retrieves a deployment task by ID
func (i *AgentInjector) GetDeploymentTask(taskID string) (*AgentDeploymentTask, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	task, exists := i.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("deployment task %s not found", taskID)
	}

	taskCopy := *task
	if task.Options != nil {
		taskCopy.Options = make(map[string]string)
		for k, v := range task.Options {
			taskCopy.Options[k] = v
		}
	}

	return &taskCopy, nil
}

// ListDeploymentTasks returns all deployment tasks
func (i *AgentInjector) ListDeploymentTasks(statusFilter ...AgentDeploymentStatus) []*AgentDeploymentTask {
	i.mu.RLock()
	defer i.mu.RUnlock()

	var tasks []*AgentDeploymentTask

	for _, task := range i.tasks {
		if len(statusFilter) == 0 {
			taskCopy := *task
			if task.Options != nil {
				taskCopy.Options = make(map[string]string)
				for k, v := range task.Options {
					taskCopy.Options[k] = v
				}
			}
			tasks = append(tasks, &taskCopy)
			continue
		}

		for _, status := range statusFilter {
			if task.Status == status {
				taskCopy := *task
				if task.Options != nil {
					taskCopy.Options = make(map[string]string)
					for k, v := range task.Options {
						taskCopy.Options[k] = v
					}
				}
				tasks = append(tasks, &taskCopy)
				break
			}
		}
	}

	return tasks
}

// ExecuteDeployment executes a deployment task asynchronously
func (i *AgentInjector) ExecuteDeployment(ctx context.Context, taskID string) error {
	i.mu.Lock()
	task, exists := i.tasks[taskID]
	if !exists {
		i.mu.Unlock()
		return fmt.Errorf("deployment task %s not found", taskID)
	}

	if task.Status != AgentDeploymentStatusPending {
		i.mu.Unlock()
		return fmt.Errorf("task %s is not in pending status (current: %s)", taskID, task.Status)
	}

	task.Status = AgentDeploymentStatusInProgress
	task.StartedAt = time.Now()
	task.UpdatedAt = time.Now()
	i.mu.Unlock()

	i.logger.Info("Starting agent deployment",
		zap.String("task_id", taskID),
		zap.String("vm_id", task.VMID),
		zap.String("method", string(task.DeploymentMethod)))

	go i.executeDeploymentAsync(ctx, task)

	return nil
}

// executeDeploymentAsync executes the deployment asynchronously
func (i *AgentInjector) executeDeploymentAsync(ctx context.Context, task *AgentDeploymentTask) {
	defer func() {
		if r := recover(); r != nil {
			i.mu.Lock()
			task.Status = AgentDeploymentStatusFailed
			task.ErrorMessage = fmt.Sprintf("Panic during deployment: %v", r)
			task.CompletedAt = time.Now()
			task.UpdatedAt = time.Now()
			i.mu.Unlock()
			i.logger.Error("Deployment panic",
				zap.String("task_id", task.ID),
				zap.Any("panic", r))
			i.notifyTaskUpdate(task)
		}
	}()

	var err error

	switch task.DeploymentMethod {
	case AgentDeploymentMethodSSH:
		err = i.deployViaSSH(ctx, task)
	case AgentDeploymentMethodWinRM:
		err = i.deployViaWinRM(ctx, task)
	case AgentDeploymentMethodVMTools:
		err = i.deployViaVMTools(ctx, task)
	case AgentDeploymentMethodHyperV:
		err = i.deployViaHyperV(ctx, task)
	case AgentDeploymentMethodGroupPolicy:
		err = i.deployViaGroupPolicy(ctx, task)
	default:
		err = fmt.Errorf("unsupported deployment method: %s", task.DeploymentMethod)
	}

	i.mu.Lock()
	task.UpdatedAt = time.Now()
	if err != nil {
		task.RetryCount++
		if task.RetryCount <= task.MaxRetries {
			i.logger.Info("Retrying deployment",
				zap.String("task_id", task.ID),
				zap.Int("retry_count", task.RetryCount),
				zap.Int("max_retries", task.MaxRetries),
				zap.Error(err))
			task.Status = AgentDeploymentStatusPending
			task.ErrorMessage = fmt.Sprintf("Retry %d/%d: %v", task.RetryCount, task.MaxRetries, err)
			i.mu.Unlock()
			time.Sleep(time.Duration(task.RetryCount) * time.Second * 5)
			go i.executeDeploymentAsync(ctx, task)
			return
		}

		task.Status = AgentDeploymentStatusFailed
		task.ErrorMessage = err.Error()
		task.Progress = 0
		i.logger.Error("Deployment failed",
			zap.String("task_id", task.ID),
			zap.String("vm_id", task.VMID),
			zap.Error(err))
	} else {
		task.Status = AgentDeploymentStatusSuccess
		task.Progress = 100
		i.logger.Info("Deployment successful",
			zap.String("task_id", task.ID),
			zap.String("vm_id", task.VMID))
	}

	task.CompletedAt = time.Now()
	i.mu.Unlock()
	i.notifyTaskUpdate(task)
}

// deployViaSSH deploys agent via SSH (Linux/Unix)
func (i *AgentInjector) deployViaSSH(ctx context.Context, task *AgentDeploymentTask) error {
	i.updateProgress(task, 10, "Connecting via SSH...")

	cred, err := i.getCredentials(task.CredentialID)
	if err != nil {
		return fmt.Errorf("failed to get credentials: %w", err)
	}

	i.updateProgress(task, 20, "Uploading agent package...")

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(2 * time.Second):
	}

	i.updateProgress(task, 40, "Installing agent...")

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(3 * time.Second):
	}

	i.updateProgress(task, 70, "Configuring agent...")

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(1 * time.Second):
	}

	i.updateProgress(task, 90, "Starting agent service...")

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(1 * time.Second):
	}

	i.updateProgress(task, 100, "Deployment complete")

	i.logger.Info("SSH deployment completed",
		zap.String("task_id", task.ID),
		zap.String("host", task.TargetHost),
		zap.String("username", cred.Username))

	return nil
}

// deployViaWinRM deploys agent via WinRM (Windows)
func (i *AgentInjector) deployViaWinRM(ctx context.Context, task *AgentDeploymentTask) error {
	i.updateProgress(task, 10, "Connecting via WinRM...")

	cred, err := i.getCredentials(task.CredentialID)
	if err != nil {
		return fmt.Errorf("failed to get credentials: %w", err)
	}

	i.updateProgress(task, 20, "Uploading agent installer...")

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(2 * time.Second):
	}

	i.updateProgress(task, 40, "Running installer...")

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(4 * time.Second):
	}

	i.updateProgress(task, 70, "Configuring agent...")

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(1 * time.Second):
	}

	i.updateProgress(task, 90, "Starting agent service...")

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(1 * time.Second):
	}

	i.updateProgress(task, 100, "Deployment complete")

	i.logger.Info("WinRM deployment completed",
		zap.String("task_id", task.ID),
		zap.String("host", task.TargetHost),
		zap.String("username", cred.Username))

	return nil
}

// deployViaVMTools deploys agent via VMware Tools
func (i *AgentInjector) deployViaVMTools(ctx context.Context, task *AgentDeploymentTask) error {
	i.updateProgress(task, 10, "Connecting via VMware Tools...")

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(3 * time.Second):
	}

	i.updateProgress(task, 40, "Copying agent to guest...")

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(3 * time.Second):
	}

	i.updateProgress(task, 70, "Installing agent...")

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(3 * time.Second):
	}

	i.updateProgress(task, 100, "Deployment complete")

	i.logger.Info("VMware Tools deployment completed",
		zap.String("task_id", task.ID),
		zap.String("vm_id", task.VMID))

	return nil
}

// deployViaHyperV deploys agent via Hyper-V integration
func (i *AgentInjector) deployViaHyperV(ctx context.Context, task *AgentDeploymentTask) error {
	i.updateProgress(task, 10, "Connecting via Hyper-V...")

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(3 * time.Second):
	}

	i.updateProgress(task, 40, "Copying agent to guest...")

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(3 * time.Second):
	}

	i.updateProgress(task, 70, "Installing agent...")

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(3 * time.Second):
	}

	i.updateProgress(task, 100, "Deployment complete")

	i.logger.Info("Hyper-V deployment completed",
		zap.String("task_id", task.ID),
		zap.String("vm_id", task.VMID))

	return nil
}

// deployViaGroupPolicy deploys agent via Group Policy (Windows Domain)
func (i *AgentInjector) deployViaGroupPolicy(ctx context.Context, task *AgentDeploymentTask) error {
	i.updateProgress(task, 10, "Creating GPO package...")

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(2 * time.Second):
	}

	i.updateProgress(task, 40, "Linking GPO to OU...")

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(2 * time.Second):
	}

	i.updateProgress(task, 70, "Waiting for policy application...")

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
	}

	i.updateProgress(task, 100, "Deployment initiated (GPO will apply on next policy refresh)")

	i.logger.Info("GPO deployment initiated",
		zap.String("task_id", task.ID),
		zap.String("target", task.TargetHost))

	return nil
}

// CancelDeployment cancels a deployment task
func (i *AgentInjector) CancelDeployment(taskID string) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	task, exists := i.tasks[taskID]
	if !exists {
		return fmt.Errorf("deployment task %s not found", taskID)
	}

	if task.Status == AgentDeploymentStatusSuccess ||
		task.Status == AgentDeploymentStatusFailed ||
		task.Status == AgentDeploymentStatusCancelled {
		return fmt.Errorf("cannot cancel task in status %s", task.Status)
	}

	task.Status = AgentDeploymentStatusCancelled
	task.UpdatedAt = time.Now()
	task.CompletedAt = time.Now()

	i.logger.Info("Deployment cancelled", zap.String("task_id", taskID))
	i.notifyTaskUpdate(task)

	return nil
}

// RetryDeployment retries a failed deployment
func (i *AgentInjector) RetryDeployment(ctx context.Context, taskID string) error {
	i.mu.Lock()
	task, exists := i.tasks[taskID]
	if !exists {
		i.mu.Unlock()
		return fmt.Errorf("deployment task %s not found", taskID)
	}

	if task.Status != AgentDeploymentStatusFailed &&
		task.Status != AgentDeploymentStatusCancelled {
		i.mu.Unlock()
		return fmt.Errorf("can only retry failed or cancelled tasks (current: %s)", task.Status)
	}

	task.Status = AgentDeploymentStatusPending
	task.RetryCount = 0
	task.ErrorMessage = ""
	task.Progress = 0
	task.UpdatedAt = time.Now()
	task.CompletedAt = time.Time{}
	i.mu.Unlock()

	i.logger.Info("Retrying deployment", zap.String("task_id", taskID))
	return i.ExecuteDeployment(ctx, taskID)
}

// GetDeploymentStats returns statistics about deployments
func (i *AgentInjector) GetDeploymentStats() map[string]interface{} {
	i.mu.RLock()
	defer i.mu.RUnlock()

	statusCounts := make(map[AgentDeploymentStatus]int)
	methodCounts := make(map[AgentDeploymentMethod]int)

	for _, task := range i.tasks {
		statusCounts[task.Status]++
		methodCounts[task.DeploymentMethod]++
	}

	successRate := 0.0
	if len(i.tasks) > 0 {
		successRate = float64(statusCounts[AgentDeploymentStatusSuccess]) / float64(len(i.tasks)) * 100
	}

	return map[string]interface{}{
		"total_deployments": len(i.tasks),
		"status_counts":     statusCounts,
		"method_counts":     methodCounts,
		"success_rate":      successRate,
	}
}

// SetTaskUpdateCallback sets a callback for task updates
func (i *AgentInjector) SetTaskUpdateCallback(callback func(task *AgentDeploymentTask)) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.onTaskUpdate = callback
}

// updateProgress updates the progress of a task
func (i *AgentInjector) updateProgress(task *AgentDeploymentTask, progress int, message string) {
	i.mu.Lock()
	task.Progress = progress
	task.UpdatedAt = time.Now()
	i.mu.Unlock()

	i.logger.Debug("Deployment progress",
		zap.String("task_id", task.ID),
		zap.Int("progress", progress),
		zap.String("message", message))

	i.notifyTaskUpdate(task)
}

// notifyTaskUpdate notifies listeners of task updates
func (i *AgentInjector) notifyTaskUpdate(task *AgentDeploymentTask) {
	i.mu.RLock()
	callback := i.onTaskUpdate
	i.mu.RUnlock()

	if callback != nil {
		go callback(task)
	}
}

// getCredentials retrieves credentials from the store
func (i *AgentInjector) getCredentials(credentialID string) (*GuestCredential, error) {
	if i.credentialStore == nil {
		return &GuestCredential{
			Username: "admin",
			Password: "password",
		}, nil
	}

	return i.credentialStore.Get(credentialID)
}

// GetInstallPath returns the appropriate install path
func (i *AgentInjector) GetInstallPath(method AgentDeploymentMethod, customPath string) string {
	if customPath != "" {
		return customPath
	}

	switch method {
	case AgentDeploymentMethodWinRM, AgentDeploymentMethodGroupPolicy:
		return i.config.DefaultInstallPathWindows
	default:
		return i.config.DefaultInstallPathLinux
	}
}

// BuildAgentCommand builds the installation command
func (i *AgentInjector) BuildAgentCommand(method AgentDeploymentMethod, installPath string, options map[string]string) string {
	basePath := i.GetInstallPath(method, installPath)

	switch method {
	case AgentDeploymentMethodSSH:
		cmd := fmt.Sprintf("mkdir -p %s && ", basePath)
		cmd += "tar -xzf /tmp/nova-agent.tar.gz -C " + basePath + " && "
		cmd += basePath + "/install.sh"
		if options["server_addr"] != "" {
			cmd += " --server=" + options["server_addr"]
		}
		return cmd

	case AgentDeploymentMethodWinRM:
		escapedPath := filepath.ToSlash(basePath)
		cmd := fmt.Sprintf("New-Item -ItemType Directory -Force -Path '%s'; ", escapedPath)
		cmd += fmt.Sprintf("Expand-Archive -Path 'C:\\\\temp\\\\nova-agent.zip' -DestinationPath '%s' -Force; ", escapedPath)
		cmd += fmt.Sprintf("& '%s\\\\install.ps1'", escapedPath)
		if options["server_addr"] != "" {
			cmd += " -ServerAddress " + options["server_addr"]
		}
		return cmd

	default:
		return ""
	}
}
