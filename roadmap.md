# 🚀 NovaBackup v7.0 - Roadmap створення повної копії Veeam Backup & Replication

Цей документ описує стратегічний шлях перетворення поточної бази коду (Go backend + PowerShell/C#) у повноцінний клон класичного **Veeam Backup & Replication**. Враховано поточний стан: наявність часткової інтеграції VMware (govmomi) та Hyper-V (WMI), дедуплікації, стиснення, API та бази SQLite.

---

## 🏗 Кросаналіз Архітектури: Поточний стан vs Цільовий Veeam

### Veeam Architecture Components:
1. **Backup Server** (Configuration Database, Management Service)
2. **Backup Proxy** (Data Movers для VMware, Hyper-V, physical)
3. **Backup Repository** (Storage targets)
4. **Guest Interaction Proxy** (Для VSS та Application-Aware processing)
5. **Mount Server** (Для FLR - File-Level Recovery)
6. **Veeam Console** (Товстий клієнт WPF/C#)

### Наша поточна Архітектура (NovaBackup):
- **Backup Server & Repository**: Об'єднані в `nova-service.exe` / `nova-cli.exe` (Go). База даних - SQLite (`novabackup.db`).
- **Backup Proxies**: Здебільшого відсутні як окремі сутності, виконуються локально в моноліті.
- **Console**: Зачатки в `desktop/wpf/`.

---

## 🎯 СТРАТЕГІЧНІ ФАЗИ РОЗРОБКИ (ROADMAP)

### ФАЗА 1: Класична Veeam Console (WPF/C#)
**Мета**: Надати користувачеві 100% візуальний досвід Veeam. Користувач не повинен бачити CLI.

**Спринти**:
1. **Console Framework**: 
   - Розвинути поточний `desktop/wpf` у повноцінний MVVM (через Prism або CommunityToolkit.Mvvm).
   - Головне вікно: Рибон (Ribbon), класичне зелено-синє меню Veeam, статусний бар знизу.
   - Навігаційне дерево зліва: *Home, Inventory, Infrastructure, Tape Infrastructure, History*.
2. **Integration API**: 
   - Console має спілкуватися з Go-сервісом виключно через gRPC або REST API (наш повноцінний `pkg/api/`).
   - Відмова від викликів PowerShell з гуя.
3. **Wizards (Майстри)**:
   - *Backup Job Wizard* (Name -> Virtual Machines -> Storage -> Guest Processing -> Schedule -> Summary).
   - [x] *Restore Wizard* (Instant VM Recovery, Entire VM Restore, Guest files).

### ФАЗА 2: Enterprise Data Movers (Proxy Architecture)
**Мета**: Відокремити логіку читання з гіпервізорів від Backup Server для досягнення швидкості Veeam (Transport Services).

**Спринти**:
1. **RPC Framework**: 
   - Створення легковагого агента на Go (`nova-datamover.exe`), який може розгортатися на віддалених серверах Windows/Linux.
2. **VMware VADP (vStorage APIs for Data Protection)**:
   - Розширення нашої поточної govmomi інтеграції для використання *HotAdd* (підключення дисків ВМ до проксі-ВМ) та *NBD* (Network Block Device) режимів.
3. **Hyper-V Data Mover**:
   - Стабілізація WMI інтеграції (котру ми почали). Створення агента, який читає VHDX/RCT безпосередньо з CSV (Cluster Shared Volume) або локального диска хоста.

### ФАЗА 3: Guest Processing & Application-Aware Backup
**Мета**: Консистентні бекапи баз даних (SQL, Exchange, AD).

**Спринти**:
1. **Guest Interaction Proxy (GIP)**:
   - Нативний ін'єкт крихітного агента (через RPC/WMI + Admin share `C$\admin$`) всередину гостьової ОС (як робить Veeam VSS Requestor).
2. **Microsoft VSS Integration**:
   - Розширення поточних напрацювань `pkg/providers/vss`.
   - Freeze/Thaw скрипти для Linux (pre-freeze.sh, post-thaw.sh).
3. **Log Truncation**:
   - Усічення логів транзакцій SQL Server та Exchange після успішного бекапу.

### ФАЗА 4: Instant VM Recovery & Advanced Restore (vPower NFS)
**Мета**: Флагманська фіча Veeam — запуск ВМ прямо з бекап-файлу за лічені секунди.

**Спринти**:
1. **[x] Нативний NFS v3 Сервер (Go)**:
   - Реалізовано в `internal/recovery/instant_manager.go` та `pkg/providers/instantrecovery/nfs.go`.
   - Реєстрація нашого сховища дедуплікованих блоків як датастору NFS у VMware vSphere.
2. **[x] Synthetic Disk Presentation**:
   - Трансляція бекап-чанків NovaBackup у сиквенційний потік `.vhdx` / `.vmdk` на льоту через `ChunkVFS`.
3. **Storage vMotion Trigger**: 
   - Автоматичний перенос ВМ з нашого NFS на продуктивний датастор в фоновому режимі після Instant Recovery.

### ФАЗА 5: Backup Repository & Storage Integration
**Мета**: Розширити варіанти зберігання даних на рівень Enterprise.

**Спринти**:
1. **[/] Scale-Out Backup Repository (SOBR)**:
   - Об'єднання кількох локальних папок/серверів в єдиний пул (Реалізовано `RepositoryPool`).
   - Tiering: локальні диски (Performance) + S3 (Capacity - Реалізовано `S3Engine`).
2. **Immutable Backups (Hardened Repository)**:
   - Захист від Ransomware.
   - Використання Linux `chattr +i` (XFS Fast Clone / reflink) для Windows-заснованої системи (розгортання Linux-агента).
3. **Storage Snapshots**:
   - Інтеграція з SAN-протоколами (NetApp, HPE 3PAR/Nimble) для зменшення навантаження на гіпервізор (читання прямо зі снепшота LUN).

---

## 🛠 Технологічний стек
1. **Backend**: Go (GoLang). Чому? Низьке споживання пам'яті, швидка криптографія, кросплатформна паралельність.
2. **Frontend (Console)**: C# .NET 8 / WPF. Чому? Тому що Veeam console написана на .NET/WPF, і це єдиний спосіб отримати **ідентичний** Enterprise Windows досвід. Знаходитиметься у папці `desktop/wpf`.
3. **Database**: Заміна SQLite на **PostgreSQL** (як головної бази) з часом, як і Veeam перейшов з SQL Express на PostgreSQL у версії 12. Для малих установок - SQLite залишається.
4. **Data format**: Власний дедуплікований формат (.nvb = аналог .vbk).

---

## 📅 Графік виконання (Приблизний)
1. **ФУНДАМЕНТ (зараз - 1 місяць)**: 
   - Завершення WMI Hyper-V, закріплення API.
2. **ВІЗУАЛІЗАЦІЯ (1-2 місяці)**:
   - Побудова WPF консолі (зв'язка з Go Service). Це дасть продукт, який "виглядає" як Veeam.
3. **ПРОСУНУТІ ФІЧІ (2-4 місяці)**:
   - vPower NFS (Instant Recovery), Application-Aware VSS.
4. **ENTERPRISE SCALE (4-6 місяців)**:
   - SOBR (Scale-out), S3 Tiering, Hardware Snapshots.

---
*Документ створено автоматично на основі аналізу поточного коду агентом NovaBackup AI.*
