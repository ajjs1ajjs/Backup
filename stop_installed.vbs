Set objWMIService = GetObject("winmgmts:\\.\root\cimv2")

' Kill process if running
Set colItems = objWMIService.ExecQuery("SELECT * FROM Win32_Process WHERE Name LIKE '%nova%' OR Name LIKE '%backup%'")
For Each objItem in colItems
    Wscript.Echo "Killing: " & objItem.Name & " (PID: " & objItem.ProcessId & ")"
    objItem.Terminate()
Next

' Try to stop Windows services
objShell.Run "sc stop NovaBackup", 0, False
objShell.Run "sc stop ""NovaBackup Service""", 0, False

Wscript.Echo "Done!"
