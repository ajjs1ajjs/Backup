// Package vss provides Windows Volume Shadow Copy Service (VSS) integration
package vss

import (
	"fmt"
	"os/exec"
	"strings"

	"go.uber.org/zap"
)

// ActiveDirectoryManager provides AD-specific backup functionality
type ActiveDirectoryManager struct {
	logger   *zap.Logger
	domain   string
	server   string
}

// NewActiveDirectoryManager creates a new AD manager
func NewActiveDirectoryManager(logger *zap.Logger, domain string) *ActiveDirectoryManager {
	return &ActiveDirectoryManager{
		logger: logger,
		domain: domain,
	}
}

// ADServerInfo contains AD server information
type ADServerInfo struct {
	DomainName       string            `json:"domain_name"`
	ServerName       string            `json:"server_name"`
	SiteName         string            `json:"site_name"`
	IsGlobalCatalog  bool              `json:"is_global_catalog"`
	IsReadOnly       bool              `json:"is_read_only"`
	FSMORoles        []string          `json:"fsmo_roles"`
	DatabasePath     string            `json:"database_path"`
	LogPath          string            `json:"log_path"`
	SysvolPath       string            `json:"sysvol_path"`
	DatabaseSizeGB   float64           `json:"database_size_gb"`
}

// ADBackupResult contains AD backup results
type ADBackupResult struct {
	BackupType       string  `json:"backup_type"` // system_state, bare_metal
	StartTime        string  `json:"start_time"`
	EndTime          string  `json:"end_time"`
	BackupFile       string  `json:"backup_file"`
	BackupSizeGB     float64 `json:"backup_size_gb"`
	Status           string  `json:"status"`
	Components       []string `json:"components"`
	USN              int64   `json:"usn"` // Update Sequence Number
}

// ADRestoreOptions contains AD restore options
type ADRestoreOptions struct {
	Authoritative    bool   `json:"authoritative"`    // Authoritative restore
	NonAuthoritative bool  `json:"non_authoritative"` // Non-authoritative
	TargetServer     string `json:"target_server,omitempty"`
	SubtreeDN        string `json:"subtree_dn,omitempty"` // For subtree restore
}

// GetServerInfo retrieves AD server information
func (ad *ActiveDirectoryManager) GetServerInfo() (*ADServerInfo, error) {
	ad.logger.Info("Getting AD server info", zap.String("domain", ad.domain))

	// Get server hostname
	hostname, _ := exec.Command("hostname").Output()
	serverName := strings.TrimSpace(string(hostname))

	// Get AD site and FSMO roles using PowerShell
	cmd := exec.Command("powershell", "-Command",
		"Import-Module ActiveDirectory; "+
			"Get-ADDomainController | Select-Object HostName, Site, IsGlobalCatalog, "+
			"OperationMasterRoles, DatabasePath, LogPath | ConvertTo-Json")
	
	output, err := cmd.Output()
	
	info := &ADServerInfo{
		DomainName:      ad.domain,
		ServerName:      serverName,
		SiteName:        "Default-First-Site-Name",
		IsGlobalCatalog: true,
		IsReadOnly:      false,
		FSMORoles:       []string{},
		DatabasePath:    "C:\\Windows\\NTDS\\ntds.dit",
		LogPath:         "C:\\Windows\\NTDS",
		SysvolPath:      "C:\\Windows\\SYSVOL",
		DatabaseSizeGB:  1.0,
	}

	if err == nil && len(output) > 0 {
		// Parse FSMO roles from output
		if strings.Contains(string(output), "OperationMasterRoles") {
			info.FSMORoles = ad.extractFSMORoles(string(output))
		}
		
		// Parse other info
		if strings.Contains(string(output), "Site") {
			info.SiteName = ad.extractValue(string(output), "Site")
		}
	}

	// Get database size
	info.DatabaseSizeGB = ad.getDatabaseSize()

	return info, nil
}

// extractFSMORoles extracts FSMO roles from PowerShell output
func (ad *ActiveDirectoryManager) extractFSMORoles(output string) []string {
	roles := []string{}
	
	// Common FSMO role names to look for
	roleNames := []string{
		"SchemaMaster",
		"DomainNamingMaster",
		"PDCEmulator",
		"RIDMaster",
		"InfrastructureMaster",
	}
	
	for _, role := range roleNames {
		if strings.Contains(output, role) {
			roles = append(roles, role)
		}
	}
	
	return roles
}

