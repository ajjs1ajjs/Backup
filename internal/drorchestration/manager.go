package drorchestration

import (
	"context"
	"fmt"
	"sync"
	"time"

	"novabackup/internal/multitenancy"
	"novabackup/internal/storage"
)

// ResourceLimits contains resource constraints
type ResourceLimits struct {
	MaxCPU     int `json:"max_cpu"`
	MaxMemory  int `json:"max_memory"`
	MaxStorage int `json:"max_storage"`
	MaxNetwork int `json:"max_network"`
}

// TriggerType represents the type of trigger
type TriggerType string

const (
	TriggerTypeCron     TriggerType = "cron"
	TriggerTypeInterval TriggerType = "interval"
	TriggerTypeEvent    TriggerType = "event"
	TriggerTypeManual   TriggerType = "manual"
)

// HealthStatus represents health status
type HealthStatus string

const (
	HealthStatusHealthy  HealthStatus = "healthy"
	HealthStatusWarning  HealthStatus = "warning"
	HealthStatusCritical HealthStatus = "critical"
	HealthStatusDown     HealthStatus = "down"
)

// ResourceUsage contains system resource usage
type ResourceUsage struct {
	CPUUsage     float64 `json:"cpu_usage"`
	MemoryUsage  float64 `json:"memory_usage"`
	StorageUsage float64 `json:"storage_usage"`
	NetworkUsage float64 `json:"network_usage"`
}

// HealthIssue represents a system health issue
type HealthIssue struct {
	Severity    string     `json:"severity"`
	Type        string     `json:"type"`
	Description string     `json:"description"`
	NodeID      string     `json:"node_id,omitempty"`
	DetectedAt  time.Time  `json:"detected_at"`
	Resolved    bool       `json:"resolved"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
}

// Constants for execution status
const (
	ExecutionStatusQueued ExecutionStatus = "queued"
)

// DROrchestrator manages disaster recovery orchestration
type DROrchestrator interface {
	// Plan operations
	CreateDRPlan(ctx context.Context, request *DRPlanRequest) (*DRPlan, error)
	GetDRPlan(ctx context.Context, planID string) (*DRPlan, error)
	ListDRPlans(ctx context.Context, filter *DRPlanFilter) ([]*DRPlan, error)
	UpdateDRPlan(ctx context.Context, planID string, request *UpdateDRPlanRequest) (*DRPlan, error)
	DeleteDRPlan(ctx context.Context, planID string) error
	EnableDRPlan(ctx context.Context, planID string) error
	DisableDRPlan(ctx context.Context, planID string) error

	// Orchestration operations
	StartFailover(ctx context.Context, request *FailoverRequest) (*FailoverExecution, error)
	GetFailoverExecution(ctx context.Context, executionID string) (*FailoverExecution, error)
	ListFailoverExecutions(ctx context.Context, filter *ExecutionFilter) ([]*FailoverExecution, error)
	CancelFailoverExecution(ctx context.Context, executionID string) error
	StartFailback(ctx context.Context, request *FailbackRequest) (*FailbackExecution, error)
	GetFailbackExecution(ctx context.Context, executionID string) (*FailbackExecution, error)

	// Testing operations
	RunDRTest(ctx context.Context, request *DRTestRequest) (*DRTestExecution, error)
	GetDRTestExecution(ctx context.Context, testID string) (*DRTestExecution, error)
	ListDRTestExecutions(ctx context.Context, filter *TestFilter) ([]*DRTestExecution, error)
	CancelDRTest(ctx context.Context, testID string) error

	// Statistics and monitoring
	GetDRStats(ctx context.Context, tenantID string, timeRange TimeRange) (*DRStats, error)
	GetPlanStats(ctx context.Context, planID string, timeRange TimeRange) (*PlanStats, error)
	GetGlobalStats(ctx context.Context, timeRange TimeRange) (*GlobalDRStats, error)

	// Health and status
	GetDRSystemHealth(ctx context.Context) (*DRSystemHealth, error)
	GetActiveExecutions(ctx context.Context) ([]*ActiveExecution, error)
}

// DRPlan represents a disaster recovery plan
type DRPlan struct {
	ID             string             `json:"id"`
	Name           string             `json:"name"`
	TenantID       string             `json:"tenant_id"`
	Description    string             `json:"description"`
	Enabled        bool               `json:"enabled"`
	Type           DRPlanType         `json:"type"`
	Priority       DRPlanPriority     `json:"priority"`
	Sites          []DRSite           `json:"sites"`
	Workloads      []DRWorkload       `json:"workloads"`
	RecoverySteps  []RecoveryStep     `json:"recovery_steps"`
	FailoverConfig FailoverConfig     `json:"failover_config"`
	FailbackConfig FailbackConfig     `json:"failback_config"`
	TestConfig     TestConfig         `json:"test_config"`
	Notifications  NotificationConfig `json:"notifications"`
	Metadata       map[string]string  `json:"metadata"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
	LastTestAt     *time.Time         `json:"last_test_at,omitempty"`
	LastFailoverAt *time.Time         `json:"last_failover_at,omitempty"`
	NextTestAt     *time.Time         `json:"next_test_at,omitempty"`
}

// DRPlanRequest contains parameters for creating a DR plan
type DRPlanRequest struct {
	Name           string             `json:"name"`
	TenantID       string             `json:"tenant_id"`
	Description    string             `json:"description"`
	Type           DRPlanType         `json:"type"`
	Priority       DRPlanPriority     `json:"priority"`
	Sites          []DRSite           `json:"sites"`
	Workloads      []DRWorkload       `json:"workloads"`
	RecoverySteps  []RecoveryStep     `json:"recovery_steps"`
	FailoverConfig FailoverConfig     `json:"failover_config"`
	FailbackConfig FailbackConfig     `json:"failback_config"`
	TestConfig     TestConfig         `json:"test_config"`
	Notifications  NotificationConfig `json:"notifications"`
	Enabled        bool               `json:"enabled"`
	Metadata       map[string]string  `json:"metadata"`
}

// UpdateDRPlanRequest contains parameters for updating a DR plan
type UpdateDRPlanRequest struct {
	Name           *string             `json:"name,omitempty"`
	Description    *string             `json:"description,omitempty"`
	Enabled        *bool               `json:"enabled,omitempty"`
	Type           *DRPlanType         `json:"type,omitempty"`
	Priority       *DRPlanPriority     `json:"priority,omitempty"`
	Sites          []DRSite            `json:"sites,omitempty"`
	Workloads      []DRWorkload        `json:"workloads,omitempty"`
	RecoverySteps  []RecoveryStep      `json:"recovery_steps,omitempty"`
	FailoverConfig *FailoverConfig     `json:"failover_config,omitempty"`
	FailbackConfig *FailbackConfig     `json:"failback_config,omitempty"`
	TestConfig     *TestConfig         `json:"test_config,omitempty"`
	Notifications  *NotificationConfig `json:"notifications,omitempty"`
	Metadata       map[string]string   `json:"metadata,omitempty"`
}

// DRPlanFilter contains filters for listing DR plans
type DRPlanFilter struct {
	TenantID      string         `json:"tenant_id,omitempty"`
	Enabled       *bool          `json:"enabled,omitempty"`
	Type          DRPlanType     `json:"type,omitempty"`
	Priority      DRPlanPriority `json:"priority,omitempty"`
	CreatedAfter  *time.Time     `json:"created_after,omitempty"`
	CreatedBefore *time.Time     `json:"created_before,omitempty"`
}

// DRPlanType represents the type of DR plan
type DRPlanType string

const (
	DRPlanTypeSiteLevel     DRPlanType = "site_level"
	DRPlanTypeWorkloadLevel DRPlanType = "workload_level"
	DRPlanTypeApplication   DRPlanType = "application"
	DRPlanTypeHybrid        DRPlanType = "hybrid"
)

// DRPlanPriority represents the priority of DR plan
type DRPlanPriority string

const (
	DRPlanPriorityCritical DRPlanPriority = "critical"
	DRPlanPriorityHigh     DRPlanPriority = "high"
	DRPlanPriorityNormal   DRPlanPriority = "normal"
	DRPlanPriorityLow      DRPlanPriority = "low"
)

// DRSite represents a site in DR plan
type DRSite struct {
	ID         string       `json:"id"`
	Name       string       `json:"name"`
	Type       SiteType     `json:"type"`
	Location   string       `json:"location"`
	Role       SiteRole     `json:"role"`
	Status     SiteStatus   `json:"status"`
	Capacity   SiteCapacity `json:"capacity"`
	Network    SiteNetwork  `json:"network"`
	Storage    SiteStorage  `json:"storage"`
	Compute    SiteCompute  `json:"compute"`
	LastSyncAt *time.Time   `json:"last_sync_at,omitempty"`
	CreatedAt  time.Time    `json:"created_at"`
}

// SiteType represents the type of site
type SiteType string

const (
	SiteTypePrimary   SiteType = "primary"
	SiteTypeSecondary SiteType = "secondary"
	SiteTypeTertiary  SiteType = "tertiary"
	SiteTypeCloud     SiteType = "cloud"
	SiteTypeDRaaS     SiteType = "draas"
)

// SiteRole represents the role of site
type SiteRole string

const (
	SiteRoleActive      SiteRole = "active"
	SiteRolePassive     SiteRole = "passive"
	SiteRoleStandby     SiteRole = "standby"
	SiteRoleHotStandby  SiteRole = "hot_standby"
	SiteRoleColdStandby SiteRole = "cold_standby"
)

// SiteStatus represents the status of site
type SiteStatus string

const (
	SiteStatusOnline      SiteStatus = "online"
	SiteStatusOffline     SiteStatus = "offline"
	SiteStatusDegraded    SiteStatus = "degraded"
	SiteStatusMaintenance SiteStatus = "maintenance"
	SiteStatusFailed      SiteStatus = "failed"
)

// SiteCapacity contains site capacity information
type SiteCapacity struct {
	MaxVMs         int   `json:"max_vms"`
	MaxStorageGB   int64 `json:"max_storage_gb"`
	MaxCPU         int   `json:"max_cpu"`
	MaxMemoryGB    int   `json:"max_memory_gb"`
	MaxNetworkMbps int   `json:"max_network_mbps"`
}

