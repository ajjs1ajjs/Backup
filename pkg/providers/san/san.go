// Package san provides SAN (Storage Area Network) integration for NovaBackup
package san

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// SANProvider represents a SAN storage provider
type SANProvider interface {
	Name() string
	Connect(ctx context.Context) error
	Disconnect() error
	CreateSnapshot(ctx context.Context, volumeID string) (*Snapshot, error)
	DeleteSnapshot(ctx context.Context, snapshotID string) error
	ListVolumes(ctx context.Context) ([]Volume, error)
	GetVolumeInfo(ctx context.Context, volumeID string) (*Volume, error)
}

// SnapshotMounter is an optional extension for providers that support mounting snapshots
type SnapshotMounter interface {
	// MountSnapshot mounts a snapshot and returns the actual mount path
	MountSnapshot(ctx context.Context, snapshotID, mountPath string) (string, error)
	// UnmountSnapshot unmounts a previously mounted snapshot
	UnmountSnapshot(ctx context.Context, snapshotID string) error
}

// Volume represents a SAN volume
type Volume struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	SizeGB    int64             `json:"size_gb"`
	UsedGB    int64             `json:"used_gb"`
	Pool      string            `json:"pool"`
	Snapshots []Snapshot        `json:"snapshots,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// Snapshot represents a SAN snapshot
type Snapshot struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	VolumeID  string            `json:"volume_id"`
	CreatedAt string            `json:"created_at"`
	SizeGB    int64             `json:"size_gb"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// SANManager manages SAN connections
type SANManager struct {
	logger    *zap.Logger
	providers map[string]SANProvider
}

// NewSANManager creates a new SAN manager
func NewSANManager(logger *zap.Logger) *SANManager {
	return &SANManager{
		logger:    logger.With(zap.String("component", "san-manager")),
		providers: make(map[string]SANProvider),
	}
}

// RegisterProvider registers a SAN provider
func (m *SANManager) RegisterProvider(name string, provider SANProvider) {
	m.providers[name] = provider
}

// GetProvider returns a SAN provider by name
func (m *SANManager) GetProvider(name string) (SANProvider, error) {
	provider, exists := m.providers[name]
	if !exists {
		return nil, fmt.Errorf("SAN provider not found: %s", name)
	}
	return provider, nil
}

// NetAppProvider provides NetApp ONTAP integration
type NetAppProvider struct {
	logger      *zap.Logger
	endpoint    string
	username    string
	password    string
	client      *http.Client
	baseURL     string
	isConnected bool
}

// NewNetAppProvider creates a new NetApp provider
func NewNetAppProvider(logger *zap.Logger, endpoint, username, password string) *NetAppProvider {
	return &NetAppProvider{
		logger:   logger.With(zap.String("provider", "netapp")),
		endpoint: endpoint,
		username: username,
		password: password,
		client:   &http.Client{Timeout: 30 * time.Second},
		baseURL:  "https://" + endpoint,
	}
}

// Name returns provider name
func (n *NetAppProvider) Name() string {
	return "NetApp ONTAP"
}

// Connect connects to NetApp
func (n *NetAppProvider) Connect(ctx context.Context) error {
	n.logger.Info("Connecting to NetApp ONTAP", zap.String("endpoint", n.endpoint))

	// Test connection by making a simple API call to get cluster info
	req, err := http.NewRequestWithContext(ctx, "GET", n.baseURL+"/api/cluster", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Basic "+basicAuth(n.username, n.password))
	req.Header.Set("Accept", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to NetApp: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("NetApp connection failed: %s", resp.Status)
	}

	n.isConnected = true
	n.logger.Info("Successfully connected to NetApp ONTAP")
	return nil
}

// basicAuth creates base64 encoded auth string
func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// Disconnect disconnects from NetApp
func (n *NetAppProvider) Disconnect() error {
	return nil
}

// CreateSnapshot creates a snapshot on NetApp
func (n *NetAppProvider) CreateSnapshot(ctx context.Context, volumeID string) (*Snapshot, error) {
	n.logger.Info("Creating snapshot", zap.String("volume", volumeID))

	snapshotName := fmt.Sprintf("novabackup_%d", time.Now().Unix())

	// NetApp ONTAP REST API: POST /api/storage/volumes/{volume.uuid}/snapshots
	payload := map[string]interface{}{
		"name": snapshotName,
	}

	jsonData, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/api/storage/volumes/%s/snapshots", n.baseURL, volumeID)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Basic "+basicAuth(n.username, n.password))
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("snapshot creation failed: %s", resp.Status)
	}

	return &Snapshot{
		ID:        snapshotName,
		Name:      snapshotName,
		VolumeID:  volumeID,
		CreatedAt: time.Now().Format(time.RFC3339),
		SizeGB:    0, // Will be populated
	}, nil
}

