package scaleout

import (
	"context"
	"fmt"
	"testing"
	"time"

	"novabackup/internal/multitenancy"
	"novabackup/internal/storage"
)

// MockTenantManager for Scale-Out testing
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

// TestScaleOutManager tests scale-out management functionality
func TestScaleOutManager(t *testing.T) {
	mockTenantMgr := &MockTenantManager{}
	mockStorageMgr := *storage.NewEngine()
	manager := NewInMemoryScaleOutManager(mockTenantMgr, mockStorageMgr)

	t.Run("CreateRepository", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		request := &RepositoryRequest{
			Name:        "enterprise-scaleout-repo",
			TenantID:    "test-tenant",
			Description: "Comprehensive scale-out repository for enterprise workloads",
			Type:        RepositoryTypeScaleOut,
			Pools: []StoragePool{
				{
					ID:       "pool-ssd-1",
					Name:     "SSD Performance Pool",
					TenantID: "test-tenant",
					Type:     StoragePoolTypeSSD,
					Enabled:  true,
					Nodes: []StorageNode{
						{
							ID:   "node-ssd-1",
							Name: "ssd-node-1",
							Type: StorageNodeTypePrimary,
							Host: "ssd-host-1",
							Port: 8080,
							Configuration: NodeConfig{
								MaxConnections: 1000,
								Timeout:        30 * time.Second,
								RetryCount:     3,
								RetryDelay:     5 * time.Second,
								KeepAlive:      true,
								Compression:    true,
								Encryption:     true,
							},
							Capacity: NodeCapacity{
								TotalStorage:      1000000000000, // 1TB
								UsedStorage:       100000000000,  // 100GB
								AvailableStorage:  900000000000,  // 900GB
								MaxIOPS:           100000,
								CurrentIOPS:       5000,
								MaxThroughput:     1000000000, // 1GB/s
								CurrentThroughput: 50000000,   // 50MB/s
							},
							Performance: NodePerformance{
								Latency:      5 * time.Millisecond,
								Throughput:   50000000, // 50MB/s
								IOPS:         5000,
								CPUUsage:     25.5,
								MemoryUsage:  45.2,
								NetworkUsage: 15.8,
								DiskUsage:    30.1,
								ResponseTime: 2 * time.Millisecond,
							},
							HealthStatus: NodeHealthStatus{
								Status:     HealthStatusHealthy,
								LastCheck:  time.Now(),
								ErrorCount: 0,
								Uptime:     time.Hour * 24 * 7, // 1 week
								Version:    "1.0.0",
							},
							NetworkConfig: NetworkConfig{
								Subnets: []SubnetConfig{
									{
										ID:         "subnet-1",
										Name:       "storage-subnet",
										CIDR:       "10.0.1.0/24",
										Gateway:    "10.0.1.1",
										DNSServers: []string{"10.0.1.10", "10.0.1.11"},
										VLAN:       100,
									},
								},
								Firewall: []FirewallRule{
									{
										Protocol: "tcp",
										Port:     8080,
										Source:   "any",
										Dest:     "0.0.0.0/0",
										Action:   "allow",
									},
								},
								DNS: DNSConfig{
									Servers: []string{"10.0.1.10", "10.0.1.11"},
									Records: []DNSRecord{
										{
											Name:  "storage-node",
											Type:  "A",
											Value: "10.0.1.50",
											TTL:   300,
										},
									},
								},
							},
							StorageConfig: StorageConfig{
								Datastores: []DatastoreConfig{
									{
										ID:        "ds-ssd-1",
										Name:      "SSD Datastore",
										Type:      "ssd",
										Capacity:  1000000000000, // 1TB
										Used:      100000000000,  // 100GB
										Available: 900000000000,  // 900GB
										Tier:      "performance",
									},
								},
								StoragePolicy: StoragePolicy{
									Name:        "ssd-performance",
									Type:        "ssd",
									Tier:        "performance",
									Replication: 1,
								},
								ThinProvision: true,
								Encryption:    true,
								Compression:   true,
							},
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
							LastSeen:  &[]time.Time{time.Now()}[0],
						},
					},
					Configuration: PoolConfig{
						Replication:   2,
						Consistency:   ConsistencyTypeStrong,
						Quorum:        2,
						Timeout:       30 * time.Second,
						RetryCount:    3,
						RetryDelay:    5 * time.Second,
						HealthCheck:   true,
						AutoRebalance: true,
						LoadBalancing: true,
					},
					LoadBalancing: LoadBalancingConfig{
						Algorithm: AlgorithmWeighted,
						Weights: map[string]float64{
							"node-ssd-1": 1.0,
						},
						HealthCheck: HealthCheckConfig{
							Type:     "tcp",
							Endpoint: "10.0.1.50:8080",
							Interval: 30 * time.Second,
							Timeout:  10 * time.Second,
							Retries:  3,
						},
						Failover: FailoverConfig{
							Enabled:    true,
							Threshold:  2,
							Timeout:    30 * time.Second,
							RetryCount: 3,
							RetryDelay: 5 * time.Second,
						},
					},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			Configuration: RepositoryConfig{
				Capacity:      10000000000000, // 10TB
				Replication:   2,
				Consistency:   ConsistencyTypeStrong,
				Quorum:        2,
				Timeout:       60 * time.Second,
				RetryCount:    3,
				RetryDelay:    10 * time.Second,
				HealthCheck:   true,
				AutoRebalance: true,
				RebalancePolicy: RebalancePolicy{
					Enabled:        true,
					Threshold:      0.8,
					Interval:       time.Hour,
					MaxConcurrency: 5,
					Priority:       RebalancePriorityNormal,
				},
			},
			LoadBalancing: LoadBalancingConfig{
				Algorithm: AlgorithmLeastConnections,
				Weights: map[string]float64{
					"pool-ssd-1": 1.0,
				},
				HealthCheck: HealthCheckConfig{
					Type:     "tcp",
					Endpoint: "10.0.1.50:8080",
					Interval: 30 * time.Second,
					Timeout:  10 * time.Second,
					Retries:  3,
				},
				Failover: FailoverConfig{
					Enabled:    true,
					Threshold:  1,
					Timeout:    30 * time.Second,
					RetryCount: 3,
					RetryDelay: 5 * time.Second,
				},
			},
			DataMovers: []DataMover{
				{
					ID:       "mover-backup-1",
					Name:     "Backup Data Mover",
					TenantID: "test-tenant",
					Type:     DataMoverTypeBackup,
					Enabled:  true,
					Nodes: []ProcessingNode{
						{
							ID:   "proc-node-1",
							Name: "processing-node-1",
							Type: ProcessingNodeTypePrimary,
							Host: "proc-host-1",
							Port: 9090,
							Configuration: NodeConfig{
								MaxConnections: 500,
								Timeout:        60 * time.Second,
								RetryCount:     5,
								RetryDelay:     10 * time.Second,
								KeepAlive:      true,
								Compression:    true,
								Encryption:     true,
							},
							Capacity: NodeCapacity{
								TotalStorage:      2000000000000, // 2TB
								UsedStorage:       200000000000,  // 200GB
								AvailableStorage:  1800000000000, // 1.8TB
								MaxIOPS:           50000,
								CurrentIOPS:       10000,
								MaxThroughput:     2000000000, // 2GB/s
								CurrentThroughput: 100000000,  // 100MB/s
							},
							Performance: NodePerformance{
								Latency:      10 * time.Millisecond,
								Throughput:   100000000, // 100MB/s
								IOPS:         10000,
								CPUUsage:     35.5,
								MemoryUsage:  55.2,
								NetworkUsage: 25.8,
								DiskUsage:    40.1,
								ResponseTime: 5 * time.Millisecond,
							},
							HealthStatus: NodeHealthStatus{
								Status:     HealthStatusHealthy,
								LastCheck:  time.Now(),
								ErrorCount: 0,
								Uptime:     time.Hour * 24 * 14, // 2 weeks
								Version:    "2.0.0",
							},
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
							LastSeen:  &[]time.Time{time.Now()}[0],
						},
					},
					Configuration: MoverConfig{
						MaxConcurrency: 10,
						ChunkSize:      1048576,  // 1MB
						BufferSize:     10485760, // 10MB
						Timeout:        300 * time.Second,
						RetryCount:     5,
						RetryDelay:     10 * time.Second,
						Compression:    true,
						Encryption:     true,
						BandwidthLimit: 1000000000, // 1GB/s
					},
					LoadBalancing: LoadBalancingConfig{
						Algorithm: AlgorithmRoundRobin,
						Weights: map[string]float64{
							"proc-node-1": 1.0,
						},
					},
					Scheduling: SchedulingConfig{
						Algorithm:    SchedulingAlgorithmPriority,
						Priority:     SchedulingPriorityHigh,
						QueueSize:    1000,
						WorkerCount:  5,
						BatchSize:    10,
						BatchTimeout: 60 * time.Second,
					},
					Monitoring: MonitoringConfig{
						Enabled:   true,
						Metrics:   []string{"throughput", "latency", "error_rate", "queue_size"},
						Interval:  30 * time.Second,
						Retention: 7 * 24 * time.Hour,
						Alerting: AlertingConfig{
							Enabled: true,
							Thresholds: map[string]float64{
								"throughput": 100000000,   // 100MB/s
								"latency":    10000000000, // 10s
								"error_rate": 0.05,
								"queue_size": 800,
							},
							Channels: []string{"email", "slack"},
							Severity: "warning",
						},
					},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					Status:    MoverStatusStopped,
				},
			},
			Tiers: []StorageTier{
				{
					ID:       "tier-performance",
					Name:     "Performance Tier",
					TenantID: "test-tenant",
					Type:     StorageTierTypePerformance,
					Enabled:  true,
					Performance: TierPerformance{
						Latency:      5 * time.Millisecond,
						Throughput:   200000000, // 200MB/s
						IOPS:         20000,
						Availability: 99.9,
						Reliability:  99.95,
					},
					Cost: TierCost{
						StorageCost:   0.10,
						IOCost:        0.01,
						BandwidthCost: 0.05,
						TransferCost:  0.02,
						Currency:      "USD",
					},
					Policy: TierPolicy{
						DataRetention: 30 * 24 * time.Hour,
						AccessPattern: "hot",
						Compression:   true,
						Encryption:    true,
						Replication:   true,
						AutoMigration: true,
					},
					Configuration: TierConfig{
						MinSize:   1000000000,    // 1GB
						MaxSize:   1000000000000, // 1TB
						Quota:     100000000000,  // 100GB
						Threshold: 0.8,
						MigrationPolicy: MigrationPolicy{
							Enabled:   true,
							Direction: MigrationDirectionDown,
							Trigger:   MigrationTriggerThreshold,
							Schedule:  "0 2 * * *",
							Conditions: []MigrationCondition{
								{
									Type:      "age",
									Target:    "data",
									Operator:  "greater_than",
									Value:     "30d",
									Threshold: 0.8,
								},
							},
						},
					},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			Replication: ReplicationConfig{
				Type:        ReplicationTypeSynchronous,
				Factor:      2,
				SyncType:    SyncTypeFull,
				Consistency: ConsistencyTypeStrong,
				Compression: true,
				Encryption:  true,
				Bandwidth:   1000000000, // 1GB/s
			},
			Compression: CompressionConfig{
				Enabled:    true,
				Algorithm:  "lz4",
				Level:      4,
				BlockSize:  65536, // 64KB
				Window:     32768, // 32KB
				Threshold:  0.1,
				Exclusions: []string{"*.zip", "*.mp4", "*.avi"},
			},
			Encryption: EncryptionConfig{
				Enabled:     true,
				Algorithm:   "aes-256",
				KeySize:     256,
				Mode:        "gcm",
				IVSize:      96,
				KeyRotation: true,
			},
			Enabled: true,
			Metadata: map[string]string{
				"environment": "production",
				"team":        "storage",
				"category":    "scale-out",
			},
		}

		repo, err := manager.CreateRepository(ctx, request)
		if err != nil {
			t.Fatalf("Failed to create repository: %v", err)
		}

		if repo.ID == "" {
			t.Error("Repository ID should not be empty")
		}

		if repo.Name != request.Name {
			t.Errorf("Expected name %s, got %s", request.Name, repo.Name)
		}

		if repo.TenantID != request.TenantID {
			t.Errorf("Expected tenant ID %s, got %s", request.TenantID, repo.TenantID)
		}

		if repo.Type != request.Type {
			t.Errorf("Expected type %s, got %s", request.Type, repo.Type)
		}

		if !repo.Enabled {
			t.Error("Repository should be enabled")
		}

		if len(repo.Pools) != len(request.Pools) {
			t.Errorf("Expected %d pools, got %d", len(request.Pools), len(repo.Pools))
		}
	})

	t.Run("GetRepository", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// First create a repository
		request := &RepositoryRequest{
			Name:     "get-test-repo",
			TenantID: "test-tenant",
			Type:     RepositoryTypeDistributed,
			Enabled:  false,
		}

		created, err := manager.CreateRepository(ctx, request)
		if err != nil {
			t.Fatalf("Failed to create repository: %v", err)
		}

		// Retrieve the repository
		retrieved, err := manager.GetRepository(ctx, created.ID)
		if err != nil {
			t.Fatalf("Failed to get repository: %v", err)
		}

		if retrieved.ID != created.ID {
			t.Errorf("Expected ID %s, got %s", created.ID, retrieved.ID)
		}

		if retrieved.Name != created.Name {
			t.Errorf("Expected name %s, got %s", created.Name, retrieved.Name)
		}
	})

	t.Run("ListRepositories", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create multiple repositories
		for i := 0; i < 3; i++ {
			request := &RepositoryRequest{
				Name:     fmt.Sprintf("list-test-repo-%d", i),
				TenantID: "test-tenant",
				Type:     RepositoryTypeScaleOut,
				Enabled:  i%2 == 0, // Enable every other repository
			}

			_, err := manager.CreateRepository(ctx, request)
			if err != nil {
				t.Fatalf("Failed to create repository %d: %v", i, err)
			}
		}

		// List all repositories for tenant
		repos, err := manager.ListRepositories(ctx, &RepositoryFilter{
			TenantID: "test-tenant",
		})
		if err != nil {
			t.Fatalf("Failed to list repositories: %v", err)
		}

		if len(repos) < 3 {
			t.Errorf("Expected at least 3 repositories, got %d", len(repos))
		}

		// Filter by enabled status
		enabledRepos, err := manager.ListRepositories(ctx, &RepositoryFilter{
			TenantID: "test-tenant",
			Enabled:  &[]bool{true}[0],
		})
		if err != nil {
			t.Fatalf("Failed to list enabled repositories: %v", err)
		}

		// Should have some enabled repositories
		if len(enabledRepos) < 1 {
			t.Errorf("Expected at least 1 enabled repository, got %d", len(enabledRepos))
		}
	})

	t.Run("UpdateRepository", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create a repository
		request := &RepositoryRequest{
			Name:     "update-test-repo",
			TenantID: "test-tenant",
			Type:     RepositoryTypeHybrid,
			Enabled:  false,
		}

		repo, err := manager.CreateRepository(ctx, request)
		if err != nil {
			t.Fatalf("Failed to create repository: %v", err)
		}

		// Update the repository
		newName := "updated-repository-name"
		newDescription := "Updated repository description"
		enabled := true

		updateRequest := &UpdateRepositoryRequest{
			Name:        &newName,
			Description: &newDescription,
			Enabled:     &enabled,
		}

		updated, err := manager.UpdateRepository(ctx, repo.ID, updateRequest)
		if err != nil {
			t.Fatalf("Failed to update repository: %v", err)
		}

		if updated.Name != newName {
			t.Errorf("Expected name %s, got %s", newName, updated.Name)
		}

		if updated.Description != newDescription {
			t.Errorf("Expected description %s, got %s", newDescription, updated.Description)
		}

		if !updated.Enabled {
			t.Error("Repository should be enabled after update")
		}
	})

	t.Run("EnableRepository", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create a disabled repository
		request := &RepositoryRequest{
			Name:     "enable-test-repo",
			TenantID: "test-tenant",
			Type:     RepositoryTypeCloud,
			Enabled:  false,
		}

		repo, err := manager.CreateRepository(ctx, request)
		if err != nil {
			t.Fatalf("Failed to create repository: %v", err)
		}

		// Enable the repository
		err = manager.EnableRepository(ctx, repo.ID)
		if err != nil {
			t.Fatalf("Failed to enable repository: %v", err)
		}

		// Verify it's enabled
		enabled, err := manager.GetRepository(ctx, repo.ID)
		if err != nil {
			t.Fatalf("Failed to get repository: %v", err)
		}

		if !enabled.Enabled {
			t.Error("Repository should be enabled")
		}
	})

	t.Run("DisableRepository", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create an enabled repository
		request := &RepositoryRequest{
			Name:     "disable-test-repo",
			TenantID: "test-tenant",
			Type:     RepositoryTypeScaleOut,
			Enabled:  true,
		}

		repo, err := manager.CreateRepository(ctx, request)
		if err != nil {
			t.Fatalf("Failed to create repository: %v", err)
		}

		// Disable the repository
		err = manager.DisableRepository(ctx, repo.ID)
		if err != nil {
			t.Fatalf("Failed to disable repository: %v", err)
		}

		// Verify it's disabled
		disabled, err := manager.GetRepository(ctx, repo.ID)
		if err != nil {
			t.Fatalf("Failed to get repository: %v", err)
		}

		if disabled.Enabled {
			t.Error("Repository should be disabled")
		}
	})

	t.Run("DeleteRepository", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create a repository
		request := &RepositoryRequest{
			Name:     "delete-test-repo",
			TenantID: "test-tenant",
			Type:     RepositoryTypeDistributed,
		}

		repo, err := manager.CreateRepository(ctx, request)
		if err != nil {
			t.Fatalf("Failed to create repository: %v", err)
		}

		// Delete the repository
		err = manager.DeleteRepository(ctx, repo.ID)
		if err != nil {
			t.Fatalf("Failed to delete repository: %v", err)
		}

		// Verify it's gone
		_, err = manager.GetRepository(ctx, repo.ID)
		if err == nil {
			t.Error("Expected error when getting deleted repository")
		}
	})

	t.Run("CreateStoragePool", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		request := &StoragePoolRequest{
			Name:        "hdd-storage-pool",
			TenantID:    "test-tenant",
			Description: "HDD storage pool for archive data",
			Type:        StoragePoolTypeHDD,
			Nodes: []StorageNode{
				{
					ID:   "node-hdd-1",
					Name: "hdd-node-1",
					Type: StorageNodeTypePrimary,
					Host: "hdd-host-1",
					Port: 8081,
					Configuration: NodeConfig{
						MaxConnections: 500,
						Timeout:        60 * time.Second,
						RetryCount:     5,
						RetryDelay:     10 * time.Second,
						KeepAlive:      true,
						Compression:    false,
						Encryption:     true,
					},
					Capacity: NodeCapacity{
						TotalStorage:      2000000000000, // 2TB
						UsedStorage:       200000000000,  // 200GB
						AvailableStorage:  1800000000000, // 1.8TB
						MaxIOPS:           5000,
						CurrentIOPS:       500,
						MaxThroughput:     200000000, // 200MB/s
						CurrentThroughput: 10000000,  // 10MB/s
					},
					Performance: NodePerformance{
						Latency:      20 * time.Millisecond,
						Throughput:   10000000, // 10MB/s
						IOPS:         500,
						CPUUsage:     15.5,
						MemoryUsage:  35.2,
						NetworkUsage: 10.8,
						DiskUsage:    20.1,
						ResponseTime: 10 * time.Millisecond,
					},
					HealthStatus: NodeHealthStatus{
						Status:     HealthStatusHealthy,
						LastCheck:  time.Now(),
						ErrorCount: 0,
						Uptime:     time.Hour * 24 * 30, // 30 days
						Version:    "1.5.0",
					},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					LastSeen:  &[]time.Time{time.Now()}[0],
				},
			},
			Configuration: PoolConfig{
				Replication:   2,
				Consistency:   ConsistencyTypeEventual,
				Quorum:        1,
				Timeout:       60 * time.Second,
				RetryCount:    5,
				RetryDelay:    10 * time.Second,
				HealthCheck:   true,
				AutoRebalance: false,
				LoadBalancing: true,
			},
			LoadBalancing: LoadBalancingConfig{
				Algorithm: AlgorithmRoundRobin,
				Weights: map[string]float64{
					"node-hdd-1": 1.0,
				},
				HealthCheck: HealthCheckConfig{
					Type:     "tcp",
					Endpoint: "10.0.2.50:8081",
					Interval: 60 * time.Second,
					Timeout:  15 * time.Second,
					Retries:  3,
				},
				Failover: FailoverConfig{
					Enabled:    true,
					Threshold:  1,
					Timeout:    60 * time.Second,
					RetryCount: 5,
					RetryDelay: 10 * time.Second,
				},
			},
			Enabled: true,
		}

		pool, err := manager.CreateStoragePool(ctx, request)
		if err != nil {
			t.Fatalf("Failed to create storage pool: %v", err)
		}

		if pool.ID == "" {
			t.Error("Pool ID should not be empty")
		}

		if pool.Name != request.Name {
			t.Errorf("Expected name %s, got %s", request.Name, pool.Name)
		}

		if pool.TenantID != request.TenantID {
			t.Errorf("Expected tenant ID %s, got %s", request.TenantID, pool.TenantID)
		}

		if pool.Type != request.Type {
			t.Errorf("Expected type %s, got %s", request.Type, pool.Type)
		}

		if !pool.Enabled {
			t.Error("Pool should be enabled")
		}

		if len(pool.Nodes) != len(request.Nodes) {
			t.Errorf("Expected %d nodes, got %d", len(request.Nodes), len(pool.Nodes))
		}
	})

	t.Run("GetStoragePool", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// First create a pool
		request := &StoragePoolRequest{
			Name:     "get-test-pool",
			TenantID: "test-tenant",
			Type:     StoragePoolTypeSSD,
			Enabled:  false,
		}

		created, err := manager.CreateStoragePool(ctx, request)
		if err != nil {
			t.Fatalf("Failed to create storage pool: %v", err)
		}

		// Retrieve the pool
		retrieved, err := manager.GetStoragePool(ctx, created.ID)
		if err != nil {
			t.Fatalf("Failed to get storage pool: %v", err)
		}

		if retrieved.ID != created.ID {
			t.Errorf("Expected ID %s, got %s", created.ID, retrieved.ID)
		}

		if retrieved.Name != created.Name {
			t.Errorf("Expected name %s, got %s", created.Name, retrieved.Name)
		}
	})

	t.Run("ListStoragePools", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create multiple pools
		for i := 0; i < 3; i++ {
			request := &StoragePoolRequest{
				Name:     fmt.Sprintf("list-test-pool-%d", i),
				TenantID: "test-tenant",
				Type:     StoragePoolTypeHybrid,
				Enabled:  i%2 == 0, // Enable every other pool
			}

			_, err := manager.CreateStoragePool(ctx, request)
			if err != nil {
				t.Fatalf("Failed to create storage pool %d: %v", i, err)
			}
		}

		// List all pools for tenant
		pools, err := manager.ListStoragePools(ctx, &StoragePoolFilter{
			TenantID: "test-tenant",
		})
		if err != nil {
			t.Fatalf("Failed to list storage pools: %v", err)
		}

		if len(pools) < 3 {
			t.Errorf("Expected at least 3 pools, got %d", len(pools))
		}

		// Filter by enabled status
		enabledPools, err := manager.ListStoragePools(ctx, &StoragePoolFilter{
			TenantID: "test-tenant",
			Enabled:  &[]bool{true}[0],
		})
		if err != nil {
			t.Fatalf("Failed to list enabled storage pools: %v", err)
		}

		// Should have some enabled pools
		if len(enabledPools) < 1 {
			t.Errorf("Expected at least 1 enabled pool, got %d", len(enabledPools))
		}
	})

	t.Run("AddStorageNode", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// First create a pool
		poolRequest := &StoragePoolRequest{
			Name:     "add-node-test-pool",
			TenantID: "test-tenant",
			Type:     StoragePoolTypeSSD,
			Enabled:  true,
		}

		pool, err := manager.CreateStoragePool(ctx, poolRequest)
		if err != nil {
			t.Fatalf("Failed to create storage pool: %v", err)
		}

		// Add a node to the pool
		nodeRequest := &AddNodeRequest{
			Name: "new-ssd-node",
			Type: StorageNodeTypeSecondary,
			Host: "ssd-host-2",
			Port: 8082,
			Configuration: NodeConfig{
				MaxConnections: 1000,
				Timeout:        30 * time.Second,
				RetryCount:     3,
				RetryDelay:     5 * time.Second,
				KeepAlive:      true,
				Compression:    true,
				Encryption:     true,
			},
			NetworkConfig: NetworkConfig{
				Subnets: []SubnetConfig{
					{
						ID:         "subnet-2",
						Name:       "storage-subnet-2",
						CIDR:       "10.0.2.0/24",
						Gateway:    "10.0.2.1",
						DNSServers: []string{"10.0.2.10", "10.0.2.11"},
						VLAN:       200,
					},
				},
			},
			StorageConfig: StorageConfig{
				Datastores: []DatastoreConfig{
					{
						ID:        "ds-ssd-2",
						Name:      "SSD Datastore 2",
						Type:      "ssd",
						Capacity:  1000000000000, // 1TB
						Used:      0,
						Available: 1000000000000, // 1TB
						Tier:      "performance",
					},
				},
				StoragePolicy: StoragePolicy{
					Name:        "ssd-performance-2",
					Type:        "ssd",
					Tier:        "performance",
					Replication: 1,
				},
				ThinProvision: true,
				Encryption:    true,
				Compression:   true,
			},
		}

		err = manager.AddStorageNode(ctx, pool.ID, nodeRequest)
		if err != nil {
			t.Fatalf("Failed to add storage node: %v", err)
		}

		// Verify the node was added
		updatedPool, err := manager.GetStoragePool(ctx, pool.ID)
		if err != nil {
			t.Fatalf("Failed to get updated storage pool: %v", err)
		}

		if len(updatedPool.Nodes) != 1 {
			t.Errorf("Expected 1 node, got %d", len(updatedPool.Nodes))
		}
	})

	t.Run("CreateDataMover", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		request := &DataMoverRequest{
			Name:        "restore-data-mover",
			TenantID:    "test-tenant",
			Description: "Data mover for restore operations",
			Type:        DataMoverTypeRestore,
			Nodes: []ProcessingNode{
				{
					ID:   "proc-node-restore-1",
					Name: "restore-processing-node-1",
					Type: ProcessingNodeTypeWorker,
					Host: "proc-host-restore-1",
					Port: 9091,
					Configuration: NodeConfig{
						MaxConnections: 200,
						Timeout:        120 * time.Second,
						RetryCount:     10,
						RetryDelay:     20 * time.Second,
						KeepAlive:      true,
						Compression:    false,
						Encryption:     true,
					},
					Capacity: NodeCapacity{
						TotalStorage:      500000000000, // 500GB
						UsedStorage:       50000000000,  // 50GB
						AvailableStorage:  450000000000, // 450GB
						MaxIOPS:           25000,
						CurrentIOPS:       5000,
						MaxThroughput:     500000000, // 500MB/s
						CurrentThroughput: 50000000,  // 50MB/s
					},
					Performance: NodePerformance{
						Latency:      50 * time.Millisecond,
						Throughput:   50000000, // 50MB/s
						IOPS:         5000,
						CPUUsage:     45.5,
						MemoryUsage:  65.2,
						NetworkUsage: 35.8,
						DiskUsage:    50.1,
						ResponseTime: 20 * time.Millisecond,
					},
					HealthStatus: NodeHealthStatus{
						Status:     HealthStatusHealthy,
						LastCheck:  time.Now(),
						ErrorCount: 0,
						Uptime:     time.Hour * 24 * 7, // 1 week
						Version:    "3.0.0",
					},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					LastSeen:  &[]time.Time{time.Now()}[0],
				},
			},
			Configuration: MoverConfig{
				MaxConcurrency: 5,
				ChunkSize:      2097152,  // 2MB
				BufferSize:     20971520, // 20MB
				Timeout:        600 * time.Second,
				RetryCount:     10,
				RetryDelay:     20 * time.Second,
				Compression:    false,
				Encryption:     true,
				BandwidthLimit: 500000000, // 500MB/s
			},
			LoadBalancing: LoadBalancingConfig{
				Algorithm: AlgorithmLeastConnections,
				Weights: map[string]float64{
					"proc-node-restore-1": 1.0,
				},
			},
			Scheduling: SchedulingConfig{
				Algorithm:    SchedulingAlgorithmFIFO,
				Priority:     SchedulingPriorityNormal,
				QueueSize:    500,
				WorkerCount:  3,
				BatchSize:    5,
				BatchTimeout: 120 * time.Second,
			},
			Monitoring: MonitoringConfig{
				Enabled:   true,
				Metrics:   []string{"throughput", "latency", "error_rate", "queue_size"},
				Interval:  60 * time.Second,
				Retention: 14 * 24 * time.Hour,
				Alerting: AlertingConfig{
					Enabled: true,
					Thresholds: map[string]float64{
						"throughput": 50000000,    // 50MB/s
						"latency":    20000000000, // 20s
						"error_rate": 0.1,
						"queue_size": 400,
					},
					Channels: []string{"email", "slack"},
					Severity: "error",
				},
			},
			Enabled: true,
		}

		mover, err := manager.CreateDataMover(ctx, request)
		if err != nil {
			t.Fatalf("Failed to create data mover: %v", err)
		}

		if mover.ID == "" {
			t.Error("Data mover ID should not be empty")
		}

		if mover.Name != request.Name {
			t.Errorf("Expected name %s, got %s", request.Name, mover.Name)
		}

		if mover.TenantID != request.TenantID {
			t.Errorf("Expected tenant ID %s, got %s", request.TenantID, mover.TenantID)
		}

		if mover.Type != request.Type {
			t.Errorf("Expected type %s, got %s", request.Type, mover.Type)
		}

		if !mover.Enabled {
			t.Error("Data mover should be enabled")
		}

		if len(mover.Nodes) != len(request.Nodes) {
			t.Errorf("Expected %d nodes, got %d", len(request.Nodes), len(mover.Nodes))
		}
	})

	t.Run("StartDataMover", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create a data mover
		moverRequest := &DataMoverRequest{
			Name:     "start-test-mover",
			TenantID: "test-tenant",
			Type:     DataMoverTypeReplication,
			Enabled:  true,
		}

		mover, err := manager.CreateDataMover(ctx, moverRequest)
		if err != nil {
			t.Fatalf("Failed to create data mover: %v", err)
		}

		// Start the data mover
		err = manager.StartDataMover(ctx, mover.ID)
		if err != nil {
			t.Fatalf("Failed to start data mover: %v", err)
		}

		// Verify it's running
		running, err := manager.GetDataMover(ctx, mover.ID)
		if err != nil {
			t.Fatalf("Failed to get data mover: %v", err)
		}

		if running.Status != MoverStatusRunning {
			t.Errorf("Expected status %s, got %s", MoverStatusRunning, running.Status)
		}
	})

	t.Run("StopDataMover", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create a data mover
		moverRequest := &DataMoverRequest{
			Name:     "stop-test-mover",
			TenantID: "test-tenant",
			Type:     DataMoverTypeSync,
			Enabled:  true,
		}

		mover, err := manager.CreateDataMover(ctx, moverRequest)
		if err != nil {
			t.Fatalf("Failed to create data mover: %v", err)
		}

		// Start the data mover first
		err = manager.StartDataMover(ctx, mover.ID)
		if err != nil {
			t.Fatalf("Failed to start data mover: %v", err)
		}

		// Stop the data mover
		err = manager.StopDataMover(ctx, mover.ID)
		if err != nil {
			t.Fatalf("Failed to stop data mover: %v", err)
		}

		// Verify it's stopped
		stopped, err := manager.GetDataMover(ctx, mover.ID)
		if err != nil {
			t.Fatalf("Failed to get data mover: %v", err)
		}

		if stopped.Status != MoverStatusStopped {
			t.Errorf("Expected status %s, got %s", MoverStatusStopped, stopped.Status)
		}
	})

	t.Run("CreateLoadBalancer", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		request := &LoadBalancerRequest{
			Name:        "api-load-balancer",
			TenantID:    "test-tenant",
			Description: "Load balancer for API services",
			Type:        LoadBalancerTypeSoftware,
			Algorithm:   AlgorithmLeastResponseTime,
			Targets: []LoadBalancerTarget{
				{
					ID:      "target-api-1",
					Name:    "api-server-1",
					Host:    "api-host-1",
					Port:    8080,
					Weight:  1,
					Enabled: true,
					HealthCheck: HealthCheckConfig{
						Type:     "http",
						Endpoint: "/health",
						Interval: 30 * time.Second,
						Timeout:  10 * time.Second,
						Retries:  3,
					},
					Metadata: map[string]string{
						"service": "api",
						"version": "v1.0",
					},
				},
				{
					ID:      "target-api-2",
					Name:    "api-server-2",
					Host:    "api-host-2",
					Port:    8080,
					Weight:  1,
					Enabled: true,
					HealthCheck: HealthCheckConfig{
						Type:     "http",
						Endpoint: "/health",
						Interval: 30 * time.Second,
						Timeout:  10 * time.Second,
						Retries:  3,
					},
					Metadata: map[string]string{
						"service": "api",
						"version": "v1.0",
					},
				},
			},
			Configuration: BalancerConfig{
				MaxConnections: 5000,
				Timeout:        30 * time.Second,
				RetryCount:     3,
				RetryDelay:     5 * time.Second,
				KeepAlive:      true,
				HealthCheck:    true,
				Failover:       true,
				StickySessions: false,
			},
			HealthCheck: HealthCheckConfig{
				Type:     "tcp",
				Endpoint: "10.0.1.100:8080",
				Interval: 15 * time.Second,
				Timeout:  5 * time.Second,
				Retries:  3,
			},
			SessionConfig: SessionConfig{
				Enabled:     false,
				Type:        SessionTypeCookie,
				Timeout:     30 * time.Minute,
				Persistence: true,
				CookieName:  "session_id",
			},
			Monitoring: MonitoringConfig{
				Enabled:   true,
				Metrics:   []string{"connections", "response_time", "error_rate", "throughput"},
				Interval:  30 * time.Second,
				Retention: 7 * 24 * time.Hour,
				Alerting: AlertingConfig{
					Enabled: true,
					Thresholds: map[string]float64{
						"connections":   4000,
						"response_time": 5000000000, // 5s
						"error_rate":    0.01,
						"throughput":    10000000000, // 10MB/s
					},
					Channels: []string{"email", "slack"},
					Severity: "warning",
				},
			},
			Enabled: true,
		}

		balancer, err := manager.CreateLoadBalancer(ctx, request)
		if err != nil {
			t.Fatalf("Failed to create load balancer: %v", err)
		}

		if balancer.ID == "" {
			t.Error("Load balancer ID should not be empty")
		}

		if balancer.Name != request.Name {
			t.Errorf("Expected name %s, got %s", request.Name, balancer.Name)
		}

		if balancer.TenantID != request.TenantID {
			t.Errorf("Expected tenant ID %s, got %s", request.TenantID, balancer.TenantID)
		}

		if balancer.Type != request.Type {
			t.Errorf("Expected type %s, got %s", request.Type, balancer.Type)
		}

		if !balancer.Enabled {
			t.Error("Load balancer should be enabled")
		}

		if len(balancer.Targets) != len(request.Targets) {
			t.Errorf("Expected %d targets, got %d", len(request.Targets), len(balancer.Targets))
		}
	})

	t.Run("GetScaleOutStats", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create some repositories and pools
		for i := 0; i < 2; i++ {
			repoRequest := &RepositoryRequest{
				Name:     fmt.Sprintf("stats-test-repo-%d", i),
				TenantID: "test-tenant",
				Type:     RepositoryTypeScaleOut,
				Enabled:  true,
			}

			_, err := manager.CreateRepository(ctx, repoRequest)
			if err != nil {
				t.Fatalf("Failed to create repository %d: %v", i, err)
			}

			// Create a storage pool
			poolRequest := &StoragePoolRequest{
				Name:     fmt.Sprintf("stats-test-pool-%d", i),
				TenantID: "test-tenant",
				Type:     StoragePoolTypeSSD,
				Enabled:  true,
			}

			_, err = manager.CreateStoragePool(ctx, poolRequest)
			if err != nil {
				t.Fatalf("Failed to create storage pool %d: %v", i, err)
			}

			// Create a data mover
			moverRequest := &DataMoverRequest{
				Name:     fmt.Sprintf("stats-test-mover-%d", i),
				TenantID: "test-tenant",
				Type:     DataMoverTypeBackup,
				Enabled:  true,
			}

			_, err = manager.CreateDataMover(ctx, moverRequest)
			if err != nil {
				t.Fatalf("Failed to create data mover %d: %v", i, err)
			}

			// Create a load balancer
			balancerRequest := &LoadBalancerRequest{
				Name:     fmt.Sprintf("stats-test-balancer-%d", i),
				TenantID: "test-tenant",
				Type:     LoadBalancerTypeSoftware,
				Enabled:  true,
			}

			_, err = manager.CreateLoadBalancer(ctx, balancerRequest)
			if err != nil {
				t.Fatalf("Failed to create load balancer %d: %v", i, err)
			}
		}

		// Get tenant stats
		timeRange := TimeRange{
			From: time.Now().Add(-2 * time.Hour),
			To:   time.Now(),
		}

		stats, err := manager.GetScaleOutStats(ctx, "test-tenant", timeRange)
		if err != nil {
			t.Fatalf("Failed to get scale-out stats: %v", err)
		}

		if stats.TenantID != "test-tenant" {
			t.Errorf("Expected tenant ID %s, got %s", "test-tenant", stats.TenantID)
		}

		if stats.TotalRepositories < 2 {
			t.Errorf("Expected at least 2 total repositories, got %d", stats.TotalRepositories)
		}

		if stats.ActiveRepositories < 2 {
			t.Errorf("Expected at least 2 active repositories, got %d", stats.ActiveRepositories)
		}

		if stats.TotalPools < 2 {
			t.Errorf("Expected at least 2 total pools, got %d", stats.TotalPools)
		}

		if stats.TotalMovers < 2 {
			t.Errorf("Expected at least 2 total movers, got %d", stats.TotalMovers)
		}

		if stats.TotalBalancers < 2 {
			t.Errorf("Expected at least 2 total balancers, got %d", stats.TotalBalancers)
		}
	})

	t.Run("GetScaleOutSystemHealth", func(t *testing.T) {
		ctx := context.Background()

		// Get system health
		health, err := manager.GetScaleOutSystemHealth(ctx)
		if err != nil {
			t.Fatalf("Failed to get scale-out system health: %v", err)
		}

		if health.Status != HealthStatusHealthy {
			t.Errorf("Expected status %s, got %s", HealthStatusHealthy, health.Status)
		}

		if len(health.RepositoryHealth) == 0 {
			t.Error("Expected at least one repository health entry")
		}

		if len(health.PoolHealth) == 0 {
			t.Error("Expected at least one pool health entry")
		}

		if len(health.NodeHealth) == 0 {
			t.Error("Expected at least one node health entry")
		}

		if len(health.MoverHealth) == 0 {
			t.Error("Expected at least one mover health entry")
		}

		if len(health.BalancerHealth) == 0 {
			t.Error("Expected at least one balancer health entry")
		}

		if health.ErrorRate < 0 || health.ErrorRate > 1 {
			t.Errorf("Expected error rate between 0 and 1, got %f", health.ErrorRate)
		}

		if health.ResponseTime <= 0 {
			t.Errorf("Expected positive response time, got %v", health.ResponseTime)
		}
	})

	t.Run("GetActiveOperations", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create a data mover
		moverRequest := &DataMoverRequest{
			Name:     "active-operations-test-mover",
			TenantID: "test-tenant",
			Type:     DataMoverTypeMigration,
			Enabled:  true,
		}

		mover, err := manager.CreateDataMover(ctx, moverRequest)
		if err != nil {
			t.Fatalf("Failed to create data mover: %v", err)
		}

		// Start the data mover
		err = manager.StartDataMover(ctx, mover.ID)
		if err != nil {
			t.Fatalf("Failed to start data mover: %v", err)
		}

		// Get active operations
		activeOperations, err := manager.GetActiveOperations(ctx)
		if err != nil {
			t.Fatalf("Failed to get active operations: %v", err)
		}

		if len(activeOperations) < 1 {
			t.Errorf("Expected at least 1 active operation, got %d", len(activeOperations))
		}

		// Verify operation properties
		for _, operation := range activeOperations {
			if operation.Type != OperationTypeReplication {
				t.Errorf("Expected operation type %s, got %s", OperationTypeReplication, operation.Type)
			}
			if operation.TenantID != "test-tenant" {
				t.Errorf("Expected tenant ID %s, got %s", "test-tenant", operation.TenantID)
			}
			if operation.Status != ExecutionStatusRunning && operation.Status != ExecutionStatusPending {
				t.Errorf("Expected running or pending status, got %s", operation.Status)
			}
		}
	})
}