// SiteNetwork contains site network information
type SiteNetwork struct {
	Subnets       []string       `json:"subnets"`
	VLANs         []int          `json:"vlans"`
	VPNEndpoints  []string       `json:"vpn_endpoints"`
	FirewallRules []FirewallRule `json:"firewall_rules"`
	BandwidthMbps int            `json:"bandwidth_mbps"`
	LatencyMs     int            `json:"latency_ms"`
}

// FirewallRule represents a firewall rule
type FirewallRule struct {
	Protocol string `json:"protocol"`
	Port     int    `json:"port"`
	Source   string `json:"source"`
	Dest     string `json:"dest"`
	Action   string `json:"action"`
}

// SiteStorage contains site storage information
type SiteStorage struct {
	Backends    []StorageBackend `json:"backends"`
	Replication bool             `json:"replication"`
	Encryption  bool             `json:"encryption"`
	Compression bool             `json:"compression"`
	Tiering     bool             `json:"tiering"`
}

// StorageBackend represents a storage backend
type StorageBackend struct {
	Type     string            `json:"type"`
	Config   map[string]string `json:"config"`
	Capacity int64             `json:"capacity"`
	Used     int64             `json:"used"`
}

// SiteCompute contains site compute information
type SiteCompute struct {
	Hosts       []ComputeHost `json:"hosts"`
	Clusters    []Cluster     `json:"clusters"`
	VCenters    []vCenter     `json:"vcenters"`
	Hypervisors []HyperV      `json:"hypervisors"`
}

// ComputeHost represents a compute host
type ComputeHost struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	CPU       int    `json:"cpu"`
	MemoryGB  int    `json:"memory_gb"`
	StorageGB int64  `json:"storage_gb"`
	Status    string `json:"status"`
}

// Cluster represents a compute cluster
type Cluster struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Hosts      []string `json:"hosts"`
	Datastores []string `json:"datastores"`
	Networks   []string `json:"networks"`
}

// vCenter represents a vCenter server
type vCenter struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Host     string   `json:"host"`
	Username string   `json:"username"`
	Clusters []string `json:"clusters"`
}

// HyperV represents a Hyper-V server
type HyperV struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Host     string   `json:"host"`
	Clusters []string `json:"clusters"`
}

// DRWorkload represents a workload in DR plan
type DRWorkload struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Type         WorkloadType      `json:"type"`
	Priority     WorkloadPriority  `json:"priority"`
	Dependencies []string          `json:"dependencies"`
	Resources    WorkloadResources `json:"resources"`
	BackupPolicy BackupPolicy      `json:"backup_policy"`
	Recovery     RecoveryPolicy    `json:"recovery_policy"`
	Applications []Application     `json:"applications"`
	VMs          []VM              `json:"vms"`
	Containers   []Container       `json:"containers"`
	Databases    []Database        `json:"databases"`
}

// WorkloadType represents the type of workload
type WorkloadType string

const (
	WorkloadTypeVM          WorkloadType = "vm"
	WorkloadTypeContainer   WorkloadType = "container"
	WorkloadTypeDatabase    WorkloadType = "database"
	WorkloadTypeApplication WorkloadType = "application"
	WorkloadTypeFileServer  WorkloadType = "fileserver"
	WorkloadTypeWebServer   WorkloadType = "webserver"
)

// WorkloadPriority represents the priority of workload
type WorkloadPriority string

const (
	WorkloadPriorityCritical WorkloadPriority = "critical"
	WorkloadPriorityHigh     WorkloadPriority = "high"
	WorkloadPriorityNormal   WorkloadPriority = "normal"
	WorkloadPriorityLow      WorkloadPriority = "low"
)

// WorkloadResources contains workload resource requirements
type WorkloadResources struct {
	CPU         int   `json:"cpu"`
	MemoryGB    int   `json:"memory_gb"`
	StorageGB   int64 `json:"storage_gb"`
	NetworkMbps int   `json:"network_mbps"`
}

// BackupPolicy contains backup policy information
type BackupPolicy struct {
	Schedule    string        `json:"schedule"`
	Retention   time.Duration `json:"retention"`
	Type        string        `json:"type"`
	Compression bool          `json:"compression"`
	Encryption  bool          `json:"encryption"`
	Replication bool          `json:"replication"`
}

// RecoveryPolicy contains recovery policy information
type RecoveryPolicy struct {
	RTO       time.Duration    `json:"rto"`
	RPO       time.Duration    `json:"rpo"`
	Priority  WorkloadPriority `json:"priority"`
	Order     int              `json:"order"`
	Parallel  bool             `json:"parallel"`
	AutoStart bool             `json:"auto_start"`
}

// Application represents an application in workload
type Application struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Type         string            `json:"type"`
	Version      string            `json:"version"`
	Config       map[string]string `json:"config"`
	Dependencies []string          `json:"dependencies"`
	HealthCheck  HealthCheck       `json:"health_check"`
}

// HealthCheck represents application health check
type HealthCheck struct {
	Type     string        `json:"type"`
	Endpoint string        `json:"endpoint"`
	Interval time.Duration `json:"interval"`
	Timeout  time.Duration `json:"timeout"`
	Retries  int           `json:"retries"`
}

// VM represents a virtual machine in workload
type VM struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Platform  string            `json:"platform"`
	Host      string            `json:"host"`
	Datastore string            `json:"datastore"`
	Network   []string          `json:"network"`
	Config    map[string]string `json:"config"`
	Snapshots []Snapshot        `json:"snapshots"`
}

// Snapshot represents a VM snapshot
type Snapshot struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	Size      int64     `json:"size"`
	Parent    string    `json:"parent"`
}

// Container represents a container in workload
type Container struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Image    string            `json:"image"`
	Platform string            `json:"platform"`
	Config   map[string]string `json:"config"`
	Volumes  []Volume          `json:"volumes"`
	Network  []string          `json:"network"`
}

// Volume represents a container volume
type Volume struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	SizeGB int64  `json:"size_gb"`
	Type   string `json:"type"`
}

// Database represents a database in workload
type Database struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Type     string            `json:"type"`
	Version  string            `json:"version"`
	Host     string            `json:"host"`
	Port     int               `json:"port"`
	Config   map[string]string `json:"config"`
	Replicas []Replica         `json:"replicas"`
	Backups  []Backup          `json:"backups"`
}

// Replica represents a database replica
type Replica struct {
	ID     string `json:"id"`
	Host   string `json:"host"`
	Port   int    `json:"port"`
	Role   string `json:"role"`
	Status string `json:"status"`
}

// Backup represents a database backup
type Backup struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	Size      int64     `json:"size"`
	Location  string    `json:"location"`
}

// RecoveryStep represents a step in recovery process
type RecoveryStep struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Type         StepType      `json:"type"`
	Order        int           `json:"order"`
	Parallel     bool          `json:"parallel"`
	Dependencies []string      `json:"dependencies"`
	Config       StepConfig    `json:"config"`
	Timeout      time.Duration `json:"timeout"`
	RetryCount   int           `json:"retry_count"`
	Conditions   []Condition   `json:"conditions"`
	Actions      []Action      `json:"actions"`
	Validations  []Validation  `json:"validations"`
}

// StepType represents the type of recovery step
type StepType string

const (
	StepTypePreCheck     StepType = "pre_check"
	StepTypeSiteFailover StepType = "site_failover"
	StepTypeVMRecovery   StepType = "vm_recovery"
	StepTypeAppStart     StepType = "app_start"
	StepTypeDataSync     StepType = "data_sync"
	StepTypePostCheck    StepType = "post_check"
	StepTypeCleanup      StepType = "cleanup"
)

// StepConfig contains step-specific configuration
type StepConfig struct {
	Parameters  map[string]string `json:"parameters"`
	Environment map[string]string `json:"environment"`
	Resources   ResourceLimits    `json:"resources"`
}

// Condition represents a condition for step execution
type Condition struct {
	Type     string        `json:"type"`
	Target   string        `json:"target"`
	Operator string        `json:"operator"`
	Value    interface{}   `json:"value"`
	Timeout  time.Duration `json:"timeout"`
}

// Action represents an action in recovery step
type Action struct {
	Type       string            `json:"type"`
	Target     string            `json:"target"`
	Parameters map[string]string `json:"parameters"`
	Timeout    time.Duration     `json:"timeout"`
	RetryCount int               `json:"retry_count"`
}

// Validation represents a validation in recovery step
type Validation struct {
	Type       string            `json:"type"`
	Target     string            `json:"target"`
	Parameters map[string]string `json:"parameters"`
	Timeout    time.Duration     `json:"timeout"`
	RetryCount int               `json:"retry_count"`
}

// FailoverConfig contains failover configuration
type FailoverConfig struct {
	Trigger       TriggerConfig `json:"trigger"`
	AutoFailover  bool          `json:"auto_failover"`
	FailoverDelay time.Duration `json:"failover_delay"`
	DataSync      bool          `json:"data_sync"`
	PreCheck      bool          `json:"pre_check"`
	PostCheck     bool          `json:"post_check"`
	Rollback      bool          `json:"rollback"`
	Notification  bool          `json:"notification"`
}

// FailbackConfig contains failback configuration
type FailbackConfig struct {
	Trigger       TriggerConfig `json:"trigger"`
	AutoFailback  bool          `json:"auto_failback"`
	FailbackDelay time.Duration `json:"failback_delay"`
	DataSync      bool          `json:"data_sync"`
	Validation    bool          `json:"validation"`
	PreCheck      bool          `json:"pre_check"`
	PostCheck     bool          `json:"post_check"`
	Rollback      bool          `json:"rollback"`
}

// TestConfig contains test configuration
type TestConfig struct {
	Schedule        TriggerConfig `json:"schedule"`
	TestType        TestType      `json:"test_type"`
	Scope           TestScope     `json:"scope"`
	Isolation       bool          `json:"isolation"`
	DataValidation  bool          `json:"data_validation"`
	PerformanceTest bool          `json:"performance_test"`
	Notification    bool          `json:"notification"`
}

