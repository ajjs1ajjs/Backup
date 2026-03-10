@echo off  
  
echo ========================================  
echo   NovaBackup v6.0 Uninstallation      
echo ========================================  
echo.  
  
echo [1/3] Stopping Service...  
net stop NovaBackup  
  
echo [2/3] Removing Service...  
nova-service.exe remove  
  
echo [3/3] Cleaning up...  
del /F \"C:\Program Files\NovaBackup\*\" /Q  
rmdir /S /Q \"C:\Program Files\NovaBackup\"  
  
echo Uninstallation complete!  
pause  
