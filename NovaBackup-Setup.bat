@echo off
chcp 1251 >nul
title NovaBackup v6.0 Enterprise - Professional Installer
color 0B

:: Перевірка прав адміністратора
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo.
    echo ========================================
    echo    ПОМИЛКА: Потрібні права адміністратора!
    echo ========================================
    echo.
    echo Будь ласка, запустіть інсталятор від імені адміністратора.
    echo Клікніть правою кнопкою миші на файл і виберіть "Запустити від імені адміністратора".
    echo.
    pause
    exit /b 1
)

cls
echo ========================================
echo    NovaBackup v6.0 Enterprise
echo    Професійний інсталятор
echo ========================================
echo.

echo Перевірка системних вимог...
ver | find "Windows 10" >nul
if %errorLevel% equ 0 (
    echo [OK] Windows 10/11 - сумісно
) else (
    echo [INFO] Версія Windows: %os%
)

echo.
echo Створення директорій для інсталяції...
if not exist "%ProgramFiles%\NovaBackup" (
    mkdir "%ProgramFiles%\NovaBackup"
    echo [OK] Створено директорію програми
)

if not exist "%ProgramFiles%\NovaBackup\bin" mkdir "%ProgramFiles%\NovaBackup\bin"
if not exist "%ProgramFiles%\NovaBackup\gui" mkdir "%ProgramFiles%\NovaBackup\gui"
if not exist "%ProgramFiles%\NovaBackup\agent" mkdir "%ProgramFiles%\NovaBackup\agent"
if not exist "%ProgramFiles%\NovaBackup\config" mkdir "%ProgramFiles%\NovaBackup\config"
if not exist "%ProgramFiles%\NovaBackup\logs" mkdir "%ProgramFiles%\NovaBackup\logs"

echo.
echo Копіювання файлів програми...

:: Копіювання основних файлів
if exist "nova.exe" (
    copy "nova.exe" "%ProgramFiles%\NovaBackup\bin\" >nul 2>&1
    if %errorLevel% equ 0 (
        echo [OK] Скопійовано nova.exe
    ) else (
        echo [ПОМИЛКА] Помилка копіювання nova.exe
    )
)

if exist "NovaBackup-Enterprise.bat" (
    copy "NovaBackup-Enterprise.bat" "%ProgramFiles%\NovaBackup\bin\NovaBackup.exe" >nul 2>&1
    if %errorLevel% equ 0 (
        echo [OK] Скопійовано GUI launcher
    ) else (
        echo [ПОМИЛКА] Помилка копіювання GUI launcher
    )
)

:: Копіювання GUI файлів
if exist "gui\app.py" (
    copy "gui\app.py" "%ProgramFiles%\NovaBackup\gui\" >nul 2>&1
    echo [OK] Скопійовано GUI backend
)

if exist "gui\templates\veeam-style.html" (
    copy "gui\templates\veeam-style.html" "%ProgramFiles%\NovaBackup\gui\templates\" >nul 2>&1
    echo [OK] Скопійовано Veeam-style interface
)

if exist "requirements.txt" (
    copy "requirements.txt" "%ProgramFiles%\NovaBackup\gui\" >nul 2>&1
    echo [OK] Скопійовано Python dependencies
)

:: Копіювання агентів
if exist "NovaBackupAgent.ps1" (
    copy "NovaBackupAgent.ps1" "%ProgramFiles%\NovaBackup\agent\" >nul 2>&1
    echo [OK] Скопійовано NovaBackup Agent
)

if exist "nova-gui-manager.ps1" (
    copy "nova-gui-manager.ps1" "%ProgramFiles%\NovaBackup\agent\" >nul 2>&1
    echo [OK] Скопійовано GUI Manager
)

echo.
echo Створення ярликів...

:: Ярлик в меню Пуск
if not exist "%APPDATA%\Microsoft\Windows\Start Menu\Programs\NovaBackup" (
    mkdir "%APPDATA%\Microsoft\Windows\Start Menu\Programs\NovaBackup"
)

echo [InternetShortcut] > "%APPDATA%\Microsoft\Windows\Start Menu\Programs\NovaBackup\NovaBackup Enterprise.url"
echo URL=file:///%ProgramFiles%/NovaBackup/bin/NovaBackup.exe >> "%APPDATA%\Microsoft\Windows\Start Menu\Programs\NovaBackup\NovaBackup Enterprise.url"
echo IconFile=%ProgramFiles%/NovaBackup/bin/NovaBackup.exe >> "%APPDATA%\Microsoft\Windows\Start Menu\Programs\NovaBackup\NovaBackup Enterprise.url"
echo IconIndex=0 >> "%APPDATA%\Microsoft\Windows\Start Menu\Programs\NovaBackup\NovaBackup Enterprise.url"

echo [OK] Створено ярлик в меню Пуск

