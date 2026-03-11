// Package vmware provides VMware vSphere integration for NovaBackup
package vmware

import (
	"context"
	"fmt"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"go.uber.org/zap"
)

// Inventory provides VM inventory management
type Inventory struct {
	logger *zap.Logger
	client *Client
}

// InventoryItem represents an inventory item (VM, Host, Datastore, etc.)
type InventoryItem struct {
	Name       string                       `json:"name"`
	Type       string                       `json:"type"`
	Reference  types.ManagedObjectReference `json:"reference"`
	Parent     *InventoryItem               `json:"parent,omitempty"`
	Children   []*InventoryItem             `json:"children,omitempty"`
	Properties map[string]interface{}       `json:"properties,omitempty"`
}

// VMListItem represents a VM in list format
type VMListItem struct {
	Name         string `json:"name"`
	UUID         string `json:"uuid"`
	InstanceUUID string `json:"instance_uuid"`
	PowerState   string `json:"power_state"`
	GuestOS      string `json:"guest_os"`
	IP           string `json:"ip_address"`
	Host         string `json:"host"`
	Datastore    string `json:"datastore"`
	ResourcePool string `json:"resource_pool"`
	Folder       string `json:"folder"`
	NumCPU       int32  `json:"num_cpu"`
	MemoryMB     int32  `json:"memory_mb"`
	DiskCount    int    `json:"disk_count"`
	NetworkCount int    `json:"network_count"`
}

// NewInventory creates a new inventory manager
func NewInventory(client *Client) *Inventory {
	return &Inventory{
		logger: client.logger.With(zap.String("component", "inventory")),
		client: client,
	}
}

// ListVirtualMachines returns a flat list of all VMs
func (i *Inventory) ListVirtualMachines(ctx context.Context) ([]VMListItem, error) {
	i.logger.Info("Listing all virtual machines")

	// Create container view for VirtualMachines
	m := view.NewManager(i.client.GetClient().Client)
	v, err := m.CreateContainerView(ctx, i.client.GetClient().ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		return nil, fmt.Errorf("failed to create container view: %w", err)
	}
	defer v.Destroy(ctx)

	// Retrieve VM summary for all VMs
	var vms []mo.VirtualMachine
	err = v.Retrieve(ctx, []string{"VirtualMachine"}, []string{"summary", "config", "guest", "runtime", "datastore"}, &vms)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve VMs: %w", err)
	}

	result := make([]VMListItem, 0, len(vms))

	for _, vm := range vms {
		if vm.Config == nil {
			continue // Skip VMs that are being deleted
		}

		item := VMListItem{
			Name:         vm.Config.Name,
			UUID:         vm.Config.Uuid,
			InstanceUUID: vm.Config.InstanceUuid,
			GuestOS:      vm.Guest.GuestFullName,
			IP:           vm.Guest.IpAddress,
			NumCPU:       vm.Config.Hardware.NumCPU,
			MemoryMB:     vm.Config.Hardware.MemoryMB,
		}

		item.PowerState = string(vm.Runtime.PowerState)

		// Count disks
		for _, device := range vm.Config.Hardware.Device {
			if _, ok := device.(*types.VirtualDisk); ok {
				item.DiskCount++
			}
		}

		// Count networks
		for _, device := range vm.Config.Hardware.Device {
			switch device.(type) {
			case *types.VirtualE1000, *types.VirtualE1000e, *types.VirtualVmxnet,
				*types.VirtualVmxnet2, *types.VirtualVmxnet3, *types.VirtualPCNet32,
				*types.VirtualSriovEthernetCard:
				item.NetworkCount++
			}
		}

		// Get primary datastore
		if len(vm.Datastore) > 0 {
			item.Datastore = vm.Datastore[0].Value
		}

		result = append(result, item)
	}

	i.logger.Info("Found virtual machines", zap.Int("count", len(result)))
	return result, nil
}

