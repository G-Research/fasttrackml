package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/aim/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim/common"
	"github.com/G-Research/fasttrackml/pkg/api/aim/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/aim/query"
	"github.com/G-Research/fasttrackml/pkg/common/dao/repositories"
	"github.com/G-Research/fasttrackml/pkg/common/dao/types"
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
	repositories.BaseRepositoryProvider
	// GetMetricKeysAndContextsByExperiments returns metric keys and contexts by provided experiments.
	GetMetricKeysAndContextsByExperiments(
		ctx context.Context, namespaceID uint, experiments []int,
	) ([]models.LatestMetric, error)
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
	repositories.BaseRepositoryProvider
}

// NewMetricRepository creates repository to work with models.Metric entity.
func NewMetricRepository(db *gorm.DB) *MetricRepository {
	return &MetricRepository{
		repositories.NewBaseRepository(db),
	}
}

// GetMetricKeysAndContextsByExperiments returns metric keys and contexts by provided experiments.
func (r MetricRepository) GetMetricKeysAndContextsByExperiments(
	ctx context.Context, namespaceID uint, experiments []int,
) ([]models.LatestMetric, error) {
	query := r.GetDB().WithContext(ctx).Distinct().Select(
		"key", "context_id",
	).Model(
		&models.LatestMetric{},
	).Joins(
		"JOIN runs USING(run_uuid)",
	).Joins(
		"INNER JOIN experiments ON experiments.experiment_id = runs.experiment_id AND experiments.namespace_id = ?",
		namespaceID,
	).Preload(
		"Context",
	).Where(
		"runs.lifecycle_stage = ?", models.LifecycleStageActive,
	)
	if len(experiments) != 0 {
		query = query.Where("experiments.experiment_id IN ?", experiments)
	}
	var metrics []models.LatestMetric
	if err := query.Find(&metrics).Error; err != nil {
		return nil, eris.Wrap(err, "error getting metrics by provided experiments")
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
		Dialector: r.GetDB().Dialector.Name(),
	}
	pq, err := qp.Parse(req.Query)
	if err != nil {
		return nil, 0, nil, err
	}

	if req.Metrics == nil || len(req.Metrics) == 0 {
		return nil, 0, nil, eris.New("No metrics are selected")
	}

	var totalRuns int64
	if err := r.GetDB().WithContext(ctx).Model(&models.Run{}).Count(&totalRuns).Error; err != nil {
		return nil, 0, nil, eris.Wrap(err, "error counting metrics")
	}

	var runs []models.Run
	if tx := r.GetDB().WithContext(ctx).
		InnerJoins(
			"Experiment",
			r.GetDB().WithContext(ctx).Select(
				"ID", "Name",
			).Where(&models.Experiment{NamespaceID: namespaceID}),
		).
		Preload("Params").
		Preload("Tags").
		Where("run_uuid IN (?)", pq.Filter(r.GetDB().WithContext(ctx).
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
			params[p.Key] = p.ValueAny()
		}
		tags := make(map[string]string, len(r.Tags))
		for _, t := range r.Tags {
			tags[t.Key] = t.Value
		}
		params["tags"] = tags
		run["params"] = params

		result[r.ID] = SearchResult{int64(r.RowNum), run}
	}

	ids, err := r.findContextIDs(ctx, &req)
	if err != nil {
		return nil, 0, nil, eris.Wrap(err, "error finding context ids")
	}

	var metricKeyContextConditionSlice []string
	for i, tuple := range req.Metrics {
		condition := fmt.Sprintf("(latest_metrics.key = '%s' AND contexts.id = %d)", tuple.Key, ids[i])
		metricKeyContextConditionSlice = append(metricKeyContextConditionSlice, condition)
	}
	metricKeyContextCondition := strings.Join(metricKeyContextConditionSlice, " OR ")

	subQuery := r.GetDB().WithContext(ctx).
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
		Joins("LEFT JOIN contexts ON latest_metrics.context_id = contexts.id").
		Where(metricKeyContextCondition)

	tx := r.GetDB().WithContext(ctx).
		Select(`
        metrics.*,
        runmetrics.context_json`,
		).
		Table("metrics").
		Joins(
			"INNER JOIN (?) runmetrics USING(run_uuid, key, context_id)",
			pq.Filter(subQuery),
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
	query := r.GetDB().WithContext(ctx)
	for _, context := range contextsMap {
		query = query.Or("contexts.json = ?", context)
	}
	var contexts []models.Context
	if err := query.Find(&contexts).Error; err != nil {
		return nil, eris.Wrap(err, "error getting contexts information")
	}
	return contexts, nil
}

func (r MetricRepository) findContextIDs(ctx context.Context, req *request.SearchMetricsRequest) ([]uint, error) {
	contextList := []types.JSONB{}
	contextsMap := map[string]types.JSONB{}
	for _, r := range req.Metrics {
		data := types.JSONB(r.Context)
		contextList = append(contextList, data)
		contextsMap[string(data)] = data
	}

	contexts, err := r.GetContextListByContextObjects(ctx, contextsMap)
	if err != nil {
		return nil, fmt.Errorf("error getting context list: %w", err)
	}

	ids := make([]uint, len(contextList))
	for _, context := range contexts {
		for i := 0; i < len(contextList); i++ {
			if common.CompareJson(contextList[i], context.Json) {
				ids[i] = context.ID
			}
		}
	}

	return ids, nil
}
