package repositories

import (
	"context"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/database"
)

// MetricRepositoryProvider provides an interface to work with models.Metric entity.
type MetricRepositoryProvider interface {
	BaseRepositoryProvider
	// CreateBatch creates []models.Metric entities in batch.
	CreateBatch(ctx context.Context, run *models.Run, batchSize int, params []models.Metric) error
	// GetMetricHistoryBulk returns metrics history bulk.
	GetMetricHistoryBulk(ctx context.Context, runIDs []string, key string, limit int) ([]models.Metric, error)
	// GetMetricHistoryByRunIDAndKey returns metrics history by RunID and Key.
	GetMetricHistoryByRunIDAndKey(ctx context.Context, runID, key string) ([]models.Metric, error)
}

// MetricRepository repository to work with models.Metric entity.
type MetricRepository struct {
	BaseRepository
}

// NewMetricRepository creates repository to work with models.Metric entity.
func NewMetricRepository(db *gorm.DB) *MetricRepository {
	return &MetricRepository{
		BaseRepository{
			db: db,
		},
	}
}

// CreateBatch creates []models.Metric entities in batch.
func (r MetricRepository) CreateBatch(
	ctx context.Context, run *models.Run, batchSize int, metrics []models.Metric,
) error {
	if len(metrics) == 0 {
		return nil
	}

	metricKeys := make([]string, 0)
	for _, m := range metrics {
		metricKeys = append(metricKeys, m.Key)
	}

	// get the latest metrics by requested Run ID and metric keys.
	lastMetrics, err := r.getLatestMetricsByRunIDAndKeys(ctx, run.ID, metricKeys)
	if err != nil {
		return eris.Wrap(err, "error getting latest metrics")
	}

	lastIters := make(map[string]int64)
	for _, lastMetric := range lastMetrics {
		lastIters[lastMetric.Key] = lastMetric.LastIter
	}

	latestMetrics := make(map[string]models.LatestMetric)
	for n, metric := range metrics {
		metrics[n].Iter = lastIters[metric.Key] + 1
		lastIters[metric.Key] = metric.Iter
		lm, ok := latestMetrics[metric.Key]
		if !ok ||
			metric.Step > lm.Step ||
			(metric.Step == lm.Step && metric.Timestamp > lm.Timestamp) ||
			(metric.Step == lm.Step && metric.Timestamp == lm.Timestamp && metric.Value > lm.Value) {
			latestMetrics[metric.Key] = models.LatestMetric{
				RunID:     metric.RunID,
				Key:       metric.Key,
				Value:     metric.Value,
				Timestamp: metric.Timestamp,
				Step:      metric.Step,
				IsNan:     metric.IsNan,
				LastIter:  metric.Iter,
			}
		}
	}

	if err := r.db.WithContext(ctx).CreateInBatches(&metrics, batchSize).Error; err != nil {
		return eris.Wrapf(err, "error creating metrics for run: %s", run.ID)
	}

	// TODO update latest metrics in the background?

	currentLatestMetricsMap := make(map[string]models.LatestMetric, len(latestMetrics))
	for _, m := range latestMetrics {
		currentLatestMetricsMap[m.Key] = m
	}

	updatedLatestMetrics := make([]models.LatestMetric, 0, len(latestMetrics))
	for k, m := range latestMetrics {
		lm, ok := currentLatestMetricsMap[k]
		if !ok ||
			m.Step > lm.Step ||
			(m.Step == lm.Step && m.Timestamp > lm.Timestamp) ||
			(m.Step == lm.Step && m.Timestamp == lm.Timestamp && m.Value > lm.Value) {
			updatedLatestMetrics = append(updatedLatestMetrics, m)
		} else {
			lm.LastIter = lastIters[k]
			updatedLatestMetrics = append(updatedLatestMetrics, lm)
		}
	}

	if len(updatedLatestMetrics) > 0 {
		if err := database.DB.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&updatedLatestMetrics).Error; err != nil {
			return eris.Wrapf(err, "error updating latest metrics for run: %s", run.ID)
		}
	}

	return nil
}

// getLatestMetricsByRunIDAndKeys returns the latest metrics by requested Run ID and keys.
func (r MetricRepository) getLatestMetricsByRunIDAndKeys(
	ctx context.Context, runID string, keys []string,
) ([]models.LatestMetric, error) {
	var metrics []models.LatestMetric
	if err := r.db.WithContext(ctx).Where(
		"run_uuid = ?", runID,
	).Where(
		"key IN ?", keys,
	).Find(&metrics).Error; err != nil {
		return nil, eris.Wrapf(err, "error getting latest metrics by run id: %s and keys: %v", runID, keys)
	}
	return metrics, nil
}

// GetMetricHistoryByRunIDAndKey returns metrics history by RunID and Key.
func (r MetricRepository) GetMetricHistoryByRunIDAndKey(
	ctx context.Context, runID, key string,
) ([]models.Metric, error) {
	var metrics []models.Metric
	if err := r.db.WithContext(ctx).Where(
		"run_uuid = ?", runID,
	).Where(
		"key = ?", key,
	).Find(&metrics).Error; err != nil {
		return nil, eris.Wrapf(err, "error getting metric history by run id: %s and key: %s", runID, key)
	}
	return metrics, nil
}

// GetMetricHistoryBulk returns metrics history bulk.
func (r MetricRepository) GetMetricHistoryBulk(
	ctx context.Context, runIDs []string, key string, limit int,
) ([]models.Metric, error) {
	var metrics []models.Metric
	if err := r.db.WithContext(ctx).Where(
		"run_uuid IN ?", runIDs,
	).Where(
		"key = ?", key,
	).Order(
		"run_uuid",
	).Order(
		"timestamp",
	).Order(
		"step",
	).Order(
		"value",
	).Limit(
		limit,
	).Find(
		&metrics,
	).Error; err != nil {
		return nil, eris.Wrapf(err, "error getting metric history by run ids: %v and key: %s", runIDs, key)
	}
	return metrics, nil
}
