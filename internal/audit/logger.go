package audit

import (
	"encoding/json"
	"fmt"
	"time"
)

type AuditLog struct {
	Timestamp time.Time
	User      string
	Action    string
	Resource  string
	Details   string
}

func LogAction(user, action, resource, details string) (*AuditLog, error) {
	log := &AuditLog{
		Timestamp: time.Now(),
		User:      user,
		Action:    action,
		Resource:  resource,
		Details:   details,
	}
	data, _ := json.Marshal(log)
	fmt.Println(string(data))
	return log, nil
}

func (l *AuditLog) Save() error {
	fmt.Printf("Audit: %s by %s on %s\n", l.Action, l.User, l.Resource)
	return nil
}
