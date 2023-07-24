package fixtures

import (
	"context"
	"fmt"
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
	tagRepository repositories.TagRepositoryProvider
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
		tagRepository: repositories.NewTagRepository(db.DB),
	}, nil
}

// CreateRun creates a new test Run.
func (f RunFixtures) CreateRun(
	ctx context.Context, run *models.Run,
) (*models.Run, error) {
	if err := f.runRepository.Create(ctx, run); err != nil {
		return nil, eris.Wrap(err, "error creating test run")
	}
	return run, nil
}

// UpdateRun creates a new test Run.
func (f RunFixtures) UpdateRun(
	ctx context.Context, run *models.Run,
) (*models.Run, error) {
	if err := f.runRepository.Update(ctx, run); err != nil {
		return nil, eris.Wrap(err, "error updating test run")
	}
	return run, nil
}

// CreateTag creates a new Tag for a run
func (f RunFixtures) CreateTag(
	ctx context.Context, tag models.Tag,
) error {
	if err := f.tagRepository.CreateRunTagWithTransaction(ctx, f.db, tag.RunID, tag.Key, tag.Value); err != nil {
		return eris.Wrap(err, "error creating run tag")
	}
	return nil
}

// CreateRuns creates some num runs belonging to the experiment
func (f RunFixtures) CreateRuns(
	ctx context.Context, exp *models.Experiment, num int,
) ([]*models.Run, error) {
	var runs []*models.Run
	// create runs for the experiment
	for i := 0; i < num; i++ {
		run := &models.Run{
			ID:             strings.ReplaceAll(uuid.New().String(), "-", ""),
			Name:           fmt.Sprintf("TestRun_%d", i),
			Status:         models.StatusRunning,
			SourceType:     "JOB",
			ExperimentID:   *exp.ID,
			LifecycleStage: models.LifecycleStageActive,
		}
		run, err := f.CreateRun(ctx, run)
		if err != nil {
			return nil, err
		}
		tag := models.Tag{
			Key: "my tag key",
			Value: "my tag value",
			RunID: run.ID,
		}
	        err = f.CreateTag(ctx, tag)
		if err != nil {
			return nil, err
		}
		run.Tags = []models.Tag{ tag }
			
		runs = append(runs, run)
	}
	return runs, nil
}

// GetTestRuns fetches all runs for an experiment
func (f RunFixtures) GetTestRuns(
	ctx context.Context, experimentID int32,
) ([]models.Run, error) {
	var runs []models.Run
	if err := f.db.WithContext(ctx).Where(
		"experiment_id = ?", experimentID,
	).Order(
		"start_time desc",
	).Find(
		&runs,
	).Error; err != nil {
		return nil, eris.Wrapf(err, "error getting `run` entities by experiment id: %v", experimentID)
	}
	return runs, nil
}

// FindMinMaxRowNums finds min and max rownum for an experiment's runs
func (f RunFixtures) FindMinMaxRowNums(
	ctx context.Context, experimentID int32,
) (int64, int64, error) {
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