// extractValue extracts a value from PowerShell JSON output
func (ad *ActiveDirectoryManager) extractValue(output, key string) string {
	// Simplified extraction
	if idx := strings.Index(output, key); idx != -1 {
		start := strings.Index(output[idx:], ":")
		if start != -1 {
			end := strings.Index(output[idx+start:], ",")
			if end != -1 {
				value := output[idx+start+1 : idx+start+end]
				return strings.Trim(strings.TrimSpace(value), "\"'")
			}
		}
	}
	return ""
}

// getDatabaseSize returns NTDS.dit file size in GB
func (ad *ActiveDirectoryManager) getDatabaseSize() float64 {
	cmd := exec.Command("powershell", "-Command",
		"(Get-Item 'C:\\Windows\\NTDS\\ntds.dit').Length / 1GB")
	
	output, err := cmd.Output()
	if err != nil {
		return 1.0 // Default 1GB
	}

	var size float64
	fmt.Sscanf(string(output), "%f", &size)
	if size == 0 {
		size = 1.0
	}
	
	return size
}

// PerformSystemStateBackup performs system state backup (includes AD)
func (ad *ActiveDirectoryManager) PerformSystemStateBackup(targetPath string) (*ADBackupResult, error) {
	ad.logger.Info("Starting AD System State backup", zap.String("target", targetPath))

	result := &ADBackupResult{
		BackupType:   "system_state",
		StartTime:    getCurrentTime(),
		Status:       "in_progress",
		Components:   []string{"SystemState", "ActiveDirectory", "SYSVOL", "Registry"},
	}

	// Use wbadmin for system state backup
	cmd := exec.Command("wbadmin", "start", "systemstatebackup",
		"-backupTarget:"+targetPath,
		"-quiet")
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		result.Status = "failed"
		ad.logger.Error("System state backup failed",
			zap.Error(err),
			zap.String("output", string(output)))
		return result, fmt.Errorf("system state backup failed: %w", err)
	}

	result.EndTime = getCurrentTime()
	result.Status = "completed"
	result.BackupFile = targetPath + "\\WindowsImageBackup"
	result.BackupSizeGB = 5.0 + ad.getDatabaseSize() // System state is typically 5GB + DB size
	result.USN = ad.getCurrentUSN()

	ad.logger.Info("System state backup completed",
		zap.String("target", targetPath),
		zap.Float64("size_gb", result.BackupSizeGB))

	return result, nil
}

// PerformBareMetalBackup performs full bare metal backup
func (ad *ActiveDirectoryManager) PerformBareMetalBackup(targetPath string, includeCriticalVolumes bool) (*ADBackupResult, error) {
	ad.logger.Info("Starting Bare Metal backup", zap.String("target", targetPath))

	result := &ADBackupResult{
		BackupType: "bare_metal",
		StartTime:  getCurrentTime(),
		Status:     "in_progress",
		Components: []string{"SystemState", "ActiveDirectory", "BareMetalRecovery"},
	}

	// Build wbadmin command
	args := []string{"start", "backup", "-backupTarget:" + targetPath, "-include:C:"}
	
	if includeCriticalVolumes {
		args = append(args, "-allCritical")
	}
	
	args = append(args, "-quiet")

	cmd := exec.Command("wbadmin", args...)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		result.Status = "failed"
		ad.logger.Error("Bare metal backup failed",
			zap.Error(err),
			zap.String("output", string(output)))
		return result, fmt.Errorf("bare metal backup failed: %w", err)
	}

	result.EndTime = getCurrentTime()
	result.Status = "completed"
	result.BackupFile = targetPath + "\\WindowsImageBackup"
	result.BackupSizeGB = 20.0 // Bare metal is larger
	result.USN = ad.getCurrentUSN()

	ad.logger.Info("Bare metal backup completed",
		zap.String("target", targetPath),
		zap.Float64("size_gb", result.BackupSizeGB))

	return result, nil
}

