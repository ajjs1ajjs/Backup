package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"novabackup/pkg/providers/vmware"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// vmwareCmd represents the vmware command
var vmwareCmd = &cobra.Command{
	Use:   "vmware",
	Short: "VMware vSphere management commands",
	Long:  `Commands for managing VMware vSphere infrastructure including VMs, snapshots, and backups.`,
}

// vmwareConnectCmd connects to vCenter/ESXi
var vmwareConnectCmd = &cobra.Command{
	Use:   "connect [host]",
	Short: "Connect to vCenter or ESXi host",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		host := args[0]
		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")
		datacenter, _ := cmd.Flags().GetString("datacenter")
		insecure, _ := cmd.Flags().GetBool("insecure")

		logger, _ := zap.NewDevelopment()
		config := &vmware.ConnectionConfig{
			Host:       host,
			Username:   username,
			Password:   password,
			Datacenter: datacenter,
			Insecure:   insecure,
		}

		client, err := vmware.NewClient(logger, config)
		if err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}
		defer client.Close()

		fmt.Printf("✅ Successfully connected to %s\n", host)
		fmt.Printf("   User: %s\n", username)
		if datacenter != "" {
			fmt.Printf("   Datacenter: %s\n", datacenter)
		}
		return nil
	},
}

// vmwareListCmd lists VMs
var vmwareListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all virtual machines",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger, _ := zap.NewDevelopment()

		// Load connection from config or flags
		config := loadVMwareConfig(cmd)
		client, err := vmware.NewClient(logger, config)
		if err != nil {
			return err
		}
		defer client.Close()

		inventory := vmware.NewInventory(client)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		vms, err := inventory.ListVirtualMachines(ctx)
		if err != nil {
			return fmt.Errorf("failed to list VMs: %w", err)
		}

		// Output as table
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tPOWER STATE\tGUEST OS\tIP ADDRESS\tCPU\tMEMORY\tDISKS")
		fmt.Fprintln(w, "----\t-----------\t--------\t----------\t---\t------\t-----")

		for _, vm := range vms {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%d MB\t%d\n",
				vm.Name,
				vm.PowerState,
				vm.GuestOS,
				vm.IP,
				vm.NumCPU,
				vm.MemoryMB,
				vm.DiskCount,
			)
		}
		w.Flush()

		fmt.Printf("\nTotal: %d virtual machines\n", len(vms))
		return nil
	},
}

// vmwareInfoCmd shows VM details
var vmwareInfoCmd = &cobra.Command{
	Use:   "info [vm-name]",
	Short: "Show detailed information about a VM",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmName := args[0]
		logger, _ := zap.NewDevelopment()

		config := loadVMwareConfig(cmd)
		client, err := vmware.NewClient(logger, config)
		if err != nil {
			return err
		}
		defer client.Close()

		inventory := vmware.NewInventory(client)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		vm, err := inventory.GetVirtualMachine(ctx, vmName)
		if err != nil {
			return fmt.Errorf("VM not found: %w", err)
		}

		info, err := vm.GetInfo()
		if err != nil {
			return fmt.Errorf("failed to get VM info: %w", err)
		}

		// Output as JSON or formatted text
		jsonOutput, _ := cmd.Flags().GetBool("json")
		if jsonOutput {
			data, _ := json.MarshalIndent(info, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		// Formatted output
		fmt.Printf("Virtual Machine: %s\n", info.Name)
		fmt.Printf("=====================================\n")
		fmt.Printf("UUID: %s\n", info.UUID)
		fmt.Printf("Instance UUID: %s\n", info.InstanceUUID)
		fmt.Printf("Guest OS: %s\n", info.GuestFullName)
		fmt.Printf("Power State: %s\n", info.PowerState)
		fmt.Printf("CPU: %d vCPUs\n", info.NumCPU)
		fmt.Printf("Memory: %d MB\n", info.MemoryMB)
		fmt.Printf("CBT Enabled: %v\n", info.CBTEnabled)
		fmt.Printf("\nDisks:\n")
		for i, disk := range info.Disks {
			fmt.Printf("  %d. %s - %d GB\n", i+1, disk.Label, disk.CapacityGB)
		}
		fmt.Printf("\nNetworks:\n")
		for i, net := range info.Networks {
			fmt.Printf("  %d. %s - %s\n", i+1, net.Label, net.MacAddress)
		}

		return nil
	},
}

// vmwareSnapshotCmd manages snapshots
var vmwareSnapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Manage VM snapshots",
}

