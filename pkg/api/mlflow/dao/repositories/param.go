package repositories

import (
	"context"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ParamRepositoryProvider provides an interface to work with models.Param entity.
type ParamRepositoryProvider interface {
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

// CreateBatch creates []models.Param entities in batch.
func (r ParamRepository) CreateBatch(ctx context.Context, batchSize int, params []models.Param) error {
	tx := r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "run_uuid"}, {Name: "key"}, {Name: "value"}},
		DoNothing: true,
	}).CreateInBatches(params, batchSize)
	if tx.Error != nil {
		return eris.Wrap(tx.Error, "error creating params in batch")
	}
	if tx.RowsAffected != int64(len(params)) {
		return eris.New("error inserting params (duplicate key/different value?)")
	}
	return nil
}
