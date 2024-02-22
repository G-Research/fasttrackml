package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/G-Research/fasttrackml/pkg/api/aim2/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/query"
	"github.com/G-Research/fasttrackml/pkg/common/db/types"
	"github.com/G-Research/fasttrackml/pkg/database"
)

const (
	MetricHistoriesDefaultLimit   = 10000000
	MetricHistoryBulkDefaultLimit = 25000
)

// SearchResult is a helper for reporting result progress.
type SearchResult struct {
	RowNum int64
	Info   fiber.Map
}

// SearchResultMap is a helper for reporting result progress.
type SearchResultMap = map[string]SearchResult

// MetricRepositoryProvider provides an interface to work with models.Metric entity.
type MetricRepositoryProvider interface {
	BaseRepositoryProvider
	// CreateBatch creates []models.Metric entities in batch.
	CreateBatch(ctx context.Context, run *models.Run, batchSize int, params []models.Metric) error
	// GetLatestMetricsByExperiments returns latest metrics by provided experiments.
	GetLatestMetricsByExperiments(
		ctx context.Context, namespaceID uint, experiments []int,
	) ([]models.LatestMetric, error)
	// GetMetricHistoryBulk returns metrics history bulk.
	GetMetricHistoryBulk(
		ctx context.Context, namespaceID uint, runIDs []string, key string, limit int,
	) ([]models.Metric, error)
	// GetMetricHistoryByRunIDAndKey returns metrics history by RunID and Key.
	GetMetricHistoryByRunIDAndKey(ctx context.Context, runID, key string) ([]models.Metric, error)
	// SearchMetrics returns a sql.Rows cursor for streaming the metrics matching the request.
	SearchMetrics(
		ctx context.Context, namespaceID uint, timeZoneOffset int, req request.SearchMetricsRequest,
	) (*sql.Rows, int64, SearchResultMap, error)
	// GetContextListByContextObjects returns list of context by provided map of contexts.
	GetContextListByContextObjects(
		ctx context.Context, contextsMap map[string]types.JSONB,
	) ([]models.Context, error)
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
		lastIters[lastMetric.UniqueKey()] = lastMetric.LastIter
	}
	allContexts := make([]*models.Context, len(metrics))
	uniqueContexts := make([]*models.Context, 0, len(metrics))
	contextProcessed := make(map[string]*models.Context)
	latestMetrics := make(map[string]models.LatestMetric)
	for n := range metrics {
		ctxHash := metrics[n].Context.GetJsonHash()
		ctxRef, ok := contextProcessed[ctxHash]
		if ok {
			allContexts[n] = ctxRef
		} else {
			uniqueContexts = append(uniqueContexts, &metrics[n].Context)
			allContexts[n] = &metrics[n].Context
			contextProcessed[ctxHash] = &metrics[n].Context
		}
	}

	if err := r.db.WithContext(ctx).Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "json"}},
			UpdateAll: true,
		},
	).CreateInBatches(&uniqueContexts, batchSize).Error; err != nil {
		return eris.Wrapf(err, "error creating contexts")
	}

	for n := range metrics {
		metrics[n].ContextID = allContexts[n].ID
		metrics[n].Context = *allContexts[n]
		metrics[n].Iter = lastIters[metrics[n].UniqueKey()] + 1
		lastIters[metrics[n].UniqueKey()] = metrics[n].Iter
		lm, ok := latestMetrics[metrics[n].UniqueKey()]
		if !ok ||
			metrics[n].Step > lm.Step ||
			(metrics[n].Step == lm.Step && metrics[n].Timestamp > lm.Timestamp) ||
			(metrics[n].Step == lm.Step && metrics[n].Timestamp == lm.Timestamp && metrics[n].Value > lm.Value) {
			latestMetrics[metrics[n].UniqueKey()] = models.LatestMetric{
				RunID:     metrics[n].RunID,
				Key:       metrics[n].Key,
				Value:     metrics[n].Value,
				Timestamp: metrics[n].Timestamp,
				Step:      metrics[n].Step,
				IsNan:     metrics[n].IsNan,
				LastIter:  metrics[n].Iter,
				ContextID: allContexts[n].ID,
				Context:   *allContexts[n],
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
	for k, m := range latestMetrics {
		currentLatestMetricsMap[k] = m
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
			Columns:   []clause.Column{{Name: "run_uuid"}, {Name: "key"}, {Name: "context_id"}},
			UpdateAll: true,
		}).Create(&updatedLatestMetrics).Error; err != nil {
			return eris.Wrapf(err, "error updating latest metrics for run: %s", run.ID)
		}
	}
	return nil
}

