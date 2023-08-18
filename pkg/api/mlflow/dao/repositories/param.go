package repositories

import (
	"context"
	"fmt"

	"github.com/G-Research/fasttrackml/pkg/common/dao/models"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"
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
	// try to create params in batch; error condition requires special handling
	// to allow certain duplicates
	if err := r.db.CreateInBatches(params, batchSize).Error; err != nil {
		// remove duplicate rows and try again
		dedupedParams, errRemovingMatches := r.removeExactMatches(ctx, params)
		if errRemovingMatches != nil {
			return eris.Wrap(errRemovingMatches, "error removing duplicates in batch")
		}
		if err := r.db.CreateInBatches(dedupedParams, batchSize).Error; err != nil {
			return eris.Wrap(err, "error creating params in batch after removing duplicates")
		}
	}
	return nil
}

// removeExactMatches will return a new slice of params which excludes exact matches
func (r ParamRepository) removeExactMatches(ctx context.Context, params []models.Param) ([]models.Param, error) {
	var keys []string
	paramMap := map[string]models.Param{}
	for _, param := range params {
		key := fmt.Sprintf("%v-%v-%v", param.RunID, param.Key, param.Value)
		keys = append(keys, key)
		paramMap[key] = param
	}

	var foundKeys []string
	if err := r.db.Raw(`
		select run_uuid || '-' || key || '-' || value
		from params
		where run_uuid || '-' || key || '-' || value in ?`, keys).
		Find(&foundKeys).Error; err != nil {
		return []models.Param{}, eris.Wrap(err, "problem selecting existing params")
	}

	for _, foundKey := range foundKeys {
		delete(paramMap, foundKey)
	}

	paramsToReturn := []models.Param{}
	for _, v := range paramMap {
		paramsToReturn = append(paramsToReturn, v)
	}

	return paramsToReturn, nil
}
