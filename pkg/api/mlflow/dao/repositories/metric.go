package repositories

import (
	"context"
	"database/sql"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/database"
)

const (
	MetricHistoriesDefaultLimit   = 10000000
	MetricHistoryBulkDefaultLimit = 25000
)

// MetricRepositoryProvider provides an interface to work with models.Metric entity.
type MetricRepositoryProvider interface {
	BaseRepositoryProvider
	// CreateBatch creates []models.Metric entities in batch.
	CreateBatch(ctx context.Context, run *models.Run, batchSize int, params []models.Metric) error
	// GetMetricHistories returns metric histories by request parameters.
	GetMetricHistories(
		ctx context.Context,
		namespaceID uint,
		experimentIDs []string, runIDs []string, metricKeys []string,
		viewType request.ViewType,
		limit int32,
	) (*sql.Rows, func(*sql.Rows, interface{}) error, error)
	// GetMetricHistoryBulk returns metrics history bulk.
	GetMetricHistoryBulk(
		ctx context.Context, namespaceID uint, runIDs []string, key string, limit int,
	) ([]models.Metric, error)
	// GetMetricHistoryByRunIDAndKey returns metrics history by RunID and Key.
	GetMetricHistoryByRunIDAndKey(ctx context.Context, runID, key string) ([]models.Metric, error)
	// CreateContext creates new models.Context entity.
	CreateContext(ctx context.Context, context *models.Context) error
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
// TODO:get back and fix `gocyclo` problem.
//
//nolint:gocyclo
func (r MetricRepository) CreateBatch(
	ctx context.Context, run *models.Run, batchSize int, metrics []models.Metric,
) error {
	if len(metrics) == 0 {
		return nil
	}

	metricKeysMap := make(map[string]any)
	for _, m := range metrics {
		metricKeysMap[m.Key] = nil
	}
	metricKeys := make([]string, 0, len(metricKeysMap))
	for k := range metricKeysMap {
		metricKeys = append(metricKeys, k)
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
		lastIters[metric.Key] = metrics[n].Iter
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

	if err := r.db.WithContext(ctx).Clauses(
		clause.OnConflict{DoNothing: true},
	).CreateInBatches(&metrics, batchSize).Error; err != nil {
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
		if err := r.db.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&updatedLatestMetrics).Error; err != nil {
			return eris.Wrapf(err, "error updating latest metrics for run: %s", run.ID)
		}
	}

	return nil
}

// GetMetricHistories returns metric histories by request parameters.
// TODO think about to use interface instead of underlying type for -> func(*sql.Rows, interface{})
func (r MetricRepository) GetMetricHistories(
	ctx context.Context,
	namespaceID uint,
	experimentIDs []string, runIDs []string, metricKeys []string,
	viewType request.ViewType,
	limit int32,
) (*sql.Rows, func(*sql.Rows, interface{}) error, error) {
	// if experimentIDs has been provided then firstly get the runs by provided experimentIDs.
	if len(experimentIDs) > 0 {
		query := r.db.WithContext(ctx).Model(
			&database.Run{},
		).Joins(
			"INNER JOIN experiments ON experiments.experiment_id = runs.experiment_id AND experiments.namespace_id = ?",
			namespaceID,
		).Where(
			"runs.experiment_id IN ?", experimentIDs,
		)

		switch viewType {
		case request.ViewTypeActiveOnly, "":
			query.Where("runs.lifecycle_stage IN ?", []models.LifecycleStage{
				models.LifecycleStageActive,
			})
		case request.ViewTypeDeletedOnly:
			query.Where("runs.lifecycle_stage IN ?", []models.LifecycleStage{
				models.LifecycleStageDeleted,
			})
		case request.ViewTypeAll:
			query.Where("runs.lifecycle_stage IN ?", []models.LifecycleStage{
				models.LifecycleStageActive,
				models.LifecycleStageDeleted,
			})
		}
		if err := query.Pluck("run_uuid", &runIDs).Error; err != nil {
			return nil, nil, eris.Wrapf(
				err, "error getting runs by experimentIDs: %v, viewType: %s", experimentIDs, viewType,
			)
		}
	}

	// if experimentIDs has been provided then runIDs contains values from previous step,
	// otherwise runIDs may or may not contain values.
	query := r.db.WithContext(ctx).Model(
		&database.Metric{},
	).Where(
		"metrics.run_uuid IN ?", runIDs,
	).Joins(
		"JOIN runs on runs.run_uuid = metrics.run_uuid",
	).Joins(
		"INNER JOIN experiments ON experiments.experiment_id = runs.experiment_id AND experiments.namespace_id = ?",
		namespaceID,
	).Order(
		"runs.start_time DESC",
	).Order(
		"metrics.run_uuid",
	).Order(
		"metrics.key",
	).Order(
		"metrics.step",
	).Order(
		"metrics.timestamp",
	).Order(
		"metrics.value",
	)

	if limit == 0 {
		limit = MetricHistoriesDefaultLimit
	}
	query.Limit(int(limit))

	if len(metricKeys) > 0 {
		query.Where("metrics.key IN ?", metricKeys)
	}

	rows, err := query.Rows()
	if err != nil {
		return nil, nil, eris.Wrapf(
			err, "error getting metrics by experimentIDs: %v, runIDs: %v, metricKeys: %v, viewType: %s",
			experimentIDs,
			runIDs,
			metricKeys,
			viewType,
		)
	}
	return rows, r.db.ScanRows, nil
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
	ctx context.Context, namespaceID uint, runIDs []string, key string, limit int,
) ([]models.Metric, error) {
	var metrics []models.Metric
	query := r.db.WithContext(ctx).Where(
		"runs.run_uuid IN ?", runIDs,
	).Joins(
		"LEFT JOIN runs ON runs.run_uuid = metrics.run_uuid",
	).Joins(
		"INNER JOIN experiments ON experiments.experiment_id = runs.experiment_id AND experiments.namespace_id = ?",
		namespaceID,
	).Where(
		"key = ?", key,
	).Order(
		"metrics.run_uuid",
	).Order(
		"metrics.timestamp",
	).Order(
		"metrics.step",
	).Order(
		"metrics.value",
	)

	if limit == 0 {
		limit = MetricHistoryBulkDefaultLimit
	}
	query.Limit(limit)

	if err := query.Find(
		&metrics,
	).Error; err != nil {
		return nil, eris.Wrapf(err, "error getting metric history by run ids: %v and key: %s", runIDs, key)
	}
	return metrics, nil
}

// CreateContext creates new models.Context entity.
func (r MetricRepository) CreateContext(ctx context.Context, context *models.Context) error {
	if err := r.db.WithContext(ctx).Where(context).FirstOrCreate(context).Error; err != nil {
		return eris.Wrapf(err, "error creating or retrieving context.")
	}
	return nil
}
