package drorchestration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"novabackup/internal/multitenancy"
	"novabackup/internal/storage"
)

// MockTenantManager for DR Orchestration testing
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
		Backups:         []multitenancy.TenantResource{},
		VMs:             []multitenancy.TenantResource{},
		Storage:         []multitenancy.TenantResource{},
		Jobs:            []multitenancy.TenantResource{},
		Users:           []multitenancy.TenantResource{},
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

// TestDROrchestrator tests DR orchestration functionality
func TestDROrchestrator(t *testing.T) {
	mockTenantMgr := &MockTenantManager{}
	mockStorageMgr := *storage.NewEngine()
	orchestrator := NewInMemoryDROrchestrator(mockTenantMgr, mockStorageMgr)

	t.Run("CreateDRPlan", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		request := &DRPlanRequest{
			Name:        "enterprise-dr-plan",
			TenantID:    "test-tenant",
			Description: "Comprehensive DR plan for enterprise workloads",
			Type:        DRPlanTypeSiteLevel,
			Priority:    DRPlanPriorityCritical,
			Sites: []DRSite{
				{
					ID:       "site-primary",
					Name:     "Primary Data Center",
					Type:     SiteTypePrimary,
					Location: "New York",
					Role:     SiteRoleActive,
					Status:   SiteStatusOnline,
					Capacity: SiteCapacity{
						MaxVMs:         100,
						MaxStorageGB:   10000,
						MaxCPU:         200,
						MaxMemoryGB:    400,
						MaxNetworkMbps: 10000,
					},
					Network: SiteNetwork{
						Subnets:       []string{"10.0.1.0/24", "10.0.2.0/24"},
						VLANs:         []int{100, 200},
						VPNEndpoints:  []string{"vpn.primary.com"},
						BandwidthMbps: 10000,
						LatencyMs:     5,
					},
					Storage: SiteStorage{
						Backends: []StorageBackend{
							{
								Type:     "s3",
								Config:   map[string]string{"bucket": "primary-backups"},
								Capacity: 10000000000000, // 10TB
								Used:     5000000000000,  // 5TB
							},
						},
						Replication: true,
						Encryption:  true,
						Compression: true,
						Tiering:     true,
					},
					Compute: SiteCompute{
						Hosts: []ComputeHost{
							{
								ID:        "host-1",
								Name:      "esxi-host-1",
								Type:      "vmware",
								CPU:       24,
								MemoryGB:  128,
								StorageGB: 1000,
								Status:    "online",
							},
						},
					},
					CreatedAt: time.Now(),
				},
				{
					ID:       "site-secondary",
					Name:     "Secondary Data Center",
					Type:     SiteTypeSecondary,
					Location: "Los Angeles",
					Role:     SiteRolePassive,
					Status:   SiteStatusOnline,
					Capacity: SiteCapacity{
						MaxVMs:         80,
						MaxStorageGB:   8000,
						MaxCPU:         160,
						MaxMemoryGB:    320,
						MaxNetworkMbps: 8000,
					},
					CreatedAt: time.Now(),
				},
			},
			Workloads: []DRWorkload{
				{
					ID:           "workload-webapp",
					Name:         "Web Application",
					Type:         WorkloadTypeApplication,
					Priority:     WorkloadPriorityHigh,
					Dependencies: []string{"workload-database"},
					Resources: WorkloadResources{
						CPU:         4,
						MemoryGB:    8,
						StorageGB:   100,
						NetworkMbps: 1000,
					},
					BackupPolicy: BackupPolicy{
						Schedule:    "0 2 * * *",
						Retention:   30 * 24 * time.Hour,
						Type:        "incremental",
						Compression: true,
						Encryption:  true,
						Replication: true,
					},
					Recovery: RecoveryPolicy{
						RTO:       15 * time.Minute,
						RPO:       5 * time.Minute,
						Priority:  WorkloadPriorityHigh,
						Order:     1,
						Parallel:  true,
						AutoStart: true,
					},
					Applications: []Application{
						{
							ID:      "app-web",
							Name:    "Web Server",
							Type:    "nginx",
							Version: "1.20",
							Config:  map[string]string{"workers": "4"},
							HealthCheck: HealthCheck{
								Type:     "http",
								Endpoint: "/health",
								Interval: 30 * time.Second,
								Timeout:  10 * time.Second,
								Retries:  3,
							},
						},
					},
					VMs: []VM{
						{
							ID:        "vm-web-1",
							Name:      "web-server-1",
							Platform:  "vmware",
							Host:      "esxi-host-1",
							Datastore: "datastore1",
							Network:   []string{"web-network"},
							Config:    map[string]string{"cpu": "2", "memory": "4GB"},
							Snapshots: []Snapshot{
								{
									ID:        "snap-1",
									Name:      "pre-failover",
									CreatedAt: time.Now().Add(-1 * time.Hour),
									Size:      10000000000, // 10GB
								},
							},
						},
					},
				},
			},
			RecoverySteps: []RecoveryStep{
				{
					ID:         "step-pre-check",
					Name:       "Pre-Flight Checks",
					Type:       StepTypePreCheck,
					Order:      1,
					Parallel:   false,
					Timeout:    5 * time.Minute,
					RetryCount: 3,
					Conditions: []Condition{
						{
							Type:     "site_status",
							Target:   "secondary",
							Operator: "equals",
							Value:    "online",
							Timeout:  2 * time.Minute,
						},
					},
					Actions: []Action{
						{
							Type:    "check_connectivity",
							Target:  "secondary-site",
							Timeout: 1 * time.Minute,
						},
					},
					Validations: []Validation{
						{
							Type:    "site_health",
							Target:  "secondary",
							Timeout: 2 * time.Minute,
						},
					},
				},
				{
					ID:         "step-site-failover",
					Name:       "Site Failover",
					Type:       StepTypeSiteFailover,
					Order:      2,
					Parallel:   false,
					Timeout:    30 * time.Minute,
					RetryCount: 2,
					Actions: []Action{
						{
							Type:    "activate_site",
							Target:  "secondary",
							Timeout: 15 * time.Minute,
						},
					},
				},
				{
					ID:         "step-vm-recovery",
					Name:       "VM Recovery",
					Type:       StepTypeVMRecovery,
					Order:      3,
					Parallel:   true,
					Timeout:    20 * time.Minute,
					RetryCount: 2,
					Actions: []Action{
						{
							Type:    "restore_vm",
							Target:  "vm-web-1",
							Timeout: 10 * time.Minute,
						},
					},
				},
				{
					ID:         "step-app-start",
					Name:       "Application Start",
					Type:       StepTypeAppStart,
					Order:      4,
					Parallel:   true,
					Timeout:    10 * time.Minute,
					RetryCount: 3,
					Actions: []Action{
						{
							Type:    "start_app",
							Target:  "app-web",
							Timeout: 5 * time.Minute,
						},
					},
				},
				{
					ID:         "step-post-check",
					Name:       "Post-Failover Validation",
					Type:       StepTypePostCheck,
					Order:      5,
					Parallel:   false,
					Timeout:    15 * time.Minute,
					RetryCount: 2,
					Validations: []Validation{
						{
							Type:    "app_health",
							Target:  "app-web",
							Timeout: 5 * time.Minute,
						},
					},
				},
			},
			FailoverConfig: FailoverConfig{
				Trigger: TriggerConfig{
					Type:    "manual",
					Enabled: true,
				},
				AutoFailover:  false,
				FailoverDelay: 0,
				DataSync:      true,
				PreCheck:      true,
				PostCheck:     true,
				Rollback:      true,
				Notification:  true,
			},
			FailbackConfig: FailbackConfig{
				Trigger: TriggerConfig{
					Type:    "manual",
					Enabled: true,
				},
				AutoFailback:  false,
				FailbackDelay: 1 * time.Hour,
				DataSync:      true,
				Validation:    true,
				PreCheck:      true,
				PostCheck:     true,
				Rollback:      true,
			},
			TestConfig: TestConfig{
				Schedule: TriggerConfig{
					Type:    "cron",
					Enabled: true,
					Config:  map[string]string{"expression": "0 3 * * 0"}, // Weekly on Sunday 3 AM
				},
				TestType:        TestTypePartial,
				Scope:           TestScopeWorkload,
				Isolation:       true,
				DataValidation:  true,
				PerformanceTest: false,
				Notification:    true,
			},
			Notifications: NotificationConfig{
				OnSuccess: []string{"admin@example.com"},
				OnFailure: []string{"admin@example.com", "ops@example.com"},
				OnError:   []string{"devops@example.com"},
			},
			Enabled: true,
			Metadata: map[string]string{
				"environment": "production",
				"team":        "disaster-recovery",
			},
		}

		plan, err := orchestrator.CreateDRPlan(ctx, request)
		if err != nil {
			t.Fatalf("Failed to create DR plan: %v", err)
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

		if len(plan.Sites) != len(request.Sites) {
			t.Errorf("Expected %d sites, got %d", len(request.Sites), len(plan.Sites))
		}

		if len(plan.RecoverySteps) != len(request.RecoverySteps) {
			t.Errorf("Expected %d recovery steps, got %d", len(request.RecoverySteps), len(plan.RecoverySteps))
		}
	})

	t.Run("GetDRPlan", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// First create a plan
		request := &DRPlanRequest{
			Name:     "get-test-plan",
			TenantID: "test-tenant",
			Type:     DRPlanTypeWorkloadLevel,
			Enabled:  false,
		}

		created, err := orchestrator.CreateDRPlan(ctx, request)
		if err != nil {
			t.Fatalf("Failed to create DR plan: %v", err)
		}

		// Retrieve the plan
		retrieved, err := orchestrator.GetDRPlan(ctx, created.ID)
		if err != nil {
			t.Fatalf("Failed to get DR plan: %v", err)
		}

		if retrieved.ID != created.ID {
			t.Errorf("Expected ID %s, got %s", created.ID, retrieved.ID)
		}

		if retrieved.Name != created.Name {
			t.Errorf("Expected name %s, got %s", created.Name, retrieved.Name)
		}
	})

	t.Run("ListDRPlans", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create multiple plans
		for i := 0; i < 3; i++ {
			request := &DRPlanRequest{
				Name:     fmt.Sprintf("list-test-plan-%d", i),
				TenantID: "test-tenant",
				Type:     DRPlanTypeSiteLevel,
				Enabled:  i%2 == 0, // Enable every other plan
			}

			_, err := orchestrator.CreateDRPlan(ctx, request)
			if err != nil {
				t.Fatalf("Failed to create DR plan %d: %v", i, err)
			}
		}

		// List all plans for tenant
		plans, err := orchestrator.ListDRPlans(ctx, &DRPlanFilter{
			TenantID: "test-tenant",
		})
		if err != nil {
			t.Fatalf("Failed to list DR plans: %v", err)
		}

		if len(plans) < 3 {
			t.Errorf("Expected at least 3 plans, got %d", len(plans))
		}

		// Filter by enabled status
		enabledPlans, err := orchestrator.ListDRPlans(ctx, &DRPlanFilter{
			TenantID: "test-tenant",
			Enabled:  &[]bool{true}[0],
		})
		if err != nil {
			t.Fatalf("Failed to list enabled DR plans: %v", err)
		}

		// Should have some enabled plans
		if len(enabledPlans) < 1 {
			t.Errorf("Expected at least 1 enabled plan, got %d", len(enabledPlans))
		}
	})

	t.Run("UpdateDRPlan", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create a plan
		request := &DRPlanRequest{
			Name:     "update-test-plan",
			TenantID: "test-tenant",
			Type:     DRPlanTypeSiteLevel,
			Enabled:  false,
		}

		plan, err := orchestrator.CreateDRPlan(ctx, request)
		if err != nil {
			t.Fatalf("Failed to create DR plan: %v", err)
		}

		// Update the plan
		newName := "updated-dr-plan-name"
		newDescription := "Updated DR plan description"
		enabled := true

		updateRequest := &UpdateDRPlanRequest{
			Name:        &newName,
			Description: &newDescription,
			Enabled:     &enabled,
		}

		updated, err := orchestrator.UpdateDRPlan(ctx, plan.ID, updateRequest)
		if err != nil {
			t.Fatalf("Failed to update DR plan: %v", err)
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

	t.Run("EnableDRPlan", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create a disabled plan
		request := &DRPlanRequest{
			Name:     "enable-test-plan",
			TenantID: "test-tenant",
			Type:     DRPlanTypeWorkloadLevel,
			Enabled:  false,
		}

		plan, err := orchestrator.CreateDRPlan(ctx, request)
		if err != nil {
			t.Fatalf("Failed to create DR plan: %v", err)
		}

		// Enable the plan
		err = orchestrator.EnableDRPlan(ctx, plan.ID)
		if err != nil {
			t.Fatalf("Failed to enable DR plan: %v", err)
		}

		// Verify it's enabled
		enabled, err := orchestrator.GetDRPlan(ctx, plan.ID)
		if err != nil {
			t.Fatalf("Failed to get DR plan: %v", err)
		}

		if !enabled.Enabled {
			t.Error("Plan should be enabled")
		}
	})

	t.Run("DisableDRPlan", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create an enabled plan
		request := &DRPlanRequest{
			Name:     "disable-test-plan",
			TenantID: "test-tenant",
			Type:     DRPlanTypeSiteLevel,
			Enabled:  true,
		}

		plan, err := orchestrator.CreateDRPlan(ctx, request)
		if err != nil {
			t.Fatalf("Failed to create DR plan: %v", err)
		}

		// Disable the plan
		err = orchestrator.DisableDRPlan(ctx, plan.ID)
		if err != nil {
			t.Fatalf("Failed to disable DR plan: %v", err)
		}

		// Verify it's disabled
		disabled, err := orchestrator.GetDRPlan(ctx, plan.ID)
		if err != nil {
			t.Fatalf("Failed to get DR plan: %v", err)
		}

		if disabled.Enabled {
			t.Error("Plan should be disabled")
		}
	})

	t.Run("DeleteDRPlan", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create a plan
		request := &DRPlanRequest{
			Name:     "delete-test-plan",
			TenantID: "test-tenant",
			Type:     DRPlanTypeApplication,
		}

		plan, err := orchestrator.CreateDRPlan(ctx, request)
		if err != nil {
			t.Fatalf("Failed to create DR plan: %v", err)
		}

		// Delete the plan
		err = orchestrator.DeleteDRPlan(ctx, plan.ID)
		if err != nil {
			t.Fatalf("Failed to delete DR plan: %v", err)
		}

		// Verify it's gone
		_, err = orchestrator.GetDRPlan(ctx, plan.ID)
		if err == nil {
			t.Error("Expected error when getting deleted plan")
		}
	})

	t.Run("StartFailover", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create a DR plan
		planRequest := &DRPlanRequest{
			Name:     "failover-test-plan",
			TenantID: "test-tenant",
			Type:     DRPlanTypeSiteLevel,
			Enabled:  true,
			RecoverySteps: []RecoveryStep{
				{
					ID:    "step-1",
					Name:  "Test Step",
					Type:  StepTypePreCheck,
					Order: 1,
				},
			},
		}

		plan, err := orchestrator.CreateDRPlan(ctx, planRequest)
		if err != nil {
			t.Fatalf("Failed to create DR plan: %v", err)
		}

		// Start failover
		failoverRequest := &FailoverRequest{
			PlanID:        plan.ID,
			TenantID:      "test-tenant",
			TriggerType:   FailoverTriggerManual,
			TriggerReason: "Manual test failover",
			Scope:         FailoverScopePartial,
			Options: FailoverOptions{
				DataSync:   true,
				PreCheck:   true,
				PostCheck:  true,
				Validation: true,
				Rollback:   true,
				Timeout:    30 * time.Minute,
				Parallel:   true,
				DryRun:     false,
			},
			Metadata: map[string]string{
				"test": "true",
			},
		}

		execution, err := orchestrator.StartFailover(ctx, failoverRequest)
		if err != nil {
			t.Fatalf("Failed to start failover: %v", err)
		}

		if execution.ID == "" {
			t.Error("Execution ID should not be empty")
		}

		if execution.PlanID != plan.ID {
			t.Errorf("Expected plan ID %s, got %s", plan.ID, execution.PlanID)
		}

		if execution.TenantID != failoverRequest.TenantID {
			t.Errorf("Expected tenant ID %s, got %s", failoverRequest.TenantID, execution.TenantID)
		}

		if execution.TriggerType != failoverRequest.TriggerType {
			t.Errorf("Expected trigger type %s, got %s", failoverRequest.TriggerType, execution.TriggerType)
		}

		if execution.Status != ExecutionStatusPending {
			t.Errorf("Expected status %s, got %s", ExecutionStatusPending, execution.Status)
		}
	})

	t.Run("GetFailoverExecution", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create a DR plan
		planRequest := &DRPlanRequest{
			Name:     "get-failover-test-plan",
			TenantID: "test-tenant",
			Type:     DRPlanTypeSiteLevel,
			Enabled:  true,
		}

		plan, err := orchestrator.CreateDRPlan(ctx, planRequest)
		if err != nil {
			t.Fatalf("Failed to create DR plan: %v", err)
		}

		// Start failover
		failoverRequest := &FailoverRequest{
			PlanID:        plan.ID,
			TenantID:      "test-tenant",
			TriggerType:   FailoverTriggerManual,
			TriggerReason: "Test failover",
			Scope:         FailoverScopeFull,
		}

		created, err := orchestrator.StartFailover(ctx, failoverRequest)
		if err != nil {
			t.Fatalf("Failed to start failover: %v", err)
		}

		// Retrieve the execution
		retrieved, err := orchestrator.GetFailoverExecution(ctx, created.ID)
		if err != nil {
			t.Fatalf("Failed to get failover execution: %v", err)
		}

		if retrieved.ID != created.ID {
			t.Errorf("Expected ID %s, got %s", created.ID, retrieved.ID)
		}

		if retrieved.PlanID != created.PlanID {
			t.Errorf("Expected plan ID %s, got %s", created.PlanID, retrieved.PlanID)
		}
	})

	t.Run("ListFailoverExecutions", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create a DR plan
		planRequest := &DRPlanRequest{
			Name:     "list-failover-test-plan",
			TenantID: "test-tenant",
			Type:     DRPlanTypeSiteLevel,
			Enabled:  true,
		}

		plan, err := orchestrator.CreateDRPlan(ctx, planRequest)
		if err != nil {
			t.Fatalf("Failed to create DR plan: %v", err)
		}

		// Create multiple failover executions
		for i := 0; i < 3; i++ {
			failoverRequest := &FailoverRequest{
				PlanID:        plan.ID,
				TenantID:      "test-tenant",
				TriggerType:   FailoverTriggerManual,
				TriggerReason: fmt.Sprintf("Test failover %d", i),
				Scope:         FailoverScopePartial,
			}

			_, err := orchestrator.StartFailover(ctx, failoverRequest)
			if err != nil {
				t.Fatalf("Failed to start failover %d: %v", i, err)
			}
		}

		// List all executions for tenant
		executions, err := orchestrator.ListFailoverExecutions(ctx, &ExecutionFilter{
			TenantID: "test-tenant",
		})
		if err != nil {
			t.Fatalf("Failed to list failover executions: %v", err)
		}

		if len(executions) < 3 {
			t.Errorf("Expected at least 3 executions, got %d", len(executions))
		}

		// Filter by status
		pendingExecutions, err := orchestrator.ListFailoverExecutions(ctx, &ExecutionFilter{
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

	t.Run("CancelFailoverExecution", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create a DR plan
		planRequest := &DRPlanRequest{
			Name:     "cancel-failover-test-plan",
			TenantID: "test-tenant",
			Type:     DRPlanTypeSiteLevel,
			Enabled:  true,
		}

		plan, err := orchestrator.CreateDRPlan(ctx, planRequest)
		if err != nil {
			t.Fatalf("Failed to create DR plan: %v", err)
		}

		// Start failover
		failoverRequest := &FailoverRequest{
			PlanID:        plan.ID,
			TenantID:      "test-tenant",
			TriggerType:   FailoverTriggerManual,
			TriggerReason: "Test failover for cancellation",
			Scope:         FailoverScopePartial,
		}

		execution, err := orchestrator.StartFailover(ctx, failoverRequest)
		if err != nil {
			t.Fatalf("Failed to start failover: %v", err)
		}

		// Cancel the execution
		err = orchestrator.CancelFailoverExecution(ctx, execution.ID)
		if err != nil {
			t.Fatalf("Failed to cancel failover execution: %v", err)
		}

		// Verify it's cancelled
		cancelled, err := orchestrator.GetFailoverExecution(ctx, execution.ID)
		if err != nil {
			t.Fatalf("Failed to get failover execution: %v", err)
		}

		if cancelled.Status != ExecutionStatusCancelled {
			t.Errorf("Expected status %s, got %s", ExecutionStatusCancelled, cancelled.Status)
		}
	})

	t.Run("StartFailback", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create a DR plan
		planRequest := &DRPlanRequest{
			Name:     "failback-test-plan",
			TenantID: "test-tenant",
			Type:     DRPlanTypeSiteLevel,
			Enabled:  true,
		}

		plan, err := orchestrator.CreateDRPlan(ctx, planRequest)
		if err != nil {
			t.Fatalf("Failed to create DR plan: %v", err)
		}

		// First start a failover
		failoverRequest := &FailoverRequest{
			PlanID:        plan.ID,
			TenantID:      "test-tenant",
			TriggerType:   FailoverTriggerManual,
			TriggerReason: "Test failover",
			Scope:         FailoverScopeFull,
		}

		failover, err := orchestrator.StartFailover(ctx, failoverRequest)
		if err != nil {
			t.Fatalf("Failed to start failover: %v", err)
		}

		// Wait a moment for failover to start
		time.Sleep(100 * time.Millisecond)

		// Start failback
		failbackRequest := &FailbackRequest{
			PlanID:        plan.ID,
			TenantID:      "test-tenant",
			ExecutionID:   failover.ID,
			TriggerType:   FailbackTriggerManual,
			TriggerReason: "Manual test failback",
			Scope:         FailbackScopeFull,
			Options: FailbackOptions{
				DataSync:   true,
				Validation: true,
				PreCheck:   true,
				PostCheck:  true,
				Timeout:    30 * time.Minute,
				Parallel:   true,
				DryRun:     false,
			},
		}

		execution, err := orchestrator.StartFailback(ctx, failbackRequest)
		if err != nil {
			t.Fatalf("Failed to start failback: %v", err)
		}

		if execution.ID == "" {
			t.Error("Execution ID should not be empty")
		}

		if execution.PlanID != plan.ID {
			t.Errorf("Expected plan ID %s, got %s", plan.ID, execution.PlanID)
		}

		if execution.ExecutionID != failover.ID {
			t.Errorf("Expected execution ID %s, got %s", failover.ID, execution.ExecutionID)
		}

		if execution.TriggerType != failbackRequest.TriggerType {
			t.Errorf("Expected trigger type %s, got %s", failbackRequest.TriggerType, execution.TriggerType)
		}
	})

	t.Run("RunDRTest", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create a DR plan
		planRequest := &DRPlanRequest{
			Name:     "test-dr-test-plan",
			TenantID: "test-tenant",
			Type:     DRPlanTypeSiteLevel,
			Enabled:  true,
		}

		plan, err := orchestrator.CreateDRPlan(ctx, planRequest)
		if err != nil {
			t.Fatalf("Failed to create DR plan: %v", err)
		}

		// Run DR test
		testRequest := &DRTestRequest{
			PlanID:   plan.ID,
			TenantID: "test-tenant",
			TestType: TestTypeSimulation,
			Scope:    TestScopeWorkload,
			Options: TestOptions{
				Isolation:       true,
				DataValidation:  true,
				PerformanceTest: false,
				Parallel:        true,
				Timeout:         15 * time.Minute,
				DryRun:          true,
			},
			Metadata: map[string]string{
				"test": "true",
			},
		}

		execution, err := orchestrator.RunDRTest(ctx, testRequest)
		if err != nil {
			t.Fatalf("Failed to run DR test: %v", err)
		}

		if execution.ID == "" {
			t.Error("Execution ID should not be empty")
		}

		if execution.PlanID != plan.ID {
			t.Errorf("Expected plan ID %s, got %s", plan.ID, execution.PlanID)
		}

		if execution.TestType != testRequest.TestType {
			t.Errorf("Expected test type %s, got %s", testRequest.TestType, execution.TestType)
		}

		if execution.Status != ExecutionStatusPending {
			t.Errorf("Expected status %s, got %s", ExecutionStatusPending, execution.Status)
		}
	})

	t.Run("GetDRStats", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create some plans and executions
		for i := 0; i < 2; i++ {
			planRequest := &DRPlanRequest{
				Name:     fmt.Sprintf("stats-test-plan-%d", i),
				TenantID: "test-tenant",
				Type:     DRPlanTypeSiteLevel,
				Enabled:  true,
			}

			plan, err := orchestrator.CreateDRPlan(ctx, planRequest)
			if err != nil {
				t.Fatalf("Failed to create DR plan %d: %v", i, err)
			}

			// Create a failover execution
			failoverRequest := &FailoverRequest{
				PlanID:        plan.ID,
				TenantID:      "test-tenant",
				TriggerType:   FailoverTriggerManual,
				TriggerReason: "Test failover",
				Scope:         FailoverScopePartial,
			}

			failover, err := orchestrator.StartFailover(ctx, failoverRequest)
			if err != nil {
				t.Fatalf("Failed to start failover %d: %v", i, err)
			}

			// Create a test execution
			testRequest := &DRTestRequest{
				PlanID:   plan.ID,
				TenantID: "test-tenant",
				TestType: TestTypeSimulation,
				Scope:    TestScopeWorkload,
			}

			_, err = orchestrator.RunDRTest(ctx, testRequest)
			if err != nil {
				t.Fatalf("Failed to run DR test %d: %v", i, err)
			}

			// Add execution to stats (simulate completion)
			failover.Status = ExecutionStatusCompleted
			failover.Timing.ExecutionDuration = 5 * time.Minute
			orchestrator.failovers[failover.ID] = failover
		}

		// Get tenant stats
		timeRange := TimeRange{
			From: time.Now().Add(-2 * time.Hour),
			To:   time.Now(),
		}

		stats, err := orchestrator.GetDRStats(ctx, "test-tenant", timeRange)
		if err != nil {
			t.Fatalf("Failed to get DR stats: %v", err)
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

		if stats.TotalFailovers < 2 {
			t.Errorf("Expected at least 2 total failovers, got %d", stats.TotalFailovers)
		}
	})

	t.Run("GetDRSystemHealth", func(t *testing.T) {
		ctx := context.Background()

		// Get system health
		health, err := orchestrator.GetDRSystemHealth(ctx)
		if err != nil {
			t.Fatalf("Failed to get DR system health: %v", err)
		}

		if health.Status != HealthStatusHealthy {
			t.Errorf("Expected status %s, got %s", HealthStatusHealthy, health.Status)
		}

		if len(health.SiteHealth) == 0 {
			t.Error("Expected at least one site health entry")
		}

		if len(health.PlanHealth) == 0 {
			t.Error("Expected at least one plan health entry")
		}

		if health.ErrorRate < 0 || health.ErrorRate > 1 {
			t.Errorf("Expected error rate between 0 and 1, got %f", health.ErrorRate)
		}

		if health.ResponseTime <= 0 {
			t.Errorf("Expected positive response time, got %v", health.ResponseTime)
		}
	})

	t.Run("GetActiveExecutions", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create a DR plan
		planRequest := &DRPlanRequest{
			Name:     "active-executions-test-plan",
			TenantID: "test-tenant",
			Type:     DRPlanTypeSiteLevel,
			Enabled:  true,
		}

		plan, err := orchestrator.CreateDRPlan(ctx, planRequest)
		if err != nil {
			t.Fatalf("Failed to create DR plan: %v", err)
		}

		// Start multiple executions
		failoverRequest := &FailoverRequest{
			PlanID:        plan.ID,
			TenantID:      "test-tenant",
			TriggerType:   FailoverTriggerManual,
			TriggerReason: "Test failover",
			Scope:         FailoverScopePartial,
		}

		_, err = orchestrator.StartFailover(ctx, failoverRequest)
		if err != nil {
			t.Fatalf("Failed to start failover: %v", err)
		}

		testRequest := &DRTestRequest{
			PlanID:   plan.ID,
			TenantID: "test-tenant",
			TestType: TestTypeSimulation,
			Scope:    TestScopeWorkload,
		}

		_, err = orchestrator.RunDRTest(ctx, testRequest)
		if err != nil {
			t.Fatalf("Failed to run DR test: %v", err)
		}

		// Get active executions
		activeExecutions, err := orchestrator.GetActiveExecutions(ctx)
		if err != nil {
			t.Fatalf("Failed to get active executions: %v", err)
		}

		if len(activeExecutions) < 2 {
			t.Errorf("Expected at least 2 active executions, got %d", len(activeExecutions))
		}

		// Verify execution types
		hasFailover := false
		hasTest := false
		for _, exec := range activeExecutions {
			if exec.Type == ExecutionTypeFailover {
				hasFailover = true
			}
			if exec.Type == ExecutionTypeTest {
				hasTest = true
			}
		}

		if !hasFailover {
			t.Error("Expected at least one failover execution")
		}

		if !hasTest {
			t.Error("Expected at least one test execution")
		}
	})
}
