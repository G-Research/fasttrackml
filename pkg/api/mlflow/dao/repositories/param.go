package repositories

import (
	"context"
	"fmt"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ParamConflictError is returned when there is a conflict in the params (same key, different value).
type ParamConflictError struct {
	Message string
}

// Error returns the ParamConflictError message.
func (e ParamConflictError) Error() string {
	return e.Message
}

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
		conflictingParams, err := r.findConflictingParams(ctx, params)
		if err != nil {
			return eris.Wrap(err, "error checking for conflicting params")
		}
		if len(conflictingParams) > 0 {
			return ParamConflictError{
				Message: fmt.Sprintf("conflicting params found: %v", conflictingParams),
			}
		}
	}
	return nil
}

// findConflictingParams checks if there are conflicting values for the input params. If a key does not
// yet exist in the db, or if the same key and value already exist for the run, it is not a conflict.
// If the key already exists for the run but with a different value, it is a conflict. Conflicting keys are returned.
func (r ParamRepository) findConflictingParams(ctx context.Context, params []models.Param) ([]string, error) {
	dbParams := []models.Param{}
	paramKeysInError := []string{}
	if err := r.db.WithContext(ctx).
		Model(&models.Param{}).
		Where("run_uuid = ?", params[0].RunID).
		Where("key IN ?", r.collectKeys(params)).
		Find(&dbParams).Error; err != nil {
		return nil, eris.New("error fetching params from db")
	}
	dbParamsAsMap := r.collectKeyValues(dbParams)
	for _, param := range params {
		if value, ok := dbParamsAsMap[param.Key]; ok && value != param.Value {
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

// collectKeyValues collects the keys and values as a map from the params.
func (r ParamRepository) collectKeyValues(params []models.Param) map[string]string {
	keyValueMap := make(map[string]string)
	for _, param := range params {
		keyValueMap[param.Key] = param.Value
	}
	return keyValueMap
}
