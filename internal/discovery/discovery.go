package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"novabackup/internal/database"
	"novabackup/internal/providers"
	"github.com/google/uuid"
)

// DiscoveryService handles scanning infrastructure nodes
type DiscoveryService struct {
	db *database.Connection
}

// NewDiscoveryService creates a new discovery service
func NewDiscoveryService(db *database.Connection) *DiscoveryService {
	return &DiscoveryService{db: db}
}

// DiscoverNode scans a specific host for virtual machines
func (d *DiscoveryService) DiscoverNode(nodeID uuid.UUID) error {
	// 1. Get node info from DB
	// Querying raw DB since we don't have GetNode method yet
	var name, address, nodeType, user, pass string
	query := `SELECT name, ip_address, node_type, username, password_encrypted FROM infrastructure_nodes WHERE id = ?`
	err := d.db.GetDB().QueryRow(query, nodeID.String()).Scan(&name, &address, &nodeType, &user, &pass)
	if err != nil {
		return fmt.Errorf("node not found: %w", err)
	}

	// 2. Run discovery based on type
	if nodeType == "Hyper-V" {
		provider := providers.NewHyperVBackupProvider(providers.HyperVConfig{
			Host:     address,
			Username: user,
			Password: pass,
		})

		vms, err := provider.ListVMs(context.Background())
		if err != nil {
			return fmt.Errorf("hyper-v scan failed: %w", err)
		}

		// 3. Save to database
		// First clear old objects
		_, _ = d.db.GetDB().Exec(`DELETE FROM infrastructure_objects WHERE node_id = ?`, nodeID.String())

		for _, vm := range vms {
			metadata, _ := json.Marshal(vm)
			err = d.db.CreateInfrastructureObject(
				uuid.New(),
				nodeID,
				vm.Name,
				"VM",
				vm.ID,
				string(metadata),
			)
			if err != nil {
				fmt.Printf("Failed to save VM %s: %v\n", vm.Name, err)
			}
		}
	}

	return nil
}
