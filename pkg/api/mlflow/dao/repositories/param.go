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
		Columns:   []clause.Column{{Name: "run_uuid"}, {Name: "key"}},
		DoNothing: true,
	}).CreateInBatches(params, batchSize)
	if tx.Error != nil {
		return eris.Wrap(tx.Error, "error creating params in batch")
	}

	// if there are conflicting params, ignore if the values are the same
	if tx.RowsAffected != int64(len(params)) {
		conflictingParams, err := r.conflictingParams(ctx, params)
		if err != nil {
			return eris.Wrap(err, "error checking for conflicting params")
		}
		if len(conflictingParams) > 0 {
			return eris.Errorf("conflicting params found: %v", conflictingParams)
		}
	}
	return nil
}

// conflictingParams checks if there are conflicting values for the params.
func (r ParamRepository) conflictingParams(ctx context.Context, params []models.Param) ([]string, error) {
	paramsInDB := []models.Param{}
	paramKeysInError := []string{}
	if err := r.db.WithContext(ctx).
		Model(&models.Param{}).
		Where("run_uuid = ?", params[0].RunID).
		Where("key IN ?", r.collectKeys(params)).
		Find(&paramsInDB).Error; err != nil {
		return nil, eris.New("error fetching params from db")
	}
	paramsInDbKeyValueMap := r.collectKeyValues(paramsInDB)

	for _, param := range params {
		if value, ok := paramsInDbKeyValueMap[param.Key]; ok && value != param.Value {
			paramKeysInError = append(paramKeysInError, param.Key)
		}
	}
	return paramKeysInError, nil
}

// collectKeys collects the keys from the params.
func (r ParamRepository) collectKeys(params []models.Param) []string {
	keys := make([]string, len(params))
	for i, param := range params {
		keys[i] = param.Key
	}
	return keys
}

// collectKeys collects the keys from the params.
func (r ParamRepository) collectKeyValues(params []models.Param) map[string]string {
	keyValueMap := make(map[string]string)
	for _, param := range params {
		keyValueMap[param.Key] = param.Value
	}
	return keyValueMap
}
