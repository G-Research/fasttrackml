package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

// ExperimentRepositoryProvider provides an interface to work with `experiment` entity.
type ExperimentRepositoryProvider interface {
	// Create creates new models.Experiment entity.
	Create(ctx context.Context, experiment *models.Experiment) error
	// Update updates existing models.Experiment entity.
	Update(ctx context.Context, experiment *models.Experiment) error
	// Delete removes the existing models.Experiment from the db.
	Delete(ctx context.Context, experiment *models.Experiment) error
	// DeleteBatch removes existing []models.Experiment in batch from the db.
	DeleteBatch(ctx context.Context, ids []*int32) error
	// GetByNamespaceIDAndName returns experiment by Namespace ID and Experiment name.
	GetByNamespaceIDAndName(ctx context.Context, namespaceID uint, name string) (*models.Experiment, error)
	// GetByNamespaceIDAndExperimentID returns experiment by Namespace ID and Experiment ID.
	GetByNamespaceIDAndExperimentID(
		ctx context.Context, namespaceID uint, experimentID int32,
	) (*models.Experiment, error)
	// UpdateWithTransaction updates existing models.Experiment entity in scope of transaction.
	UpdateWithTransaction(ctx context.Context, tx *gorm.DB, experiment *models.Experiment) error
}

// ExperimentRepository repository to work with `experiment` entity.
type ExperimentRepository struct {
	db *gorm.DB
}

// NewExperimentRepository creates repository to work with `experiment` entity.
func NewExperimentRepository(db *gorm.DB) *ExperimentRepository {
	return &ExperimentRepository{
		db: db,
	}
}

// Create creates new models.Experiment entity.
func (r ExperimentRepository) Create(ctx context.Context, experiment *models.Experiment) error {
	if err := r.db.WithContext(ctx).Create(&experiment).Error; err != nil {
		return eris.Wrap(err, "error creating experiment entity")
	}
	if experiment.ArtifactLocation == "" {
		if err := r.db.Model(
			&experiment,
		).Update(
			"ArtifactLocation", experiment.ArtifactLocation,
		).Error; err != nil {
			return eris.Wrapf(err, `error updating artifact_location: '%s'`, experiment.ArtifactLocation)
		}
	}
	return nil
}

// GetByNamespaceIDAndExperimentID returns experiment by Namespace ID and Experiment ID.
func (r ExperimentRepository) GetByNamespaceIDAndExperimentID(
	ctx context.Context, namespaceID uint, experimentID int32,
) (*models.Experiment, error) {
	var experiment models.Experiment
	if err := r.db.WithContext(ctx).Preload(
		"Tags",
	).Where(
		models.Experiment{ID: &experimentID},
	).Where(
		"experiments.namespace_id = ?", namespaceID,
	).First(&experiment).Error; err != nil {
		return nil, eris.Wrapf(err, "error getting experiment by id: %d", experimentID)
	}
	return &experiment, nil
}

// GetByNamespaceIDAndName returns experiment by Namespace ID and Experiment name.
func (r ExperimentRepository) GetByNamespaceIDAndName(
	ctx context.Context, namespaceID uint, name string,
) (*models.Experiment, error) {
	var experiment models.Experiment
	if err := r.db.WithContext(ctx).Preload(
		"Tags",
	).Where(
		models.Experiment{Name: name},
	).Where(
		"experiments.namespace_id = ?", namespaceID,
	).First(&experiment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, eris.Wrapf(err, "error getting experiment by name: %s", name)
	}
	return &experiment, nil
}

// Update updates existing models.Experiment entity.
func (r ExperimentRepository) Update(ctx context.Context, experiment *models.Experiment) error {
	if err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&experiment).Updates(experiment).Error; err != nil {
			return eris.Wrapf(err, "error updating experiment with id: %d", *experiment.ID)
		}

		// also archive experiment runs if experiment is being archived
		if experiment.LifecycleStage == models.LifecycleStageDeleted {
			run := models.Run{
				LifecycleStage: experiment.LifecycleStage,
				DeletedTime:    experiment.LastUpdateTime,
			}

			if err := tx.WithContext(
				ctx,
			).Model(
				&run,
			).Where(
				"experiment_id = ?", experiment.ID,
			).Updates(&run).Error; err != nil {
				return eris.Wrapf(err, "error updating existing runs with experiment id: %d", *experiment.ID)
			}
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

// Delete removes the existing models.Experiment from the db.
func (r ExperimentRepository) Delete(ctx context.Context, experiment *models.Experiment) error {
	return r.DeleteBatch(ctx, []*int32{experiment.ID})
}

// DeleteBatch removes existing []models.Experiment in batch from the db.
func (r ExperimentRepository) DeleteBatch(ctx context.Context, ids []*int32) error {
	if err := r.db.Transaction(func(tx *gorm.DB) error {
		// finding all the runs
		var minRowNum sql.NullInt64
		if err := tx.Model(
			&models.Run{},
		).Where(
			"experiment_id IN (?)", ids,
		).Pluck("MIN(row_num)", &minRowNum).Error; err != nil {
			return err
		}

		experiments := make([]models.Experiment, 0, len(ids))
		if err := tx.Clauses(
			clause.Returning{Columns: []clause.Column{{Name: "experiment_id"}}},
		).Where(
			"experiment_id IN ?", ids,
		).Delete(&experiments).Error; err != nil {
			return eris.Wrapf(err, "error deleting existing experiments with ids: %d", ids)
		}

		// verify deletion
		if len(experiments) != len(ids) {
			return eris.New("count of deleted experiments does not match length of ids input (invalid experiment ID?)")
		}

		// renumbering the remainder runs
		if minRowNum.Valid {
			runRepo := NewRunRepository(tx)
			if err := runRepo.renumberRows(tx, models.RowNum(minRowNum.Int64)); err != nil {
				return eris.Wrapf(err, "error renumbering runs.row_num")
			}
		}

		return nil
	}); err != nil {
		return eris.Wrapf(err, "error deleting experiments")
	}

	return nil
}

// UpdateWithTransaction updates existing models.Experiment entity in scope of transaction.
func (r ExperimentRepository) UpdateWithTransaction(
	ctx context.Context,
	tx *gorm.DB,
	experiment *models.Experiment,
) error {
	if err := tx.WithContext(ctx).Model(&experiment).Updates(experiment).Error; err != nil {
		return eris.Wrapf(err, "error updating existing experiment with id: %d", experiment.ID)
	}

	return nil
}
