package repositories

import (
	"context"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/database"
)

// TagRepositoryProvider provides an interface to work with models.Tag entity.
type TagRepositoryProvider interface {
	BaseRepositoryProvider
	// Delete deletes existing models.Tag entity.
	Delete(ctx context.Context, tag *models.Tag) error
	// GetByRunIDAndKey returns models.Tag by provided RunID and Tag Key.
	GetByRunIDAndKey(ctx context.Context, runID, key string) (*models.Tag, error)
	// CreateWithTransaction creates new models.Tag entity connected to models.Run.
	CreateWithTransaction(ctx context.Context, tx *gorm.DB, runID, key, value string) error
}

// TagRepository repository to work with models.Tag entity.
type TagRepository struct {
	BaseRepository
}

// NewTagRepository creates repository to work with models.Tag entity.
func NewTagRepository(db *gorm.DB) *TagRepository {
	return &TagRepository{
		BaseRepository{
			db: db,
		},
	}
}

// Delete deletes existing models.Tag entity.
func (r TagRepository) Delete(ctx context.Context, tag *models.Tag) error {
	if err := database.DB.Delete(tag).Error; err != nil {
		return eris.Wrapf(err, "error deleting tag by run id: %s and key: %s", tag.RunID, tag.Key)
	}
	return nil
}

// GetByRunIDAndKey returns models.Tag by provided RunID and Tag Key.
func (r TagRepository) GetByRunIDAndKey(ctx context.Context, runID, key string) (*models.Tag, error) {
	tag := models.Tag{RunID: runID, Key: key}
	if err := r.db.WithContext(ctx).First(&tag).Error; err != nil {
		return nil, eris.Wrapf(err, "error getting tag by run id: %s and tag key: %s", runID, key)
	}
	return &tag, nil
}

// CreateWithTransaction creates new models.Tag entity connected to models.Run.
func (r TagRepository) CreateWithTransaction(
	ctx context.Context, tx *gorm.DB, runID, key, value string,
) error {
	if err := tx.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create([]database.Tag{{
		Key:   key,
		Value: value,
		RunID: runID,
	}}).Error; err != nil {
		return eris.Wrapf(err, "error creating tag for run with id: %s", runID)
	}
	return nil
}
