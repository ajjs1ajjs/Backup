package recovery

import (
	"context"
	"fmt"
	"sync"
	"time"

	"novabackup/internal/multitenancy"
	"novabackup/internal/storage"

	"github.com/google/uuid"
)

// RecoveryPlanManager manages automated recovery plans
type RecoveryPlanManager interface {
	// Plan operations
	CreateRecoveryPlan(ctx context.Context, request *RecoveryPlanRequest) (*RecoveryPlan, error)
	GetRecoveryPlan(ctx context.Context, planID string) (*RecoveryPlan, error)
	ListRecoveryPlans(ctx context.Context, filter *RecoveryPlanFilter) ([]*RecoveryPlan, error)
	UpdateRecoveryPlan(ctx context.Context, planID string, request *UpdateRecoveryPlanRequest) (*RecoveryPlan, error)
	DeleteRecoveryPlan(ctx context.Context, planID string) error
	EnableRecoveryPlan(ctx context.Context, planID string) error
	DisableRecoveryPlan(ctx context.Context, planID string) error

	// Recovery execution operations
	StartRecovery(ctx context.Context, request *RecoveryRequest) (*RecoveryExecution, error)
	GetRecoveryExecution(ctx context.Context, executionID string) (*RecoveryExecution, error)
	ListRecoveryExecutions(ctx context.Context, filter *ExecutionFilter) ([]*RecoveryExecution, error)
	CancelRecoveryExecution(ctx context.Context, executionID string) error
	RerunRecoveryExecution(ctx context.Context, executionID string) error

	// VM sequence operations
	CreateVMSequence(ctx context.Context, request *VMSequenceRequest) (*VMSequence, error)
	GetVMSequence(ctx context.Context, sequenceID string) (*VMSequence, error)
	ListVMSequences(ctx context.Context, filter *VMSequenceFilter) ([]*VMSequence, error)
	UpdateVMSequence(ctx context.Context, sequenceID string, request *UpdateVMSequenceRequest) (*VMSequence, error)
	DeleteVMSequence(ctx context.Context, sequenceID string) error

	// Statistics and monitoring
	GetRecoveryStats(ctx context.Context, tenantID string, timeRange TimeRange) (*RecoveryStats, error)
	GetPlanStats(ctx context.Context, planID string, timeRange TimeRange) (*PlanStats, error)
	GetGlobalStats(ctx context.Context, timeRange TimeRange) (*GlobalRecoveryStats, error)

	// Health and status
	GetRecoverySystemHealth(ctx context.Context) (*RecoverySystemHealth, error)
	GetActiveRecoveries(ctx context.Context) ([]*ActiveRecovery, error)
}

// RecoveryPlan represents an automated recovery plan
type RecoveryPlan struct {
	ID            string             `json:"id"`
	Name          string             `json:"name"`
	TenantID      string             `json:"tenant_id"`
	Description   string             `json:"description"`
	Enabled       bool               `json:"enabled"`
	Type          RecoveryPlanType   `json:"type"`
	Priority      RecoveryPriority   `json:"priority"`
	VMSequences   []VMSequence       `json:"vm_sequences"`
	RecoveryOrder RecoveryOrder      `json:"recovery_order"`
	Dependencies  []Dependency       `json:"dependencies"`
	Configuration RecoveryConfig     `json:"configuration"`
	Notifications NotificationConfig `json:"notifications"`
	Retention     RetentionPolicy    `json:"retention"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
	LastRunAt     *time.Time         `json:"last_run_at,omitempty"`
	NextRunAt     *time.Time         `json:"next_run_at,omitempty"`
	Metadata      map[string]string  `json:"metadata"`
}

// RecoveryPlanRequest contains parameters for creating a recovery plan
type RecoveryPlanRequest struct {
	Name          string             `json:"name"`
	TenantID      string             `json:"tenant_id"`
	Description   string             `json:"description"`
	Type          RecoveryPlanType   `json:"type"`
	Priority      RecoveryPriority   `json:"priority"`
	VMSequences   []VMSequence       `json:"vm_sequences"`
	RecoveryOrder RecoveryOrder      `json:"recovery_order"`
	Dependencies  []Dependency       `json:"dependencies"`
	Configuration RecoveryConfig     `json:"configuration"`
	Notifications NotificationConfig `json:"notifications"`
	Retention     RetentionPolicy    `json:"retention"`
	Enabled       bool               `json:"enabled"`
	Metadata      map[string]string  `json:"metadata"`
}

// UpdateRecoveryPlanRequest contains parameters for updating a recovery plan
type UpdateRecoveryPlanRequest struct {
	Name          *string             `json:"name,omitempty"`
	Description   *string             `json:"description,omitempty"`
	Enabled       *bool               `json:"enabled,omitempty"`
	Type          *RecoveryPlanType   `json:"type,omitempty"`
	Priority      *RecoveryPriority   `json:"priority,omitempty"`
	VMSequences   []VMSequence        `json:"vm_sequences,omitempty"`
	RecoveryOrder *RecoveryOrder      `json:"recovery_order,omitempty"`
	Dependencies  []Dependency        `json:"dependencies,omitempty"`
	Configuration *RecoveryConfig     `json:"configuration,omitempty"`
	Notifications *NotificationConfig `json:"notifications,omitempty"`
	Retention     *RetentionPolicy    `json:"retention,omitempty"`
	Metadata      map[string]string   `json:"metadata,omitempty"`
}

// RecoveryPlanFilter contains filters for listing recovery plans
type RecoveryPlanFilter struct {
	TenantID      string           `json:"tenant_id,omitempty"`
	Enabled       *bool            `json:"enabled,omitempty"`
	Type          RecoveryPlanType `json:"type,omitempty"`
	Priority      RecoveryPriority `json:"priority,omitempty"`
	CreatedAfter  *time.Time       `json:"created_after,omitempty"`
	CreatedBefore *time.Time       `json:"created_before,omitempty"`
}

// RecoveryPlanType represents the type of recovery plan
type RecoveryPlanType string

const (
	RecoveryPlanTypeDisaster    RecoveryPlanType = "disaster"
	RecoveryPlanTypeMaintenance RecoveryPlanType = "maintenance"
	RecoveryPlanTypeMigration   RecoveryPlanType = "migration"
	RecoveryPlanTypeTesting     RecoveryPlanType = "testing"
	RecoveryPlanTypeHybrid      RecoveryPlanType = "hybrid"
)

// RecoveryPriority represents the priority of recovery plan
type RecoveryPriority string

const (
	RecoveryPriorityCritical RecoveryPriority = "critical"
	RecoveryPriorityHigh     RecoveryPriority = "high"
	RecoveryPriorityNormal   RecoveryPriority = "normal"
	RecoveryPriorityLow      RecoveryPriority = "low"
)

// VMSequence represents a VM recovery sequence
type VMSequence struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	TenantID      string            `json:"tenant_id"`
	Description   string            `json:"description"`
	Enabled       bool              `json:"enabled"`
	Type          VMSequenceType    `json:"type"`
	VMs           []VMRecovery      `json:"vms"`
	RecoveryOrder VMRecoveryOrder   `json:"recovery_order"`
	Dependencies  []VMDependency    `json:"dependencies"`
	Configuration VMRecoveryConfig  `json:"configuration"`
	NetworkConfig NetworkConfig     `json:"network_config"`
	StorageConfig StorageConfig     `json:"storage_config"`
	Validation    ValidationConfig  `json:"validation"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	LastRunAt     *time.Time        `json:"last_run_at,omitempty"`
	Metadata      map[string]string `json:"metadata"`
}

// VMSequenceRequest contains parameters for creating a VM sequence
type VMSequenceRequest struct {
	Name          string            `json:"name"`
	TenantID      string            `json:"tenant_id"`
	Description   string            `json:"description"`
	Type          VMSequenceType    `json:"type"`
	VMs           []VMRecovery      `json:"vms"`
	RecoveryOrder VMRecoveryOrder   `json:"recovery_order"`
	Dependencies  []VMDependency    `json:"dependencies"`
	Configuration VMRecoveryConfig  `json:"configuration"`
	NetworkConfig NetworkConfig     `json:"network_config"`
	StorageConfig StorageConfig     `json:"storage_config"`
	Validation    ValidationConfig  `json:"validation"`
	Enabled       bool              `json:"enabled"`
	Metadata      map[string]string `json:"metadata"`
}