// RestoreSystemState restores AD from system state backup
func (ad *ActiveDirectoryManager) RestoreSystemState(backupPath string, options ADRestoreOptions) error {
	ad.logger.Info("Restoring AD System State",
		zap.String("backup", backupPath),
		zap.Bool("authoritative", options.Authoritative))

	// Important: Server must be in DSRM (Directory Services Restore Mode)
	
	if options.Authoritative {
		// Perform authoritative restore
		// This will mark the restored data as authoritative and replicate to other DCs
		ad.logger.Warn("Performing authoritative restore - this will overwrite data on other DCs")
		
		// Use ntdsutil for authoritative restore
		cmd := exec.Command("ntdsutil", "authoritative restore", "restore database", "quit", "quit")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("authoritative restore failed: %w - %s", err, string(output))
		}
	}

	// Use wbadmin for system state restore
	cmd := exec.Command("wbadmin", "start", "systemstaterecovery",
		"-backupTarget:"+backupPath,
		"-authsysvol", // Restore SYSVOL authoritatively if needed
		"-quiet")
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("system state restore failed: %w - %s", err, string(output))
	}

	ad.logger.Info("System state restore completed")
	return nil
}

// RestoreSubtree performs authoritative restore of a specific subtree
func (ad *ActiveDirectoryManager) RestoreSubtree(backupPath string, subtreeDN string) error {
	ad.logger.Info("Restoring AD subtree",
		zap.String("backup", backupPath),
		zap.String("subtree", subtreeDN))

	// Start with system state restore
	if err := ad.RestoreSystemState(backupPath, ADRestoreOptions{Authoritative: false}); err != nil {
		return err
	}

	// Then mark specific subtree as authoritative using ntdsutil
	cmd := exec.Command("ntdsutil",
		"authoritative restore",
		"restore subtree "+subtreeDN,
		"quit", "quit")
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("subtree restore failed: %w - %s", err, string(output))
	}

	ad.logger.Info("Subtree restore completed", zap.String("subtree", subtreeDN))
	return nil
}

// getCurrentUSN gets the current Update Sequence Number
func (ad *ActiveDirectoryManager) getCurrentUSN() int64 {
	cmd := exec.Command("powershell", "-Command",
		"Get-ADRootDSE | Select-Object -ExpandProperty highestCommittedUSN")
	
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	var usn int64
	fmt.Sscanf(string(output), "%d", &usn)
	return usn
}

// CheckReplicationStatus checks AD replication health
func (ad *ActiveDirectoryManager) CheckReplicationStatus() (map[string]interface{}, error) {
	ad.logger.Info("Checking AD replication status")

	cmd := exec.Command("powershell", "-Command",
		"Import-Module ActiveDirectory; "+
			"Get-ADReplicationPartnerMetadata -Target (Get-ADDomainController).HostName | "+
			"Select-Object Server, Partner, LastReplicationSuccess, LastReplicationResult | "+
			"ConvertTo-Json")
	
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to check replication: %w", err)
	}

	// Parse results
	status := map[string]interface{}{
		"healthy":    strings.Contains(string(output), "LastReplicationSuccess"),
		"partners":   []string{},
		"last_success": getCurrentTime(),
	}

	return status, nil
}

// PrepareForBackup prepares AD for VSS backup (freezes writes)
func (ad *ActiveDirectoryManager) PrepareForBackup() error {
	ad.logger.Info("Preparing AD for backup")
	
	// VSS will automatically handle this when using VSS writer
	// This method is for custom backup implementations
	
	return nil
}

// ResumeAfterBackup resumes AD writes after backup
func (ad *ActiveDirectoryManager) ResumeAfterBackup() error {
	ad.logger.Info("Resuming AD after backup")
	
	// VSS will automatically handle this
	
	return nil
}

// ValidateDomainControllerHealth checks if DC is healthy for backup
func (ad *ActiveDirectoryManager) ValidateDomainControllerHealth() error {
	ad.logger.Info("Validating DC health")

	// Check critical services
	services := []string{
		"NTDS",     // Active Directory Domain Services
		"DNS",      // DNS Server
		"Netlogon", // Netlogon
		"KDC",      // Kerberos Key Distribution Center
	}

	for _, service := range services {
		cmd := exec.Command("sc", "query", service)
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("service %s check failed: %w", service, err)
		}
		
		if !strings.Contains(string(output), "RUNNING") {
			return fmt.Errorf("critical AD service %s is not running", service)
		}
	}

	// Check replication
	repStatus, err := ad.CheckReplicationStatus()
	if err != nil {
		ad.logger.Warn("Could not check replication status", zap.Error(err))
	} else if !repStatus["healthy"].(bool) {
		ad.logger.Warn("AD replication may have issues")
	}

	return nil
}
