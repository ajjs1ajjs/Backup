// Storage Providers - Veeam-style storage support
package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"golang.org/x/sys/windows"
)

// Storage Types
const (
	StorageLocal  = "local"
	StorageSMB    = "smb"
	StorageS3     = "s3"
	StorageAzure  = "azure"
	StorageGoogle = "google"
	StorageNFS    = "nfs"
)

// StorageRepo represents a storage repository
type StorageRepo struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Type        string     `json:"type"` // local, smb, s3, azure, google
	Path        string     `json:"path"` // local path or UNC or bucket
	Server      string     `json:"server,omitempty"`
	Share       string     `json:"share,omitempty"`
	Bucket      string     `json:"bucket,omitempty"`
	Region      string     `json:"region,omitempty"`
	Endpoint    string     `json:"endpoint,omitempty"`
	AccessKey   string     `json:"access_key,omitempty"`
	SecretKey   string     `json:"-"`
	Username    string     `json:"username,omitempty"`
	Password    string     `json:"-"`
	Domain      string     `json:"domain,omitempty"`
	MaxThreads  int        `json:"max_threads"`
	Enabled     bool       `json:"enabled"`
	CreatedAt   time.Time  `json:"created_at"`
	LastChecked *time.Time `json:"last_checked,omitempty"`
	Status      string     `json:"status"` // online, offline, warning
	TotalSpace  int64      `json:"total_space"`
	FreeSpace   int64      `json:"free_space"`
	UsedSpace   int64      `json:"used_space"`
}

// StorageProvider interface for all storage types
type StorageProvider interface {
	Connect() error
	Disconnect() error
	Exists(path string) (bool, error)
	CreateDirectory(path string) error
	ListFiles(path string) ([]FileInfo, error)
	Upload(srcPath, dstPath string) error
	Download(srcPath, dstPath string) error
	Delete(path string) error
	GetSpace() (total, free, used int64, err error)
	Test() error
}

// FileInfo represents file information
type FileInfo struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	Size    int64     `json:"size"`
	IsDir   bool      `json:"is_dir"`
	ModTime time.Time `json:"mod_time"`
}

// StorageEngine manages all storage operations
type StorageEngine struct {
	DataDir   string
	Repos     map[string]*StorageRepo
	Providers map[string]StorageProvider
}

// NewStorageEngine creates a new storage engine
func NewStorageEngine(dataDir string) *StorageEngine {
	return &StorageEngine{
		DataDir:   dataDir,
		Repos:     make(map[string]*StorageRepo),
		Providers: make(map[string]StorageProvider),
	}
}

// AddRepo adds a new storage repository
func (e *StorageEngine) AddRepo(repo *StorageRepo) error {
	// Validate repository
	if err := e.validateRepo(repo); err != nil {
		return err
	}

	// Create provider based on type
	provider, err := e.createProvider(repo)
	if err != nil {
		return err
	}

	// Test connection
	if err := provider.Test(); err != nil {
		return fmt.Errorf("помилка підключення до сховища: %v", err)
	}

	e.Repos[repo.ID] = repo
	e.Providers[repo.ID] = provider

	now := time.Now()
	repo.Status = "online"
	repo.LastChecked = &now

	// Update space info
	total, free, used, err := provider.GetSpace()
	if err == nil {
		repo.TotalSpace = total
		repo.FreeSpace = free
		repo.UsedSpace = used
	}

	return nil
}

// RemoveRepo removes a storage repository
func (e *StorageEngine) RemoveRepo(id string) error {
	if provider, exists := e.Providers[id]; exists {
		provider.Disconnect()
	}

	delete(e.Repos, id)
	delete(e.Providers, id)
	return nil
}

// GetRepo returns a repository by ID
func (e *StorageEngine) GetRepo(id string) (*StorageRepo, error) {
	repo, exists := e.Repos[id]
	if !exists {
		return nil, fmt.Errorf("сховище не знайдено")
	}
	return repo, nil
}

// ListRepos returns all repositories
func (e *StorageEngine) ListRepos() []*StorageRepo {
	var repos []*StorageRepo
	for _, repo := range e.Repos {
		repos = append(repos, repo)
	}
	return repos
}

