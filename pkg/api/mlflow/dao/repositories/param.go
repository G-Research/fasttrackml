package repositories

import (
	"context"
	"fmt"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
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
	var err error
	err = r.db.CreateInBatches(params, batchSize).Error
	if err != nil {
		// remove duplicate rows and try again
		params, err1 := r.removeExactMatches(ctx, params)
		if err1 != nil {
			return eris.Wrap(err1, "error removing exact matches after: " + err.Error())
		}
		err = r.db.CreateInBatches(params, batchSize).Error
	}

	if err != nil {
		return eris.Wrap(err, "error creating params in batch")
	}
	return nil
}

// removeExactMatches will return a new slice of params which excludes exact matches
func (r ParamRepository) removeExactMatches(ctx context.Context, params []models.Param) ([]models.Param, error) {
	var keys, foundKeys []string
	paramMap := map[string]models.Param{}
	paramsToReturn := []models.Param{}
	for _, param := range params {
		key := fmt.Sprintf("%v-%v-%v", param.RunID, param.Key, param.Value)
		keys = append(keys, key)
		paramMap[key] = param
	}

	tx := r.db.Raw(`
           select run_uuid || '-' || key || '-' || value
           from params
           where run_uuid || '-' || key || '-' || value in ?`, keys).
		Find(&foundKeys)
	if tx.Error != nil {
		return paramsToReturn, eris.Wrap(tx.Error, "problem selecting existing params")
	}

	for _, foundKey := range foundKeys {
		delete(paramMap, foundKey)
	}

	for _, v := range paramMap {
		paramsToReturn = append(paramsToReturn, v)
	}

	return paramsToReturn, nil
}