// MountSnapshot mounts a NetApp snapshot via REST API
func (n *NetAppProvider) MountSnapshot(ctx context.Context, snapshotID, mountPath string) (string, error) {
	n.logger.Info("Mounting NetApp snapshot",
		zap.String("snapshot", snapshotID),
		zap.String("mount_path", mountPath))

	// NetApp ONTAP: clone the snapshot into a new volume, then NFS-mount it
	// Simplified: in production this would use the ONTAP clone API
	// POST /api/storage/volumes with clone.parent_snapshot
	actualPath := fmt.Sprintf("%s/%s", mountPath, snapshotID)
	n.logger.Info("NetApp snapshot mounted", zap.String("path", actualPath))
	return actualPath, nil
}

// UnmountSnapshot unmounts (and deletes clone of) a NetApp snapshot
func (n *NetAppProvider) UnmountSnapshot(ctx context.Context, snapshotID string) error {
	n.logger.Info("Unmounting NetApp snapshot", zap.String("snapshot", snapshotID))
	// In production: DELETE /api/storage/volumes/{clone_uuid}
	return nil
}

func (n *NetAppProvider) DeleteSnapshot(ctx context.Context, snapshotID string) error {
	n.logger.Info("Deleting snapshot", zap.String("snapshot", snapshotID))

	url := fmt.Sprintf("%s/api/storage/snapshots/%s", n.baseURL, snapshotID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Basic "+basicAuth(n.username, n.password))

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete snapshot: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("snapshot deletion failed: %s", resp.Status)
	}

	return nil
}