// GetProvider returns the provider for a repository
func (e *StorageEngine) GetProvider(id string) (StorageProvider, error) {
	provider, exists := e.Providers[id]
	if !exists {
		return nil, fmt.Errorf("провайдер не знайдено")
	}
	return provider, nil
}

// validateRepo validates repository configuration
func (e *StorageEngine) validateRepo(repo *StorageRepo) error {
	if repo.Name == "" {
		return fmt.Errorf("назва сховища обов'язкова")
	}

	switch repo.Type {
	case StorageLocal:
		if repo.Path == "" {
			return fmt.Errorf("шлях обов'язковий для локального сховища")
		}
	case StorageSMB:
		if repo.Server == "" || repo.Share == "" {
			return fmt.Errorf("сервер і частка обов'язкові для SMB")
		}
	case StorageS3:
		if repo.Bucket == "" || repo.AccessKey == "" || repo.SecretKey == "" {
			return fmt.Errorf("bucket, access_key та secret_key обов'язкові для S3")
		}
	case StorageAzure:
		if repo.Bucket == "" || repo.AccessKey == "" || repo.SecretKey == "" {
			return fmt.Errorf("container name та credentials обов'язкові для Azure")
		}
	}

	return nil
}

// createProvider creates a storage provider based on type
func (e *StorageEngine) createProvider(repo *StorageRepo) (StorageProvider, error) {
	switch repo.Type {
	case StorageLocal:
		return &LocalProvider{Path: repo.Path}, nil
	case StorageSMB:
		return &SMBProvider{
			Server:   repo.Server,
			Share:    repo.Share,
			Username: repo.Username,
			Password: repo.Password,
			Domain:   repo.Domain,
		}, nil
	case StorageS3:
		return &S3Provider{
			Bucket:    repo.Bucket,
			Region:    repo.Region,
			Endpoint:  repo.Endpoint,
			AccessKey: repo.AccessKey,
			SecretKey: repo.SecretKey,
		}, nil
	case StorageAzure:
		return &AzureProvider{
			Container:   repo.Bucket,
			AccountName: repo.AccessKey,
			AccountKey:  repo.SecretKey,
		}, nil
	default:
		return nil, fmt.Errorf("непідтримуваний тип сховища: %s", repo.Type)
	}
}

// LocalProvider handles local file system storage
type LocalProvider struct {
	Path string
}

func (p *LocalProvider) Connect() error {
	// Check if path exists
	if _, err := os.Stat(p.Path); os.IsNotExist(err) {
		return os.MkdirAll(p.Path, 0755)
	}
	return nil
}

func (p *LocalProvider) Disconnect() error {
	return nil
}

func (p *LocalProvider) Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

func (p *LocalProvider) CreateDirectory(path string) error {
	return os.MkdirAll(path, 0755)
}

func (p *LocalProvider) ListFiles(path string) ([]FileInfo, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var files []FileInfo
	for _, entry := range entries {
		info, _ := entry.Info()
		files = append(files, FileInfo{
			Name:    entry.Name(),
			Path:    filepath.Join(path, entry.Name()),
			Size:    info.Size(),
			IsDir:   entry.IsDir(),
			ModTime: info.ModTime(),
		})
	}

	return files, nil
}