// TestType represents the type of DR test
type TestType string

const (
	TestTypeFull       TestType = "full"
	TestTypePartial    TestType = "partial"
	TestTypeSimulation TestType = "simulation"
	TestTypeReadiness  TestType = "readiness"
)

// TestScope represents the scope of DR test
type TestScope string

const (
	TestScopeAllSites    TestScope = "all_sites"
	TestScopeSingleSite  TestScope = "single_site"
	TestScopeWorkload    TestScope = "workload"
	TestScopeApplication TestScope = "application"
)

// TriggerConfig contains trigger configuration
type TriggerConfig struct {
	Type    TriggerType       `json:"type"`
	Config  map[string]string `json:"config"`
	Enabled bool              `json:"enabled"`
}

// NotificationConfig contains notification configuration
type NotificationConfig struct {
	OnSuccess []string `json:"on_success"`
	OnFailure []string `json:"on_failure"`
	OnError   []string `json:"on_error"`
	OnWarning []string `json:"on_warning"`
}

// FailoverRequest contains parameters for starting failover
type FailoverRequest struct {
	PlanID        string            `json:"plan_id"`
	TenantID      string            `json:"tenant_id"`
	TriggerType   FailoverTrigger   `json:"trigger_type"`
	TriggerReason string            `json:"trigger_reason"`
	Scope         FailoverScope     `json:"scope"`
	Options       FailoverOptions   `json:"options"`
	Metadata      map[string]string `json:"metadata"`
}

// FailoverTrigger represents the trigger type for failover
type FailoverTrigger string

const (
	FailoverTriggerManual    FailoverTrigger = "manual"
	FailoverTriggerAutomatic FailoverTrigger = "automatic"
	FailoverTriggerScheduled FailoverTrigger = "scheduled"
	FailoverTriggerEmergency FailoverTrigger = "emergency"
)

// FailoverScope represents the scope of failover
type FailoverScope string

const (
	FailoverScopeFull      FailoverScope = "full"
	FailoverScopePartial   FailoverScope = "partial"
	FailoverScopeSiteLevel FailoverScope = "site_level"
	FailoverScopeWorkload  FailoverScope = "workload"
)

// FailoverOptions contains failover options
type FailoverOptions struct {
	DataSync   bool          `json:"data_sync"`
	PreCheck   bool          `json:"pre_check"`
	PostCheck  bool          `json:"post_check"`
	Validation bool          `json:"validation"`
	Rollback   bool          `json:"rollback"`
	Timeout    time.Duration `json:"timeout"`
	Parallel   bool          `json:"parallel"`
	DryRun     bool          `json:"dry_run"`
}

// FailoverExecution represents a failover execution
type FailoverExecution struct {
	ID            string             `json:"id"`
	PlanID        string             `json:"plan_id"`
	TenantID      string             `json:"tenant_id"`
	TriggerType   FailoverTrigger    `json:"trigger_type"`
	TriggerReason string             `json:"trigger_reason"`
	Status        ExecutionStatus    `json:"status"`
	Scope         FailoverScope      `json:"scope"`
	Options       FailoverOptions    `json:"options"`
	Steps         []StepExecution    `json:"steps"`
	Results       *FailoverResults   `json:"results,omitempty"`
	Error         string             `json:"error,omitempty"`
	Progress      ExecutionProgress  `json:"progress"`
	Timing        ExecutionTiming    `json:"timing"`
	Resources     ExecutionResources `json:"resources"`
	RollbackInfo  RollbackInfo       `json:"rollback_info"`
	CreatedAt     time.Time          `json:"created_at"`
	StartedAt     *time.Time         `json:"started_at,omitempty"`
	CompletedAt   *time.Time         `json:"completed_at,omitempty"`
	ExpiresAt     *time.Time         `json:"expires_at,omitempty"`
	Metadata      map[string]string  `json:"metadata"`
}

// ExecutionStatus represents the status of execution
type ExecutionStatus string

const (
	ExecutionStatusPending     ExecutionStatus = "pending"
	ExecutionStatusRunning     ExecutionStatus = "running"
	ExecutionStatusCompleted   ExecutionStatus = "completed"
	ExecutionStatusFailed      ExecutionStatus = "failed"
	ExecutionStatusCancelled   ExecutionStatus = "cancelled"
	ExecutionStatusRollingBack ExecutionStatus = "rolling_back"
	ExecutionStatusRolledBack  ExecutionStatus = "rolled_back"
)

// StepExecution represents execution of a recovery step
type StepExecution struct {
	ID         string         `json:"id"`
	StepID     string         `json:"step_id"`
	Status     StepStatus     `json:"status"`
	StartTime  time.Time      `json:"start_time"`
	EndTime    *time.Time     `json:"end_time,omitempty"`
	Duration   time.Duration  `json:"duration"`
	Output     string         `json:"output"`
	Error      string         `json:"error,omitempty"`
	RetryCount int            `json:"retry_count"`
	Results    []ActionResult `json:"results"`
}

// StepStatus represents the status of step execution
type StepStatus string

const (
	StepStatusPending   StepStatus = "pending"
	StepStatusRunning   StepStatus = "running"
	StepStatusCompleted StepStatus = "completed"
	StepStatusFailed    StepStatus = "failed"
	StepStatusSkipped   StepStatus = "skipped"
)

// ActionResult represents result of an action
type ActionResult struct {
	ActionID string        `json:"action_id"`
	Status   string        `json:"status"`
	Output   string        `json:"output"`
	Error    string        `json:"error,omitempty"`
	Duration time.Duration `json:"duration"`
}

// FailoverResults contains failover execution results
type FailoverResults struct {
	Success         bool             `json:"success"`
	TotalSteps      int              `json:"total_steps"`
	CompletedSteps  int              `json:"completed_steps"`
	FailedSteps     int              `json:"failed_steps"`
	SkippedSteps    int              `json:"skipped_steps"`
	ExecutionTime   time.Duration    `json:"execution_time"`
	DataSyncTime    time.Duration    `json:"data_sync_time"`
	ValidationTime  time.Duration    `json:"validation_time"`
	Summary         ExecutionSummary `json:"summary"`
	Issues          []ExecutionIssue `json:"issues"`
	Recommendations []string         `json:"recommendations"`
}

// ExecutionSummary contains execution summary
type ExecutionSummary struct {
	SitesRecovered     int  `json:"sites_recovered"`
	WorkloadsRecovered int  `json:"workloads_recovered"`
	VMsRecovered       int  `json:"vms_recovered"`
	AppsStarted        int  `json:"apps_started"`
	DataValidated      bool `json:"data_validated"`
	PerformanceOK      bool `json:"performance_ok"`
}

