# 🚀 NovaBackup Enterprise v7.0 - Інструкція з Встановлення

## 📥 Швидка Установка з GitHub

### Windows

#### Спосіб 1: Автоматична установка (Рекомендовано)

```powershell
# 1. Завантажити інсталятор
wget https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install.bat -OutFile install.bat

# 2. Запустити ВІД ІМЕНІ АДМІНІСТРАТОРА
.\install.bat
```

#### Одна команда (ВІД ІМЕНІ АДМІНІСТРАТОРА)

```powershell
# Встановити
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install.bat" -OutFile "install.bat"; .\install.bat

# Оновити
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/update.bat" -OutFile "update.bat"; .\update.bat
```

#### Один скрипт (інтерактивно)

```powershell
# Завантажити та запустити (встановлення/оновлення/видалення)
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/novabackup-setup.bat" -OutFile "novabackup-setup.bat"; .\novabackup-setup.bat
```

Примітка: якщо ви не використовуєте GitHub Releases, тримайте `novabackup.exe` оновленим у корені репозиторію.

#### Спосіб 2: Ручна установка

```powershell
# 1. Завантажити останній реліз
wget https://github.com/ajjs1ajjs/Backup/releases/latest/download/novabackup-windows-amd64.exe -OutFile NovaBackup.exe

# 2. Встановити службу
.\NovaBackup.exe install

# 3. Запустити службу
.\NovaBackup.exe start

# 4. Відкрити веб-інтерфейс
start http://localhost:8050
```

---

### Linux (Ubuntu/Debian)

```bash
# 1. Завантажити інсталятор
wget https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install.sh -O install.sh

# 2. Запустити від root
sudo bash install.sh

# 3. Відкрити веб-інтерфейс
xdg-open http://localhost:8050
```

#### Одна команда

```bash
# Встановити
curl -fsSL https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install.sh | sudo bash

# Оновити
curl -fsSL https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/update.sh | sudo bash
```

---

## 🔐 Перший Вхід

1. **Відкрийте** http://localhost:8050
2. **Увійдіть** як:
   - Логін: `admin`
   - Пароль: `admin123`
3. **Змініть пароль** на надійний (обов'язково!)
   - Мінімум 8 символів
   - Великі та малі літери
   - Цифри

---

## 📋 Перші Кроки

### 1. Створіть Перше Резервне Копіювання

**Швидкий спосіб:**
1. Натисніть **"Швидке Резервне Копіювання"** в меню
2. Оберіть тип джерела:
   - 📁 **Локальні файли** - файли на цьому сервері
   - 🌐 **Мережеве джерело** - SMB/CIFS частка
   - 🖥️ **Віддалений сервер** - інший сервер
   - ☁️ **Хмарне джерело** - S3, Azure, Google Cloud
3. Вкажіть шлях (наприклад: `C:\Documents` або `\\server\share`)
4. Оберіть сховище призначення (наприклад: `D:\Backups`)
5. Налаштуйте політику зберігання:
   - **Зберігати:** 30 днів
   - **Макс. копій:** 10
6. Натисніть **"Запустити Резервне Копіювання"**

**Приклади шляхів:**
```
Локально:  C:\Users\Admin\Documents
Мережа:    \\192.168.1.100\Backup
S3:        my-backup-bucket (регіон: us-east-1)
```

---

### 2. Додайте Сховище

1. Перейдіть в **"Сховища"**
2. Натисніть **"➕ Додати сховище"**
3. Оберіть тип:
   - 📂 **Локальне** - папка на диску
   - 🌐 **SMB/CIFS** - мережева частка
   - ☁️ **Amazon S3** - хмарне сховище
   - 🔷 **Azure Blob** - Microsoft Azure
   - 📁 **NFS** - NFS сховище
4. Заповніть параметри
5. Натисніть **"Додати сховище"**

---

### 3. Створіть Користувача

1. Перейдіть в **"Користувачі"**
2. Натисніть **"➕ Додати користувача"**
3. Заповніть:
   - **Ім'я користувача:** `ivan`
   - **Пароль:** `Ivan1234!` (мінімум 6 символів)
   - **Повне ім'я:** `Ivan Petrenko`
   - **Email:** `ivan@example.com`
   - **Роль:** `Backup User`
4. Натисніть **"Додати користувача"**

**Доступні ролі:**
- 👑 **Адміністратор** - повний доступ
- 🔧 **Адмін бекапів** - управління бекапами
- 💾 **Користувач бекапів** - виконання бекапів
- 📖 **Тільки читання** - перегляд

---

### 4. Налаштуйте Сповіщення

1. Перейдіть в **"Сповіщення"**
2. Оберіть канали:
   - 📧 **Email** - SMTP сервер
   - 💬 **Telegram** - бот токен + chat ID
   - 🔷 **MS Teams** - webhook URL
   - 💬 **Slack** - webhook URL
3. Оберіть події:
   - ✅ Початок бекапу
   - ✅ Успішний бекап
   - ❌ Помилка бекапу
   - ⚠️ Мало місця
4. Натисніть **"Зберегти налаштування"**

---

## 🔄 Оновлення

### Windows

```powershell
# Завантажити оновлення
wget https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/update.bat -OutFile update.bat

# Запустити від адміністратора
.\update.bat
```

### Linux

```bash
# Завантажити оновлення
wget https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/update.sh -O update.sh

# Запустити
sudo bash update.sh
```

---

## 🛠️ Команди

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

# Перезапустити службу
NovaBackup.exe restart

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

---

## 🔐 Безпека

### Зміна Паролю Адміна

1. Увійдіть як `admin` / `admin123`
2. Система автоматично запропонує змінити пароль
3. Введіть новий пароль (мінімум 8 символів)
4. Збережіть

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

## 🐛 Вирішення Проблем

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

# Перезапустити
sudo systemctl restart novabackup
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

Перезапустіть службу після змін.

### Веб-інтерфейс не вантажиться

1. Перевірте чи працює служба:
   ```powershell
   Get-Service NovaBackup
   ```

2. Перевірте логи:
   ```powershell
   Get-Content "C:\ProgramData\NovaBackup\Logs\novabackup.log" -Tail 50
   ```

3. Перезапустіть службу:
   ```powershell
   Restart-Service NovaBackup
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

MIT License - див. файл LICENSE

---

<div align="center">

**NovaBackup Enterprise v7.0**

[Завантажити](https://github.com/ajjs1ajjs/Backup/releases) • [Документація](https://github.com/ajjs1ajjs/Backup/wiki) • [Звіти](https://github.com/ajjs1ajjs/Backup/issues)

🇺🇦 Зроблено в Україні

</div>
