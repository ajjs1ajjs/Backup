[Setup]
AppName=NOVA Backup
AppVersion=1.0.0
DefaultDirName={pf}\NovaBackup
DefaultGroupName=NOVA Backup
OutputDir=installer
OutputSetupFilename=NovaBackupSetup.exe
SetupIconFile=app\icons\NovaBackup.ico
Compression=lzma
SolidCompression=yes
PrivilegesRequired=admin

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"
Name: "ukrainian"; MessagesFile: "compiler:Languages\Ukrainian.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked
Name: "quicklaunchicon"; Description: "{cm:CreateQuickLaunchIcon}"; GroupDescription: "{cm:AdditionalIcons}"; OnlyBelowVersion: 6.1; Flags: unchecked
Name: "service"; Description: "Install Windows Service"; GroupDescription: "Service Options"; Flags: unchecked
Name: "webconsole"; Description: "Enable Web Console"; GroupDescription: "Web Options"; Flags: unchecked

[Files]
Source: "app\*"; DestDir: "{app}"; Flags: ignoreversion recursesubdirs createallsubfolders
Source: "services\*"; DestDir: "{app}\services"; Flags: ignoreversion recursesubdirs createallsubfolders
Source: "web-ui\*"; DestDir: "{app}\web-ui"; Flags: ignoreversion recursesubdirs createallsubfolders
Source: "app\icons\NovaBackup.ico"; DestDir: "{app}\icons"; Flags: ignoreversion

[Dirs]
Name: "{app}\logs"
Name: "{app}\config"
Name: "{app}\backups"

[Icons]
Name: "{group}\NOVA Backup"; Filename: "{app}\NovaBackup.exe"; IconFilename: "{app}\icons\NovaBackup.ico"
Name: "{group}\Web Console"; Filename: "http://localhost:8080"; IconFilename: "{app}\icons\NovaBackup.ico"
Name: "{group}\{cm:UninstallProgram,NOVA Backup}"; Filename: "{uninstallexe}"
Name: "{commondesktop}\NOVA Backup"; Filename: "{app}\NovaBackup.exe"; IconFilename: "{app}\icons\NovaBackup.ico"; Tasks: desktopicon
Name: "{userappdata}\Microsoft\Internet Explorer\Quick Launch\NOVA Backup"; Filename: "{app}\NovaBackup.exe"; Tasks: quicklaunchicon

[Run]
Filename: "{app}\NovaBackup.exe"; Description: "Launch NOVA Backup"; Flags: nowait postinstall skipifsilent
Filename: "{app}\services\NovaBackupService.exe"; Parameters: "install"; Description: "Install Windows Service"; Tasks: service; Flags: runhidden
Filename: "{app}\services\NovaBackupService.exe"; Parameters: "start"; Description: "Start Windows Service"; Tasks: service; Flags: runhidden
Filename: "netsh"; Parameters: "advfirewall firewall add rule name=""NOVA Backup Web Console"" dir=in action=allow protocol=TCP localport=8080"; Description: "Configure firewall for web console"; Tasks: webconsole; Flags: runhidden

[UninstallRun]
Filename: "{app}\services\NovaBackupService.exe"; Parameters: "stop"; RunOnceId: "StopService"; Flags: runhidden
Filename: "{app}\services\NovaBackupService.exe"; Parameters: "uninstall"; RunOnceId: "UninstallService"; Flags: runhidden

[Registry]
Root: HKLM; Subkey: "SOFTWARE\NovaBackup"; ValueType: string; ValueName: "InstallPath"; ValueData: "{app}"
Root: HKLM; Subkey: "SOFTWARE\NovaBackup"; ValueType: string; ValueName: "Version"; ValueData: "1.0.0"
Root: HKLM; Subkey: "SOFTWARE\NovaBackup"; ValueType: dword; ValueName: "WebConsoleEnabled"; ValueData: "1"; Tasks: webconsole
Root: HKLM; Subkey: "SOFTWARE\NovaBackup"; ValueType: dword; ValueName: "ServiceInstalled"; ValueData: "1"; Tasks: service

[Code]
function IsDotNetInstalled: Boolean;
var
  ErrorCode: Integer;
  IsInstalled: Boolean;