// GetLatestMetricsByExperiments returns latest metrics by provided experiments.
func (r MetricRepository) GetLatestMetricsByExperiments(
	ctx context.Context, namespaceID uint, experiments []int,
) ([]models.LatestMetric, error) {
	query := r.db.WithContext(ctx).Distinct().Model(
		&database.LatestMetric{},
	).Joins(
		"JOIN runs USING(run_uuid)",
	).Joins(
		"INNER JOIN experiments ON experiments.experiment_id = runs.experiment_id AND experiments.namespace_id = ?",
		namespaceID,
	).Joins(
		"Context",
	).Where(
		"runs.lifecycle_stage = ?", database.LifecycleStageActive,
	)
	if len(experiments) != 0 {
		query.Where("experiments.experiment_id IN ?", experiments)
	}
	var metrics []models.LatestMetric
	if err := query.Find(&metrics).Error; err != nil {
		return nil, eris.Wrap(err, "error getting metrics by provided experiments")
	}
	return metrics, nil
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
	if err := r.db.WithContext(
		ctx,
	).Joins(
		"Context",
	).Where(
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

// SearchMetrics returns a metrics cursor according to the SearchMetricsRequest.
func (r MetricRepository) SearchMetrics(
	ctx context.Context, namespaceID uint, timeZoneOffset int, req request.SearchMetricsRequest,
) (*sql.Rows, int64, SearchResultMap, error) {
	qp := query.QueryParser{
		Default: query.DefaultExpression{
			Contains:   "run.archived",
			Expression: "not run.archived",
		},
		Tables: map[string]string{
			"runs":        "runs",
			"experiments": "experiments",
			"metrics":     "latest_metrics",
		},
		TzOffset:  timeZoneOffset,
		Dialector: r.db.Dialector.Name(),
	}
	pq, err := qp.Parse(req.Query)
	if err != nil {
		return nil, 0, nil, err
	}

	if !pq.IsMetricSelected() {
		return nil, 0, nil, eris.New("No metrics are selected")
	}

	var totalRuns int64
	if err := r.db.WithContext(ctx).Model(&models.Run{}).Count(&totalRuns).Error; err != nil {
		return nil, 0, nil, eris.Wrap(err, "error counting metrics")
	}

	var runs []models.Run
	if tx := r.db.WithContext(ctx).
		InnerJoins(
			"Experiment",
			r.db.WithContext(ctx).Select(
				"ID", "Name",
			).Where(&models.Experiment{NamespaceID: namespaceID}),
		).
		Preload("Params").
		Preload("Tags").
		Where("run_uuid IN (?)", pq.Filter(r.db.WithContext(ctx).
			Select("runs.run_uuid").
			Table("runs").
			Joins(
				"INNER JOIN experiments ON experiments.experiment_id = runs.experiment_id AND experiments.namespace_id = ?",
				namespaceID,
			).
			Joins("JOIN latest_metrics USING(run_uuid)").
			Joins("JOIN contexts ON latest_metrics.context_id = contexts.id"),
		)).
		Order("runs.row_num DESC").
		Find(&runs); tx.Error != nil {
		return nil, 0, nil, eris.Wrap(err, "error searching metrics")
	}

	result := make(SearchResultMap, len(runs))
	for _, r := range runs {
		run := fiber.Map{
			"props": fiber.Map{
				"name":        r.Name,
				"description": nil,
				"experiment": fiber.Map{
					"id":   fmt.Sprintf("%d", *r.Experiment.ID),
					"name": r.Experiment.Name,
				},
				"tags":          []string{}, // TODO insert real tags
				"creation_time": float64(r.StartTime.Int64) / 1000,
				"end_time":      float64(r.EndTime.Int64) / 1000,
				"archived":      r.LifecycleStage == models.LifecycleStageDeleted,
				"active":        r.Status == models.StatusRunning,
			},
		}

		params := make(fiber.Map, len(r.Params)+1)
		for _, p := range r.Params {
			params[p.Key] = p.Value
		}
		tags := make(map[string]string, len(r.Tags))
		for _, t := range r.Tags {
			tags[t.Key] = t.Value
		}
		params["tags"] = tags
		run["params"] = params

		result[r.ID] = SearchResult{int64(r.RowNum), run}
	}

	tx := r.db.WithContext(ctx).
		Select(`
			metrics.*,
			runmetrics.context_json`,
		).
		Table("metrics").
		Joins(
			"INNER JOIN (?) runmetrics USING(run_uuid, key, context_id)",
			pq.Filter(r.db.WithContext(ctx).
				Select(
					"runs.run_uuid",
					"runs.row_num",
					"latest_metrics.key",
					"latest_metrics.context_id",
					"contexts.json AS context_json",
					fmt.Sprintf("(latest_metrics.last_iter + 1)/ %f AS interval", float32(req.Steps)),
				).
				Table("runs").
				Joins(
					"INNER JOIN experiments ON experiments.experiment_id = runs.experiment_id AND experiments.namespace_id = ?",
					namespaceID,
				).
				Joins("LEFT JOIN latest_metrics USING(run_uuid)").
				Joins("LEFT JOIN contexts ON latest_metrics.context_id = contexts.id")),
		).
		Where("MOD(metrics.iter + 1 + runmetrics.interval / 2, runmetrics.interval) < 1").
		Order("runmetrics.row_num DESC").
		Order("metrics.key").
		Order("metrics.context_id").
		Order("metrics.iter")

	if req.XAxis != "" {
		tx.
			Select("metrics.*", "runmetrics.context_json", "x_axis.value as x_axis_value", "x_axis.is_nan as x_axis_is_nan").
			Joins(
				"LEFT JOIN metrics x_axis ON metrics.run_uuid = x_axis.run_uuid AND "+
					"metrics.iter = x_axis.iter AND x_axis.context_id = metrics.context_id AND x_axis.key = ?",
				req.XAxis,
			)
	}
	rows, err := tx.Rows()
	if err != nil {
		return nil, 0, nil, eris.Wrap(err, "error searching metrics")
	}
	if err := rows.Err(); err != nil {
		return nil, 0, nil, eris.Wrap(err, "error getting metrics rows cursor")
	}

	return rows, totalRuns, result, nil
}

// GetContextListByContextObjects returns list of context by provided map of contexts.
func (r MetricRepository) GetContextListByContextObjects(
	ctx context.Context, contextsMap map[string]types.JSONB,
) ([]models.Context, error) {
	query := r.db.WithContext(ctx)
	for _, context := range contextsMap {
		query = query.Or("contexts.json = ?", context)
	}
	var contexts []models.Context
	if err := query.Find(&contexts).Error; err != nil {
		return nil, eris.Wrap(err, "error getting contexts information")
	}
	return contexts, nil
}
