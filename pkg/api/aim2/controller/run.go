package controller

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/aim/encoding"
	"github.com/G-Research/fasttrackml/pkg/api/aim/query"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/api/response"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/repositories"
	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/pkg/common/db/types"
	"github.com/G-Research/fasttrackml/pkg/common/middleware/namespace"
	"github.com/G-Research/fasttrackml/pkg/database"
)

// GetRunInfo handles `GET /runs/:id/info` endpoint.
func (c Controller) GetRunInfo(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("getRunInfo namespace: %s", ns.Code)

	req := request.GetRunInfoRequest{}
	if err := ctx.QueryParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}
	if err := ctx.ParamsParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	runInfo, err := c.runService.GetRunInfo(ctx.Context(), ns, &req)
	if err != nil {
		return err
	}

	resp := response.NewGetRunInfoResponse(runInfo)
	log.Debugf("getRunInfo response: %#v", resp)
	return ctx.JSON(resp)
}

// GetRunMetrics handles `GET /runs/:id/metric/get-batch` endpoint.
func (c Controller) GetRunMetrics(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("getRunMetrics namespace: %s", ns.Code)

	req := request.GetRunMetricsRequest{}
	if err := ctx.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	metrics, metricKeysMap, err := c.runService.GetRunMetrics(ctx.Context(), ns, ctx.Params("id"), &req)
	if err != nil {
		return err
	}

	resp := response.NewGetRunMetricsResponse(metrics, metricKeysMap)
	log.Debugf("getRunMetrics response: %#v", resp)
	return ctx.JSON(resp)
}

func (c Controller) GetRunsActive(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("getRunsActive namespace: %s", ns.Code)

	req := request.GetRunsActiveRequest{}
	if err := ctx.QueryParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	if ctx.Query("report_progress") == "" {
		req.ReportProgress = true
	}

	var runs []database.Run
	if err := database.DB.
		Where("status = ?", database.StatusRunning).
		InnerJoins(
			"Experiment",
			database.DB.Select(
				"ID", "Name",
			).Where(
				&models.Experiment{NamespaceID: ns.ID},
			),
		).
		Preload("LatestMetrics.Context").
		Limit(50).
		Order("start_time DESC").
		Find(&runs).Error; err != nil {
		return fmt.Errorf("error retrieving active runs: %w", err)
	}

	ctx.Set("Content-Type", "application/octet-stream")
	ctx.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		start := time.Now()
		if err := func() error {
			for i, r := range runs {
				props := fiber.Map{
					"name":        r.Name,
					"description": nil,
					"experiment": fiber.Map{
						"id":   fmt.Sprintf("%d", *r.Experiment.ID),
						"name": r.Experiment.Name,
					},
					"tags":          []string{}, // TODO insert real tags
					"creation_time": float64(r.StartTime.Int64) / 1000,
					"end_time":      float64(r.EndTime.Int64) / 1000,
					"archived":      r.LifecycleStage == database.LifecycleStageDeleted,
					"active":        r.Status == database.StatusRunning,
				}

				metrics := make([]fiber.Map, len(r.LatestMetrics))
				for i, m := range r.LatestMetrics {
					v := m.Value
					if m.IsNan {
						v = math.NaN()
					}
					data := fiber.Map{
						"name": m.Key,
						"last_value": fiber.Map{
							"dtype":      "float",
							"first_step": 0,
							"last_step":  m.LastIter,
							"last":       v,
							"version":    2,
							"context":    fiber.Map{},
						},
					}
					// to be properly decoded by AIM UI, json should be represented as a key:value object.
					context := fiber.Map{}
					if err := json.Unmarshal(m.Context.Json, &context); err != nil {
						return eris.Wrap(err, "error unmarshalling `context` json to `fiber.Map` object")
					}
					data["context"] = context
					metrics[i] = data
				}

				if err := encoding.EncodeTree(w, fiber.Map{
					r.ID: fiber.Map{
						"props": props,
						"traces": fiber.Map{
							"metric": metrics,
						},
					},
				}); err != nil {
					return err
				}

				if req.ReportProgress {
					if err := encoding.EncodeTree(w, fiber.Map{
						fmt.Sprintf("progress_%d", i): []int{i + 1, len(runs)},
					}); err != nil {
						return err
					}
				}

				if err := w.Flush(); err != nil {
					return err
				}
			}

			// if q.ReportProgress && err == nil {
			// 	err = encoding.EncodeTree(w, fiber.Map{
			// 		fmt.Sprintf("progress_%d", len(runs)): []int{len(runs), len(runs)},
			// 	})
			// 	if err != nil {
			// 		err = w.Flush()
			// 	}
			// }

			return nil
		}(); err != nil {
			log.Errorf("Error encountered in %s %s: error streaming active runs: %s", ctx.Method(), ctx.Path(), err)
		}

		log.Infof("body - %s %s %s", time.Since(start), ctx.Method(), ctx.Path())
	})

	return nil
}