// UpdateVMSequenceRequest contains parameters for updating a VM sequence
type UpdateVMSequenceRequest struct {
	Name          *string           `json:"name,omitempty"`
	Description   *string           `json:"description,omitempty"`
	Enabled       *bool             `json:"enabled,omitempty"`
	Type          *VMSequenceType   `json:"type,omitempty"`
	VMs           []VMRecovery      `json:"vms,omitempty"`
	RecoveryOrder *VMRecoveryOrder  `json:"recovery_order,omitempty"`
	Dependencies  []VMDependency    `json:"dependencies,omitempty"`
	Configuration *VMRecoveryConfig `json:"configuration,omitempty"`
	NetworkConfig *NetworkConfig    `json:"network_config,omitempty"`
	StorageConfig *StorageConfig    `json:"storage_config,omitempty"`
	Validation    *ValidationConfig `json:"validation,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// VMSequenceFilter contains filters for listing VM sequences
type VMSequenceFilter struct {
	TenantID      string         `json:"tenant_id,omitempty"`
	Enabled       *bool          `json:"enabled,omitempty"`
	Type          VMSequenceType `json:"type,omitempty"`
	CreatedAfter  *time.Time     `json:"created_after,omitempty"`
	CreatedBefore *time.Time     `json:"created_before,omitempty"`
}

// VMSequenceType represents the type of VM sequence
type VMSequenceType string

const (
	VMSequenceTypeSequential  VMSequenceType = "sequential"
	VMSequenceTypeParallel    VMSequenceType = "parallel"
	VMSequenceTypeConditional VMSequenceType = "conditional"
	VMSequenceTypePriority    VMSequenceType = "priority"
)

// VMRecovery represents a VM in recovery sequence
type VMRecovery struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	SourceVM      SourceVM       `json:"source_vm"`
	TargetVM      TargetVM       `json:"target_vm"`
	RecoveryType  VMRecoveryType `json:"recovery_type"`
	Backup        BackupInfo     `json:"backup"`
	Configuration VMConfig       `json:"configuration"`
	Requirements  VMRequirements `json:"requirements"`
	Validation    VMValidation   `json:"validation"`
	PostActions   []PostAction   `json:"post_actions"`
}

// SourceVM represents source VM information
type SourceVM struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Platform   string            `json:"platform"`
	Host       string            `json:"host"`
	Datacenter string            `json:"datacenter"`
	Network    []string          `json:"network"`
	Config     map[string]string `json:"config"`
}

// TargetVM represents target VM information
type TargetVM struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Platform   string            `json:"platform"`
	Host       string            `json:"host"`
	Datacenter string            `json:"datacenter"`
	Network    []string          `json:"network"`
	Config     map[string]string `json:"config"`
	Resources  VMResources       `json:"resources"`
}

// VMRecoveryType represents the type of VM recovery
type VMRecoveryType string

const (
	VMRecoveryTypeFull         VMRecoveryType = "full"
	VMRecoveryTypeIncremental  VMRecoveryType = "incremental"
	VMRecoveryTypeDifferential VMRecoveryType = "differential"
	VMRecoveryTypeBareMetal    VMRecoveryType = "bare_metal"
	VMRecoveryTypeClone        VMRecoveryType = "clone"
)

// BackupInfo represents backup information for VM recovery
type BackupInfo struct {
	BackupID    string    `json:"backup_id"`
	Type        string    `json:"type"`
	Timestamp   time.Time `json:"timestamp"`
	Size        int64     `json:"size"`
	Checksum    string    `json:"checksum"`
	Storage     string    `json:"storage"`
	Encryption  bool      `json:"encryption"`
	Compression bool      `json:"compression"`
}

// VMConfig contains VM configuration
type VMConfig struct {
	CPU          int               `json:"cpu"`
	MemoryGB     int               `json:"memory_gb"`
	StorageGB    int64             `json:"storage_gb"`
	Network      []NetworkConfig   `json:"network"`
	Disks        []DiskConfig      `json:"disks"`
	NICs         []NICConfig       `json:"nics"`
	CustomConfig map[string]string `json:"custom_config"`
}

// VMRequirements contains VM requirements
type VMRequirements struct {
	MinCPU           int    `json:"min_cpu"`
	MinMemoryGB      int    `json:"min_memory_gb"`
	MinStorageGB     int64  `json:"min_storage_gb"`
	NetworkBandwidth int64  `json:"network_bandwidth"`
	StorageType      string `json:"storage_type"`
	StorageTier      string `json:"storage_tier"`
	NetworkType      string `json:"network_type"`
	NetworkTier      string `json:"network_tier"`
}

// VMValidation contains VM validation rules
type VMValidation struct {
	PreCheck         []ValidationRule  `json:"pre_check"`
	PostCheck        []ValidationRule  `json:"post_check"`
	HealthCheck      HealthCheckConfig `json:"health_check"`
	PerformanceCheck PerformanceCheck  `json:"performance_check"`
}

// PostAction represents post-recovery action
type PostAction struct {
	Type       string            `json:"type"`
	Target     string            `json:"target"`
	Config     map[string]string `json:"config"`
	Timeout    time.Duration     `json:"timeout"`
	RetryCount int               `json:"retry_count"`
}

// VMRecoveryOrder defines VM recovery order
type VMRecoveryOrder struct {
	Type        OrderType `json:"type"`
	Groups      []VMGroup `json:"groups"`
	Parallel    bool      `json:"parallel"`
	MaxParallel int       `json:"max_parallel"`
}

// OrderType represents the type of order
type OrderType string

const (
	OrderTypeSequential  OrderType = "sequential"
	OrderTypeParallel    OrderType = "parallel"
	OrderTypePriority    OrderType = "priority"
	OrderTypeConditional OrderType = "conditional"
)

// VMGroup represents a group of VMs
type VMGroup struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	VMIDs        []string `json:"vm_ids"`
	Priority     int      `json:"priority"`
	Dependencies []string `json:"dependencies"`
}

// VMDependency represents VM dependency
type VMDependency struct {
	VMID      string `json:"vm_id"`
	DependsOn string `json:"depends_on"`
	Type      string `json:"type"`
	Condition string `json:"condition"`
}

// VMRecoveryConfig contains VM recovery configuration
type VMRecoveryConfig struct {
	Timeout    time.Duration `json:"timeout"`
	RetryCount int           `json:"retry_count"`
	RetryDelay time.Duration `json:"retry_delay"`
	PreCheck   bool          `json:"pre_check"`
	PostCheck  bool          `json:"post_check"`
	PowerOn    bool          `json:"power_on"`
	Snapshot   bool          `json:"snapshot"`
	Validation bool          `json:"validation"`
	Monitoring bool          `json:"monitoring"`
}

// NetworkConfig contains network configuration
type NetworkConfig struct {
	Subnets      []SubnetConfig     `json:"subnets"`
	VLANs        []VLANConfig       `json:"vlans"`
	Firewall     []FirewallRule     `json:"firewall"`
	LoadBalancer LoadBalancerConfig `json:"load_balancer"`
	DNS          DNSConfig          `json:"dns"`
}

// FirewallRule represents a firewall rule
type FirewallRule struct {
	Protocol string `json:"protocol"`
	Port     int    `json:"port"`
	Source   string `json:"source"`
	Dest     string `json:"dest"`
	Action   string `json:"action"`
}

// SubnetConfig contains subnet configuration
type SubnetConfig struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	CIDR       string   `json:"cidr"`
	Gateway    string   `json:"gateway"`
	DNSServers []string `json:"dns_servers"`
	VLAN       int      `json:"vlan"`
}

// VLANConfig contains VLAN configuration
type VLANConfig struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Subnet      string `json:"subnet"`
	Gateway     string `json:"gateway"`
	Description string `json:"description"`
}

// LoadBalancerConfig contains load balancer configuration
type LoadBalancerConfig struct {
	Type        string            `json:"type"`
	Config      map[string]string `json:"config"`
	Algorithm   string            `json:"algorithm"`
	HealthCheck HealthCheckConfig `json:"health_check"`
}

// DNSConfig contains DNS configuration
type DNSConfig struct {
	Servers    []string          `json:"servers"`
	Records    []DNSRecord       `json:"records"`
	SearchPath []string          `json:"search_path"`
	Options    map[string]string `json:"options"`
}

// DNSRecord represents a DNS record
type DNSRecord struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
	TTL   int    `json:"ttl"`
}

// StorageConfig contains storage configuration
type StorageConfig struct {
	Datastores    []DatastoreConfig `json:"datastores"`
	StoragePolicy StoragePolicy     `json:"storage_policy"`
	ThinProvision bool              `json:"thin_provision"`
	Encryption    bool              `json:"encryption"`
	Compression   bool              `json:"compression"`
}

// DatastoreConfig contains datastore configuration
type DatastoreConfig struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Type      string            `json:"type"`
	Capacity  int64             `json:"capacity"`
	Used      int64             `json:"used"`
	Available int64             `json:"available"`
	Tier      string            `json:"tier"`
	Config    map[string]string `json:"config"`
}

// StoragePolicy contains storage policy
type StoragePolicy struct {
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Tier        string            `json:"tier"`
	Replication int               `json:"replication"`
	Config      map[string]string `json:"config"`
}

// DiskConfig contains disk configuration
type DiskConfig struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Type          string            `json:"type"`
	SizeGB        int64             `json:"size_gb"`
	Format        string            `json:"format"`
	ThinProvision bool              `json:"thin_provision"`
	StorageTier   string            `json:"storage_tier"`
	Config        map[string]string `json:"config"`
}

// NICConfig contains NIC configuration
type NICConfig struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	Network    string            `json:"network"`
	MACAddress string            `json:"mac_address"`
	IPAddress  string            `json:"ip_address"`
	Config     map[string]string `json:"config"`
}

// VMResources contains VM resource information
type VMResources struct {
	CPU         int   `json:"cpu"`
	MemoryGB    int   `json:"memory_gb"`
	StorageGB   int64 `json:"storage_gb"`
	NetworkMbps int   `json:"network_mbps"`
}

// ValidationConfig contains validation configuration
type ValidationConfig struct {
	PreCheck         []ValidationRule  `json:"pre_check"`
	PostCheck        []ValidationRule  `json:"post_check"`
	HealthCheck      HealthCheckConfig `json:"health_check"`
	PerformanceCheck PerformanceCheck  `json:"performance_check"`
}

// ValidationRule represents a validation rule
type ValidationRule struct {
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	Target     string            `json:"target"`
	Parameters map[string]string `json:"parameters"`
	Timeout    time.Duration     `json:"timeout"`
	RetryCount int               `json:"retry_count"`
	Required   bool              `json:"required"`
}

// HealthCheckConfig contains health check configuration
type HealthCheckConfig struct {
	Type       string            `json:"type"`
	Endpoint   string            `json:"endpoint"`
	Interval   time.Duration     `json:"interval"`
	Timeout    time.Duration     `json:"timeout"`
	Retries    int               `json:"retries"`
	Parameters map[string]string `json:"parameters"`
}

// PerformanceCheck contains performance check configuration
type PerformanceCheck struct {
	Metrics    []string           `json:"metrics"`
	Thresholds map[string]float64 `json:"thresholds"`
	Duration   time.Duration      `json:"duration"`
	Interval   time.Duration      `json:"interval"`
}

// RecoveryOrder defines overall recovery order
type RecoveryOrder struct {
	Type        OrderType       `json:"type"`
	Groups      []RecoveryGroup `json:"groups"`
	Parallel    bool            `json:"parallel"`
	MaxParallel int             `json:"max_parallel"`
}

// RecoveryGroup represents a recovery group
type RecoveryGroup struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	SequenceIDs  []string `json:"sequence_ids"`
	Priority     int      `json:"priority"`
	Dependencies []string `json:"dependencies"`
}

// Dependency represents a dependency
type Dependency struct {
	SequenceID string `json:"sequence_id"`
	DependsOn  string `json:"depends_on"`
	Type       string `json:"type"`
	Condition  string `json:"condition"`
}

// RecoveryConfig contains recovery configuration
type RecoveryConfig struct {
	Timeout    time.Duration `json:"timeout"`
	RetryCount int           `json:"retry_count"`
	RetryDelay time.Duration `json:"retry_delay"`
	PreCheck   bool          `json:"pre_check"`
	PostCheck  bool          `json:"post_check"`
	Validation bool          `json:"validation"`
	Monitoring bool          `json:"monitoring"`
	AutoStart  bool          `json:"auto_start"`
	Rollback   bool          `json:"rollback"`
}

// NotificationConfig contains notification configuration
type NotificationConfig struct {
	OnSuccess []string `json:"on_success"`
	OnFailure []string `json:"on_failure"`
	OnError   []string `json:"on_error"`
	OnWarning []string `json:"on_warning"`
}

// RetentionPolicy contains retention policy
type RetentionPolicy struct {
	ExecutionDays int `json:"execution_days"`
	ReportDays    int `json:"report_days"`
	LogDays       int `json:"log_days"`
	ArtifactsDays int `json:"artifacts_days"`
	MaxExecutions int `json:"max_executions"`
	MaxReports    int `json:"max_reports"`
	MaxLogs       int `json:"max_logs"`
	MaxArtifacts  int `json:"max_artifacts"`
}

// RecoveryRequest contains parameters for starting recovery
type RecoveryRequest struct {
	PlanID        string            `json:"plan_id"`
	TenantID      string            `json:"tenant_id"`
	TriggerType   RecoveryTrigger   `json:"trigger_type"`
	TriggerReason string            `json:"trigger_reason"`
	Scope         RecoveryScope     `json:"scope"`
	Options       RecoveryOptions   `json:"options"`
	Sequences     []string          `json:"sequences,omitempty"`
	Metadata      map[string]string `json:"metadata"`
}

// RecoveryTrigger represents the trigger type for recovery
type RecoveryTrigger string

const (
	RecoveryTriggerManual    RecoveryTrigger = "manual"
	RecoveryTriggerAutomatic RecoveryTrigger = "automatic"
	RecoveryTriggerScheduled RecoveryTrigger = "scheduled"
	RecoveryTriggerEmergency RecoveryTrigger = "emergency"
)

// RecoveryScope represents the scope of recovery
type RecoveryScope string

const (
	RecoveryScopeFull     RecoveryScope = "full"
	RecoveryScopePartial  RecoveryScope = "partial"
	RecoveryScopeSequence RecoveryScope = "sequence"
	RecoveryScopeVM       RecoveryScope = "vm"
)

// RecoveryOptions contains recovery options
type RecoveryOptions struct {
	PreCheck    bool          `json:"pre_check"`
	PostCheck   bool          `json:"post_check"`
	Validation  bool          `json:"validation"`
	Monitoring  bool          `json:"monitoring"`
	Timeout     time.Duration `json:"timeout"`
	Parallel    bool          `json:"parallel"`
	MaxParallel int           `json:"max_parallel"`
	DryRun      bool          `json:"dry_run"`
	Rollback    bool          `json:"rollback"`
}

// RecoveryExecution represents a recovery execution
type RecoveryExecution struct {
	ID            string              `json:"id"`
	PlanID        string              `json:"plan_id"`
	TenantID      string              `json:"tenant_id"`
	TriggerType   RecoveryTrigger     `json:"trigger_type"`
	TriggerReason string              `json:"trigger_reason"`
	Status        ExecutionStatus     `json:"status"`
	Scope         RecoveryScope       `json:"scope"`
	Options       RecoveryOptions     `json:"options"`
	Sequences     []SequenceExecution `json:"sequences"`
	Results       *RecoveryResults    `json:"results,omitempty"`
	Error         string              `json:"error,omitempty"`
	Progress      ExecutionProgress   `json:"progress"`
	Timing        ExecutionTiming     `json:"timing"`
	Resources     ExecutionResources  `json:"resources"`
	RollbackInfo  RollbackInfo        `json:"rollback_info"`
	CreatedAt     time.Time           `json:"created_at"`
	StartedAt     *time.Time          `json:"started_at,omitempty"`
	CompletedAt   *time.Time          `json:"completed_at,omitempty"`
	ExpiresAt     *time.Time          `json:"expires_at,omitempty"`
	Metadata      map[string]string   `json:"metadata"`
}

// SequenceExecution represents execution of a VM sequence
type SequenceExecution struct {
	ID           string         `json:"id"`
	SequenceID   string         `json:"sequence_id"`
	Status       SequenceStatus `json:"status"`
	StartTime    time.Time      `json:"start_time"`
	EndTime      *time.Time     `json:"end_time,omitempty"`
	Duration     time.Duration  `json:"duration"`
	VMExecutions []VMExecution  `json:"vm_executions"`
	Output       string         `json:"output"`
	Error        string         `json:"error,omitempty"`
	RetryCount   int            `json:"retry_count"`
}

// SequenceStatus represents the status of sequence execution
type SequenceStatus string

const (
	SequenceStatusPending   SequenceStatus = "pending"
	SequenceStatusRunning   SequenceStatus = "running"
	SequenceStatusCompleted SequenceStatus = "completed"
	SequenceStatusFailed    SequenceStatus = "failed"
	SequenceStatusSkipped   SequenceStatus = "skipped"
)

// VMExecution represents execution of a VM recovery
type VMExecution struct {
	ID         string          `json:"id"`
	VMID       string          `json:"vm_id"`
	Status     VMStatus        `json:"status"`
	StartTime  time.Time       `json:"start_time"`
	EndTime    *time.Time      `json:"end_time,omitempty"`
	Duration   time.Duration   `json:"duration"`
	Output     string          `json:"output"`
	Error      string          `json:"error,omitempty"`
	RetryCount int             `json:"retry_count"`
	Steps      []StepExecution `json:"steps"`
}

// VMStatus represents the status of VM execution
type VMStatus string

const (
	VMStatusPending   VMStatus = "pending"
	VMStatusRunning   VMStatus = "running"
	VMStatusCompleted VMStatus = "completed"
	VMStatusFailed    VMStatus = "failed"
	VMStatusSkipped   VMStatus = "skipped"
)

// StepExecution represents execution of a recovery step
type StepExecution struct {
	StepID     string        `json:"step_id"`
	Status     StepStatus    `json:"status"`
	StartTime  time.Time     `json:"start_time"`
	EndTime    *time.Time    `json:"end_time,omitempty"`
	Duration   time.Duration `json:"duration"`
	Output     string        `json:"output"`
	Error      string        `json:"error,omitempty"`
	RetryCount int           `json:"retry_count"`
}

// RecoveryResults contains recovery execution results
type RecoveryResults struct {
	Success            bool             `json:"success"`
	TotalSequences     int              `json:"total_sequences"`
	CompletedSequences int              `json:"completed_sequences"`
	FailedSequences    int              `json:"failed_sequences"`
	SkippedSequences   int              `json:"skipped_sequences"`
	TotalVMs           int              `json:"total_vms"`
	CompletedVMs       int              `json:"completed_vms"`
	FailedVMs          int              `json:"failed_vms"`
	SkippedVMs         int              `json:"skipped_vms"`
	ExecutionTime      time.Duration    `json:"execution_time"`
	Summary            ExecutionSummary `json:"summary"`
	Issues             []RecoveryIssue  `json:"issues"`
	Recommendations    []string         `json:"recommendations"`
}

// ExecutionSummary contains execution summary
type ExecutionSummary struct {
	SequencesRecovered int  `json:"sequences_recovered"`
	VMsRecovered       int  `json:"vms_recovered"`
	DataValidated      bool `json:"data_validated"`
	PerformanceOK      bool `json:"performance_ok"`
	NetworkOK          bool `json:"network_ok"`
	StorageOK          bool `json:"storage_ok"`
}

// RecoveryIssue represents an issue during recovery
type RecoveryIssue struct {
	Severity   string     `json:"severity"`
	Type       string     `json:"type"`
	SequenceID string     `json:"sequence_id,omitempty"`
	VMID       string     `json:"vm_id,omitempty"`
	StepID     string     `json:"step_id,omitempty"`
	Message    string     `json:"message"`
	DetectedAt time.Time  `json:"detected_at"`
	Resolved   bool       `json:"resolved"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
}

