package repositories

import (
	"context"
	"database/sql"
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
	// GetByID returns models.Run entity bt its ID.
	GetByID(ctx context.Context, id string) (*models.Run, error)
	// Create creates new models.Run entity.
	Create(ctx context.Context, run *models.Run) error
	// Archive marks existing models.Run entity as archived.
	Archive(ctx context.Context, run *models.Run) error
	// Delete removes the existing models.Run
	Delete(ctx context.Context, run *models.Run) error
	// Restore marks existing models.Run entity as active.
	Restore(ctx context.Context, run *models.Run) error
	// SetRunTagsBatch sets Run tags in batch.
	SetRunTagsBatch(ctx context.Context, batchSize int, run *models.Run, tags []models.Tag) error
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

// GetByID returns models.Run entity bt its ID.
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
		return nil, eris.Wrapf(err, "error getting `run` entity by id: %s", id)
	}
	return &run, nil
}

// Create creates new models.Run entity.
func (r RunRepository) Create(ctx context.Context, run *models.Run) error {
	//TODO:DSuhinin - purpose of lock here?
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if tx.Dialector.Name() == "postgres" {
			if err := tx.Exec("LOCK TABLE runs").Error; err != nil {
				return err
			}
		}
		return tx.Create(&run).Error
	}); err != nil {
		return eris.Wrap(err, "error creating new `run` entity")
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

// Delete removes the existing models.Run
func (r RunRepository) Delete(ctx context.Context, run *models.Run) error {
	if err := r.db.WithContext(ctx).Model(&run).Delete(run).Error; err != nil {
		return eris.Wrapf(err, "error deleting run with id: %s", run.ID)
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

// UpdateWithTransaction updates existing models.Run entity in scope of transaction.
func (r RunRepository) UpdateWithTransaction(ctx context.Context, tx *gorm.DB, run *models.Run) error {
	if err := tx.WithContext(ctx).Model(&run).Updates(run).Error; err != nil {
		return eris.Wrapf(err, "error updating existing run with id: %s", run.ID)
	}

	return nil
}

// SetRunTagsBatch sets Run tags in batch.
func (r RunRepository) SetRunTagsBatch(ctx context.Context, batchSize int, run *models.Run, tags []models.Tag) error {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, tag := range tags {
			switch tag.Key {
			case "mlflow.user":
				run.UserID = tag.Value
				if err := r.UpdateWithTransaction(ctx, tx, run); err != nil {
					return eris.Wrap(err, "error updating run `user_id` field")
				}
			case "mlflow.runName":
				run.Name = tag.Value
				if err := r.UpdateWithTransaction(ctx, tx, run); err != nil {
					return eris.Wrap(err, "error updating run `name` field")
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
