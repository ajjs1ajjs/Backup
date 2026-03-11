package recovery

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"novabackup/internal/database"
	"novabackup/internal/compression"
	"novabackup/internal/storage"
	"novabackup/pkg/models"
	"github.com/google/uuid"
	"github.com/go-git/go-billy/v5"
)

// ChunkVFS implements a virtual file system that reads from deduplicated chunks
type ChunkVFS struct {
	db      *database.Connection
	comp    compression.Compressor
	storage *storage.Engine
	mu      sync.RWMutex
	files   map[string]*VirtualFile
}

type VirtualFile struct {
	Name    string
	Size    int64
	Chunks  []string // List of chunk hashes
	RelPath string
}

func NewChunkVFS(db *database.Connection, comp compression.Compressor, storage *storage.Engine) *ChunkVFS {
	return &ChunkVFS{
		db:      db,
		comp:    comp,
		storage: storage,
		files:   make(map[string]*VirtualFile),
	}
}

// AddFile adds a virtual file to the VFS
func (v *ChunkVFS) AddFile(name string, size int64, chunks []string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.files[name] = &VirtualFile{
		Name:   name,
		Size:   size,
		Chunks: chunks,
	}
}

// LoadRestorePoint populates the VFS with files from a specific restore point
func (v *ChunkVFS) LoadRestorePoint(ctx context.Context, rpID uuid.UUID) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	chunks, err := v.db.GetChunksForRestorePoint(rpID)
	if err != nil {
		return err
	}

	v.files = make(map[string]*VirtualFile)
	// We need the mapping to know which chunks belong to which file path
	mappings, err := v.db.GetRestorePointMapping(rpID)
	if err != nil {
		return err
	}

	// Create a map of hashes to metadata for quick lookup
	chunkMetadata := make(map[string]models.ChunkInfo)
	for _, c := range chunks {
		chunkMetadata[c.Hash] = c
	}

	for _, m := range mappings {
		vf, ok := v.files[m.OriginalPath]
		if !ok {
			vf = &VirtualFile{
				Name:    m.OriginalPath,
				RelPath: m.OriginalPath,
				Chunks:  []string{},
				Size:    0,
			}
			v.files[m.OriginalPath] = vf
		}
		vf.Chunks = append(vf.Chunks, m.ChunkHash)
		if meta, ok := chunkMetadata[m.ChunkHash]; ok {
			vf.Size += meta.SizeBytes
		}
	}

	return nil
}

// ReadAt reads data from the virtual file system at a specific offset
func (v *ChunkVFS) ReadAt(relPath string, p []byte, off int64) (n int, err error) {
	v.mu.RLock()
	vf, ok := v.files[relPath]
	v.mu.RUnlock()

	if !ok {
		return 0, os.ErrNotExist
	}

	if off >= vf.Size {
		return 0, io.EOF
	}

	bytesRead := 0
	currentTotalOffset := int64(0)

	for _, hash := range vf.Chunks {
		// Need chunk size to calculate overlaps correctly
		// Let's assume chunks were stored with their size.
		// For now, we'll fetch the whole chunk and then slice it.
		
		data, err := v.storage.GetChunk(hash)
		if err != nil {
			return bytesRead, err
		}

		// Decompress
		decompressed, err := v.comp.Decompress(data)
		if err != nil {
			return bytesRead, err
		}

		chunkSize := int64(len(decompressed))
		
		// Does this chunk overlap with requested range?
		if currentTotalOffset + chunkSize > off {
			startInChunk := off - currentTotalOffset
			if startInChunk < 0 {
				startInChunk = 0
			}

			available := chunkSize - startInChunk
			toCopy := int64(len(p)) - int64(bytesRead)
			if toCopy > available {
				toCopy = available
			}

			copy(p[bytesRead:], decompressed[startInChunk:startInChunk+toCopy])
			bytesRead += int(toCopy)
			off += toCopy // Update offset to point to the next needed data

			if bytesRead >= len(p) {
				break
			}
		}
		currentTotalOffset += chunkSize
	}
	return bytesRead, nil
}

// --- billy.Filesystem Implementation ---

func (v *ChunkVFS) Create(filename string) (billy.File, error) {
	return nil, billy.ErrReadOnly
}