// TODO:get back and fix `gocyclo` problem.
//
//nolint:gocyclo
func (c Controller) SearchRuns(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("searchRuns namespace: %s", ns.Code)

	req := request.SearchRunsRequest{}
	if err = ctx.QueryParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	if ctx.Query("report_progress") == "" {
		req.ReportProgress = true
	}

	tzOffset, err := strconv.Atoi(ctx.Get("x-timezone-offset", "0"))
	if err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "x-timezone-offset header is not a valid integer")
	}

	qp := query.QueryParser{
		Default: query.DefaultExpression{
			Contains:   "run.archived",
			Expression: "not run.archived",
		},
		Tables: map[string]string{
			"runs":        "runs",
			"experiments": "Experiment",
		},
		TzOffset:  tzOffset,
		Dialector: database.DB.Dialector.Name(),
	}
	pq, err := qp.Parse(req.Query)
	if err != nil {
		return err
	}

	var total int64
	if tx := database.DB.
		Model(&database.Run{}).
		Count(&total); tx.Error != nil {
		return fmt.Errorf("unable to count total runs: %w", tx.Error)
	}

	log.Debugf("Total runs: %d", total)

	tx := database.DB.
		InnerJoins(
			"Experiment",
			database.DB.Select(
				"ID", "Name",
			).Where(
				&models.Experiment{NamespaceID: ns.ID},
			),
		).
		Order("row_num DESC")

	if !req.ExcludeParams {
		tx.Preload("Params")
		tx.Preload("Tags")
	}

	if !req.ExcludeTraces {
		tx.Preload("LatestMetrics.Context")
	}

	switch req.Action {
	case "export":
		var runs []database.Run
		if err := pq.Filter(tx).Find(&runs).Error; err != nil {
			return fmt.Errorf("error searching runs: %w", err)
		}
		log.Debugf("found %d runs", len(runs))
		RunsSearchAsCSVResponse(ctx, runs, req.ExcludeTraces, req.ExcludeParams)
	default:
		if req.Limit > 0 {
			tx.Limit(req.Limit)
		}
		if req.Offset != "" {
			run := &database.Run{
				ID: req.Offset,
			}
			if err := database.DB.Select(
				"row_num",
			).First(
				&run,
			).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("unable to find search runs offset %q: %w", req.Offset, err)
			}
			tx.Where("row_num < ?", run.RowNum)
		}
		var runs []database.Run
		if err := pq.Filter(tx).Find(&runs).Error; err != nil {
			return fmt.Errorf("error searching runs: %w", err)
		}
		log.Debugf("found %d runs", len(runs))
		RunsSearchAsStreamResponse(ctx, runs, total, req.ExcludeTraces, req.ExcludeParams, req.ReportProgress)
	}

	return nil
}

