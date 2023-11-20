package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/database"
)

// RunRepositoryProvider provides an interface to work with models.Run entity.
type RunRepositoryProvider interface {
	BaseRepositoryProvider
	// GetByID returns models.Run entity by its ID.
	GetByID(ctx context.Context, id string) (*models.Run, error)
	// GetByNamespaceIDRunIDAndLifecycleStage returns models.Run entity by Namespace ID, its ID and Lifecycle Stage.
	GetByNamespaceIDRunIDAndLifecycleStage(
		ctx context.Context, namespaceID uint, runID string, lifecycleStage models.LifecycleStage,
	) (*models.Run, error)
	// GetByNamespaceIDAndRunID returns models.Run entity by Namespace ID and its ID.
	GetByNamespaceIDAndRunID(
		ctx context.Context, namespaceID uint, runID string,
	) (*models.Run, error)
	// Create creates new models.Run entity.
	Create(ctx context.Context, run *models.Run) error
	// Update updates existing models.Experiment entity.
	Update(ctx context.Context, run *models.Run) error
	// Archive marks existing models.Run entity as archived.
	Archive(ctx context.Context, run *models.Run) error
	// Delete removes the existing models.Run
	Delete(ctx context.Context, namespaceID uint, run *models.Run) error
	// Restore marks existing models.Run entity as active.
	Restore(ctx context.Context, run *models.Run) error
	// ArchiveBatch marks existing models.Run entities as archived.
	ArchiveBatch(ctx context.Context, namespaceID uint, ids []string) error
	// DeleteBatch removes the existing models.Run from the db.
	DeleteBatch(ctx context.Context, namespaceID uint, ids []string) error
	// RestoreBatch marks existing models.Run entities as active.
	RestoreBatch(ctx context.Context, namespaceID uint, ids []string) error
	// SetRunTagsBatch sets Run tags in batch.
	SetRunTagsBatch(ctx context.Context, run *models.Run, batchSize int, tags []models.Tag) error
	// UpdateWithTransaction updates existing models.Run entity in scope of transaction.
	UpdateWithTransaction(ctx context.Context, tx *gorm.DB, run *models.Run) error
}

// RunRepository repository to work with models.Run entity.
type RunRepository struct {
	BaseRepository
}

// NewRunRepository creates repository to work with models.Run entity.
func NewRunRepository(db *gorm.DB) *RunRepository {
	return &RunRepository{
		BaseRepository{
			db: db,
		},
	}
}

// GetByID returns models.Run entity by its ID.
func (r RunRepository) GetByID(ctx context.Context, id string) (*models.Run, error) {
	run := models.Run{ID: id}
	if err := r.db.WithContext(
		ctx,
	).Preload(
		"LatestMetrics",
	).Preload(
		"Params",
	).Preload(
		"Tags",
	).First(&run).Error; err != nil {
		return nil, eris.Wrapf(err, "error getting 'run' entity by id: %s", id)
	}
	return &run, nil
}

// GetByNamespaceIDRunIDAndLifecycleStage returns models.Run entity by Namespace ID, its ID and Lifecycle Stage..
func (r RunRepository) GetByNamespaceIDRunIDAndLifecycleStage(
	ctx context.Context, namespaceID uint, runID string, lifecycleStage models.LifecycleStage,
) (*models.Run, error) {
	run := models.Run{ID: runID}
	if err := r.db.WithContext(
		ctx,
	).Preload(
		"LatestMetrics",
	).Preload(
		"Params",
	).Preload(
		"Tags",
	).Joins(
		"INNER JOIN experiments ON experiments.experiment_id = runs.experiment_id AND experiments.namespace_id = ?",
		namespaceID,
	).Where(
		`runs.lifecycle_stage = ?`, lifecycleStage,
	).First(&run).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, eris.Wrapf(err, "error getting 'run' entity by id: %s", runID)
	}
	return &run, nil
}

