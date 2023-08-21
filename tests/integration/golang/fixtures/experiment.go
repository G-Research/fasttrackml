package fixtures

import (
	"context"
	"database/sql"
	"time"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/repositories"
	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
)

// ExperimentFixtures represents data fixtures object.
type ExperimentFixtures struct {
	baseFixtures
	experimentRepository repositories.ExperimentRepositoryProvider
}

// NewExperimentFixtures creates new instance of ExperimentFixtures.
func NewExperimentFixtures(db *gorm.DB) (*ExperimentFixtures, error) {
	return &ExperimentFixtures{
		baseFixtures:         baseFixtures{db: db},
		experimentRepository: repositories.NewExperimentRepository(db),
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

// CreateExperiments creates some num new test experiments.
func (f ExperimentFixtures) CreateExperiments(
	ctx context.Context, num int,
) ([]*models.Experiment, error) {
	var experiments []*models.Experiment
	for i := 0; i < num; i++ {
		experiment := &models.Experiment{
			Name: "Test Experiment",
			Tags: []models.ExperimentTag{
				{
					Key:   "key1",
					Value: "value1",
				},
			},
			CreationTime: sql.NullInt64{
				Int64: time.Now().UTC().UnixMilli(),
				Valid: true,
			},
			LastUpdateTime: sql.NullInt64{
				Int64: time.Now().UTC().UnixMilli(),
				Valid: true,
			},
			LifecycleStage:   models.LifecycleStageActive,
			ArtifactLocation: "/artifact/location",
		}
		experiment, err := f.CreateExperiment(ctx, experiment)
		if err != nil {
			return nil, err
		}
		experiments = append(experiments, experiment)
	}
	return experiments, nil
}

// GetTestExperiments fetches all experiments.
func (f ExperimentFixtures) GetTestExperiments(
	ctx context.Context,
) ([]models.Experiment, error) {
	var experiments []models.Experiment
	if err := f.db.WithContext(ctx).
		Find(&experiments).Error; err != nil {
		return nil, eris.Wrapf(err, "error getting 'experiment' entities")
	}
	return experiments, nil
}

// GetByNamespaceIDAndExperimentID returns the experiment by Namespace ID and the given Experiment id.
func (f ExperimentFixtures) GetByNamespaceIDAndExperimentID(
	ctx context.Context, namespaceID uint, experimentID int32,
) (*models.Experiment, error) {
	experiment, err := f.experimentRepository.GetByNamespaceIDAndExperimentID(ctx, namespaceID, experimentID)
	if err != nil {
		return nil, eris.Wrapf(err, "error getting experiment with ID %d", experimentID)
	}
	return experiment, nil
}
