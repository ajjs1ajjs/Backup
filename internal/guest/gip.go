package guest

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// GuestAgentStatus represents the current status of a guest agent
type GuestAgentStatus string

const (
	GuestAgentStatusUnknown   GuestAgentStatus = "Unknown"
	GuestAgentStatusOnline    GuestAgentStatus = "Online"
	GuestAgentStatusOffline   GuestAgentStatus = "Offline"
	GuestAgentStatusUnhealthy GuestAgentStatus = "Unhealthy"
	GuestAgentStatusStarting  GuestAgentStatus = "Starting"
	GuestAgentStatusStopping  GuestAgentStatus = "Stopping"
)

// GuestAgentType represents the type of guest agent
type GuestAgentType string

const (
	GuestAgentTypeWindows GuestAgentType = "windows"
	GuestAgentTypeLinux   GuestAgentType = "linux"
)

// GuestAgent represents a registered guest agent running inside a VM
type GuestAgent struct {
	ID              string         `json:"id"`
	VMID            string         `json:"vm_id"`
	VMName          string         `json:"vm_name"`
	Hostname        string         `json:"hostname"`
	IPAddress       string         `json:"ip_address"`
	AgentType       GuestAgentType `json:"agent_type"`
	AgentVersion    string         `json:"agent_version"`
	Status          GuestAgentStatus `json:"status"`
	LastHeartbeat   time.Time      `json:"last_heartbeat"`
	RegisteredAt    time.Time      `json:"registered_at"`
	Capabilities    []string       `json:"capabilities"`
	Metadata        map[string]string `json:"metadata"`
	HealthScore     int            `json:"health_score"` // 0-100
	ErrorMessage    string         `json:"error_message,omitempty"`
}

// GuestSession represents an active session with a guest agent
type GuestSession struct {
	ID           string            `json:"id"`
	AgentID      string            `json:"agent_id"`
	CreatedAt    time.Time         `json:"created_at"`
	LastActivity time.Time         `json:"last_activity"`
	Status       string            `json:"status"` // "active", "closed", "timeout", "error"
	Username     string            `json:"username,omitempty"`
	Environment  map[string]string `json:"environment,omitempty"`
	WorkingDir   string            `json:"working_dir,omitempty"`
}

