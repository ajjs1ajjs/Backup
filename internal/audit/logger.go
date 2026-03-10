package audit 
import \"time\" 
import \"encoding/json\" 
type AuditLog struct{Timestamp time.Time;User string;Action string;Resource string;Details string} 
func LogAction(user,action,resource,details string)(*AuditLog,error){log:=&AuditLog{Timestamp:time.Now(),User:user,Action:action,Resource:resource,Details:details};data,_:=json.Marshal(log);println(string(data));return log,nil} 
func(l*AuditLog)Save()error{println(\"Audit:\",l.Action,\"by\",l.User,\"on\",l.Resource);return nil} 