// TODO:get back and fix `gocyclo` problem.
//
//nolint:gocyclo
func (c Controller) SearchMetrics(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("searchMetrics namespace: %s", ns.Code)

	req := request.SearchMetricsRequest{}
	if err = ctx.QueryParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	if ctx.Query("report_progress") == "" {
		req.ReportProgress = true
	}

	if ctx.Query("p") == "" {
		req.Steps = 50
	}

	tzOffset, err := strconv.Atoi(ctx.Get("x-timezone-offset", "0"))
	if err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "x-timezone-offset header is not a valid integer")
	}

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
		TzOffset:  tzOffset,
		Dialector: database.DB.Dialector.Name(),
	}
	pq, err := qp.Parse(req.Query)
	if err != nil {
		return err
	}

	if !pq.IsMetricSelected() {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "No metrics are selected")
	}

	var totalRuns int64
	if tx := database.DB.Model(&database.Run{}).Count(&totalRuns); tx.Error != nil {
		return fmt.Errorf("error searching run metrics: %w", tx.Error)
	}

	var runs []database.Run
	if tx := database.DB.
		InnerJoins(
			"Experiment",
			database.DB.Select(
				"ID", "Name",
			).Where(&models.Experiment{NamespaceID: ns.ID}),
		).
		Preload("Params").
		Preload("Tags").
		Where("run_uuid IN (?)", pq.Filter(database.DB.
			Select("runs.run_uuid").
			Table("runs").
			Joins(
				"INNER JOIN experiments ON experiments.experiment_id = runs.experiment_id AND experiments.namespace_id = ?",
				ns.ID,
			).
			Joins("JOIN latest_metrics USING(run_uuid)").
			Joins("JOIN contexts ON latest_metrics.context_id = contexts.id"),
		)).
		Order("runs.row_num DESC").
		Find(&runs); tx.Error != nil {
		return fmt.Errorf("error searching run metrics: %w", tx.Error)
	}

	result := make(map[string]struct {
		RowNum int64
		Info   fiber.Map
	}, len(runs))
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
				"archived":      r.LifecycleStage == database.LifecycleStageDeleted,
				"active":        r.Status == database.StatusRunning,
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

		result[r.ID] = struct {
			RowNum int64
			Info   fiber.Map
		}{int64(r.RowNum), run}
	}

	tx := database.DB.
		Select(`
			metrics.*,
			runmetrics.context_json`,
		).
		Table("metrics").
		Joins(
			"INNER JOIN (?) runmetrics USING(run_uuid, key, context_id)",
			pq.Filter(database.DB.
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
					ns.ID,
				).
				Joins("LEFT JOIN latest_metrics USING(run_uuid)").
				Joins("LEFT JOIN contexts ON latest_metrics.context_id = contexts.id")),
		).
		Where("MOD(metrics.iter + 1 + runmetrics.interval / 2, runmetrics.interval) < 1").
		Order("runmetrics.row_num DESC").
		Order("metrics.key").
		Order("metrics.context_id").
		Order("metrics.iter")

	var xAxis bool
	if req.XAxis != "" {
		tx.
			Select("metrics.*", "runmetrics.context_json", "x_axis.value as x_axis_value", "x_axis.is_nan as x_axis_is_nan").
			Joins(
				"LEFT JOIN metrics x_axis ON metrics.run_uuid = x_axis.run_uuid AND "+
					"metrics.iter = x_axis.iter AND x_axis.context_id = metrics.context_id AND x_axis.key = ?",
				req.XAxis,
			)
		xAxis = true
	}

	rows, err := tx.Rows()
	if err != nil {
		return fmt.Errorf("error searching run metrics: %w", err)
	}
	if err := rows.Err(); err != nil {
		return api.NewInternalError("error getting query result: %s", err)
	}

	ctx.Set("Content-Type", "application/octet-stream")
	ctx.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		//nolint:errcheck
		defer rows.Close()

		start := time.Now()
		if err := func() error {
			var (
				id          string
				key         string
				context     fiber.Map
				contextID   uint
				metrics     []fiber.Map
				values      []float64
				iters       []float64
				epochs      []float64
				timestamps  []float64
				xAxisValues []float64
				progress    int
			)
			reportProgress := func(cur int64) error {
				if !req.ReportProgress {
					return nil
				}
				err := encoding.EncodeTree(w, fiber.Map{
					fmt.Sprintf("progress_%d", progress): []int64{cur, totalRuns},
				})
				if err != nil {
					return err
				}
				progress++
				return w.Flush()
			}
			addMetrics := func() {
				if key != "" {
					metric := fiber.Map{
						"name":          key,
						"context":       context,
						"slice":         []int{0, 0, req.Steps},
						"values":        toNumpy(values),
						"iters":         toNumpy(iters),
						"epochs":        toNumpy(epochs),
						"timestamps":    toNumpy(timestamps),
						"x_axis_values": nil,
						"x_axis_iters":  nil,
					}
					if xAxis {
						metric["x_axis_values"] = toNumpy(xAxisValues)
						metric["x_axis_iters"] = metric["iters"]
					}
					metrics = append(metrics, metric)
				}
			}
			flushMetrics := func() error {
				if id == "" {
					return nil
				}
				if err := encoding.EncodeTree(w, fiber.Map{
					id: fiber.Map{
						"traces": metrics,
					},
				}); err != nil {
					return err
				}
				if err := reportProgress(totalRuns - result[id].RowNum); err != nil {
					return err
				}
				return w.Flush()
			}
			for rows.Next() {
				var metric struct {
					database.Metric
					Context    datatypes.JSON `gorm:"column:context_json"`
					XAxisValue float64        `gorm:"column:x_axis_value"`
					XAxisIsNaN bool           `gorm:"column:x_axis_is_nan"`
				}
				if err := database.DB.ScanRows(rows, &metric); err != nil {
					return err
				}

				if metric.Key != key || metric.RunID != id || metric.ContextID != contextID {
					addMetrics()

					if metric.RunID != id {
						if err := flushMetrics(); err != nil {
							return err
						}

						metrics = make([]fiber.Map, 0)

						if err := encoding.EncodeTree(w, fiber.Map{
							metric.RunID: result[metric.RunID].Info,
						}); err != nil {
							return err
						}

						id = metric.RunID
					}

					values = make([]float64, 0, req.Steps)
					iters = make([]float64, 0, req.Steps)
					epochs = make([]float64, 0, req.Steps)
					context = fiber.Map{}
					timestamps = make([]float64, 0, req.Steps)
					if xAxis {
						xAxisValues = make([]float64, 0, req.Steps)
					}
					key = metric.Key
				}

				v := metric.Value
				if metric.IsNan {
					v = math.NaN()
				}
				values = append(values, v)
				iters = append(iters, float64(metric.Iter))
				epochs = append(epochs, float64(metric.Step))
				timestamps = append(timestamps, float64(metric.Timestamp)/1000)
				if xAxis {
					x := metric.XAxisValue
					if metric.XAxisIsNaN {
						x = math.NaN()
					}
					xAxisValues = append(xAxisValues, x)
				}
				// to be properly decoded by AIM UI, json should be represented as a key:value object.
				if err := json.Unmarshal(metric.Context, &context); err != nil {
					return eris.Wrap(err, "error unmarshalling `context` json to `fiber.Map` object")
				}
				contextID = metric.ContextID
			}

			addMetrics()
			if err := flushMetrics(); err != nil {
				return err
			}

			if err := reportProgress(totalRuns); err != nil {
				return err
			}

			return nil
		}(); err != nil {
			log.Errorf("Error encountered in %s %s: error streaming metrics: %s", ctx.Method(), ctx.Path(), err)
		}

		log.Infof("body - %s %s %s", time.Since(start), ctx.Method(), ctx.Path())
	})

	return nil
}