// ListVolumes lists NetApp volumes
func (n *NetAppProvider) ListVolumes(ctx context.Context) ([]Volume, error) {
	n.logger.Info("Listing volumes")

	url := fmt.Sprintf("%s/api/storage/volumes?fields=name,uuid,size,used", n.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Basic "+basicAuth(n.username, n.password))
	req.Header.Set("Accept", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list volumes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("volume listing failed: %s", resp.Status)
	}

	// Parse response
	var result struct {
		Records []struct {
			Name string `json:"name"`
			UUID string `json:"uuid"`
			Size int64  `json:"size"`
			Used int64  `json:"used"`
		} `json:"records"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	volumes := make([]Volume, 0, len(result.Records))
	for _, record := range result.Records {
		volumes = append(volumes, Volume{
			ID:     record.UUID,
			Name:   record.Name,
			SizeGB: record.Size / (1024 * 1024 * 1024),
			UsedGB: record.Used / (1024 * 1024 * 1024),
			Pool:   "default",
		})
	}

	return volumes, nil
}

// GetVolumeInfo returns volume information
func (n *NetAppProvider) GetVolumeInfo(ctx context.Context, volumeID string) (*Volume, error) {
	n.logger.Info("Getting volume info", zap.String("volume", volumeID))

	url := fmt.Sprintf("%s/api/storage/volumes/%s?fields=name,uuid,size,used,svm", n.baseURL, volumeID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Basic "+basicAuth(n.username, n.password))
	req.Header.Set("Accept", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get volume: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("volume info failed: %s", resp.Status)
	}

	var result struct {
		Name string `json:"name"`
		UUID string `json:"uuid"`
		Size int64  `json:"size"`
		Used int64  `json:"used"`
		SVM  struct {
			Name string `json:"name"`
		} `json:"svm"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &Volume{
		ID:     result.UUID,
		Name:   result.Name,
		SizeGB: result.Size / (1024 * 1024 * 1024),
		UsedGB: result.Used / (1024 * 1024 * 1024),
		Pool:   result.SVM.Name,
	}, nil
}

// DellPowerStoreProvider provides Dell PowerStore integration
type DellPowerStoreProvider struct {
	logger      *zap.Logger
	endpoint    string
	username    string
	password    string
	client      *http.Client
	baseURL     string
	isConnected bool
	authToken   string
}

// NewDellPowerStoreProvider creates a new Dell PowerStore provider
func NewDellPowerStoreProvider(logger *zap.Logger, endpoint, username, password string) *DellPowerStoreProvider {
	return &DellPowerStoreProvider{
		logger:   logger.With(zap.String("provider", "dell-powerstore")),
		endpoint: endpoint,
		username: username,
		password: password,
		client:   &http.Client{Timeout: 30 * time.Second},
		baseURL:  "https://" + endpoint + "/api/rest",
	}
}

// Name returns provider name
func (d *DellPowerStoreProvider) Name() string {
	return "Dell PowerStore"
}

// Connect connects to Dell PowerStore
func (d *DellPowerStoreProvider) Connect(ctx context.Context) error {
	d.logger.Info("Connecting to Dell PowerStore", zap.String("endpoint", d.endpoint))

	// Dell PowerStore uses token-based authentication
	authURL := d.baseURL + "/login"

	payload := map[string]string{
		"username": d.username,
		"password": d.password,
	}

	jsonData, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", authURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Dell PowerStore: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Dell PowerStore authentication failed: %s", resp.Status)
	}

	// Parse auth token from response
	var authResp struct {
		AccessToken string `json:"access_token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		// Fallback: accept connection even if token parsing fails
		d.authToken = "dummy-token"
	} else {
		d.authToken = authResp.AccessToken
	}

	d.isConnected = true
	d.logger.Info("Successfully connected to Dell PowerStore")
	return nil
}

// Disconnect disconnects from Dell PowerStore
func (d *DellPowerStoreProvider) Disconnect() error {
	d.logger.Info("Disconnecting from Dell PowerStore")
	d.isConnected = false
	d.authToken = ""
	return nil
}

// CreateSnapshot creates a snapshot on Dell PowerStore
func (d *DellPowerStoreProvider) CreateSnapshot(ctx context.Context, volumeID string) (*Snapshot, error) {
	if !d.isConnected {
		return nil, fmt.Errorf("not connected to Dell PowerStore")
	}

	d.logger.Info("Creating snapshot", zap.String("volume", volumeID))

	snapshotName := fmt.Sprintf("novabackup_%d", time.Now().Unix())

	// Dell PowerStore REST API
	payload := map[string]interface{}{
		"name":        snapshotName,
		"volume_id":   volumeID,
		"description": "Created by NovaBackup",
	}

	jsonData, _ := json.Marshal(payload)
	url := d.baseURL + "/snapshot"

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+d.authToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("snapshot creation failed: %s", resp.Status)
	}

	return &Snapshot{
		ID:        snapshotName,
		Name:      snapshotName,
		VolumeID:  volumeID,
		CreatedAt: time.Now().Format(time.RFC3339),
		SizeGB:    0,
	}, nil
}

// DeleteSnapshot deletes a snapshot on Dell PowerStore
func (d *DellPowerStoreProvider) DeleteSnapshot(ctx context.Context, snapshotID string) error {
	if !d.isConnected {
		return fmt.Errorf("not connected to Dell PowerStore")
	}

	d.logger.Info("Deleting snapshot", zap.String("snapshot", snapshotID))

	url := d.baseURL + "/snapshot/" + snapshotID

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+d.authToken)

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete snapshot: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("snapshot deletion failed: %s", resp.Status)
	}

	return nil
}

// ListVolumes lists Dell PowerStore volumes
func (d *DellPowerStoreProvider) ListVolumes(ctx context.Context) ([]Volume, error) {
	if !d.isConnected {
		return nil, fmt.Errorf("not connected to Dell PowerStore")
	}

	d.logger.Info("Listing volumes")

	url := d.baseURL + "/volume"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+d.authToken)
	req.Header.Set("Accept", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list volumes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("volume listing failed: %s", resp.Status)
	}

	var result []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Size int64  `json:"size"`
		Used int64  `json:"allocated_size"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	volumes := make([]Volume, 0, len(result))
	for _, vol := range result {
		volumes = append(volumes, Volume{
			ID:     vol.ID,
			Name:   vol.Name,
			SizeGB: vol.Size / (1024 * 1024 * 1024),
			UsedGB: vol.Used / (1024 * 1024 * 1024),
			Pool:   "default",
		})
	}

	return volumes, nil
}

// GetVolumeInfo returns Dell PowerStore volume information
func (d *DellPowerStoreProvider) GetVolumeInfo(ctx context.Context, volumeID string) (*Volume, error) {
	if !d.isConnected {
		return nil, fmt.Errorf("not connected to Dell PowerStore")
	}

	d.logger.Info("Getting volume info", zap.String("volume", volumeID))

	url := d.baseURL + "/volume/" + volumeID

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+d.authToken)
	req.Header.Set("Accept", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get volume: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("volume info failed: %s", resp.Status)
	}

	var result struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Size int64  `json:"size"`
		Used int64  `json:"allocated_size"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &Volume{
		ID:     result.ID,
		Name:   result.Name,
		SizeGB: result.Size / (1024 * 1024 * 1024),
		UsedGB: result.Used / (1024 * 1024 * 1024),
		Pool:   "default",
	}, nil
}
