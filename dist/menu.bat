@echo off 
 
echo ======================================== 
echo   NovaBackup v6.0 - Control Panel 
echo ======================================== 
echo. 
echo [1] Backup Files 
echo [2] Create Backup Job 
echo [3] List Jobs 
echo [4] Restore Files 
echo [5] Start API Server 
echo [6] Install Service 
echo [7] Exit 
echo. 
set /p choice=Enter choice:  
if \"%%choice%%\"==\"1\" goto backup 
if \"%%choice%%\"==\"2\" goto createjob 
if \"%%choice%%\"==\"3\" goto listjobs 
if \"%%choice%%\"==\"4\" goto restore 
if \"%%choice%%\"==\"5\" goto api 
if \"%%choice%%\"==\"6\" goto service 
if \"%%choice%%\"==\"7\" goto end 
goto menu 
 
:backup 
nova.exe backup run -s C:\Data -d D:\Backups -c 
pause 
goto menu 
 
:createjob 
set /p name=Job Name:  
set /p source=Source:  
set /p dest=Destination:  
nova.exe backup create -n %%name%% -s %%source%% -d %%dest%% 
pause 
goto menu 
 
:listjobs 
nova.exe backup list 
pause 
goto menu 
 
:restore 
set /p source=Backup ID:  
set /p dest=Destination:  
nova.exe restore files -s %%source%% -d %%dest%% 
pause 
goto menu 
 
:api 
start nova.exe api start 
goto menu 
 
:service 
nova.exe service install 
pause 
goto menu 
 
:end 
exit 