// GuestOperation represents an operation to be executed on the guest
type GuestOperation struct {
	ID          string                 `json:"id"`
	SessionID   string                 `json:"session_id"`
	Type        string                 `json:"type"` // "command", "script", "file_upload", "file_download", "app_aware_backup"
	Command     string                 `json:"command,omitempty"`
	Script      string                 `json:"script,omitempty"`
	Arguments   []string               `json:"arguments,omitempty"`
	Timeout     time.Duration          `json:"timeout"`
	Environment map[string]string      `json:"environment,omitempty"`
	Options     map[string]interface{} `json:"options,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   time.Time              `json:"started_at,omitempty"`
	CompletedAt time.Time              `json:"completed_at,omitempty"`
	Status      string                 `json:"status"` // "pending", "running", "completed", "failed", "timeout"
	ExitCode    int                    `json:"exit_code,omitempty"`
	Stdout      string                 `json:"stdout,omitempty"`
	Stderr      string                 `json:"stderr,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// GuestOperationResult represents the result of a guest operation
type GuestOperationResult struct {
	OperationID string        `json:"operation_id"`
	Success     bool          `json:"success"`
	ExitCode    int           `json:"exit_code"`
	Stdout      string        `json:"stdout"`
	Stderr      string        `json:"stderr"`
	Duration    time.Duration `json:"duration"`
	Error       string        `json:"error,omitempty"`
}

// HealthCheckResult represents the result of a health check on a guest agent
type HealthCheckResult struct {
	AgentID      string            `json:"agent_id"`
	Timestamp    time.Time         `json:"timestamp"`
	Healthy      bool              `json:"healthy"`
	HealthScore  int               `json:"health_score"` // 0-100
	Checks       map[string]bool   `json:"checks"`       // check_name -> passed
	Latency      time.Duration     `json:"latency"`
	ErrorMessage string            `json:"error_message,omitempty"`
	Metrics      map[string]string `json:"metrics,omitempty"`
}

// GuestInteractionProxy manages interactions with guest agents running inside VMs
type GuestInteractionProxy struct {
	mu              sync.RWMutex
	agents          map[string]*GuestAgent
	sessions        map[string]*GuestSession
	operations      map[string]*GuestOperation
	logger          *zap.Logger
	healthCheckInterval time.Duration
	heartbeatTimeout    time.Duration
	sessionTimeout      time.Duration
	onAgentStatusChange func(agentID string, oldStatus, newStatus GuestAgentStatus)
	ctx                 context.Context
	cancel              context.CancelFunc
}

// GuestInteractionProxyConfig holds configuration for the proxy
type GuestInteractionProxyConfig struct {
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	HeartbeatTimeout    time.Duration `json:"heartbeat_timeout"`
	SessionTimeout      time.Duration `json:"session_timeout"`
	Logger              *zap.Logger   `json:"-"`
}

// DefaultGuestInteractionProxyConfig returns default configuration
func DefaultGuestInteractionProxyConfig() GuestInteractionProxyConfig {
	return GuestInteractionProxyConfig{
		HealthCheckInterval: 30 * time.Second,
		HeartbeatTimeout:    2 * time.Minute,
		SessionTimeout:      30 * time.Minute,
		Logger:              zap.NewNop(),
	}
}

// NewGuestInteractionProxy creates a new GuestInteractionProxy
func NewGuestInteractionProxy(config GuestInteractionProxyConfig) *GuestInteractionProxy {
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}

	ctx, cancel := context.WithCancel(context.Background())

	proxy := &GuestInteractionProxy{
		agents:              make(map[string]*GuestAgent),
		sessions:            make(map[string]*GuestSession),
		operations:          make(map[string]*GuestOperation),
		logger:              config.Logger,
		healthCheckInterval: config.HealthCheckInterval,
		heartbeatTimeout:    config.HeartbeatTimeout,
		sessionTimeout:      config.SessionTimeout,
		ctx:                 ctx,
		cancel:              cancel,
	}

	// Start background health monitoring
	go proxy.healthMonitor()
	go proxy.sessionCleanup()

	proxy.logger.Info("GuestInteractionProxy initialized",
		zap.Duration("health_check_interval", config.HealthCheckInterval),
		zap.Duration("heartbeat_timeout", config.HeartbeatTimeout),
		zap.Duration("session_timeout", config.SessionTimeout))

	return proxy
}

// RegisterAgent registers a new guest agent
func (p *GuestInteractionProxy) RegisterAgent(agent *GuestAgent) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if agent.ID == "" {
		return fmt.Errorf("agent ID cannot be empty")
	}

	if _, exists := p.agents[agent.ID]; exists {
		p.logger.Warn("Agent already registered, updating", zap.String("agent_id", agent.ID))
	}

	agent.Status = GuestAgentStatusStarting
	agent.RegisteredAt = time.Now()
	agent.LastHeartbeat = time.Now()
	agent.HealthScore = 50 // Start with moderate health score

	p.agents[agent.ID] = agent

	p.logger.Info("Agent registered",
		zap.String("agent_id", agent.ID),
		zap.String("vm_id", agent.VMID),
		zap.String("vm_name", agent.VMName),
		zap.String("ip_address", agent.IPAddress))

	// Trigger status change callback if registered
	if p.onAgentStatusChange != nil {
		go p.onAgentStatusChange(agent.ID, GuestAgentStatusUnknown, GuestAgentStatusStarting)
	}

	return nil
}

