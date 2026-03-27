# 🚀 NovaBackup Enterprise v8.0.0 - Інструкція з Встановлення

## 📋 Зміст

- [Швидка установка](#-швидка-установка)
- [Встановлення з Git](#-встановлення-з-git)
- [Docker установка](#-docker-установка)
- [Перший вхід та налаштування](#-перший-вхід-та-налаштування)
- [Конфігурація безпеки](#-конфігурація-безпеки)
- [Команди та управління](#-команди-та-управління)
- [Вирішення проблем](#-вирішення-проблем)

---

## 📥 Швидка установка

### Windows (Python версія - Production)

#### Спосіб 1: Автоматична установка

```powershell
# 1. Завантажити інсталятор
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install.bat" -OutFile "install.bat"

# 2. Запустити ВІД ІМЕНІ АДМІНІСТРАТОРА
.\install.bat
```

#### Одна команда

```powershell
# Встановити
iwr -useb https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install.bat | iex
```

#### Спосіб 2: Ручна установка

```powershell
# 1. Встановити Python 3.9+ (якщо не встановлено)
winget install Python.Python.3.12

# 2. Клонувати репозиторій
git clone https://github.com/ajjs1ajjs/Backup.git
cd Backup

# 3. Створити віртуальне оточення
python -m venv venv
.\venv\Scripts\Activate.ps1

# 4. Встановити залежності
pip install --upgrade pip
pip install -e ".[api,dev]"

# 5. Створити файл конфігурації
copy .env.example .env

# 6. Згенерувати секрети
.\generate-secrets.ps1 -All

# 7. Запустити сервер
python -m uvicorn novabackup.api:app --host 0.0.0.0 --port 8000
```

### Linux (Ubuntu/Debian)

```bash
# 1. Встановити залежності
sudo apt update
sudo apt install -y python3 python3-pip python3-venv git curl

# 2. Клонувати репозиторій
git clone https://github.com/ajjs1ajjs/Backup.git
cd Backup

# 3. Створити віртуальне оточення
python3 -m venv venv
source venv/bin/activate

# 4. Встановити залежності
pip install --upgrade pip
pip install -e ".[api,dev]"

# 5. Створити конфігурацію
cp .env.example .env

# 6. Згенерувати секрети
python generate-secrets.py --all

# 7. Запустити сервер
python -m uvicorn novabackup.api:app --host 0.0.0.0 --port 8000
```

---

## 🐳 Docker установка

### Розробка (docker-compose)

```bash
# 1. Клонувати репозиторій
git clone https://github.com/ajjs1ajjs/Backup.git
cd Backup

# 2. Створити .env файл з секретами
cp .env.example .env
# Відредагуйте .env і змініть паролі за замовчуванням!

# 3. Запустити всі сервіси
docker-compose up -d

# 4. Перевірити статус
docker-compose ps

# 5. Переглянути логи
docker-compose logs -f api
```

### Production (docker-compose-prod)

```bash
# 1. Створити .env файл
cp .env.example .env

# 2. Згенерувати безпечні секрети
# Windows:
.\generate-secrets.ps1 -UpdateEnv

# Linux:
python3 generate-secrets.py --update-env

# 3. Запустити production версію
docker-compose -f docker-compose-prod.yml up -d

# 4. Перевірити здоров'я сервісів
docker-compose -f docker-compose-prod.yml ps
```

### Окремий Docker контейнер

```bash
# Побудувати образ
docker build -t novabackup:latest .

# Запустити з секретами
docker run -d \
  --name novabackup \
  -p 8000:8000 \
  --env-file .env \
  -v novabackup_data:/app/data \
  novabackup:latest
```

---

## 🔐 Конфігурація безпеки

### Генерація секретів

**Windows:**
```powershell
# Згенерувати всі секрети
.\generate-secrets.ps1 -All

# Або по окремості:
.\generate-secrets.ps1 -MasterKey      # Master ключ шифрування
.\generate-secrets.ps1 -JwtSecret      # JWT секрет
.\generate-secrets.ps1 -ApiKey         # API ключ
.\generate-secrets.ps1 -Password       # Безпечний пароль
```

**Linux:**
```bash
# Вручну з OpenSSL
openssl rand -hex 32  # Для JWT secret (64 символи)
openssl rand -hex 16  # Для API key (32 символи)
```

### Налаштування .env файлу

```bash
# Обов'язково змініть ці значення!
NOVABACKUP_MASTER_KEY=your-64-char-hex-key-here
NOVABACKUP_JWT_SECRET=your-64-char-random-string
NOVABACKUP_API_KEY=your-32-char-api-key

# PostgreSQL (для production)
POSTGRES_USER=novabackup
POSTGRES_PASSWORD=your-secure-postgres-password
POSTGRES_DB=novabackup
```

### Firewall правила

**Windows:**
```powershell
# Дозволити порт 8000
netsh advfirewall firewall add rule name="NovaBackup API" dir=in action=allow protocol=TCP localport=8000

# Дозволити порт 8080 (Dashboard)
netsh advfirewall firewall add rule name="NovaBackup Dashboard" dir=in action=allow protocol=TCP localport=8080
```

**Linux (UFW):**
```bash
sudo ufw allow 8000/tcp
sudo ufw allow 8080/tcp
sudo ufw enable
```

---

## 🔑 Перший вхід та налаштування

### 1. Відкрийте веб-інтерфейс

```
http://localhost:8000/docs    # API документація (Swagger)
http://localhost:8080         # Dashboard (якщо використовується)
```

### 2. Створіть першого користувача-адміна

**Через API:**
```bash
curl -X POST http://localhost:8000/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "YourSecurePassword123!",
    "email": "admin@example.com",
    "roles": ["admin"]
  }'
```

**Або через CLI:**
```bash
source venv/bin/activate  # Windows: .\venv\Scripts\Activate.ps1
novabackup create-user admin --password "YourSecurePassword123!" --role admin
```

### ⚠️ Важливо!

- **Змініть паролі за замовчуванням** перед production deployment
- **Ніколи не зберігайте .env файл в Git** (вже додано до .gitignore)
- **Використовуйте HTTPS** в production

---

## 🛠️ Команди та управління

### Універсальний менеджер (Windows)

```powershell
# Статус
.\nova-manager.ps1 -Action status

# Запуск
.\nova-manager.ps1 -Action start

# Зупинка
.\nova-manager.ps1 -Action stop

# Перезапуск
.\nova-manager.ps1 -Action restart

# Повне вбивство процесів
.\nova-manager.ps1 -Action kill

# Встановити як службу Windows
.\nova-manager.ps1 -Action install

# Видалити службу
.\nova-manager.ps1 -Action uninstall
```

### Docker команди

```bash
# Запуск
docker-compose up -d

# Зупинка
docker-compose down

# Перезапуск
docker-compose restart

# Перегляд логів
docker-compose logs -f

# Оновлення
docker-compose pull
docker-compose up -d --force-recreate
```

### Linux Systemd (production)

```bash
# Створити файл оточення
sudo mkdir -p /etc/novabackup
sudo cp .env /etc/novabackup/env
sudo chmod 600 /etc/novabackup/env

# Встановити службу
sudo cp deploy/systemd/novabackup.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable novabackup
sudo systemctl start novabackup

# Перевірити статус
sudo systemctl status novabackup

# Переглянути логи
sudo journalctl -u novabackup -f
```

---

## 📁 Структура директорій

### Windows
```
C:\Program Files\NovaBackup\          # Програма
C:\ProgramData\NovaBackup\            # Дані
  ├── Config\                         # Конфігурація
  ├── Logs\                           # Журнали
  ├── Backups\                        # Резервні копії
  └── novabackup.db                   # База даних
```

### Linux
```
/opt/novabackup/                      # Програма
/var/lib/novabackup/                  # Дані
  ├── config/                         # Конфігурація
  ├── logs/                           # Журнали
  ├── backups/                        # Резервні копії
  └── novabackup.db                   # База даних
```

### Docker volumes
```
postgres_data:/var/lib/postgresql/data    # PostgreSQL дані
novabackup_data:/app/data                 # NovaBackup дані
```

---

## 🐛 Вирішення проблем

### Служба не запускається

**Windows:**
```powershell
# Переглянути логи подій
Get-EventLog -LogName Application -Source *nova* -Newest 20

# Перевірити чи порт вільний
netstat -ano | findstr ":8000"

# Перевстановити службу
.\nova-manager.ps1 -Action uninstall
.\nova-manager.ps1 -Action install
.\nova-manager.ps1 -Action start
```

**Linux:**
```bash
# Переглянути логи
sudo journalctl -u novabackup -n 50 --no-pager

# Перевірити порт
sudo netstat -tlnp | grep 8000

# Перезапустити
sudo systemctl restart novabackup
```

### Помилки аутентифікації

```bash
# Переглянути JWT секрети в .env
cat .env | grep SECRET

# Перезгенерувати секрети
.\generate-secrets.ps1 -JwtSecret

# Перезапустити сервіс
docker-compose restart api
```

### Порт 8000 зайнятий

Змініть порт в конфігурації:

**docker-compose.yml:**
```yaml
ports:
  - "8001:8000"  # Замість 8000:8000
```

**.env:**
```bash
NOVABACKUP_PORT=8001
```

### Docker контейнер не запускається

```bash
# Переглянути логи
docker logs novabackup_api

# Перевірити .env файл
docker-compose config

# Перезапустити з rebuild
docker-compose up -d --build --force-recreate
```

### База даних не підключається

```bash
# Перевірити PostgreSQL
docker-compose exec db psql -U novabackup -c "SELECT 1"

# Скинути пароль
docker-compose exec db psql -U postgres -c "ALTER USER novabackup PASSWORD 'new-password';"
```

---

## 📞 Підтримка

- 📧 **Email:** support@novabackup.local
- 💬 **Telegram:** @novabackup
- 📚 **Wiki:** https://github.com/ajjs1ajjs/Backup/wiki
- 🐛 **Issues:** https://github.com/ajjs1ajjs/Backup/issues
- 💻 **GitHub:** https://github.com/ajjs1ajjs/Backup

---

## 📄 Ліцензія

MIT License - див. файл [LICENSE](LICENSE)

---

<div align="center">

**NovaBackup Enterprise v8.0.0**

[Завантажити](https://github.com/ajjs1ajjs/Backup/releases) • [Документація](https://github.com/ajjs1ajjs/Backup/wiki) • [Звіти](https://github.com/ajjs1ajjs/Backup/issues)

🇺🇦 Зроблено в Україні

</div>
