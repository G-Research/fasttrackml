package fixtures

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/repositories"
)

// RunFixtures represents data fixtures object.
type RunFixtures struct {
	baseFixtures
	runRepository    repositories.RunRepositoryProvider
	tagRepository    repositories.TagRepositoryProvider
	metricRepository repositories.MetricRepositoryProvider
}

// NewRunFixtures creates new instance of RunFixtures.
func NewRunFixtures(db *gorm.DB) (*RunFixtures, error) {
	return &RunFixtures{
		baseFixtures:     baseFixtures{db: db},
		runRepository:    repositories.NewRunRepository(db),
		tagRepository:    repositories.NewTagRepository(db),
		metricRepository: repositories.NewMetricRepository(db),
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

// ArchiveRun archive existing runs by their ids.
func (f RunFixtures) ArchiveRun(ctx context.Context, ids []string) error {
	return f.runRepository.ArchiveBatch(ctx, ids)
}

// UpdateRun updates existing Run.
func (f RunFixtures) UpdateRun(
	ctx context.Context, run *models.Run,
) error {
	if err := f.runRepository.Update(ctx, run); err != nil {
		return eris.Wrap(err, "error updating test run")
	}
	return nil
}

// ArchiveRuns soft-deletes existing Runs.
func (f RunFixtures) ArchiveRuns(
	ctx context.Context, runIDs []string,
) error {
	if err := f.runRepository.ArchiveBatch(ctx, runIDs); err != nil {
		return eris.Wrap(err, "error archiving runs")
	}
	return nil
}

// CreateExampleRun creates one example run belonging to the experiment, with tags and metrics.
func (f RunFixtures) CreateExampleRun(
	ctx context.Context, exp *models.Experiment,
) (*models.Run, error) {
	runs, err := f.CreateExampleRuns(ctx, exp, 1)
	if err != nil {
		return nil, err
	}
	return runs[0], err
}

// CreateExampleRuns creates some example runs belonging to the experiment, with tags and metrics.
func (f RunFixtures) CreateExampleRuns(
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
			ArtifactURI:    "artifact_uri",
			LifecycleStage: models.LifecycleStageActive,
		}
		run, err := f.CreateRun(ctx, run)
		if err != nil {
			return nil, err
		}
		tag := models.Tag{
			Key:   "my tag key",
			Value: "my tag value",
			RunID: run.ID,
		}
		err = f.CreateTag(ctx, tag)
		if err != nil {
			return nil, err
		}
		run.Tags = []models.Tag{tag}

		err = f.CreateMetrics(ctx, run, 2)
		if err != nil {
			return nil, err
		}

		err = f.CreateParams(ctx, run, 2)
		if err != nil {
			return nil, err
		}

		runs = append(runs, run)
	}
	return runs, nil
}

// GetRun returns run by requested run id.
func (f RunFixtures) GetRun(ctx context.Context, runID string) (*models.Run, error) {
	run, err := f.runRepository.GetByID(ctx, runID)
	if err != nil {
		return nil, eris.Wrap(err, "error getting test run")
	}
	return run, nil
}

// GetRuns fetches all runs for an experiment.
func (f RunFixtures) GetRuns(
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

// FindMinMaxRowNums finds min and max rownum for an experiment's runs.
func (f RunFixtures) FindMinMaxRowNums(
	ctx context.Context, experimentID int32,
) (int64, int64, error) {
	runs, err := f.GetRuns(ctx, experimentID)
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

// CreateTag creates a new Tag for a run.
func (f RunFixtures) CreateTag(
	ctx context.Context, tag models.Tag,
) error {
	if err := f.tagRepository.CreateRunTagWithTransaction(ctx, f.db, tag.RunID, tag.Key, tag.Value); err != nil {
		return eris.Wrap(err, "error creating run tag")
	}
	return nil
}

// CreateMetric creates a new test Metric.
func (f RunFixtures) CreateMetric(ctx context.Context, metric *models.Metric) error {
	if err := f.baseFixtures.db.WithContext(ctx).Create(metric).Error; err != nil {
		return eris.Wrap(err, "error creating test metric")
	}
	return nil
}

// CreateMetrics creates some example metrics for a Run, up to count.
func (f RunFixtures) CreateMetrics(
	ctx context.Context, run *models.Run, count int,
) error {
	for i := 1; i <= count; i++ {
		// create test `metric` and test `latest metric` and connect to run.

		for iter := 1; iter <= count; iter++ {
			if err := f.baseFixtures.db.WithContext(ctx).Create(&models.Metric{
				Key:       fmt.Sprintf("key%d", i),
				Value:     123.1 + float64(iter),
				Timestamp: 1234567890 + int64(iter),
				RunID:     run.ID,
				Step:      int64(iter),
				IsNan:     false,
				Iter:      int64(iter),
			}).Error; err != nil {
				return err
			}
		}
		err := f.baseFixtures.db.WithContext(ctx).Create(&models.LatestMetric{
			Key:       fmt.Sprintf("key%d", i),
			Value:     123.1 + float64(count),
			Timestamp: 1234567890 + int64(count),
			Step:      int64(count),
			IsNan:     false,
			RunID:     run.ID,
			LastIter:  int64(count),
		}).Error
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateParams creates some example params for a Run, up to count.
func (f RunFixtures) CreateParams(
	ctx context.Context, run *models.Run, count int,
) error {
	for i := 1; i <= count; i++ {
		err := f.baseFixtures.db.WithContext(ctx).Create(&models.Param{
			Key:   fmt.Sprintf("key%d", i),
			Value: fmt.Sprintf("val%d", i),
			RunID: run.ID,
		}).Error
		if err != nil {
			return err
		}
	}
	return nil
}
