package deduplication

import (
	"context"
	"log"
	"os"
	"novabackup/internal/database"
)

// GarbageCollector cleans up unused chunks from disk
type GarbageCollector struct {
	db *database.Connection
}

func NewGarbageCollector(db *database.Connection) *GarbageCollector {
	return &GarbageCollector{db: db}
}

// Run scans database for orphaned chunks and deletes them from disk
func (gc *GarbageCollector) Run(ctx context.Context) error {
	log.Printf("[GC] Starting garbage collection...")

	hashes, err := gc.db.GetOrphanedChunks()
	if err != nil {
		return err
	}

	for _, hash := range hashes {
		// Fetch path for safety/log (or just delete from storage path pattern)
		// For now we assume we know where they are or we can fetch them
		
		// Let's assume GetOrphanedChunks returns full path or we can fetch it
		// For this implementation, let's fetch storage path from DB
		// (Actually GetOrphanedChunks only returns hashes in current implementation)
		
		// Better: Update GetOrphanedChunks to return path too or fetch by hash
		path, err := gc.db.GetChunkPath(hash)
		if err != nil {
			log.Printf("[GC] Error finding path for chunk %s: %v", hash, err)
			continue
		}

		if path != "" {
			log.Printf("[GC] Deleting unused chunk: %s", path)
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				log.Printf("[GC] Error deleting file %s: %v", path, err)
			}
		}

		// Final removal from DB
		if err := gc.db.DeleteChunk(hash); err != nil {
			log.Printf("[GC] Error deleting chunk %s from DB: %v", hash, err)
		}
	}

	log.Printf("[GC] Finished. Removed %d chunks.", len(hashes))
	return nil
}
