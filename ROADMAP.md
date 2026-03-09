# 🚀 NovaBackup v6.0 — Roadmap

## 📋 Огляд

**NovaBackup v6.0** — повна заміна **Veeam Backup & Replication** з використанням **Go** (backend) та **Python** (GUI + AI).

### 🎯 Ціль: Повний аналог Veeam

| Функція Veeam | NovaBackup v6.0 | Статус |
|---------------|-----------------|--------|
| **VM Backup (VMware/Hyper-V)** | ✅ Повна підтримка | Plan |
| **Agent Backup (Windows/Linux)** | ✅ Агенти для ОС | Plan |
| **Backup Copy Jobs** | ✅ Реплікація між сховищами | Plan |
| **Replication** | ✅ Реплікація VM | Plan |
| **Instant VM Recovery** | ✅ Запуск VM з backup | Plan |
| **SureBackup** | ✅ Авто-перевірка backup | Plan |
| **Deduplication** | ✅ Глобальна дедуплікація | Plan |
| **Compression** | ✅ Стиснення даних | Plan |
| **Encryption** | ✅ AES-256 шифрування | Plan |
| **WAN Acceleration** | ✅ Оптимізація WAN | Plan |
| **Cloud Connect** | ✅ Інтеграція з хмарою | Plan |
| **Monitoring & Alerts** | ✅ Сповіщення | Plan |
| **Reporting** | ✅ Звіти та аналітика | Plan |

### Технічний стек

| Компонент | Технологія | Призначення |
|-----------|------------|-------------|
| **Backend** | Go 1.22+ | API, backup engine, scheduler |
| **GUI Desktop** | Python + PyQt6 / CustomTkinter | Користувацький інтерфейс |
| **Web UI** | React + TypeScript + Recharts | Веб-інтерфейс |
| **AI Analytics** | Python + scikit-learn + pandas | ML прогнози |
| **Database** | SQLite + PostgreSQL | Локальне + серверне зберігання |
| **Message Queue** | NATS / Redis | Асинхронні задачі |

---

## 🖥️ Інтерфейс (UI/UX)

### 📌 Головне вікно (Dashboard)

```
┌─────────────────────────────────────────────────────────────────┐
│  ☰  NovaBackup v6.0                    🔔 3  ⚙️  👤 Admin  ❌  │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  📊 OVERVIEW                                                    │
│  ┌─────────────┬─────────────┬─────────────┬─────────────┐    │
│  │ 📦 Jobs     │ ✅ Success  │ ⚠️ Warnings │ ❌ Failed   │    │
│  │    24       │    18       │     4       │     2       │    │
│  └─────────────┴─────────────┴─────────────┴─────────────┘    │
│                                                                 │
│  📈 BACKUP STATUS (24h)                                        │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │  [████████████████░░░░░░░░░░████████] 68% Complete     │  │
│  │   00:00    04:00    08:00   12:00   16:00    20:00     │  │
│  └─────────────────────────────────────────────────────────┘  │
│                                                                 │
│  🖥️ INFRASTRUCTURE                                             │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │ VMware vCenter  │  Hyper-V  │  Agents  │  Cloud        │  │
│  │ ● cluster-01    │  ● hv-01  │  ● 12    │  ● AWS S3     │  │
│  │ ● cluster-02    │  ● hv-02  │  ● 8     │  ● Azure Blob │  │
│  └─────────────────────────────────────────────────────────┘  │
│                                                                 │
│  📋 RECENT JOBS                                                 │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │ Job Name          │ Status   │ Progress │ Next Run     │  │
│  ├───────────────────┼──────────┼──────────┼──────────────┤  │
│  │ Daily Backup      │ ✅ Done  │ 100%     │ Tomorrow 02:00│ │
│  │ DB Backup         │ ⚠️ Warn  │ 100%     │ Today 23:00  │  │
│  │ VM Replication    │ 🔄 Run   │ 68%      │ —            │  │
│  │ Agent Backup #5   │ ❌ Fail  │ 45%      │ Retry in 1h  │  │
│  └─────────────────────────────────────────────────────────┘  │
│                                                                 │
├─────────────────────────────────────────────────────────────────┤
│  🏠 Home  📦 Jobs  🖥️ Infrastructure  📊 Reports  ⚙️ Settings  │
└─────────────────────────────────────────────────────────────────┘
```

