package tape

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type TapeLibrary struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Model      string `json:"model"`
	Slots      int    `json:"slots"`
	DriveCount int    `json:"drive_count"`
	Status     string `json:"status"` // "Online", "Offline", "Error"
}

type TapeDrive struct {
	ID          string `json:"id"`
	LibraryID   string `json:"library_id"`
	DriveNumber int    `json:"drive_number"`
	Status      string `json:"status"` // "Idle", "Reading", "Writing", "Error"
	CurrentTape string `json:"current_tape"`
}

type TapeCartridge struct {
	ID         string     `json:"id"`
	Barcode    string     `json:"barcode"`
	Slot       int        `json:"slot"`
	MediaType  string     `json:"media_type"` // "LTO-8", "LTO-9", "LTO-10"
	CapacityGB int64      `json:"capacity_gb"`
	UsedGB     int64      `json:"used_gb"`
	Status     string     `json:"status"` // "Free", "InUse", "Archived", "Expired"
	Label      string     `json:"label"`
	CreatedAt  time.Time  `json:"created_at"`
	ArchivedAt *time.Time `json:"archived_at,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

type TapeVault struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	Contact     string    `json:"contact"`
	CreatedAt   time.Time `json:"created_at"`
}

type TapeBackupJob struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Source      string    `json:"source"`
	TargetVault string    `json:"target_vault"`
	Schedule    string    `json:"schedule"` // "Daily", "Weekly", "Monthly"
	Retention   int       `json:"retention_days"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
}

type TapeManager interface {
	// Libraries
	ListLibraries(ctx context.Context) ([]TapeLibrary, error)
	GetLibrary(ctx context.Context, id string) (*TapeLibrary, error)

	// Drives
	ListDrives(ctx context.Context, libraryID string) ([]TapeDrive, error)
	GetDriveStatus(ctx context.Context, driveID string) (*TapeDrive, error)

	// Cartridges
	ListCartridges(ctx context.Context) ([]TapeCartridge, error)
	GetCartridge(ctx context.Context, id string) (*TapeCartridge, error)
	Inventory(ctx context.Context, libraryID string) ([]TapeCartridge, error)
	EjectTape(ctx context.Context, cartridgeID string) error
	ImportTape(ctx context.Context, libraryID string, slot int) error

	// Vaults
	ListVaults(ctx context.Context) ([]TapeVault, error)
	CreateVault(ctx context.Context, vault *TapeVault) error
	DeleteVault(ctx context.Context, id string) error

	// Jobs
	ListJobs(ctx context.Context) ([]TapeBackupJob, error)
	CreateJob(ctx context.Context, job *TapeBackupJob) error
	DeleteJob(ctx context.Context, id string) error
	RunJob(ctx context.Context, id string) error
}

type InMemoryTapeManager struct {
	libraries  map[string]*TapeLibrary
	drives     map[string]*TapeDrive
	cartridges map[string]*TapeCartridge
	vaults     map[string]*TapeVault
	jobs       map[string]*TapeBackupJob
	mu         sync.RWMutex
}

func NewInMemoryTapeManager() *InMemoryTapeManager {
	m := &InMemoryTapeManager{
		libraries:  make(map[string]*TapeLibrary),
		drives:     make(map[string]*TapeDrive),
		cartridges: make(map[string]*TapeCartridge),
		vaults:     make(map[string]*TapeVault),
		jobs:       make(map[string]*TapeBackupJob),
	}
	m.initDemoData()
	return m
}

func (m *InMemoryTapeManager) initDemoData() {
	m.libraries["lib-1"] = &TapeLibrary{
		ID: "lib-1", Name: "TL4000", Model: "IBM TS3400", Slots: 48, DriveCount: 2, Status: "Online",
	}
	m.libraries["lib-2"] = &TapeLibrary{
		ID: "lib-2", Name: "Dell TL2000", Model: "Dell PowerVault TL2000", Slots: 24, DriveCount: 1, Status: "Online",
	}

	m.drives["drive-1"] = &TapeDrive{ID: "drive-1", LibraryID: "lib-1", DriveNumber: 0, Status: "Idle"}
	m.drives["drive-2"] = &TapeDrive{ID: "drive-2", LibraryID: "lib-1", DriveNumber: 1, Status: "Idle"}

	for i := 1; i <= 10; i++ {
		m.cartridges[fmt.Sprintf("tape-%d", i)] = &TapeCartridge{
			ID: fmt.Sprintf("tape-%d", i), Barcode: fmt.Sprintf("LTO%d%04d", 9, i),
			Slot: i, MediaType: "LTO-9", CapacityGB: 30000, UsedGB: int64(i * 1000),
			Status: "Free", Label: fmt.Sprintf("Tape-%d", i), CreatedAt: time.Now().Add(-time.Duration(i) * 24 * time.Hour),
		}
	}

	m.vaults["vault-1"] = &TapeVault{
		ID: "vault-1", Name: "Offsite Vault", Description: "Offsite tape storage",
		Location: "Building B", Contact: "tape-admin@company.com", CreatedAt: time.Now(),
	}

	m.jobs["job-1"] = &TapeBackupJob{
		ID: "job-1", Name: "Weekly Archive", Source: "Backup-Job-1", TargetVault: "vault-1",
		Schedule: "Weekly", Retention: 365, Enabled: true, CreatedAt: time.Now(),
	}
}