var vmwareSnapshotCreateCmd = &cobra.Command{
	Use:   "create [vm-name] [snapshot-name]",
	Short: "Create a snapshot of a VM",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmName := args[0]
		snapshotName := args[1]

		memory, _ := cmd.Flags().GetBool("memory")
		quiesce, _ := cmd.Flags().GetBool("quiesce")
		description, _ := cmd.Flags().GetString("description")

		logger, _ := zap.NewDevelopment()
		config := loadVMwareConfig(cmd)
		client, err := vmware.NewClient(logger, config)
		if err != nil {
			return err
		}
		defer client.Close()

		inventory := vmware.NewInventory(client)
		ctx := context.Background()

		vm, err := inventory.GetVirtualMachine(ctx, vmName)
		if err != nil {
			return err
		}

		task, err := vm.CreateSnapshot(snapshotName, description, memory, quiesce)
		if err != nil {
			return fmt.Errorf("failed to create snapshot: %w", err)
		}

		fmt.Printf("Creating snapshot '%s'...\n", snapshotName)
		if err := task.Wait(ctx); err != nil {
			return fmt.Errorf("snapshot creation failed: %w", err)
		}

		fmt.Printf("✅ Snapshot '%s' created successfully\n", snapshotName)
		return nil
	},
}

// vmwareBackupCmd backs up a VM
var vmwareBackupCmd = &cobra.Command{
	Use:   "backup [vm-name]",
	Short: "Backup a virtual machine",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmName := args[0]
		destination, _ := cmd.Flags().GetString("destination")
		compression, _ := cmd.Flags().GetBool("compression")
		encryption, _ := cmd.Flags().GetBool("encryption")
		incremental, _ := cmd.Flags().GetBool("incremental")

		if destination == "" {
			destination = "./backups"
		}

		logger, _ := zap.NewDevelopment()
		config := loadVMwareConfig(cmd)
		client, err := vmware.NewClient(logger, config)
		if err != nil {
			return err
		}
		defer client.Close()

		inventory := vmware.NewInventory(client)
		ctx := context.Background()

		vm, err := inventory.GetVirtualMachine(ctx, vmName)
		if err != nil {
			return err
		}

		backupConfig := &vmware.BackupConfig{
			Name:        fmt.Sprintf("%s-backup", vmName),
			Destination: destination,
			Compression: compression,
			Encryption:  encryption,
			Quiesce:     true,
		}

		if incremental {
			// Use incremental backup engine
			incEngine := vmware.NewIncrementalBackupEngine(client, filepath.Join(destination, "state"))
			result, err := incEngine.PerformIncrementalBackup(ctx, vm, backupConfig, func(progress vmware.BackupProgress) {
				fmt.Printf("\r%s: %.1f%% - %s", progress.Phase, progress.Percent, progress.Message)
			})
			if err != nil {
				return fmt.Errorf("backup failed: %w", err)
			}
			fmt.Printf("\n✅ Incremental backup completed: %s\n", result.BackupID)
		} else {
			// Use full backup
			engine := vmware.NewBackupEngine(client)
			result, err := engine.FullBackup(ctx, vm, backupConfig, func(progress vmware.BackupProgress) {
				fmt.Printf("\r%s: %.1f%% - %s", progress.Phase, progress.Percent, progress.Message)
			})
			if err != nil {
				return fmt.Errorf("backup failed: %w", err)
			}
			fmt.Printf("\n✅ Full backup completed: %s\n", result.BackupID)
		}

		return nil
	},
}

// vmwareCBTCmd manages CBT
var vmwareCBTCmd = &cobra.Command{
	Use:   "cbt",
	Short: "Manage Changed Block Tracking (CBT)",
}

var vmwareCBTEnableCmd = &cobra.Command{
	Use:   "enable [vm-name]",
	Short: "Enable CBT for a VM",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmName := args[0]
		logger, _ := zap.NewDevelopment()
		config := loadVMwareConfig(cmd)
		client, err := vmware.NewClient(logger, config)
		if err != nil {
			return err
		}
		defer client.Close()

		inventory := vmware.NewInventory(client)
		ctx := context.Background()

		vm, err := inventory.GetVirtualMachine(ctx, vmName)
		if err != nil {
			return err
		}

		cbtMgr := vmware.NewCBTManager(client)
		if err := cbtMgr.EnableCBTForVM(ctx, vm); err != nil {
			return fmt.Errorf("failed to enable CBT: %w", err)
		}

		fmt.Printf("✅ CBT enabled for VM '%s'\n", vmName)
		return nil
	},
}

