package repositories

import (
	"context"
	"database/sql"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/aim/dao/models"
	"github.com/G-Research/fasttrackml/pkg/common/dao/repositories"
)

// LogRepositoryProvider provides an interface to work with models.Log entity.
type LogRepositoryProvider interface {
	// GetLogsByNamespaceIDAndRunID returns logs by Run ID.
	GetLogsByNamespaceIDAndRunID(
		ctx context.Context, namespaceID uint, runID string,
	) (*sql.Rows, func(rows *sql.Rows) (*models.Log, error), error)
}

// LogRepository repository to work with models.Log entity.
type LogRepository struct {
	repositories.BaseRepositoryProvider
}

// NewLogRepository creates a repository to work with models.Log entity.
func NewLogRepository(db *gorm.DB) *LogRepository {
	return &LogRepository{
		repositories.NewBaseRepository(db),
	}
}

// GetLogsByNamespaceIDAndRunID returns logs by Namespace ID and Run ID.
func (r LogRepository) GetLogsByNamespaceIDAndRunID(
	ctx context.Context, namespaceID uint, runID string,
) (*sql.Rows, func(rows *sql.Rows) (*models.Log, error), error) {
	rows, err := r.GetDB().WithContext(ctx).Model(
		&models.Log{},
	).Joins(
		"LEFT JOIN runs ON runs.run_uuid = logs.run_uuid",
	).Joins(
		"LEFT JOIN experiments ON experiments.experiment_id = runs.experiment_id",
	).Where(
		"experiments.namespace_id = ?", namespaceID,
	).Rows()
	if err != nil {
		return nil, nil, eris.Wrap(err, "error getting run logs")
	}
	if err := rows.Err(); err != nil {
		return nil, nil, eris.Wrap(err, "error getting query result")
	}

	return rows, func(rows *sql.Rows) (*models.Log, error) {
		var runLog models.Log
		if err := r.GetDB().ScanRows(rows, &runLog); err != nil {
			return nil, eris.Wrapf(err, "error getting logs by run id: %s", runID)
		}
		return &runLog, nil
	}, nil
}
