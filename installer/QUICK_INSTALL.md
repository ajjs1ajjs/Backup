# 🚀 NovaBackup - Швидка Інструкція з Встановлення

## ✅ Швидке Встановлення (1 Крок!)

### Варіант 1: Двічі клацнути на MSI
```
1. Двічі клацніть на: installer\NovaBackup-1.0.0.msi
2. Натисніть "Install"
3. Готово!
```

### Варіант 2: PowerShell (автоматично)
```powershell
cd D:\WORK_CODE\Backup\installer
.\BuildInstaller.ps1 -Install
```

### Варіант 3: Batch скрипт
```batch
cd D:\WORK_CODE\Backup\installer
Install.bat
```

**Все! Служба встановлюється автоматично через MSI.**

---

## 📦 Що Встановлюється

| Компонент | Опис |
|-----------|------|
| `NovaBackup.GUI.exe` | Графічний інтерфейс |
| `NovaBackup.Agent.exe` | Фоновий агент |
| `NovaBackup Agent Service` | Служба Windows (автоматично) |
| `NovaBackup.*.dll` | Основні модулі |
| Ярлик в меню Пуск | Для запуску програми |
| Ярлик на робочому столі | Для швидкого доступу |

---

## 🎯 Перевірка Встановлення

### Перевірити програму:
```
C:\Program Files\NovaBackup\NovaBackup.GUI.exe
```

### Перевірити службу:
```cmd
sc query "NovaBackup Agent"
```

Має показати:
```
SERVICE_NAME: NovaBackup Agent
        TYPE               : 10  WIN32_OWN_PROCESS
        STATE              : 4  RUNNING
```

---

## 🔧 Управління Службою

### Запустити службу:
```cmd
net start "NovaBackup Agent"
```

### Зупинити службу:
```cmd
net stop "NovaBackup Agent"
```

### Видалити службу:
```cmd
sc delete "NovaBackup Agent"
```

---

## ❌ Видалення Програми

### Через Панель Керування:
1. Панель керування → Програми та компоненти
2. Знайдіть "NovaBackup"
3. Натисніть "Видалити"

### Через командний рядок:
```cmd
msiexec /x {DD271EB5-FC16-4B71-813D-E6F37EDE51E4} /quiet
```

---

## 🐛 Усунення Проблем

### "Access Denied" / "Потрібні права адміністратора"
**Рішення:** Запустіть MSI або скрипт від імені адміністратора:
- Right-click → "Run as administrator"

### "Service failed to start" (Помилка 1053)
**Рішення:**
1. Відкрийте Event Viewer → Windows Logs → Application
2. Перевірте логи NovaBackup
3. Переустановіть: `msiexec /x {ProductCode}` → `msiexec /i NovaBackup-1.0.0.msi`

### Програма не запускається
**Рішення:** Встановіть .NET Desktop Runtime 8.0:
https://dotnet.microsoft.com/download/dotnet/8.0

---

## 📝 Логи

| Лог | Розташування |
|-----|--------------|
| Інсталятор | `installer\install.log` |
| Служба | Event Viewer → Application |
| Програма | `%ProgramData%\NovaBackup\Logs\` |
| База даних | `%ProgramData%\NovaBackup\backup.db` |

---

## 🔐 Безпека

Інсталятор вимагає права адміністратора тому що:
- Встановлює в `Program Files`
- Створює службу Windows
- Записує в HKLM реєстр

Це необхідно для коректної роботи фонових backup jobs.
