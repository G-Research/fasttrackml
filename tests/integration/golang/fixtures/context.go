package fixtures

import (
	"context"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

// ContextFixtures represents data fixtures object.
type ContextFixtures struct {
	baseFixtures
}

// NewContextFixtures creates new instance of ContextFixtures.
func NewContextFixtures(db *gorm.DB) (*ContextFixtures, error) {
	return &ContextFixtures{
		baseFixtures: baseFixtures{db: db},
	}, nil
}

// CreateContext creates new test Context.
func (f ContextFixtures) CreateContext(ctx context.Context, context *models.Context) (*models.Context, error) {
	if err := f.baseFixtures.db.WithContext(ctx).Create(context).Error; err != nil {
		return nil, eris.Wrap(err, "error creating test context")
	}
	return context, nil
}

// GetContextByMetricKey returns the associated Context of a Metric.
func (f ContextFixtures) GetContextByMetricKey(ctx context.Context, key string) (*models.Context, error) {
	var metric models.Metric
	if err := f.db.WithContext(ctx).Preload("Context").Where(
		"key = ?", key,
	).Find(&metric).Error; err != nil {
		return nil, eris.Wrapf(err, "error getting metric by key: %s ", key)
	}

	return metric.Context, nil
}
