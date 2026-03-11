// +build windows

package hyperv

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/yusufpapurcu/wmi"
	"go.uber.org/zap"
)

// WMI Namespace for Hyper-V
const wmiNamespace = `root\virtualization\v2`

// Client represents a Hyper-V client
type Client struct {
	logger    *zap.Logger
	server    string // Hyper-V server hostname (empty for local)
	namespace string
}

// ConnectionConfig holds connection parameters for Hyper-V
type ConnectionConfig struct {
	Server   string // Hyper-V server hostname (empty for local)
	Username string // For remote Hyper-V (not fully implemented in wmi lib yet, assumes local for now)
	Password string // For remote Hyper-V
}

// VM represents a Hyper-V virtual machine
type VM struct {
	Name           string `json:"name"`
	State          string `json:"state"`
	Uptime         string `json:"uptime"`
	CPUUsage       int    `json:"cpu_usage"`
	MemoryAssigned int64  `json:"memory_assigned"`
	MemoryDemand   int64  `json:"memory_demand"`
	Status         string `json:"status"`
	Generation     int    `json:"generation"`
	Version        string `json:"version"`
	Path           string `json:"path"`
}

// VMInfo contains detailed information about a Hyper-V VM
type VMInfo struct {
	VM
	NetworkAdapters []NetworkAdapter `json:"network_adapters"`
	HardDrives      []HardDrive      `json:"hard_drives"`
	DVDDrives       []DVDDrive       `json:"dvd_drives"`
	ProcessorCount  int              `json:"processor_count"`
}

// NetworkAdapter represents a Hyper-V network adapter
type NetworkAdapter struct {
	Name       string `json:"name"`
	MacAddress string `json:"mac_address"`
	IPAddress  string `json:"ip_address"`
	SwitchName string `json:"switch_name"`
}

// HardDrive represents a Hyper-V virtual hard disk
type HardDrive struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	PoolName string `json:"pool_name"`
}

// DVDDrive represents a Hyper-V DVD drive
type DVDDrive struct {
	Name string `json:"name"`
	Path string `json:"path,omitempty"`
}

// Checkpoint represents a Hyper-V checkpoint (snapshot)
type Checkpoint struct {
	Name        string `json:"name"`
	ID          string `json:"id"`
	ParentID    string `json:"parent_id,omitempty"`
	CreatedTime string `json:"created_time"`
}

// RCTInfo contains Resilient Change Tracking information
type RCTInfo struct {
	VMName         string         `json:"vm_name"`
	Enabled        bool           `json:"enabled"`
	ReferencePoint string         `json:"reference_point,omitempty"`
	ChangedBlocks  []ChangedBlock `json:"changed_blocks,omitempty"`
}

// ChangedBlock represents changed blocks tracked by RCT
type ChangedBlock struct {
	Offset   int64  `json:"offset"`
	Length   int64  `json:"length"`
	VHDXPath string `json:"vhdx_path"`
}

// NewClient creates a new Hyper-V client
func NewClient(logger *zap.Logger, config *ConnectionConfig) (*Client, error) {
	// Initialize COM for WMI
	err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED)
	if err != nil {
		oleErr, ok := err.(*ole.OleError)
		// RPC_E_CHANGED_MODE means COM is already initialized
		if !ok || oleErr.Code() != ole.S_OK && oleErr.Code() != 0x80010106 {
			return nil, fmt.Errorf("failed to initialize COM: %w", err)
		}
	}
	// Note: We don't defer ole.CoUninitialize() here because the client is meant to be long-lived
	// Caller should handle cleanup if necessary, but typically WMI calls in Go initialize per thread

	ns := wmiNamespace
	if config.Server != "" && strings.ToLower(config.Server) != "localhost" && config.Server != "127.0.0.1" {
		ns = fmt.Sprintf(`\\%s\%s`, config.Server, wmiNamespace)
	}

	client := &Client{
		logger:    logger.With(zap.String("component", "hyperv-client")),
		server:    config.Server,
		namespace: ns,
	}

	// Test connection by checking if Hyper-V WMI namespace is accessible
	var vss []struct {
		Name string
	}
	err = wmi.QueryNamespace("SELECT Name FROM Msvm_ComputerSystem WHERE Name LIKE '%'", &vss, client.namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Hyper-V WMI namespace (is Hyper-V role installed?): %w", err)
	}

	return client, nil
}

// vmStateToString converts Hyper-V EnabledState enum to string
func vmStateToString(state uint16) string {
	switch state {
	case 2:
		return "Running"
	case 3:
		return "Off"
	case 32768:
		return "Paused"
	case 32769:
		return "Suspended"
	case 32770:
		return "Starting"
	case 32771:
		return "Snapshotting"
	case 32773:
		return "Saving"
	case 32774:
		return "Stopping"
	default:
		return "Unknown"
	}
}

// IsHyperVInstalled checks if Hyper-V is installed on the system
func IsHyperVInstalled() bool {
	cmd := exec.Command("powershell", "-Command", "Get-WindowsOptionalFeature -Online -FeatureName Microsoft-Hyper-V | Select-Object State")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "Enabled")
}

// invokeWmiMethod helper for invoking WMI methods that return a Job
func (c *Client) invokeWmiMethod(servicePath string, methodName string, inParams ...interface{}) (int32, string, error) {
	unknown, err := oleutil.CreateObject("WbemScripting.SWbemLocator")
	if err != nil {
		return 0, "", fmt.Errorf("CreateObject: %w", err)
	}
	defer unknown.Release()

	disp, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return 0, "", fmt.Errorf("QueryInterface: %w", err)
	}
	defer disp.Release()

	// ... Simplified for now. For complex WMI method invocations (like tracking Jobs), 
	// we will use PowerShell as a fallback for the specific action, or direct WMI queries.
	// We'll implement direct queries where possible in the other files.
	return 0, "", fmt.Errorf("not fully implemented")
}
