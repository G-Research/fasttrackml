package fixtures

import (
	"context"
	"time"

	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/repositories"
	"github.com/G-Research/fasttrackml/pkg/database"
)

// MetricFixtures represents data fixtures object.
type MetricFixtures struct {
	baseFixtures
	metricRepository repositories.MetricRepositoryProvider
}

// NewMetricFixtures creates new instance of MetricFixtures.
func NewMetricFixtures(databaseDSN string) (*MetricFixtures, error) {
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
	return &MetricFixtures{
		baseFixtures:     baseFixtures{db: db.DB},
		metricRepository: repositories.NewMetricRepository(db.DB),
	}, nil
}

// CreateMetric creates new test Metric.
func (f MetricFixtures) CreateMetric(ctx context.Context, metric *models.Metric) (*models.Metric, error) {
	if err := f.baseFixtures.db.WithContext(ctx).Create(metric).Error; err != nil {
		return nil, eris.Wrap(err, "error creating test metric")
	}
	return metric, nil
}

// CreateLatestMetric creates new test Latest Metric.
func (f MetricFixtures) CreateLatestMetric(
	ctx context.Context, metric *models.LatestMetric,
) (*models.LatestMetric, error) {
	if err := f.baseFixtures.db.WithContext(ctx).Create(metric).Error; err != nil {
		return nil, eris.Wrap(err, "error creating test latest metric")
	}
	return metric, nil
}

// GetLatestMetricByKey returns the latest metric by provided key.
func (f MetricFixtures) GetLatestMetricByKey(ctx context.Context, key string) (*models.LatestMetric, error) {
	var metric models.LatestMetric
	if err := f.db.WithContext(ctx).Where(
		"key = ?", key,
	).First(&metric).Error; err != nil {
		return nil, eris.Wrapf(err, "error getting latest metric by key: %v", key)
	}
	return &metric, nil
}