// GetVirtualMachine returns a single VM by name or UUID
func (i *Inventory) GetVirtualMachine(ctx context.Context, identifier string) (*VM, error) {
	i.logger.Info("Getting virtual machine", zap.String("identifier", identifier))

	// Try to find by name first
	vm, err := i.findVMByName(ctx, identifier)
	if err == nil {
		return vm, nil
	}

	// Try to find by UUID
	vm, err = i.findVMByUUID(ctx, identifier)
	if err == nil {
		return vm, nil
	}

	return nil, fmt.Errorf("virtual machine not found: %s", identifier)
}

// findVMByName finds a VM by its name
func (i *Inventory) findVMByName(ctx context.Context, name string) (*VM, error) {
	ref, err := i.client.GetFinder().VirtualMachine(ctx, name)
	if err != nil {
		return nil, err
	}

	return NewVM(i.client, ref.Reference(), name), nil
}

// findVMByUUID finds a VM by its UUID
func (i *Inventory) findVMByUUID(ctx context.Context, uuid string) (*VM, error) {
	// Search all VMs
	vms, err := i.ListVirtualMachines(ctx)
	if err != nil {
		return nil, err
	}

	for _, vm := range vms {
		if vm.UUID == uuid || vm.InstanceUUID == uuid {
			return i.GetVirtualMachine(ctx, vm.Name)
		}
	}

	return nil, fmt.Errorf("VM with UUID %s not found", uuid)
}

// GetInventoryTree returns the full vCenter inventory tree
func (i *Inventory) GetInventoryTree(ctx context.Context) (*InventoryItem, error) {
	i.logger.Info("Building inventory tree")

	root := &InventoryItem{
		Name:      "vCenter",
		Type:      "Folder",
		Reference: i.client.GetClient().ServiceContent.RootFolder,
		Children:  []*InventoryItem{},
	}

	// Get all datacenters
	datacenters, err := i.client.GetFinder().DatacenterList(ctx, "*")
	if err != nil {
		return nil, fmt.Errorf("failed to list datacenters: %w", err)
	}

	for _, dc := range datacenters {
		dcItem := &InventoryItem{
			Name:      dc.Name(),
			Type:      "Datacenter",
			Reference: dc.Reference(),
			Parent:    root,
			Children:  []*InventoryItem{},
		}
		root.Children = append(root.Children, dcItem)

		// Get clusters in datacenter
		clusters, err := i.client.GetFinder().ClusterComputeResourceList(ctx, dc.InventoryPath+"/*")
		if err == nil {
			for _, cluster := range clusters {
				clusterItem := &InventoryItem{
					Name:      cluster.Name(),
					Type:      "ClusterComputeResource",
					Reference: cluster.Reference(),
					Parent:    dcItem,
					Children:  []*InventoryItem{},
				}
				dcItem.Children = append(dcItem.Children, clusterItem)

				// Get hosts in cluster
				hosts, err := cluster.Hosts(ctx)
				if err == nil {
					for _, host := range hosts {
						hostItem := &InventoryItem{
							Name:      host.Name(),
							Type:      "HostSystem",
							Reference: host.Reference(),
							Parent:    clusterItem,
							Children:  []*InventoryItem{},
						}
						clusterItem.Children = append(clusterItem.Children, hostItem)
					}
				}
			}
		}

		// Get standalone hosts
		hosts, err := i.client.GetFinder().HostSystemList(ctx, dc.InventoryPath+"/*")
		if err == nil {
			for _, host := range hosts {
				hostItem := &InventoryItem{
					Name:      host.Name(),
					Type:      "HostSystem",
					Reference: host.Reference(),
					Parent:    dcItem,
					Children:  []*InventoryItem{},
				}
				dcItem.Children = append(dcItem.Children, hostItem)
			}
		}

		// Get datastores
		datastores, err := i.client.GetFinder().DatastoreList(ctx, dc.InventoryPath+"/*")
		if err == nil {
			for _, ds := range datastores {
				dsItem := &InventoryItem{
					Name:      ds.Name(),
					Type:      "Datastore",
					Reference: ds.Reference(),
					Parent:    dcItem,
					Children:  []*InventoryItem{},
				}
				dcItem.Children = append(dcItem.Children, dsItem)
			}
		}

		// Get VMs
		vms, err := i.client.GetFinder().VirtualMachineList(ctx, dc.InventoryPath+"/*")
		if err == nil {
			for _, vm := range vms {
				vmItem := &InventoryItem{
					Name:      vm.Name(),
					Type:      "VirtualMachine",
					Reference: vm.Reference(),
					Parent:    dcItem,
					Children:  []*InventoryItem{},
				}
				dcItem.Children = append(dcItem.Children, vmItem)
			}
		}

		// Get resource pools
		pools, err := i.client.GetFinder().ResourcePoolList(ctx, dc.InventoryPath+"/host/*")
		if err == nil {
			for _, pool := range pools {
				poolItem := &InventoryItem{
					Name:      pool.Name(),
					Type:      "ResourcePool",
					Reference: pool.Reference(),
					Parent:    dcItem,
					Children:  []*InventoryItem{},
				}
				dcItem.Children = append(dcItem.Children, poolItem)
			}
		}
	}

	return root, nil
}