// ExecutionProgress contains progress information
type ExecutionProgress struct {
	Percentage         int           `json:"percentage"`
	CurrentSequence    string        `json:"current_sequence"`
	CurrentVM          string        `json:"current_vm"`
	TotalSequences     int           `json:"total_sequences"`
	CompletedSequences int           `json:"completed_sequences"`
	TotalVMs           int           `json:"total_vms"`
	CompletedVMs       int           `json:"completed_vms"`
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
	Hosts        []string `json:"hosts,omitempty"`
	Datastores   []string `json:"datastores,omitempty"`
}

// RollbackInfo contains rollback information
type RollbackInfo struct {
	Available           bool          `json:"available"`
	Triggered           bool          `json:"triggered"`
	Reason              string        `json:"reason,omitempty"`
	StartTime           *time.Time    `json:"start_time,omitempty"`
	EndTime             *time.Time    `json:"end_time,omitempty"`
	Duration            time.Duration `json:"duration,omitempty"`
	Success             bool          `json:"success"`
	SequencesRolledBack int           `json:"sequences_rolled_back"`
	VMsRolledBack       int           `json:"vms_rolled_back"`
}

// ExecutionFilter contains filters for listing executions
type ExecutionFilter struct {
	TenantID      string          `json:"tenant_id,omitempty"`
	PlanID        string          `json:"plan_id,omitempty"`
	Status        ExecutionStatus `json:"status,omitempty"`
	TriggerType   RecoveryTrigger `json:"trigger_type,omitempty"`
	Scope         RecoveryScope   `json:"scope,omitempty"`
	CreatedAfter  *time.Time      `json:"created_after,omitempty"`
	CreatedBefore *time.Time      `json:"created_before,omitempty"`
}