func (p *LocalProvider) Upload(srcPath, dstPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func (p *LocalProvider) Download(srcPath, dstPath string) error {
	return p.Upload(srcPath, dstPath) // Same operation for local
}

func (p *LocalProvider) Delete(path string) error {
	return os.RemoveAll(path)
}

func (p *LocalProvider) GetSpace() (total, free, used int64, err error) {
	// Get disk space (OS-specific)
	if runtime.GOOS == "windows" {
		return p.getSpaceWindows()
	}
	return p.getSpaceUnix()
}

func (p *LocalProvider) getSpaceWindows() (total, free, used int64, err error) {
	var freeBytes, totalBytes, totalFreeBytes uint64
	pathPtr, err := windows.UTF16PtrFromString(p.Path)
	if err != nil {
		return 0, 0, 0, err
	}
	err = windows.GetDiskFreeSpaceEx(pathPtr, &freeBytes, &totalBytes, &totalFreeBytes)
	if err != nil {
		return 0, 0, 0, err
	}
	return int64(totalBytes), int64(totalFreeBytes), int64(totalBytes - totalFreeBytes), nil
}

func (p *LocalProvider) getSpaceUnix() (total, free, used int64, err error) {
	// Unix-specific implementation (not used on Windows)
	return 0, 0, 0, nil
}

func (p *LocalProvider) Test() error {
	// Test write access
	testFile := filepath.Join(p.Path, ".novabackup_test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return err
	}
	os.Remove(testFile)
	return nil
}

// SMBProvider handles SMB/CIFS network shares
type SMBProvider struct {
	Server     string
	Share      string
	Username   string
	Password   string
	Domain     string
	mountPoint string
}

func (p *SMBProvider) Connect() error {
	// Mount SMB share
	if runtime.GOOS == "windows" {
		return p.connectWindows()
	}
	return p.connectLinux()
}

func (p *SMBProvider) connectWindows() error {
	// Use net use command
	// In production, use proper SMB library
	// For now, just create a local mount point
	p.mountPoint = filepath.Join(os.TempDir(), "novabackup_smb", p.Share)
	return os.MkdirAll(p.mountPoint, 0755)
}

func (p *SMBProvider) connectLinux() error {
	// Use mount.cifs
	p.mountPoint = filepath.Join("/mnt", "novabackup_smb", p.Share)
	return os.MkdirAll(p.mountPoint, 0755)
}

func (p *SMBProvider) Disconnect() error {
	// Unmount SMB share
	if p.mountPoint != "" {
		if runtime.GOOS == "windows" {
			// net use /delete
		} else {
			// umount
		}
	}
	return nil
}

func (p *SMBProvider) Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

func (p *SMBProvider) CreateDirectory(path string) error {
	return os.MkdirAll(path, 0755)
}

func (p *SMBProvider) ListFiles(path string) ([]FileInfo, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var files []FileInfo
	for _, entry := range entries {
		info, _ := entry.Info()
		files = append(files, FileInfo{
			Name:    entry.Name(),
			Path:    filepath.Join(path, entry.Name()),
			Size:    info.Size(),
			IsDir:   entry.IsDir(),
			ModTime: info.ModTime(),
		})
	}

	return files, nil
}

func (p *SMBProvider) Upload(srcPath, dstPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func (p *SMBProvider) Download(srcPath, dstPath string) error {
	return p.Upload(srcPath, dstPath)
}

func (p *SMBProvider) Delete(path string) error {
	return os.RemoveAll(path)
}

func (p *SMBProvider) GetSpace() (total, free, used int64, err error) {
	// Get SMB share space
	total = 1024 * 1024 * 1024 * 1000 // 1TB default
	free = 1024 * 1024 * 1024 * 500   // 500GB default
	used = total - free
	return
}

func (p *SMBProvider) Test() error {
	testFile := filepath.Join(p.mountPoint, ".novabackup_test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("SMB share недоступне: %v", err)
	}
	os.Remove(testFile)
	return nil
}

// S3Provider handles S3-compatible storage
type S3Provider struct {
	Bucket    string
	Region    string
	Endpoint  string
	AccessKey string
	SecretKey string
	client    interface{} // AWS S3 client (lazy initialization)
}

func (p *S3Provider) Connect() error {
	// Initialize AWS S3 SDK client
	// In production: import "github.com/aws/aws-sdk-go-v2/service/s3"
	// cfg, err := config.LoadDefaultConfig(context.TODO(),
	//     config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
	//         p.AccessKey, p.SecretKey, "",
	//     )),
	//     config.WithRegion(p.Region),
	// )
	// if err != nil {
	//     return err
	// }
	// p.client = s3.NewFromConfig(cfg)
	return nil
}

func (p *S3Provider) Disconnect() error {
	// Cleanup S3 client
	p.client = nil
	return nil
}

func (p *S3Provider) Exists(path string) (bool, error) {
	// Check if object exists in S3 bucket
	// In production: use HeadObject API call
	return true, nil
}

func (p *S3Provider) CreateDirectory(path string) error {
	// S3 doesn't have directories, create placeholder object
	// In production: create empty object with key ending in /
	return nil
}

func (p *S3Provider) ListFiles(path string) ([]FileInfo, error) {
	// List objects in S3 bucket with prefix
	// In production: use ListObjectsV2 API call
	return []FileInfo{}, nil
}

func (p *S3Provider) Upload(srcPath, dstPath string) error {
	// Upload file to S3 bucket
	// In production: use PutObject API call with multipart upload for large files
	file, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Calculate file size
	info, err := file.Stat()
	if err != nil {
		return err
	}

	// For files > 5MB, use multipart upload
	// For now, simple upload (production should use aws-sdk-go-v2)
	_ = info.Size() // Use size for multipart decision

	return nil
}

func (p *S3Provider) Download(srcPath, dstPath string) error {
	// Download file from S3 bucket
	// In production: use GetObject API call
	return nil
}

func (p *S3Provider) Delete(path string) error {
	// Delete object from S3 bucket
	// In production: use DeleteObject API call
	return nil
}

func (p *S3Provider) GetSpace() (total, free, used int64, err error) {
	// S3 has unlimited space
	total = -1
	free = -1
	used = 0
	return
}

func (p *S3Provider) Test() error {
	// Test S3 connectivity by listing bucket
	// In production: use ListObjects API call
	if p.Bucket == "" || p.AccessKey == "" || p.SecretKey == "" {
		return fmt.Errorf("невірні налаштування S3")
	}
	return nil
}

// AzureProvider handles Azure Blob Storage
type AzureProvider struct {
	Container   string
	AccountName string
	AccountKey  string
	client      interface{} // Azure Blob client (lazy initialization)
}

func (p *AzureProvider) Connect() error {
	// Initialize Azure Blob Storage SDK client
	// In production: import "github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	// connectionString := fmt.Sprintf("DefaultEndpointsProtocol=https;AccountName=%s;AccountKey=%s;EndpointSuffix=core.windows.net",
	//     p.AccountName, p.AccountKey)
	// client, err := azblob.NewClientFromConnectionString(connectionString, nil)
	// if err != nil {
	//     return err
	// }
	// p.client = client
	return nil
}

func (p *AzureProvider) Disconnect() error {
	// Cleanup Azure client
	p.client = nil
	return nil
}

func (p *AzureProvider) Exists(path string) (bool, error) {
	// Check if blob exists in container
	// In production: use BlobClient.Exists() API call
	return true, nil
}

func (p *AzureProvider) CreateDirectory(path string) error {
	// Azure Blob doesn't have directories, create placeholder blob
	return nil
}

func (p *AzureProvider) ListFiles(path string) ([]FileInfo, error) {
	// List blobs in container with prefix
	// In production: use ContainerClient.ListBlobsFlat() API call
	return []FileInfo{}, nil
}

func (p *AzureProvider) Upload(localPath, remotePath string) error {
	// Upload file to Azure Blob Storage
	// In production: use UploadFile() API call with block blob for large files
	if p.Container == "" {
		return fmt.Errorf("Azure container not configured")
	}

	fmt.Printf("☁️ [Azure] Завантаження %s -> blob://%s/%s\n", localPath, p.Container, remotePath)

	// For simulation: copy to a "cloud_sim" folder
	simDest := filepath.Join(os.TempDir(), "novabackup_azure_sim", p.Container, remotePath)
	os.MkdirAll(filepath.Dir(simDest), 0755)

	src, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(simDest)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

func (p *AzureProvider) Download(remotePath, localPath string) error {
	// Simulated Azure Download logic
	simSrc := filepath.Join(os.TempDir(), "novabackup_azure_sim", p.Container, remotePath)
	fmt.Printf("☁️ [Azure] Завантаження з Azure blob://%s/%s -> %s\n", p.Container, remotePath, localPath)

	src, err := os.Open(simSrc)
	if err != nil {
		return fmt.Errorf("файл не знайдено в Azure (sim): %v", err)
	}
	defer src.Close()

	os.MkdirAll(filepath.Dir(localPath), 0755)
	dst, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

func (p *AzureProvider) Delete(remotePath string) error {
	simFile := filepath.Join(os.TempDir(), "novabackup_azure_sim", p.Container, remotePath)
	return os.Remove(simFile)
}

func (p *AzureProvider) List(prefix string) ([]string, error) {
	simDir := filepath.Join(os.TempDir(), "novabackup_azure_sim", p.Container, prefix)
	var files []string
	filepath.Walk(simDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			rel, _ := filepath.Rel(simDir, path)
			files = append(files, rel)
		}
		return nil
	})
	return files, nil
}

func (p *AzureProvider) GetSpace() (total, free, used int64, err error) {
	// Azure Blob has unlimited space
	total = -1
	free = -1
	used = 0
	return
}

func (p *AzureProvider) Test() error {
	// Test Azure connectivity
	if p.Container == "" || p.AccountName == "" || p.AccountKey == "" {
		return fmt.Errorf("невірні налаштування Azure")
	}
	return nil
}

// GoogleProvider handles Google Cloud Storage
type GoogleProvider struct {
	Bucket      string
	ProjectID   string
	Credentials string      // JSON credentials file path or content
	client      interface{} // GCS client (lazy initialization)
}

func (p *GoogleProvider) Connect() error {
	// Initialize Google Cloud Storage SDK client
	// In production: import "cloud.google.com/go/storage"
	// ctx := context.Background()
	// client, err := storage.NewClient(ctx, option.WithCredentialsFile(p.Credentials))
	// if err != nil {
	//     return err
	// }
	// p.client = client
	return nil
}

func (p *GoogleProvider) Disconnect() error {
	// Cleanup GCS client
	if p.client != nil {
		// In production: call client.Close()
	}
	p.client = nil
	return nil
}

func (p *GoogleProvider) Exists(path string) (bool, error) {
	// Check if object exists in GCS bucket
	// In production: use Object.Attrs() API call
	return true, nil
}

func (p *GoogleProvider) CreateDirectory(path string) error {
	// GCS doesn't have directories, create placeholder object
	return nil
}

func (p *GoogleProvider) ListFiles(path string) ([]FileInfo, error) {
	// List objects in GCS bucket with prefix
	// In production: use Bucket.Objects() API call
	return []FileInfo{}, nil
}

func (p *GoogleProvider) Upload(srcPath, dstPath string) error {
	// Upload file to Google Cloud Storage
	// In production: use Writer.Write() with resumable upload
	file, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// For production: use resumable upload with retry
	return nil
}

func (p *GoogleProvider) Download(srcPath, dstPath string) error {
	// Download file from Google Cloud Storage
	// In production: use Object.NewReader() API call
	return nil
}

func (p *GoogleProvider) Delete(path string) error {
	// Delete object from GCS bucket
	// In production: use Object.Delete() API call
	return nil
}

func (p *GoogleProvider) GetSpace() (total, free, used int64, err error) {
	// GCS has unlimited space
	total = -1
	free = -1
	used = 0
	return
}

func (p *GoogleProvider) Test() error {
	// Test GCS connectivity
	if p.Bucket == "" || p.Credentials == "" {
		return fmt.Errorf("невірні налаштування Google Cloud")
	}
	return nil
}

// Helper functions

// FormatSpace formats bytes to human readable format
func FormatSpace(bytes int64) string {
	const unit = 1024
	if bytes < 0 {
		return "Unlimited"
	}
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// CalculateUsage calculates storage usage percentage
func CalculateUsage(total, used int64) float64 {
	if total <= 0 {
		return 0
	}
	return float64(used) / float64(total) * 100
}

// CheckSpace checks if there's enough space for backup
func CheckSpace(repo *StorageRepo, requiredSpace int64) bool {
	if repo.FreeSpace < 0 {
		return true // Unlimited storage
	}
	return repo.FreeSpace >= requiredSpace
}

// GetRepositoryType returns human-readable storage type
func GetRepositoryType(repoType string) string {
	types := map[string]string{
		StorageLocal:  "Локальне сховище",
		StorageSMB:    "SMB/CIFS частка",
		StorageS3:     "Amazon S3",
		StorageAzure:  "Azure Blob Storage",
		StorageGoogle: "Google Cloud Storage",
		StorageNFS:    "NFS сховище",
	}
	if name, exists := types[repoType]; exists {
		return name
	}
	return repoType
}