// TODO:get back and fix `gocyclo` problem.
//
//nolint:gocyclo
func (c Controller) SearchAlignedMetrics(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("searchAlignedMetrics namespace: %s", ns.Code)

	req := request.SearchAlignedMetricsRequest{}
	if err := ctx.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	values, capacity, contextsMap := []any{}, 0, map[string]types.JSONB{}
	for _, r := range req.Runs {
		for _, t := range r.Traces {
			l := t.Slice[2]
			if l > capacity {
				capacity = l
			}
			// collect map of unique contexts.
			data, err := json.Marshal(t.Context)
			if err != nil {
				return api.NewInternalError("error serializing context: %s", err)
			}
			sum := sha256.Sum256(data)
			contextHash := fmt.Sprintf("%x", sum)
			_, ok := contextsMap[contextHash]
			if !ok {
				contextsMap[contextHash] = data
			}
			values = append(values, r.ID, t.Name, data, float32(l))
		}
	}

	// map context values to context ids
	query := database.DB
	for _, context := range contextsMap {
		query = query.Or("contexts.json = ?", context)
	}
	var contexts []database.Context
	if err := query.Find(&contexts).Error; err != nil {
		return api.NewInternalError("error getting context information: %s", err)
	}

	// add context ids to `values`
	for _, context := range contexts {
		for i := 2; i < len(values); i += 4 {
			if CompareJson(values[i].([]byte), context.Json) {
				values[i] = context.ID
			}
		}
	}

	var valuesStmt strings.Builder
	length := len(values) / 4
	for i := 0; i < length; i++ {
		valuesStmt.WriteString("(?, ?, CAST(? AS numeric), CAST(? AS numeric))")
		if i < length-1 {
			valuesStmt.WriteString(",")
		}
	}

	// TODO this should probably be batched

	values = append(values, ns.ID, req.AlignBy)

	rows, err := database.DB.Raw(
		fmt.Sprintf("WITH params(run_uuid, key, context_id, steps) AS (VALUES %s)", &valuesStmt)+
			"        SELECT m.run_uuid, "+
			"				rm.key, "+
			"				m.iter, "+
			"				m.value, "+
			"				m.is_nan, "+
			"				rm.context_id, "+
			"				rm.context_json"+
			"		 FROM metrics AS m"+
			"        RIGHT JOIN ("+
			"          SELECT p.run_uuid, "+
			"				  p.key, "+
			"				  p.context_id, "+
			"				  lm.last_iter AS max, "+
			"				  (lm.last_iter + 1) / p.steps AS interval, "+
			"				  contexts.json AS context_json"+
			"          FROM params AS p"+
			"          LEFT JOIN latest_metrics AS lm USING(run_uuid, key, context_id)"+
			"          INNER JOIN contexts ON contexts.id = lm.context_id"+
			"        ) rm USING(run_uuid, context_id)"+
			"		 INNER JOIN runs AS r ON m.run_uuid = r.run_uuid"+
			"		 INNER JOIN experiments AS e ON r.experiment_id = e.experiment_id AND e.namespace_id = ?"+
			"        WHERE m.key = ?"+
			"          AND m.iter <= rm.max"+
			"          AND MOD(m.iter + 1 + rm.interval / 2, rm.interval) < 1"+
			"        ORDER BY r.row_num DESC, rm.key, rm.context_id, m.iter",
		values...,
	).Rows()
	if err != nil {
		return fmt.Errorf("error searching aligned run metrics: %w", err)
	}
	if err := rows.Err(); err != nil {
		return api.NewInternalError("error getting query result: %s", err)
	}

	ctx.Set("Content-Type", "application/octet-stream")
	ctx.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		//nolint:errcheck
		defer rows.Close()

		start := time.Now()
		if err := func() error {
			var id string
			var key string
			var context fiber.Map
			var contextID uint
			metrics := make([]fiber.Map, 0)
			values := make([]float64, 0, capacity)
			iters := make([]float64, 0, capacity)

			addMetrics := func() {
				if key != "" {
					metric := fiber.Map{
						"name":          key,
						"context":       context,
						"x_axis_values": toNumpy(values),
						"x_axis_iters":  toNumpy(iters),
					}
					metrics = append(metrics, metric)
				}
			}

			flushMetrics := func() error {
				if id == "" {
					return nil
				}
				if err := encoding.EncodeTree(w, fiber.Map{
					id: metrics,
				}); err != nil {
					return err
				}
				return w.Flush()
			}

			for rows.Next() {
				var metric struct {
					database.Metric
					Context datatypes.JSON `gorm:"column:context_json"`
				}
				if err := database.DB.ScanRows(rows, &metric); err != nil {
					return err
				}

				// New series of metrics
				if metric.Key != key || metric.RunID != id || metric.ContextID != contextID {
					addMetrics()

					if metric.RunID != id {
						if err := flushMetrics(); err != nil {
							return err
						}
						metrics = metrics[:0]
						id = metric.RunID
					}

					key = metric.Key
					values = values[:0]
					iters = iters[:0]
					context = fiber.Map{}
				}

				v := metric.Value
				if metric.IsNan {
					v = math.NaN()
				}
				values = append(values, v)
				iters = append(iters, float64(metric.Iter))
				if metric.Context != nil {
					// to be properly decoded by AIM UI, json should be represented as a key:value object.
					if err := json.Unmarshal(metric.Context, &context); err != nil {
						return eris.Wrap(err, "error unmarshalling `context` json to `fiber.Map` object")
					}
					contextID = metric.ContextID
				}
			}

			addMetrics()
			if err := flushMetrics(); err != nil {
				return err
			}

			return nil
		}(); err != nil {
			log.Errorf("Error encountered in %s %s: error streaming metrics: %s", ctx.Method(), ctx.Path(), err)
		}

		log.Infof("body - %s %s %s", time.Since(start), ctx.Method(), ctx.Path())
	})

	return nil
}

