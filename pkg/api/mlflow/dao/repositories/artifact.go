package repositories

import (
	"context"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

// ArtifactRepositoryProvider provides an interface to work with `artifact` entity.
type ArtifactRepositoryProvider interface {
	// Create creates a new database.App object.
	Create(ctx context.Context, artifact *models.Artifact) error
}

// ArtifactRepository repository to work with `artifact` entity.
type ArtifactRepository struct {
	db *gorm.DB
}

// NewArtifactRepository creates a repository to work with `artifact` entity.
func NewArtifactRepository(db *gorm.DB) *ArtifactRepository {
	return &ArtifactRepository{
		db: db,
	}
}

// Create creates a new database.Artifact object.
func (r ArtifactRepository) Create(ctx context.Context, artifact *models.Artifact) error {
	if err := r.db.WithContext(ctx).Create(&artifact).Error; err != nil {
		return eris.Wrap(err, "error creating artifact entity")
	}
	return nil
}
