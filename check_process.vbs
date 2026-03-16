Set objWMIService = GetObject("winmgmts:\\.\root\cimv2")
Set colItems = objWMIService.ExecQuery("SELECT * FROM Win32_Process WHERE ProcessId = '12428'")
For Each objItem in colItems
    Wscript.Echo "Process Name: " & objItem.Name
    Wscript.Echo "Executable: " & objItem.ExecutablePath
    Wscript.Echo "Command Line: " & objItem.CommandLine
    Wscript.Echo "Started by User: " & objItem.GetOwner()
Next