// UnregisterAgent removes a guest agent from the proxy
func (p *GuestInteractionProxy) UnregisterAgent(agentID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	agent, exists := p.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	oldStatus := agent.Status
	agent.Status = GuestAgentStatusStopping

	// Close all sessions for this agent
	for sessionID, session := range p.sessions {
		if session.AgentID == agentID {
			session.Status = "closed"
			delete(p.sessions, sessionID)
			p.logger.Debug("Session closed due to agent unregister",
				zap.String("session_id", sessionID),
				zap.String("agent_id", agentID))
		}
	}

	delete(p.agents, agentID)

	p.logger.Info("Agent unregistered", zap.String("agent_id", agentID))

	if p.onAgentStatusChange != nil {
		go p.onAgentStatusChange(agentID, oldStatus, GuestAgentStatusOffline)
	}

	return nil
}

// GetAgent retrieves a guest agent by ID
func (p *GuestInteractionProxy) GetAgent(agentID string) (*GuestAgent, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	agent, exists := p.agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent %s not found", agentID)
	}

	// Return a copy to prevent race conditions
	agentCopy := *agent
	return &agentCopy, nil
}

// ListAgents returns all registered agents, optionally filtered by status
func (p *GuestInteractionProxy) ListAgents(statusFilter ...GuestAgentStatus) []*GuestAgent {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var agents []*GuestAgent

	for _, agent := range p.agents {
		if len(statusFilter) == 0 {
			agentCopy := *agent
			agents = append(agents, &agentCopy)
			continue
		}

		for _, status := range statusFilter {
			if agent.Status == status {
				agentCopy := *agent
				agents = append(agents, &agentCopy)
				break
			}
		}
	}

	return agents
}

// GetAgentByVMID retrieves an agent by VM ID
func (p *GuestInteractionProxy) GetAgentByVMID(vmID string) (*GuestAgent, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, agent := range p.agents {
		if agent.VMID == vmID {
			agentCopy := *agent
			return &agentCopy, nil
		}
	}

	return nil, fmt.Errorf("no agent found for VM %s", vmID)
}

// UpdateAgentHeartbeat updates the last heartbeat time for an agent
func (p *GuestInteractionProxy) UpdateAgentHeartbeat(agentID string, metadata map[string]string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	agent, exists := p.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	agent.LastHeartbeat = time.Now()
	agent.Status = GuestAgentStatusOnline
	agent.HealthScore = 100 // Reset health score on successful heartbeat

	if metadata != nil {
		if agent.Metadata == nil {
			agent.Metadata = make(map[string]string)
		}
		for k, v := range metadata {
			agent.Metadata[k] = v
		}
	}

	return nil
}

// StartSession creates a new session with a guest agent
func (p *GuestInteractionProxy) StartSession(ctx context.Context, agentID string, username string, environment map[string]string) (*GuestSession, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	agent, exists := p.agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent %s not found", agentID)
	}

	if agent.Status != GuestAgentStatusOnline {
		return nil, fmt.Errorf("agent %s is not online (status: %s)", agentID, agent.Status)
	}

	sessionID := fmt.Sprintf("session-%s-%d", agentID, time.Now().UnixNano())
	session := &GuestSession{
		ID:           sessionID,
		AgentID:      agentID,
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
		Status:       "active",
		Username:     username,
		Environment:  environment,
	}

	p.sessions[sessionID] = session

	p.logger.Info("Session started",
		zap.String("session_id", sessionID),
		zap.String("agent_id", agentID),
		zap.String("username", username))

	return session, nil
}

// GetSession retrieves a session by ID
func (p *GuestInteractionProxy) GetSession(sessionID string) (*GuestSession, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	session, exists := p.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	sessionCopy := *session
	return &sessionCopy, nil
}

// CloseSession closes an active session
func (p *GuestInteractionProxy) CloseSession(sessionID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	session, exists := p.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	session.Status = "closed"
	session.LastActivity = time.Now()
	delete(p.sessions, sessionID)

	p.logger.Info("Session closed", zap.String("session_id", sessionID))

	return nil
}