// DeleteRun will remove the Run from the repo
func (c Controller) DeleteRun(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("deleteRun namespace: %s", ns.Code)

	req := request.DeleteRunRequest{}
	if err = ctx.ParamsParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	// TODO this code should move to service
	runRepository := repositories.NewRunRepository(database.DB)
	run, err := runRepository.GetByNamespaceIDAndRunID(ctx.Context(), ns.ID, req.ID)
	if err != nil {
		return fiber.NewError(
			fiber.StatusInternalServerError, fmt.Sprintf("unable to find run '%s': %s", req.ID, err),
		)
	}
	if run == nil {
		return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("unable to find run '%s'", req.ID))
	}

	// TODO this code should move to service with injected repository
	if err = runRepository.Delete(ctx.Context(), ns.ID, run); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError,
			fmt.Sprintf("unable to delete run %q: %s", req.ID, err),
		)
	}

	return ctx.JSON(fiber.Map{
		"id":     req.ID,
		"status": "OK",
	})
}

// UpdateRun will update the run name, description, and lifecycle stage
func (c Controller) UpdateRun(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("updateRun namespace: %s", ns.Code)

	req := request.UpdateRunRequest{}
	if err = ctx.ParamsParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	if err = ctx.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	// TODO this code should move to service
	runRepository := repositories.NewRunRepository(database.DB)
	run, err := runRepository.GetByNamespaceIDAndRunID(ctx.Context(), ns.ID, req.ID)
	if err != nil {
		return fiber.NewError(
			fiber.StatusInternalServerError, fmt.Sprintf("unable to find run '%s': %s", req.ID, err),
		)
	}
	if run == nil {
		return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("unable to find run '%s'", req.ID))
	}

	if req.Archived != nil {
		if *req.Archived {
			if err := runRepository.Archive(ctx.Context(), run); err != nil {
				return fiber.NewError(fiber.StatusInternalServerError,
					fmt.Sprintf("unable to archive/restore run %q: %s", req.ID, err))
			}
		} else {
			if err := runRepository.Restore(ctx.Context(), run); err != nil {
				return fiber.NewError(fiber.StatusInternalServerError,
					fmt.Sprintf("unable to archive/restore run %q: %s", req.ID, err))
			}
		}
	}

	if req.Name != nil {
		run.Name = *req.Name
		// TODO:DSuhinin - transaction?
		if err := database.DB.Transaction(func(tx *gorm.DB) error {
			if err := runRepository.UpdateWithTransaction(ctx.Context(), tx, run); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError,
				fmt.Sprintf("unable to update run %q: %s", req.ID, err))
		}
	}

	return ctx.JSON(fiber.Map{
		"id":     req.ID,
		"status": "OK",
	})
}

