package repositories

import (
	"context"
	"errors"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/database"
)

// ExperimentRepositoryProvider provides an interface to work with `experiment` entity.
type ExperimentRepositoryProvider interface {
	// Create creates new models.Experiment entity.
	Create(ctx context.Context, experiment *models.Experiment) error
	// GetByID returns experiment by its ID.
	GetByID(ctx context.Context, experimentID int32) (*models.Experiment, error)
	// GetByName returns experiment by its name.
	GetByName(ctx context.Context, name string) (*models.Experiment, error)
	// Update updates existing models.Experiment entity.
	Update(ctx context.Context, experiment *models.Experiment) error
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
	if err := r.db.Create(&experiment).Error; err != nil {
		return eris.Wrap(err, "error creating experiment entity")
	}
	if experiment.ArtifactLocation == "" {
		if err := database.DB.Model(
			&experiment,
		).Update(
			"ArtifactLocation", experiment.ArtifactLocation,
		).Error; err != nil {
			return eris.Wrapf(err, `error updating artifact_location: '%s'`, experiment.ArtifactLocation)
		}
	}
	return nil
}

// GetByID returns experiment by its ID.
func (r ExperimentRepository) GetByID(ctx context.Context, experimentID int32) (*models.Experiment, error) {
	var experiment models.Experiment
	if err := r.db.WithContext(ctx).Where(
		models.Experiment{ID: &experimentID},
	).Preload("Tags").First(&experiment).Error; err != nil {
		return nil, eris.Wrapf(err, "error getting experiment by id: %d", experimentID)
	}
	return &experiment, nil
}

// GetByName returns experiment by its name.
func (r ExperimentRepository) GetByName(ctx context.Context, name string) (*models.Experiment, error) {
	var experiment models.Experiment
	if err := r.db.WithContext(ctx).Preload(
		"Tags",
	).Where(
		models.Experiment{Name: name},
	).First(&experiment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, eris.Wrapf(err, "error getting experiment by id: %s", name)
	}
	return &experiment, nil
}

// Update updates existing models.Experiment entity.
func (r ExperimentRepository) Update(ctx context.Context, experiment *models.Experiment) error {

	if err := r.db.Transaction(func(tx *gorm.DB) error {

		if err := r.db.WithContext(ctx).Model(&experiment).Updates(experiment).Error; err != nil {
			return eris.Wrapf(err, "error updating experiment with id: %d", *experiment.ID)
		}

		// also archive experiment runs if experiment is being archived
		if experiment.LifecycleStage == models.LifecycleStageDeleted {
			run := models.Run{
				LifecycleStage: experiment.LifecycleStage,
				DeletedTime:    experiment.LastUpdateTime,
			}

			if err := r.db.WithContext(ctx).Model(&run).Where("experiment_id = ?", experiment.ID).Updates(&run).Error; err != nil {
				return eris.Wrapf(err, "error updating existing runs with experiment id: %d", *experiment.ID)
			}
		}
		return nil

	}); err != nil {
		return err
	}

	return nil
}
