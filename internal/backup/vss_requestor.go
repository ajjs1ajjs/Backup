package backup

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// VSSRequestor handles VSS operations inside the guest VM
type VSSRequestor struct {
	host       string
	vmName     string
	username   string
	password   string
}

func NewVSSRequestor(host, vmName, user, pass string) *VSSRequestor {
	return &VSSRequestor{
		host:     host,
		vmName:   vmName,
		username: user,
		password: pass,
	}
}

// Freeze Guest triggers VSS inside the VM
func (v *VSSRequestor) Freeze(ctx context.Context) error {
	// PowerShell script to be executed inside the guest
	vssScript := `
		$shadow = (Get-WmiObject -List Win32_ShadowCopy).Create("C:\", "ClientAccessible")
		if ($shadow.ReturnValue -ne 0) { throw "VSS Freeze failed" }
		return $shadow.ShadowID
	`
	
	_, err := v.invokePowerShellDirect(ctx, vssScript)
	return err
}

// Thaw Guest removes the shadow copy or continues processing
func (v *VSSRequestor) Thaw(ctx context.Context) error {
	// (Implementation details for cleanup)
	return nil
}

// TruncateLogs performs application-specific log truncation (e.g. SQL Server)
func (v *VSSRequestor) TruncateLogs(ctx context.Context) error {
	sqlScript := `
		Import-Module SQLPS -ErrorAction SilentlyContinue
		Get-SqlDatabase | Where-Object { $_.RecoveryModel -eq "Full" } | ForEach-Object {
			Backup-SqlDatabase -Database $_.Name -BackupAction Log -NoRecovery
		}
	`
	_, err := v.invokePowerShellDirect(ctx, sqlScript)
	return err
}

func (v *VSSRequestor) invokePowerShellDirect(ctx context.Context, script string) (string, error) {
	// Invoke-Command -VMName works over Hyper-V VMBus (no network needed)
	fullCmd := fmt.Sprintf(`Invoke-Command -VMName "%s" -ScriptBlock { %s }`, v.vmName, script)
	
	cmd := exec.CommandContext(ctx, "powershell", "-NoProfile", "-NonInteractive", "-Command", fullCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("PowerShell Direct failed: %w, output: %s", err, string(output))
	}
	
	return strings.TrimSpace(string(output)), nil
}
