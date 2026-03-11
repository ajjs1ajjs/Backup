package datamover

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
)

// ChunkHasher calculates hashes for source-side deduplication
type ChunkHasher struct {
}

func NewChunkHasher() *ChunkHasher {
	return &ChunkHasher{}
}

// HashChunk returns the SHA-256 hash of the provided data
func (h *ChunkHasher) HashChunk(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// HashStream reads from a stream and returns its hash
func (h *ChunkHasher) HashStream(r io.Reader) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, r); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
