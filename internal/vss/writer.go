package vss

import (
	"context"
	"fmt"
	"time"
)

type VSSWriterType string

const (
	VSSWriterSQLServer  VSSWriterType = "SQLServerWriter"
	VSSWriterExchange   VSSWriterType = "ExchangeWriter"
	VSSWriterActiveDir  VSSWriterType = "NTDS"
	VSSWriterSharePoint VSSWriterType = "SharePoint"
	VSSWriterHyperV     VSSWriterType = "Hyper-V"
)

type VSSSnapshot struct {
	ID           string            `json:"id"`
	WriterType   VSSWriterType     `json:"writer_type"`
	Volume       string            `json:"volume"`
	SnapshotID   string            `json:"snapshot_id"`
	CreationTime time.Time         `json:"creation_time"`
	Metadata     map[string]string `json:"metadata"`
}

type VSSWriterStatus struct {
	WriterType VSSWriterType `json:"writer_type"`
	Name       string        `json:"name"`
	State      string        `json:"state"` // "Stable", "Failed", "WaitingForCompletion"
	Error      string        `json:"error,omitempty"`
	LastSeen   time.Time     `json:"last_seen"`
}

type VSSRequest struct {
	Volume      string            `json:"volume" binding:"required"`
	WriterTypes []VSSWriterType   `json:"writer_types"`
	BackupType  string            `json:"backup_type"` // "Full", "Incremental", "Copy"
	TargetPath  string            `json:"target_path"`
	Options     map[string]string `json:"options"`
}

type VSSResult struct {
	SnapshotID   string            `json:"snapshot_id"`
	Success      bool              `json:"success"`
	WriterStatus []VSSWriterStatus `json:"writer_status"`
	Error        string            `json:"error,omitempty"`
	Duration     time.Duration     `json:"duration"`
}

type VSSManager interface {
	CreateSnapshot(ctx context.Context, req *VSSRequest) (*VSSResult, error)
	DeleteSnapshot(ctx context.Context, snapshotID string) error
	ListSnapshots(ctx context.Context, volume string) ([]VSSSnapshot, error)
	GetWriterStatus(ctx context.Context) ([]VSSWriterStatus, error)
}

type InMemoryVSSManager struct {
	snapshots map[string]*VSSSnapshot
}

func NewInMemoryVSSManager() *InMemoryVSSManager {
	return &InMemoryVSSManager{
		snapshots: make(map[string]*VSSSnapshot),
	}
}

func (m *InMemoryVSSManager) CreateSnapshot(ctx context.Context, req *VSSRequest) (*VSSResult, error) {
	result := &VSSResult{
		Success:      true,
		WriterStatus: []VSSWriterStatus{},
	}

	for _, wt := range req.WriterTypes {
		status := VSSWriterStatus{
			WriterType: wt,
			Name:       string(wt),
			State:      "Stable",
			LastSeen:   time.Now(),
		}
		result.WriterStatus = append(result.WriterStatus, status)
	}

	snapshot := &VSSSnapshot{
		ID:           fmt.Sprintf("snap-%d", time.Now().Unix()),
		Volume:       req.Volume,
		SnapshotID:   result.SnapshotID,
		CreationTime: time.Now(),
		Metadata:     req.Options,
	}

	m.snapshots[snapshot.ID] = snapshot
	result.SnapshotID = snapshot.ID
	result.Duration = time.Second

	return result, nil
}

func (m *InMemoryVSSManager) DeleteSnapshot(ctx context.Context, snapshotID string) error {
	delete(m.snapshots, snapshotID)
	return nil
}

func (m *InMemoryVSSManager) ListSnapshots(ctx context.Context, volume string) ([]VSSSnapshot, error) {
	var snaps []VSSSnapshot
	for _, s := range m.snapshots {
		if s.Volume == volume || volume == "" {
			snaps = append(snaps, *s)
		}
	}
	return snaps, nil
}

func (m *InMemoryVSSManager) GetWriterStatus(ctx context.Context) ([]VSSWriterStatus, error) {
	return []VSSWriterStatus{
		{WriterType: VSSWriterSQLServer, Name: "SQLServerWriter", State: "Stable", LastSeen: time.Now()},
		{WriterType: VSSWriterExchange, Name: "Microsoft Exchange Writer", State: "Stable", LastSeen: time.Now()},
		{WriterType: VSSWriterActiveDir, Name: "NTDS", State: "Stable", LastSeen: time.Now()},
		{WriterType: VSSWriterSharePoint, Name: "SharePoint Services VSS Writer", State: "Stable", LastSeen: time.Now()},
		{WriterType: VSSWriterHyperV, Name: "Microsoft Hyper-V VSS Writer", State: "Stable", LastSeen: time.Now()},
	}, nil
}
