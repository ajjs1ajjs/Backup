' Check if NovaBackup process is running
' Usage: cscript check_process.vbs

Dim objWMIService, colItems, objItem

Set objWMIService = GetObject("winmgmts:\\.\root\cimv2")

' Find all nova processes (not hardcoded PID)
Set colItems = objWMIService.ExecQuery("SELECT * FROM Win32_Process WHERE Name LIKE '%nova%'")

If colItems.Count > 0 Then
    Wscript.Echo "NovaBackup processes found:"
    For Each objItem in colItems
        Wscript.Echo "  - " & objItem.Name & " (PID: " & objItem.ProcessId & ")"
        Wscript.Echo "    Path: " & objItem.ExecutablePath
        Wscript.Echo "    Memory: " & FormatNumber(objItem.WorkingSetSize / 1024 / 1024, 2) & " MB"
    Next
Else
    Wscript.Echo "No NovaBackup process running"
End If
