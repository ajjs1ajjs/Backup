//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Embed the binary files
//
//go:embed nova.exe
const (
	installPath = `C:\Program Files\NovaBackup`
	dataPath    = `C:\ProgramData\NovaBackup`
)

func main() {
	fmt.Println("╔══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║              NovaBackup Enterprise v6.0 Setup                    ║")
	fmt.Println("║                   Professional Installer                         ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	if !isAdmin() {
		fmt.Println("[ERROR] Administrator rights required!")
		fmt.Println("Please right-click this file and select 'Run as administrator'")
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}

	fmt.Println("[*] Starting installation...")
	fmt.Println()

	// Step 1: Create directories
	fmt.Println("[1/5] Creating directories...")
	createDirs()
	fmt.Println("      OK")

	// Step 2: Extract files
	fmt.Println("[2/5] Extracting program files...")
	if err := extractFile(novaExe, filepath.Join(installPath, "nova.exe")); err != nil {
		fmt.Printf("      ERROR: %v\n", err)
		time.Sleep(3 * time.Second)
		os.Exit(1)
	}
	if err := extractFile(novaCliExe, filepath.Join(installPath, "nova-cli.exe")); err != nil {
		fmt.Printf("      ERROR: %v\n", err)
		time.Sleep(3 * time.Second)
		os.Exit(1)
	}
	fmt.Println("      OK")

	// Step 3: Create config
	fmt.Println("[3/5] Creating configuration...")
	createConfig()
	fmt.Println("      OK")

	// Step 4: Install service
	fmt.Println("[4/5] Installing Windows Service...")
	installService()
	fmt.Println("      OK")

	// Step 5: Start service
	fmt.Println("[5/5] Starting service...")
	startService()
	fmt.Println("      OK")

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║              Installation Complete Successfully!                 ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("Installation Path:", installPath)
	fmt.Println("Data Path:", dataPath)
	fmt.Println()
	fmt.Println("Web Console: http://localhost:8080")
	fmt.Println()
	fmt.Println("Default Login:")
	fmt.Println("  Username: admin")
	fmt.Println("  Password: admin123")
	fmt.Println()
	fmt.Println("CLI: \"", installPath+`\nova-cli.exe`, "\" --help")
	fmt.Println()
	fmt.Println("Press any key to exit...")
	time.Sleep(500 * time.Millisecond)
	exec.Command("cmd", "/c", "pause").Run()
}

func isAdmin() bool {
	_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	return err == nil
}

func createDirs() {
	dirs := []string{
		installPath,
		dataPath,
		dataPath + `\Logs`,
		dataPath + `\Backups`,
		dataPath + `\Config`,
	}
	for _, dir := range dirs {
		os.MkdirAll(dir, 0755)
	}
}

func extractFile(data []byte, dest string) error {
	file, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}

func createConfig() {
	config := `{
  "server": {
    "http_port": 8080,
    "https_port": 8443,
    "bind_address": "0.0.0.0"
  },
  "logging": {
    "level": "info",
    "file": "` + dataPath + `\\Logs\\novabackup.log"
  },
  "backup": {
    "default_path": "` + dataPath + `\\Backups",
    "retention_days": 30
  },
  "database": {
    "path": "` + dataPath + `\\novabackup.db"
  },
  "version": "6.0.0"
}`

	configFile := filepath.Join(dataPath, "Config", "config.json")
	os.WriteFile(configFile, []byte(config), 0644)
}

func installService() {
	// Stop and delete existing service
	exec.Command("cmd", "/c", "net", "stop", "NovaBackupService").Run()
	exec.Command("cmd", "/c", "sc", "delete", "NovaBackupService").Run()
	time.Sleep(2 * time.Second)

	// Create new service
	cmd := exec.Command("sc", "create", "NovaBackupService",
		"binPath=", `"`+installPath+`\nova.exe" --service`,
		"displayName=", "NovaBackup Enterprise Service",
		"start=", "auto")
	cmd.Run()
}

func startService() {
	cmd := exec.Command("net", "start", "NovaBackupService")
	cmd.Run()
}
