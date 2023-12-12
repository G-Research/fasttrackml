package fixtures

import (
	"context"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"

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
func NewMetricFixtures(db *gorm.DB) (*MetricFixtures, error) {
	return &MetricFixtures{
		baseFixtures:     baseFixtures{db: db},
		metricRepository: repositories.NewMetricRepository(db),
	}, nil
}

// CreateMetric creates new test Metric.
func (f MetricFixtures) CreateMetric(ctx context.Context, metric *models.Metric) (*models.Metric, error) {
	if metric.Context != nil {
		if err := f.baseFixtures.db.WithContext(ctx).FirstOrCreate(
			&metric.Context, metric.Context,
		).Error; err != nil {
			return nil, eris.Wrap(err, "error creating metric context")
		}
		metric.ContextID = &metric.Context.ID
	}
	if err := f.baseFixtures.db.WithContext(ctx).Create(metric).Error; err != nil {
		return nil, eris.Wrap(err, "error creating metric")
	}
	return metric, nil
}

// GetMetricsByRunID returns the metrics by Run ID.
func (f MetricFixtures) GetMetricsByRunID(ctx context.Context, runID string) ([]*models.Metric, error) {
	var metrics []*models.Metric
	if err := f.db.WithContext(ctx).Where(
		"run_uuid = ?", runID,
	).Find(&metrics).Error; err != nil {
		return nil, eris.Wrapf(err, "error getting metric by run_uuid: %v", runID)
	}
	return metrics, nil
}

// GetMetricsByContext returns metric by a context partial match.
func (f MetricFixtures) GetMetricsByContext(
	ctx context.Context,
	metricContext map[string]string,
) ([]*models.Metric, error) {
	var metrics []*models.Metric
	tx := f.db.WithContext(ctx).Model(
		&database.Metric{},
	).Joins(
		"LEFT JOIN contexts ON metrics.context_id = contexts.id",
	)
	sql, args := repositories.BuildJsonCondition(tx.Dialector.Name(), "contexts.json", metricContext)
	if err := tx.Where(sql, args...).Find(&metrics).Error; err != nil {
		return nil, eris.Wrapf(err, "error getting metrics by context: %v", metricContext)
	}
	return metrics, nil
}

// CreateLatestMetric creates new test Latest Metric.
func (f MetricFixtures) CreateLatestMetric(
	ctx context.Context, metric *models.LatestMetric,
) (*models.LatestMetric, error) {
	if metric.Context != nil {
		if err := f.baseFixtures.db.WithContext(ctx).FirstOrCreate(
			&metric.Context, metric.Context,
		).Error; err != nil {
			return nil, eris.Wrap(err, "error creating latest metric context")
		}
		metric.ContextID = &metric.Context.ID
	}
	if err := f.baseFixtures.db.WithContext(ctx).Create(metric).Error; err != nil {
		return nil, eris.Wrap(err, "error creating latest metric")
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

// GetLatestMetricByRunID returns the latest metric by provide Run ID.
func (f MetricFixtures) GetLatestMetricByRunID(ctx context.Context, runID string) (*models.LatestMetric, error) {
	var metric models.LatestMetric
	if err := f.db.WithContext(ctx).Preload("Context").Where(
		"run_uuid = ?", runID,
	).First(&metric).Error; err != nil {
		return nil, eris.Wrapf(err, "error getting latest metric by run_uuid: %v", runID)
	}
	return &metric, nil
}
