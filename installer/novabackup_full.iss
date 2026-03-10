; NovaBackup v6.0 - Inno Setup Script 
#define MyAppName \"NovaBackup\" 
#define MyAppVersion \"6.0.0\" 
  
[Setup] 
AppId={{A3B5C7D9-E1F2-3456-7890-ABCDEF123456} 
AppName={#MyAppName} 
AppVersion={#MyAppVersion} 
DefaultDirName={autopf}\{#MyAppName} 
OutputDir=.\output 
OutputBaseFilename=NovaBackup-{#MyAppVersion}-Setup 
  
[Files] 
Source: \"..\dist\nova.exe\"; DestDir: \"{app}\"; Flags: ignoreversion 
Source: \"..\dist\menu.bat\"; DestDir: \"{app}\"; Flags: ignoreversion 
  
[Icons] 
Name: \"{group}\{#MyAppName}\"; Filename: \"{app}\nova.exe\" 
Name: \"{group}\{#MyAppName} Menu\"; Filename: \"{app}\menu.bat\" 
  
[Run] 
Filename: \"{app}\nova.exe\"; Parameters: \"service install\"; Flags: runhidden 
  