begin
  // Check for .NET 6.0 or later
  IsInstalled := RegQueryDwordValue(HKLM, 'SOFTWARE\Microsoft\NET Framework Setup\NDP\v6\Full', 'Install', ErrorCode);
  if not IsInstalled or (ErrorCode <> 1) then
  begin
    Result := False;
    Exit;
  end;
  
  Result := True;
end;

function InitializeSetup(): Boolean;
begin
  Result := True;
  
  // Check .NET installation
  if not IsDotNetInstalled then
  begin
    if MsgBox('NOVA Backup requires .NET 6.0 or later. Would you like to download and install it now?', mbConfirmation, MB_YESNO) = IDYES then
    begin
      ShellExec('open', 'https://dotnet.microsoft.com/download/dotnet/6.0', '', '', SW_SHOWNORMAL, ewNoWait, ErrorCode);
    end;
    Result := False;
  end;
end;

procedure CurStepChanged(CurStep: TSetupStep);
var
  ErrorCode: Integer;
begin
  if CurStep = ssPostInstall then
  then
  begin
    // Create default configuration
    CreateDefaultConfig();
    
    // Set up firewall rules if web console is enabled
    if WizardIsTaskSelected('webconsole') then
    begin
      ShellExec('', 'netsh', 'advfirewall firewall add rule name="NOVA Backup Web Console" dir=in action=allow protocol=TCP localport=8080', '', SW_HIDE, ewWaitUntilTerminated, ErrorCode);
    end;
  end;
end;

procedure CreateDefaultConfig;
var
  ConfigFile: String;
  ConfigContent: String;
begin
  ConfigFile := ExpandConstant('{app}\config\backup-config.json');
  ConfigContent := '{' +
    '"BackupJobs": [],' +
    '"Schedules": [],' +
    '"Settings": {' +
    '"DefaultBackupPath": "' + ExpandConstant('{app}\backups') + '",' +
    '"MaxConcurrentBackups": 3,' +
    '"CompressionLevel": "normal",' +
    '"EnableEncryption": true,' +
    '"EnableNotifications": true,' +
    '"WebConsoleEnabled": ' + IIf(WizardIsTaskSelected('webconsole'), 'true', 'false') + ',' +
    '"WebConsolePort": 8080' +
    '},' +
    '"UpdatedAt": "' + FormatDateTime('yyyy-mm-dd hh:nn:ss', Now) + '"' +
    '}';
  
  SaveStringToFile(ConfigFile, ConfigContent, False);
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
var
  ErrorCode: Integer;
begin
  if CurUninstallStep = usPostUninstall then
  begin
    // Remove firewall rules
    ShellExec('', 'netsh', 'advfirewall firewall delete rule name="NOVA Backup Web Console"', '', SW_HIDE, ewWaitUntilTerminated, ErrorCode);
    
    // Clean up registry
    RegDeleteKeyIncludingSubkeys(HKLM, 'SOFTWARE\NovaBackup');
  end;
end;

// Custom messages
; Ukrainian custom messages
[CustomMessages]
uk.CreateDesktopIcon=Створити іконку на робочому столі
uk.CreateQuickLaunchIcon=Створити іконку швидкого запуску
uk.AdditionalIcons=Додаткові іконки
uk.ServiceOptions=Опції служби
uk.WebOptions=Веб опції
uk.InstallWindowsService=Встановити службу Windows
uk.EnableWebConsole=Увімкнути веб-консоль
uk.LaunchOnStartup=Запускати при старті
uk.AssociateFileTypes=Асоціювати типи файлів
uk.AutoStart=Автоматичний запуск

[Types]
Name: "full"; Description: "Full installation"
Name: "compact"; Description: "Compact installation"
Name: "custom"; Description: "Custom installation"; Flags: iscustom

[Components]
Name: "main"; Description: "Main Application"; Types: full compact custom; Flags: fixed
Name: "service"; Description: "Windows Service"; Types: full
Name: "webconsole"; Description: "Web Console"; Types: full
Name: "documentation"; Description: "Documentation"; Types: full

[Components]
; Ukrainian component descriptions
uk.Name: "main"; Description: "Основна програма"
uk.Name: "service"; Description: "Служба Windows"
uk.Name: "webconsole"; Description: "Веб-консоль"
uk.Name: "documentation"; Description: "Документація"
