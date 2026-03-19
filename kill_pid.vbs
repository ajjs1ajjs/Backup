Set objWMIService = GetObject("winmgmts:\\.\root\cimv2")
objWMIService.Get("Win32_Process.Handle='12428'").Terminate()
Wscript.Echo "Killed process 12428!"
