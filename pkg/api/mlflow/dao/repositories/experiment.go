package repositories

import (
	"context"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

// ExperimentRepositoryProvider provides an interface to work with `experiment` entity.
type ExperimentRepositoryProvider interface {
	// GetByID returns experiment by its ID.
	GetByID(ctx context.Context, experimentID int32) (*models.Experiment, error)
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

// GetByID returns experiment by its ID.
func (r ExperimentRepository) GetByID(ctx context.Context, experimentID int32) (*models.Experiment, error) {
	experiment := models.Experiment{ID: &experimentID}
	if err := r.db.WithContext(ctx).First(&experiment).Error; err != nil {
		return nil, eris.Wrapf(err, "error getting experiment by id: %d", experimentID)
	}
	return &experiment, nil
}
