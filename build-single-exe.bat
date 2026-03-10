@echo off  
  
echo Building NovaBackup Single EXE...  
  
REM Build CLI with service support  
set GOOS=windows  
set GOARCH=amd64  
  
go build -o nova.exe ./cmd/nova-cli/  
  
if %0% EQU 0 (  
    echo SUCCESS: nova.exe created  
    dir nova.exe  
) else (  
    echo FAILED: Build error  
)  
pause  
