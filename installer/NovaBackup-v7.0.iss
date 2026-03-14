[Setup]
AppName=NovaBackup v7.0 Enterprise
AppVersion=7.0.0
AppPublisher=NovaBackup Technologies
AppPublisherURL=https://novabackup.local
AppSupportURL=https://novabackup.local/support
AppUpdatesURL=https://novabackup.local/updates
DefaultDirName={pf}\NovaBackup
DefaultGroupName=NovaBackup
AllowNoIcons=yes
LicenseFile=license.txt
InfoBeforeFile=readme.txt
OutputDir=output
OutputBaseFilename=NovaBackup-v7.0-Enterprise-Setup
Compression=lzma2
SolidCompression=yes
WizardStyle=modern
PrivilegesRequired=admin
ArchitecturesAllowed=x64
ArchitecturesInstallIn64BitMode=x64

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"
Name: "ukrainian"; MessagesFile: "compiler:Languages\Ukrainian.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked
Name: "quicklaunchicon"; Description: "{cm:CreateQuickLaunchIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked; OnlyBelowVersion: 6.1; Check: not IsAdminInstallMode

[Files]
Source: "publish\v7.0\*"; DestDir: "{app}"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "license.txt"; DestDir: "{app}"; Flags: ignoreversion
Source: "readme.txt"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\NovaBackup v7.0"; Filename: "{app}\NovaBackup.exe"
Name: "{group}\{cm:UninstallProgram,NovaBackup v7.0}"; Filename: "{uninstallexe}"
Name: "{autodesktop}\NovaBackup v7.0"; Filename: "{app}\NovaBackup.exe"; Tasks: desktopicon
Name: "{userappdata}\Microsoft\Internet Explorer\Quick Launch\NovaBackup v7.0"; Filename: "{app}\NovaBackup.exe"; Tasks: quicklaunchicon

[Run]
Filename: "{app}\NovaBackup.exe"; Description: "{cm:LaunchProgram,NovaBackup v7.0 Enterprise}"; Flags: nowait postinstall skipifsilent

[Code]
function InitializeSetup(): Boolean;
var
  ResultCode: Integer;
begin
  Result := True;
end;

function IsUpgrade(): Boolean;
begin
  Result := RegKeyExists(HKEY_LOCAL_MACHINE, 'SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\NovaBackup v7.0_is1');
end;

procedure CurStepChanged(CurStep: TSetupStep);
begin
  if CurStep = ssPostInstall then
  begin
    // Post-installation tasks
  end;
end;