:: Ярлик на робочому столі
echo Set WshShell = CreateObject("WScript.Shell") > "%TEMP%\DesktopShortcut.vbs"
echo strDesktop = WshShell.SpecialFolders("Desktop") >> "%TEMP%\DesktopShortcut.vbs"
echo Set oShellLink = WshShell.CreateShortcut(strDesktop & "\NovaBackup Enterprise.lnk") >> "%TEMP%\DesktopShortcut.vbs"
echo oShellLink.TargetPath = "%ProgramFiles%\NovaBackup\bin\NovaBackup.exe" >> "%TEMP%\DesktopShortcut.vbs"
echo oShellLink.WorkingDirectory = "%ProgramFiles%\NovaBackup\bin" >> "%TEMP%\DesktopShortcut.vbs"
echo oShellLink.Description = "NovaBackup v6.0 Enterprise - Система резервного копіювання" >> "%TEMP%\DesktopShortcut.vbs"
echo oShellLink.Save >> "%TEMP%\DesktopShortcut.vbs"
cscript //nologo "%TEMP%\DesktopShortcut.vbs"
del "%TEMP%\DesktopShortcut.vbs" >nul 2>&1

echo [OK] Створено ярлик на робочому столі

echo.
echo Встановлення служби Windows...
if exist "%ProgramFiles%\NovaBackup\bin\nova.exe" (
    "%ProgramFiles%\NovaBackup\bin\nova.exe" service install >nul 2>&1
    if %errorLevel% equ 0 (
        echo [OK] Службу NovaBackup встановлено
        "%ProgramFiles%\NovaBackup\bin\nova.exe" service start >nul 2>&1
        if %errorLevel% equ 0 (
            echo [OK] Службу NovaBackup запущено
        ) else (
            echo [INFO] Службу не вдалося запустити автоматично
        )
    ) else (
        echo [ПОМИЛКА] Не вдалося встановити службу
    )
)

echo.
echo Створення конфігураційних файлів...

:: Основний конфігураційний файл
echo [NovaBackup] > "%ProgramFiles%\NovaBackup\config\novabackup.ini"
echo Version=6.0.0 >> "%ProgramFiles%\NovaBackup\config\novabackup.ini"
echo InstallPath=%ProgramFiles%\NovaBackup >> "%ProgramFiles%\NovaBackup\config\novabackup.ini"
echo InstallDate=%date% %time% >> "%ProgramFiles%\NovaBackup\config\novabackup.ini"
echo Language=Ukrainian >> "%ProgramFiles%\NovaBackup\config\novabackup.ini"

echo [GUI] >> "%ProgramFiles%\NovaBackup\config\novabackup.ini"
echo Theme=Veeam >> "%ProgramFiles%\NovaBackup\config\novabackup.ini"
echo AutoStart=true >> "%ProgramFiles%\NovaBackup\config\novabackup.ini"

echo [Backup] >> "%ProgramFiles%\NovaBackup\config\novabackup.ini"
echo DefaultCompression=Optimal >> "%ProgramFiles%\NovaBackup\config\novabackup.ini"
echo DefaultDeduplication=true >> "%ProgramFiles%\NovaBackup\config\novabackup.ini"
echo DefaultEncryption=false >> "%ProgramFiles%\NovaBackup\config\novabackup.ini"
echo MaxConcurrentJobs=4 >> "%ProgramFiles%\NovaBackup\config\novabackup.ini"

echo [OK] Створено конфігураційний файл

:: Реєстр для інтеграції з системою
reg add "HKLM\SOFTWARE\NovaBackup" /v "InstallPath" /t REG_SZ /d "%ProgramFiles%\NovaBackup" /f >nul 2>&1
reg add "HKLM\SOFTWARE\NovaBackup" /v "Version" /t REG_SZ /d "6.0.0" /f >nul 2>&1
reg add "HKLM\SOFTWARE\NovaBackup" /v "InstallDate" /t REG_SZ /d "%date% %time%" /f >nul 2>&1

echo [OK] Додано записи в реєстр

:: Асоціація файлів
reg add "HKLM\SOFTWARE\Classes\.nbk" /ve /d "NovaBackup Job" /f >nul 2>&1
reg add "HKLM\SOFTWARE\Classes\.nbk\DefaultIcon" /ve /d "%ProgramFiles%\NovaBackup\bin\NovaBackup.exe,0" /f >nul 2>&1
reg add "HKLM\SOFTWARE\Classes\.nbk\shell\open\command" /ve /d "\"%ProgramFiles%\NovaBackup\bin\NovaBackup.exe\" \"%%1\"" /f >nul 2>&1

echo [OK] Асоційовано типи файлів

echo.
echo ========================================
echo    ІНСТАЛЯЦІЮ ЗАВЕРШЕНО!
echo ========================================
echo.
echo NovaBackup v6.0 Enterprise успішно встановлено!
echo.
echo Місце інсталяції: %ProgramFiles%\NovaBackup
echo.
echo Що було встановлено:
echo   - Veeam-style графічний інтерфейс
echo   - Веб-інтерфейс управління  
echo   - Фоновий агент-служба
echo   - Інтеграція з Windows
echo   - Ярлики в меню Пуск та на робочому столі
echo   - Асоціація файлів (.nbk)
echo.
echo Тепер ви можете:
echo 1. Запустити NovaBackup з меню Пуск
echo 2. Двічі клікнути на ярлик на робочому столі
echo 3. Запустити: "%ProgramFiles%\NovaBackup\bin\NovaBackup.exe"
echo.
echo Дякуємо, що обрали NovaBackup Enterprise!
echo.

echo Запускаємо NovaBackup...
timeout /t 3 /nobreak >nul
start "" "%ProgramFiles%\NovaBackup\bin\NovaBackup.exe"

echo.
echo Інсталятор завершив роботу.
pause
