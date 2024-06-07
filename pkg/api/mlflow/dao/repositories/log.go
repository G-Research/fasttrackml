package repositories

import (
	"context"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/common/dao/repositories"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

// LogRepositoryProvider provides an interface to work with models.Log entity.
type LogRepositoryProvider interface {
	repositories.BaseRepositoryProvider
	// SaveLog creates new models.Log entity connected to models.Run.
	SaveLog(ctx context.Context, log *models.Log) error
}

// LogRepository repository to work with models.Log entity.
type LogRepository struct {
	repositories.BaseRepositoryProvider
	maxRowsPerRun int
}

// NewLogRepository creates repository to work with models.Log entity.
func NewLogRepository(db *gorm.DB, maxRowsPerRun int) *LogRepository {
	return &LogRepository{
		repositories.NewBaseRepository(db),
		maxRowsPerRun,
	}
}

// SaveLog will store a row of log.
func (r LogRepository) SaveLog(ctx context.Context, log *models.Log) error {
	if err := r.GetDB().WithContext(ctx).Create(log).Error; err != nil {
		return eris.Wrapf(err, "error creating log row for run %s", log.RunID)
	}
	return r.enforceMaxRowsPerRun(ctx, log.RunID)
}

// enforceMaxRowsPerRun will truncate the log rows for the run if needed.
func (r LogRepository) enforceMaxRowsPerRun(ctx context.Context, runID string) error {
	var rowCount int64
	if err := r.GetDB().WithContext(ctx).Model(models.Log{}).Where("run_uuid = ?", runID).Count(&rowCount).Error; err != nil {
		return eris.Wrapf(err, "error counting log rows for run %s", runID)
	}
	if rowCount < int64(r.maxRowsPerRun) {
		return nil
	}
	if err := r.GetDB().WithContext(ctx).Exec(`
                DELETE FROM logs
                WHERE run_uuid = ?
                AND timestamp IN (
                     SELECT timestamp
                     FROM logs
                     WHERE run_uuid = ?
                     ORDER BY timestamp ASC
                     LIMIT ?
                )`, runID, runID, rowCount-int64(r.maxRowsPerRun)).Error; err != nil {
		return eris.Wrapf(err, "error counting log rows for run %s", runID)
	}
	return nil
}
