# NovaBackup v6.0 - Installation 
  
## Quick Start (Portable - Recommended)  
  
Simply run:  
```  
menu.bat  
```  
  
## Full Installation  
  
Run as Administrator:  
```batch  
mkdir \"C:\Program Files\NovaBackup\"  
copy nova.exe \"C:\Program Files\NovaBackup\\\"  
\"C:\Program Files\NovaBackup\nova.exe\" service install  
net start NovaBackup  
```  
  
## Usage  
  
### Interactive Menu  
```batch  
menu.bat  
```  
  
### Command Line  
```batch  
nova backup run -s C:\Data -d D:\Backups -c  
nova api start  
nova --help  
```  
