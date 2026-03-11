package factory

import (
	"fmt"
	"novabackup/internal/database"
	"novabackup/internal/storage"
	"novabackup/internal/storage/s3"
	"novabackup/pkg/models"
)

// InitializeRepos loads repositories from database and registers providers in the engine
func InitializeRepos(db *database.Connection, engine *storage.Engine) error {
	repos, err := db.GetAllRepositories()
	if err != nil {
		return err
	}

	for _, repo := range repos {
		var p storage.Provider
		var err error

		switch repo.Type {
		case models.RepositoryTypeLocal:
			p, err = storage.NewLocalProvider(repo.Path)
		case models.RepositoryTypeS3:
			p, err = s3.NewS3Engine(&s3.S3Config{
				Endpoint:        repo.Endpoint,
				Bucket:          repo.Bucket,
				Region:          repo.Region,
				AccessKeyID:     repo.AccessKey,
				SecretAccessKey: repo.SecretKey,
				ForcePathStyle:  true,
			})
		}

		if err != nil {
			fmt.Printf("Failed to initialize repository %s: %v\n", repo.Name, err)
			continue
		}

		if p != nil {
			engine.RegisterProvider(repo.Name, p)
			if repo.IsSOBR {
				if repo.Tier == "performance" {
					engine.AddPerformanceExtent(p)
				} else {
					engine.AddCapacityExtent(p)
				}
			}
		}
	}
	return nil
}
