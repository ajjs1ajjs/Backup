@echo off  
echo ========================================  
echo   NovaBackup v6.0 Uninstallation      
echo ========================================  
echo.  
net stop NovaBackup  
\"C:\Program Files\NovaBackup\nova.exe\" service remove  
rmdir /s /q \"C:\Program Files\NovaBackup\"  
  
Uninstallation complete!  
pause  
