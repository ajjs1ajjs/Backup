package guest

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"sync"
	"time"

	"go.uber.org/zap"
)

// GuestCredential represents credentials for accessing guest operating systems
type GuestCredential struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Username    string            `json:"username"`
	Password    string            `json:"password,omitempty"`
	Domain      string            `json:"domain,omitempty"`
	Type        string            `json:"type"` // "windows", "linux", "vmware", "hyperv"
	Description string            `json:"description,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	LastUsedAt  time.Time         `json:"last_used_at,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Encrypted   bool              `json:"encrypted"`
}

// CredentialStore defines the interface for credential storage
type CredentialStore interface {
	Store(cred *GuestCredential) error
	Get(id string) (*GuestCredential, error)
	List() []*GuestCredential
	Delete(id string) error
	Update(cred *GuestCredential) error
	FindByName(name string) []*GuestCredential
	Close() error
}

// InMemoryCredentialStore provides in-memory credential storage
type InMemoryCredentialStore struct {
	mu          sync.RWMutex
	credentials map[string]*GuestCredential
	logger      *zap.Logger
}

// InMemoryCredentialStoreConfig holds configuration for the store
type InMemoryCredentialStoreConfig struct {
	Logger *zap.Logger `json:"-"`
}

// NewInMemoryCredentialStore creates a new in-memory credential store
func NewInMemoryCredentialStore(config InMemoryCredentialStoreConfig) *InMemoryCredentialStore {
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}

	store := &InMemoryCredentialStore{
		credentials: make(map[string]*GuestCredential),
		logger:      config.Logger,
	}

	store.logger.Info("InMemoryCredentialStore initialized")
	return store
}

