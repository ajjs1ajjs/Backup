Set objWMIService = GetObject("winmgmts:\\.\root\cimv2")
Set colItems = objWMIService.ExecQuery("SELECT * FROM Win32_Process WHERE ProcessId = '12428'")

For Each objItem in colItems
    Wscript.Echo "Process: " & objItem.Name
    Wscript.Echo "PID: " & objItem.ProcessId
    Wscript.Echo "Executable: " & objItem.ExecutablePath

    ' Get owner
    strOwner = objItem.GetOwner()
    Wscript.Echo "Owner: " & strOwner

    ' Check if we can terminate
    Wscript.Echo "Attempting termination..."
    intReturn = objItem.Terminate()
    Wscript.Echo "Return code: " & intReturn

    Select Case intReturn
        Case 0
            Wscript.Echo "Result: SUCCESS"
        Case 2
            Wscript.Echo "Result: ACCESS DENIED - Need administrator privileges"
        Case 8
            Wscript.Echo "Result: OUT OF MEMORY"
        Case 9
            Wscript.Echo "Result: INVALID PATH"
        Case 10
            Wscript.Echo "Result: INVALID PARAMETER"
        Case Else
            Wscript.Echo "Result: UNKNOWN ERROR"
    End Select
Next
