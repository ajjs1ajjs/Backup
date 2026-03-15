# 🚀 NovaBackup Enterprise v7.0 - Installation Guide

## 📥 Швидка установка

### Windows

**Автоматична установка з GitHub:**

```powershell
# Завантажити install.bat
wget https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install.bat -OutFile install.bat

# Запустити від імені адміністратора
.\install.bat
```

**Ручна установка:**

```powershell
# Завантажити
wget https://github.com/ajjs1ajjs/Backup/releases/latest/download/novabackup-windows-amd64.exe -OutFile NovaBackup.exe

# Встановити службу
.\NovaBackup.exe install

# Запустити
.\NovaBackup.exe start
```

---

### Linux (Ubuntu/Debian)

**Автоматична установка:**

```bash
# Завантажити
wget https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install.sh -O install.sh

# Запустити
sudo bash install.sh
```

**Ручна установка:**

```bash
# Завантажити
wget https://github.com/ajjs1ajjs/Backup/releases/latest/download/novabackup-linux-amd64 -O novabackup

# Встановити
chmod +x novabackup
sudo cp novabackup /opt/novabackup/NovaBackup

# Створити systemd службу
sudo nano /etc/systemd/system/novabackup.service
```

Вміст служби:
```ini
[Unit]
Description=NovaBackup Enterprise v7.0
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/novabackup
ExecStart=/opt/novabackup/NovaBackup server
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
# Увімкнути та запустити
sudo systemctl daemon-reload
sudo systemctl enable novabackup
sudo systemctl start novabackup
```

---

### macOS

**Homebrew (recommended):**

```bash
# Install (coming soon)
brew install novabackup

# Start
brew services start novabackup
```

**Manual:**

```bash
# Download
wget https://github.com/ajjs1ajjs/Backup/releases/latest/download/novabackup-macos-amd64 -O novabackup
# For M1/M2/M3:
# wget https://github.com/ajjs1ajjs/Backup/releases/latest/download/novabackup-macos-arm64 -O novabackup

# Install
chmod +x novabackup
sudo cp novabackup /usr/local/bin/NovaBackup

# Run
NovaBackup server
```

---

## 🔄 Оновлення

### Windows

```powershell
# Завантажити
wget https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/update.bat -OutFile update.bat

# Запустити від імені адміністратора
.\update.bat
```

### Linux

```bash
# Завантажити
wget https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/update.sh -O update.sh

# Запустити
sudo bash update.sh
```

---

## 🌐 Доступ до веб-інтерфейсу

Після встановлення відкрийте:

```
http://localhost:8050

Login: admin
Password: admin123
```

---

## 🔧 Команди

### Windows Service

```powershell
# Встановити службу
NovaBackup.exe install

# Видалити службу
NovaBackup.exe remove

# Запустити службу
NovaBackup.exe start

# Зупинити службу
NovaBackup.exe stop

# Запустити в режимі консолі
NovaBackup.exe debug
```

### Linux Systemd

```bash
# Перевірити статус
sudo systemctl status novabackup

# Запустити
sudo systemctl start novabackup

# Зупинити
sudo systemctl stop novabackup

# Перезапустити
sudo systemctl restart novabackup

# Увімкнути автозапуск
sudo systemctl enable novabackup

# Переглянути логи
sudo journalctl -u novabackup -f
```

---

## 📁 Директорії

### Windows
```
C:\Program Files\NovaBackup\          # Програма
C:\ProgramData\NovaBackup\            # Дані
  ├── Logs\                           # Логи
  ├── Backups\                        # Резервні копії
  └── Config\                         # Конфігурація
```

### Linux
```
/opt/novabackup/                      # Програма
/var/lib/novabackup/                  # Дані
  ├── logs/                           # Логи
  ├── backups/                        # Резервні копії
  └── config/                         # Конфігурація
```

---

## 🔐 Безпека

### Змінити пароль за замовчуванням:

1. Відкрийте Web UI: http://localhost:8050
2. Увійдіть як `admin` / `admin123`
3. Перейдіть в ⚙️ Налаштування → 👥 Користувачі
4. Змініть пароль

### Firewall Rules

**Windows:**
```powershell
# Дозволити порт 8050
netsh advfirewall firewall add rule name="NovaBackup" dir=in action=allow protocol=TCP localport=8050
```

**Linux (UFW):**
```bash
sudo ufw allow 8050/tcp
```

**Linux (firewalld):**
```bash
sudo firewall-cmd --permanent --add-port=8050/tcp
sudo firewall-cmd --reload
```

---

## 🐛 Вирішення проблем

### Служба не запускається

**Windows:**
```powershell
# Перевірити логи
Get-EventLog -LogName Application -Source NovaBackup -Newest 20

# Перевстановити службу
NovaBackup.exe remove
NovaBackup.exe install
NovaBackup.exe start
```

**Linux:**
```bash
# Перевірити логи
sudo journalctl -u novabackup -n 50 --no-pager

# Перевстановити службу
sudo systemctl stop novabackup
sudo /opt/novabackup/NovaBackup server
```

### Порт 8050 зайнятий

Змініть порт в конфігурації:

**Windows:** `C:\ProgramData\NovaBackup\Config\config.json`
**Linux:** `/var/lib/novabackup/config/config.json`

```json
{
  "server": {
    "port": 8051
  }
}
```

---

## 📞 Підтримка

- 📧 Email: support@novabackup.local
- 💬 Telegram: @novabackup
- 📚 Wiki: https://github.com/ajjs1ajjs/Backup/wiki
- 🐛 Issues: https://github.com/ajjs1ajjs/Backup/issues

---

<div align="center">

**NovaBackup Enterprise v7.0**

[Завантажити](https://github.com/ajjs1ajjs/Backup/releases) • [Документація](https://github.com/ajjs1ajjs/Backup/wiki)

🇺🇦 Зроблено в Україні

</div>
