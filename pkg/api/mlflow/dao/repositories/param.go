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

// paramConflict represents a conflicting parameter.
type paramConflict struct {
	RunID    string `gorm:"column:run_uuid"`
	Key      string
	OldValue string
	NewValue string
}

// String renders the paramConflict for error messages.
func (pc paramConflict) String() string {
	return fmt.Sprintf("{run_id: %s, key: %s, old_value: %s, new_value: %s}", pc.RunID, pc.Key, pc.OldValue, pc.NewValue)
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
		// if there were ignored conflicts, verify to be exact duplicates
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
// If the key already exists for the run but with a different value, it is a conflict. Conflicts are returned.
func findConflictingParams(tx *gorm.DB, params []models.Param) ([]paramConflict, error) {
	var conflicts []paramConflict
	if err := tx.Raw(
		`WITH new(key, value, run_uuid) AS (VALUES `+makeSqlPlaceholders(params)+`)
		     SELECT current.run_uuid, current.key, current.value as old_value, new.value as new_value
		     FROM params AS current
		     INNER JOIN new USING (run_uuid, key)
		     WHERE new.value != current.value`,
		makeSqlValues(params)...).
		Find(&conflicts).Error; err != nil {
		return nil, eris.Wrap(err, "error fetching params from db")
	}
	return conflicts, nil
}

// makeSqlPlaceholders collects a string of (?,?,?), (?,?,?), etc, to the length of the params.
func makeSqlPlaceholders(params []models.Param) string {
	valuesArray := make([]string, len(params))
	for i := range params {
		valuesArray[i] = "(?,?,?)"
	}
	return strings.Join(valuesArray, ",")
}

// makeSqlValues concatenates Key, Value, RunID from each input Param for use in sql values replacement
func makeSqlValues(params []models.Param) []interface{} {
	// values array is params * 3 in length since using 3 fields from each
	valuesArray := make([]interface{}, len(params)*3)
	index := 0
	for _, param := range params {
		valuesArray[index] = param.Key
		valuesArray[index+1] = param.Value
		valuesArray[index+2] = param.RunID
		index = index + 3
	}
	return valuesArray
}
