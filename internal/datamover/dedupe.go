package datamover

import (
	"sync"
)

// Deduplicator handles source-side deduplication logic
type Deduplicator struct {
	hasher *ChunkHasher
	mu     sync.RWMutex
	cache  map[string]bool // Local cache of hashes already seen in this session
}

func NewDeduplicator() *Deduplicator {
	return &Deduplicator{
		hasher: NewChunkHasher(),
		cache:  make(map[string]bool),
	}
}

// CheckAndHash hashes the data and returns the hash and whether it's new locally
func (d *Deduplicator) CheckAndHash(data []byte) (string, bool) {
	hash := d.hasher.HashChunk(data)
	
	d.mu.RLock()
	seen := d.cache[hash]
	d.mu.RUnlock()
	
	if seen {
		return hash, false
	}
	
	return hash, true
}

// MarkAsTransfered marks a hash as successfully transfered to storage
func (d *Deduplicator) MarkAsTransfered(hash string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.cache[hash] = true
}

// ClearSession clears the local deduplication cache for a new job session
func (d *Deduplicator) ClearSession() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.cache = make(map[string]bool)
}
