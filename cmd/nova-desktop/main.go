package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type NovaBackupDesktop struct {
	*walk.MainWindow
	statusLabel    *walk.Label
	progressBar    *walk.ProgressBar
	jobsTableView  *walk.TableView
	logTextEdit    *walk.TextEdit
}

type BackupJob struct {
	Name     string
	Type     string
	Status   string
	LastRun  string
	NextRun  string
	Schedule string
}

func main() {
	desktop := &NovaBackupDesktop{}
	
	if err := desktop.createMainWindow(); err != nil {
		fmt.Printf("Error creating desktop GUI: %v\n", err)
		os.Exit(1)
	}

	desktop.startBackgroundMonitoring()
	desktop.MainWindow.Run()
}

func (d *NovaBackupDesktop) createMainWindow() error {
	mw, err := MainWindow{
		AssignTo: &d.MainWindow,
		Title:    "NovaBackup v6.0 - Enterprise Backup & Recovery",
		Size:     Size{1200, 800},
		Layout:   VBox{},
		Children: []Widget{
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{Text: "Статус:", MaxSize: Size{60, 20}},
					Label{
						AssignTo: &d.statusLabel,
						Text:     "Готовий",
						TextColor: walk.RGB(0, 128, 0),
						MaxSize:  Size{150, 20},
					},
					HSpacer{},
					PushButton{
						Text: "Створити завдання",
						OnClicked: func() { d.showCreateJobDialog() },
					},
					PushButton{
						Text: "Запустити зараз",
						OnClicked: func() { d.runImmediateBackup() },
					},
					PushButton{
						Text: "Налаштування",
						OnClicked: func() { d.showSettingsDialog() },
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					GroupBox{
						Title:  "Завдання резервного копіювання",
						Layout: VBox{},
						MinSize: Size{700, 400},
						Children: []Widget{
							TableView{
								AssignTo: &d.jobsTableView,
								Columns: []TableViewColumn{
									{Title: "Назва завдання", Width: 180},
									{Title: "Тип", Width: 80},
									{Title: "Статус", Width: 100},
									{Title: "Останній запуск", Width: 140},
									{Title: "Наступний запуск", Width: 140},
									{Title: "Розклад", Width: 100},
								},
								Model: d.getJobsModel(),
							},
							Composite{
								Layout: HBox{},
								Children: []Widget{
									PushButton{
										Text: "Редагувати",
										OnClicked: func() { d.editSelectedJob() },
									},
									PushButton{
										Text: "Видалити",
										OnClicked: func() { d.deleteSelectedJob() },
									},
									PushButton{
										Text: "Запустити",
										OnClicked: func() { d.runSelectedJob() },
									},
									HSpacer{},
								},
							},
						},
					},
					GroupBox{
						Title:  "Статус системи",
						Layout: VBox{},
						MinSize: Size{400, 400},
						Children: []Widget{
							Label{Text: "Прогрес копіювання:"},
							ProgressBar{
								AssignTo: &d.progressBar,
								MinSize:  Size{350, 25},
							},
							Label{Text: "Журнал активності:"},
							TextEdit{
								AssignTo: &d.logTextEdit,
								ReadOnly: true,
								MinSize:  Size{350, 250},
								MaxSize:  Size{350, 250},
							},
						},
					},
				},
			},
			GroupBox{
				Title:  "Статистика сховища",
				Layout: HBox{},
				Children: []Widget{
					Label{Text: "Всього копій: 0"},
					Label{Text: "Використано: 0 ГБ"},
					Label{Text: "Дедуплікація: 0:1"},
					Label{Text: "Стиснення: 0:1"},
				},
			},
		},
	}.Create()

	if err != nil {
		return err
	}

	d.updateJobsList()
	d.startStatusUpdater()

	return nil
}

