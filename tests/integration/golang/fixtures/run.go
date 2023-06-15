package fixtures

import (
	"context"
	"time"

	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/repositories"
	"github.com/G-Research/fasttrackml/pkg/database"
)

// RunFixtures represents data fixtures object.
type RunFixtures struct {
	baseFixtures
	runRepository        repositories.RunRepositoryProvider
}

// NewRunFixtures creates new instance of RunFixtures.
func NewRunFixtures(databaseDSN string) (*RunFixtures, error) {
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
	return &RunFixtures{
		baseFixtures:  baseFixtures{db: db},
		runRepository: repositories.NewRunRepository(db),
	}, nil
}

// CreateTestRun creates a new test Run.
func (f RunFixtures) CreateTestRun(
	ctx context.Context, run *models.Run,
) (*models.Run, error) {
	if err := f.runRepository.Create(ctx, run); err != nil {
		return nil, eris.Wrap(err, "error creating test run")
	}
	return run, nil
}