// FindVMsByPattern finds VMs matching a pattern (glob)
func (i *Inventory) FindVMsByPattern(ctx context.Context, pattern string) ([]*VM, error) {
	i.logger.Info("Finding VMs by pattern", zap.String("pattern", pattern))

	vms, err := i.client.GetFinder().VirtualMachineList(ctx, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to find VMs: %w", err)
	}

	result := make([]*VM, 0, len(vms))
	for _, vm := range vms {
		result = append(result, NewVM(i.client, vm.Reference(), vm.Name()))
	}

	return result, nil
}

// GetVMsByTag finds VMs by vCenter tag
func (i *Inventory) GetVMsByTag(ctx context.Context, tag string) ([]*VM, error) {
	// This requires vSphere tagging API
	// Implementation depends on vCenter version
	return nil, fmt.Errorf("tag-based search not yet implemented")
}

// GetVMsInFolder returns all VMs in a specific folder
func (i *Inventory) GetVMsInFolder(ctx context.Context, folderPath string) ([]*VM, error) {
	i.logger.Info("Getting VMs in folder", zap.String("folder", folderPath))

	folder, err := i.client.GetFinder().Folder(ctx, folderPath)
	if err != nil {
		return nil, fmt.Errorf("folder not found: %w", err)
	}

	// Create container view for this folder
	m := view.NewManager(i.client.GetClient().Client)
	v, err := m.CreateContainerView(ctx, folder.Reference(), []string{"VirtualMachine"}, true)
	if err != nil {
		return nil, fmt.Errorf("failed to create container view: %w", err)
	}
	defer v.Destroy(ctx)

	// Get VMs
	var vms []mo.VirtualMachine
	err = v.Retrieve(ctx, []string{"VirtualMachine"}, []string{"name"}, &vms)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve VMs: %w", err)
	}

	result := make([]*VM, 0, len(vms))
	for _, vm := range vms {
		result = append(result, NewVM(i.client, vm.Reference(), vm.Name))
	}

	return result, nil
}

// GetVMsInResourcePool returns all VMs in a resource pool
func (i *Inventory) GetVMsInResourcePool(ctx context.Context, poolPath string) ([]*VM, error) {
	i.logger.Info("Getting VMs in resource pool", zap.String("pool", poolPath))

	pool, err := i.client.GetFinder().ResourcePool(ctx, poolPath)
	if err != nil {
		return nil, fmt.Errorf("resource pool not found: %w", err)
	}

	// Get all VMs in the pool
	var rp mo.ResourcePool
	err = i.client.GetClient().RetrieveOne(ctx, pool.Reference(), []string{"vm"}, &rp)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve resource pool: %w", err)
	}

	result := make([]*VM, 0, len(rp.Vm))
	for _, ref := range rp.Vm {
		vm := object.NewVirtualMachine(i.client.GetClient().Client, ref)
		name, _ := vm.ObjectName(ctx)
		result = append(result, NewVM(i.client, ref, name))
	}

	return result, nil
}

// Refresh updates the inventory cache
func (i *Inventory) Refresh(ctx context.Context) error {
	// Reset any internal caches
	i.logger.Info("Refreshing inventory")
	return nil
}
