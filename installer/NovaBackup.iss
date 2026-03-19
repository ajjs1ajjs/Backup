; NovaBackup Enterprise v6.0 - Inno Setup Script
; Single EXE Installer

#define MyAppName "NovaBackup Enterprise"
#define MyAppVersion "6.0.0"
#define MyAppPublisher "NovaBackup Team"
#define MyAppURL "https://github.com/ajjs1ajjs/Backup"
#define MyAppExeName "NovaBackup.exe"

[Setup]
; Basic Settings
AppId={{A1B2C3D4-E5F6-7890-ABCD-EF1234567890}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}
AppUpdatesURL={#MyAppURL}
DefaultDirName={autopf}\NovaBackup
DefaultGroupName=NovaBackup Enterprise
AllowNoIcons=yes
LicenseFile=..\installer\License.rtf
OutputDir=.\output
OutputBaseFilename=NovaBackup-6.0.0-Setup
Compression=lzma2/max
SolidCompression=yes
WizardStyle=modern
ArchitecturesAllowed=x64
ArchitecturesInstallIn64BitMode=x64
PrivilegesRequired=admin
PrivilegesRequiredOverridesAllowed=dialog

; Disable uninstall
Uninstallable=yes
UninstallDisplayIcon={app}\NovaBackup.exe

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Files]
; Main application files
Source: "..\build\installer\NovaBackup.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\build\installer\NovaBackup.dll"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\build\installer\NovaBackup.pdb"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\build\installer\NovaBackup.deps.json"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\build\installer\NovaBackup.runtimeconfig.json"; DestDir: "{app}"; Flags: ignoreversion

; Windows Service
Source: "..\build\installer\nova.exe"; DestDir: "{app}"; Flags: ignoreversion

; Dependencies
Source: "..\build\installer\MaterialDesignColors.dll"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\build\installer\MaterialDesignThemes.Wpf.dll"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\build\installer\Microsoft.Xaml.Behaviors.dll"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\build\installer\System.ServiceProcess.ServiceController.dll"; DestDir: "{app}"; Flags: ignoreversion

; Configuration
Source: "..\installer\config.json"; DestDir: "{app}\Config"; Flags: ignoreversion

; Note: Don't need to include MSVC runtime or .NET, they're usually already installed

[Icons]
Name: "{group}\{cm:UninstallProgram,{#MyAppName}}"; Filename: "{uninstallexe}"
Name: "{group}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; WorkingDir: "{app}"
Name: "{autodesktop}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; WorkingDir: "{app}"

[Run]
; Install and start Windows Service
Filename: "{app}\nova.exe"; Parameters: "install"; Flags: runhidden waituntilterminated
Filename: "{app}\nova.exe"; Parameters: "start"; Flags: runhidden waituntilterminated
Filename: "{app}\{#MyAppExeName}"; Description: "{cm:LaunchProgram,{#StringChange(MyAppName, '&', '&&')}}"; Flags: nowait postinstall skipifsilent

[Code]
var
  ServicePage: TWizardPage;
  ServiceStatusLabel: TLabel;

procedure InitializeWizard;
begin
  ServicePage := CreateCustomPage(wpInstalling, 'Installing Service', 'Please wait while the NovaBackup service is being installed...');
  ServiceStatusLabel := TLabel.Create(WizardForm);
  ServiceStatusLabel.Parent := ServicePage.Surface;
  ServiceStatusLabel.Left := 20;
  ServiceStatusLabel.Top := 20;
  ServiceStatusLabel.Width := 400;
  ServiceStatusLabel.Caption := 'Installing Windows Service...';
  ServiceStatusLabel.Font.Style := [fsBold];
end;

procedure CurStepChanged(CurStep: TSetupStep);
var
  ResultCode: Integer;
begin
  if CurStep = ssPostInstall then
  begin
    // Service installation is handled by [Run] section
  end;
end;

function InitializeSetup(): Boolean;
var
  ResultCode: Integer;
begin
  Result := True;

  // Check for .NET 8.0 (optional, app will show error if not installed)
  // This is just a warning, not a hard requirement
end;

procedure DeinitializeSetup();
begin
  // Cleanup if needed
end;

[UninstallRun]
; Stop and remove service on uninstall
Filename: "{app}\nova.exe"; Parameters: "stop"; Flags: runhidden waituntilterminated
Filename: "{app}\nova.exe"; Parameters: "remove"; Flags: runhidden waituntilterminated

[UninstallDelete]
; Clean up data directories
Type: filesandordirs; Name: "{commonappdata}\NovaBackup\Logs"
Type: filesandordirs; Name: "{commonappdata}\NovaBackup\Backups"
Type: filesandordirs; Name: "{commonappdata}\NovaBackup\Config"