// GetByNamespaceIDAndRunID returns models.Run entity by Namespace ID and its ID.
func (r RunRepository) GetByNamespaceIDAndRunID(
	ctx context.Context, namespaceID uint, runID string,
) (*models.Run, error) {
	run := models.Run{ID: runID}
	if err := r.db.WithContext(
		ctx,
	).Preload(
		"LatestMetrics",
	).Preload(
		"Params",
	).Preload(
		"Tags",
	).Joins(
		"INNER JOIN experiments ON experiments.experiment_id = runs.experiment_id AND experiments.namespace_id = ?",
		namespaceID,
	).First(&run).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, eris.Wrapf(err, "error getting 'run' entity by id: %s", runID)
	}
	return &run, nil
}

// Create creates new models.Run entity.
func (r RunRepository) Create(ctx context.Context, run *models.Run) error {
	// Lock need to calculate row_num
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if tx.Dialector.Name() == "postgres" {
			if err := tx.Exec("LOCK TABLE runs").Error; err != nil {
				return err
			}
		}
		return tx.Create(&run).Error
	}); err != nil {
		return eris.Wrap(err, "error creating new 'run' entity")
	}
	return nil
}

// Update updates existing models.Run entity.
func (r RunRepository) Update(ctx context.Context, run *models.Run) error {
	if err := r.db.WithContext(ctx).Model(&run).Updates(run).Error; err != nil {
		return eris.Wrapf(err, "error updating run with id: %s", run.ID)
	}
	return nil
}

// Archive marks existing models.Run entity as archived.
func (r RunRepository) Archive(ctx context.Context, run *models.Run) error {
	run.DeletedTime = sql.NullInt64{
		Int64: time.Now().UTC().UnixMilli(),
		Valid: true,
	}
	run.LifecycleStage = models.LifecycleStageDeleted
	if err := r.db.WithContext(ctx).Model(&run).Updates(run).Error; err != nil {
		return eris.Wrapf(err, "error updating existing run with id: %s", run.ID)
	}

	return nil
}

// ArchiveBatch marks existing models.Run entities as archived.
func (r RunRepository) ArchiveBatch(ctx context.Context, namespaceID uint, ids []string) error {
	if err := r.db.WithContext(
		ctx,
	).Model(
		models.Run{},
	).Where(
		"run_uuid IN (?)",
		r.db.Model(
			models.Run{},
		).Select(
			"run_uuid",
		).Joins(
			"INNER JOIN experiments ON experiments.experiment_id = runs.experiment_id AND experiments.namespace_id = ?",
			namespaceID,
		).Where(
			"run_uuid IN (?)", ids,
		),
	).Updates(models.Run{
		DeletedTime: sql.NullInt64{
			Int64: time.Now().UTC().UnixMilli(),
			Valid: true,
		},
		LifecycleStage: models.LifecycleStageDeleted,
	}).Error; err != nil {
		return eris.Wrapf(err, "error updating existing runs with ids: %s", ids)
	}

	return nil
}

// Delete removes the existing models.Run from the db.
func (r RunRepository) Delete(ctx context.Context, namespaceID uint, run *models.Run) error {
	return r.DeleteBatch(ctx, namespaceID, []string{run.ID})
}

// DeleteBatch removes existing models.Run from the db.
func (r RunRepository) DeleteBatch(ctx context.Context, namespaceID uint, ids []string) error {
	if err := r.db.Transaction(func(tx *gorm.DB) error {
		runs := make([]models.Run, 0, len(ids))
		if err := tx.Clauses(
			clause.Returning{Columns: []clause.Column{{Name: "row_num"}}},
		).Where(
			"run_uuid IN (?)",
			r.db.Model(
				models.Run{},
			).Select(
				"run_uuid",
			).Joins(
				"INNER JOIN experiments ON experiments.experiment_id = runs.experiment_id AND experiments.namespace_id = ?",
				namespaceID,
			).Where(
				"run_uuid IN (?)", ids,
			),
		).Delete(
			&runs,
		).Error; err != nil {
			return eris.Wrapf(err, "error deleting existing runs with ids: %s", ids)
		}

		// verify deletion
		// NOTE: tx.RowsAffected does not provide correct number of deleted, using the returning slice instead
		if len(runs) != len(ids) {
			return eris.New("count of deleted runs does not match length of ids input (invalid run ID?)")
		}

		// renumber the remainder
		if err := r.renumberRows(tx, getMinRowNum(runs)); err != nil {
			return eris.Wrapf(err, "error renumbering runs.row_num")
		}
		return nil
	}); err != nil {
		return eris.Wrapf(err, "error deleting runs")
	}

	return nil
}

