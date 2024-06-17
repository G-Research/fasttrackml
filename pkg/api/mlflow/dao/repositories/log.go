package repositories

import (
	"context"
	"time"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/common/dao/repositories"
)

// LogRepositoryProvider provides an interface to work with models.Log entity.
type LogRepositoryProvider interface {
	repositories.BaseRepositoryProvider
	// Create creates new models.Log entity connected to models.Run.
	Create(ctx context.Context, log *models.Log) error
	// CleanExpired delete expired Run log outputs.
	CleanExpired(ctx context.Context, period time.Duration) (int64, error)
	// GetFinishedRuns returns finished runs with theirs logs.
	GetFinishedRuns(ctx context.Context) ([]models.Run, error)
}

// LogRepository repository to work with models.Log entity.
type LogRepository struct {
	repositories.BaseRepositoryProvider
	maxRowsPerRun int
}

// NewLogRepository creates a repository to work with models.Log entity.
func NewLogRepository(db *gorm.DB, maxRowsPerRun int) *LogRepository {
	return &LogRepository{
		repositories.NewBaseRepository(db),
		maxRowsPerRun,
	}
}

// Create creates new models.Log entity connected to models.Run.
func (r LogRepository) Create(ctx context.Context, log *models.Log) error {
	if err := r.GetDB().WithContext(ctx).Create(log).Error; err != nil {
		return eris.Wrapf(err, "error creating log row for run %s", log.RunID)
	}
	return r.enforceMaxRowsPerRun(ctx, log.RunID)
}

// enforceMaxRowsPerRun will truncate the log rows for the run if needed.
func (r LogRepository) enforceMaxRowsPerRun(ctx context.Context, runID string) error {
	var rowCount int64
	if err := r.GetDB().WithContext(
		ctx,
	).Model(
		models.Log{},
	).Where(
		"run_uuid = ?", runID,
	).Count(&rowCount).Error; err != nil {
		return eris.Wrapf(err, "error counting log rows for run %s", runID)
	}
	if rowCount <= int64(r.maxRowsPerRun) {
		return nil
	}
	if err := r.GetDB().WithContext(ctx).Exec(`
		DELETE FROM logs
		WHERE id IN (
			 SELECT id
			 FROM logs
			 WHERE run_uuid = ?
			 ORDER BY timestamp ASC
			 LIMIT ?
		)`,
		runID, rowCount-int64(r.maxRowsPerRun),
	).Error; err != nil {
		return eris.Wrapf(err, "error deleting excess log rows for run %s", runID)
	}
	return nil
}

// CleanExpired delete expired Run log outputs.
func (r LogRepository) CleanExpired(ctx context.Context, period time.Duration) (int64, error) {
	result := r.GetDB().WithContext(ctx).Exec(`
		DELETE FROM logs
		WHERE id IN (
			 SELECT id
			 FROM logs
			 LEFT JOIN runs ON runs.run_uuid = logs.run_uuid
			 WHERE (runs.lifecycle_stage = ?) AND timestamp < ?
		)`,
		models.LifecycleStageDeleted,
		time.Now().Add(-period).Unix(),
	)
	if err := result.Error; err != nil {
		return 0, eris.Wrap(err, "error deleting run logs")
	}

	return result.RowsAffected, nil
}

// GetFinishedRuns returns finished runs with theirs logs.
func (r LogRepository) GetFinishedRuns(ctx context.Context) ([]models.Run, error) {
	var runs []models.Run
	if err := r.GetDB().WithContext(
		ctx,
	).Select(
		"DISTINCT runs.*",
	).Model(
		&models.Log{},
	).Joins(
		"JOIN runs ON runs.run_uuid = logs.run_uuid",
	).Preload(
		"Logs",
	).Where(
		"runs.status = ?", models.StatusFinished,
	).Find(&runs).Error; err != nil {
		return nil, eris.Wrap(err, "error getting finished runs")
	}
	return runs, nil
}