// ExecutionIssue represents an issue during execution
type ExecutionIssue struct {
	Severity   string     `json:"severity"`
	Type       string     `json:"type"`
	StepID     string     `json:"step_id"`
	ActionID   string     `json:"action_id"`
	Message    string     `json:"message"`
	DetectedAt time.Time  `json:"detected_at"`
	Resolved   bool       `json:"resolved"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
}

// ExecutionProgress contains progress information
type ExecutionProgress struct {
	Percentage         int           `json:"percentage"`
	CurrentStep        string        `json:"current_step"`
	TotalSteps         int           `json:"total_steps"`
	CompletedSteps     int           `json:"completed_steps"`
	FailedSteps        int           `json:"failed_steps"`
	EstimatedRemaining time.Duration `json:"estimated_remaining"`
}

// ExecutionTiming contains timing information
type ExecutionTiming struct {
	QueuedDuration    time.Duration `json:"queued_duration"`
	ExecutionDuration time.Duration `json:"execution_duration"`
	TotalDuration     time.Duration `json:"total_duration"`
	TimeoutDuration   time.Duration `json:"timeout_duration"`
}

// ExecutionResources contains resource usage information
type ExecutionResources struct {
	CPUUsage     float64  `json:"cpu_usage"`
	MemoryUsage  int64    `json:"memory_usage"`
	StorageUsage int64    `json:"storage_usage"`
	NetworkUsage int64    `json:"network_usage"`
	SitesUsed    []string `json:"sites_used,omitempty"`
}

// RollbackInfo contains rollback information
type RollbackInfo struct {
	Available       bool          `json:"available"`
	Triggered       bool          `json:"triggered"`
	Reason          string        `json:"reason,omitempty"`
	StartTime       *time.Time    `json:"start_time,omitempty"`
	EndTime         *time.Time    `json:"end_time,omitempty"`
	Duration        time.Duration `json:"duration,omitempty"`
	Success         bool          `json:"success"`
	StepsRolledBack int           `json:"steps_rolled_back"`
}

// ExecutionFilter contains filters for listing executions
type ExecutionFilter struct {
	TenantID      string          `json:"tenant_id,omitempty"`
	PlanID        string          `json:"plan_id,omitempty"`
	Status        ExecutionStatus `json:"status,omitempty"`
	TriggerType   FailoverTrigger `json:"trigger_type,omitempty"`
	Scope         FailoverScope   `json:"scope,omitempty"`
	CreatedAfter  *time.Time      `json:"created_after,omitempty"`
	CreatedBefore *time.Time      `json:"created_before,omitempty"`
}

// FailbackRequest contains parameters for starting failback
type FailbackRequest struct {
	PlanID        string            `json:"plan_id"`
	TenantID      string            `json:"tenant_id"`
	ExecutionID   string            `json:"execution_id"`
	TriggerType   FailbackTrigger   `json:"trigger_type"`
	TriggerReason string            `json:"trigger_reason"`
	Scope         FailbackScope     `json:"scope"`
	Options       FailbackOptions   `json:"options"`
	Metadata      map[string]string `json:"metadata"`
}

// FailbackTrigger represents the trigger type for failback
type FailbackTrigger string

const (
	FailbackTriggerManual    FailbackTrigger = "manual"
	FailbackTriggerAutomatic FailbackTrigger = "automatic"
	FailbackTriggerScheduled FailbackTrigger = "scheduled"
)

// FailbackScope represents the scope of failback
type FailbackScope string

const (
	FailbackScopeFull      FailbackScope = "full"
	FailbackScopePartial   FailbackScope = "partial"
	FailbackScopeSiteLevel FailbackScope = "site_level"
	FailbackScopeWorkload  FailbackScope = "workload"
)

// FailbackOptions contains failback options
type FailbackOptions struct {
	DataSync   bool          `json:"data_sync"`
	Validation bool          `json:"validation"`
	PreCheck   bool          `json:"pre_check"`
	PostCheck  bool          `json:"post_check"`
	Timeout    time.Duration `json:"timeout"`
	Parallel   bool          `json:"parallel"`
	DryRun     bool          `json:"dry_run"`
}

// FailbackExecution represents a failback execution
type FailbackExecution struct {
	ID            string             `json:"id"`
	PlanID        string             `json:"plan_id"`
	TenantID      string             `json:"tenant_id"`
	ExecutionID   string             `json:"execution_id"`
	TriggerType   FailbackTrigger    `json:"trigger_type"`
	TriggerReason string             `json:"trigger_reason"`
	Status        ExecutionStatus    `json:"status"`
	Scope         FailbackScope      `json:"scope"`
	Options       FailbackOptions    `json:"options"`
	Steps         []StepExecution    `json:"steps"`
	Results       *FailbackResults   `json:"results,omitempty"`
	Error         string             `json:"error,omitempty"`
	Progress      ExecutionProgress  `json:"progress"`
	Timing        ExecutionTiming    `json:"timing"`
	Resources     ExecutionResources `json:"resources"`
	CreatedAt     time.Time          `json:"created_at"`
	StartedAt     *time.Time         `json:"started_at,omitempty"`
	CompletedAt   *time.Time         `json:"completed_at,omitempty"`
	Metadata      map[string]string  `json:"metadata"`
}

// FailbackResults contains failback execution results
type FailbackResults struct {
	Success         bool             `json:"success"`
	TotalSteps      int              `json:"total_steps"`
	CompletedSteps  int              `json:"completed_steps"`
	FailedSteps     int              `json:"failed_steps"`
	ExecutionTime   time.Duration    `json:"execution_time"`
	DataSyncTime    time.Duration    `json:"data_sync_time"`
	ValidationTime  time.Duration    `json:"validation_time"`
	Summary         ExecutionSummary `json:"summary"`
	Issues          []ExecutionIssue `json:"issues"`
	Recommendations []string         `json:"recommendations"`
}

// DRTestRequest contains parameters for DR testing
type DRTestRequest struct {
	PlanID   string            `json:"plan_id"`
	TenantID string            `json:"tenant_id"`
	TestType TestType          `json:"test_type"`
	Scope    TestScope         `json:"scope"`
	Options  TestOptions       `json:"options"`
	Metadata map[string]string `json:"metadata"`
}

// TestOptions contains test options
type TestOptions struct {
	Isolation       bool          `json:"isolation"`
	DataValidation  bool          `json:"data_validation"`
	PerformanceTest bool          `json:"performance_test"`
	Parallel        bool          `json:"parallel"`
	Timeout         time.Duration `json:"timeout"`
	DryRun          bool          `json:"dry_run"`
}

// DRTestExecution represents a DR test execution
type DRTestExecution struct {
	ID          string             `json:"id"`
	PlanID      string             `json:"plan_id"`
	TenantID    string             `json:"tenant_id"`
	TestType    TestType           `json:"test_type"`
	Scope       TestScope          `json:"scope"`
	Status      ExecutionStatus    `json:"status"`
	Options     TestOptions        `json:"options"`
	Steps       []StepExecution    `json:"steps"`
	Results     *TestResults       `json:"results,omitempty"`
	Error       string             `json:"error,omitempty"`
	Progress    ExecutionProgress  `json:"progress"`
	Timing      ExecutionTiming    `json:"timing"`
	Resources   ExecutionResources `json:"resources"`
	CreatedAt   time.Time          `json:"created_at"`
	StartedAt   *time.Time         `json:"started_at,omitempty"`
	CompletedAt *time.Time         `json:"completed_at,omitempty"`
	Metadata    map[string]string  `json:"metadata"`
}

// TestResults contains test execution results
type TestResults struct {
	Success          bool          `json:"success"`
	TotalSteps       int           `json:"total_steps"`
	CompletedSteps   int           `json:"completed_steps"`
	FailedSteps      int           `json:"failed_steps"`
	ExecutionTime    time.Duration `json:"execution_time"`
	DataValidationOK bool          `json:"data_validation_ok"`
	PerformanceOK    bool          `json:"performance_ok"`
	Summary          TestSummary   `json:"summary"`
	Issues           []TestIssue   `json:"issues"`
	Recommendations  []string      `json:"recommendations"`
}

// TestSummary contains test summary
type TestSummary struct {
	SitesTested     int           `json:"sites_tested"`
	WorkloadsTested int           `json:"workloads_tested"`
	VMsTested       int           `json:"vms_tested"`
	AppsTested      int           `json:"apps_tested"`
	DataValidated   bool          `json:"data_validated"`
	PerformanceOK   bool          `json:"performance_ok"`
	RecoveryTime    time.Duration `json:"recovery_time"`
	Downtime        time.Duration `json:"downtime"`
}

// TestIssue represents an issue during test
type TestIssue struct {
	Severity   string     `json:"severity"`
	Type       string     `json:"type"`
	StepID     string     `json:"step_id"`
	Message    string     `json:"message"`
	DetectedAt time.Time  `json:"detected_at"`
	Resolved   bool       `json:"resolved"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
}

// TestFilter contains filters for listing test executions
type TestFilter struct {
	TenantID      string          `json:"tenant_id,omitempty"`
	PlanID        string          `json:"plan_id,omitempty"`
	Status        ExecutionStatus `json:"status,omitempty"`
	TestType      TestType        `json:"test_type,omitempty"`
	Scope         TestScope       `json:"scope,omitempty"`
	CreatedAfter  *time.Time      `json:"created_after,omitempty"`
	CreatedBefore *time.Time      `json:"created_before,omitempty"`
}

// TimeRange defines a time range for statistics
type TimeRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// DRStats contains DR statistics
type DRStats struct {
	TenantID            string        `json:"tenant_id"`
	TimeRange           TimeRange     `json:"time_range"`
	TotalPlans          int64         `json:"total_plans"`
	ActivePlans         int64         `json:"active_plans"`
	TotalFailovers      int64         `json:"total_failovers"`
	SuccessfulFailovers int64         `json:"successful_failovers"`
	FailedFailovers     int64         `json:"failed_failovers"`
	TotalFailbacks      int64         `json:"total_failbacks"`
	SuccessfulFailbacks int64         `json:"successful_failbacks"`
	FailedFailbacks     int64         `json:"failed_failbacks"`
	TotalTests          int64         `json:"total_tests"`
	SuccessfulTests     int64         `json:"successful_tests"`
	FailedTests         int64         `json:"failed_tests"`
	AverageFailoverTime time.Duration `json:"average_failover_time"`
	AverageFailbackTime time.Duration `json:"average_failback_time"`
	AverageTestTime     time.Duration `json:"average_test_time"`
	LastUpdated         time.Time     `json:"last_updated"`
}

// PlanStats contains plan-specific statistics
type PlanStats struct {
	PlanID              string        `json:"plan_id"`
	TimeRange           TimeRange     `json:"time_range"`
	TotalFailovers      int64         `json:"total_failovers"`
	SuccessfulFailovers int64         `json:"successful_failovers"`
	FailedFailovers     int64         `json:"failed_failovers"`
	TotalFailbacks      int64         `json:"total_failbacks"`
	SuccessfulFailbacks int64         `json:"successful_failbacks"`
	FailedFailbacks     int64         `json:"failed_failbacks"`
	TotalTests          int64         `json:"total_tests"`
	SuccessfulTests     int64         `json:"successful_tests"`
	FailedTests         int64         `json:"failed_tests"`
	AverageFailoverTime time.Duration `json:"average_failover_time"`
	AverageFailbackTime time.Duration `json:"average_failback_time"`
	AverageTestTime     time.Duration `json:"average_test_time"`
	LastFailoverAt      *time.Time    `json:"last_failover_at,omitempty"`
	LastFailbackAt      *time.Time    `json:"last_failback_at,omitempty"`
	LastTestAt          *time.Time    `json:"last_test_at,omitempty"`
	SuccessRate         float64       `json:"success_rate"`
	LastUpdated         time.Time     `json:"last_updated"`
}

// GlobalDRStats contains global DR statistics
type GlobalDRStats struct {
	TimeRange           TimeRange     `json:"time_range"`
	TotalTenants        int64         `json:"total_tenants"`
	TotalPlans          int64         `json:"total_plans"`
	ActivePlans         int64         `json:"active_plans"`
	TotalSites          int64         `json:"total_sites"`
	TotalFailovers      int64         `json:"total_failovers"`
	SuccessfulFailovers int64         `json:"successful_failovers"`
	FailedFailovers     int64         `json:"failed_failovers"`
	TotalFailbacks      int64         `json:"total_failbacks"`
	SuccessfulFailbacks int64         `json:"successful_failbacks"`
	FailedFailbacks     int64         `json:"failed_failbacks"`
	TotalTests          int64         `json:"total_tests"`
	SuccessfulTests     int64         `json:"successful_tests"`
	FailedTests         int64         `json:"failed_tests"`
	AverageFailoverTime time.Duration `json:"average_failover_time"`
	AverageFailbackTime time.Duration `json:"average_failback_time"`
	AverageTestTime     time.Duration `json:"average_test_time"`
	LastUpdated         time.Time     `json:"last_updated"`
}

// DRSystemHealth contains DR system health information
type DRSystemHealth struct {
	Status          HealthStatus    `json:"status"`
	SiteHealth      []SiteHealth    `json:"site_health"`
	PlanHealth      []PlanHealth    `json:"plan_health"`
	ExecutionHealth ExecutionHealth `json:"execution_health"`
	ResourceUsage   ResourceUsage   `json:"resource_usage"`
	ErrorRate       float64         `json:"error_rate"`
	ResponseTime    time.Duration   `json:"response_time"`
	LastHealthCheck time.Time       `json:"last_health_check"`
	Issues          []HealthIssue   `json:"issues"`
}