var vmwareCBTStatusCmd = &cobra.Command{
	Use:   "status [vm-name]",
	Short: "Show CBT status for a VM",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmName := args[0]
		logger, _ := zap.NewDevelopment()
		config := loadVMwareConfig(cmd)
		client, err := vmware.NewClient(logger, config)
		if err != nil {
			return err
		}
		defer client.Close()

		inventory := vmware.NewInventory(client)
		ctx := context.Background()

		vm, err := inventory.GetVirtualMachine(ctx, vmName)
		if err != nil {
			return err
		}

		cbtMgr := vmware.NewCBTManager(client)
		status, err := cbtMgr.GetCBTStatus(ctx, vm)
		if err != nil {
			return fmt.Errorf("failed to get CBT status: %w", err)
		}

		fmt.Printf("CBT Status for VM '%s':\n", vmName)
		fmt.Printf("  Enabled: %v\n", status.CBTEnabled)
		fmt.Printf("  Supported: %v\n", status.Supported)
		fmt.Printf("  Disks:\n")
		for _, disk := range status.Disks {
			fmt.Printf("    - %s (Key: %d)\n", disk.DiskName, disk.DiskKey)
			if disk.CurrentChangeID != "" {
				fmt.Printf("      Change ID: %s\n", disk.CurrentChangeID)
			}
		}

		return nil
	},
}

// Helper functions
func loadVMwareConfig(cmd *cobra.Command) *vmware.ConnectionConfig {
	// Try to load from flags first, then from config file
	host, _ := cmd.Flags().GetString("host")
	username, _ := cmd.Flags().GetString("username")
	password, _ := cmd.Flags().GetString("password")
	datacenter, _ := cmd.Flags().GetString("datacenter")
	insecure, _ := cmd.Flags().GetBool("insecure")

	return &vmware.ConnectionConfig{
		Host:       host,
		Username:   username,
		Password:   password,
		Datacenter: datacenter,
		Insecure:   insecure,
	}
}

func init() {
	rootCmd.AddCommand(vmwareCmd)

	// Add subcommands
	vmwareCmd.AddCommand(vmwareConnectCmd)
	vmwareCmd.AddCommand(vmwareListCmd)
	vmwareCmd.AddCommand(vmwareInfoCmd)
	vmwareCmd.AddCommand(vmwareSnapshotCmd)
	vmwareCmd.AddCommand(vmwareBackupCmd)
	vmwareCmd.AddCommand(vmwareCBTCmd)

	// Snapshot subcommands
	vmwareSnapshotCmd.AddCommand(vmwareSnapshotCreateCmd)

	// CBT subcommands
	vmwareCBTCmd.AddCommand(vmwareCBTEnableCmd)
	vmwareCBTCmd.AddCommand(vmwareCBTStatusCmd)

	// Global flags for VMware commands
	vmwareCmd.PersistentFlags().String("host", "", "vCenter/ESXi hostname")
	vmwareCmd.PersistentFlags().String("username", "", "Username for authentication")
	vmwareCmd.PersistentFlags().String("password", "", "Password for authentication")
	vmwareCmd.PersistentFlags().String("datacenter", "", "Datacenter name (optional)")
	vmwareCmd.PersistentFlags().Bool("insecure", true, "Skip SSL certificate verification")

	// Connect command flags
	vmwareConnectCmd.Flags().String("username", "", "Username (required)")
	vmwareConnectCmd.Flags().String("password", "", "Password (required)")
	vmwareConnectCmd.Flags().String("datacenter", "", "Datacenter")
	vmwareConnectCmd.Flags().Bool("insecure", true, "Skip SSL verification")

	// Snapshot create flags
	vmwareSnapshotCreateCmd.Flags().Bool("memory", false, "Include memory in snapshot")
	vmwareSnapshotCreateCmd.Flags().Bool("quiesce", true, "Quiesce guest filesystem")
	vmwareSnapshotCreateCmd.Flags().String("description", "", "Snapshot description")

	// Backup flags
	vmwareBackupCmd.Flags().String("destination", "./backups", "Backup destination path")
	vmwareBackupCmd.Flags().Bool("compression", true, "Enable compression")
	vmwareBackupCmd.Flags().Bool("encryption", false, "Enable encryption")
	vmwareBackupCmd.Flags().Bool("incremental", false, "Perform incremental backup")
}
