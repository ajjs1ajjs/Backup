package recovery

import (
	"context"
	"fmt"
	"testing"
	"time"

	"novabackup/internal/multitenancy"
	"novabackup/internal/storage"
)

// MockTenantManager for Recovery Plans testing
type MockTenantManager struct{}

func (m *MockTenantManager) GetTenantFromContext(ctx context.Context) string {
	if tenantID, ok := ctx.Value("tenant_id").(string); ok {
		return tenantID
	}
	return ""
}

func (m *MockTenantManager) WithTenant(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, "tenant_id", tenantID)
}

func (m *MockTenantManager) GetTenantQuota(tenantID string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"storage_quota": int64(1000000000), // 1GB
		"backup_quota":  int(10),
	}, nil
}

func (m *MockTenantManager) CreateTenant(ctx context.Context, tenant *multitenancy.Tenant) error {
	return nil
}

func (m *MockTenantManager) GetTenant(ctx context.Context, tenantID string) (*multitenancy.Tenant, error) {
	return &multitenancy.Tenant{
		ID:   tenantID,
		Name: "Test Tenant",
	}, nil
}

func (m *MockTenantManager) UpdateTenant(ctx context.Context, tenant *multitenancy.Tenant) error {
	return nil
}

func (m *MockTenantManager) DeleteTenant(ctx context.Context, tenantID string) error {
	return nil
}

func (m *MockTenantManager) ListTenants(ctx context.Context) ([]multitenancy.Tenant, error) {
	return []multitenancy.Tenant{}, nil
}

func (m *MockTenantManager) CheckQuota(ctx context.Context, tenantID string, quotaType multitenancy.QuotaType, amount int64) (bool, error) {
	return true, nil
}

func (m *MockTenantManager) GetQuotaUsage(ctx context.Context, tenantID string) (*multitenancy.QuotaUsage, error) {
	return &multitenancy.QuotaUsage{
		BackupsCount:  int64(5),
		StorageUsedGB: 0.5, // 0.5GB
	}, nil
}

func (m *MockTenantManager) UpdateQuota(ctx context.Context, tenantID string, quotas multitenancy.TenantQuotas) error {
	return nil
}

func (m *MockTenantManager) GetTenantResources(ctx context.Context, tenantID string) (*multitenancy.TenantResources, error) {
	return &multitenancy.TenantResources{
		Backups: []multitenancy.TenantResource{},
		VMs: []multitenancy.TenantResource{},
		Storage: []multitenancy.TenantResource{},
		Jobs: []multitenancy.TenantResource{},
		Users: []multitenancy.TenantResource{},
		CustomResources: make(map[string][]multitenancy.TenantResource),
	}, nil
}

func (m *MockTenantManager) AssignResource(ctx context.Context, tenantID string, resourceType string, resourceID string) error {
	return nil
}

func (m *MockTenantManager) UnassignResource(ctx context.Context, tenantID string, resourceType string, resourceID string) error {
	return nil
}

func (m *MockTenantManager) ValidateTenantAccess(ctx context.Context, tenantID string) error {
	return nil
}

