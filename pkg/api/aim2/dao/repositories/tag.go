package repositories

import (
	"context"

	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/models"
)

// TagRepositoryProvider provides an interface to work with models.Tag entity.
type TagRepositoryProvider interface {
	BaseRepositoryProvider
	// GetTagsByNamespace returns the list of tags.
	GetTagsByNamespace(ctx context.Context, namespaceID uint) ([]models.Tag, error)
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

// GetTagsByNamespace returns the list of tags.
// TODO not really implemented
func (r TagRepository) GetTagsByNamespace(ctx context.Context, namespaceID uint) ([]models.Tag, error) {
	var tags []models.Tag
	if err := r.db.WithContext(ctx).Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}
