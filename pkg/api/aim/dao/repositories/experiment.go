package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/G-Research/fasttrackml/pkg/api/aim/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim/common"
	"github.com/G-Research/fasttrackml/pkg/api/aim/dao/models"
	"github.com/G-Research/fasttrackml/pkg/database"
)

// ExperimentRepositoryProvider provides an interface to work with `experiment` entity.
type ExperimentRepositoryProvider interface {
	// Update updates existing experiment.
	Update(ctx context.Context, experiment *models.Experiment) error
	// Delete deletes existing experiment.
	Delete(ctx context.Context, experiment *models.Experiment) error
	// GetExperiments returns list of experiments.
	GetExperiments(ctx context.Context, namespaceID uint) ([]models.ExperimentExtended, error)
	// GetExperimentRuns returns list of runs which belong to experiment.
	GetExperimentRuns(ctx context.Context, req *request.GetExperimentRunsRequest) ([]models.Run, error)
	// GetExperimentActivity returns experiment activity.
	GetExperimentActivity(
		ctx context.Context, namespaceID uint, experimentID int32, tzOffset int,
	) (*models.ExperimentActivity, error)
	// GetExperimentByNamespaceIDAndExperimentID returns experiment by Namespace ID and Experiment ID.
	GetExperimentByNamespaceIDAndExperimentID(
		ctx context.Context, namespaceID uint, experimentID int32,
	) (*models.Experiment, error)
	// GetCountOfActiveExperiments returns count of active experiments.
	GetCountOfActiveExperiments(ctx context.Context, namespaceID uint) (int64, error)
	// GetExtendedExperimentByNamespaceIDAndExperimentID returns extended experiment by Namespace ID and Experiment ID.
	GetExtendedExperimentByNamespaceIDAndExperimentID(
		ctx context.Context, namespaceID uint, experimentID int32,
	) (*models.ExperimentExtended, error)
}

// ExperimentRepository repository to work with `experiment` entity.
type ExperimentRepository struct {
	db *gorm.DB
}

// NewExperimentRepository creates a repository to work with `experiment` entity.
func NewExperimentRepository(db *gorm.DB) *ExperimentRepository {
	return &ExperimentRepository{
		db: db,
	}
}

// Update updates existing experiment.
func (r ExperimentRepository) Update(ctx context.Context, experiment *models.Experiment) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&experiment).Updates(experiment).Error; err != nil {
			return eris.Wrapf(err, "error updating experiment with id: %d", *experiment.ID)
		}

		// also archive experiment runs if experiment is being archived
		if experiment.LifecycleStage == models.LifecycleStageDeleted {
			if err := tx.WithContext(
				ctx,
			).Model(
				&models.Run{},
			).Where(
				"experiment_id = ?", experiment.ID,
			).Updates(&models.Run{
				LifecycleStage: experiment.LifecycleStage,
				DeletedTime:    experiment.LastUpdateTime,
			}).Error; err != nil {
				return eris.Wrapf(err, "error updating existing runs with experiment id: %d", *experiment.ID)
			}
		}
		return nil
	})
}

// Delete deletes existing experiment.
func (r ExperimentRepository) Delete(ctx context.Context, experiment *models.Experiment) error {
	if err := r.db.Transaction(func(tx *gorm.DB) error {
		// finding all the related runs
		var minRowNum sql.NullInt64
		if err := tx.Model(
			&models.Run{},
		).Where(
			"experiment_id  = ?", *experiment.ID,
		).Pluck(
			"MIN(row_num)", &minRowNum,
		).Error; err != nil {
			return err
		}

		// delete current experiment
		if err := tx.Clauses(
			clause.Returning{Columns: []clause.Column{{Name: "experiment_id"}}},
		).Where(
			experiment,
		).Delete(
			&models.Experiment{},
		).Error; err != nil {
			return eris.Wrapf(err, "error deleting existing experiment with id: %d", *experiment.ID)
		}

		// renumbering the remainder runs
		if minRowNum.Valid {
			if models.RowNum(minRowNum.Int64) < models.RowNum(0) {
				return eris.New("attempting to renumber with less than 0 row number value")
			}

			if tx.Dialector.Name() == database.PostgresDialectorName {
				if err := tx.Exec("LOCK TABLE runs").Error; err != nil {
					return eris.Wrap(err, "unable to lock table")
				}
			}

			if err := tx.Exec(
				`UPDATE runs
				 SET row_num = rows.new_row_num
					 FROM (
					   SELECT run_uuid, ROW_NUMBER() OVER (ORDER BY start_time) + ? - 1 as new_row_num
					   FROM runs
					   WHERE runs.row_num >= ?
					 ) as rows
				 WHERE runs.run_uuid = rows.run_uuid`,
				minRowNum.Int64,
				minRowNum.Int64,
			).Error; err != nil {
				return eris.Wrap(err, "error updating runs.row_num")
			}
		}

		return nil
	}); err != nil {
		return eris.Wrapf(err, "error deleting experiment with id: %d", *experiment.ID)
	}

	return nil
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
	).Order(
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
	ctx context.Context, req *request.GetExperimentRunsRequest,
) ([]models.Run, error) {
	query := r.db.WithContext(ctx)
	if req.Limit > 0 {
		query = query.Limit(req.Limit)
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
		query = query.Where("row_num < ?", run.RowNum)
	}

	var runs []models.Run
	if err := query.Where(
		"experiment_id = ?", req.ID,
	).Order(
		"row_num DESC",
	).Find(&runs).Error; err != nil {
		return nil, eris.Wrapf(err, "error getting runs of experiment: %d", req.ID)
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
		return nil, eris.Wrapf(err, "error getting runs of experiment: %d", experimentID)
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
) (*models.Experiment, error) {
	var experiment models.Experiment
	if err := r.db.WithContext(ctx).Preload(
		"Tags",
	).Where(
		models.Experiment{ID: &experimentID, NamespaceID: namespaceID},
	).First(&experiment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, eris.Wrapf(err, "error getting experiment by id: %d", experimentID)
	}
	return &experiment, nil
}

// GetCountOfActiveExperiments returns count of active experiments.
func (r ExperimentRepository) GetCountOfActiveExperiments(ctx context.Context, namespaceID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(
		&database.Experiment{},
	).Where(
		"lifecycle_stage = ?", database.LifecycleStageActive,
	).Where(
		"namespace_id = ?", namespaceID,
	).Count(&count).Error; err != nil {
		return 0, eris.Wrap(err, "error counting experiments")
	}
	return count, nil
}

// GetExtendedExperimentByNamespaceIDAndExperimentID returns experiment by Namespace ID and Experiment ID.
// TODO:dsuhinin this moment needs to be discussed.
func (r ExperimentRepository) GetExtendedExperimentByNamespaceIDAndExperimentID(
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
		return nil, eris.Wrapf(err, "error getting experiment by id: %d", experimentID)
	}
	return &experiment, nil
}
