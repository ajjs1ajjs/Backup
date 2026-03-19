Set objWMIService = GetObject("winmgmts:\\.\root\cimv2")

Wscript.Echo "Stopping all NovaBackup processes..."

' Kill all nova/backup processes
Set colItems = objWMIService.ExecQuery("SELECT * FROM Win32_Process WHERE Name LIKE '%nova%' OR Name LIKE '%backup%' OR Name LIKE '%novabackup%'")
For Each objItem in colItems
    Wscript.Echo "Killing: " & objItem.Name & " (PID: " & objItem.ProcessId & ") - Path: " & objItem.ExecutablePath
    objItem.Terminate()
Next

Wscript.Sleep 2000

' Verify all killed
Set colItems = objWMIService.ExecQuery("SELECT * FROM Win32_Process WHERE Name LIKE '%nova%' OR Name LIKE '%backup%' OR Name LIKE '%novabackup%'")
If colItems.Count = 0 Then
    Wscript.Echo "SUCCESS: All NovaBackup processes stopped!"
Else
    Wscript.Echo "WARNING: Some processes still running:"
    For Each objItem in colItems
        Wscript.Echo "  - " & objItem.Name & " (PID: " & objItem.ProcessId & ")"
    Next
End If
