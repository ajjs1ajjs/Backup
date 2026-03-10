# NovaBackup v6.0 - Windows Service Installation  
  
## Architecture  
  
NovaBackup runs as a **Windows Service** which provides:  
  
- ? Automatic startup on boot  
- ? Runs in background (no user login required)  
- ? Survives user logoff  
- ? Automatic restart on failure  
- ? Managed via Windows Services Manager  
  
## Quick Install  
  
**1. Run install.bat as Administrator**  
  
This will:  
- Install NovaBackup Windows Service  
- Start the service automatically  
- Add to PATH  
  
**2. Verify installation**  
  
```batch  
nova-cli --help  
```  
  
## Management  
  
### Via Command Line  
  
```batch  
# Start service  
net start NovaBackup  
  
# Stop service  
net stop NovaBackup  
  
# Check status  
sc query NovaBackup  
```  
  
### Via Services Manager  
  
1. Press Win+R  
2. Type services.msc  
3. Find NovaBackup service  
4. Right-click  Start/Stop/Restart  
  
## Web Access  
  
After installation:  
  
- **API**: http://localhost:8080  
- **Swagger UI**: http://localhost:8080/swagger  
- **Web Dashboard**: http://localhost:3000 (if running)  
  
## Backup Commands  
  
```batch  
# File backup  
nova-cli backup run -s C:\Data -d D:\Backups -c  
  
# Create scheduled job  
nova-cli backup create -n Daily -s C:\Data -d D:\Backups --schedule \"0 2 * * *\"  
  
# View jobs  
nova-cli backup list  
```  
  
