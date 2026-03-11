package backup

import (
	"context"
	"log"
	"novabackup/internal/database"
	"github.com/google/uuid"
)

// RetentionService manages cleanup of old restore points
type RetentionService struct {
	db *database.Connection
}

func NewRetentionService(db *database.Connection) *RetentionService {
	return &RetentionService{db: db}
}

// ApplyPolicy deletes restore points older than retention period
func (rs *RetentionService) ApplyPolicy(ctx context.Context, jobID uuid.UUID, retentionDays int) error {
	log.Printf("[Retention] Applying policy for job %s: %d days", jobID, retentionDays)

	points, err := rs.db.GetExpiredRestorePoints(jobID, retentionDays)
	if err != nil {
		return err
	}

	for _, rp := range points {
		log.Printf("[Retention] Expiring restore point: %s (Time: %s)", rp.ID, rp.PointTime)
		
		// 1. Get chunks associated with this point
		chunks, err := rs.db.GetRestorePointChunks(rp.ID)
		if err != nil {
			log.Printf("[Retention] Error getting chunks for RP %s: %v", rp.ID, err)
			continue
		}

		// 2. Decrement chunk ref counts
		for _, hash := range chunks {
			if err := rs.db.DecrementChunkRef(hash); err != nil {
				log.Printf("[Retention] Error decrementing ref for chunk %s: %v", hash, err)
			}
		}

		// 3. Delete restore point from DB
		if err := rs.db.DeleteRestorePoint(rp.ID); err != nil {
			log.Printf("[Retention] Error deleting RP %s: %v", rp.ID, err)
		}
	}

	return nil
}
