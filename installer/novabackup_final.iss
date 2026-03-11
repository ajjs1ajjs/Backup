[Setup] 
AppId={{A3B5C7D9-E1F2-3456-7890-ABCDEF123456}} 
AppName=NovaBackup 
AppVersion=6.0.0 
AppPublisher=NovaBackup Team 
DefaultDirName={autopf}\NovaBackup 
OutputDir=.\output 
OutputBaseFilename=NovaBackup-6.0.0-Setup 
Compression=lzma2/max 
SolidCompression=yes 
WizardStyle=modern 
 
[Languages] 
Name: english; MessagesFile: compiler:Default.isl 
Name: ukrainian; MessagesFile: compiler:Languages\Ukrainian.isl 
 
[Tasks] 
Name: desktopicon; Description: Create desktop shortcut; GroupDescription: Additional icons; Flags: unchecked 
 
[Files] 
Source: ..\nova.exe; DestDir: {app} 
Source: ..\dist\menu.bat; DestDir: {app} 
Source: ..\dist\README.md; DestDir: {app} 
Source: ..\dist\INSTALL.md; DestDir: {app} 
Source: ..\web-ui\*; DestDir: {app}\web-ui; Flags: recursesubdirs 
Source: ..\launcher.bat; DestDir: {app} 
Source: ..\NovaBackupAgent.ps1; DestDir: {app} 
Source: ..\nova-gui-manager.ps1; DestDir: {app} 
Source: ..\NovaBackup-Manager.bat; DestDir: {app} 
 
[Icons] 
Name: {group}\NovaBackup Manager; Filename: {app}\nova-gui-manager.ps1 
Name: {group}\NovaBackup Web GUI; Filename: {app}\launcher.bat 
Name: {group}\NovaBackup Menu; Filename: {app}\menu.bat 
Name: {group}\NovaBackup CMD; Filename: {cmd}; WorkingDir: {app} 
Name: {group}\Uninstall; Filename: {uninstallexe} 
Name: {autodesktop}\NovaBackup Manager; Filename: {app}\nova-gui-manager.ps1; Tasks: desktopicon 
 
[Run] 
Filename: {app}\nova.exe; Parameters: service install; Flags: runhidden 
