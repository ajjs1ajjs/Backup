# 🛡️ NovaBackup Enterprise v7.0

**Повнофункціональна система резервного копіювання як Veeam Backup & Replication**

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.25-blue)](https://golang.org)
[![Ukraine](https://img.shields.io/badge/Made%20in-%F0%9F%87%BA%F0%9F%87%A6-blue)](https://ukraine.ua)

---

## 🚀 Можливості

### ✅ Резервне копіювання
- 📁 **Файли і папки** - з дедуплікацією та стисненням
- 🗄️ **Бази даних** - MySQL, PostgreSQL, SQLite
- 🖥️ **Віртуальні машини** - Hyper-V (експорт VM)
- ☁️ **Хмарні сховища** - S3, Azure Blob, Google Cloud

### ♻️ Відновлення
- 📁 **Відновлення файлів** - вибіркове або повне
- 🗄️ **Відновлення БД** - автоматичне розгортання
- 🖥️ **Відновлення VM** - з можливістю миттєвого запуску
- ⚡ **Instant Recovery** - запуск VM з бекапу

### 💾 Сховища даних
- 📂 **Локальні** - прямий доступ до диску
- 🌐 **SMB/CIFS** - мережеві частки
- ☁️ **S3-compatible** - AWS, MinIO, Wasabi
- 🔷 **Azure Blob** - Microsoft Azure
- 📦 **Google Cloud** - GCS

### 🔐 Безпека
- 🔒 **Шифрування** - AES-256
- 🎭 **Дедуплікація** - блокова на рівні 1MB
- 🗜️ **Стиснення** - gzip з різними рівнями
- 👥 **RBAC** - ролі та дозволи

### ⏰ Планування
- 📅 **Щодня/Щотижня/Щомісяця**
- ⏲️ **Cron вирази**
- 🔄 **Інкрементальні бекапи**
- 📊 **GFS ротація**

### 🔔 Сповіщення
- 📧 **Email** - SMTP з HTML шаблонами
- 💬 **Telegram** - бот сповіщення
- 🔗 **Webhook** - інтеграція з іншими системами
- 💬 **Slack** - канали сповіщень

### 👨‍💼 Користувачі
- 🎭 **Ролі** - Admin, Backup Admin, Backup User, ReadOnly
- 🔑 **Аутентифікація** - паролі з хешуванням
- 📝 **Аудит** - повне логування дій
- 🔒 **Сесії** - 24-годинні токени

---

## 🎨 Веб-інтерфейс

Сучасний темний інтерфейс українською мовою:

- 📊 **Панель керування** - статистика та швидкі дії
- 💾 **Завдання** - створення та управління
- 🔄 **Відновлення** - вибір точок відновлення
- 📁 **Сховища** - конфігурація сховищ
- 📋 **Сесії** - історія операцій
- ⚙️ **Налаштування** - система та сповіщення

---

## 🚀 Швидкий старт

### Windows
```powershell
# Завантажити
wget https://github.com/ajjs1ajjs/Backup/releases/latest/download/novabackup.exe

# Встановити службу
.\novabackup.exe install

# Запустити
.\novabackup.exe start

# Відкрити веб-інтерфейс
http://localhost:8050
```

### Linux
```bash
# Завантажити
wget https://github.com/ajjs1ajjs/Backup/releases/latest/download/novabackup-linux-amd64

# Встановити
chmod +x novabackup-linux-amd64
sudo ./novabackup-linux-amd64 install

# Відкрити веб-інтерфейс
http://localhost:8050
```

**Login:** `admin`  
**Password:** `admin123`

---

## 📋 API Endpoints

### Authentication
```bash
POST /api/auth/login
POST /api/auth/logout
```

### Backup Jobs
```bash
GET    /api/jobs              # Список завдань
POST   /api/jobs              # Створити завдання
PUT    /api/jobs/:id          # Оновити завдання
DELETE /api/jobs/:id          # Видалити завдання
POST   /api/jobs/:id/run      # Запустити завдання
```

### Backup & Restore
```bash
POST /api/backup/run          # Запустити бекап
GET  /api/backup/sessions     # Список сесій
GET  /api/restore/points      # Точки відновлення
GET  /api/restore/files       # Перегляд файлів
POST /api/restore/files       # Відновити файли
POST /api/restore/database    # Відновити БД
```

### Storage
```bash
GET    /api/storage/repos     # Список сховищ
POST   /api/storage/repos     # Додати сховище
DELETE /api/storage/repos/:id # Видалити сховище
```

### Settings & Users
```bash
GET    /api/settings          # Налаштування
PUT    /api/settings          # Оновити налаштування
GET    /api/users             # Користувачі
POST   /api/users             # Створити користувача
GET    /api/audit/logs        # Аудит логи
```

---

## 🔧 Приклад використання API

### Створити завдання резервного копіювання
```bash
curl -X POST http://localhost:8050/api/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Daily Backup",
    "type": "file",
    "sources": ["C:\\Users\\Documents"],
    "destination": "D:\\Backups",
    "compression": true,
    "encryption": true,
    "schedule": "daily",
    "schedule_time": "02:00"
  }'
```

### Запустити резервне копіювання
```bash
curl -X POST http://localhost:8050/api/jobs/:id/run
```

### Відновити файли
```bash
curl -X POST http://localhost:8050/api/restore/files \
  -H "Content-Type: application/json" \
  -d '{
    "backup_path": "D:\\Backups\\Daily Backup\\2026-03-15_020000",
    "destination": "C:\\Restored",
    "files": ["document1.docx", "document2.xlsx"]
  }'
```

---

## 📊 Архітектура

```
┌─────────────────────────────────────────────────────────┐
│                  Web UI (React/HTML)                     │
│                  http://localhost:8050                   │
└────────────────────┬────────────────────────────────────┘
                     │ REST API
┌────────────────────▼────────────────────────────────────┐
│                  Go Web Server (Gin)                     │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐  │
│  │  Backup  │ │ Restore  │ │Scheduler │ │   RBAC   │  │
│  │  Engine  │ │ Engine   │ │          │ │          │  │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘  │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐  │
│  │ Storage  │ │Notifications│ │  Audit  │ │  Users  │  │
│  │Providers │ │           │ │         │ │         │  │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘  │
└────────────────────┬────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────┐
│              SQLite Database + File System               │
└─────────────────────────────────────────────────────────┘
```

---

## 🛠️ Розробка

### Вимоги
- Go 1.25+
- Node.js 18+ (для Web UI)
- Git

### Збірка
```bash
# Встановити залежності
go mod tidy

# Зібрати
go build -o novabackup.exe ./cmd/novabackup/

# Запустити
./novabackup.exe server
```

### Тести
```bash
go test ./...
```

---

## 📁 Структура проекту

```
Backup/
├── cmd/
│   └── novabackup/         # Головний додаток
├── internal/
│   ├── api/                # REST API handlers
│   ├── backup/             # Backup Engine
│   ├── restore/            # Restore Engine
│   ├── database/           # SQLite Database
│   ├── scheduler/          # Job Scheduler
│   ├── storage/            # Storage Providers
│   ├── notifications/      # Email, Telegram, Webhook
│   ├── rbac/               # Users & Roles
│   └── audit/              # Audit Logging
├── web/
│   └── index.html          # Web UI
├── go.mod
└── README.md
```

---

## 🔐 Безпека

### Ролі користувачів
| Роль | Дозволи |
|------|---------|
| **Admin** | Повний доступ до всіх функцій |
| **Backup Admin** | Управління бекапами, без доступу до користувачів |
| **Backup User** | Виконання бекапів, перегляд |
| **ReadOnly** | Тільки перегляд |

### Шифрування
- Паролі: SHA-256 з сіллю
- Бекапи: AES-256 CBC режим
- Сесії: JWT токени

---

## 📈 Порівняння з Veeam

| Функція | Veeam | NovaBackup |
|---------|-------|------------|
| Файлові бекапи | ✅ | ✅ |
| Бази даних | ✅ | ✅ |
| VM бекапи | ✅ | ✅ (Hyper-V) |
| Хмарні сховища | ✅ | ✅ |
| Дедуплікація | ✅ | ✅ |
| Стиснення | ✅ | ✅ |
| Шифрування | ✅ | ✅ |
| Веб-інтерфейс | ✅ | ✅ |
| Українська мова | ❌ | ✅ |
| RBAC | ✅ | ✅ |
| Сповіщення | ✅ | ✅ |
| Безкоштовно | ❌ | ✅ |

---

## 🤝 Внесок

1. Fork репозиторій
2. Створіть feature branch (`git checkout -b feature/amazing-feature`)
3. Commit зміни (`git commit -m 'Add amazing feature'`)
4. Push до branch (`git push origin feature/amazing-feature`)
5. Відкрийте Pull Request

---

## 📄 Ліцензія

MIT License - див. [LICENSE](LICENSE) файл

---

## 📞 Контакти

- 📧 Email: support@novabackup.local
- 💬 Telegram: @novabackup
- 🌐 Website: https://github.com/ajjs1ajjs/Backup
- 📚 Документація: https://github.com/ajjs1ajjs/Backup/wiki

---

## 🙏 Подяки

- Натхненно [Veeam Backup & Replication](https://www.veeam.com)
- Зроблено з ❤️ в Україні 🇺🇦

---

<div align="center">

**NovaBackup Enterprise v7.0**

Сучасна система резервного копіювання українською мовою

[Завантажити](https://github.com/ajjs1ajjs/Backup/releases) • [Документація](https://github.com/ajjs1ajjs/Backup/wiki) • [Звіти](https://github.com/ajjs1ajjs/Backup/issues)

🇺🇦 Зроблено в Україні

</div>
