package repositories

import (
	"context"
	"fmt"
	"strings"

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

// ParamConflict represents a conflicting parameter.
type ParamConflict struct {
	Key              string
	PreviousValue    string
	ConflictingValue string
}

func (pc ParamConflict) String() string {
	return fmt.Sprintf("key: %s, previous value: %s, conflicting value: %s", pc.Key, pc.PreviousValue, pc.ConflictingValue)
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
	if err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "run_uuid"}, {Name: "key"}},
			DoNothing: true,
		}).CreateInBatches(params, batchSize).Error; err != nil {
			return eris.Wrap(tx.Error, "error creating params in batch")
		}
		// if there are conflicting params, ignore if the values are the same
		if tx.RowsAffected != int64(len(params)) {
			conflictingParams, err := findConflictingParams(tx, params)
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
	}); err != nil {
		return err
	}
	return nil
}

// findConflictingParams checks if there are conflicting values for the input params. If a key does not
// yet exist in the db, or if the same key and value already exist for the run, it is not a conflict.
// If the key already exists for the run but with a different value, it is a conflict. Conflicting keys
// are returned.
func findConflictingParams(
	tx *gorm.DB,
	params []models.Param,
) ([]ParamConflict, error) {
	var conflicts []ParamConflict
	if err := tx.Raw(fmt.Sprintf(`
		    WITH new(key, value, run_uuid) AS (VALUES %s)
		    SELECT current.key as key, current.value as old_value, new.value as new_value
		    FROM params AS current
		    INNER JOIN new ON new.run_uuid = current.run_uuid AND new.key = current.key
		    WHERE new.value != current.value;`,
		makeSqlValues(params))).
		Find(&conflicts).Error; err != nil {
		return nil, eris.Wrap(err, "error fetching params from db")
	}
	return conflicts, nil
}

// makeSqlValues collects a string of (key, value, run_uuid), (key, value, run_uuid), etc, from the params.
func makeSqlValues(params []models.Param) string {
	valuesArray := make([]string, len(params))
	for i, param := range params {
		valuesArray[i] = fmt.Sprintf("('%s', '%s', '%s')", param.Key, param.Value, param.RunID)
	}
	return strings.Join(valuesArray, ",")
}
