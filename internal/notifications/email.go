package notifications 
import \"net/smtp\" 
type EmailConfig struct{SMTPHost string;SMTPPort int;Username string;Password string;From string;To []string} 
func SendEmail(c EmailConfig,subject,body string)error{msg:=[]byte(\"Subject: \"+subject+\"\r\n\r\n\"+body);return smtp.SendMail(c.SMTPHost+\":587\",smtp.PlainAuth(\"\",c.Username,c.Password,c.SMTPHost),c.From,c.To,msg)} 
func SendJobSuccess(c EmailConfig,jobName string)error{return SendEmail(c,\"Backup Success\",\"Job \"+jobName+\" completed successfully\")} 
func SendJobFailure(c EmailConfig,jobName string,err string)error{return SendEmail(c,\"Backup FAILED\",\"Job \"+jobName+\" failed: \"+err)} 
