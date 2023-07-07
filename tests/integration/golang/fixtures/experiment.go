package fixtures

import (
	"context"
	"time"

	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/repositories"
	"github.com/G-Research/fasttrackml/pkg/database"
)

// ExperimentFixtures represents data fixtures object.
type ExperimentFixtures struct {
	baseFixtures
	experimentRepository repositories.ExperimentRepositoryProvider
}

// NewExperimentFixtures creates new instance of ExperimentFixtures.
func NewExperimentFixtures(databaseDSN string) (*ExperimentFixtures, error) {
	db, err := database.ConnectDB(
		databaseDSN,
		1*time.Second,
		20,
		false,
		false,
		"",
	)
	if err != nil {
		return nil, eris.Wrap(err, "error connection to database")
	}
	return &ExperimentFixtures{
		baseFixtures:         baseFixtures{db: db.DB},
		experimentRepository: repositories.NewExperimentRepository(db.DB),
	}, nil
}

// CreateExperiment creates a new test Experiment.
func (f ExperimentFixtures) CreateExperiment(
	ctx context.Context, experiment *models.Experiment,
) (*models.Experiment, error) {
	if err := f.experimentRepository.Create(ctx, experiment); err != nil {
		return nil, eris.Wrap(err, "error creating test experiment")
	}
	return experiment, nil
}

// GetTestExperiments fetches all experiments
func (f ExperimentFixtures) GetTestExperiments(
	ctx context.Context) ([]models.Experiment, error) {
	experiments := []models.Experiment{}
	if err := f.db.WithContext(ctx).
		Find(&experiments).Error; err != nil {
		return nil, eris.Wrapf(err, "error getting `experiment` entities")
	}
	return experiments, nil
}

// GetExperimentByID returns the experiment by the given id
func (f ExperimentFixtures) GetExperimentByID(ctx context.Context, experimentID int32) (*models.Experiment, error) {
	experiment, err := f.experimentRepository.GetByID(ctx, experimentID)
	if err != nil {
		return nil, eris.Wrapf(err, "error getting experiment with ID %d", experimentID)
	}
	return experiment, nil
}