// TimeRange defines a time range for statistics
type TimeRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// RecoveryStats contains recovery statistics
type RecoveryStats struct {
	TenantID             string        `json:"tenant_id"`
	TimeRange            TimeRange     `json:"time_range"`
	TotalPlans           int64         `json:"total_plans"`
	ActivePlans          int64         `json:"active_plans"`
	TotalSequences       int64         `json:"total_sequences"`
	TotalExecutions      int64         `json:"total_executions"`
	CompletedExecutions  int64         `json:"completed_executions"`
	FailedExecutions     int64         `json:"failed_executions"`
	CancelledExecutions  int64         `json:"cancelled_executions"`
	RunningExecutions    int64         `json:"running_executions"`
	QueuedExecutions     int64         `json:"queued_executions"`
	TotalVMs             int64         `json:"total_vms"`
	RecoveredVMs         int64         `json:"recovered_vms"`
	FailedVMs            int64         `json:"failed_vms"`
	SuccessRate          float64       `json:"success_rate"`
	AverageExecutionTime time.Duration `json:"average_execution_time"`
	MinExecutionTime     time.Duration `json:"min_execution_time"`
	MaxExecutionTime     time.Duration `json:"max_execution_time"`
	TotalExecutionTime   time.Duration `json:"total_execution_time"`
	LastUpdated          time.Time     `json:"last_updated"`
}

// PlanStats contains plan-specific statistics
type PlanStats struct {
	PlanID               string        `json:"plan_id"`
	TimeRange            TimeRange     `json:"time_range"`
	TotalExecutions      int64         `json:"total_executions"`
	CompletedExecutions  int64         `json:"completed_executions"`
	FailedExecutions     int64         `json:"failed_executions"`
	CancelledExecutions  int64         `json:"cancelled_executions"`
	TotalVMs             int64         `json:"total_vms"`
	RecoveredVMs         int64         `json:"recovered_vms"`
	FailedVMs            int64         `json:"failed_vms"`
	AverageExecutionTime time.Duration `json:"average_execution_time"`
	MinExecutionTime     time.Duration `json:"min_execution_time"`
	MaxExecutionTime     time.Duration `json:"max_execution_time"`
	LastExecutionAt      *time.Time    `json:"last_execution_at,omitempty"`
	SuccessRate          float64       `json:"success_rate"`
	LastUpdated          time.Time     `json:"last_updated"`
}

