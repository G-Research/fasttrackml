package fixtures

import (
	"context"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

// LogFixtures represents data fixtures object.
type LogFixtures struct {
	baseFixtures
}

// NewLogFixtures creates new instance of LogFixtures.
func NewLogFixtures(db *gorm.DB) (*LogFixtures, error) {
	return &LogFixtures{
		baseFixtures: baseFixtures{db: db},
	}, nil
}

// CreateLog creates new Log.
func (f LogFixtures) CreateLog(ctx context.Context, log *models.Log) (*models.Log, error) {
	if err := f.baseFixtures.db.WithContext(ctx).Create(log).Error; err != nil {
		return nil, eris.Wrap(err, "error creating test tag")
	}
	return log, nil
}

// GetByRunID returns log collection by requested Run ID.
func (f LogFixtures) GetByRunID(ctx context.Context, runID string) ([]models.Log, error) {
	var logs []models.Log
	if err := f.db.WithContext(ctx).Where(
		models.Log{RunID: runID},
	).Order("timestamp").Find(&logs).Error; err != nil {
		return nil, eris.Wrapf(err, "error getting logs by run id: %s", runID)
	}
	return logs, nil
}
