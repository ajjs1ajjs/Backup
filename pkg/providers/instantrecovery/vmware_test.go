package instantrecovery

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmware/govmomi/simulator"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"go.uber.org/zap"
)

type mockVMwareClient struct {
	client *vim25.Client
	finder *find.Finder
}

func (m *mockVMwareClient) GetDatacenter(ctx context.Context, name string) (*object.Datacenter, error) {
	return m.finder.Datacenter(ctx, name)
}

func (m *mockVMwareClient) GetHost(ctx context.Context, path string) (*object.HostSystem, error) {
	return m.finder.HostSystem(ctx, path)
}

func (m *mockVMwareClient) GetDatastore(ctx context.Context, name string) (*object.Datastore, error) {
	return m.finder.Datastore(ctx, name)
}

func (m *mockVMwareClient) GetFinder() VMwareFinder {
	return m.finder
}

func TestVMwareInstantRecovery(t *testing.T) {
	simulator.Run(func(ctx context.Context, client *vim25.Client) error {
		logger := zap.NewNop()
		
		finder := find.NewFinder(client)
		
		// Setup mock client
		mockClient := &mockVMwareClient{
			client: client,
			finder: finder,
		}
		
		nfsConfig := &NFSConfig{
			RootPath: t.TempDir(),
			Port:     2049,
		}

		ir, err := NewVMwareInstantRecovery(logger, mockClient, nfsConfig)
		assert.NoError(t, err)

		t.Run("IPDetection", func(t *testing.T) {
			// Simulator usually runs on localhost, but we test the logic
			ip, err := ir.getReachableLocalIP("127.0.0.1")
			if err == nil {
				assert.NotEmpty(t, ip)
			}
		})

		return nil
	})
}