// GlobalRecoveryStats contains global recovery statistics
type GlobalRecoveryStats struct {
	TimeRange            TimeRange     `json:"time_range"`
	TotalTenants         int64         `json:"total_tenants"`
	TotalPlans           int64         `json:"total_plans"`
	ActivePlans          int64         `json:"active_plans"`
	TotalSequences       int64         `json:"total_sequences"`
	TotalExecutions      int64         `json:"total_executions"`
	CompletedExecutions  int64         `json:"completed_executions"`
	FailedExecutions     int64         `json:"failed_executions"`
	CancelledExecutions  int64         `json:"cancelled_executions"`
	RunningExecutions    int64         `json:"running_executions"`
	QueuedExecutions     int64         `json:"queued_executions"`
	TotalVMs             int64         `json:"total_vms"`
	RecoveredVMs         int64         `json:"recovered_vms"`
	FailedVMs            int64         `json:"failed_vms"`
	SuccessRate          float64       `json:"success_rate"`
	AverageExecutionTime time.Duration `json:"average_execution_time"`
	LastUpdated          time.Time     `json:"last_updated"`
}

// RecoverySystemHealth contains recovery system health information
type RecoverySystemHealth struct {
	Status          HealthStatus    `json:"status"`
	HostHealth      []HostHealth    `json:"host_health"`
	StorageHealth   []StorageHealth `json:"storage_health"`
	NetworkHealth   []NetworkHealth `json:"network_health"`
	ExecutionHealth ExecutionHealth `json:"execution_health"`
	ResourceUsage   ResourceUsage   `json:"resource_usage"`
	ErrorRate       float64         `json:"error_rate"`
	ResponseTime    time.Duration   `json:"response_time"`
	LastHealthCheck time.Time       `json:"last_health_check"`
	Issues          []HealthIssue   `json:"issues"`
}

// HostHealth contains health information for a host
type HostHealth struct {
	HostID       string       `json:"host_id"`
	Name         string       `json:"name"`
	Status       HealthStatus `json:"status"`
	CPUUsage     float64      `json:"cpu_usage"`
	MemoryUsage  float64      `json:"memory_usage"`
	StorageUsage float64      `json:"storage_usage"`
	NetworkUsage float64      `json:"network_usage"`
	ActiveVMs    int          `json:"active_vms"`
	MaxVMs       int          `json:"max_vms"`
	LastSeen     time.Time    `json:"last_seen"`
	Version      string       `json:"version"`
}

// StorageHealth contains health information for storage
type StorageHealth struct {
	StorageID   string       `json:"storage_id"`
	Name        string       `json:"name"`
	Type        string       `json:"type"`
	Status      HealthStatus `json:"status"`
	Capacity    int64        `json:"capacity"`
	Used        int64        `json:"used"`
	Available   int64        `json:"available"`
	Performance float64      `json:"performance"`
	LastSync    *time.Time   `json:"last_sync,omitempty"`
	LastSeen    time.Time    `json:"last_seen"`
}