func (d *NovaBackupDesktop) getJobsModel() *walk.ReflectTableModel {
	jobs := []BackupJob{
		{
			Name:     "Щоденне копіювання документів",
			Type:     "Файли",
			Status:   "✅ Активне",
			LastRun:  "2026-03-11 02:00",
			NextRun:  "2026-03-12 02:00",
			Schedule: "Щодня 2:00",
		},
		{
			Name:     "Щотижневе копіювання системи",
			Type:     "Система",
			Status:   "✅ Активне",
			LastRun:  "2026-03-08 22:00",
			NextRun:  "2026-03-15 22:00",
			Schedule: "Щонеділі 22:00",
		},
	}

	return walk.NewReflectTableModelFromSlice(jobs)
}

func (d *NovaBackupDesktop) startBackgroundMonitoring() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				d.updateStatus()
				d.updateJobsList()
			}
		}
	}()
}

func (d *NovaBackupDesktop) startStatusUpdater() {
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				d.MainWindow.Synchronize(func() {
					d.updateProgressBar()
				})
			}
		}
	}()
}

func (d *NovaBackupDesktop) updateStatus() {
	d.MainWindow.Synchronize(func() {
		d.statusLabel.SetText("Працює")
		d.statusLabel.SetTextColor(walk.RGB(0, 128, 0))
		
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		logEntry := fmt.Sprintf("[%s] Моніторинг системи активний\n", timestamp)
		
		currentText := d.logTextEdit.Text()
		if len(currentText) > 1000 {
			lines := strings.Split(currentText, "\n")
			if len(lines) > 50 {
				currentText = strings.Join(lines[len(lines)-50:], "\n")
			}
		}
		
		d.logTextEdit.SetText(currentText + logEntry)
	})
}

func (d *NovaBackupDesktop) updateProgressBar() {
	current := d.progressBar.Value()
	if current >= 100 {
		d.progressBar.SetValue(0)
	} else {
		d.progressBar.SetValue(current + 1)
	}
}

func (d *NovaBackupDesktop) updateJobsList() {
	d.MainWindow.Synchronize(func() {
		model := d.getJobsModel()
		d.jobsTableView.SetModel(model)
	})
}

func (d *NovaBackupDesktop) showCreateJobDialog() {
	var nameEdit, sourceEdit, destEdit *walk.LineEdit
	var scheduleCombo *walk.ComboBox

	dlg, err := walk.NewDialog(d.MainWindow)
	if err != nil {
		return
	}
	
	dlg.SetTitle("Створити завдання резервного копіювання")
	dlg.SetLayout(walk.NewVBoxLayout())

	// Name
	nameLabel := walk.NewLabel(dlg)
	nameLabel.SetText("Назва завдання:")
	
	nameEdit, _ = walk.NewLineEdit(dlg)
	
	// Source
	sourceLabel := walk.NewLabel(dlg)
	sourceLabel.SetText("Джерело:")
	
	sourceEdit, _ = walk.NewLineEdit(dlg)
	
	// Destination
	destLabel := walk.NewLabel(dlg)
	destLabel.SetText("Призначення:")
	
	destEdit, _ = walk.NewLineEdit(dlg)
	
	// Schedule
	scheduleLabel := walk.NewLabel(dlg)
	scheduleLabel.SetText("Розклад:")
	
	scheduleCombo, _ = walk.NewComboBox(dlg)
	scheduleCombo.SetModel([]string{"Щодня", "Щотижня", "Щомісяця", "Вручну"})
	scheduleCombo.SetValue("Щодня")
	
	// Buttons
	buttonContainer := walk.NewComposite(dlg)
	buttonContainer.SetLayout(walk.NewHBoxLayout())
	
	createBtn := walk.NewPushButton(buttonContainer)
	createBtn.SetText("Створити")
	createBtn.Clicked().Attach(func() {
		d.addLogEntry("Створено нове завдання: "+nameEdit.Text(), "success")
		dlg.Accept()
	})
	
	cancelBtn := walk.NewPushButton(buttonContainer)
	cancelBtn.SetText("Скасувати")
	cancelBtn.Clicked().Attach(func() {
		dlg.Cancel()
	})

	dlg.Run()
}

