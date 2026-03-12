[Setup]
; Визначення інсталятора
AppName=NovaBackup v7.0 Enterprise
AppVersion=7.0.0
AppPublisher=NovaBackup Technologies
AppPublisherURL=https://novabackup.com
AppSupportURL=https://novabackup.com/support
AppUpdatesURL=https://novabackup.com/updates
DefaultDirName={pf}\NovaBackup
DefaultGroupName=NovaBackup
AllowNoIcons=yes
LicenseFile=license.txt
InfoBeforeFile=readme.txt
OutputDir=installer\output
OutputBaseFilename=NovaBackup-Enterprise-Setup
SetupIconFile=icons\novabackup.ico
UninstallDisplayIcon={app}\nova-unified.exe
Compression=lzma2
SolidCompression=yes
WizardStyle=modern
WizardImageFile=images\wizard.bmp
WizardSmallImageFile=images\wizard-small.bmp
DisableStartupPrompt=yes
PrivilegesRequired=admin
ArchitecturesAllowed=x64
ArchitecturesInstallIn64BitMode=x64

[Languages]
Name: "ukrainian"; MessagesFile: "compiler:Languages\Ukrainian.isl"
Name: "english"; MessagesFile: "compiler:Default.isl"

[Types]
Name: "full"; Description: "Повна інсталяція з усіма компонентами"; Flags: iscustom
Name: "compact"; Description: "Мінімальна інсталяція"; Flags: iscustom
Name: "custom"; Description: "Вибіркова інсталяція"; Flags: iscustom

[Components]
Name: "core"; Description: "Основні компоненти NovaBackup"; Types: full compact custom; Flags: fixed
Name: "gui"; Description: "Графічний інтерфейс (Veeam Style)"; Types: full custom
Name: "web"; Description: "Веб-інтерфейс"; Types: full custom
Name: "agent"; Description: "Фоновий агент"; Types: full custom
Name: "docs"; Description: "Документація"; Types: full custom

[Dirs]
; Створення директорій
Name: "{app}"; Components: core
Name: "{app}\bin"; Components: core
Name: "{app}\gui"; Components: gui
Name: "{app}\web"; Components: web
Name: "{app}\agent"; Components: agent
Name: "{app}\docs"; Components: docs
Name: "{app}\logs"; Components: core
Name: "{app}\jobs"; Components: core
Name: "{app}\repository"; Components: core
Name: "{app}\config"; Components: core
Name: "{commonappdata}\NovaBackup"; Components: core
Name: "{commonappdata}\NovaBackup\logs"; Components: core
Name: "{commonappdata}\NovaBackup\jobs"; Components: core
Name: "{commonappdata}\NovaBackup\repository"; Components: core

[Files]
; Основні файли програми
Source: "..\nova.exe"; DestDir: "{app}\bin"; Components: core; Flags: ignoreversion
Source: "..\nova-service.exe"; DestDir: "{app}\bin"; Components: core; Flags: ignoreversion
Source: "..\nova-cli.exe"; DestDir: "{app}\bin"; Components: core; Flags: ignoreversion

; GUI компоненти
Source: "..\NovaBackup-Enterprise.bat"; DestDir: "{app}\bin"; DestName: "NovaBackup.exe"; Components: gui; Flags: ignoreversion
Source: "..\gui\app.py"; DestDir: "{app}\gui"; Components: gui; Flags: ignoreversion
Source: "..\gui\templates\veeam-style.html"; DestDir: "{app}\gui\templates"; Components: gui; Flags: ignoreversion
Source: "..\requirements.txt"; DestDir: "{app}\gui"; Components: gui; Flags: ignoreversion

; Веб-інтерфейс
Source: "..\web-ui\*"; DestDir: "{app}\web"; Components: web; Flags: ignoreversion recursesubdirs createallsubdirs

; Агент
Source: "..\NovaBackupAgent.ps1"; DestDir: "{app}\agent"; Components: agent; Flags: ignoreversion
Source: "..\nova-gui-manager.ps1"; DestDir: "{app}\agent"; Components: agent; Flags: ignoreversion

; Документація
Source: "..\README.md"; DestDir: "{app}\docs"; Components: docs; Flags: ignoreversion
Source: "..\docs\*"; DestDir: "{app}\docs"; Components: docs; Flags: ignoreversion recursesubdirs createallsubdirs

; Іконки та ресурси
Source: "icons\novabackup.ico"; DestDir: "{app}"; Flags: ignoreversion
Source: "images\*"; DestDir: "{app}\images"; Flags: ignoreversion recursesubdirs createallsubdirs

