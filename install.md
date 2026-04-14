# 📖 Встановлення Fortress Backup Enterprise

Система підтримує повністю автоматичне розгортання на Windows.

## Автоматична інсталяція (Windows)

1. Відкрийте **PowerShell** від імені Адміністратора.
2. Виконайте наступні команди для встановлення системи "з нуля":

```powershell
# 1. Завантажити оновлений інсталятор
iwr -useb https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/installer.ps1 -OutFile installer.ps1

# 2. Запустити інсталяцію (встановить .NET, Node, PostgreSQL, Build Tools та запустить сервер як службу)
.\installer.ps1 -AutoStart
```

## Що робить інсталятор?
- **Залежності:** Автоматично перевіряє та встановлює через `winget`: `.NET 8 SDK`, `Node.js`, `PostgreSQL 16`, `MSVC Build Tools`.
- **Збірка:** Клонує актуальну версію коду, очищує кеші та збирає сервер і **C++ Агент з підтримкою VSS**.
- **Сервіс:** Створює та запускає Windows-службу `BackupServer`.
- **Конфігурація:** Створює `appsettings.json` із підключенням до локального `PostgreSQL`.

---

## 🛠️ Після інсталяції
1. **Веб-інтерфейс:** Перейдіть на `http://localhost:8000`.
2. **Перший вхід:** Використовуйте `Admin` / `Lkmo291263@`.
3. **Налаштування:**
   - Перейдіть у розділ **Settings**, щоб підключити HashiCorp Vault для шифрування.
   - Встановіть **2FA** у профілі користувача для підвищення безпеки.
   - Налаштуйте репозиторії (S3, Azure або Local).

## 💡 Troubleshooting
Якщо збірка не проходить (наприклад, помилки `CS0246` або `NU1202`), це означає, що середовище кешує старі файли. Виконайте:
```powershell
# Очищення тимчасових файлів
Remove-Item -Path "C:\Users\$env:USERNAME\AppData\Local\Temp\Backup-*" -Recurse -Force
# Перезапустіть installer.ps1
```