### 📦 Створення Job (Wizard)

```
┌─────────────────────────────────────────────────────────┐
│  New Backup Job                              ❌         │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  Step 1 of 4: Name & Description                        │
│  ┌───────────────────────────────────────────────────┐ │
│  │ Name:        [Daily VM Backup____________]        │ │
│  │ Description: [Backup all production VMs___]        │ │
│  │                                                    │ │
│  │ [◀ Back]              [Next ▶]                    │ │
│  └───────────────────────────────────────────────────┘ │
│                                                         │
│  Step 2 of 4: Virtual Machines                          │
│  ┌───────────────────────────────────────────────────┐ │
│  │ Add VMs:                                           │ │
│  │ ☐ vm-web-01 (192.168.1.10)                        │ │
│  │ ☐ vm-web-02 (192.168.1.11)                        │ │
│  │ ☑ vm-db-01  (192.168.1.20)                        │ │
│  │ ☑ vm-db-02  (192.168.1.21)                        │ │
│  │ ☐ vm-app-01 (192.168.1.30)                        │ │
│  │                                                    │ │
│  │ Selected: 2 VMs (Total: 500 GB)                   │ │
│  │ [◀ Back]              [Next ▶]                    │ │
│  └───────────────────────────────────────────────────┘ │
│                                                         │
│  Step 3 of 4: Storage & Schedule                        │
│  ┌───────────────────────────────────────────────────┐ │
│  │ Backup Repository: [Backup-Repo-01 ▼]             │ │
│  │ Retention:         [7] days                       │ │
│  │ Enable deduplication: ☑                           │ │
│  │ Enable compression:  ☑ High                       │ │
│  │ Enable encryption:   ☑                             │ │
│  │                                                    │ │
│  │ Schedule:                                          │ │
│  │ ☑ Daily at: [02:00 ▼]                             │ │
│  │ ☐ Weekly on: [ ] Mon [ ] Tue [ ] Wed...           │ │
│  │ ☐ Monthly on: [1st ▼] [Monday ▼]                  │ │
│  │                                                    │ │
│  │ [◀ Back]              [Next ▶]                    │ │
│  └───────────────────────────────────────────────────┘ │
│                                                         │
│  Step 4 of 4: Advanced Options                          │
│  ┌───────────────────────────────────────────────────┐ │
│  │ Application-aware processing: ☑                    │ │
│  │ Guest processing: ☐                               │ │
│  │ Backup storage: ☑ Reverse incremental             │ │
│  │ Storage optimization: ☑ LAN                       │ │
│  │                                                    │ │
│  │ Notifications:                                     │ │
│  │ ☑ Send email on: ☑ Success ☑ Warning ☑ Failure   │ │
│  │ Email: [admin@company.com________________]         │ │
│  │                                                    │ │
│  │ [◀ Back]              [Finish ✓]                  │ │
│  └───────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

### 🎨 Дизайн-система

#### Кольорова схема:
```css
/* Dark Theme (Default) */
--bg-primary: #1a1a2e;
--bg-secondary: #16213e;
--bg-tertiary: #0f3460;
--accent: #e94560;
--success: #00d26a;
--warning: #ffc107;
--error: #dc3545;
--text-primary: #ffffff;
--text-secondary: #a0a0a0;

/* Light Theme */
--bg-primary: #ffffff;
--bg-secondary: #f5f5f5;
--bg-tertiary: #e0e0e0;
--accent: #d63031;
--success: #00b894;
--warning: #fdcb6e;
--error: #d63031;
--text-primary: #2d3436;
--text-secondary: #636e72;
```

#### Компоненти:
- **Картки** — з тінню, rounded corners (8px)
- **Таблиці** — sortable, filterable, pagination
- **Графіки** — Recharts / Chart.js
- **Сповіщення** — toast notifications
- **Progress bars** — з відсотками та часом
- **Дерева** — для інфраструктури (VMware, Hyper-V)

---

## 📅 Етапи розробки

### **Етап 1: Go Core (2 тижні)**

**Мета:** Базовий рушій резервного копіювання

#### Завдання:
- [ ] Ініціалізація Go модуля
- [ ] Структура проєкту (cmd/, internal/, pkg/)
- [ ] Базовий backup engine
- [ ] File provider (резервне копіювання файлів)
- [ ] SQLite інтеграція
- [ ] CLI утиліта

#### Файли:
```
cmd/
  nova-cli/main.go          # CLI точка входу