// SiteHealth contains health information for a site
type SiteHealth struct {
	SiteID       string       `json:"site_id"`
	Name         string       `json:"name"`
	Status       HealthStatus `json:"status"`
	CPUUsage     float64      `json:"cpu_usage"`
	MemoryUsage  float64      `json:"memory_usage"`
	StorageUsage float64      `json:"storage_usage"`
	NetworkUsage float64      `json:"network_usage"`
	LastSync     *time.Time   `json:"last_sync,omitempty"`
	LastSeen     time.Time    `json:"last_seen"`
	Version      string       `json:"version"`
}

// PlanHealth contains health information for a plan
type PlanHealth struct {
	PlanID       string       `json:"plan_id"`
	Name         string       `json:"name"`
	Status       HealthStatus `json:"status"`
	Enabled      bool         `json:"enabled"`
	LastTest     *time.Time   `json:"last_test,omitempty"`
	LastTestPass bool         `json:"last_test_pass"`
	TestCount    int          `json:"test_count"`
	PassCount    int          `json:"pass_count"`
}

// ExecutionHealth contains execution health information
type ExecutionHealth struct {
	Status          HealthStatus  `json:"status"`
	RunningJobs     int           `json:"running_jobs"`
	QueuedJobs      int           `json:"queued_jobs"`
	FailedJobs24h   int           `json:"failed_jobs_24h"`
	SuccessRate     float64       `json:"success_rate"`
	AverageDuration time.Duration `json:"average_duration"`
}

// ActiveExecution represents an active execution
type ActiveExecution struct {
	ID        string            `json:"id"`
	Type      ExecutionType     `json:"type"`
	PlanID    string            `json:"plan_id"`
	TenantID  string            `json:"tenant_id"`
	Status    ExecutionStatus   `json:"status"`
	Progress  ExecutionProgress `json:"progress"`
	StartTime time.Time         `json:"start_time"`
	Duration  time.Duration     `json:"duration"`
	Eta       *time.Time        `json:"eta,omitempty"`
}

// ExecutionType represents the type of execution
type ExecutionType string

const (
	ExecutionTypeFailover ExecutionType = "failover"
	ExecutionTypeFailback ExecutionType = "failback"
	ExecutionTypeTest     ExecutionType = "test"
)

// InMemoryDROrchestrator implements DROrchestrator interface
type InMemoryDROrchestrator struct {
	plans          map[string]*DRPlan
	failovers      map[string]*FailoverExecution
	failbacks      map[string]*FailbackExecution
	tests          map[string]*DRTestExecution
	stats          map[string]*DRStats
	planStats      map[string]*PlanStats
	globalStats    *GlobalDRStats
	tenantManager  multitenancy.TenantManager
	storageManager storage.Engine
	mutex          sync.RWMutex
}

// NewInMemoryDROrchestrator creates a new in-memory DR orchestrator
func NewInMemoryDROrchestrator(tenantMgr multitenancy.TenantManager, storageMgr storage.Engine) *InMemoryDROrchestrator {
	return &InMemoryDROrchestrator{
		plans:          make(map[string]*DRPlan),
		failovers:      make(map[string]*FailoverExecution),
		failbacks:      make(map[string]*FailbackExecution),
		tests:          make(map[string]*DRTestExecution),
		stats:          make(map[string]*DRStats),
		planStats:      make(map[string]*PlanStats),
		globalStats:    &GlobalDRStats{LastUpdated: time.Now()},
		tenantManager:  tenantMgr,
		storageManager: storageMgr,
	}
}

// CreateDRPlan creates a new DR plan
func (m *InMemoryDROrchestrator) CreateDRPlan(ctx context.Context, request *DRPlanRequest) (*DRPlan, error) {
	if err := m.validateTenantAccess(ctx, request.TenantID); err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	plan := &DRPlan{
		ID:             generatePlanID(),
		Name:           request.Name,
		TenantID:       request.TenantID,
		Description:    request.Description,
		Enabled:        request.Enabled,
		Type:           request.Type,
		Priority:       request.Priority,
		Sites:          request.Sites,
		Workloads:      request.Workloads,
		RecoverySteps:  request.RecoverySteps,
		FailoverConfig: request.FailoverConfig,
		FailbackConfig: request.FailbackConfig,
		TestConfig:     request.TestConfig,
		Notifications:  request.Notifications,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Metadata:       request.Metadata,
	}

	m.plans[plan.ID] = plan

	// Initialize plan statistics
	m.planStats[plan.ID] = &PlanStats{
		PlanID:      plan.ID,
		TimeRange:   TimeRange{From: time.Now(), To: time.Now().Add(24 * time.Hour)},
		LastUpdated: time.Now(),
	}

	return plan, nil
}

// GetDRPlan retrieves a DR plan by ID
func (m *InMemoryDROrchestrator) GetDRPlan(ctx context.Context, planID string) (*DRPlan, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	plan, exists := m.plans[planID]
	if !exists {
		return nil, fmt.Errorf("DR plan %s not found", planID)
	}

	if err := m.validateTenantAccess(ctx, plan.TenantID); err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	return plan, nil
}

