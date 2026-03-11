package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"novabackup/pkg/providers/hyperv"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var hypervCmd = &cobra.Command{
	Use:   "hyperv",
	Short: "Microsoft Hyper-V management commands",
	Long:  `Commands for managing Microsoft Hyper-V virtual machines.`,
}

var hypervListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all Hyper-V VMs",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger, _ := zap.NewDevelopment()
		client, err := hyperv.NewClient(logger, &hyperv.ConnectionConfig{})
		if err != nil {
			return err
		}

		ctx := context.Background()
		vms, err := client.ListVMs(ctx)
		if err != nil {
			return fmt.Errorf("failed to list VMs: %w", err)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tSTATE\tCPU\tMEMORY\tUPTIME")
		fmt.Fprintln(w, "----\t-----\t---\t------\t------")

		for _, vm := range vms {
			fmt.Fprintf(w, "%s\t%s\t%d%%\t%d MB\t%s\n",
				vm.Name, vm.State, vm.CPUUsage, vm.MemoryAssigned/1024/1024, vm.Uptime)
		}
		w.Flush()

		fmt.Printf("\nTotal: %d virtual machines\n", len(vms))
		return nil
	},
}

var hypervBackupCmd = &cobra.Command{
	Use:   "backup [vm-name]",
	Short: "Backup a Hyper-V VM",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmName := args[0]
		backupPath, _ := cmd.Flags().GetString("destination")

		logger, _ := zap.NewDevelopment()
		client, err := hyperv.NewClient(logger, &hyperv.ConnectionConfig{})
		if err != nil {
			return err
		}

		ctx := context.Background()
		fmt.Printf("Backing up VM '%s' to '%s'...\n", vmName, backupPath)

		if err := client.BackupVM(ctx, vmName, backupPath, false); err != nil {
			return fmt.Errorf("backup failed: %w", err)
		}

		fmt.Printf("✅ VM '%s' backed up successfully to '%s'\n", vmName, backupPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(hypervCmd)
	hypervCmd.AddCommand(hypervListCmd)
	hypervCmd.AddCommand(hypervBackupCmd)

	hypervBackupCmd.Flags().String("destination", "./backups/hyperv", "Backup destination path")
}
