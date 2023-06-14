package fixtures

import (
	"context"
	"time"

	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/models"
	"github.com/G-Research/fasttrackml/pkg/repositories"
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
		baseFixtures:         baseFixtures{db: db},
		experimentRepository: repositories.NewExperimentRepository(db),
	}, nil
}

// CreateTestExperiment creates a new test Experiment.
func (f ExperimentFixtures) CreateTestExperiment(
	ctx context.Context, experiment *models.Experiment,
) (*models.Experiment, error) {
	if err := f.experimentRepository.Create(ctx, experiment); err != nil {
		return nil, eris.Wrap(err, "error creating test experiment")
	}
	return experiment, nil
}