// Restore marks existing models.Run entity as active.
func (r RunRepository) Restore(ctx context.Context, run *models.Run) error {
	// Use UpdateColumns so we can reset DeletedTime to null
	if err := r.db.WithContext(ctx).Model(&run).UpdateColumns(map[string]any{
		"DeletedTime":    sql.NullInt64{},
		"LifecycleStage": database.LifecycleStageActive,
	}).Error; err != nil {
		return eris.Wrapf(err, "error updating existing run with id: %s", run.ID)
	}

	return nil
}

// RestoreBatch marks existing models.Run entities as active.
func (r RunRepository) RestoreBatch(ctx context.Context, namespaceID uint, ids []string) error {
	if err := r.db.WithContext(
		ctx,
	).Where(
		"run_uuid IN (?)",
		r.db.Model(
			models.Run{},
		).Select(
			"run_uuid",
		).Joins(
			"INNER JOIN experiments ON experiments.experiment_id = runs.experiment_id AND experiments.namespace_id = ?",
			namespaceID,
		).Where(
			"run_uuid IN (?)", ids,
		),
	).Updates(models.Run{
		DeletedTime:    sql.NullInt64{},
		LifecycleStage: models.LifecycleStageActive,
	}).Error; err != nil {
		return eris.Wrapf(err, "error updating existing runs with ids: %s", ids)
	}

	return nil
}

// UpdateWithTransaction updates existing models.Run entity in scope of transaction.
func (r RunRepository) UpdateWithTransaction(ctx context.Context, tx *gorm.DB, run *models.Run) error {
	if err := tx.WithContext(ctx).Model(&run).Updates(run).Error; err != nil {
		return eris.Wrapf(err, "error updating existing run with id: %s", run.ID)
	}

	return nil
}

// SetRunTagsBatch sets Run tags in batch.
func (r RunRepository) SetRunTagsBatch(ctx context.Context, run *models.Run, batchSize int, tags []models.Tag) error {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, tag := range tags {
			switch tag.Key {
			case "mlflow.user":
				run.UserID = tag.Value
				if err := r.UpdateWithTransaction(ctx, tx, run); err != nil {
					return eris.Wrap(err, "error updating run 'user_id' field")
				}
			case "mlflow.runName":
				run.Name = tag.Value
				if err := r.UpdateWithTransaction(ctx, tx, run); err != nil {
					return eris.Wrap(err, "error updating run 'name' field")
				}
			}
		}

		if err := tx.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).CreateInBatches(&tags, batchSize).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// getMinRowNum will find the lowest row_num for the slice of runs
// or 0 for an empty slice
func getMinRowNum(runs []models.Run) models.RowNum {
	var minRowNum models.RowNum
	for _, run := range runs {
		if minRowNum == models.RowNum(0) || run.RowNum < minRowNum {
			minRowNum = run.RowNum
		}
	}
	return minRowNum
}

// renumberRows will update the runs.row_num field with the correct ordinal
func (r RunRepository) renumberRows(tx *gorm.DB, startWith models.RowNum) error {
	if startWith < models.RowNum(0) {
		return eris.New("attempting to renumber with less than 0 row number value")
	}

	if tx.Dialector.Name() == "postgres" {
		if err := tx.Exec("LOCK TABLE runs").Error; err != nil {
			return eris.Wrap(err, "unable to lock table")
		}
	}

	if err := tx.Exec(
		`UPDATE runs
	         SET row_num = rows.new_row_num
                 FROM (
                   SELECT run_uuid, ROW_NUMBER() OVER (ORDER BY start_time) + ? - 1 as new_row_num
                   FROM runs
                   WHERE runs.row_num >= ?
                 ) as rows
	         WHERE runs.run_uuid = rows.run_uuid`,
		int64(startWith),
		int64(startWith)).Error; err != nil {
		return eris.Wrap(err, "error updating runs.row_num")
	}
	return nil
}
