package notifications

import (
	"fmt"
	"net/smtp"
)

type EmailConfig struct {
	SMTPHost string
	SMTPPort int
	Username string
	Password string
	From     string
	To       []string
}

func SendEmail(c EmailConfig, subject, body string) error {
	msg := []byte("Subject: " + subject + "\r\n\r\n" + body)
	addr := fmt.Sprintf("%s:%d", c.SMTPHost, c.SMTPPort)
	if c.SMTPPort == 0 {
		addr = c.SMTPHost + ":587"
	}
	auth := smtp.PlainAuth("", c.Username, c.Password, c.SMTPHost)
	return smtp.SendMail(addr, auth, c.From, c.To, msg)
}

func SendJobSuccess(c EmailConfig, jobName string) error {
	return SendEmail(c, "Backup Success", "Job "+jobName+" completed successfully")
}

func SendJobFailure(c EmailConfig, jobName string, err string) error {
	return SendEmail(c, "Backup FAILED", "Job "+jobName+" failed: "+err)
}