internal/
  backup/
    engine.go               # BackupEngine
    restore.go              # RestoreEngine
  providers/
    file.go                 # FileBackupProvider
  database/
    sqlite.go               # SQLite connection
pkg/
  models/
    models.go               # Structs: Job, Result, Config
```

#### Результат:
```bash
./nova-cli backup --source /data --dest /backup
./nova-cli restore --backup /backup/2026-03-08 --dest /restore
```

---

### **Етап 2: Go API (1 тиждень)**

**Мета:** REST API для керування backup

#### Завдання:
- [ ] Gin/Echo framework
- [ ] REST endpoints (jobs, backups, restore)
- [ ] Swagger документація
- [ ] JWT authentication
- [ ] Middleware (logging, CORS, rate limiting)

#### API Endpoints:
```
POST   /api/v1/jobs          # Створити job
GET    /api/v1/jobs          # Список jobs
POST   /api/v1/jobs/:id/run  # Запустити job
DELETE /api/v1/jobs/:id      # Видалити job
GET    /api/v1/backups       # Список backups
POST   /api/v1/restore       # Відновлення
GET    /swagger              # Swagger UI
```

#### Результат:
```bash
./nova-api --port 8080
curl http://localhost:8080/api/v1/jobs
```

---

### **Етап 3: Python GUI Basic (2 тижні)**

**Мета:** PyQt6 додаток для керування

#### Завдання:
- [ ] PyQt6 setup
- [ ] Головне вікно (список jobs)
- [ ] Діалог створення job
- [ ] API client (httpx)
- [ ] Progress bar для backup

#### Файли:
```
python/
  gui/
    main.py                 # Точка входу
    main_window.py          # MainWindow class
    widgets.py              # Custom widgets
    styles.py               # QSS styles
  api_wrapper/
    client.py               # APIClient class
```

#### Результат:
```bash
python python/gui/main.py
```

---

### **Етап 4: Database Providers (1 тиждень)**

**Мета:** Backup баз даних

#### Завдання:
- [x] MySQL provider (mysqldump)
- [x] PostgreSQL provider (pg_dump)
- [x] SQLite provider (VACUUM INTO)
- [x] SQL Server provider (stub)
- [x] Compression (gzip)
- [x] Encryption (AES-256)
- [x] Database provider interface
- [x] Database backup service

#### Файли:
```
internal/providers/
  database.go            # Provider interface
  mysql.go               # MySQL provider
  postgresql.go          # PostgreSQL provider
  sqlite.go              # SQLite provider
  mssql.go               # SQL Server provider (stub)
internal/backup/
  database_backup.go     # Database backup service
pkg/models/
  models.go              # DatabaseConfig struct
```

#### Приклад:
```go
// MySQL backup
provider := providers.NewMySQLProvider()
config := &providers.DatabaseConfig{
    Type: providers.DatabaseTypeMySQL,
    Host: "localhost", Port: 3306,
    Username: "user", Password: "pass",
    Database: "mydb",
}
err := provider.Backup(ctx, config, "/backup/mydb.sql")

// PostgreSQL backup
provider := providers.NewPostgreSQLProvider()
err := provider.Backup(ctx, config, "/backup/mydb.sql")

// SQLite backup
provider := providers.NewSQLiteProvider()
err := provider.Backup(ctx, config, "/backup/sqlite.db")
```

#### Результат:
```bash
# CLI database backup
nova-cli db-backup --type mysql --host localhost --database mydb --dest /backup
nova-cli db-backup --type postgresql --host localhost --database mydb --dest /backup
nova-cli db-backup --type sqlite --source /data/app.db --dest /backup
```

---

### **Етап 5: Scheduler (1 тиждень)**

**Мета:** Автоматичне виконання за розкладом

#### Завдання:
- [x] gocron integration
- [x] Розклади (щодня, щотижня, щомісяця)
- [x] Windows Service
- [x] Linux systemd service
- [x] Логіка retry при помилках

#### Файли:
```
internal/backup/
  scheduler.go           # Scheduler with cron integration