// ListSessions returns all active sessions, optionally filtered by agent ID
func (p *GuestInteractionProxy) ListSessions(agentIDFilter ...string) []*GuestSession {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var sessions []*GuestSession

	for _, session := range p.sessions {
		if len(agentIDFilter) == 0 {
			sessionCopy := *session
			sessions = append(sessions, &sessionCopy)
			continue
		}

		for _, filterID := range agentIDFilter {
			if session.AgentID == filterID {
				sessionCopy := *session
				sessions = append(sessions, &sessionCopy)
				break
			}
		}
	}

	return sessions
}

// ExecuteOperation executes an operation on a guest agent
func (p *GuestInteractionProxy) ExecuteOperation(ctx context.Context, sessionID string, opType string, command string, timeout time.Duration) (*GuestOperation, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	session, exists := p.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	if session.Status != "active" {
		return nil, fmt.Errorf("session %s is not active (status: %s)", sessionID, session.Status)
	}

	// Verify agent is still online
	agent, exists := p.agents[session.AgentID]
	if !exists || agent.Status != GuestAgentStatusOnline {
		return nil, fmt.Errorf("agent %s is not available", session.AgentID)
	}

	opID := fmt.Sprintf("op-%s-%d", sessionID, time.Now().UnixNano())
	operation := &GuestOperation{
		ID:          opID,
		SessionID:   sessionID,
		Type:        opType,
		Command:     command,
		Timeout:     timeout,
		CreatedAt:   time.Now(),
		Status:      "pending",
		Environment: session.Environment,
	}

	p.operations[opID] = operation
	session.LastActivity = time.Now()

	p.logger.Info("Operation queued",
		zap.String("operation_id", opID),
		zap.String("session_id", sessionID),
		zap.String("type", opType),
		zap.Duration("timeout", timeout))

	return operation, nil
}

// ExecuteOperationWithScript executes a script operation on a guest agent
func (p *GuestInteractionProxy) ExecuteOperationWithScript(ctx context.Context, sessionID string, script string, args []string, timeout time.Duration) (*GuestOperation, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	session, exists := p.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	if session.Status != "active" {
		return nil, fmt.Errorf("session %s is not active (status: %s)", sessionID, session.Status)
	}

	agent, exists := p.agents[session.AgentID]
	if !exists || agent.Status != GuestAgentStatusOnline {
		return nil, fmt.Errorf("agent %s is not available", session.AgentID)
	}

	opID := fmt.Sprintf("op-%s-%d", sessionID, time.Now().UnixNano())
	operation := &GuestOperation{
		ID:          opID,
		SessionID:   sessionID,
		Type:        "script",
		Script:      script,
		Arguments:   args,
		Timeout:     timeout,
		CreatedAt:   time.Now(),
		Status:      "pending",
		Environment: session.Environment,
	}

	p.operations[opID] = operation
	session.LastActivity = time.Now()

	p.logger.Info("Script operation queued",
		zap.String("operation_id", opID),
		zap.String("session_id", sessionID),
		zap.Int("script_length", len(script)),
		zap.Duration("timeout", timeout))

	return operation, nil
}

// GetOperation retrieves an operation by ID
func (p *GuestInteractionProxy) GetOperation(operationID string) (*GuestOperation, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	op, exists := p.operations[operationID]
	if !exists {
		return nil, fmt.Errorf("operation %s not found", operationID)
	}

	opCopy := *op
	return &opCopy, nil
}

// UpdateOperationStatus updates the status of an operation
func (p *GuestInteractionProxy) UpdateOperationStatus(operationID string, status string, exitCode int, stdout, stderr, errMsg string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	op, exists := p.operations[operationID]
	if !exists {
		return fmt.Errorf("operation %s not found", operationID)
	}

	oldStatus := op.Status
	op.Status = status
	op.ExitCode = exitCode
	op.Stdout = stdout
	op.Stderr = stderr
	op.Error = errMsg

	if status == "running" && op.StartedAt.IsZero() {
		op.StartedAt = time.Now()
	}

	if status == "completed" || status == "failed" || status == "timeout" {
		op.CompletedAt = time.Now()
	}

	p.logger.Info("Operation status updated",
		zap.String("operation_id", operationID),
		zap.String("old_status", oldStatus),
		zap.String("new_status", status),
		zap.Int("exit_code", exitCode))

	return nil
}

