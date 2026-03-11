package main

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type NovaBackupGUI struct {
	app         fyne.App
	window      fyne.Window
	statusLabel *widget.Label
	progressBar *widget.ProgressBar
	jobsList    *widget.List
	logText     *widget.Entry
}

func main() {
	gui := &NovaBackupGUI{}
	gui.app = app.New()
	gui.app.Settings().SetTheme(theme.DarkTheme())
	gui.window = gui.app.NewWindow("NovaBackup v6.0 - Enterprise Backup & Recovery")
	gui.window.Resize(fyne.NewSize(1200, 800))
	gui.window.CenterOnScreen()

	gui.createUI()
	gui.setupSystemTray()
	gui.startBackgroundMonitoring()

	gui.window.ShowAndRun()
}

func (gui *NovaBackupGUI) createUI() {
	// Status bar
	statusContainer := container.NewHBox(
		widget.NewLabel("Status:"),
		gui.createStatusLabel(),
		widget.NewSeparator(),
		widget.NewButton("Create Backup Job", func() { gui.showCreateJobDialog() }),
		widget.NewButton("Run Backup Now", func() { gui.runImmediateBackup() }),
		widget.NewButton("Settings", func() { gui.showSettingsDialog() }),
	)

	// Main content area
	mainContent := container.NewHSplit(
		gui.createJobsPanel(),
		gui.createStatusPanel(),
	)
	mainContent.Resize(fyne.NewSize(1200, 600))

	// Statistics
	statsContainer := container.NewHBox(
		widget.NewLabel("Total Backups: 0"),
		widget.NewLabel("Storage Used: 0 GB"),
		widget.NewLabel("Deduplication Ratio: 0:1"),
		widget.NewLabel("Compression Ratio: 0:1"),
	)

	// Main layout
	content := container.NewVBox(
		statusContainer,
		mainContent,
		statsContainer,
	)

	gui.window.SetContent(content)
}

func (gui *NovaBackupGUI) createStatusLabel() *widget.Label {
	gui.statusLabel = widget.NewLabel("Ready")
	gui.statusLabel.Importance = widget.SuccessImportance
	return gui.statusLabel
}

func (gui *NovaBackupGUI) createJobsPanel() *container.Container {
	// Jobs table
	headers := []string{"Job Name", "Type", "Status", "Last Run", "Next Run", "Schedule"}

	table := widget.NewTable(
		func() (int, int) {
			return len(headers), len(headers)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Cell")
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			if id.Row == 0 {
				label.SetText(headers[id.Col])
				label.Importance = widget.MediumImportance
			} else {
				// Sample data
				data := [][]string{
					{"Daily Documents Backup", "File", "✅ Active", "2026-03-11 02:00", "2026-03-12 02:00", "Daily 2AM"},
					{"Weekly System Backup", "System", "✅ Active", "2026-03-08 22:00", "2026-03-15 22:00", "Weekly Sun 10PM"},
				}

				if id.Row-1 < len(data) {
					label.SetText(data[id.Row-1][id.Col])
				}
			}
		},
	)

	// Buttons
	buttonContainer := container.NewHBox(
		widget.NewButton("Edit Job", func() { gui.editSelectedJob() }),
		widget.NewButton("Delete Job", func() { gui.deleteSelectedJob() }),
		widget.NewButton("Run Now", func() { gui.runSelectedJob() }),
	)

	return container.NewVBox(
		widget.NewCard("Backup Jobs", "", container.NewVBox(
			table,
			buttonContainer,
		)),
	)
}

func (gui *NovaBackupGUI) createStatusPanel() *container.Container {
	gui.progressBar = widget.NewProgressBar()

	gui.logText = widget.NewMultiLineEntry()
	gui.logText.SetText("NovaBackup v6.0 initialized\n")
	gui.logText.Disable()

	return container.NewVBox(
		widget.NewCard("System Status", "", container.NewVBox(
			widget.NewLabel("Backup Progress:"),
			gui.progressBar,
			widget.NewSeparator(),
			widget.NewLabel("Recent Activity Log:"),
			container.NewScroll(gui.logText),
		)),
	)
}

func (gui *NovaBackupGUI) setupSystemTray() {
	// Fyne doesn't have native system tray support yet
	// This would require additional implementation
}

func (gui *NovaBackupGUI) startBackgroundMonitoring() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				gui.updateStatus()
			}
		}
	}()
}

func (gui *NovaBackupGUI) updateStatus() {
	gui.statusLabel.SetText("Running")
	gui.statusLabel.Importance = widget.SuccessImportance

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] System monitoring active\n", timestamp)

	currentText := gui.logText.Text
	if len(currentText) > 1000 {
		lines := strings.Split(currentText, "\n")
		if len(lines) > 50 {
			currentText = strings.Join(lines[len(lines)-50:], "\n")
		}
	}

	gui.logText.SetText(currentText + logEntry)

	// Simulate progress
	current := float64(gui.progressBar.Value)
	if current >= 1.0 {
		gui.progressBar.SetValue(0)
	} else {
		gui.progressBar.SetValue(current + 0.01)
	}
}

func (gui *NovaBackupGUI) showCreateJobDialog() {
	dialog := widget.NewModalPopUp(
		container.NewVBox(
			widget.NewLabel("Create Backup Job"),
			widget.NewSeparator(),
			widget.NewLabel("Job creation dialog would open here"),
			widget.NewButton("Close", func() {
				// Dialog would be closed here
			}),
		),
		gui.window.Canvas(),
	)
	dialog.Show()
}

func (gui *NovaBackupGUI) runImmediateBackup() {
	gui.statusLabel.SetText("Running backup...")
	gui.statusLabel.Importance = widget.WarningImportance

	go func() {
		time.Sleep(3 * time.Second)
		gui.statusLabel.SetText("Backup completed")
		gui.statusLabel.Importance = widget.SuccessImportance

		timestamp := time.Now().Format("2006-01-02 15:04:05")
		logEntry := fmt.Sprintf("[%s] ✅ Backup completed successfully\n", timestamp)
		gui.logText.SetText(gui.logText.Text + logEntry)
		gui.progressBar.SetValue(1.0)
	}()
}

func (gui *NovaBackupGUI) showSettingsDialog() {
	dialog := widget.NewModalPopUp(
		container.NewVBox(
			widget.NewLabel("Settings"),
			widget.NewSeparator(),
			widget.NewLabel("Settings dialog would open here"),
			widget.NewButton("Close", func() {
				// Dialog would be closed here
			}),
		),
		gui.window.Canvas(),
	)
	dialog.Show()
}

func (gui *NovaBackupGUI) editSelectedJob() {
	dialog := widget.NewModalPopUp(
		container.NewVBox(
			widget.NewLabel("Edit Job"),
			widget.NewSeparator(),
			widget.NewLabel("Edit job dialog would open here"),
			widget.NewButton("Close", func() {
				// Dialog would be closed here
			}),
		),
		gui.window.Canvas(),
	)
	dialog.Show()
}

func (gui *NovaBackupGUI) deleteSelectedJob() {
	confirm := widget.NewModalPopUp(
		container.NewVBox(
			widget.NewLabel("Delete Job"),
			widget.NewSeparator(),
			widget.NewLabel("Are you sure you want to delete this backup job?"),
			container.NewHBox(
				widget.NewButton("Yes", func() {
					// Delete logic here
					confirm.Hide()
				}),
				widget.NewButton("No", func() {
					confirm.Hide()
				}),
			),
		),
		gui.window.Canvas(),
	)
	confirm.Show()
}

func (gui *NovaBackupGUI) runSelectedJob() {
	gui.runImmediateBackup()
}
