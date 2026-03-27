# 🌟 NovaBackup

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Python Version](https://img.shields.io/badge/python-3.9%2B-blue)](https://www.python.org/)
[![Version](https://img.shields.io/badge/version-8.0.0-blue.svg)](https://github.com/ajjs1ajjs/Backup/releases)
[![Status](https://img.shields.io/badge/status-production-green)](https://github.com/ajjs1ajjs/Backup)

> **🇺🇦 Українська версія** | [English version](#english-version)

**Production-ready система резервного копіювання на Python з підтримкою хмарних провайдерів та безпекою RBAC**

---

## 📖 Зміст

- [Огляд](#-огляд)
- [Можливості](#-ключові-можливості)
- [Швидкий старт](#-швидкий-старт)
- [Встановлення](#-детальне-встановлення)
- [Конфігурація](#-конфігурація)
- [Використання](#-використання)
- [API Довідник](#-api-довідник)
- [Архітектура](#-архітектура)
- [Хмарні провайдери](#-хмарні-провайдери)
- [Docker](#-docker-розгортання)
- [Тестування](#-тестування)
- [Вирішення проблем](#-вирішення-проблем)
- [Внесок](#-внесок)

---

## 🎯 Огляд

**NovaBackup** — це сучасна модульна система резервного копіювання з повноцінним API, CLI та веб-інтерфейсом для керування бекапами віртуальних машин та даних.

### Статус реалізації

| Stage | Опис | Статус |
|-------|------|--------|
| **Stage 1** | RBAC/OAuth2/JWT з оновленням токенів | ✅ Завершено |
| **Stage 2** | Оркестрація хмарних провайдерів (AWS/Azure/GCP) | ✅ Завершено |
| **Stage 3** | Веб-панель керування (Dashboard) | ✅ Завершено |
| **Stage 4** | Production Docker/CI pipelines | ✅ Завершено |
| **Stage 5** | Міграція на Python (повна) | ✅ Завершено |
| **Stage 6** | Розширений функціонал (Security+, Audit, Scopes) | ✅ Завершено |
| **Stage 7** | Повні хмарні інтеграції (AWS/Azure/GCP) | ✅ Завершено |
| **Stage 8** | Сучасний веб-інтерфейс (Vue.js 3) | ✅ Завершено |
| **Stage 9** | Моніторинг та сповіщення | ✅ Завершено |
| **Stage 10+** | Розширений UI та автоматизація | 🚧 В розробці |

### ✨ Нові можливості Stage 6

- 🔐 **Розширена RBAC система**
  - Ролі: `admin`, `user`, `service`
  - Scopes: `read`, `write`, `delete`, `manage_users`, `backup`, `restore`
  - Token blacklisting для logout
  - Унікальні JWT token ID (jti)

- 📋 **Audit логування**
  - Всі критичні події логуються
  - Перегляд логів через `/audit/logs`
  - Збереження в пам'яті (останні 1000 записів)

- 🔑 **OAuth2 flow**
  - Access + Refresh токени
  - Автоматичне оновлення токенів
  - Відстеження активних сесій

### 🌩️ Нові можливості Stage 7

- **AWS Integration**
  - EC2 instance listing з деталями (type, AZ, platform)
  - EBS snapshot creation для всіх дисків VM
  - Snapshot tagging (Name, BackupType, InstanceId, ManagedBy)
  - Restore з snapshot
  - Delete backup (cleanup snapshots)

- **Azure Integration**
  - Azure VM listing з resource group та location
  - OS disk + Data disk snapshot creation
  - VM power state monitoring
  - Restore з snapshot
  - Delete backup (cleanup snapshots)

- **GCP Integration**
  - Compute Engine VM listing з zone та machine type
  - Persistent disk snapshot creation
  - Cross-zone backup support
  - Restore з snapshot
  - Delete backup (cleanup snapshots)

### 🎨 Нові можливості Stage 8

- **Сучасний веб-інтерфейс на Vue.js 3**
  - SPA (Single Page Application)
  - Vue Router для навігації
  - Pinia для управління станом
  - Axios для API запитів

- **Компоненти**
  - Dashboard з метриками та статистикою
  - Сторінка віртуальних машин
  - Сторінка резервних копій
  - Audit логи (admin only)
  - Модальні вікна для створення бекапів

- **UI/UX**
  - Адаптивний дизайн (mobile-friendly)
  - Toast сповіщення
  - Іконки Bootstrap Icons
  - Статус бейджі
  - Spinner завантаження

- **Безпека**
  - JWT автентифікація
  - Auto-refresh токенів
  - Logout з відкликанням токену
  - RBAC на рівні сторінок

- **Real-time**
  - WebSocket підтримка
  - Real-time оновлення статусу
  - Моніторинг активних клієнтів

### 🔔 Нові можливості Stage 9

- **Система сповіщень (Notification Manager)**
  - Email сповіщення через SMTP
  - Telegram бот сповіщення
  - Webhook сповіщення (Slack, Teams, custom)
  - Console сповіщення (development)
  - Історія сповіщень

- **Планувальник бекапів (Scheduler)**
  - Cron-like розклад (хвилина година день місяць день тижня)
  - Interval-based розклад (кожен N секунд)
  - One-time заплановані бекапи
  - Persistent збереження розкладу
  - Автоматичне виконання бекапів
  - Статистика виконань (успішні/помилки)

- **Нові API ендпоінти**
  - `POST /notifications/test` - тестове сповіщення
  - `GET /notifications/history` - історія сповіщень
  - `GET /notifications/channels` - активні канали
  - `GET /scheduler/jobs` - список запланованих jobs
  - `POST /scheduler/jobs` - створити job
  - `DELETE /scheduler/jobs/{id}` - видалити job
  - `POST /scheduler/jobs/{id}/enable` - увімкнути job
  - `POST /scheduler/jobs/{id}/disable` - вимкнути job
  - `GET /scheduler/status` - статус планувальника

- **Змінні оточення для сповіщень**
  ```ini
  # Email
  NOVABACKUP_SMTP_HOST=smtp.gmail.com
  NOVABACKUP_SMTP_PORT=587
  NOVABACKUP_SMTP_USER=user@gmail.com
  NOVABACKUP_SMTP_PASSWORD=password
  NOVABACKUP_SMTP_FROM=user@gmail.com
  NOVABACKUP_SMTP_TO=admin@example.com,manager@example.com
  
  # Telegram
  NOVABACKUP_TELEGRAM_BOT_TOKEN=123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11
  NOVABACKUP_TELEGRAM_CHAT_IDS=-1001234567890,-1009876543210
  
  # Webhook
  NOVABACKUP_WEBHOOK_URL=https://hooks.slack.com/services/xxx
  NOVABACKUP_WEBHOOK_AUTH_TOKEN=secret-token
  
  # Debug mode
  NOVABACKUP_DEBUG=true
  ```

> ⚠️ **Важливо**: Починаючи з версії 8.0.0, основна реалізація — **Python**. Go версія застаріла та видалена.

---

## 🔑 Ключові можливості

| Можливість | Опис |
|------------|------|
| 🔐 **RBAC Security** | Контроль доступу на основі ролей з JWT токенами |
| ☁️ **Multi-Cloud** | Підтримка AWS S3, Azure Blob, Google Cloud Storage |
| 🔄 **Backup & Restore** | Повний життєвий цикл резервних копій |
| 📊 **Web Dashboard** | Зручний веб-інтерфейс керування |
| 🚀 **Docker Ready** | Production образи з health checks |
| 📈 **Monitoring** | Інтеграція з Prometheus для метрик |
| 🛡️ **Шифрування** | Підтримка шифрування резервних копій |
| 📧 **Сповіщення** | Telegram, Email сповіщення про події |

---

## ⚡ Швидкий старт

### Windows (PowerShell)

```powershell
# 1. Клонувати репозиторій
git clone https://github.com/ajjs1ajjs/Backup.git
cd Backup

# 2. Створити віртуальне оточення
python -m venv venv
.\venv\Scripts\Activate.ps1

# 3. Встановити залежності
pip install -e ".[api,dev]"

# 4. Створити конфігурацію
copy .env.example .env
.\generate-secrets.ps1 -All

# 5. Запустити сервер
python -m uvicorn novabackup.api:get_app --host 0.0.0.0 --port 8000
```

### Linux / macOS

```bash
# 1. Клонувати репозиторій
git clone https://github.com/ajjs1ajjs/Backup.git
cd Backup

# 2. Створити віртуальне оточення
python3 -m venv venv
source venv/bin/activate

# 3. Встановити залежності
pip install -e ".[api,dev]"

# 4. Створити конфігурацію
cp .env.example .env
python3 generate-secrets.py --all

# 5. Запустити сервер
python3 -m uvicorn novabackup.api:get_app --host 0.0.0.0 --port 8000
```

### 📚 Повна інструкція

Дивіться **[INSTALL_UA.md](INSTALL_UA.md)** для детальної інструкції з встановлення.

---

## 📥 Детальне встановлення

### Вимоги

| Компонент | Версія | Примітка |
|-----------|--------|----------|
| Python | 3.9+ | Обов'язково |
| Git | Будь-яка | Для клонування |
| Docker | 20.10+ | Опціонально |
| Docker Compose | 2.0+ | Опціонально |

### Варіант 1: Автоматичне встановлення

#### Linux/macOS
```bash
curl -fsSL https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install.sh | bash
```

#### Windows (PowerShell)
```powershell
iwr -useb https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install.bat | iex
```

### Варіант 2: Ручне встановлення

```bash
# Клонувати репозиторій
git clone https://github.com/ajjs1ajjs/Backup.git
cd Backup

# Створити віртуальне оточення
python -m venv venv

# Активувати (Windows)
.\venv\Scripts\activate

# Активувати (Linux/macOS)
source venv/bin/activate

# Встановити пакет з API та dev залежностями
pip install -e ".[api,dev]"
```

### Варіант 3: Docker

```bash
# Завантажити готовий образ
docker pull ajjs1ajjs/novabackup:latest

# Або зібрати локально
docker-compose build

# Запустити
docker-compose up -d
```

---

## ⚙️ Конфігурація

### Основні змінні оточення

Створіть файл `.env` на основі `.env.example`:

```bash
# Windows
copy .env.example .env

# Linux/macOS
cp .env.example .env
```

### Критичні налаштування

```ini
# ===========================================
# 🔐 Безпека - ЗМІНІТЬ ЦІ ЗНАЧЕННЯ!
# ===========================================

# Майстер-ключ для шифрування (мін. 32 символи)
# Генерація: openssl rand -hex 32
NOVABACKUP_MASTER_KEY=ваш-безпечний-ключ-мін-32-символи

# JWT секрет для підпису токенів (мін. 32 символи)
# Генерація: openssl rand -hex 32
NOVABACKUP_JWT_SECRET=ваш-jwt-секрет-мін-32-символи

# API ключ для зовнішнього доступу
NOVABACKUP_API_KEY=ваш-api-ключ

# ===========================================
# 💾 База даних
# ===========================================

# SQLite (для розробки)
NOVABACKUP_DATABASE_URL=sqlite:///./novabackup.db

# PostgreSQL (для production)
# NOVABACKUP_DATABASE_URL=postgresql://user:password@localhost:5432/novabackup

# ===========================================
# 🖥️ Сервер
# ===========================================

NOVABACKUP_HOST=0.0.0.0
NOVABACKUP_PORT=8000

# ===========================================
# ☁️ Хмарні провайдери
# ===========================================

# Доступні: MOCK, AWS, AZURE, GCP
NOVABACKUP_CLOUD_PROVIDERS=MOCK
```

### Генерація секретів

#### Windows (PowerShell)
```powershell
.\generate-secrets.ps1 -All
```

#### Linux/macOS
```bash
# Автоматична генерація
python3 generate-secrets.py --all

# Або вручну
openssl rand -hex 32  # Для NOVABACKUP_MASTER_KEY
openssl rand -hex 32  # Для NOVABACKUP_JWT_SECRET
```

---

## 🚀 Використання

### Запуск API сервера

```bash
# Розробка (з авто-релоадом)
python -m uvicorn novabackup.api:get_app --reload --host 0.0.0.0 --port 8000

# Production
python -m uvicorn novabackup.api:get_app --host 0.0.0.0 --port 8000 --workers 4
```

### CLI команди

```bash
# Переглянути список ВМ
novabackup list-vms

# Створити бекап
novabackup create-backup --vm-id vm-123 --dest /backups

# Відновити з бекапу
novabackup restore --backup-id backup-456 --dest /restore

# Переглянути метрики
novabackup metrics
```

### Доступ до веб-інтерфейсу

1. Відкрийте браузер
2. Перейдіть на `http://localhost:8000` (або `http://localhost:8000/static/index.html`)
3. Увійдіть з credentials:
   - **Користувач:** `alice`
   - **Пароль:** `secret`

**Test users:**
| User | Password | Roles | Scopes |
|------|----------|-------|--------|
| alice | secret | admin | read, write, delete, manage_users |
| bob | secret | user | read, write |
| service | service-secret | service | read, backup, restore |

**Веб-інтерфейс включає:**
- 📊 Dashboard з метриками та статистикою
- 💻 Сторінка віртуальних машин
- 🗄️ Сторінка резервних копій
- 🛡️ Audit логи (admin only)
- 🔔 Toast сповіщення
- 📱 Адаптивний дизайн

---

## 📚 API Довідник

### Автентифікація

| Метод | Ендпоінт | Опис |
|-------|----------|------|
| `POST` | `/token` | Отримати access та refresh токени |
| `POST` | `/token/refresh` | Оновити access токен |
| `GET` | `/auth/me` | Отримати інформацію про користувача |
| `POST` | `/auth/logout` | Вийти з системи (відкликати токен) |

### Аудит та безпека

| Метод | Ендпоінт | Опис | Обмеження |
|-------|----------|------|-----------|
| `GET` | `/audit/logs?limit=100` | Перегляд audit логів | Тільки `admin` |

### Віртуальні машини

| Метод | Ендпоінт | Опис |
|-------|----------|------|
| `GET` | `/vms` | Список всіх ВМ |
| `GET` | `/normalize/{vm_type}` | Нормалізація типу ВМ |

### Резервні копії

| Метод | Ендпоінт | Опис |
|-------|----------|------|
| `POST` | `/backups` | Створити нову резервну копію |
| `GET` | `/backups` | Список всіх бекапів |
| `POST` | `/backups/{id}/restore` | Відновити з бекапу |

### Моніторинг

| Метод | Ендпоінт | Опис |
|-------|----------|------|
| `GET` | `/metrics` | Prometheus метрики |
| `GET` | `/health` | Перевірка здоров'я |

### Приклад запиту

```bash
# Отримати токен
curl -X POST http://localhost:8000/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "username=admin&password=secret"

# Отримати список ВМ
curl -X GET http://localhost:8000/vms \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

---

## 🏗️ Архітектура

```
novabackup/
├── api.py           # FastAPI REST ендпоінти
├── cli.py           # CLI інтерфейс (Typer)
├── core.py          # Основна логіка
├── security.py      # RBAC, JWT, автентифікація
├── backup.py        # Менеджер бекапів
├── models.py        # Pydantic моделі
├── db.py            # Робота з БД
├── cloudops.py      # Хмарна оркестрація
├── providers/       # Провайдери (AWS, Azure, GCP)
├── observability.py # Моніторинг та логування
└── static/          # Веб-інтерфейс
```

### Компоненти

| Компонент | Опис |
|-----------|------|
| **Core** | Базова логіка переліку ВМ |
| **API** | REST ендпоінти з RBAC захистом |
| **Security** | JWT токени, ролі, дозволи |
| **Cloud** | Оркестрація хмарних провайдерів |
| **Persistence** | Зберігання в БД або JSON |
| **UI** | Статичний веб-дашборд |

---

## ☁️ Хмарні провайдери

### Підтримувані провайдери

| Провайдер | Статус | Посилання |
|-----------|--------|-----------|
| **Mock** | ✅ | Для тестування та CI |
| **AWS S3** | ✅ | Потрібні credentials |
| **Azure Blob** | ✅ | Потрібні credentials |
| **Google Cloud** | ✅ | Потрібні credentials |

### Налаштування AWS

```ini
NOVABACKUP_CLOUD_PROVIDERS=AWS
AWS_ACCESS_KEY_ID=ваш-access-key
AWS_SECRET_ACCESS_KEY=ваш-secret-key
AWS_DEFAULT_REGION=us-east-1

# Або через profile
# AWS_PROFILE=your-profile-name
```

**Приклад використання CLI:**
```bash
novabackup create-backup --vm-id i-1234567890abcdef0 \
  --destination-type cloud \
  --cloud-provider AWS \
  --cloud-region us-east-1 \
  --cloud-dest s3://my-backup-bucket/
```

### Налаштування Azure

```ini
NOVABACKUP_CLOUD_PROVIDERS=AZURE
AZURE_TENANT_ID=ваш-tenant-id
AZURE_CLIENT_ID=ваш-client-id
AZURE_CLIENT_SECRET=ваш-client-secret
AZURE_SUBSCRIPTION_ID=ваш-subscription-id

# Або через DefaultAzureCredential (рекомендовано для AKS/VM)
# Використовує managed identity або Azure CLI credentials
```

**Приклад використання CLI:**
```bash
novabackup create-backup --vm-id /subscriptions/xxx/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm1 \
  --destination-type cloud \
  --cloud-provider Azure \
  --cloud-region eastus \
  --cloud-dest azure://my-backup-rg/
```

### Налаштування Google Cloud

```ini
NOVABACKUP_CLOUD_PROVIDERS=GCP
GOOGLE_APPLICATION_CREDENTIALS=/шлях/до/service-account.json
GOOGLE_CLOUD_PROJECT=ваш-project-id
```

**Приклад використання CLI:**
```bash
novabackup create-backup --vm-id my-instance-1 \
  --destination-type cloud \
  --cloud-provider GCP \
  --cloud-region us-central1-a \
  --cloud-dest gcp://my-project/backups/
```

---

## 🐳 Docker розгортання

### Локальна розробка

```bash
# Зібрати та запустити
docker-compose up -d

# Переглянути логи
docker-compose logs -f api

# Зупинити
docker-compose down
```

### Production

```bash
# Використовувати production конфігурацію
docker-compose -f docker-compose-prod.yml up -d
```

### Змінні оточення для Docker

```yaml
services:
  api:
    environment:
      - NOVABACKUP_JWT_SECRET=${NOVABACKUP_JWT_SECRET}
      - NOVABACKUP_DATABASE_URL=${NOVABACKUP_DATABASE_URL}
      - NOVABACKUP_CLOUD_PROVIDERS=${NOVABACKUP_CLOUD_PROVIDERS}
```

---

## 🧪 Тестування

```bash
# Запустити всі тести
pytest -v

# Запустити з покриттям
pytest --cov=novabackup --cov-report=html

# Запустити конкретний тест
pytest tests/test_api.py -v
```

### Вимоги для тестів

- Python 3.9+
- pytest 7.0+
- Mock провайдер (за замовчуванням)

---

## 🔧 Вирішення проблем

### Поширені проблеми

#### Помилка імпорту модулів

```bash
# Переконайтесь, що віртуальне оточення активоване
python -m pip install -e ".[api,dev]"
```

#### Порт 8000 зайнятий

```bash
# Змінити порт
python -m uvicorn novabackup.api:get_app --port 8080
```

#### Проблеми з JWT токенами

```bash
# Згенерувати нові секрети
.\generate-secrets.ps1 -All  # Windows
python3 generate-secrets.py --all  # Linux
```

#### Помилки бази даних

```bash
# Видалити стару БД та перезапустити
rm novabackup.db
python -m uvicorn novabackup.api:get_app --reload
```

---

## 🤝 Внесок

### Як допомогти

1. Fork репозиторій
2. Створіть гілку (`git checkout -b feature/AmazingFeature`)
3. Зробіть зміни та тести
4. Закомітьте (`git commit -m 'Add AmazingFeature'`)
5. Push до гілки (`git push origin feature/AmazingFeature`)
6. Відкрийте Pull Request

### Вимоги до коду

```bash
# Форматування
black novabackup/

# Лінтинг
ruff check novabackup/

# Типи
mypy novabackup/

# Тести
pytest --cov=novabackup
```

---

## 📄 Ліцензія

Цей проект ліцензовано за **MIT License** - дивіться файл [LICENSE](LICENSE) для деталей.

---

## 📞 Контакти

- **GitHub**: [@ajjs1ajjs](https://github.com/ajjs1ajjs)
- **Issues**: [GitHub Issues](https://github.com/ajjs1ajjs/Backup/issues)
- **Документація**: [Wiki](https://github.com/ajjs1ajjs/Backup/wiki)

---

## 🇬🇧 English Version

### Quick Start

```bash
git clone https://github.com/ajjs1ajjs/Backup.git
cd Backup
python -m venv venv
source venv/bin/activate  # Windows: .\venv\Scripts\activate
pip install -e ".[api,dev]"
cp .env.example .env
python -m uvicorn novabackup.api:get_app --host 0.0.0.0 --port 8000
```

### API Endpoints

- `POST /token` - Get JWT tokens
- `GET /vms` - List virtual machines
- `POST /backups` - Create backup
- `GET /backups` - List backups
- `POST /backups/{id}/restore` - Restore from backup
- `GET /metrics` - Prometheus metrics

### Documentation

See full English documentation in the repository files.

---

<div align="center">

**Made with ❤️ by OpenCode**

[⬆️ Повернутись до початку](#-novabackup)

</div>
