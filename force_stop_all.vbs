Set objWMIService = GetObject("winmgmts:\\.\root\cimv2")

Wscript.Echo "=== FORCE STOP ALL NOVABACKUP ==="

' First try to stop via WMIC
Set colItems = objWMIService.ExecQuery("SELECT * FROM Win32_Process WHERE Name LIKE '%nova%' OR Name LIKE '%backup%'")
For Each objItem in colItems
    Wscript.Echo "Attempting to kill: " & objItem.Name & " (PID: " & objItem.ProcessId & ")"

    ' Try terminate method
    intReturn = objItem.Terminate()
    If intReturn = 0 Then
        Wscript.Echo "  -> Terminated successfully"
    Else
        Wscript.Echo "  -> Terminate failed with code: " & intReturn

        ' Try harder - use WMIC directly
        Set objShell = CreateObject("WScript.Shell")
        strCommand = "wmic process where ProcessId='" & objItem.ProcessId & "' call terminate"
        objShell.Run strCommand, 0, True
        Wscript.Sleep 1000

        ' Check if still running
        Set colCheck = objWMIService.ExecQuery("SELECT * FROM Win32_Process WHERE ProcessId = '" & objItem.ProcessId & "'")
        If colCheck.Count = 0 Then
            Wscript.Echo "  -> Killed via WMIC"
        Else
            Wscript.Echo "  -> STILL RUNNING! Manual intervention required."
        End If
    End If
Next

Wscript.Sleep 3000

' Final check
Set colItems = objWMIService.ExecQuery("SELECT * FROM Win32_Process WHERE Name LIKE '%nova%' OR Name LIKE '%backup%'")
If colItems.Count = 0 Then
    Wscript.Echo ""
    Wscript.Echo "✅ SUCCESS: All processes stopped!"
Else
    Wscript.Echo ""
    Wscript.Echo "❌ WARNING: Processes still running:"
    For Each objItem in colItems
        Wscript.Echo "   PID: " & objItem.ProcessId & " - " & objItem.Name
    Next
End If
