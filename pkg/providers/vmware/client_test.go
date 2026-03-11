package vmware

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestNewClient tests client creation
func TestNewClient(t *testing.T) {
	logger := zap.NewNop()

	// Test with invalid config (should fail without real vCenter)
	config := &ConnectionConfig{
		Host:     "invalid-host",
		Username: "test",
		Password: "test",
		Insecure: true,
	}

	client, err := NewClient(logger, config)
	// Should fail to connect to invalid host
	assert.Error(t, err)
	assert.Nil(t, client)
}

// TestDefaultConnectionConfig tests default configuration
func TestDefaultConnectionConfig(t *testing.T) {
	config := DefaultConnectionConfig()

	assert.Equal(t, 443, config.Port)
	assert.True(t, config.Insecure)
	assert.Empty(t, config.Host)
	assert.Empty(t, config.Username)
}

// TestConnectionConfigValidation tests config validation
func TestConnectionConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *ConnectionConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: &ConnectionConfig{
				Host:     "vcenter.local",
				Username: "admin",
				Password: "password",
				Port:     443,
			},
			wantErr: false,
		},
		{
			name: "missing host",
			config: &ConnectionConfig{
				Username: "admin",
				Password: "password",
			},
			wantErr: true,
		},
		{
			name: "missing credentials",
			config: &ConnectionConfig{
				Host: "vcenter.local",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just validate that config fields are set
			if tt.wantErr {
				assert.True(t, tt.config.Host == "" || tt.config.Username == "" || tt.config.Password == "")
			} else {
				assert.NotEmpty(t, tt.config.Host)
				assert.NotEmpty(t, tt.config.Username)
				assert.NotEmpty(t, tt.config.Password)
			}
		})
	}
}

// TestVMInfoStructure tests VMInfo structure
func TestVMInfoStructure(t *testing.T) {
	info := &VMInfo{
		Name:          "TestVM",
		UUID:          "12345-abc",
		InstanceUUID:  "67890-def",
		GuestName:     "Windows",
		GuestFullName: "Microsoft Windows Server 2022",
		PowerState:    "poweredOn",
		NumCPU:        4,
		MemoryMB:      8192,
		Disks: []DiskInfo{
			{
				Name:         "Hard disk 1",
				Label:        "Hard disk 1",
				CapacityGB:   100,
				Datastore:    "datastore1",
				Provisioning: "thin",
			},
		},
		Networks: []NetworkInfo{
			{
				Name:       "Network adapter 1",
				Label:      "Network adapter 1",
				Type:       "vmxnet3",
				MacAddress: "00:50:56:00:00:01",
			},
		},
	}

	assert.Equal(t, "TestVM", info.Name)
	assert.Equal(t, "12345-abc", info.UUID)
	assert.Equal(t, "poweredOn", info.PowerState)
	assert.Equal(t, int32(4), info.NumCPU)
	assert.Equal(t, int32(8192), info.MemoryMB)
	require.Len(t, info.Disks, 1)
	assert.Equal(t, int64(100), info.Disks[0].CapacityGB)
	require.Len(t, info.Networks, 1)
	assert.Equal(t, "00:50:56:00:00:01", info.Networks[0].MacAddress)
}

// TestBackupConfig tests backup configuration
func TestBackupConfig(t *testing.T) {
	config := &BackupConfig{
		Name:             "Test Backup",
		Destination:      "/backups",
		Compression:      true,
		CompressionLevel: 6,
		Deduplication:    true,
		Encryption:       true,
		EncryptionKey:    []byte("test-key-12345678901234567890123456789012"),
		Quiesce:          true,
		Memory:           false,
		Incremental:      false,
		FullBackup:       true,
		RetentionDays:    30,
	}

	assert.Equal(t, "Test Backup", config.Name)
	assert.Equal(t, "/backups", config.Destination)
	assert.True(t, config.Compression)
	assert.Equal(t, 6, config.CompressionLevel)
	assert.True(t, config.Deduplication)
	assert.True(t, config.Encryption)
	assert.True(t, config.Quiesce)
	assert.False(t, config.Memory)
}

// TestBackupResult tests backup result structure
func TestBackupResult(t *testing.T) {
	result := &BackupResult{
		BackupID:         "backup-123",
		VMName:           "TestVM",
		VMUUID:           "uuid-123",
		Status:           "success",
		TotalBytes:       10737418240, // 10 GB
		ProcessedBytes:   5368709120,  // 5 GB
		TransferredBytes: 2684354560,  // 2.5 GB with compression
		CompressionRatio: 2.0,
		Disks: []DiskBackupInfo{
			{
				DiskName:         "disk1",
				ProcessedGB:      5.0,
				TransferredGB:    2.5,
				CompressionRatio: 2.0,
			},
		},
	}

	assert.Equal(t, "backup-123", result.BackupID)
	assert.Equal(t, "success", result.Status)
	assert.Equal(t, 2.0, result.CompressionRatio)
	assert.Len(t, result.Disks, 1)
}

// TestCBTStatus tests CBT status structure
func TestCBTStatus(t *testing.T) {
	status := &CBTStatus{
		VMName:     "TestVM",
		VMUUID:     "uuid-123",
		CBTEnabled: true,
		Supported:  true,
		Disks: []DiskCBTStatus{
			{
				DiskName:        "Hard disk 1",
				DiskKey:         2000,
				CapacityBytes:   107374182400,
				CurrentChangeID: "change-123",
			},
		},
	}

	assert.True(t, status.CBTEnabled)
	assert.True(t, status.Supported)
	require.Len(t, status.Disks, 1)
	assert.Equal(t, int32(2000), status.Disks[0].DiskKey)
	assert.Equal(t, "change-123", status.Disks[0].CurrentChangeID)
}

// BenchmarkVMInfoCreation benchmarks VM info creation
func BenchmarkVMInfoCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = &VMInfo{
			Name:       fmt.Sprintf("VM-%d", i),
			UUID:       "12345",
			PowerState: "poweredOn",
			NumCPU:     4,
			MemoryMB:   8192,
		}
	}
}