func (v *ChunkVFS) Open(filename string) (billy.File, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	// Clean filename (remove leading slash)
	filename = filepath.Clean(filename)
	if filename != "." && filename[0] == filepath.Separator {
		filename = filename[1:]
	}

	vf, ok := v.files[filename]
	if !ok {
		return nil, os.ErrNotExist
	}

	return &VirtualFileHandle{vf: vf, vfs: v, off: 0}, nil
}

func (v *ChunkVFS) OpenFile(filename string, flag int, perm os.FileMode) (billy.File, error) {
	if flag&os.O_CREATE != 0 || flag&os.O_WRONLY != 0 || flag&os.O_RDWR != 0 {
		return nil, billy.ErrReadOnly
	}
	return v.Open(filename)
}

func (v *ChunkVFS) Stat(filename string) (os.FileInfo, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	filename = filepath.Clean(filename)
	if filename == "." {
		return &VirtualFileInfo{name: "/", size: 0, isDir: true}, nil
	}

	if filename[0] == filepath.Separator {
		filename = filename[1:]
	}

	vf, ok := v.files[filename]
	if !ok {
		return nil, os.ErrNotExist
	}

	return &VirtualFileInfo{name: vf.Name, size: vf.Size, isDir: false}, nil
}

func (v *ChunkVFS) ReadDir(path string) ([]os.FileInfo, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	var infos []os.FileInfo
	for _, vf := range v.files {
		infos = append(infos, &VirtualFileInfo{name: vf.Name, size: vf.Size, isDir: false})
	}
	return infos, nil
}

// Dummy methods to satisfy billy.Filesystem
func (v *ChunkVFS) Rename(oldpath, newpath string) error { return billy.ErrReadOnly }
func (v *ChunkVFS) Remove(filename string) error         { return billy.ErrReadOnly }
func (v *ChunkVFS) Join(elem ...string) string           { return filepath.Join(elem...) }
func (v *ChunkVFS) TempFile(dir, prefix string) (billy.File, error) { return nil, billy.ErrReadOnly }
func (v *ChunkVFS) Readlink(link string) (string, error) { return "", billy.ErrNotSupported }
func (v *ChunkVFS) Symlink(target, link string) error    { return billy.ErrReadOnly }
func (v *ChunkVFS) Lstat(filename string) (os.FileInfo, error) { return v.Stat(filename) }
func (v *ChunkVFS) MkdirAll(filename string, perm os.FileMode) error { return billy.ErrReadOnly }
func (v *ChunkVFS) Chroot(path string) (billy.Filesystem, error) { return nil, billy.ErrNotSupported }
func (v *ChunkVFS) Root() string { return "/" }

// --- VirtualFileHandle Implementation ---

type VirtualFileHandle struct {
	vf  *VirtualFile
	vfs *ChunkVFS
	off int64
}

func (h *VirtualFileHandle) Name() string { return h.vf.Name }
func (h *VirtualFileHandle) Write(p []byte) (n int, err error) { return 0, billy.ErrReadOnly }
func (h *VirtualFileHandle) Read(p []byte) (n int, err error) {
	n, err = h.vfs.ReadAt(h.vf.Name, p, h.off)
	h.off += int64(n)
	return n, err
}
func (h *VirtualFileHandle) ReadAt(p []byte, off int64) (n int, err error) {
	return h.vfs.ReadAt(h.vf.Name, p, off)
}
func (h *VirtualFileHandle) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		h.off = offset
	case io.SeekCurrent:
		h.off += offset
	case io.SeekEnd:
		h.off = h.vf.Size + offset
	}
	return h.off, nil
}
func (h *VirtualFileHandle) Close() error { return nil }
func (h *VirtualFileHandle) Lock() error  { return nil }
func (h *VirtualFileHandle) Unlock() error { return nil }
func (h *VirtualFileHandle) Truncate(size int64) error { return billy.ErrReadOnly }

// --- VirtualFileInfo Implementation ---

type VirtualFileInfo struct {
	name  string
	size  int64
	isDir bool
}

func (i *VirtualFileInfo) Name() string       { return i.name }
func (i *VirtualFileInfo) Size() int64        { return i.size }
func (i *VirtualFileInfo) Mode() os.FileMode { 
	if i.isDir {
		return os.ModeDir | 0555
	}
	return 0444 
}
func (i *VirtualFileInfo) ModTime() time.Time { return time.Now() }
func (i *VirtualFileInfo) IsDir() bool        { return i.isDir }
func (i *VirtualFileInfo) Sys() interface{}   { return nil }
