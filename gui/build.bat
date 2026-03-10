@echo off  
  
echo Building NovaBackup GUI...  
  
dotnet restore  
dotnet build -c Release  
  
if %0% EQU 0 (  
    echo SUCCESS!  
    dir bin\Release\*.exe  
) else (  
    echo FAILED  
)  
pause  
