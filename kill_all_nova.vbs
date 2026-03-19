Set objWMIService = GetObject("winmgmts:\\.\root\cimv2")
Set colItems = objWMIService.ExecQuery("SELECT * FROM Win32_Process WHERE Name LIKE '%nova%' OR Name LIKE '%backup%'")
For Each objItem in colItems
    Wscript.Echo "Killing: " & objItem.Name & " (PID: " & objItem.ProcessId & ")"
    objItem.Terminate()
Next
Wscript.Echo "Done!"
