package guest

// ProcessingMode represents the guest processing mode
type ProcessingMode string

const (
	ProcessingModeDisabled       ProcessingMode = "disabled"
	ProcessingModeCrashConsistent ProcessingMode = "crash_consistent"
	ProcessingModeAppAware       ProcessingMode = "application_aware"
)

// ApplicationType represents a supported application type for guest processing
type ApplicationType string

const (
	ApplicationTypeSQLServer  ApplicationType = "sql_server"
	ApplicationTypeExchange   ApplicationType = "exchange"
	ApplicationTypeActiveDir  ApplicationType = "active_directory"
	ApplicationTypeSharePoint ApplicationType = "sharepoint"
)

// ProcessingRequest represents a request to process a guest VM
type ProcessingRequest struct {
	VMID            string           `json:"vm_id"`
	VMName          string           `json:"vm_name"`
	Mode            ProcessingMode   `json:"mode"`
	Applications    []ApplicationType `json:"applications"`
	CredentialsID   string           `json:"credentials_id"`
	EnableQuiesce   bool             `json:"enable_quiesce"`
	TruncateLogs    bool             `json:"truncate_logs"`
	PreFreezeScript string           `json:"pre_freeze_script"`
	PostThawScript  string           `json:"post_thaw_script"`
	Timeout         int              `json:"timeout_seconds"` // timeout in seconds
}

// ProcessingResult represents the result of guest processing
type ProcessingResult struct {
	Success         bool              `json:"success"`
	TaskID          string            `json:"task_id"`
	VSSSnapshotID   string            `json:"vss_snapshot_id,omitempty"`
	Applications    []string          `json:"applications"`
	ApplicationStatus map[string]string `json:"application_status,omitempty"`
	ErrorMessage    string            `json:"error,omitempty"`
	Duration        int64             `json:"duration_ms"`
}
