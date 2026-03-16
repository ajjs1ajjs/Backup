Set objWMIService = GetObject("winmgmts:\\.\root\cimv2")
Set colItems = objWMIService.ExecQuery("SELECT * FROM Win32_Process WHERE Name = 'nova-backup.exe'")
For Each objItem in colItems
    objItem.Terminate()
    Wscript.Echo "Killed process: " & objItem.ProcessId
Next
Wscript.Echo "Done!"