func (c Controller) ArchiveBatch(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("archiveBatch namespace: %s", ns.Code)

	req := request.ArchiveBatchRequest{}
	if err := ctx.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	// TODO this code should move to service
	runRepo := repositories.NewRunRepository(database.DB)
	if ctx.Query("archive") == "true" {
		if err := runRepo.ArchiveBatch(ctx.Context(), ns.ID, req); err != nil {
			return err
		}
	} else {
		if err := runRepo.RestoreBatch(ctx.Context(), ns.ID, req); err != nil {
			return err
		}
	}

	return ctx.JSON(fiber.Map{
		"status": "OK",
	})
}

func (c Controller) DeleteBatch(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("deleteBatch namespace: %s", ns.Code)

	req := request.DeleteBatchRequest{}
	if err := ctx.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	// TODO this code should move to service
	runRepo := repositories.NewRunRepository(database.DB)
	if err := runRepo.DeleteBatch(ctx.Context(), ns.ID, req); err != nil {
		return err
	}
	return ctx.JSON(fiber.Map{
		"status": "OK",
	})
}

func toNumpy(values []float64) fiber.Map {
	buf := bytes.NewBuffer(make([]byte, 0, len(values)*8))
	for _, v := range values {
		switch v {
		case math.MaxFloat64:
			v = math.Inf(1)
		case -math.MaxFloat64:
			v = math.Inf(-1)
		}
		//nolint:gosec,errcheck
		binary.Write(buf, binary.LittleEndian, v)
	}
	return fiber.Map{
		"type":  "numpy",
		"dtype": "float64",
		"shape": len(values),
		"blob":  buf.Bytes(),
	}
}