[Icons]
; Ярлики в меню Пуск
Name: "{group}\NovaBackup Enterprise"; Filename: "{app}\bin\NovaBackup.exe"; Components: gui; IconFilename: "{app}\novabackup.ico"
Name: "{group}\NovaBackup Web Interface"; Filename: "{app}\web\index.html"; Components: web; IconFilename: "{app}\novabackup.ico"
Name: "{group}\NovaBackup Agent Manager"; Filename: "{app}\agent\nova-gui-manager.ps1"; Components: agent; IconFilename: "{app}\novabackup.ico"
Name: "{group}\Documentation"; Filename: "{app}\docs\README.md"; Components: docs; IconFilename: "{app}\novabackup.ico"
Name: "{group}\Uninstall NovaBackup"; Filename: "{uninstallexe}"; Components: core

; Ярлики на робочому столі
Name: "{commondesktop}\NovaBackup Enterprise"; Filename: "{app}\bin\NovaBackup.exe"; Components: gui; IconFilename: "{app}\novabackup.ico"; Tasks: desktopicon

[Tasks]
Name: "desktopicon"; Description: "Створити ярлик на робочому столі"; GroupDescription: "Додаткові ярлики:"; Components: gui
Name: "quicklaunchicon"; Description: "Створити ярлик у швидкому запуску"; GroupDescription: "Додаткові ярлики:"; Components: gui
Name: "autostart"; Description: "Запускати NovaBackup при старті системи"; GroupDescription: "Автозапуск:"; Components: agent

[Run]
; Файли, що виконуються після інсталяції
Filename: "{app}\bin\nova.exe"; Parameters: "service install"; Components: core; Flags: runhidden runascurrentuser
Filename: "{app}\bin\nova.exe"; Parameters: "service start"; Components: core; Flags: runhidden runascurrentuser
Filename: "{app}\bin\NovaBackup.exe"; Description: "Запустити NovaBackup Enterprise"; Components: gui; Flags: nowait postinstall runascurrentuser

[UninstallRun]
; Файли, що виконуються при деінсталяції
Filename: "{app}\bin\nova.exe"; Parameters: "service stop"; Flags: runhidden
Filename: "{app}\bin\nova.exe"; Parameters: "service remove"; Flags: runhidden

[Registry]
; Реєстр для інтеграції з системою
Root: HKLM; Subkey: "SOFTWARE\NovaBackup"; ValueType: string; ValueName: "InstallPath"; ValueData: "{app}"; Components: core
Root: HKLM; Subkey: "SOFTWARE\NovaBackup"; ValueType: string; ValueName: "Version"; ValueData: "6.0.0"; Components: core
Root: HKLM; Subkey: "SOFTWARE\NovaBackup"; ValueType: string; ValueName: "InstallDate"; ValueData: "{code:GetDateTime|Now}"; Components: core
Root: HKLM; Subkey: "SOFTWARE\NovaBackup\Components"; ValueType: dword; ValueName: "GUI"; ValueData: 1; Components: gui
Root: HKLM; Subkey: "SOFTWARE\NovaBackup\Components"; ValueType: dword; ValueName: "Web"; ValueData: 1; Components: web
Root: HKLM; Subkey: "SOFTWARE\NovaBackup\Components"; ValueType: dword; ValueName: "Agent"; ValueData: 1; Components: agent

[Code]
function GetDateTime(Param: String): String;
var
  Year, Month, Day, Hour, Min, Sec: Word;
begin
  DecodeDate(Now, Year, Month, Day);
  DecodeTime(Now, Hour, Min, Sec, 0);
  Result := Format('%d-%.2d-%.2d %.2d:%.2d:%.2d', [Year, Month, Day, Hour, Min, Sec]);
end;

// Перевірка системних вимог
function InitializeSetup(): Boolean;
var
  ResultCode: Integer;
begin
  // Перевірка Windows версії
  if UsingWinNT() = False then
  begin
    MsgBox('NovaBackup v6.0 вимагає Windows 7 або новішу версію.', mbError, MB_OK);
    Result := False;
    Exit;
  end;
  
  // Перевірка адміністраторських прав
  if not IsAdminLoggedOn then
  begin
    MsgBox('Для інсталяції NovaBackup потрібні адміністраторські права.', mbError, MB_OK);
    Result := False;
    Exit;
  end;
  
  Result := True;
end;

// Створення служби Windows
procedure CreateNovaBackupService();
var
  ResultCode: Integer;
begin
  if Exec(ExpandConstant('{app}\bin\nova-service.exe'), 
           ExpandConstant('install'), 
           '', 
           SW_SHOW, 
           ewWaitUntilTerminated, 
           ResultCode) then
  begin
    if ResultCode = 0 then
      Log('NovaBackup service installed successfully')
    else
      Log('Failed to install NovaBackup service. Error code: ' + IntToStr(ResultCode));
  end;
end;

// Запуск служби
procedure StartNovaBackupService();
var
  ResultCode: Integer;
