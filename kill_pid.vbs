' Kill process by PID
' Usage: cscript kill_pid.vbs [PID]

Dim objWMIService, colItems, objItem
Dim targetPID

If WScript.Arguments.Count > 0 Then
    targetPID = WScript.Arguments(0)
Else
    Wscript.Echo "Usage: cscript kill_pid.vbs [PID]"
    Wscript.Echo "Example: cscript kill_pid.vbs 12345"
    WScript.Quit(1)
End If

Set objWMIService = GetObject("winmgmts:\\.\root\cimv2")
Set colItems = objWMIService.ExecQuery("SELECT * FROM Win32_Process WHERE ProcessId = '" & targetPID & "'")

If colItems.Count > 0 Then
    For Each objItem in colItems
        Wscript.Echo "Killing: " & objItem.Name & " (PID: " & targetPID & ")"
        objItem.Terminate()
    Next
    Wscript.Echo "Done!"
Else
    Wscript.Echo "Process " & targetPID & " not found"
End If
