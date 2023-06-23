package fixtures

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/repositories"
	"github.com/G-Research/fasttrackml/pkg/database"
)

// RunFixtures represents data fixtures object.
type RunFixtures struct {
	baseFixtures
	runRepository repositories.RunRepositoryProvider
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
		baseFixtures:  baseFixtures{db: db.DB},
		runRepository: repositories.NewRunRepository(db.DB),
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

// CreateTestRuns creates some num runs belonging to the experiment
func (f RunFixtures) CreateTestRuns(
	ctx context.Context, exp *models.Experiment, num int,
) ([]*models.Run, error) {
	var runs []*models.Run
	// create runs for the experiment
	for i := 0; i < num; i++ {
		run := &models.Run{
			ID:             strings.ReplaceAll(uuid.New().String(), "-", ""),
			ExperimentID:   *exp.ID,
			SourceType:     "JOB",
			LifecycleStage: models.LifecycleStageActive,
			Status:         models.StatusRunning,
		}
		run, err := f.CreateTestRun(context.Background(), run)
		if err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	return runs, nil
}

// GetTestRuns fetches all runs for an experiment
func (f RunFixtures) GetTestRuns(
	ctx context.Context, experimentID int32) ([]models.Run, error) {
	runs := []models.Run{}
	if err := f.db.WithContext(ctx).
		Where("experiment_id = ?", experimentID).
		Order("start_time desc").
		Find(&runs).Error; err != nil {
		return nil, eris.Wrapf(err, "error getting `run` entities by experiment id: %v", experimentID)
	}
	return runs, nil
}

// FindMinMaxRowNums finds min and max rownum for an experiment's runs
func (f RunFixtures) FindMinMaxRowNums(
	ctx context.Context, experimentID int32) (int64, int64, error) {
	runs, err := f.GetTestRuns(ctx, experimentID)
	if err != nil {
		return 0, 0, eris.Wrap(err, "error fetching test runs")
	}
	var min, max models.RowNum
	for _, run := range runs {
		if run.RowNum < min {
			min = run.RowNum
		}
		if run.RowNum > max {
			max = run.RowNum
		}
	}
	return int64(min), int64(max), nil
}
