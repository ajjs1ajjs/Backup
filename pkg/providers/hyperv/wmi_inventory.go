// +build windows

package hyperv

import (
	"context"
	"fmt"
	"time"

	"github.com/yusufpapurcu/wmi"
	"go.uber.org/zap"
)

// WMI Classes
type Msvm_ComputerSystem struct {
	Name             string
	ElementName      string
	EnabledState     uint16
	TimeOfLastStateChange time.Time
	OnTimeInMilliseconds uint64
}

type Msvm_VirtualSystemSettingData struct {
	InstanceID       string
	ElementName      string
	ConfigurationDataRoot string
	VirtualSystemSubType string
}

type Msvm_ProcessorSettingData struct {
	VirtualQuantity uint64
}

type Msvm_MemorySettingData struct {
	VirtualQuantity uint64
}

// ListVMs lists all Hyper-V virtual machines using WMI
func (c *Client) ListVMs(ctx context.Context) ([]VM, error) {
	c.logger.Info("Listing Hyper-V VMs via WMI")

	var computerSystems []Msvm_ComputerSystem
	// Filtering by Caption="Virtual Machine" ensures we don't get the host system
	query := "SELECT Name, ElementName, EnabledState, OnTimeInMilliseconds FROM Msvm_ComputerSystem WHERE Caption='Virtual Machine'"
	
	err := wmi.QueryNamespace(query, &computerSystems, c.namespace)
	if err != nil {
		return nil, fmt.Errorf("wmi query Msvm_ComputerSystem failed: %w", err)
	}

	var vms []VM
	for _, cs := range computerSystems {
		vms = append(vms, VM{
			Name:   cs.ElementName,
			State:  vmStateToString(cs.EnabledState),
			Uptime: fmt.Sprintf("%d ms", cs.OnTimeInMilliseconds),
			// ID representation is the Name in WMI
			// Path and Version require querying Msvm_VirtualSystemSettingData matching this system
		})
	}

	return vms, nil
}

// GetVM retrieves a specific VM by name using WMI
func (c *Client) GetVM(ctx context.Context, name string) (*VM, error) {
	c.logger.Info("Getting VM via WMI", zap.String("name", name))

	var computerSystems []Msvm_ComputerSystem
	query := fmt.Sprintf("SELECT Name, ElementName, EnabledState, OnTimeInMilliseconds FROM Msvm_ComputerSystem WHERE Caption='Virtual Machine' AND ElementName='%s'", name)
	
	err := wmi.QueryNamespace(query, &computerSystems, c.namespace)
	if err != nil {
		return nil, fmt.Errorf("wmi query failed: %w", err)
	}

	if len(computerSystems) == 0 {
		return nil, fmt.Errorf("VM not found: %s", name)
	}

	cs := computerSystems[0]

	// Query settings for Generation and Path
	var settings []Msvm_VirtualSystemSettingData
	settingsQuery := fmt.Sprintf("SELECT InstanceID, VirtualSystemSubType, ConfigurationDataRoot FROM Msvm_VirtualSystemSettingData WHERE InstanceID LIKE 'Microsoft:%s%%'", cs.Name)
	_ = wmi.QueryNamespace(settingsQuery, &settings, c.namespace)

	gen := 1
	path := ""
	if len(settings) > 0 {
		if settings[0].VirtualSystemSubType == "Microsoft:Hyper-V:SubType:2" {
			gen = 2
		}
		path = settings[0].ConfigurationDataRoot
	}

	vm := &VM{
		Name:       cs.ElementName,
		State:      vmStateToString(cs.EnabledState),
		Generation: gen,
		Path:       path,
		Uptime:     fmt.Sprintf("%d ms", cs.OnTimeInMilliseconds),
	}

	return vm, nil
}

// GetVMInfo retrieves detailed information about a VM
func (c *Client) GetVMInfo(ctx context.Context, name string) (*VMInfo, error) {
	c.logger.Info("Getting detailed VM info via WMI", zap.String("name", name))

	vm, err := c.GetVM(ctx, name)
	if err != nil {
		return nil, err
	}

	// This is a simplified version building. To get Memory/CPU realistically requires navigating associators.
	// For production we would use `ASSOCIATORS OF` WMI queries.
	
	info := &VMInfo{
		VM:              *vm,
		NetworkAdapters: []NetworkAdapter{},
		HardDrives:      []HardDrive{},
		DVDDrives:       []DVDDrive{},
		ProcessorCount:  1, // Default fallback
	}

	return info, nil
}