cmd/nova-cli/
  main.go                # Scheduler commands
docs/
  windows-service.md     # Windows service installation
  linux-systemd-service.md  # Linux systemd service
  novabackup-scheduler.service  # systemd service template
```

#### Приклад:
```go
// internal/backup/scheduler.go
func (s *Scheduler) Start() {
    // Load jobs from database
    jobs, _ := s.db.GetAllJobs()
    
    for _, job := range jobs {
        if job.Enabled {
            s.scheduleJob(job)
        }
    }
    
    s.cron.Start()
}

// Schedule examples:
// - Daily at 02:00:  "0 0 2 * * *"
// - Weekly Monday 03:00: "0 0 3 * * 1"
// - Monthly 1st at 04:00: "0 0 4 1 * *"
```

#### CLI Commands:
```bash
# Start scheduler
nova-cli scheduler start

# Check status
nova-cli scheduler status

# Run job immediately
nova-cli scheduler run-now --id <job-id>

# Create scheduled job
nova-cli job create \
  --name "Daily Backup" \
  --source /data \
  --destination /backup \
  --schedule daily \
  --time 02:00
```

#### Retry Configuration:
```go
// Default retry settings
RetryConfig:
  MaxRetries: 3
  RetryDelay: 5 minutes
  BackoffFactor: 2.0 (exponential backoff)
```

#### Результат:
```bash
# Start scheduler (runs until Ctrl+C)
nova-cli scheduler start

# Install as Windows Service
# See docs/windows-service.md

# Install as Linux systemd Service
# See docs/linux-systemd-service.md
```

---

### **Етап 6: Virtualization (2 тижні)**

**Мета:** Backup віртуальних машин

#### Завдання:
- [x] VMware vCenter API (govmomi)
- [x] Hyper-V WMI (PowerShell)
- [x] Proxmox API (опціонально)
- [x] Snapshot management
- [x] Changed Block Tracking (CBT) support

#### Файли:
```
internal/providers/
  vmware.go              # VMware vSphere provider
  hyperv.go              # Hyper-V provider (Windows only)
internal/backup/
  vm_backup.go           # VM backup service
cmd/nova-cli/
  main.go                # VM commands
docs/
  vm-backup-guide.md     # VM backup documentation
```

#### VMware Commands:
```bash
# List VMs
nova-cli vm list-vmware \
  --vcenter https://vcenter.company.com \
  --username admin@vsphere.local \
  --password "password"

# Backup VM
nova-cli vm backup \
  --type vmware \
  --name "vm-db-01" \
  --destination /backups/vmware \
  --vcenter https://vcenter.company.com \
  --username admin@vsphere.local \
  --password "password"
```

#### Hyper-V Commands:
```bash
# List VMs
nova-cli vm list-hyperv --host localhost

# Backup VM
nova-cli vm backup \
  --type hyperv \
  --name "vm-db-01" \
  --destination D:\backups\hyperv \
  --host localhost
```

#### VM Backup Service:
```go
vmService := backup.NewVMBackupService()

// VMware backup
config := &backup.VMBackupConfig{
    SourceType:     models.SourceTypeVMware,
    VMName:         "vm-db-01",
    Destination:    "/backups/vmware",
    VCenterURL:     "https://vcenter.company.com",
    Username:       "admin@vsphere.local",
    Password:       "password",
    InsecureVerify: true,
}
result, err := vmService.PerformVMBackup(ctx, config)

// Hyper-V backup
config := &backup.VMBackupConfig{
    SourceType:  models.SourceTypeHyperV,
    VMName:      "vm-db-01",
    Destination: "D:\\backups\\hyperv",
    HyperVHost:  "localhost",
}
result, err := vmService.PerformVMBackup(ctx, config)
```

#### Snapshot Management:
```go
// Create snapshot
snapshotRef, err := vmwareProvider.CreateSnapshot(ctx, vmName, "NovaBackup_20260309", false)

// Remove snapshot
err := vmwareProvider.RemoveSnapshot(ctx, vmName, snapshotRef, true)

// List snapshots
snapshots, err := vmwareProvider.ListSnapshots(ctx, vmName)
```

#### CBT Support:
```go
// Check if CBT is supported
cbtSupported, err := vmwareProvider.IsCBTSupported(ctx, vmName)

