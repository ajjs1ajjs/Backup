@echo off 
echo ========================================  
echo   NovaBackup v6.0 Installation        
echo ========================================  
echo.  
echo IMPORTANT: Run as Administrator!  
echo Right-click this file -> Run as Administrator  
echo.  
pause  
  
echo [1/3] Installing to Program Files...  
mkdir \"C:\Program Files\NovaBackup\"  
copy /Y nova.exe \"C:\Program Files\NovaBackup\\\"  
  
echo [2/3] Installing Windows Service...  
\"C:\Program Files\NovaBackup\nova.exe\" service install  
  
echo [3/3] Starting Service...  
net start NovaBackup  
  
echo ========================================  
echo   Installation Complete!  
echo ========================================  
echo.  
pause  