// TestRecoveryPlanManager tests recovery plan management functionality
func TestRecoveryPlanManager(t *testing.T) {
	mockTenantMgr := &MockTenantManager{}
	mockStorageMgr := *storage.NewEngine()
	manager := NewInMemoryRecoveryPlanManager(mockTenantMgr, mockStorageMgr)

	t.Run("CreateRecoveryPlan", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		request := &RecoveryPlanRequest{
			Name:        "enterprise-recovery-plan",
			TenantID:    "test-tenant",
			Description: "Comprehensive recovery plan for enterprise workloads",
			Type:        RecoveryPlanTypeDisaster,
			Priority:    RecoveryPriorityCritical,
			VMSequences: []VMSequence{
				{
					ID:          "seq-webapp",
					Name:        "Web Application Recovery",
					TenantID:    "test-tenant",
					Type:        VMSequenceTypeSequential,
					Enabled:     true,
					VMs: []VMRecovery{
						{
							ID:           "vm-web-1",
							Name:         "web-server-1",
							RecoveryType: VMRecoveryTypeFull,
							SourceVM: SourceVM{
								ID:         "vm-source-web-1",
								Name:       "web-server-source",
								Platform:   "vmware",
								Host:       "esxi-host-1",
								Datacenter: "dc-primary",
								Network:    []string{"web-network"},
								Config:     map[string]string{"cpu": "2", "memory": "4GB"},
							},
							TargetVM: TargetVM{
								ID:         "vm-target-web-1",
								Name:       "web-server-recovered",
								Platform:   "vmware",
								Host:       "esxi-host-2",
								Datacenter: "dc-secondary",
								Network:    []string{"recovery-network"},
								Config:     map[string]string{"cpu": "2", "memory": "4GB"},
								Resources: VMResources{
									CPU:        2,
									MemoryGB:   4,
									StorageGB:  100,
									NetworkMbps: 1000,
								},
							},
							Backup: BackupInfo{
								BackupID:    "backup-web-1-20240311",
								Type:        "full",
								Timestamp:   time.Now().Add(-2 * time.Hour),
								Size:        50000000000, // 50GB
								Checksum:    "sha256:abc123",
								Storage:     "s3://backups/web",
								Encryption:  true,
								Compression: true,
							},
							Configuration: VMConfig{
								CPU:      2,
								MemoryGB: 4,
								StorageGB: 100,
								Network: []NetworkConfig{
									{
										Subnets: []SubnetConfig{
											{
												ID:         "subnet-1",
												Name:       "recovery-subnet",
												CIDR:       "10.0.1.0/24",
												Gateway:    "10.0.1.1",
												DNSServers: []string{"10.0.1.10", "10.0.1.11"},
												VLAN:       100,
											},
										},
										Firewall: []FirewallRule{
											{
												Protocol: "tcp",
												Port:     80,
												Source:   "any",
												Dest:     "0.0.0.0/0",
												Action:   "allow",
											},
											{
												Protocol: "tcp",
												Port:     443,
												Source:   "any",
												Dest:     "0.0.0.0/0",
												Action:   "allow",
											},
										},
									},
								},
								Disks: []DiskConfig{
									{
										ID:          "disk-1",
										Name:        "system-disk",
										Type:        "ssd",
										SizeGB:      100,
										Format:      "vmdk",
										ThinProvision: true,
										StorageTier: "ssd-tier-1",
									},
								},
								NICs: []NICConfig{
									{
										ID:        "nic-1",
										Name:      "primary-nic",
										Type:      "vmxnet3",
										Network:   "recovery-network",
										IPAddress: "10.0.1.50",
									},
								},
							},
							Requirements: VMRequirements{
								MinCPU:         2,
								MinMemoryGB:    4,
								MinStorageGB:   100,
								NetworkBandwidth: 1000,
								StorageType:    "ssd",
								StorageTier:    "ssd-tier-1",
								NetworkType:    "vlan",
								NetworkTier:    "standard",
							},
							Validation: VMValidation{
								PreCheck: []ValidationRule{
									{
										Name:       "target_host_check",
										Type:       "host_connectivity",
										Target:     "esxi-host-2",
										Timeout:    30 * time.Second,
										RetryCount: 3,
										Required:   true,
									},
								},
								PostCheck: []ValidationRule{
									{
										Name:       "vm_power_check",
										Type:       "vm_status",
										Target:     "vm-target-web-1",
										Timeout:    60 * time.Second,
										RetryCount: 5,
										Required:   true,
									},
								},
								HealthCheck: HealthCheckConfig{
									Type:     "ping",
									Endpoint: "10.0.1.50",
									Interval: 30 * time.Second,
									Timeout:  10 * time.Second,
									Retries:  3,
								},
								PerformanceCheck: PerformanceCheck{
									Metrics: []string{"cpu_usage", "memory_usage", "disk_io"},
									Thresholds: map[string]float64{
										"cpu_usage":    80.0,
										"memory_usage": 85.0,
										"disk_io":      1000.0,
									},
									Duration: 5 * time.Minute,
									Interval: 30 * time.Second,
								},
							},
							PostActions: []PostAction{
								{
									Type:       "start_services",
									Target:     "web-services",
									Timeout:    2 * time.Minute,
									RetryCount: 3,
								},
								{
									Type:       "health_check",
									Target:     "application",
									Timeout:    1 * time.Minute,
									RetryCount: 5,
								},
							},
						},
					},
					RecoveryOrder: VMRecoveryOrder{
						Type:        OrderTypeSequential,
						Parallel:    false,
						MaxParallel: 1,
					},
					Configuration: VMRecoveryConfig{
						Timeout:    30 * time.Minute,
						RetryCount: 3,
						RetryDelay: 30 * time.Second,
						PreCheck:   true,
						PostCheck:  true,
						PowerOn:    true,
						Snapshot:   true,
						Validation: true,
						Monitoring: true,
					},
					NetworkConfig: NetworkConfig{
						Subnets: []SubnetConfig{
							{
								ID:         "recovery-subnet",
								Name:       "Recovery Network",
								CIDR:       "10.0.1.0/24",
								Gateway:    "10.0.1.1",
								DNSServers: []string{"10.0.1.10", "10.0.1.11"},
								VLAN:       100,
							},
						},
						Firewall: []FirewallRule{
							{
								Protocol: "tcp",
								Port:     80,
								Source:   "any",
								Dest:     "0.0.0.0/0",
								Action:   "allow",
							},
						},
						DNS: DNSConfig{
							Servers: []string{"10.0.1.10", "10.0.1.11"},
							Records: []DNSRecord{
								{
									Name:  "web-server",
									Type:  "A",
									Value: "10.0.1.50",
									TTL:   300,
								},
							},
							SearchPath: []string{"recovery.local"},
						},
					},
					StorageConfig: StorageConfig{
						Datastores: []DatastoreConfig{
							{
								ID:        "ds-recovery-1",
								Name:      "Recovery Datastore",
								Type:      "vmfs",
								Capacity:  10000000000000, // 10TB
								Used:      5000000000000,  // 5TB
								Available: 5000000000000,  // 5TB
								Tier:      "ssd-tier-1",
							},
						},
						StoragePolicy: StoragePolicy{
							Name:        "recovery-policy",
							Type:        "ssd",
							Tier:        "tier-1",
							Replication: 1,
						},
						ThinProvision: true,
						Encryption:    true,
						Compression:   true,
					},
					Validation: ValidationConfig{
						PreCheck: []ValidationRule{
							{
								Name:       "storage_check",
								Type:       "storage_capacity",
								Target:     "ds-recovery-1",
								Timeout:    30 * time.Second,
								RetryCount: 3,
								Required:   true,
							},
						},
						PostCheck: []ValidationRule{
							{
								Name:       "network_check",
								Type:       "network_connectivity",
								Target:     "recovery-network",
								Timeout:    60 * time.Second,
								RetryCount: 5,
								Required:   true,
							},
						},
					},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			RecoveryOrder: RecoveryOrder{
				Type:     OrderTypeSequential,
				Parallel: false,
				Groups: []RecoveryGroup{
					{
						ID:          "group-1",
						Name:        "Web Application Group",
						SequenceIDs: []string{"seq-webapp"},
						Priority:    1,
					},
				},
			},
			Dependencies: []Dependency{
				{
					SequenceID: "seq-webapp",
					DependsOn:  "",
					Type:       "none",
					Condition:  "always",
				},
			},
			Configuration: RecoveryConfig{
				Timeout:    120 * time.Minute,
				RetryCount: 3,
				RetryDelay: 60 * time.Second,
				PreCheck:   true,
				PostCheck:  true,
				Validation: true,
				Monitoring: true,
				AutoStart:  true,
				Rollback:   true,
			},
			Notifications: NotificationConfig{
				OnSuccess: []string{"admin@example.com"},
				OnFailure: []string{"admin@example.com", "ops@example.com"},
				OnError:   []string{"devops@example.com"},
				OnWarning: []string{"ops@example.com"},
			},
			Retention: RetentionPolicy{
				ExecutionDays: 30,
				ReportDays:    90,
				LogDays:       7,
				ArtifactsDays: 14,
				MaxExecutions: 100,
				MaxReports:    50,
				MaxLogs:       1000,
				MaxArtifacts:  500,
			},
			Enabled: true,
			Metadata: map[string]string{
				"environment": "production",
				"team":        "disaster-recovery",
				"category":    "critical",
			},
		}

		plan, err := manager.CreateRecoveryPlan(ctx, request)
		if err != nil {
			t.Fatalf("Failed to create recovery plan: %v", err)
		}

		if plan.ID == "" {
			t.Error("Plan ID should not be empty")
		}

		if plan.Name != request.Name {
			t.Errorf("Expected name %s, got %s", request.Name, plan.Name)
		}

		if plan.TenantID != request.TenantID {
			t.Errorf("Expected tenant ID %s, got %s", request.TenantID, plan.TenantID)
		}

		if plan.Type != request.Type {
			t.Errorf("Expected type %s, got %s", request.Type, plan.Type)
		}

		if !plan.Enabled {
			t.Error("Plan should be enabled")
		}

		if len(plan.VMSequences) != len(request.VMSequences) {
			t.Errorf("Expected %d VM sequences, got %d", len(request.VMSequences), len(plan.VMSequences))
		}
	})

	t.Run("GetRecoveryPlan", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// First create a plan
		request := &RecoveryPlanRequest{
			Name:     "get-test-plan",
			TenantID: "test-tenant",
			Type:     RecoveryPlanTypeMaintenance,
			Enabled:  false,
		}

		created, err := manager.CreateRecoveryPlan(ctx, request)
		if err != nil {
			t.Fatalf("Failed to create recovery plan: %v", err)
		}

		// Retrieve the plan
		retrieved, err := manager.GetRecoveryPlan(ctx, created.ID)
		if err != nil {
			t.Fatalf("Failed to get recovery plan: %v", err)
		}

		if retrieved.ID != created.ID {
			t.Errorf("Expected ID %s, got %s", created.ID, retrieved.ID)
		}

		if retrieved.Name != created.Name {
			t.Errorf("Expected name %s, got %s", created.Name, retrieved.Name)
		}
	})

	t.Run("ListRecoveryPlans", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create multiple plans
		for i := 0; i < 3; i++ {
			request := &RecoveryPlanRequest{
				Name:     fmt.Sprintf("list-test-plan-%d", i),
				TenantID: "test-tenant",
				Type:     RecoveryPlanTypeDisaster,
				Enabled:  i%2 == 0, // Enable every other plan
			}

			_, err := manager.CreateRecoveryPlan(ctx, request)
			if err != nil {
				t.Fatalf("Failed to create recovery plan %d: %v", i, err)
			}
		}

		// List all plans for tenant
		plans, err := manager.ListRecoveryPlans(ctx, &RecoveryPlanFilter{
			TenantID: "test-tenant",
		})
		if err != nil {
			t.Fatalf("Failed to list recovery plans: %v", err)
		}

		if len(plans) < 3 {
			t.Errorf("Expected at least 3 plans, got %d", len(plans))
		}

		// Filter by enabled status
		enabledPlans, err := manager.ListRecoveryPlans(ctx, &RecoveryPlanFilter{
			TenantID: "test-tenant",
			Enabled:  &[]bool{true}[0],
		})
		if err != nil {
			t.Fatalf("Failed to list enabled recovery plans: %v", err)
		}

		// Should have some enabled plans
		if len(enabledPlans) < 1 {
			t.Errorf("Expected at least 1 enabled plan, got %d", len(enabledPlans))
		}
	})

	t.Run("UpdateRecoveryPlan", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create a plan
		request := &RecoveryPlanRequest{
			Name:     "update-test-plan",
			TenantID: "test-tenant",
			Type:     RecoveryPlanTypeDisaster,
			Enabled:  false,
		}

		plan, err := manager.CreateRecoveryPlan(ctx, request)
		if err != nil {
			t.Fatalf("Failed to create recovery plan: %v", err)
		}

		// Update the plan
		newName := "updated-recovery-plan-name"
		newDescription := "Updated recovery plan description"
		enabled := true

		updateRequest := &UpdateRecoveryPlanRequest{
			Name:        &newName,
			Description: &newDescription,
			Enabled:     &enabled,
		}

		updated, err := manager.UpdateRecoveryPlan(ctx, plan.ID, updateRequest)
		if err != nil {
			t.Fatalf("Failed to update recovery plan: %v", err)
		}

		if updated.Name != newName {
			t.Errorf("Expected name %s, got %s", newName, updated.Name)
		}

		if updated.Description != newDescription {
			t.Errorf("Expected description %s, got %s", newDescription, updated.Description)
		}

		if !updated.Enabled {
			t.Error("Plan should be enabled after update")
		}
	})

	t.Run("EnableRecoveryPlan", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create a disabled plan
		request := &RecoveryPlanRequest{
			Name:     "enable-test-plan",
			TenantID: "test-tenant",
			Type:     RecoveryPlanTypeTesting,
			Enabled:  false,
		}

		plan, err := manager.CreateRecoveryPlan(ctx, request)
		if err != nil {
			t.Fatalf("Failed to create recovery plan: %v", err)
		}

		// Enable the plan
		err = manager.EnableRecoveryPlan(ctx, plan.ID)
		if err != nil {
			t.Fatalf("Failed to enable recovery plan: %v", err)
		}

		// Verify it's enabled
		enabled, err := manager.GetRecoveryPlan(ctx, plan.ID)
		if err != nil {
			t.Fatalf("Failed to get recovery plan: %v", err)
		}

		if !enabled.Enabled {
			t.Error("Plan should be enabled")
		}
	})

	t.Run("DisableRecoveryPlan", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create an enabled plan
		request := &RecoveryPlanRequest{
			Name:     "disable-test-plan",
			TenantID: "test-tenant",
			Type:     RecoveryPlanTypeMigration,
			Enabled:  true,
		}

		plan, err := manager.CreateRecoveryPlan(ctx, request)
		if err != nil {
			t.Fatalf("Failed to create recovery plan: %v", err)
		}

		// Disable the plan
		err = manager.DisableRecoveryPlan(ctx, plan.ID)
		if err != nil {
			t.Fatalf("Failed to disable recovery plan: %v", err)
		}

		// Verify it's disabled
		disabled, err := manager.GetRecoveryPlan(ctx, plan.ID)
		if err != nil {
			t.Fatalf("Failed to get recovery plan: %v", err)
		}

		if disabled.Enabled {
			t.Error("Plan should be disabled")
		}
	})

	t.Run("DeleteRecoveryPlan", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create a plan
		request := &RecoveryPlanRequest{
			Name:     "delete-test-plan",
			TenantID: "test-tenant",
			Type:     RecoveryPlanTypeHybrid,
		}

		plan, err := manager.CreateRecoveryPlan(ctx, request)
		if err != nil {
			t.Fatalf("Failed to create recovery plan: %v", err)
		}

		// Delete the plan
		err = manager.DeleteRecoveryPlan(ctx, plan.ID)
		if err != nil {
			t.Fatalf("Failed to delete recovery plan: %v", err)
		}

		// Verify it's gone
		_, err = manager.GetRecoveryPlan(ctx, plan.ID)
		if err == nil {
			t.Error("Expected error when getting deleted plan")
		}
	})

	t.Run("StartRecovery", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create a recovery plan
		planRequest := &RecoveryPlanRequest{
			Name:     "recovery-test-plan",
			TenantID: "test-tenant",
			Type:     RecoveryPlanTypeDisaster,
			Enabled:  true,
		}

		plan, err := manager.CreateRecoveryPlan(ctx, planRequest)
		if err != nil {
			t.Fatalf("Failed to create recovery plan: %v", err)
		}

		// Start recovery
		recoveryRequest := &RecoveryRequest{
			PlanID:        plan.ID,
			TenantID:      "test-tenant",
			TriggerType:   RecoveryTriggerManual,
			TriggerReason: "Manual test recovery",
			Scope:         RecoveryScopeFull,
			Options: RecoveryOptions{
				PreCheck:    true,
				PostCheck:   true,
				Validation:  true,
				Monitoring:  true,
				Timeout:     120 * time.Minute,
				Parallel:    true,
				MaxParallel: 5,
				DryRun:      false,
				Rollback:    true,
			},
			Metadata: map[string]string{
				"test": "true",
			},
		}

		execution, err := manager.StartRecovery(ctx, recoveryRequest)
		if err != nil {
			t.Fatalf("Failed to start recovery: %v", err)
		}

		if execution.ID == "" {
			t.Error("Execution ID should not be empty")
		}

		if execution.PlanID != plan.ID {
			t.Errorf("Expected plan ID %s, got %s", plan.ID, execution.PlanID)
		}

		if execution.TenantID != recoveryRequest.TenantID {
			t.Errorf("Expected tenant ID %s, got %s", recoveryRequest.TenantID, execution.TenantID)
		}

		if execution.TriggerType != recoveryRequest.TriggerType {
			t.Errorf("Expected trigger type %s, got %s", recoveryRequest.TriggerType, execution.TriggerType)
		}

		if execution.Status != ExecutionStatusPending {
			t.Errorf("Expected status %s, got %s", ExecutionStatusPending, execution.Status)
		}
	})

	t.Run("GetRecoveryExecution", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create a recovery plan
		planRequest := &RecoveryPlanRequest{
			Name:     "get-recovery-test-plan",
			TenantID: "test-tenant",
			Type:     RecoveryPlanTypeDisaster,
			Enabled:  true,
		}

		plan, err := manager.CreateRecoveryPlan(ctx, planRequest)
		if err != nil {
			t.Fatalf("Failed to create recovery plan: %v", err)
		}

		// Start recovery
		recoveryRequest := &RecoveryRequest{
			PlanID:        plan.ID,
			TenantID:      "test-tenant",
			TriggerType:   RecoveryTriggerManual,
			TriggerReason: "Test recovery",
			Scope:         RecoveryScopePartial,
		}

		created, err := manager.StartRecovery(ctx, recoveryRequest)
		if err != nil {
			t.Fatalf("Failed to start recovery: %v", err)
		}

		// Retrieve the execution
		retrieved, err := manager.GetRecoveryExecution(ctx, created.ID)
		if err != nil {
			t.Fatalf("Failed to get recovery execution: %v", err)
		}

		if retrieved.ID != created.ID {
			t.Errorf("Expected ID %s, got %s", created.ID, retrieved.ID)
		}

		if retrieved.PlanID != created.PlanID {
			t.Errorf("Expected plan ID %s, got %s", created.PlanID, retrieved.PlanID)
		}
	})

	t.Run("ListRecoveryExecutions", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create a recovery plan
		planRequest := &RecoveryPlanRequest{
			Name:     "list-recovery-test-plan",
			TenantID: "test-tenant",
			Type:     RecoveryPlanTypeDisaster,
			Enabled:  true,
		}

		plan, err := manager.CreateRecoveryPlan(ctx, planRequest)
		if err != nil {
			t.Fatalf("Failed to create recovery plan: %v", err)
		}

		// Create multiple recovery executions
		for i := 0; i < 3; i++ {
			recoveryRequest := &RecoveryRequest{
				PlanID:        plan.ID,
				TenantID:      "test-tenant",
				TriggerType:   RecoveryTriggerManual,
				TriggerReason: fmt.Sprintf("Test recovery %d", i),
				Scope:         RecoveryScopePartial,
			}

			_, err := manager.StartRecovery(ctx, recoveryRequest)
			if err != nil {
				t.Fatalf("Failed to start recovery %d: %v", i, err)
			}
		}

		// List all executions for tenant
		executions, err := manager.ListRecoveryExecutions(ctx, &ExecutionFilter{
			TenantID: "test-tenant",
		})
		if err != nil {
			t.Fatalf("Failed to list recovery executions: %v", err)
		}

		if len(executions) < 3 {
			t.Errorf("Expected at least 3 executions, got %d", len(executions))
		}

		// Filter by status
		pendingExecutions, err := manager.ListRecoveryExecutions(ctx, &ExecutionFilter{
			TenantID: "test-tenant",
			Status:   ExecutionStatusPending,
		})
		if err != nil {
			t.Fatalf("Failed to list pending executions: %v", err)
		}

		// Should have some pending executions
		if len(pendingExecutions) < 1 {
			t.Errorf("Expected at least 1 pending execution, got %d", len(pendingExecutions))
		}
	})

	t.Run("CancelRecoveryExecution", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create a recovery plan
		planRequest := &RecoveryPlanRequest{
			Name:     "cancel-recovery-test-plan",
			TenantID: "test-tenant",
			Type:     RecoveryPlanTypeDisaster,
			Enabled:  true,
		}

		plan, err := manager.CreateRecoveryPlan(ctx, planRequest)
		if err != nil {
			t.Fatalf("Failed to create recovery plan: %v", err)
		}

		// Start recovery
		recoveryRequest := &RecoveryRequest{
			PlanID:        plan.ID,
			TenantID:      "test-tenant",
			TriggerType:   RecoveryTriggerManual,
			TriggerReason: "Test recovery for cancellation",
			Scope:         RecoveryScopePartial,
		}

		execution, err := manager.StartRecovery(ctx, recoveryRequest)
		if err != nil {
			t.Fatalf("Failed to start recovery: %v", err)
		}

		// Cancel the execution
		err = manager.CancelRecoveryExecution(ctx, execution.ID)
		if err != nil {
			t.Fatalf("Failed to cancel recovery execution: %v", err)
		}

		// Verify it's cancelled
		cancelled, err := manager.GetRecoveryExecution(ctx, execution.ID)
		if err != nil {
			t.Fatalf("Failed to get recovery execution: %v", err)
		}

		if cancelled.Status != ExecutionStatusCancelled {
			t.Errorf("Expected status %s, got %s", ExecutionStatusCancelled, cancelled.Status)
		}
	})

	t.Run("GetRecoveryStats", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create some plans and executions
		for i := 0; i < 2; i++ {
			planRequest := &RecoveryPlanRequest{
				Name:     fmt.Sprintf("stats-test-plan-%d", i),
				TenantID: "test-tenant",
				Type:     RecoveryPlanTypeDisaster,
				Enabled:  true,
			}

			plan, err := manager.CreateRecoveryPlan(ctx, planRequest)
			if err != nil {
				t.Fatalf("Failed to create recovery plan %d: %v", i, err)
			}

			// Create a recovery execution
			recoveryRequest := &RecoveryRequest{
				PlanID:        plan.ID,
				TenantID:      "test-tenant",
				TriggerType:   RecoveryTriggerManual,
				TriggerReason: "Test recovery",
				Scope:         RecoveryScopePartial,
			}

			recovery, err := manager.StartRecovery(ctx, recoveryRequest)
			if err != nil {
				t.Fatalf("Failed to start recovery %d: %v", i, err)
			}

			// Add execution to stats (simulate completion)
			recovery.Status = ExecutionStatusCompleted
			recovery.Timing.ExecutionDuration = 10 * time.Minute
			manager.executions[recovery.ID] = recovery
		}

		// Get tenant stats
		timeRange := TimeRange{
			From: time.Now().Add(-2 * time.Hour),
			To:   time.Now(),
		}

		stats, err := manager.GetRecoveryStats(ctx, "test-tenant", timeRange)
		if err != nil {
			t.Fatalf("Failed to get recovery stats: %v", err)
		}

		if stats.TenantID != "test-tenant" {
			t.Errorf("Expected tenant ID %s, got %s", "test-tenant", stats.TenantID)
		}

		if stats.TotalPlans < 2 {
			t.Errorf("Expected at least 2 total plans, got %d", stats.TotalPlans)
		}

		if stats.ActivePlans < 2 {
			t.Errorf("Expected at least 2 active plans, got %d", stats.ActivePlans)
		}

		if stats.TotalExecutions < 2 {
			t.Errorf("Expected at least 2 total executions, got %d", stats.TotalExecutions)
		}
	})

	t.Run("GetRecoverySystemHealth", func(t *testing.T) {
		ctx := context.Background()

		// Get system health
		health, err := manager.GetRecoverySystemHealth(ctx)
		if err != nil {
			t.Fatalf("Failed to get recovery system health: %v", err)
		}

		if health.Status != HealthStatusHealthy {
			t.Errorf("Expected status %s, got %s", HealthStatusHealthy, health.Status)
		}

		if len(health.HostHealth) == 0 {
			t.Error("Expected at least one host health entry")
		}

		if len(health.StorageHealth) == 0 {
			t.Error("Expected at least one storage health entry")
		}

		if len(health.NetworkHealth) == 0 {
			t.Error("Expected at least one network health entry")
		}

		if health.ErrorRate < 0 || health.ErrorRate > 1 {
			t.Errorf("Expected error rate between 0 and 1, got %f", health.ErrorRate)
		}

		if health.ResponseTime <= 0 {
			t.Errorf("Expected positive response time, got %v", health.ResponseTime)
		}
	})

	t.Run("GetActiveRecoveries", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create a recovery plan
		planRequest := &RecoveryPlanRequest{
			Name:     "active-recoveries-test-plan",
			TenantID: "test-tenant",
			Type:     RecoveryPlanTypeDisaster,
			Enabled:  true,
		}

		plan, err := manager.CreateRecoveryPlan(ctx, planRequest)
		if err != nil {
			t.Fatalf("Failed to create recovery plan: %v", err)
		}

		// Start multiple recoveries
		recoveryRequest := &RecoveryRequest{
			PlanID:        plan.ID,
			TenantID:      "test-tenant",
			TriggerType:   RecoveryTriggerManual,
			TriggerReason: "Test recovery",
			Scope:         RecoveryScopePartial,
		}

		_, err = manager.StartRecovery(ctx, recoveryRequest)
		if err != nil {
			t.Fatalf("Failed to start recovery: %v", err)
		}

		// Get active recoveries
		activeRecoveries, err := manager.GetActiveRecoveries(ctx)
		if err != nil {
			t.Fatalf("Failed to get active recoveries: %v", err)
		}

		if len(activeRecoveries) < 1 {
			t.Errorf("Expected at least 1 active recovery, got %d", len(activeRecoveries))
		}

		// Verify recovery properties
		for _, recovery := range activeRecoveries {
			if recovery.PlanID != plan.ID {
				t.Errorf("Expected plan ID %s, got %s", plan.ID, recovery.PlanID)
			}
			if recovery.TenantID != "test-tenant" {
				t.Errorf("Expected tenant ID %s, got %s", "test-tenant", recovery.TenantID)
			}
			if recovery.Status != ExecutionStatusRunning && recovery.Status != ExecutionStatusPending {
				t.Errorf("Expected running or pending status, got %s", recovery.Status)
			}
		}
	})
}