// WaitForOperation waits for an operation to complete with a timeout
func (p *GuestInteractionProxy) WaitForOperation(ctx context.Context, operationID string, checkInterval time.Duration) (*GuestOperation, error) {
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			op, err := p.GetOperation(operationID)
			if err != nil {
				return nil, err
			}

			if op.Status == "completed" || op.Status == "failed" || op.Status == "timeout" {
				return op, nil
			}
		}
	}
}

// CheckHealth performs a health check on a specific agent
func (p *GuestInteractionProxy) CheckHealth(ctx context.Context, agentID string) (*HealthCheckResult, error) {
	p.mu.RLock()
	agent, exists := p.agents[agentID]
	p.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("agent %s not found", agentID)
	}

	startTime := time.Now()
	result := &HealthCheckResult{
		AgentID:     agentID,
		Timestamp:   startTime,
		Healthy:     true,
		HealthScore: 100,
		Checks:      make(map[string]bool),
	}

	// Check 1: Agent is registered and has recent heartbeat
	timeSinceHeartbeat := time.Since(agent.LastHeartbeat)
	result.Checks["heartbeat_recent"] = timeSinceHeartbeat < p.heartbeatTimeout
	if !result.Checks["heartbeat_recent"] {
		result.Healthy = false
		result.HealthScore -= 30
		result.ErrorMessage = "Heartbeat timeout"
	}

	// Check 2: Agent status is online
	result.Checks["status_online"] = agent.Status == GuestAgentStatusOnline
	if !result.Checks["status_online"] {
		result.Healthy = false
		result.HealthScore -= 20
	}

	// Check 3: Agent has valid IP address
	result.Checks["valid_ip"] = agent.IPAddress != ""
	if !result.Checks["valid_ip"] {
		result.HealthScore -= 10
	}

	// Check 4: Agent has reported capabilities
	result.Checks["has_capabilities"] = len(agent.Capabilities) > 0
	if !result.Checks["has_capabilities"] {
		result.HealthScore -= 10
	}

	// Calculate latency (simulated - in real implementation would ping agent)
	result.Latency = time.Since(startTime)

	// Ensure health score doesn't go below 0
	if result.HealthScore < 0 {
		result.HealthScore = 0
	}

	p.logger.Debug("Health check completed",
		zap.String("agent_id", agentID),
		zap.Bool("healthy", result.Healthy),
		zap.Int("health_score", result.HealthScore),
		zap.Duration("latency", result.Latency))

	return result, nil
}

// ListHealthyAgents returns all agents that are currently healthy
func (p *GuestInteractionProxy) ListHealthyAgents() []*GuestAgent {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var healthyAgents []*GuestAgent

	for _, agent := range p.agents {
		if agent.Status == GuestAgentStatusOnline && agent.HealthScore >= 70 {
			agentCopy := *agent
			healthyAgents = append(healthyAgents, &agentCopy)
		}
	}

	return healthyAgents
}

// SetAgentStatusChangeCallback sets a callback for agent status changes
func (p *GuestInteractionProxy) SetAgentStatusChangeCallback(callback func(agentID string, oldStatus, newStatus GuestAgentStatus)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.onAgentStatusChange = callback
}

// healthMonitor runs periodic health checks on all agents
func (p *GuestInteractionProxy) healthMonitor() {
	ticker := time.NewTicker(p.healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			p.logger.Info("Health monitor stopped")
			return
		case <-ticker.C:
			p.performHealthChecks()
		}
	}
}

