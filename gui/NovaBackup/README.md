# NovaBackup GUI - WinForms Application 
 
## How to Build 
 
### Option 1: Visual Studio 
1. Open Visual Studio 2022 
2. Create New Project 
3. Windows Forms App (.NET 6 or .NET 8) 
4. Name: NovaBackup 
5. Copy files from this folder 
 
### Option 2: .NET CLI 
```bash 
dotnet new winforms -n NovaBackup -f net8.0 
cd NovaBackup 
dotnet run 
``` 
 
## Project Structure 
- MainForm.cs - Main application window 
- Dashboard - Statistics and overview 
- JobManager - Backup job management 
- Settings - Application settings 
 