// ListDRPlans lists DR plans with optional filtering
func (m *InMemoryDROrchestrator) ListDRPlans(ctx context.Context, filter *DRPlanFilter) ([]*DRPlan, error) {
	if filter != nil && filter.TenantID != "" {
		if err := m.validateTenantAccess(ctx, filter.TenantID); err != nil {
			return nil, fmt.Errorf("access denied: %w", err)
		}
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var results []*DRPlan
	for _, plan := range m.plans {
		if filter != nil {
			if filter.TenantID != "" && plan.TenantID != filter.TenantID {
				continue
			}
			if filter.Enabled != nil && plan.Enabled != *filter.Enabled {
				continue
			}
			if filter.Type != "" && plan.Type != filter.Type {
				continue
			}
			if filter.Priority != "" && plan.Priority != filter.Priority {
				continue
			}
			if filter.CreatedAfter != nil && plan.CreatedAt.Before(*filter.CreatedAfter) {
				continue
			}
			if filter.CreatedBefore != nil && plan.CreatedAt.After(*filter.CreatedBefore) {
				continue
			}
		}

		// Check tenant access
		if err := m.validateTenantAccess(ctx, plan.TenantID); err != nil {
			continue
		}

		results = append(results, plan)
	}

	return results, nil
}

// UpdateDRPlan updates an existing DR plan
func (m *InMemoryDROrchestrator) UpdateDRPlan(ctx context.Context, planID string, request *UpdateDRPlanRequest) (*DRPlan, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	plan, exists := m.plans[planID]
	if !exists {
		return nil, fmt.Errorf("DR plan %s not found", planID)
	}

	if err := m.validateTenantAccess(ctx, plan.TenantID); err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	// Update fields
	if request.Name != nil {
		plan.Name = *request.Name
	}
	if request.Description != nil {
		plan.Description = *request.Description
	}
	if request.Enabled != nil {
		plan.Enabled = *request.Enabled
	}
	if request.Type != nil {
		plan.Type = *request.Type
	}
	if request.Priority != nil {
		plan.Priority = *request.Priority
	}
	if request.Sites != nil {
		plan.Sites = request.Sites
	}
	if request.Workloads != nil {
		plan.Workloads = request.Workloads
	}
	if request.RecoverySteps != nil {
		plan.RecoverySteps = request.RecoverySteps
	}
	if request.FailoverConfig != nil {
		plan.FailoverConfig = *request.FailoverConfig
	}
	if request.FailbackConfig != nil {
		plan.FailbackConfig = *request.FailbackConfig
	}
	if request.TestConfig != nil {
		plan.TestConfig = *request.TestConfig
	}
	if request.Notifications != nil {
		plan.Notifications = *request.Notifications
	}
	if request.Metadata != nil {
		plan.Metadata = request.Metadata
	}

	plan.UpdatedAt = time.Now()

	return plan, nil
}

// DeleteDRPlan deletes a DR plan
func (m *InMemoryDROrchestrator) DeleteDRPlan(ctx context.Context, planID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	plan, exists := m.plans[planID]
	if !exists {
		return fmt.Errorf("DR plan %s not found", planID)
	}

	if err := m.validateTenantAccess(ctx, plan.TenantID); err != nil {
		return fmt.Errorf("access denied: %w", err)
	}

	delete(m.plans, planID)
	delete(m.planStats, planID)

	return nil
}

// EnableDRPlan enables a DR plan
func (m *InMemoryDROrchestrator) EnableDRPlan(ctx context.Context, planID string) error {
	_, err := m.UpdateDRPlan(ctx, planID, &UpdateDRPlanRequest{
		Enabled: &[]bool{true}[0],
	})
	return err
}

// DisableDRPlan disables a DR plan
func (m *InMemoryDROrchestrator) DisableDRPlan(ctx context.Context, planID string) error {
	_, err := m.UpdateDRPlan(ctx, planID, &UpdateDRPlanRequest{
		Enabled: &[]bool{false}[0],
	})
	return err
}

// StartFailover starts a failover execution
func (m *InMemoryDROrchestrator) StartFailover(ctx context.Context, request *FailoverRequest) (*FailoverExecution, error) {
	if err := m.validateTenantAccess(ctx, request.TenantID); err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if plan exists and is enabled
	plan, exists := m.plans[request.PlanID]
	if !exists {
		return nil, fmt.Errorf("DR plan %s not found", request.PlanID)
	}

	if !plan.Enabled {
		return nil, fmt.Errorf("DR plan %s is not enabled", request.PlanID)
	}

	execution := &FailoverExecution{
		ID:            generateExecutionID(),
		PlanID:        request.PlanID,
		TenantID:      request.TenantID,
		TriggerType:   request.TriggerType,
		TriggerReason: request.TriggerReason,
		Status:        ExecutionStatusPending,
		Scope:         request.Scope,
		Options:       request.Options,
		Progress: ExecutionProgress{
			Percentage:     0,
			CurrentStep:    "queued",
			TotalSteps:     len(plan.RecoverySteps),
			CompletedSteps: 0,
		},
		Timing: ExecutionTiming{
			TimeoutDuration: request.Options.Timeout,
		},
		RollbackInfo: RollbackInfo{
			Available: true,
		},
		CreatedAt: time.Now(),
		Metadata:  request.Metadata,
	}

	m.failovers[execution.ID] = execution

	// Start execution in background
	go m.runFailoverExecution(ctx, execution)

	return execution, nil
}

// GetFailoverExecution retrieves a failover execution by ID
func (m *InMemoryDROrchestrator) GetFailoverExecution(ctx context.Context, executionID string) (*FailoverExecution, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	execution, exists := m.failovers[executionID]
	if !exists {
		return nil, fmt.Errorf("failover execution %s not found", executionID)
	}

	if err := m.validateTenantAccess(ctx, execution.TenantID); err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	return execution, nil
}

// ListFailoverExecutions lists failover executions with optional filtering
func (m *InMemoryDROrchestrator) ListFailoverExecutions(ctx context.Context, filter *ExecutionFilter) ([]*FailoverExecution, error) {
	if filter != nil && filter.TenantID != "" {
		if err := m.validateTenantAccess(ctx, filter.TenantID); err != nil {
			return nil, fmt.Errorf("access denied: %w", err)
		}
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var results []*FailoverExecution
	for _, execution := range m.failovers {
		if filter != nil {
			if filter.TenantID != "" && execution.TenantID != filter.TenantID {
				continue
			}
			if filter.PlanID != "" && execution.PlanID != filter.PlanID {
				continue
			}
			if filter.Status != "" && execution.Status != filter.Status {
				continue
			}
			if filter.TriggerType != "" && execution.TriggerType != filter.TriggerType {
				continue
			}
			if filter.Scope != "" && execution.Scope != filter.Scope {
				continue
			}
			if filter.CreatedAfter != nil && execution.CreatedAt.Before(*filter.CreatedAfter) {
				continue
			}
			if filter.CreatedBefore != nil && execution.CreatedAt.After(*filter.CreatedBefore) {
				continue
			}
		}

		// Check tenant access
		if err := m.validateTenantAccess(ctx, execution.TenantID); err != nil {
			continue
		}

		results = append(results, execution)
	}

	return results, nil
}

// CancelFailoverExecution cancels a failover execution
func (m *InMemoryDROrchestrator) CancelFailoverExecution(ctx context.Context, executionID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	execution, exists := m.failovers[executionID]
	if !exists {
		return fmt.Errorf("failover execution %s not found", executionID)
	}

	if err := m.validateTenantAccess(ctx, execution.TenantID); err != nil {
		return fmt.Errorf("access denied: %w", err)
	}

	if execution.Status == ExecutionStatusPending || execution.Status == ExecutionStatusRunning {
		execution.Status = ExecutionStatusCancelled
		now := time.Now()
		execution.CompletedAt = &now
	}

	return nil
}

// StartFailback starts a failback execution
func (m *InMemoryDROrchestrator) StartFailback(ctx context.Context, request *FailbackRequest) (*FailbackExecution, error) {
	if err := m.validateTenantAccess(ctx, request.TenantID); err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if plan exists and is enabled
	plan, exists := m.plans[request.PlanID]
	if !exists {
		return nil, fmt.Errorf("DR plan %s not found", request.PlanID)
	}

	if !plan.Enabled {
		return nil, fmt.Errorf("DR plan %s is not enabled", request.PlanID)
	}

	// Check if failover execution exists
	failover, exists := m.failovers[request.ExecutionID]
	if !exists {
		return nil, fmt.Errorf("failover execution %s not found", request.ExecutionID)
	}

	execution := &FailbackExecution{
		ID:            generateExecutionID(),
		PlanID:        request.PlanID,
		TenantID:      request.TenantID,
		ExecutionID:   request.ExecutionID,
		TriggerType:   request.TriggerType,
		TriggerReason: request.TriggerReason,
		Status:        ExecutionStatusPending,
		Scope:         request.Scope,
		Options:       request.Options,
		Progress: ExecutionProgress{
			Percentage:     0,
			CurrentStep:    "queued",
			TotalSteps:     len(plan.RecoverySteps),
			CompletedSteps: 0,
		},
		Timing: ExecutionTiming{
			TimeoutDuration: request.Options.Timeout,
		},
		CreatedAt: time.Now(),
		Metadata:  request.Metadata,
	}

	m.failbacks[execution.ID] = execution

	// Start execution in background
	go m.runFailbackExecution(ctx, execution, failover)

	return execution, nil
}

// GetFailbackExecution retrieves a failback execution by ID
func (m *InMemoryDROrchestrator) GetFailbackExecution(ctx context.Context, executionID string) (*FailbackExecution, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	execution, exists := m.failbacks[executionID]
	if !exists {
		return nil, fmt.Errorf("failback execution %s not found", executionID)
	}

	if err := m.validateTenantAccess(ctx, execution.TenantID); err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	return execution, nil
}

// RunDRTest runs a DR test
func (m *InMemoryDROrchestrator) RunDRTest(ctx context.Context, request *DRTestRequest) (*DRTestExecution, error) {
	if err := m.validateTenantAccess(ctx, request.TenantID); err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if plan exists and is enabled
	plan, exists := m.plans[request.PlanID]
	if !exists {
		return nil, fmt.Errorf("DR plan %s not found", request.PlanID)
	}

	if !plan.Enabled {
		return nil, fmt.Errorf("DR plan %s is not enabled", request.PlanID)
	}

	execution := &DRTestExecution{
		ID:       generateExecutionID(),
		PlanID:   request.PlanID,
		TenantID: request.TenantID,
		TestType: request.TestType,
		Scope:    request.Scope,
		Status:   ExecutionStatusPending,
		Options:  request.Options,
		Progress: ExecutionProgress{
			Percentage:     0,
			CurrentStep:    "queued",
			TotalSteps:     len(plan.RecoverySteps),
			CompletedSteps: 0,
		},
		Timing: ExecutionTiming{
			TimeoutDuration: request.Options.Timeout,
		},
		CreatedAt: time.Now(),
		Metadata:  request.Metadata,
	}

	m.tests[execution.ID] = execution

	// Start test execution in background
	go m.runDRTestExecution(ctx, execution)

	return execution, nil
}

// GetDRTestExecution retrieves a DR test execution by ID
func (m *InMemoryDROrchestrator) GetDRTestExecution(ctx context.Context, testID string) (*DRTestExecution, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	execution, exists := m.tests[testID]
	if !exists {
		return nil, fmt.Errorf("DR test execution %s not found", testID)
	}

	if err := m.validateTenantAccess(ctx, execution.TenantID); err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	return execution, nil
}

// ListDRTestExecutions lists DR test executions with optional filtering
func (m *InMemoryDROrchestrator) ListDRTestExecutions(ctx context.Context, filter *TestFilter) ([]*DRTestExecution, error) {
	if filter != nil && filter.TenantID != "" {
		if err := m.validateTenantAccess(ctx, filter.TenantID); err != nil {
			return nil, fmt.Errorf("access denied: %w", err)
		}
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var results []*DRTestExecution
	for _, execution := range m.tests {
		if filter != nil {
			if filter.TenantID != "" && execution.TenantID != filter.TenantID {
				continue
			}
			if filter.PlanID != "" && execution.PlanID != filter.PlanID {
				continue
			}
			if filter.Status != "" && execution.Status != filter.Status {
				continue
			}
			if filter.TestType != "" && execution.TestType != filter.TestType {
				continue
			}
			if filter.Scope != "" && execution.Scope != filter.Scope {
				continue
			}
			if filter.CreatedAfter != nil && execution.CreatedAt.Before(*filter.CreatedAfter) {
				continue
			}
			if filter.CreatedBefore != nil && execution.CreatedAt.After(*filter.CreatedBefore) {
				continue
			}
		}

		// Check tenant access
		if err := m.validateTenantAccess(ctx, execution.TenantID); err != nil {
			continue
		}

		results = append(results, execution)
	}

	return results, nil
}

// CancelDRTest cancels a DR test
func (m *InMemoryDROrchestrator) CancelDRTest(ctx context.Context, testID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	execution, exists := m.tests[testID]
	if !exists {
		return fmt.Errorf("DR test execution %s not found", testID)
	}

	if err := m.validateTenantAccess(ctx, execution.TenantID); err != nil {
		return fmt.Errorf("access denied: %w", err)
	}

	if execution.Status == ExecutionStatusPending || execution.Status == ExecutionStatusRunning {
		execution.Status = ExecutionStatusCancelled
		now := time.Now()
		execution.CompletedAt = &now
	}

	return nil
}

// GetDRStats retrieves DR statistics for a tenant
func (m *InMemoryDROrchestrator) GetDRStats(ctx context.Context, tenantID string, timeRange TimeRange) (*DRStats, error) {
	if err := m.validateTenantAccess(ctx, tenantID); err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := &DRStats{
		TenantID:    tenantID,
		TimeRange:   timeRange,
		LastUpdated: time.Now(),
	}

	var totalFailoverTime, totalFailbackTime, totalTestTime time.Duration
	var failoverTimes, failbackTimes, testTimes []time.Duration

	for _, plan := range m.plans {
		if plan.TenantID != tenantID {
			continue
		}

		stats.TotalPlans++
		if plan.Enabled {
			stats.ActivePlans++
		}
	}

	for _, execution := range m.failovers {
		if execution.TenantID != tenantID {
			continue
		}

		// Filter by time range
		if execution.CreatedAt.Before(timeRange.From) || execution.CreatedAt.After(timeRange.To) {
			continue
		}

		stats.TotalFailovers++

		switch execution.Status {
		case ExecutionStatusCompleted:
			stats.SuccessfulFailovers++
		case ExecutionStatusFailed:
			stats.FailedFailovers++
		}

		// Calculate execution time
		if execution.Timing.ExecutionDuration > 0 {
			totalFailoverTime += execution.Timing.ExecutionDuration
			failoverTimes = append(failoverTimes, execution.Timing.ExecutionDuration)
		}
	}

	for _, execution := range m.failbacks {
		if execution.TenantID != tenantID {
			continue
		}

		// Filter by time range
		if execution.CreatedAt.Before(timeRange.From) || execution.CreatedAt.After(timeRange.To) {
			continue
		}

		stats.TotalFailbacks++

		switch execution.Status {
		case ExecutionStatusCompleted:
			stats.SuccessfulFailbacks++
		case ExecutionStatusFailed:
			stats.FailedFailbacks++
		}

		// Calculate execution time
		if execution.Timing.ExecutionDuration > 0 {
			totalFailbackTime += execution.Timing.ExecutionDuration
			failbackTimes = append(failbackTimes, execution.Timing.ExecutionDuration)
		}
	}

	for _, execution := range m.tests {
		if execution.TenantID != tenantID {
			continue
		}

		// Filter by time range
		if execution.CreatedAt.Before(timeRange.From) || execution.CreatedAt.After(timeRange.To) {
			continue
		}

		stats.TotalTests++

		switch execution.Status {
		case ExecutionStatusCompleted:
			stats.SuccessfulTests++
		case ExecutionStatusFailed:
			stats.FailedTests++
		}

		// Calculate execution time
		if execution.Timing.ExecutionDuration > 0 {
			totalTestTime += execution.Timing.ExecutionDuration
			testTimes = append(testTimes, execution.Timing.ExecutionDuration)
		}
	}

	// Calculate average times
	if len(failoverTimes) > 0 {
		stats.AverageFailoverTime = totalFailoverTime / time.Duration(len(failoverTimes))
	}
	if len(failbackTimes) > 0 {
		stats.AverageFailbackTime = totalFailbackTime / time.Duration(len(failbackTimes))
	}
	if len(testTimes) > 0 {
		stats.AverageTestTime = totalTestTime / time.Duration(len(testTimes))
	}

	return stats, nil
}

// GetPlanStats retrieves statistics for a specific DR plan
func (m *InMemoryDROrchestrator) GetPlanStats(ctx context.Context, planID string, timeRange TimeRange) (*PlanStats, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	plan, exists := m.plans[planID]
	if !exists {
		return nil, fmt.Errorf("DR plan %s not found", planID)
	}

	if err := m.validateTenantAccess(ctx, plan.TenantID); err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	stats := &PlanStats{
		PlanID:      planID,
		TimeRange:   timeRange,
		LastUpdated: time.Now(),
	}

	var totalFailoverTime, totalFailbackTime, totalTestTime time.Duration
	var failoverTimes, failbackTimes, testTimes []time.Duration

	for _, execution := range m.failovers {
		if execution.PlanID != planID {
			continue
		}

		// Filter by time range
		if execution.CreatedAt.Before(timeRange.From) || execution.CreatedAt.After(timeRange.To) {
			continue
		}

		stats.TotalFailovers++

		switch execution.Status {
		case ExecutionStatusCompleted:
			stats.SuccessfulFailovers++
		case ExecutionStatusFailed:
			stats.FailedFailovers++
		}

		// Calculate execution time
		if execution.Timing.ExecutionDuration > 0 {
			totalFailoverTime += execution.Timing.ExecutionDuration
			failoverTimes = append(failoverTimes, execution.Timing.ExecutionDuration)
		}

		// Track last failover
		if stats.LastFailoverAt == nil || execution.CreatedAt.After(*stats.LastFailoverAt) {
			stats.LastFailoverAt = &execution.CreatedAt
		}
	}

	for _, execution := range m.failbacks {
		if execution.PlanID != planID {
			continue
		}

		// Filter by time range
		if execution.CreatedAt.Before(timeRange.From) || execution.CreatedAt.After(timeRange.To) {
			continue
		}

		stats.TotalFailbacks++

		switch execution.Status {
		case ExecutionStatusCompleted:
			stats.SuccessfulFailbacks++
		case ExecutionStatusFailed:
			stats.FailedFailbacks++
		}

		// Calculate execution time
		if execution.Timing.ExecutionDuration > 0 {
			totalFailbackTime += execution.Timing.ExecutionDuration
			failbackTimes = append(failbackTimes, execution.Timing.ExecutionDuration)
		}

		// Track last failback
		if stats.LastFailbackAt == nil || execution.CreatedAt.After(*stats.LastFailbackAt) {
			stats.LastFailbackAt = &execution.CreatedAt
		}
	}

	for _, execution := range m.tests {
		if execution.PlanID != planID {
			continue
		}

		// Filter by time range
		if execution.CreatedAt.Before(timeRange.From) || execution.CreatedAt.After(timeRange.To) {
			continue
		}

		stats.TotalTests++

		switch execution.Status {
		case ExecutionStatusCompleted:
			stats.SuccessfulTests++
		case ExecutionStatusFailed:
			stats.FailedTests++
		}

		// Calculate execution time
		if execution.Timing.ExecutionDuration > 0 {
			totalTestTime += execution.Timing.ExecutionDuration
			testTimes = append(testTimes, execution.Timing.ExecutionDuration)
		}

		// Track last test
		if stats.LastTestAt == nil || execution.CreatedAt.After(*stats.LastTestAt) {
			stats.LastTestAt = &execution.CreatedAt
		}
	}

	// Calculate success rate
	totalExecutions := stats.TotalFailovers + stats.TotalFailbacks + stats.TotalTests
	if totalExecutions > 0 {
		stats.SuccessRate = float64(stats.SuccessfulFailovers+stats.SuccessfulFailbacks+stats.SuccessfulTests) / float64(totalExecutions)
	}

	// Calculate average times
	if len(failoverTimes) > 0 {
		stats.AverageFailoverTime = totalFailoverTime / time.Duration(len(failoverTimes))
	}
	if len(failbackTimes) > 0 {
		stats.AverageFailbackTime = totalFailbackTime / time.Duration(len(failbackTimes))
	}
	if len(testTimes) > 0 {
		stats.AverageTestTime = totalTestTime / time.Duration(len(testTimes))
	}

	return stats, nil
}

// GetGlobalStats retrieves global DR statistics
func (m *InMemoryDROrchestrator) GetGlobalStats(ctx context.Context, timeRange TimeRange) (*GlobalDRStats, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := &GlobalDRStats{
		TimeRange:   timeRange,
		LastUpdated: time.Now(),
	}

	tenants := make(map[string]bool)

	for _, plan := range m.plans {
		tenants[plan.TenantID] = true
		if plan.Enabled {
			stats.ActivePlans++
		}
	}

	stats.TotalTenants = int64(len(tenants))
	stats.TotalPlans = int64(len(m.plans))

	var totalFailoverTime, totalFailbackTime, totalTestTime time.Duration
	var failoverTimes, failbackTimes, testTimes []time.Duration

	for _, execution := range m.failovers {
		// Filter by time range
		if execution.CreatedAt.Before(timeRange.From) || execution.CreatedAt.After(timeRange.To) {
			continue
		}

		stats.TotalFailovers++
		tenants[execution.TenantID] = true

		switch execution.Status {
		case ExecutionStatusCompleted:
			stats.SuccessfulFailovers++
		case ExecutionStatusFailed:
			stats.FailedFailovers++
		}

		// Calculate execution time
		if execution.Timing.ExecutionDuration > 0 {
			totalFailoverTime += execution.Timing.ExecutionDuration
			failoverTimes = append(failoverTimes, execution.Timing.ExecutionDuration)
		}
	}

	for _, execution := range m.failbacks {
		// Filter by time range
		if execution.CreatedAt.Before(timeRange.From) || execution.CreatedAt.After(timeRange.To) {
			continue
		}

		stats.TotalFailbacks++
		tenants[execution.TenantID] = true

		switch execution.Status {
		case ExecutionStatusCompleted:
			stats.SuccessfulFailbacks++
		case ExecutionStatusFailed:
			stats.FailedFailbacks++
		}

		// Calculate execution time
		if execution.Timing.ExecutionDuration > 0 {
			totalFailbackTime += execution.Timing.ExecutionDuration
			failbackTimes = append(failbackTimes, execution.Timing.ExecutionDuration)
		}
	}

	for _, execution := range m.tests {
		// Filter by time range
		if execution.CreatedAt.Before(timeRange.From) || execution.CreatedAt.After(timeRange.To) {
			continue
		}

		stats.TotalTests++
		tenants[execution.TenantID] = true

		switch execution.Status {
		case ExecutionStatusCompleted:
			stats.SuccessfulTests++
		case ExecutionStatusFailed:
			stats.FailedTests++
		}

		// Calculate execution time
		if execution.Timing.ExecutionDuration > 0 {
			totalTestTime += execution.Timing.ExecutionDuration
			testTimes = append(testTimes, execution.Timing.ExecutionDuration)
		}
	}

	// Calculate average times
	if len(failoverTimes) > 0 {
		stats.AverageFailoverTime = totalFailoverTime / time.Duration(len(failoverTimes))
	}
	if len(failbackTimes) > 0 {
		stats.AverageFailbackTime = totalFailbackTime / time.Duration(len(failbackTimes))
	}
	if len(testTimes) > 0 {
		stats.AverageTestTime = totalTestTime / time.Duration(len(testTimes))
	}

	return stats, nil
}

// GetDRSystemHealth retrieves DR system health information
func (m *InMemoryDROrchestrator) GetDRSystemHealth(ctx context.Context) (*DRSystemHealth, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	health := &DRSystemHealth{
		Status:     HealthStatusHealthy,
		SiteHealth: []SiteHealth{},
		PlanHealth: []PlanHealth{},
		ExecutionHealth: ExecutionHealth{
			Status:          HealthStatusHealthy,
			RunningJobs:     0,
			QueuedJobs:      0,
			FailedJobs24h:   0,
			SuccessRate:     0.95,
			AverageDuration: 5 * time.Minute,
		},
		ResourceUsage: ResourceUsage{
			CPUUsage:     12.5,
			MemoryUsage:  48.2,
			StorageUsage: 31.8,
			NetworkUsage: 8.7,
		},
		ErrorRate:       0.03,
		ResponseTime:    200 * time.Millisecond,
		LastHealthCheck: time.Now(),
		Issues:          []HealthIssue{},
	}

	// Count running and queued jobs
	runningJobs := 0
	pendingJobs := 0
	failedJobs24h := 0
	twentyFourHoursAgo := time.Now().Add(-24 * time.Hour)

	for _, execution := range m.failovers {
		if execution.Status == ExecutionStatusRunning {
			runningJobs++
		}
		if execution.Status == ExecutionStatusPending || execution.Status == ExecutionStatusQueued {
			pendingJobs++
		}
		if execution.Status == ExecutionStatusFailed && execution.CreatedAt.After(twentyFourHoursAgo) {
			failedJobs24h++
		}
	}

	for _, execution := range m.failbacks {
		if execution.Status == ExecutionStatusRunning {
			runningJobs++
		}
		if execution.Status == ExecutionStatusPending || execution.Status == ExecutionStatusQueued {
			pendingJobs++
		}
		if execution.Status == ExecutionStatusFailed && execution.CreatedAt.After(twentyFourHoursAgo) {
			failedJobs24h++
		}
	}

	for _, execution := range m.tests {
		if execution.Status == ExecutionStatusRunning {
			runningJobs++
		}
		if execution.Status == ExecutionStatusPending || execution.Status == ExecutionStatusQueued {
			pendingJobs++
		}
		if execution.Status == ExecutionStatusFailed && execution.CreatedAt.After(twentyFourHoursAgo) {
			failedJobs24h++
		}
	}

	health.ExecutionHealth.RunningJobs = runningJobs
	health.ExecutionHealth.QueuedJobs = pendingJobs
	health.ExecutionHealth.FailedJobs24h = failedJobs24h

	// Add mock site health
	health.SiteHealth = append(health.SiteHealth, SiteHealth{
		SiteID:       "site-1",
		Name:         "Primary Site",
		Status:       HealthStatusHealthy,
		CPUUsage:     15.2,
		MemoryUsage:  42.8,
		StorageUsage: 28.1,
		NetworkUsage: 12.5,
		LastSync:     &[]time.Time{time.Now().Add(-5 * time.Minute)}[0],
		LastSeen:     time.Now().Add(-2 * time.Minute),
		Version:      "1.0.0",
	})

	// Add mock plan health
	for _, plan := range m.plans {
		planHealth := PlanHealth{
			PlanID:       plan.ID,
			Name:         plan.Name,
			Status:       HealthStatusHealthy,
			Enabled:      plan.Enabled,
			LastTestPass: true,
			TestCount:    5,
			PassCount:    4,
		}
		health.PlanHealth = append(health.PlanHealth, planHealth)
	}

	return health, nil
}

// GetActiveExecutions retrieves active executions
func (m *InMemoryDROrchestrator) GetActiveExecutions(ctx context.Context) ([]*ActiveExecution, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var activeExecutions []*ActiveExecution

	// Add active failovers
	for _, execution := range m.failovers {
		if execution.Status == ExecutionStatusRunning || execution.Status == ExecutionStatusPending {
			// Check tenant access
			if err := m.validateTenantAccess(ctx, execution.TenantID); err != nil {
				continue
			}

			activeExec := &ActiveExecution{
				ID:        execution.ID,
				Type:      ExecutionTypeFailover,
				PlanID:    execution.PlanID,
				TenantID:  execution.TenantID,
				Status:    execution.Status,
				Progress:  execution.Progress,
				StartTime: *execution.StartedAt,
			}

			if execution.StartedAt != nil {
				activeExec.Duration = time.Since(*execution.StartedAt)
			}

			activeExecutions = append(activeExecutions, activeExec)
		}
	}

	// Add active failbacks
	for _, execution := range m.failbacks {
		if execution.Status == ExecutionStatusRunning || execution.Status == ExecutionStatusPending {
			// Check tenant access
			if err := m.validateTenantAccess(ctx, execution.TenantID); err != nil {
				continue
			}

			activeExec := &ActiveExecution{
				ID:        execution.ID,
				Type:      ExecutionTypeFailback,
				PlanID:    execution.PlanID,
				TenantID:  execution.TenantID,
				Status:    execution.Status,
				Progress:  execution.Progress,
				StartTime: *execution.StartedAt,
			}

			if execution.StartedAt != nil {
				activeExec.Duration = time.Since(*execution.StartedAt)
			}

			activeExecutions = append(activeExecutions, activeExec)
		}
	}

	// Add active tests
	for _, execution := range m.tests {
		if execution.Status == ExecutionStatusRunning || execution.Status == ExecutionStatusPending {
			// Check tenant access
			if err := m.validateTenantAccess(ctx, execution.TenantID); err != nil {
				continue
			}

			activeExec := &ActiveExecution{
				ID:        execution.ID,
				Type:      ExecutionTypeTest,
				PlanID:    execution.PlanID,
				TenantID:  execution.TenantID,
				Status:    execution.Status,
				Progress:  execution.Progress,
				StartTime: *execution.StartedAt,
			}

			if execution.StartedAt != nil {
				activeExec.Duration = time.Since(*execution.StartedAt)
			}

			activeExecutions = append(activeExecutions, activeExec)
		}
	}

	return activeExecutions, nil
}

// Helper methods

func (m *InMemoryDROrchestrator) runFailoverExecution(ctx context.Context, execution *FailoverExecution) {
	m.mutex.Lock()
	execution.Status = ExecutionStatusRunning
	now := time.Now()
	execution.StartedAt = &now
	execution.Progress.CurrentStep = "starting_failover"
	m.mutex.Unlock()

	// Simulate failover execution
	time.Sleep(3 * time.Second)

	m.mutex.Lock()
	execution.Status = ExecutionStatusCompleted
	completedAt := time.Now()
	execution.CompletedAt = &completedAt
	execution.Progress.Percentage = 100
	execution.Progress.CurrentStep = "completed"
	execution.Progress.CompletedSteps = len(execution.Steps)

	// Set timing
	execution.Timing.ExecutionDuration = completedAt.Sub(*execution.StartedAt)
	execution.Timing.TotalDuration = completedAt.Sub(execution.CreatedAt)

	// Set mock results
	execution.Results = &FailoverResults{
		Success:        true,
		TotalSteps:     len(execution.Steps),
		CompletedSteps: len(execution.Steps),
		FailedSteps:    0,
		ExecutionTime:  execution.Timing.ExecutionDuration,
		Summary: ExecutionSummary{
			SitesRecovered:     2,
			WorkloadsRecovered: 5,
			VMsRecovered:       3,
			AppsStarted:        4,
			DataValidated:      true,
			PerformanceOK:      true,
		},
	}
	m.mutex.Unlock()
}

func (m *InMemoryDROrchestrator) runFailbackExecution(ctx context.Context, execution *FailbackExecution, failover *FailoverExecution) {
	m.mutex.Lock()
	execution.Status = ExecutionStatusRunning
	now := time.Now()
	execution.StartedAt = &now
	execution.Progress.CurrentStep = "starting_failback"
	m.mutex.Unlock()

	// Simulate failback execution
	time.Sleep(2 * time.Second)

	m.mutex.Lock()
	execution.Status = ExecutionStatusCompleted
	completedAt := time.Now()
	execution.CompletedAt = &completedAt
	execution.Progress.Percentage = 100
	execution.Progress.CurrentStep = "completed"
	execution.Progress.CompletedSteps = len(execution.Steps)

	// Set timing
	execution.Timing.ExecutionDuration = completedAt.Sub(*execution.StartedAt)
	execution.Timing.TotalDuration = completedAt.Sub(execution.CreatedAt)

	// Set mock results
	execution.Results = &FailbackResults{
		Success:        true,
		TotalSteps:     len(execution.Steps),
		CompletedSteps: len(execution.Steps),
		FailedSteps:    0,
		ExecutionTime:  execution.Timing.ExecutionDuration,
		Summary: ExecutionSummary{
			SitesRecovered:     2,
			WorkloadsRecovered: 5,
			VMsRecovered:       3,
			AppsStarted:        4,
			DataValidated:      true,
			PerformanceOK:      true,
		},
	}
	m.mutex.Unlock()
}

func (m *InMemoryDROrchestrator) runDRTestExecution(ctx context.Context, execution *DRTestExecution) {
	m.mutex.Lock()
	execution.Status = ExecutionStatusRunning
	now := time.Now()
	execution.StartedAt = &now
	execution.Progress.CurrentStep = "starting_test"
	m.mutex.Unlock()

	// Simulate test execution
	time.Sleep(2 * time.Second)

	m.mutex.Lock()
	execution.Status = ExecutionStatusCompleted
	completedAt := time.Now()
	execution.CompletedAt = &completedAt
	execution.Progress.Percentage = 100
	execution.Progress.CurrentStep = "completed"
	execution.Progress.CompletedSteps = len(execution.Steps)

	// Set timing
	execution.Timing.ExecutionDuration = completedAt.Sub(*execution.StartedAt)
	execution.Timing.TotalDuration = completedAt.Sub(execution.CreatedAt)

	// Set mock results
	execution.Results = &TestResults{
		Success:          true,
		TotalSteps:       len(execution.Steps),
		CompletedSteps:   len(execution.Steps),
		FailedSteps:      0,
		ExecutionTime:    execution.Timing.ExecutionDuration,
		DataValidationOK: true,
		PerformanceOK:    true,
		Summary: TestSummary{
			SitesTested:     2,
			WorkloadsTested: 5,
			VMsTested:       3,
			AppsTested:      4,
			DataValidated:   true,
			PerformanceOK:   true,
			RecoveryTime:    2 * time.Minute,
			Downtime:        30 * time.Second,
		},
	}
	m.mutex.Unlock()
}

func (m *InMemoryDROrchestrator) validateTenantAccess(ctx context.Context, tenantID string) error {
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

func generatePlanID() string {
	return fmt.Sprintf("dr-plan-%d", time.Now().UnixNano())
}

func generateExecutionID() string {
	return fmt.Sprintf("dr-exec-%d", time.Now().UnixNano())
}