// Store saves a credential to the store
func (s *InMemoryCredentialStore) Store(cred *GuestCredential) error {
	if cred.ID == "" {
		return fmt.Errorf("credential ID cannot be empty")
	}
	if cred.Username == "" {
		return fmt.Errorf("credential username cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	cred.CreatedAt = now
	cred.UpdatedAt = now

	credCopy := *cred
	if cred.Metadata != nil {
		credCopy.Metadata = make(map[string]string)
		for k, v := range cred.Metadata {
			credCopy.Metadata[k] = v
		}
	}

	s.credentials[cred.ID] = &credCopy

	s.logger.Info("Credential stored",
		zap.String("credential_id", cred.ID),
		zap.String("name", cred.Name),
		zap.String("type", cred.Type))

	return nil
}

// Get retrieves a credential by ID
func (s *InMemoryCredentialStore) Get(id string) (*GuestCredential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cred, exists := s.credentials[id]
	if !exists {
		return nil, fmt.Errorf("credential %s not found", id)
	}

	credCopy := *cred
	if cred.Metadata != nil {
		credCopy.Metadata = make(map[string]string)
		for k, v := range cred.Metadata {
			credCopy.Metadata[k] = v
		}
	}

	return &credCopy, nil
}

// List returns all credentials
func (s *InMemoryCredentialStore) List() []*GuestCredential {
	s.mu.RLock()
	defer s.mu.RUnlock()

	creds := make([]*GuestCredential, 0, len(s.credentials))
	for _, cred := range s.credentials {
		credCopy := *cred
		if cred.Metadata != nil {
			credCopy.Metadata = make(map[string]string)
			for k, v := range cred.Metadata {
				credCopy.Metadata[k] = v
			}
		}
		creds = append(creds, &credCopy)
	}

	return creds
}

// Delete removes a credential by ID
func (s *InMemoryCredentialStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.credentials[id]; !exists {
		return fmt.Errorf("credential %s not found", id)
	}

	delete(s.credentials, id)
	s.logger.Info("Credential deleted", zap.String("credential_id", id))
	return nil
}

// Update modifies an existing credential
func (s *InMemoryCredentialStore) Update(cred *GuestCredential) error {
	if cred.ID == "" {
		return fmt.Errorf("credential ID cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.credentials[cred.ID]; !exists {
		return fmt.Errorf("credential %s not found", cred.ID)
	}

	now := time.Now()
	cred.UpdatedAt = now

	if existing, ok := s.credentials[cred.ID]; ok {
		cred.CreatedAt = existing.CreatedAt
		cred.LastUsedAt = existing.LastUsedAt
	}

	credCopy := *cred
	if cred.Metadata != nil {
		credCopy.Metadata = make(map[string]string)
		for k, v := range cred.Metadata {
			credCopy.Metadata[k] = v
		}
	}

	s.credentials[cred.ID] = &credCopy
	s.logger.Info("Credential updated", zap.String("credential_id", cred.ID))
	return nil
}

// FindByName searches for credentials by name
func (s *InMemoryCredentialStore) FindByName(name string) []*GuestCredential {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*GuestCredential
	for _, cred := range s.credentials {
		if cred.Name == name {
			credCopy := *cred
			if cred.Metadata != nil {
				credCopy.Metadata = make(map[string]string)
				for k, v := range cred.Metadata {
					credCopy.Metadata[k] = v
				}
			}
			results = append(results, &credCopy)
		}
	}

	return results
}

// Close cleans up resources
func (s *InMemoryCredentialStore) Close() error {
	s.logger.Info("InMemoryCredentialStore closed")
	return nil
}

// EncryptedCredentialStore wraps a credential store with encryption
type EncryptedCredentialStore struct {
	mu        sync.RWMutex
	store     CredentialStore
	cipher    cipher.AEAD
	key       []byte
	logger    *zap.Logger
	encryptor *CredentialEncryptor
}

// EncryptedCredentialStoreConfig holds configuration
type EncryptedCredentialStoreConfig struct {
	EncryptionKey []byte      `json:"-"`
	Store         CredentialStore
	Logger        *zap.Logger `json:"-"`
}

// NewEncryptedCredentialStore creates encrypted store
func NewEncryptedCredentialStore(config EncryptedCredentialStoreConfig) (*EncryptedCredentialStore, error) {
	if config.Store == nil {
		return nil, fmt.Errorf("credential store cannot be nil")
	}
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}

	var encryptor *CredentialEncryptor
	var err error

	if len(config.EncryptionKey) > 0 {
		encryptor, err = NewCredentialEncryptor(config.EncryptionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize encryptor: %w", err)
		}
	}

	ecs := &EncryptedCredentialStore{
		store:     config.Store,
		logger:    config.Logger,
		encryptor: encryptor,
	}

	if encryptor != nil {
		ecs.cipher = encryptor.cipher
	}

	ecs.logger.Info("EncryptedCredentialStore initialized",
		zap.Bool("encryption_enabled", encryptor != nil))

	return ecs, nil
}

// Store saves an encrypted credential
func (ecs *EncryptedCredentialStore) Store(cred *GuestCredential) error {
	if ecs.encryptor == nil {
		return ecs.store.Store(cred)
	}

	ecs.mu.Lock()
	defer ecs.mu.Unlock()

	if cred.Password != "" {
		encrypted, err := ecs.encryptor.Encrypt(cred.Password)
		if err != nil {
			return fmt.Errorf("failed to encrypt password: %w", err)
		}
		cred.Password = encrypted
		cred.Encrypted = true
	}

	return ecs.store.Store(cred)
}

// Get retrieves and decrypts a credential
func (ecs *EncryptedCredentialStore) Get(id string) (*GuestCredential, error) {
	ecs.mu.Lock()
	defer ecs.mu.Unlock()

	cred, err := ecs.store.Get(id)
	if err != nil {
		return nil, err
	}

	if ecs.encryptor != nil && cred.Encrypted && cred.Password != "" {
		decrypted, err := ecs.encryptor.Decrypt(cred.Password)
		if err != nil {
			ecs.logger.Warn("Failed to decrypt password",
				zap.String("credential_id", id),
				zap.Error(err))
			return cred, nil
		}
		cred.Password = decrypted
	}

	return cred, nil
}

// List returns all credentials (passwords remain encrypted)
func (ecs *EncryptedCredentialStore) List() []*GuestCredential {
	return ecs.store.List()
}

// Delete removes a credential
func (ecs *EncryptedCredentialStore) Delete(id string) error {
	return ecs.store.Delete(id)
}

// Update modifies an encrypted credential
func (ecs *EncryptedCredentialStore) Update(cred *GuestCredential) error {
	if ecs.encryptor == nil {
		return ecs.store.Update(cred)
	}

	ecs.mu.Lock()
	defer ecs.mu.Unlock()

	existing, err := ecs.store.Get(cred.ID)
	if err != nil {
		return err
	}

	if cred.Password == "" {
		cred.Password = existing.Password
		cred.Encrypted = existing.Encrypted
	} else if !cred.Encrypted {
		encrypted, err := ecs.encryptor.Encrypt(cred.Password)
		if err != nil {
			return fmt.Errorf("failed to encrypt password: %w", err)
		}
		cred.Password = encrypted
		cred.Encrypted = true
	}

	return ecs.store.Update(cred)
}

// FindByName searches for credentials by name
func (ecs *EncryptedCredentialStore) FindByName(name string) []*GuestCredential {
	return ecs.store.FindByName(name)
}

// Close cleans up resources
func (ecs *EncryptedCredentialStore) Close() error {
	return ecs.store.Close()
}

// CredentialEncryptor handles encryption/decryption
type CredentialEncryptor struct {
	cipher cipher.AEAD
	key    []byte
}

// NewCredentialEncryptor creates a new encryptor
func NewCredentialEncryptor(key []byte) (*CredentialEncryptor, error) {
	if len(key) == 0 {
		return nil, fmt.Errorf("encryption key cannot be empty")
	}

	key, err := normalizeKey(key)
	if err != nil {
		return nil, fmt.Errorf("invalid encryption key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	return &CredentialEncryptor{
		cipher: gcm,
		key:    key,
	}, nil
}

// Encrypt encrypts plaintext
func (e *CredentialEncryptor) Encrypt(plaintext string) (string, error) {
	nonce := make([]byte, e.cipher.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := e.cipher.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts ciphertext
func (e *CredentialEncryptor) Decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	nonceSize := e.cipher.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := e.cipher.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

func normalizeKey(key []byte) ([]byte, error) {
	keyLen := len(key)
	if keyLen >= 32 {
		return key[:32], nil
	}
	if keyLen >= 24 {
		return key[:24], nil
	}
	if keyLen >= 16 {
		return key[:16], nil
	}
	return nil, fmt.Errorf("key too short: %d bytes, minimum 16 required", keyLen)
}

// ValidateCredential validates a credential
func ValidateCredential(cred *GuestCredential) error {
	if cred.ID == "" {
		return fmt.Errorf("credential ID is required")
	}
	if cred.Username == "" {
		return fmt.Errorf("username is required")
	}
	if cred.Type == "" {
		return fmt.Errorf("credential type is required")
	}

	validTypes := map[string]bool{
		"windows": true,
		"linux":   true,
		"vmware":  true,
		"hyperv":  true,
		"ssh":     true,
		"service": true,
	}

	if !validTypes[cred.Type] {
		return fmt.Errorf("invalid credential type: %s", cred.Type)
	}

	return nil
}
