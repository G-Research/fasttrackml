package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/aim2/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/common"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/models"
	"github.com/G-Research/fasttrackml/pkg/database"
)

// ExperimentRepositoryProvider provides an interface to work with `experiment` entity.
type ExperimentRepositoryProvider interface {
	// GetExperiments returns list of experiments.
	GetExperiments(ctx context.Context, namespaceID uint) ([]models.ExperimentExtended, error)
	// GetExperimentRuns returns list of runs which belong to experiment.
	GetExperimentRuns(
		ctx context.Context, namespaceID uint, req *request.GetExperimentRunsRequest,
	) ([]models.Run, error)
	// GetExperimentActivity returns experiment activity.
	GetExperimentActivity(
		ctx context.Context, namespaceID uint, experimentID int32, tzOffset int,
	) (*models.ExperimentActivity, error)
	// GetExperimentByNamespaceIDAndExperimentID returns experiment by Namespace ID and Experiment ID.
	GetExperimentByNamespaceIDAndExperimentID(
		ctx context.Context, namespaceID uint, experimentID int32,
	) (*models.ExperimentExtended, error)
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

// GetExperiments returns list of experiments.
func (r ExperimentRepository) GetExperiments(
	ctx context.Context, namespaceID uint,
) ([]models.ExperimentExtended, error) {
	var experiments []models.ExperimentExtended
	if err := r.db.WithContext(ctx).Model(
		&models.ExperimentExtended{},
	).Select(
		"experiments.experiment_id",
		"experiments.name",
		"experiments.lifecycle_stage",
		"experiments.creation_time",
		"COUNT(runs.run_uuid) AS run_count",
		"COALESCE(MAX(experiment_tags.value), '') AS description",
	).Where(
		"experiments.namespace_id = ?", namespaceID,
	).Where(
		"experiments.lifecycle_stage = ?", database.LifecycleStageActive,
	).Joins(
		"LEFT JOIN runs USING(experiment_id)",
	).Joins(
		"LEFT JOIN experiment_tags ON experiments.experiment_id = experiment_tags.experiment_id AND"+
			" experiment_tags.key = ?", common.DescriptionTagKey,
	).Group(
		"experiments.experiment_id",
	).Find(
		&experiments,
	).Error; err != nil {
		return nil, eris.Wrapf(err, "error getting experiments by namespace id: %d", namespaceID)
	}

	return experiments, nil
}

// GetExperimentRuns returns list of runs which belong to experiment.
func (r ExperimentRepository) GetExperimentRuns(
	ctx context.Context, namespaceID uint, req *request.GetExperimentRunsRequest,
) ([]models.Run, error) {
	query := r.db
	if req.Limit > 0 {
		query.Limit(req.Limit)
	}

	if req.Offset != "" {
		run := &models.Run{ID: req.Offset}
		if err := r.db.Select(
			"row_num",
		).First(
			&run,
		).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, eris.Wrapf(err, "error getting runs offset: %q", req.Offset)
		}
		query.Where("row_num < ?", run.RowNum)
	}

	var runs []models.Run
	if err := query.Where(
		"experiment_id = ?", req.ID,
	).Order(
		"row_num DESC",
	).Find(&runs).Error; err != nil {
		return nil, eris.Wrapf(err, "error getting runs of experiment: %s", req.ID)
	}
	return runs, nil
}

// GetExperimentActivity returns experiment activity.
func (r ExperimentRepository) GetExperimentActivity(
	ctx context.Context, namespaceID uint, experimentID int32, tzOffset int,
) (*models.ExperimentActivity, error) {
	var runs []models.Run
	if err := r.db.WithContext(ctx).Select(
		"runs.start_time", "runs.lifecycle_stage", "runs.status",
	).Joins(
		"LEFT JOIN experiments USING(experiment_id)",
	).Where(
		"experiments.namespace_id = ?", namespaceID,
	).Where(
		"experiments.experiment_id = ?", experimentID,
	).Find(&runs).Error; err != nil {
		return nil, eris.Wrapf(err, "error getting runs of experiment: %s", experimentID)
	}

	activity := models.ExperimentActivity{
		NumRuns:     len(runs),
		ActivityMap: map[string]int{},
	}
	for _, run := range runs {
		key := time.UnixMilli(
			run.StartTime.Int64,
		).Add(
			time.Duration(-tzOffset) * time.Minute,
		).Format("2006-01-02T15:00:00")
		activity.ActivityMap[key] += 1
		switch {
		case run.LifecycleStage == models.LifecycleStageDeleted:
			activity.NumArchivedRuns += 1
		case run.Status == models.StatusRunning:
			activity.NumActiveRuns += 1
		}
	}
	return &activity, nil
}

// GetExperimentByNamespaceIDAndExperimentID returns experiment by Namespace ID and Experiment ID.
func (r ExperimentRepository) GetExperimentByNamespaceIDAndExperimentID(
	ctx context.Context, namespaceID uint, experimentID int32,
) (*models.ExperimentExtended, error) {
	var experiment models.ExperimentExtended
	if err := r.db.WithContext(ctx).Model(
		&models.ExperimentExtended{},
	).Select(
		"experiments.experiment_id",
		"experiments.name",
		"experiments.lifecycle_stage",
		"experiments.creation_time",
		"COUNT(runs.run_uuid) AS run_count",
		"COALESCE(MAX(experiment_tags.value), '') AS description",
	).Joins(
		"LEFT JOIN runs USING(experiment_id)",
	).Joins(
		"LEFT JOIN experiment_tags ON experiments.experiment_id = experiment_tags.experiment_id AND"+
			" experiment_tags.key = ?", common.DescriptionTagKey,
	).Where(
		"experiments.namespace_id = ?", namespaceID,
	).Where(
		"experiments.experiment_id = ?", experimentID,
	).Group(
		"experiments.experiment_id",
	).First(&experiment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, eris.Wrapf(err, "error getting experiment by id: %s", experimentID)
	}
	return &experiment, nil
}