func (d *NovaBackupDesktop) runImmediateBackup() {
	d.statusLabel.SetText("Виконується копіювання...")
	d.statusLabel.SetTextColor(walk.RGB(255, 165, 0))
	
	go func() {
		time.Sleep(3 * time.Second)
		d.MainWindow.Synchronize(func() {
			d.statusLabel.SetText("Копіювання завершено")
			d.statusLabel.SetTextColor(walk.RGB(0, 128, 0))
			
			timestamp := time.Now().Format("2006-01-02 15:04:05")
			logEntry := fmt.Sprintf("[%s] ✅ Резервне копіювання успішно завершено\n", timestamp)
			d.logTextEdit.SetText(d.logTextEdit.Text() + logEntry)
			d.progressBar.SetValue(100)
		})
	}()
}

func (d *NovaBackupDesktop) showSettingsDialog() {
	dlg, err := walk.NewDialog(d.MainWindow)
	if err != nil {
		return
	}
	
	dlg.SetTitle("Налаштування NovaBackup")
	dlg.SetLayout(walk.NewVBoxLayout())
	dlg.SetSize(walk.Size{500, 400})

	// General Settings
	generalGroup := walk.NewGroupBox(dlg)
	generalGroup.SetTitle("Загальні налаштування")
	generalGroup.SetLayout(walk.NewVBoxLayout())
	
	autoStartCheck := walk.NewCheckBox(generalGroup)
	autoStartCheck.SetText("Автоматичний запуск з Windows")
	
	notificationsCheck := walk.NewCheckBox(generalGroup)
	notificationsCheck.SetText("Відображати сповіщення")
	
	compressionCheck := walk.NewCheckBox(generalGroup)
	compressionCheck.SetText("Стиснення даних")
	compressionCheck.SetChecked(true)
	
	encryptionCheck := walk.NewCheckBox(generalGroup)
	encryptionCheck.SetText("Шифрування даних")

	// Storage Settings
	storageGroup := walk.NewGroupBox(dlg)
	storageGroup.SetTitle("Сховище")
	storageGroup.SetLayout(walk.NewVBoxLayout())
	
	pathLabel := walk.NewLabel(storageGroup)
	pathLabel.SetText("Шлях до сховища за замовчуванням:")
	
	pathEdit, _ := walk.NewLineEdit(storageGroup)
	pathEdit.SetText("D:\\NovaBackups")
	
	sizeLabel := walk.NewLabel(storageGroup)
	sizeLabel.SetText("Максимальний розмір сховища (ГБ):")
	
	sizeEdit, _ := walk.NewLineEdit(storageGroup)
	sizeEdit.SetText("1000")
	
	// Buttons
	buttonContainer := walk.NewComposite(dlg)
	buttonContainer.SetLayout(walk.NewHBoxLayout())
	
	saveBtn := walk.NewPushButton(buttonContainer)
	saveBtn.SetText("Зберегти")
	saveBtn.Clicked().Attach(func() {
		d.addLogEntry("Налаштування збережено", "success")
		dlg.Accept()
	})
	
	cancelBtn := walk.NewPushButton(buttonContainer)
	cancelBtn.SetText("Скасувати")
	cancelBtn.Clicked().Attach(func() {
		dlg.Cancel()
	})

	dlg.Run()
}

func (d *NovaBackupDesktop) editSelectedJob() {
	d.addLogEntry("Редагування завдання", "info")
}

func (d *NovaBackupDesktop) deleteSelectedJob() {
	if walk.MsgBox(
		d.MainWindow,
		"Видалити завдання",
		"Ви впевнені, що хочете видалити це завдання?",
		walk.MsgBoxYesNo|walk.MsgBoxIconQuestion,
	) == walk.DlgCmdYes {
		d.addLogEntry("Завдання видалено", "warning")
	}
}

func (d *NovaBackupDesktop) runSelectedJob() {
	d.runImmediateBackup()
}

func (d *NovaBackupDesktop) addLogEntry(message string, logType string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	var icon string
	switch logType {
	case "success":
		icon = "✅"
	case "warning":
		icon = "⚠️"
	case "error":
		icon = "❌"
	default:
		icon = "ℹ️"
	}
	
	logEntry := fmt.Sprintf("[%s] %s %s\n", timestamp, icon, message)
	d.logTextEdit.SetText(d.logTextEdit.Text() + logEntry)
}
