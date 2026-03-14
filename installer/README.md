# 🚀 NovaBackup - Швидке Встановлення

## ⚡ Найшвидший Спосіб (Автоматична Установка)

### Крок 1: Збірка та Встановлення

**Варіант A: PowerShell (Рекомендовано)**
```powershell
cd D:\WORK_CODE\Backup\installer
.\BuildInstaller.ps1 -Install
```
Це автоматично:
- Збере проект
- Створить MSI інсталятор
- Запустить встановлення з правами адміністратора

**Варіант B: Тільки збірка**
```powershell
.\BuildInstaller.ps1
```
Потім двічі клацніть на `NovaBackup-1.0.0.msi`

**Варіант C: Batch скрипт (авто-elevate)**
```batch
Install.bat
```

---

## 📦 Що Створюється

| Файл | Опис |
|------|------|
| `NovaBackup-1.0.0.msi` | MSI інсталятор (вимагає права адміна) |
| `Install.bat` | Скрипт для авто-запуску з правами адміна |
| `BuildInstaller.ps1` | PowerShell скрипт для збірки та встановлення |

---

## 🔧 Вимоги

- Windows 10/11 x64
- .NET Desktop Runtime 8.0 (якщо немає - інсталятор запропонує встановити)
- **Права адміністратора** (обов'язково для встановлення служби)
- 200 MB вільного місця

---

## 🎯 Детальні Інструкції

### Спосіб 1: PowerShell Script (Рекомендовано)

```powershell
# Перейдіть до папки інсталятора
cd D:\WORK_CODE\Backup\installer

# Зберіть та встановіть (автоматично запитає права адміна)
.\BuildInstaller.ps1 -Install

# АБО тільки збірка без встановлення
.\BuildInstaller.ps1
```

### Спосіб 2: Batch Script (Авто-Elevation)

```batch
# Двічі клацніть на Install.bat
# АБО з командного рядка:
Install.bat
```

Скрипт автоматично:
1. Перевірить чи є права адміна
2. Якщо ні - запросить UAC elevation
3. Запустить MSI інсталятор

### Спосіб 3: Ручне Встановлення MSI

1. Зберіть інсталятор:
   ```powershell
   .\BuildInstaller.ps1
   ```

2. Двічі клацніть на `NovaBackup-1.0.0.msi`
   - **АБО** запустіть з правами адміна:
   ```cmd
   msiexec /i NovaBackup-1.0.0.msi
   ```

3. Натисніть "Next" → "Install"

---

## ✅ Перевірка Встановлення

### Перевірити службу:
```cmd
sc query "NovaBackup Agent"
```

### Перевірити планувальник:
```powershell
schtasks /query /fo LIST | findstr NovaBackup
```

### Запустити програму:
```
C:\Program Files\NovaBackup\NovaBackup.GUI.exe
```

---

## ❌ Видалення

### Через Панель Керування:
1. Панель керування → Програми та компоненти
2. Знайдіть "NovaBackup"
3. Натисніть "Видалити"

### Через командний рядок:
```cmd
msiexec /x {ProductCode} /quiet
```

### Або через PowerShell:
```powershell
Get-WmiObject -Class Win32_Product | Where-Object {$_.Name -like "NovaBackup"} | Remove-WmiObject
```

---

## 🐛 Усунення Проблем

### "Access Denied" / "Потрібні права адміністратора"
**Рішення:** Використовуйте `Install.bat` або запустіть MSI з правами адміна:
```cmd
Right-click NovaBackup-1.0.0.msi → Run as administrator
```

### "A newer version is already installed"
**Рішення:** Спочатку видаліть стару версію, потім встановіть нову.

### "You must install .NET Desktop Runtime"
**Рішення:** Завантажте та встановіть:
https://dotnet.microsoft.com/download/dotnet/8.0

Або використайте ZIP версію (не вимагає Runtime).

### Служба не запускається
1. Відкрийте Event Viewer
2. Перевірте логи: `Windows Logs → Application`
3. Переконайтесь що у вас права адміністратора

---

## 📝 Логи

| Лог | Розташування |
|-----|--------------|
| Інсталятор | `installer\install.log` |
| Служба | Event Viewer → Application |
| Програма | `%ProgramData%\NovaBackup\Logs\` |
| База даних | `%ProgramData%\NovaBackup\backup.db` |

---

## 🎯 Архітектура Встановлення

```
BuildInstaller.ps1
       │
       ├─► dotnet publish (збірка проекту)
       │
       └─► wix build (створення MSI)
               │
               └─► NovaBackup-1.0.0.msi
                       │
                       ├─► Install.bat (авто-elevation)
                       │
                       └─► Ручний запуск (двічі клацнути)
```

---

## 📞 Підтримка

- GitHub: https://github.com/ajjs1ajjs/Backup/issues
- Документація: `USER_GUIDE.md`
- Сайт: https://novabackup.local

---

## 🔐 Безпека

Інсталятор вимагає права адміністратора тому що:
- Встановлює службу Windows (`NovaBackup Agent Service`)
- Записує в `Program Files`
- Створює завдання в Task Scheduler
- Записує в HKLM реєстр

Це необхідно для коректної роботи фонових backup jobs.