// NetworkHealth contains health information for network
type NetworkHealth struct {
	NetworkID     string       `json:"network_id"`
	Name          string       `json:"name"`
	Type          string       `json:"type"`
	Status        HealthStatus `json:"status"`
	Bandwidth     int64        `json:"bandwidth"`
	UsedBandwidth int64        `json:"used_bandwidth"`
	Latency       int          `json:"latency"`
	LastSeen      time.Time    `json:"last_seen"`
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

// ActiveRecovery represents an active recovery
type ActiveRecovery struct {
	ID        string            `json:"id"`
	PlanID    string            `json:"plan_id"`
	TenantID  string            `json:"tenant_id"`
	Status    ExecutionStatus   `json:"status"`
	Progress  ExecutionProgress `json:"progress"`
	StartTime time.Time         `json:"start_time"`
	Duration  time.Duration     `json:"duration"`
	Eta       *time.Time        `json:"eta,omitempty"`
}

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
	HostID      string     `json:"host_id,omitempty"`
	DetectedAt  time.Time  `json:"detected_at"`
	Resolved    bool       `json:"resolved"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
}

// ExecutionStatus represents the status of execution
type ExecutionStatus string

const (
	ExecutionStatusPending   ExecutionStatus = "pending"
	ExecutionStatusRunning   ExecutionStatus = "running"
	ExecutionStatusCompleted ExecutionStatus = "completed"
	ExecutionStatusFailed    ExecutionStatus = "failed"
	ExecutionStatusCancelled ExecutionStatus = "cancelled"
	ExecutionStatusExpired   ExecutionStatus = "expired"
	ExecutionStatusRetrying  ExecutionStatus = "retrying"
)

// StepStatus represents the status of step execution
type StepStatus string

const (
	StepStatusPending   StepStatus = "pending"
	StepStatusRunning   StepStatus = "running"
	StepStatusCompleted StepStatus = "completed"
	StepStatusFailed    StepStatus = "failed"
	StepStatusSkipped   StepStatus = "skipped"
)

// InMemoryRecoveryPlanManager implements RecoveryPlanManager interface
type InMemoryRecoveryPlanManager struct {
	plans          map[string]*RecoveryPlan
	sequences      map[string]*VMSequence
	executions     map[string]*RecoveryExecution
	stats          map[string]*RecoveryStats
	planStats      map[string]*PlanStats
	globalStats    *GlobalRecoveryStats
	tenantManager  multitenancy.TenantManager
	storageManager *storage.Engine
	mutex          sync.RWMutex
}

// NewInMemoryRecoveryPlanManager creates a new in-memory recovery plan manager
func NewInMemoryRecoveryPlanManager(
	tenantMgr multitenancy.TenantManager,
	storageMgr *storage.Engine,
) *InMemoryRecoveryPlanManager {
	return &InMemoryRecoveryPlanManager{
		plans:          make(map[string]*RecoveryPlan),
		sequences:      make(map[string]*VMSequence),
		executions:     make(map[string]*RecoveryExecution),
		stats:          make(map[string]*RecoveryStats),
		planStats:      make(map[string]*PlanStats),
		globalStats:    &GlobalRecoveryStats{LastUpdated: time.Now()},
		tenantManager:  tenantMgr,
		storageManager: storageMgr,
	}
}

// CreateRecoveryPlan creates a new recovery plan
func (m *InMemoryRecoveryPlanManager) CreateRecoveryPlan(ctx context.Context, request *RecoveryPlanRequest) (*RecoveryPlan, error) {
	if err := m.validateTenantAccess(ctx, request.TenantID); err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	plan := &RecoveryPlan{
		ID:            generatePlanID(),
		Name:          request.Name,
		TenantID:      request.TenantID,
		Description:   request.Description,
		Enabled:       request.Enabled,
		Type:          request.Type,
		Priority:      request.Priority,
		VMSequences:   request.VMSequences,
		RecoveryOrder: request.RecoveryOrder,
		Dependencies:  request.Dependencies,
		Configuration: request.Configuration,
		Notifications: request.Notifications,
		Retention:     request.Retention,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Metadata:      request.Metadata,
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

// GetRecoveryPlan retrieves a recovery plan by ID
func (m *InMemoryRecoveryPlanManager) GetRecoveryPlan(ctx context.Context, planID string) (*RecoveryPlan, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	plan, exists := m.plans[planID]
	if !exists {
		return nil, fmt.Errorf("recovery plan %s not found", planID)
	}

	if err := m.validateTenantAccess(ctx, plan.TenantID); err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	return plan, nil
}

// ListRecoveryPlans lists recovery plans with optional filtering
func (m *InMemoryRecoveryPlanManager) ListRecoveryPlans(ctx context.Context, filter *RecoveryPlanFilter) ([]*RecoveryPlan, error) {
	if filter != nil && filter.TenantID != "" {
		if err := m.validateTenantAccess(ctx, filter.TenantID); err != nil {
			return nil, fmt.Errorf("access denied: %w", err)
		}
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var results []*RecoveryPlan
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

// UpdateRecoveryPlan updates an existing recovery plan
func (m *InMemoryRecoveryPlanManager) UpdateRecoveryPlan(ctx context.Context, planID string, request *UpdateRecoveryPlanRequest) (*RecoveryPlan, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	plan, exists := m.plans[planID]
	if !exists {
		return nil, fmt.Errorf("recovery plan %s not found", planID)
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
	if request.VMSequences != nil {
		plan.VMSequences = request.VMSequences
	}
	if request.RecoveryOrder != nil {
		plan.RecoveryOrder = *request.RecoveryOrder
	}
	if request.Dependencies != nil {
		plan.Dependencies = request.Dependencies
	}
	if request.Configuration != nil {
		plan.Configuration = *request.Configuration
	}
	if request.Notifications != nil {
		plan.Notifications = *request.Notifications
	}
	if request.Retention != nil {
		plan.Retention = *request.Retention
	}
	if request.Metadata != nil {
		plan.Metadata = request.Metadata
	}

	plan.UpdatedAt = time.Now()

	return plan, nil
}

// DeleteRecoveryPlan deletes a recovery plan
func (m *InMemoryRecoveryPlanManager) DeleteRecoveryPlan(ctx context.Context, planID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	plan, exists := m.plans[planID]
	if !exists {
		return fmt.Errorf("recovery plan %s not found", planID)
	}

	if err := m.validateTenantAccess(ctx, plan.TenantID); err != nil {
		return fmt.Errorf("access denied: %w", err)
	}

	delete(m.plans, planID)
	delete(m.planStats, planID)

	return nil
}

// EnableRecoveryPlan enables a recovery plan
func (m *InMemoryRecoveryPlanManager) EnableRecoveryPlan(ctx context.Context, planID string) error {
	_, err := m.UpdateRecoveryPlan(ctx, planID, &UpdateRecoveryPlanRequest{
		Enabled: &[]bool{true}[0],
	})
	return err
}

// DisableRecoveryPlan disables a recovery plan
func (m *InMemoryRecoveryPlanManager) DisableRecoveryPlan(ctx context.Context, planID string) error {
	_, err := m.UpdateRecoveryPlan(ctx, planID, &UpdateRecoveryPlanRequest{
		Enabled: &[]bool{false}[0],
	})
	return err
}

// StartRecovery starts a recovery execution
func (m *InMemoryRecoveryPlanManager) StartRecovery(ctx context.Context, request *RecoveryRequest) (*RecoveryExecution, error) {
	if err := m.validateTenantAccess(ctx, request.TenantID); err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if plan exists and is enabled
	plan, exists := m.plans[request.PlanID]
	if !exists {
		return nil, fmt.Errorf("recovery plan %s not found", request.PlanID)
	}

	if !plan.Enabled {
		return nil, fmt.Errorf("recovery plan %s is not enabled", request.PlanID)
	}

	execution := &RecoveryExecution{
		ID:            generateExecutionID(),
		PlanID:        request.PlanID,
		TenantID:      request.TenantID,
		TriggerType:   request.TriggerType,
		TriggerReason: request.TriggerReason,
		Status:        ExecutionStatusPending,
		Scope:         request.Scope,
		Options:       request.Options,
		Progress: ExecutionProgress{
			Percentage:         0,
			CurrentSequence:    "queued",
			TotalSequences:     len(plan.VMSequences),
			CompletedSequences: 0,
			TotalVMs:           int(m.countTotalVMs(plan)),
			CompletedVMs:       0,
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

	m.executions[execution.ID] = execution

	// Start execution in background
	go m.runRecoveryExecution(ctx, execution)

	return execution, nil
}

// GetRecoveryExecution retrieves a recovery execution by ID
func (m *InMemoryRecoveryPlanManager) GetRecoveryExecution(ctx context.Context, executionID string) (*RecoveryExecution, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	execution, exists := m.executions[executionID]
	if !exists {
		return nil, fmt.Errorf("recovery execution %s not found", executionID)
	}

	if err := m.validateTenantAccess(ctx, execution.TenantID); err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	return execution, nil
}

// ListRecoveryExecutions lists recovery executions with optional filtering
func (m *InMemoryRecoveryPlanManager) ListRecoveryExecutions(ctx context.Context, filter *ExecutionFilter) ([]*RecoveryExecution, error) {
	if filter != nil && filter.TenantID != "" {
		if err := m.validateTenantAccess(ctx, filter.TenantID); err != nil {
			return nil, fmt.Errorf("access denied: %w", err)
		}
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var results []*RecoveryExecution
	for _, execution := range m.executions {
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

// CancelRecoveryExecution cancels a recovery execution
func (m *InMemoryRecoveryPlanManager) CancelRecoveryExecution(ctx context.Context, executionID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	execution, exists := m.executions[executionID]
	if !exists {
		return fmt.Errorf("recovery execution %s not found", executionID)
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

// RerunRecoveryExecution reruns a recovery execution
func (m *InMemoryRecoveryPlanManager) RerunRecoveryExecution(ctx context.Context, executionID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	execution, exists := m.executions[executionID]
	if !exists {
		return fmt.Errorf("recovery execution %s not found", executionID)
	}

	if err := m.validateTenantAccess(ctx, execution.TenantID); err != nil {
		return fmt.Errorf("access denied: %w", err)
	}

	// Create new execution with same configuration
	newExecution := &RecoveryExecution{
		ID:            generateExecutionID(),
		PlanID:        execution.PlanID,
		TenantID:      execution.TenantID,
		TriggerType:   execution.TriggerType,
		TriggerReason: execution.TriggerReason,
		Status:        ExecutionStatusPending,
		Scope:         execution.Scope,
		Options:       execution.Options,
		Progress: ExecutionProgress{
			Percentage:         0,
			CurrentSequence:    "queued",
			TotalSequences:     len(execution.Sequences),
			CompletedSequences: 0,
			TotalVMs:           int(m.countTotalVMsFromExecution(execution)),
			CompletedVMs:       0,
		},
		Timing: ExecutionTiming{
			TimeoutDuration: execution.Timing.TimeoutDuration,
		},
		RollbackInfo: RollbackInfo{
			Available: true,
		},
		CreatedAt: time.Now(),
		Metadata:  execution.Metadata,
	}

	m.executions[newExecution.ID] = newExecution

	// Start execution in background
	go m.runRecoveryExecution(ctx, newExecution)

	return nil
}

// CreateVMSequence creates a new VM sequence
func (m *InMemoryRecoveryPlanManager) CreateVMSequence(ctx context.Context, request *VMSequenceRequest) (*VMSequence, error) {
	if err := m.validateTenantAccess(ctx, request.TenantID); err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	sequence := &VMSequence{
		ID:            generateSequenceID(),
		Name:          request.Name,
		TenantID:      request.TenantID,
		Description:   request.Description,
		Enabled:       request.Enabled,
		Type:          request.Type,
		VMs:           request.VMs,
		RecoveryOrder: request.RecoveryOrder,
		Dependencies:  request.Dependencies,
		Configuration: request.Configuration,
		NetworkConfig: request.NetworkConfig,
		StorageConfig: request.StorageConfig,
		Validation:    request.Validation,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Metadata:      request.Metadata,
	}

	m.sequences[sequence.ID] = sequence

	return sequence, nil
}

// GetVMSequence retrieves a VM sequence by ID
func (m *InMemoryRecoveryPlanManager) GetVMSequence(ctx context.Context, sequenceID string) (*VMSequence, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	sequence, exists := m.sequences[sequenceID]
	if !exists {
		return nil, fmt.Errorf("VM sequence %s not found", sequenceID)
	}

	if err := m.validateTenantAccess(ctx, sequence.TenantID); err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	return sequence, nil
}

// ListVMSequences lists VM sequences with optional filtering
func (m *InMemoryRecoveryPlanManager) ListVMSequences(ctx context.Context, filter *VMSequenceFilter) ([]*VMSequence, error) {
	if filter != nil && filter.TenantID != "" {
		if err := m.validateTenantAccess(ctx, filter.TenantID); err != nil {
			return nil, fmt.Errorf("access denied: %w", err)
		}
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var results []*VMSequence
	for _, sequence := range m.sequences {
		if filter != nil {
			if filter.TenantID != "" && sequence.TenantID != filter.TenantID {
				continue
			}
			if filter.Enabled != nil && sequence.Enabled != *filter.Enabled {
				continue
			}
			if filter.Type != "" && sequence.Type != filter.Type {
				continue
			}
			if filter.CreatedAfter != nil && sequence.CreatedAt.Before(*filter.CreatedAfter) {
				continue
			}
			if filter.CreatedBefore != nil && sequence.CreatedAt.After(*filter.CreatedBefore) {
				continue
			}
		}

		// Check tenant access
		if err := m.validateTenantAccess(ctx, sequence.TenantID); err != nil {
			continue
		}

		results = append(results, sequence)
	}

	return results, nil
}

// UpdateVMSequence updates an existing VM sequence
func (m *InMemoryRecoveryPlanManager) UpdateVMSequence(ctx context.Context, sequenceID string, request *UpdateVMSequenceRequest) (*VMSequence, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	sequence, exists := m.sequences[sequenceID]
	if !exists {
		return nil, fmt.Errorf("VM sequence %s not found", sequenceID)
	}

	if err := m.validateTenantAccess(ctx, sequence.TenantID); err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	// Update fields
	if request.Name != nil {
		sequence.Name = *request.Name
	}
	if request.Description != nil {
		sequence.Description = *request.Description
	}
	if request.Enabled != nil {
		sequence.Enabled = *request.Enabled
	}
	if request.Type != nil {
		sequence.Type = *request.Type
	}
	if request.VMs != nil {
		sequence.VMs = request.VMs
	}
	if request.RecoveryOrder != nil {
		sequence.RecoveryOrder = *request.RecoveryOrder
	}
	if request.Dependencies != nil {
		sequence.Dependencies = request.Dependencies
	}
	if request.Configuration != nil {
		sequence.Configuration = *request.Configuration
	}
	if request.NetworkConfig != nil {
		sequence.NetworkConfig = *request.NetworkConfig
	}
	if request.StorageConfig != nil {
		sequence.StorageConfig = *request.StorageConfig
	}
	if request.Validation != nil {
		sequence.Validation = *request.Validation
	}
	if request.Metadata != nil {
		sequence.Metadata = request.Metadata
	}

	sequence.UpdatedAt = time.Now()

	return sequence, nil
}

// DeleteVMSequence deletes a VM sequence
func (m *InMemoryRecoveryPlanManager) DeleteVMSequence(ctx context.Context, sequenceID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	sequence, exists := m.sequences[sequenceID]
	if !exists {
		return fmt.Errorf("VM sequence %s not found", sequenceID)
	}

	if err := m.validateTenantAccess(ctx, sequence.TenantID); err != nil {
		return fmt.Errorf("access denied: %w", err)
	}

	delete(m.sequences, sequenceID)

	return nil
}

// GetRecoveryStats retrieves recovery statistics for a tenant
func (m *InMemoryRecoveryPlanManager) GetRecoveryStats(ctx context.Context, tenantID string, timeRange TimeRange) (*RecoveryStats, error) {
	if err := m.validateTenantAccess(ctx, tenantID); err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := &RecoveryStats{
		TenantID:    tenantID,
		TimeRange:   timeRange,
		LastUpdated: time.Now(),
	}

	var totalExecutionTime time.Duration
	var executionTimes []time.Duration

	for _, plan := range m.plans {
		if plan.TenantID != tenantID {
			continue
		}

		stats.TotalPlans++
		if plan.Enabled {
			stats.ActivePlans++
		}

		stats.TotalSequences += int64(len(plan.VMSequences))
	}

	for _, execution := range m.executions {
		if execution.TenantID != tenantID {
			continue
		}

		// Filter by time range
		if execution.CreatedAt.Before(timeRange.From) || execution.CreatedAt.After(timeRange.To) {
			continue
		}

		stats.TotalExecutions++

		switch execution.Status {
		case ExecutionStatusCompleted:
			stats.CompletedExecutions++
		case ExecutionStatusFailed:
			stats.FailedExecutions++
		case ExecutionStatusCancelled:
			stats.CancelledExecutions++
		case ExecutionStatusRunning:
			stats.RunningExecutions++
		case ExecutionStatusPending:
			stats.QueuedExecutions++
		}

		// Calculate execution time for completed executions
		if execution.Timing.ExecutionDuration > 0 {
			totalExecutionTime += execution.Timing.ExecutionDuration
			executionTimes = append(executionTimes, execution.Timing.ExecutionDuration)
		}

		// Count VMs
		for _, sequence := range execution.Sequences {
			stats.TotalVMs += int64(len(sequence.VMExecutions))
			for _, vmExec := range sequence.VMExecutions {
				if vmExec.Status == VMStatusCompleted {
					stats.RecoveredVMs++
				} else if vmExec.Status == VMStatusFailed {
					stats.FailedVMs++
				}
			}
		}
	}

	// Calculate success rate
	if stats.TotalExecutions > 0 {
		stats.SuccessRate = float64(stats.CompletedExecutions) / float64(stats.TotalExecutions)
	}

	// Calculate average, min, max execution times
	if len(executionTimes) > 0 {
		stats.AverageExecutionTime = totalExecutionTime / time.Duration(len(executionTimes))
		stats.MinExecutionTime = executionTimes[0]
		stats.MaxExecutionTime = executionTimes[0]

		for _, et := range executionTimes {
			if et < stats.MinExecutionTime {
				stats.MinExecutionTime = et
			}
			if et > stats.MaxExecutionTime {
				stats.MaxExecutionTime = et
			}
		}
	}

	stats.TotalExecutionTime = totalExecutionTime

	return stats, nil
}

// GetPlanStats retrieves statistics for a specific recovery plan
func (m *InMemoryRecoveryPlanManager) GetPlanStats(ctx context.Context, planID string, timeRange TimeRange) (*PlanStats, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	plan, exists := m.plans[planID]
	if !exists {
		return nil, fmt.Errorf("recovery plan %s not found", planID)
	}

	if err := m.validateTenantAccess(ctx, plan.TenantID); err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	stats := &PlanStats{
		PlanID:      planID,
		TimeRange:   timeRange,
		LastUpdated: time.Now(),
	}

	var totalExecutionTime time.Duration
	var executionTimes []time.Duration

	for _, execution := range m.executions {
		if execution.PlanID != planID {
			continue
		}

		// Filter by time range
		if execution.CreatedAt.Before(timeRange.From) || execution.CreatedAt.After(timeRange.To) {
			continue
		}

		stats.TotalExecutions++

		switch execution.Status {
		case ExecutionStatusCompleted:
			stats.CompletedExecutions++
		case ExecutionStatusFailed:
			stats.FailedExecutions++
		case ExecutionStatusCancelled:
			stats.CancelledExecutions++
		}

		// Calculate execution time
		if execution.Timing.ExecutionDuration > 0 {
			totalExecutionTime += execution.Timing.ExecutionDuration
			executionTimes = append(executionTimes, execution.Timing.ExecutionDuration)
		}

		// Track last execution
		if stats.LastExecutionAt == nil || execution.CreatedAt.After(*stats.LastExecutionAt) {
			stats.LastExecutionAt = &execution.CreatedAt
		}

		// Count VMs
		for _, sequence := range execution.Sequences {
			stats.TotalVMs += int64(len(sequence.VMExecutions))
			for _, vmExec := range sequence.VMExecutions {
				if vmExec.Status == VMStatusCompleted {
					stats.RecoveredVMs++
				} else if vmExec.Status == VMStatusFailed {
					stats.FailedVMs++
				}
			}
		}
	}

	// Calculate success rate
	if stats.TotalExecutions > 0 {
		stats.SuccessRate = float64(stats.CompletedExecutions) / float64(stats.TotalExecutions)
	}

	// Calculate average times
	if len(executionTimes) > 0 {
		stats.AverageExecutionTime = totalExecutionTime / time.Duration(len(executionTimes))
		stats.MinExecutionTime = executionTimes[0]
		stats.MaxExecutionTime = executionTimes[0]

		for _, et := range executionTimes {
			if et < stats.MinExecutionTime {
				stats.MinExecutionTime = et
			}
			if et > stats.MaxExecutionTime {
				stats.MaxExecutionTime = et
			}
		}
	}

	return stats, nil
}

// GetGlobalStats retrieves global recovery statistics
func (m *InMemoryRecoveryPlanManager) GetGlobalStats(ctx context.Context, timeRange TimeRange) (*GlobalRecoveryStats, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := &GlobalRecoveryStats{
		TimeRange:   timeRange,
		LastUpdated: time.Now(),
	}

	tenants := make(map[string]bool)

	for _, plan := range m.plans {
		tenants[plan.TenantID] = true
		if plan.Enabled {
			stats.ActivePlans++
		}

		stats.TotalSequences += int64(len(plan.VMSequences))
	}

	stats.TotalTenants = int64(len(tenants))
	stats.TotalPlans = int64(len(m.plans))

	var totalExecutionTime time.Duration
	var executionTimes []time.Duration

	for _, execution := range m.executions {
		// Filter by time range
		if execution.CreatedAt.Before(timeRange.From) || execution.CreatedAt.After(timeRange.To) {
			continue
		}

		stats.TotalExecutions++
		tenants[execution.TenantID] = true

		switch execution.Status {
		case ExecutionStatusCompleted:
			stats.CompletedExecutions++
		case ExecutionStatusFailed:
			stats.FailedExecutions++
		case ExecutionStatusCancelled:
			stats.CancelledExecutions++
		case ExecutionStatusRunning:
			stats.RunningExecutions++
		case ExecutionStatusPending:
			stats.QueuedExecutions++
		}

		// Calculate execution time
		if execution.Timing.ExecutionDuration > 0 {
			totalExecutionTime += execution.Timing.ExecutionDuration
			executionTimes = append(executionTimes, execution.Timing.ExecutionDuration)
		}

		// Count VMs
		for _, sequence := range execution.Sequences {
			stats.TotalVMs += int64(len(sequence.VMExecutions))
			for _, vmExec := range sequence.VMExecutions {
				if vmExec.Status == VMStatusCompleted {
					stats.RecoveredVMs++
				} else if vmExec.Status == VMStatusFailed {
					stats.FailedVMs++
				}
			}
		}
	}

	// Calculate success rate
	if stats.TotalExecutions > 0 {
		stats.SuccessRate = float64(stats.CompletedExecutions) / float64(stats.TotalExecutions)
	}

	// Calculate average time
	if len(executionTimes) > 0 {
		stats.AverageExecutionTime = totalExecutionTime / time.Duration(len(executionTimes))
	}

	return stats, nil
}

// GetRecoverySystemHealth retrieves recovery system health information
func (m *InMemoryRecoveryPlanManager) GetRecoverySystemHealth(ctx context.Context) (*RecoverySystemHealth, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	health := &RecoverySystemHealth{
		Status:        HealthStatusHealthy,
		HostHealth:    []HostHealth{},
		StorageHealth: []StorageHealth{},
		NetworkHealth: []NetworkHealth{},
		ExecutionHealth: ExecutionHealth{
			Status:          HealthStatusHealthy,
			RunningJobs:     0,
			QueuedJobs:      0,
			FailedJobs24h:   0,
			SuccessRate:     0.95,
			AverageDuration: 10 * time.Minute,
		},
		ResourceUsage: ResourceUsage{
			CPUUsage:     15.5,
			MemoryUsage:  52.3,
			StorageUsage: 38.7,
			NetworkUsage: 12.1,
		},
		ErrorRate:       0.02,
		ResponseTime:    250 * time.Millisecond,
		LastHealthCheck: time.Now(),
		Issues:          []HealthIssue{},
	}

	// Count running and queued jobs
	runningJobs := 0
	pendingJobs := 0
	failedJobs24h := 0
	twentyFourHoursAgo := time.Now().Add(-24 * time.Hour)

	for _, execution := range m.executions {
		if execution.Status == ExecutionStatusRunning {
			runningJobs++
		}
		if execution.Status == ExecutionStatusPending {
			pendingJobs++
		}
		if execution.Status == ExecutionStatusFailed && execution.CreatedAt.After(twentyFourHoursAgo) {
			failedJobs24h++
		}
	}

	health.ExecutionHealth.RunningJobs = runningJobs
	health.ExecutionHealth.QueuedJobs = pendingJobs
	health.ExecutionHealth.FailedJobs24h = failedJobs24h

	// Add mock host health
	health.HostHealth = append(health.HostHealth, HostHealth{
		HostID:       "host-1",
		Name:         "esxi-host-1",
		Status:       HealthStatusHealthy,
		CPUUsage:     18.2,
		MemoryUsage:  45.8,
		StorageUsage: 32.1,
		NetworkUsage: 15.5,
		ActiveVMs:    5,
		MaxVMs:       20,
		LastSeen:     time.Now().Add(-2 * time.Minute),
		Version:      "1.0.0",
	})

	// Add mock storage health
	health.StorageHealth = append(health.StorageHealth, StorageHealth{
		StorageID:   "storage-1",
		Name:        "datastore-1",
		Type:        "vmfs",
		Status:      HealthStatusHealthy,
		Capacity:    10000000000000, // 10TB
		Used:        5000000000000,  // 5TB
		Available:   5000000000000,  // 5TB
		Performance: 85.5,
		LastSync:    &[]time.Time{time.Now().Add(-10 * time.Minute)}[0],
		LastSeen:    time.Now().Add(-1 * time.Minute),
	})

	// Add mock network health
	health.NetworkHealth = append(health.NetworkHealth, NetworkHealth{
		NetworkID:     "network-1",
		Name:          "vm-network",
		Type:          "vlan",
		Status:        HealthStatusHealthy,
		Bandwidth:     10000,
		UsedBandwidth: 2500,
		Latency:       5,
		LastSeen:      time.Now().Add(-30 * time.Second),
	})

	return health, nil
}

// GetActiveRecoveries retrieves active recoveries
func (m *InMemoryRecoveryPlanManager) GetActiveRecoveries(ctx context.Context) ([]*ActiveRecovery, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var activeRecoveries []*ActiveRecovery

	for _, execution := range m.executions {
		if execution.Status == ExecutionStatusRunning || execution.Status == ExecutionStatusPending {
			// Check tenant access
			if err := m.validateTenantAccess(ctx, execution.TenantID); err != nil {
				continue
			}

			activeRecovery := &ActiveRecovery{
				ID:       execution.ID,
				PlanID:   execution.PlanID,
				TenantID: execution.TenantID,
				Status:   execution.Status,
				Progress: execution.Progress,
			}

			if execution.StartedAt != nil {
				activeRecovery.StartTime = *execution.StartedAt
				activeRecovery.Duration = time.Since(*execution.StartedAt)
			}

			activeRecoveries = append(activeRecoveries, activeRecovery)
		}
	}

	return activeRecoveries, nil
}

// Helper methods

func (m *InMemoryRecoveryPlanManager) runRecoveryExecution(ctx context.Context, execution *RecoveryExecution) {
	m.mutex.Lock()
	execution.Status = ExecutionStatusRunning
	now := time.Now()
	execution.StartedAt = &now
	execution.Progress.CurrentSequence = "starting_recovery"
	m.mutex.Unlock()

	// Simulate recovery execution
	time.Sleep(4 * time.Second)

	m.mutex.Lock()
	execution.Status = ExecutionStatusCompleted
	completedAt := time.Now()
	execution.CompletedAt = &completedAt
	execution.Progress.Percentage = 100
	execution.Progress.CurrentSequence = "completed"
	execution.Progress.CompletedSequences = len(execution.Sequences)
	execution.Progress.CompletedVMs = execution.Progress.TotalVMs

	// Set timing
	execution.Timing.ExecutionDuration = completedAt.Sub(*execution.StartedAt)
	execution.Timing.TotalDuration = completedAt.Sub(execution.CreatedAt)

	// Set mock results
	execution.Results = &RecoveryResults{
		Success:            true,
		TotalSequences:     len(execution.Sequences),
		CompletedSequences: len(execution.Sequences),
		FailedSequences:    0,
		TotalVMs:           execution.Progress.TotalVMs,
		CompletedVMs:       execution.Progress.CompletedVMs,
		FailedVMs:          0,
		ExecutionTime:      execution.Timing.ExecutionDuration,
		Summary: ExecutionSummary{
			SequencesRecovered: len(execution.Sequences),
			VMsRecovered:       execution.Progress.CompletedVMs,
			DataValidated:      true,
			PerformanceOK:      true,
			NetworkOK:          true,
			StorageOK:          true,
		},
	}
	m.mutex.Unlock()
}

func (m *InMemoryRecoveryPlanManager) countTotalVMs(plan *RecoveryPlan) int64 {
	var total int64
	for _, sequence := range plan.VMSequences {
		total += int64(len(sequence.VMs))
	}
	return total
}

func (m *InMemoryRecoveryPlanManager) countTotalVMsFromExecution(execution *RecoveryExecution) int64 {
	var total int64
	for _, sequence := range execution.Sequences {
		total += int64(len(sequence.VMExecutions))
	}
	return total
}

func (m *InMemoryRecoveryPlanManager) validateTenantAccess(ctx context.Context, tenantID string) error {
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
	return fmt.Sprintf("recovery-plan-%s", uuid.New().String()[:8])
}

func generateSequenceID() string {
	return fmt.Sprintf("vm-sequence-%s", uuid.New().String()[:8])
}

func generateExecutionID() string {
	return fmt.Sprintf("recovery-exec-%s", uuid.New().String()[:8])
}