// performHealthChecks performs health checks on all agents
func (p *GuestInteractionProxy) performHealthChecks() {
	p.mu.Lock()
	agentIDs := make([]string, 0, len(p.agents))
	for agentID := range p.agents {
		agentIDs = append(agentIDs, agentID)
	}
	p.mu.Unlock()

	for _, agentID := range agentIDs {
		result, err := p.CheckHealth(context.Background(), agentID)
		if err != nil {
			p.logger.Error("Health check failed",
				zap.String("agent_id", agentID),
				zap.Error(err))
			continue
		}

		// Update agent health score
		p.mu.Lock()
		if agent, exists := p.agents[agentID]; exists {
			oldStatus := agent.Status
			agent.HealthScore = result.HealthScore

			// Update status based on health
			if result.HealthScore < 30 {
				agent.Status = GuestAgentStatusUnhealthy
			} else if result.HealthScore >= 70 {
				agent.Status = GuestAgentStatusOnline
			}

			if !result.Healthy {
				agent.ErrorMessage = result.ErrorMessage
			} else {
				agent.ErrorMessage = ""
			}

			p.mu.Unlock()

			// Trigger callback if status changed
			if p.onAgentStatusChange != nil && oldStatus != agent.Status {
				go p.onAgentStatusChange(agentID, oldStatus, agent.Status)
			}
		} else {
			p.mu.Unlock()
		}

		if !result.Healthy {
			p.logger.Warn("Agent unhealthy",
				zap.String("agent_id", agentID),
				zap.Int("health_score", result.HealthScore),
				zap.String("error", result.ErrorMessage))
		}
	}
}

// sessionCleanup periodically cleans up timed-out sessions
func (p *GuestInteractionProxy) sessionCleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			p.logger.Info("Session cleanup stopped")
			return
		case <-ticker.C:
			p.cleanupTimedOutSessions()
		}
	}
}

// cleanupTimedOutSessions removes sessions that have exceeded the timeout
func (p *GuestInteractionProxy) cleanupTimedOutSessions() {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	var closedCount int

	for sessionID, session := range p.sessions {
		if now.Sub(session.LastActivity) > p.sessionTimeout {
			session.Status = "timeout"
			delete(p.sessions, sessionID)
			closedCount++

			p.logger.Info("Session timed out",
				zap.String("session_id", sessionID),
				zap.String("agent_id", session.AgentID),
				zap.Duration("inactive_duration", now.Sub(session.LastActivity)))
		}
	}

	if closedCount > 0 {
		p.logger.Debug("Session cleanup completed", zap.Int("closed_count", closedCount))
	}
}

// GetStats returns statistics about the proxy
func (p *GuestInteractionProxy) GetStats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	statusCounts := make(map[GuestAgentStatus]int)
	for _, agent := range p.agents {
		statusCounts[agent.Status]++
	}

	sessionCounts := make(map[string]int)
	for _, session := range p.sessions {
		sessionCounts[session.Status]++
	}

	operationCounts := make(map[string]int)
	for _, op := range p.operations {
		operationCounts[op.Status]++
	}

	return map[string]interface{}{
		"total_agents":      len(p.agents),
		"agent_status":      statusCounts,
		"total_sessions":    len(p.sessions),
		"session_status":    sessionCounts,
		"total_operations":  len(p.operations),
		"operation_status":  operationCounts,
		"healthy_agents":    len(p.ListHealthyAgents()),
	}
}

// Shutdown gracefully shuts down the proxy
func (p *GuestInteractionProxy) Shutdown(ctx context.Context) error {
	p.logger.Info("Shutting down GuestInteractionProxy")

	p.cancel()

	// Wait for background goroutines to finish
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
		// Timeout waiting for graceful shutdown
	}

	// Close all active sessions
	p.mu.Lock()
	for sessionID := range p.sessions {
		p.sessions[sessionID].Status = "closed"
	}
	p.sessions = make(map[string]*GuestSession)
	p.mu.Unlock()

	p.logger.Info("GuestInteractionProxy shutdown complete")
	return nil
}
