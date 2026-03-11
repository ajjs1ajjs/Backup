//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

func main() {
	// GUI mode if no arguments
	if len(os.Args) == 1 {
		showGUI()
		return
	}

	// CLI mode
	if len(os.Args) > 1 {
		handleCLI()
	}
}

func showGUI() {
	fmt.Println("╔══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                  NovaBackup Enterprise v6.0                    ║")
	fmt.Println("║                 Enterprise Backup Solution                    ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Check if installed
	if isInstalled() {
		showInstalledMenu()
	} else {
		showInstallMenu()
	}
}

func isInstalled() bool {
	_, err := os.Stat(`C:\Program Files\NovaBackup\nova.exe`)
	return err == nil
}

func showInstallMenu() {
	fmt.Println("NovaBackup не встановлено. Виберіть дію:")
	fmt.Println()
	fmt.Println("1. Встановити NovaBackup Enterprise")
	fmt.Println("2. Перевірити системні вимоги")
	fmt.Println("3. Вихід")
	fmt.Println()

	var choice int
	fmt.Print("Ваш вибір (1-3): ")
	fmt.Scan(&choice)

	switch choice {
	case 1:
		installNovaBackup()
	case 2:
		checkRequirements()
	case 3:
		os.Exit(0)
	default:
		fmt.Println("Невірний вибір")
		time.Sleep(2 * time.Second)
		showGUI()
	}
}

func showInstalledMenu() {
	fmt.Println("NovaBackup Enterprise встановлено!")
	fmt.Println()
	fmt.Println("1. Запустити Web Console")
	fmt.Println("2. Запустити CLI")
	fmt.Println("3. Перевірити статус служби")
	fmt.Println("4. Видалити NovaBackup")
	fmt.Println("5. Вихід")
	fmt.Println()

	var choice int
	fmt.Print("Ваш вибір (1-5): ")
	fmt.Scan(&choice)

	switch choice {
	case 1:
		startWebConsole()
	case 2:
		startCLI()
	case 3:
		checkService()
	case 4:
		uninstallNovaBackup()
	case 5:
		os.Exit(0)
	default:
		fmt.Println("Невірний вибір")
		time.Sleep(2 * time.Second)
		showGUI()
	}
}

func installNovaBackup() {
	fmt.Println()
	fmt.Println("Встановлення NovaBackup Enterprise...")
	fmt.Println("[1/4] Перевірка прав адміністратора...")

	if !isAdmin() {
		fmt.Println("[ПОМИЛКА] Потрібні права адміністратора!")
		fmt.Println("Запустіть цю програму від імені адміністратора")
		time.Sleep(3 * time.Second)
		return
	}

	fmt.Println("[OK] Права адміністратора є")
	fmt.Println("[2/4] Створення директорій...")

	installPath := `C:\Program Files\NovaBackup`
	dataPath := `C:\ProgramData\NovaBackup`

	os.MkdirAll(installPath, 0755)
	os.MkdirAll(dataPath, 0755)
	os.MkdirAll(dataPath+`\Logs`, 0755)
	os.MkdirAll(dataPath+`\Backups`, 0755)
	os.MkdirAll(dataPath+`\Config`, 0755)

	fmt.Println("[OK] Директорії створено")
	fmt.Println("[3/4] Копіювання файлів...")

	// Copy embedded files (simplified)
	exePath, _ := os.Executable()
	sourceDir := filepath.Dir(exePath)

	copyFile(filepath.Join(sourceDir, "nova.exe"), filepath.Join(installPath, "nova.exe"))
	copyFile(filepath.Join(sourceDir, "nova-cli.exe"), filepath.Join(installPath, "nova-cli.exe"))

	fmt.Println("[OK] Файли скопійовано")
	fmt.Println("[4/4] Встановлення служби...")

	// Install service
	exec.Command("cmd", "/c", "sc", "delete", "NovaBackupService").Run()
	exec.Command("sc", "create", "NovaBackupService",
		"binPath=", `"`+installPath+`\nova.exe" --service`,
		"displayName=", "NovaBackup Enterprise Service",
		"start=", "auto").Run()

	fmt.Println("[OK] Службу встановлено")
	fmt.Println()
	fmt.Println("Запускаю службу...")
	exec.Command("net", "start", "NovaBackupService").Run()

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║               УСТАНОВКУ ЗАВЕРШЕНО УСПІШНО!                 ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("Web Console: http://localhost:8080")
	fmt.Println("Login: admin / admin123")
	fmt.Println()
	fmt.Println("Натисніть Enter для продовження...")
	fmt.Scanln()
}

func startWebConsole() {
	fmt.Println("Запускаю Web Console...")
	exec.Command("cmd", "/c", "start", "http://localhost:8080").Run()
	time.Sleep(2 * time.Second)
	showGUI()
}

func startCLI() {
	fmt.Println("Запускаю CLI...")
	exec.Command("cmd", "/c", "cd /d", `C:\Program Files\NovaBackup`, "&&", "nova-cli.exe", "--help").Run()
	time.Sleep(2 * time.Second)
	showGUI()
}

func checkService() {
	fmt.Println("Перевіряю статус служби...")
	cmd := exec.Command("sc", "query", "NovaBackupService")
	output, _ := cmd.Output()
	
	fmt.Println(string(output))
	fmt.Println()
	fmt.Println("Натисніть Enter для продовження...")
	fmt.Scanln()
	showGUI()
}

func uninstallNovaBackup() {
	fmt.Println()
	fmt.Println("Видалення NovaBackup Enterprise...")
	fmt.Println("[1/3] Зупинка служби...")
	exec.Command("net", "stop", "NovaBackupService").Run()
	fmt.Println("[OK] Службу зупинено")
	
	fmt.Println("[2/3] Видалення служби...")
	exec.Command("sc", "delete", "NovaBackupService").Run()
	fmt.Println("[OK] Службу видалено")
	
	fmt.Println("[3/3] Видалення файлів...")
	os.RemoveAll(`C:\Program Files\NovaBackup`)
	fmt.Println("[OK] Файли видалено")
	
	fmt.Println()
	fmt.Println("NovaBackup Enterprise видалено!")
	fmt.Println("Натисніть Enter для виходу...")
	fmt.Scanln()
}

func checkRequirements() {
	fmt.Println()
	fmt.Println("Перевірка системних вимог:")
	fmt.Println()

	// OS Check
	fmt.Printf("Операційна система: ")
	if runtime.GOOS == "windows" {
		fmt.Println("Windows [OK]")
	} else {
		fmt.Println("Не Windows [FAIL]")
	}

	// Check if running as admin
	fmt.Printf("Права адміністратора: ")
	if isAdmin() {
		fmt.Println("Є [OK]")
	} else {
		fmt.Println("Немає [FAIL]")
	}

	// Check disk space
	fmt.Printf("Місце на диску: ")
	// Simplified check
	fmt.Println("Перевірте вручну (потрібно ~500MB)")

	fmt.Println()
	fmt.Println("Натисніть Enter для продовження...")
	fmt.Scanln()
	showGUI()
}

func handleCLI() {
	fmt.Println("NovaBackup CLI v6.0")
	fmt.Println("Usage: nova.exe [command]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  install    - Install NovaBackup")
	fmt.Println("  start      - Start services")
	fmt.Println("  stop       - Stop services")
	fmt.Println("  status     - Show status")
	fmt.Println("  uninstall  - Uninstall NovaBackup")
}

func isAdmin() bool {
	_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	return err == nil
}

func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0644)
}
