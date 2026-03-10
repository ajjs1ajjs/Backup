@echo off  
  
echo ========================================  
echo   NovaBackup v6.0 - Build Installer  
echo ========================================  
echo.  
echo Looking for Inno Setup...  
  
if exist \"C:\Program Files (x86)\Inno Setup 6\ISCC.exe\" (  
    \"C:\Program Files (x86)\Inno Setup 6\ISCC.exe\" novabackup_full.iss  
) else if exist \"C:\Program Files\Inno Setup 6\ISCC.exe\" (  
    \"C:\Program Files\Inno Setup 6\ISCC.exe\" novabackup_full.iss  
) else (  
    echo Inno Setup not found!  
    echo Please download from: https://jrsoftware.org/isdl.php  
    pause  
    exit /b 1  
)  
  
echo ========================================  
echo   Build Complete!  
echo ========================================  
echo.  
echo Installer created: output\  
  
pause  
