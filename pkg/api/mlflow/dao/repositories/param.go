package repositories

import (
	"context"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

// ParamRepositoryProvider provides an interface to work with models.Param entity.
type ParamRepositoryProvider interface {
	// Create creates models.Param entity.
	Create(ctx context.Context, param *models.Param) error
	// CreateBatch creates []models.Param entities in batch.
	CreateBatch(ctx context.Context, batchSize int, params []models.Param) error
}

// ParamRepository repository to work with models.Param entity.
type ParamRepository struct {
	BaseRepository
}

// NewParamRepository creates repository to work with models.Param entity.
func NewParamRepository(db *gorm.DB) *ParamRepository {
	return &ParamRepository{
		BaseRepository{
			db: db,
		},
	}
}

// Create creates models.Param entity.
func (r ParamRepository) Create(ctx context.Context, param *models.Param) error {
	if err := r.db.Create(param).Error; err != nil {
		return eris.Wrap(err, "error creating param")
	}
	return nil
}

// CreateBatch creates []models.Param entities in batch.
func (r ParamRepository) CreateBatch(ctx context.Context, batchSize int, params []models.Param) error {
	if err := r.db.CreateInBatches(params, batchSize).Error; err != nil {
		return eris.Wrap(err, "error creating params in batch")
	}
	return nil
}
