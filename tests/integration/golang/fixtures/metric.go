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
		baseFixtures:     baseFixtures{db: db},
		metricRepository: repositories.NewMetricRepository(db),
	}, nil
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