// Enable CBT
err := vmwareProvider.EnableCBT(ctx, vmName)

// Get changed blocks (for incremental backup)
changedBlocks, err := vmwareProvider.GetChangedBlocks(ctx, vmName, snapshotRef)
```

#### Результат:
```bash
# VM backup completed
✅ VM backup completed!
  Status:  success
  Path:    /backups/vmware/vm-db-01_backup
  Time:    45s
  Size:    50 GB
```

---

### **Етап 7: AI Analytics (2 тижні)**

**Мета:** ML прогнози для оптимізації

#### Завдання:
- [ ] Збір статистики backup
- [ ] ML модель (failure prediction)
- [ ] Оптимізація розкладу
- [ ] Рекомендації щодо зберігання

#### Файли:
```
python/
  ai_analytics/
    predictor.py            # BackupPredictor
    optimizer.py            # ScheduleOptimizer
    models/
      failure_model.pkl     # Trained model
```

---

### **Етап 8: Web UI (1 тиждень)**

**Мета:** React веб-інтерфейс

#### Завдання:
- [ ] React + Vite setup
- [ ] Material UI
- [ ] Dashboard з графіками
- [ ] WebSocket для realtime
- [ ] Responsive design

---

### **Етап 9: Build + Release (1 тиждень)**

**Мета:** Готові білди для всіх платформ

#### Завдання:
- [ ] Go cross-compile (Windows, Linux, macOS)
- [ ] Python PyInstaller
- [ ] Docker image
- [ ] MSI інсталятор (WiX)
- [ ] DEB/RPM пакети

#### Команди:
```bash
# Go
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -o nova-api.exe ./cmd/nova-api
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o nova-api ./cmd/nova-api

# Python
pyinstaller python/gui/main.spec

# Docker
docker build -t novabackup/api:latest .
```

---

## 📊 Timeline

| Етап | Тривалість | Дедлайн | Статус |
|------|------------|---------|--------|
| 1. Go Core | 2 тижні | 22.03.2026 | ✅ Complete |
| 2. Go API | 1 тиждень | 29.03.2026 | ✅ Complete |
| 3. Python GUI | 2 тижні | 12.04.2026 | ✅ Complete |
| 4. DB Providers | 1 тиждень | 19.04.2026 | ✅ Complete |
| 5. Scheduler | 1 тиждень | 26.04.2026 | ✅ Complete |
| 6. Virtualization | 2 тижні | 10.05.2026 | ✅ Complete |
| 7. AI Analytics | 2 тижні | 24.05.2026 | ⬜ Pending |
| 8. Web UI | 1 тиждень | 31.05.2026 | ⬜ Pending |
| 9. Build | 1 тиждень | 07.06.2026 | ⬜ Pending |

**Всього:** 13 тижнів (~3 місяці)

**Прогрес:** 6/9 етапів завершено (67%)

---

## ✅ Критерії готовності

### v6.0.0 Alpha (Етап 3)
- [x] Go backup engine працює
- [x] CLI утиліта функціональна
- [x] API доступне
- [x] Python GUI показує список jobs

### v6.0.0 Beta (Етап 6)
- [x] Всі DB providers працюють (MySQL, PostgreSQL, SQLite)
- [x] Scheduler виконує jobs за розкладом
- [x] GUI дозволяє створювати/редагувати jobs
- [x] VMware/Hyper-V integration
- [ ] Логи та моніторинг

### v6.0.0 RC (Етап 8)
- [ ] Web UI функціональний
- [ ] AI analytics працює
- [ ] Всі тести проходять
- [ ] Документація готова

### v6.0.0 Release (Етап 9)
- [ ] Білди для Windows/Linux/macOS
- [ ] Docker image
- [ ] MSI інсталятор
- [ ] Реліз на GitHub

---

## 📝 Notes

- Зберігаємо сумісність з v5.x (міграція баз даних)
- Мінімум залежностей для Go (static binary)
- Python тільки для GUI та AI
- Тестування: go test + pytest
- CI/CD: GitHub Actions

---

## 🔗 Посилання

- [Go Documentation](https://go.dev/doc/)
- [PyQt6 Documentation](https://www.riverbankcomputing.com/static/Docs/PyQt6/)
- [Gin Framework](https://gin-gonic.com/)
- [React Documentation](https://react.dev/)