begin
  if Exec(ExpandConstant('{app}\bin\nova-service.exe'), 
           ExpandConstant('start'), 
           '', 
           SW_SHOW, 
           ewWaitUntilTerminated, 
           ResultCode) then
  begin
    if ResultCode = 0 then
      Log('NovaBackup service started successfully')
    else
      Log('Failed to start NovaBackup service. Error code: ' + IntToStr(ResultCode));
  end;
end;

// Встановлення Python залежностей
procedure InstallPythonDependencies();
var
  ResultCode: Integer;
begin
  // Перевірка чи встановлено Python
  if Exec('python', '--version', '', SW_HIDE, ewWaitUntilTerminated, ResultCode) then
  begin
    if ResultCode = 0 then
    begin
      // Встановлення Python пакетів
      Exec('pip', 'install -r requirements.txt', ExpandConstant('{app}\gui'), SW_HIDE, ewWaitUntilTerminated, ResultCode);
      Log('Python dependencies installed');
    end;
  end;
end;

procedure CurStepChanged(CurStep: TSetupStep);
begin
  if CurStep = ssPostInstall then
  begin
    // Створення служби після інсталяції
    CreateNovaBackupService();
    StartNovaBackupService();
    
    // Встановлення Python залежностей
    InstallPythonDependencies();
    
    // Створення конфігураційних файлів
    CreateConfigFiles();
  end;
end;

// Створення конфігураційних файлів
procedure CreateConfigFiles();
var
  ConfigFile: String;
begin
  ConfigFile := ExpandConstant('{app}\config\novabackup.conf');
  
  // Створення базового конфігураційного файлу
  SaveStringToFile(ConfigFile, 
    '# NovaBackup v6.0 Configuration' + #13#10 +
    '[General]' + #13#10 +
    'Version=6.0.0' + #13#10 +
    'InstallPath=' + ExpandConstant('{app}') + #13#10 +
    'LogPath=' + ExpandConstant('{commonappdata}\NovaBackup\logs') + #13#10 +
    'JobPath=' + ExpandConstant('{commonappdata}\NovaBackup\jobs') + #13#10 +
    'RepositoryPath=' + ExpandConstant('{commonappdata}\NovaBackup\repository') + #13#10 +
    '' + #13#10 +
    '[GUI]' + #13#10 +
    'Theme=Veeam' + #13#10 +
    'Language=ukrainian' + #13#10 +
    'AutoStart=true' + #13#10 +
    '' + #13#10 +
    '[Backup]' + #13#10 +
    'DefaultCompression=Optimal' + #13#10 +
    'DefaultDeduplication=true' + #13#10 +
    'DefaultEncryption=false' + #13#10 +
    'MaxConcurrentJobs=4' + #13#10 +
    '' + #13#10 +
    '[Repository]' + #13#10 +
    'DefaultRepository=Primary' + #13#10 +
    'RetentionDays=30' + #13#10 +
    'StorageLimit=1TB' + #13#10,
    False);
end;

// Перевірка попередньої версії
function ShouldSkipPage(PageID: Integer): Boolean;
begin
  Result := False;
  
  if PageID = wpReady then
  begin
    if RegKeyExists(HKEY_LOCAL_MACHINE, 'SOFTWARE\NovaBackup') then
    begin
      if MsgBox('Знайдено попередню версію NovaBackup. Бажаєте оновити її?', 
                 mbConfirmation, MB_YESNO or MB_DEFBUTTON2) = IDYES then
      begin
        Result := True; // Пропустити сторінку вибору компонентів
      end;
    end;
  end;
end;

[Messages]
ukrainian.LicenseAccepted=Я приймаю умови ліцензійної угоди
ukrainian.SetupAppTitle=Інсталятор NovaBackup v6.0 Enterprise
ukrainian.SetupWindowTitle=Інсталятор NovaBackup v6.0 Enterprise
ukrainian.WelcomeLabel1=Ласкаво просимо до інсталятора NovaBackup v6.0 Enterprise
ukrainian.WelcomeLabel2=Цей інсталятор встановить NovaBackup v6.0 Enterprise на ваш комп'ютер.%n%nNovaBackup - це професійна система резервного копіювання рівня enterprise.
ukrainian.SelectDirLabel3=NovaBackup буде встановлено в наступну директорію.
ukrainian.SelectComponentsLabel=Виберіть компоненти, які потрібно встановити:
ukrainian.SelectComponentsDesc2=Для повнофункціональної роботи рекомендується встановити усі компоненти.
ukrainian.FinishedLabel=NovaBackup v6.0 Enterprise успішно встановлено на ваш комп'ютер.
ukrainian.FinishedLabel2=Програма готова до використання. Ви можете запустити її з меню Пуск або з ярликів на робочому столі.
ukrainian.RunProgram=Запустити NovaBackup Enterprise
