' Check NovaBackup Process Owner
' Usage: cscript check_owner.vbs [PID]
' If no PID specified, finds process by name

Dim objWMIService, colItems, objItem
Dim targetPID

' Get PID from argument or find by process name
If WScript.Arguments.Count > 0 Then
    targetPID = WScript.Arguments(0)
Else
    ' Find process by name instead of hardcoded PID
    Set objWMIService = GetObject("winmgmts:\\.\root\cimv2")
    Set colItems = objWMIService.ExecQuery("SELECT * FROM Win32_Process WHERE Name LIKE '%nova%'")
    
    For Each objItem in colItems
        Wscript.Echo "Process: " & objItem.Name & " (PID: " & objItem.ProcessId & ")"
        Wscript.Echo "Owner: " & GetProcessOwner(objItem.ProcessId)
    Next
    
    If colItems.Count = 0 Then
        Wscript.Echo "No NovaBackup process found"
    End If
    WScript.Quit(0)
End If

' Check specific PID
Set objWMIService = GetObject("winmgmts:\\.\root\cimv2")
Set colItems = objWMIService.ExecQuery("SELECT * FROM Win32_Process WHERE ProcessId = '" & targetPID & "'")

If colItems.Count > 0 Then
    For Each objItem in colItems
        Wscript.Echo "Process: " & objItem.Name & " (PID: " & targetPID & ")"
        Wscript.Echo "Owner: " & GetProcessOwner(targetPID)
    Next
Else
    Wscript.Echo "Process " & targetPID & " not found"
End If

Function GetProcessOwner(pid)
    On Error Resume Next
    Set objWMIService = GetObject("winmgmts:\\.\root\cimv2")
    Set colItems = objWMIService.ExecQuery("SELECT * FROM Win32_Process WHERE ProcessId = '" & pid & "'")
    
    For Each objItem in colItems
        Dim arrUser, strDomain, strName
        If objItem.GetOwner(strName, strDomain) = 0 Then
            GetProcessOwner = strDomain & "\" & strName
        Else
            GetProcessOwner = "Unknown"
        End If
        Exit Function
    Next
    GetProcessOwner = "Not found"
End Function