func (m *InMemoryTapeManager) ListLibraries(ctx context.Context) ([]TapeLibrary, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var libs []TapeLibrary
	for _, l := range m.libraries {
		libs = append(libs, *l)
	}
	return libs, nil
}

func (m *InMemoryTapeManager) GetLibrary(ctx context.Context, id string) (*TapeLibrary, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if lib, ok := m.libraries[id]; ok {
		return lib, nil
	}
	return nil, fmt.Errorf("library %s not found", id)
}

func (m *InMemoryTapeManager) ListDrives(ctx context.Context, libraryID string) ([]TapeDrive, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var drives []TapeDrive
	for _, d := range m.drives {
		if d.LibraryID == libraryID {
			drives = append(drives, *d)
		}
	}
	return drives, nil
}

func (m *InMemoryTapeManager) GetDriveStatus(ctx context.Context, driveID string) (*TapeDrive, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if d, ok := m.drives[driveID]; ok {
		return d, nil
	}
	return nil, fmt.Errorf("drive %s not found", driveID)
}

func (m *InMemoryTapeManager) ListCartridges(ctx context.Context) ([]TapeCartridge, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var tapes []TapeCartridge
	for _, t := range m.cartridges {
		tapes = append(tapes, *t)
	}
	return tapes, nil
}

func (m *InMemoryTapeManager) GetCartridge(ctx context.Context, id string) (*TapeCartridge, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if t, ok := m.cartridges[id]; ok {
		return t, nil
	}
	return nil, fmt.Errorf("cartridge %s not found", id)
}

func (m *InMemoryTapeManager) Inventory(ctx context.Context, libraryID string) ([]TapeCartridge, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var tapes []TapeCartridge
	for _, t := range m.cartridges {
		tapes = append(tapes, *t)
	}
	return tapes, nil
}

func (m *InMemoryTapeManager) EjectTape(ctx context.Context, cartridgeID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if t, ok := m.cartridges[cartridgeID]; ok {
		t.Status = "Free"
		t.Slot = 0
		return nil
	}
	return fmt.Errorf("cartridge %s not found", cartridgeID)
}

func (m *InMemoryTapeManager) ImportTape(ctx context.Context, libraryID string, slot int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return nil
}

func (m *InMemoryTapeManager) ListVaults(ctx context.Context) ([]TapeVault, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var vaults []TapeVault
	for _, v := range m.vaults {
		vaults = append(vaults, *v)
	}
	return vaults, nil
}

func (m *InMemoryTapeManager) CreateVault(ctx context.Context, vault *TapeVault) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	vault.ID = fmt.Sprintf("vault-%d", len(m.vaults)+1)
	vault.CreatedAt = time.Now()
	m.vaults[vault.ID] = vault
	return nil
}

func (m *InMemoryTapeManager) DeleteVault(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.vaults, id)
	return nil
}

func (m *InMemoryTapeManager) ListJobs(ctx context.Context) ([]TapeBackupJob, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var jobs []TapeBackupJob
	for _, j := range m.jobs {
		jobs = append(jobs, *j)
	}
	return jobs, nil
}

func (m *InMemoryTapeManager) CreateJob(ctx context.Context, job *TapeBackupJob) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	job.ID = fmt.Sprintf("job-%d", len(m.jobs)+1)
	job.CreatedAt = time.Now()
	m.jobs[job.ID] = job
	return nil
}

func (m *InMemoryTapeManager) DeleteJob(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.jobs, id)
	return nil
}

func (m *InMemoryTapeManager) RunJob(ctx context.Context, id string) error {
	return nil
}
