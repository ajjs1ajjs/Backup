package vss

import (
	"context"
	"fmt"
	"time"
)

type SQLServerVSS struct {
	manager *InMemoryVSSManager
}

func NewSQLServerVSS() *SQLServerVSS {
	return &SQLServerVSS{
		manager: NewInMemoryVSSManager(),
	}
}

func (s *SQLServerVSS) CreateSnapshot(ctx context.Context, databases []string) (*VSSResult, error) {
	req := &VSSRequest{
		Volume:      "C:",
		WriterTypes: []VSSWriterType{VSSWriterSQLServer},
		BackupType:  "Full",
		Options: map[string]string{
			"databases": fmt.Sprintf("%v", databases),
		},
	}
	return s.manager.CreateSnapshot(ctx, req)
}

func (s *SQLServerVSS) GetDatabases(ctx context.Context) ([]string, error) {
	return []string{
		"master",
		"msdb",
		"model",
		"tempdb",
		"NovaBackupDB",
	}, nil
}

func (s *SQLServerVSS) GetStatus(ctx context.Context) (*VSSWriterStatus, error) {
	return &VSSWriterStatus{
		WriterType: VSSWriterSQLServer,
		Name:       "SQLServerWriter",
		State:      "Stable",
		LastSeen:   time.Now(),
	}, nil
}

type ExchangeVSS struct {
	manager *InMemoryVSSManager
}

func NewExchangeVSS() *ExchangeVSS {
	return &ExchangeVSS{
		manager: NewInMemoryVSSManager(),
	}
}

func (e *ExchangeVSS) CreateSnapshot(ctx context.Context, mailboxes []string) (*VSSResult, error) {
	req := &VSSRequest{
		Volume:      "C:",
		WriterTypes: []VSSWriterType{VSSWriterExchange},
		BackupType:  "Full",
		Options: map[string]string{
			"mailboxes": fmt.Sprintf("%v", mailboxes),
		},
	}
	return e.manager.CreateSnapshot(ctx, req)
}

func (e *ExchangeVSS) GetMailboxes(ctx context.Context) ([]string, error) {
	return []string{
		"administrator",
		"backup-service",
		"info@company.com",
	}, nil
}

func (e *ExchangeVSS) GetStatus(ctx context.Context) (*VSSWriterStatus, error) {
	return &VSSWriterStatus{
		WriterType: VSSWriterExchange,
		Name:       "Microsoft Exchange Writer",
		State:      "Stable",
		LastSeen:   time.Now(),
	}, nil
}

type ActiveDirectoryVSS struct {
	manager *InMemoryVSSManager
}

func NewActiveDirectoryVSS() *ActiveDirectoryVSS {
	return &ActiveDirectoryVSS{
		manager: NewInMemoryVSSManager(),
	}
}

func (a *ActiveDirectoryVSS) CreateSnapshot(ctx context.Context) (*VSSResult, error) {
	req := &VSSRequest{
		Volume:      "C:",
		WriterTypes: []VSSWriterType{VSSWriterActiveDir},
		BackupType:  "Full",
		Options: map[string]string{
			"includesystem": "true",
		},
	}
	return a.manager.CreateSnapshot(ctx, req)
}

func (a *ActiveDirectoryVSS) GetStatus(ctx context.Context) (*VSSWriterStatus, error) {
	return &VSSWriterStatus{
		WriterType: VSSWriterActiveDir,
		Name:       "NTDS",
		State:      "Stable",
		LastSeen:   time.Now(),
	}, nil
}

type GuestCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Domain   string `json:"domain"`
	Type     string `json:"type"` // "windows", "linux"
}

type GuestCredentialManager struct {
	credentials map[string]*GuestCredentials
}

func NewGuestCredentialManager() *GuestCredentialManager {
	return &GuestCredentialManager{
		credentials: make(map[string]*GuestCredentials),
	}
}

func (m *GuestCredentialManager) AddCredential(id string, cred *GuestCredentials) {
	m.credentials[id] = cred
}

func (m *GuestCredentialManager) GetCredential(id string) (*GuestCredentials, bool) {
	cred, ok := m.credentials[id]
	return cred, ok
}

func (m *GuestCredentialManager) DeleteCredential(id string) {
	delete(m.credentials, id)
}

func (m *GuestCredentialManager) ListCredentials() []*GuestCredentials {
	var creds []*GuestCredentials
	for _, c := range m.credentials {
		creds = append(creds, c)
	}
	return creds
}
