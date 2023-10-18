package aim

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/aim/encoding"
	"github.com/G-Research/fasttrackml/pkg/api/aim/query"
	"github.com/G-Research/fasttrackml/pkg/api/aim/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/repositories"
	"github.com/G-Research/fasttrackml/pkg/common/middleware/namespace"
	"github.com/G-Research/fasttrackml/pkg/database"
)

func GetRunInfo(c *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(c.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("getRunInfo namespace: %s", ns.Code)

	q := struct {
		// TODO skip_system is unused - should we keep it?
		SkipSystem bool     `query:"skip_system"`
		Sequences  []string `query:"sequence"`
	}{}

	if err := c.QueryParser(&q); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	p := struct {
		ID string `params:"id"`
	}{}

	if err := c.ParamsParser(&p); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	tx := database.DB.
		InnerJoins(
			"Experiment",
			database.DB.Where(&models.Experiment{NamespaceID: ns.ID}),
		).
		Preload("Params").
		Preload("Tags")

	if len(q.Sequences) == 0 {
		q.Sequences = []string{
			"audios",
			"distributions",
			"figures",
			"images",
			"log_records",
			"logs",
			"metric",
			"texts",
		}
	}

	traces := make(map[string][]fiber.Map, len(q.Sequences))
	for _, s := range q.Sequences {
		switch s {
		case "audios", "distributions", "figures", "images", "log_records", "logs", "texts":
			traces[s] = []fiber.Map{}
		case "metric":
			tx.Preload("LatestMetrics", func(db *gorm.DB) *gorm.DB {
				return db.Select("RunID", "Key")
			})
		default:
			return fmt.Errorf("%q is not a valid Sequence", s)
		}
	}

	r := database.Run{
		ID: p.ID,
	}

	if err := tx.First(&r).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.ErrNotFound
		}
		return fmt.Errorf("error retrieving run %q: %w", p.ID, err)
	}

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
	params := make(map[string]any, len(r.Params)+1)
	for _, p := range r.Params {
		params[p.Key] = p.Value
	}
	tags := make(map[string]string, len(r.Tags))
	for _, t := range r.Tags {
		tags[t.Key] = t.Value
	}
	params["tags"] = tags

	metrics := make([]fiber.Map, len(r.LatestMetrics))
	for i, m := range r.LatestMetrics {
		metrics[i] = fiber.Map{
			"name":       m.Key,
			"last_value": 0.1,
			"context":    fiber.Map{},
		}
	}
	traces["metric"] = metrics

	return c.JSON(fiber.Map{
		"params": params,
		"traces": traces,
		"props":  props,
	})
}

func GetRunMetrics(c *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(c.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("getRunMetrics namespace: %s", ns.Code)

	p := struct {
		ID string `params:"id"`
	}{}

	if err := c.ParamsParser(&p); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	b := []struct {
		Context fiber.Map `json:"context"`
		Name    string    `json:"name"`
	}{}

	if err := c.BodyParser(&b); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	metricKeysMap := make(fiber.Map, len(b))
	for _, m := range b {
		metricKeysMap[m.Name] = nil
	}
	metricKeys := make([]string, len(metricKeysMap))

	i := 0
	for k := range metricKeysMap {
		metricKeys[i] = k
		i++
	}

	r := database.Run{
		ID: p.ID,
	}
	if err := database.DB.
		Select("ID").
		InnerJoins(
			"Experiment",
			database.DB.Where(&models.Experiment{NamespaceID: ns.ID}),
		).
		Preload("Metrics", func(db *gorm.DB) *gorm.DB {
			return db.
				Where("key IN ?", metricKeys).
				Order("iter")
		}).
		First(&r).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.ErrNotFound
		}
		return fmt.Errorf("unable to find run %q: %w", p.ID, err)
	}

	metrics := make(map[string]struct {
		values []*float64
		iters  []int
	}, len(metricKeys))
	for _, m := range r.Metrics {
		k := metrics[m.Key]

		v := m.Value
		pv := &v
		if m.IsNan {
			pv = nil
		}

		k.values = append(k.values, pv)
		k.iters = append(k.iters, int(m.Iter))
		metrics[m.Key] = k
	}

	resp := make([]fiber.Map, len(metrics))
	for i, k := range metricKeys {
		resp[i] = fiber.Map{
			"name":    k,
			"context": fiber.Map{},
			"values":  metrics[k].values,
			"iters":   metrics[k].iters,
		}
	}

	return c.JSON(resp)
}

func GetRunsActive(c *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(c.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("getRunsActive namespace: %s", ns.Code)

	q := struct {
		ReportProgress bool `query:"report_progress"`
	}{}

	if err := c.QueryParser(&q); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	if c.Query("report_progress") == "" {
		q.ReportProgress = true
	}

	var runs []database.Run
	if tx := database.DB.
		Where("status = ?", database.StatusRunning).
		InnerJoins(
			"Experiment",
			database.DB.Where(&models.Experiment{NamespaceID: ns.ID}),
		).
		Preload("LatestMetrics").
		Find(&runs); tx.Error != nil {
		return fmt.Errorf("error retrieving active runs: %w", tx.Error)
	}

	c.Set("Content-Type", "application/octet-stream")
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
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
					metrics[i] = fiber.Map{
						"context": fiber.Map{},
						"name":    m.Key,
						"last_value": fiber.Map{
							"dtype":      "float",
							"first_step": 0,
							"last_step":  m.LastIter,
							"last":       v,
							"version":    2,
						},
					}
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

				if q.ReportProgress {
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
			log.Errorf("Error encountered in %s %s: error streaming active runs: %s", c.Method(), c.Path(), err)
		}

		log.Infof("body - %s %s %s", time.Since(start), c.Method(), c.Path())
	})

	return nil
}

// TODO:get back and fix `gocyclo` problem.
//
//nolint:gocyclo
func SearchRuns(c *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(c.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("searchRuns namespace: %s", ns.Code)

	q := struct {
		Query  string `query:"q"`
		Limit  int    `query:"limit"`
		Offset string `query:"offset"`
		// TODO skip_system is unused - should we keep it?
		SkipSystem     bool `query:"skip_system"`
		ReportProgress bool `query:"report_progress"`
		ExcludeParams  bool `query:"exclude_params"`
		ExcludeTraces  bool `query:"exclude_traces"`
	}{}

	if err = c.QueryParser(&q); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	if c.Query("report_progress") == "" {
		q.ReportProgress = true
	}

	tzOffset, err := strconv.Atoi(c.Get("x-timezone-offset", "0"))
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
	pq, err := qp.Parse(q.Query)
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
			database.DB.Where(&models.Experiment{NamespaceID: ns.ID}),
		).
		Order("row_num DESC")

	if q.Limit > 0 {
		tx.Limit(q.Limit)
	}

	if q.Offset != "" {
		run := &database.Run{
			ID: q.Offset,
		}
		// TODO:DSuhinin -> do we need `namespace` restriction? it seems like yyyyess, but ....
		if tx = database.DB.Select("row_num").First(&run); tx.Error != nil && tx.Error != gorm.ErrRecordNotFound {
			return fmt.Errorf("unable to find search runs offset %q: %w", q.Offset, tx.Error)
		}

		tx.Where("row_num < ?", run.RowNum)
	}

	if !q.ExcludeParams {
		tx.Preload("Params")
		tx.Preload("Tags")
	}

	if !q.ExcludeTraces {
		tx.Preload("LatestMetrics")
	}

	var runs []database.Run
	pq.Filter(tx).Find(&runs)
	if tx.Error != nil {
		return fmt.Errorf("error searching runs: %w", tx.Error)
	}

	log.Debugf("Found %d runs", len(runs))

	c.Set("Content-Type", "application/octet-stream")
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		start := time.Now()
		if err = func() error {
			for i, r := range runs {
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

				if !q.ExcludeTraces {
					metrics := make([]fiber.Map, len(r.LatestMetrics))
					for i, m := range r.LatestMetrics {
						v := m.Value
						if m.IsNan {
							v = math.NaN()
						}
						metrics[i] = fiber.Map{
							"context": fiber.Map{},
							"name":    m.Key,
							"last_value": fiber.Map{
								"dtype":      "float",
								"first_step": 0,
								"last_step":  m.LastIter,
								"last":       v,
								"version":    2,
							},
						}
					}
					run["traces"] = fiber.Map{
						"metric": metrics,
					}
				}

				if !q.ExcludeParams {
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
				}

				err = encoding.EncodeTree(w, fiber.Map{
					r.ID: run,
				})
				if err != nil {
					return err
				}

				if q.ReportProgress {
					err = encoding.EncodeTree(w, fiber.Map{
						fmt.Sprintf("progress_%d", i): []int64{total - int64(r.RowNum), total},
					})
					if err != nil {
						return err
					}
				}

				err = w.Flush()
				if err != nil {
					return err
				}
			}

			if q.ReportProgress && err == nil {
				err = encoding.EncodeTree(w, fiber.Map{
					fmt.Sprintf("progress_%d", len(runs)): []int64{total, total},
				})
				if err != nil {
					err = w.Flush()
				}
			}

			return nil
		}(); err != nil {
			log.Errorf("Error encountered in %s %s: error streaming runs: %s", c.Method(), c.Path(), err)
		}

		log.Infof("body - %s %s %s", time.Since(start), c.Method(), c.Path())
	})

	return nil
}

// TODO:get back and fix `gocyclo` problem.
//
//nolint:gocyclo
func SearchMetrics(c *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(c.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("searchMetrics namespace: %s", ns.Code)

	q := struct {
		Query string `query:"q"`
		Steps int    `query:"p"`
		XAxis string `query:"x_axis"`
		// TODO skip_system is unused - should we keep it?
		SkipSystem     bool `query:"skip_system"`
		ReportProgress bool `query:"report_progress"`
	}{}

	if err = c.QueryParser(&q); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	if c.Query("report_progress") == "" {
		q.ReportProgress = true
	}

	if c.Query("p") == "" {
		q.Steps = 50
	}

	tzOffset, err := strconv.Atoi(c.Get("x-timezone-offset", "0"))
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
	pq, err := qp.Parse(q.Query)
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
			database.DB.Where(&models.Experiment{NamespaceID: ns.ID}),
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
			Joins("LEFT JOIN latest_metrics USING(run_uuid)"))).
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
		Select("metrics.*").
		Table("metrics").
		Joins(
			"INNER JOIN (?) runmetrics USING(run_uuid, key)",
			pq.Filter(database.DB.
				Select(
					"runs.run_uuid",
					"runs.row_num",
					"latest_metrics.key",
					fmt.Sprintf("(latest_metrics.last_iter + 1)/ %f AS interval", float32(q.Steps)),
				).
				Table("runs").
				Joins(
					"INNER JOIN experiments ON experiments.experiment_id = runs.experiment_id AND experiments.namespace_id = ?",
					ns.ID,
				).
				Joins("LEFT JOIN latest_metrics USING(run_uuid)")),
		).
		Where("MOD(metrics.iter + 1 + runmetrics.interval / 2, runmetrics.interval) < 1").
		Order("runmetrics.row_num DESC").
		Order("metrics.key").
		Order("metrics.iter")

	var xAxis bool
	if q.XAxis != "" {
		tx.
			Select("metrics.*", "x_axis.value as x_axis_value", "x_axis.is_nan as x_axis_is_nan").
			Joins(
				"LEFT JOIN metrics x_axis ON metrics.run_uuid = x_axis.run_uuid AND metrics.iter = x_axis.iter AND x_axis.key = ?",
				q.XAxis,
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

	c.Set("Content-Type", "application/octet-stream")
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		//nolint:errcheck
		defer rows.Close()

		start := time.Now()
		if err := func() error {
			var (
				id          string
				key         string
				metrics     []fiber.Map
				values      []float64
				iters       []float64
				epochs      []float64
				timestamps  []float64
				xAxisValues []float64
				progress    int
			)
			reportProgress := func(cur int64) error {
				if !q.ReportProgress {
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
						"context":       fiber.Map{},
						"slice":         []int{0, 0, q.Steps},
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
					XAxisValue float64 `gorm:"column:x_axis_value"`
					XAxisIsNaN bool    `gorm:"column:x_axis_is_nan"`
				}
				if err := database.DB.ScanRows(rows, &metric); err != nil {
					return err
				}

				// New series of metrics
				if metric.Key != key || metric.RunID != id {
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

					values = make([]float64, 0, q.Steps)
					iters = make([]float64, 0, q.Steps)
					epochs = make([]float64, 0, q.Steps)
					timestamps = make([]float64, 0, q.Steps)
					if xAxis {
						xAxisValues = make([]float64, 0, q.Steps)
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
			log.Errorf("Error encountered in %s %s: error streaming metrics: %s", c.Method(), c.Path(), err)
		}

		log.Infof("body - %s %s %s", time.Since(start), c.Method(), c.Path())
	})

	return nil
}

// TODO:get back and fix `gocyclo` problem.
//
//nolint:gocyclo
func SearchAlignedMetrics(c *fiber.Ctx) error {
	b := struct {
		AlignBy string `json:"align_by"`
		Runs    []struct {
			ID     string `json:"run_id"`
			Traces []struct {
				Context fiber.Map `json:"context"`
				Name    string    `json:"name"`
				Slice   [3]int    `json:"slice"`
			} `json:"traces"`
		} `json:"runs"`
	}{}

	if err := c.BodyParser(&b); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	var capacity int
	var values []any
	for _, r := range b.Runs {
		for _, t := range r.Traces {
			l := t.Slice[2]
			if l > capacity {
				capacity = l
			}
			values = append(values, r.ID, t.Name, float32(l))
		}
	}

	var valuesStmt strings.Builder
	length := len(values) / 3
	for i := 0; i < length; i++ {
		valuesStmt.WriteString("(?, ?, CAST(? AS numeric))")
		if i < length-1 {
			valuesStmt.WriteString(",")
		}
	}

	// TODO this should probably be batched

	values = append(values, b.AlignBy)
	rows, err := database.DB.Raw(
		fmt.Sprintf("WITH params(run_uuid, key, steps) AS (VALUES %s)", &valuesStmt)+
			"        SELECT m.run_uuid, rm.key, m.iter, m.value, m.is_nan FROM metrics AS m"+
			"        RIGHT JOIN ("+
			"          SELECT p.run_uuid, p.key, lm.last_iter AS max, (lm.last_iter + 1) / p.steps AS interval"+
			"          FROM params AS p"+
			"          LEFT JOIN latest_metrics AS lm USING(run_uuid, key)"+
			"        ) rm USING(run_uuid)"+
			"        WHERE m.key = ?"+
			"          AND m.iter <= rm.max"+
			"          AND MOD(m.iter + 1 + rm.interval / 2, rm.interval) < 1"+
			"        ORDER BY m.run_uuid, rm.key, m.iter",
		values...,
	).Rows()
	if err != nil {
		return fmt.Errorf("error searching aligned run metrics: %w", err)
	}
	if err := rows.Err(); err != nil {
		return api.NewInternalError("error getting query result: %s", err)
	}

	c.Set("Content-Type", "application/octet-stream")
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		//nolint:errcheck
		defer rows.Close()

		start := time.Now()
		if err := func() error {
			var id string
			var key string
			metrics := make([]fiber.Map, 0)
			values := make([]float64, 0, capacity)
			iters := make([]float64, 0, capacity)

			addMetrics := func() {
				if key != "" {
					metric := fiber.Map{
						"name":          key,
						"context":       fiber.Map{},
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
				var metric database.Metric
				if err := database.DB.ScanRows(rows, &metric); err != nil {
					return err
				}

				// New series of metrics
				if metric.Key != key || metric.RunID != id {
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
				}

				v := metric.Value
				if metric.IsNan {
					v = math.NaN()
				}
				values = append(values, v)
				iters = append(iters, float64(metric.Iter))
			}

			addMetrics()
			if err := flushMetrics(); err != nil {
				return err
			}

			return nil
		}(); err != nil {
			log.Errorf("Error encountered in %s %s: error streaming metrics: %s", c.Method(), c.Path(), err)
		}

		log.Infof("body - %s %s %s", time.Since(start), c.Method(), c.Path())
	})

	return nil
}

// DeleteRun will remove the Run from the repo
func DeleteRun(c *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(c.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("deleteRun namespace: %s", ns.Code)

	params := struct {
		ID string `params:"id"`
	}{}

	if err = c.ParamsParser(&params); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	// TODO this code should move to service
	runRepository := repositories.NewRunRepository(database.DB)
	run, err := runRepository.GetByNamespaceIDAndRunID(c.Context(), ns.ID, params.ID)
	if err != nil {
		return fiber.NewError(
			fiber.StatusInternalServerError, fmt.Sprintf("unable to find run '%s': %s", params.ID, err),
		)
	}
	if run == nil {
		return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("unable to find run '%s'", params.ID))
	}

	// TODO this code should move to service with injected repository
	if err = runRepository.Delete(c.Context(), run); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError,
			fmt.Sprintf("unable to delete run %q: %s", params.ID, err),
		)
	}

	return c.JSON(fiber.Map{
		"id":     params.ID,
		"status": "OK",
	})
}

// UpdateRun will update the run name, description, and lifecycle stage
func UpdateRun(c *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(c.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("updateRun namespace: %s", ns.Code)

	params := struct {
		ID string `params:"id"`
	}{}
	if err = c.ParamsParser(&params); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	var updateRequest request.UpdateRunRequest
	if err = c.BodyParser(&updateRequest); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	// TODO this code should move to service
	runRepository := repositories.NewRunRepository(database.DB)
	run, err := runRepository.GetByNamespaceIDAndRunID(c.Context(), ns.ID, params.ID)
	if err != nil {
		return fiber.NewError(
			fiber.StatusInternalServerError, fmt.Sprintf("unable to find run '%s': %s", params.ID, err),
		)
	}
	if run == nil {
		return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("unable to find run '%s'", params.ID))
	}

	if updateRequest.Archived != nil {
		if *updateRequest.Archived {
			if err := runRepository.Archive(c.Context(), run); err != nil {
				return fiber.NewError(fiber.StatusInternalServerError,
					fmt.Sprintf("unable to archive/restore run %q: %s", params.ID, err))
			}
		} else {
			if err := runRepository.Restore(c.Context(), run); err != nil {
				return fiber.NewError(fiber.StatusInternalServerError,
					fmt.Sprintf("unable to archive/restore run %q: %s", params.ID, err))
			}
		}
	}

	if updateRequest.Name != nil {
		run.Name = *updateRequest.Name
		// TODO:DSuhinin - transaction?
		if err := database.DB.Transaction(func(tx *gorm.DB) error {
			if err := runRepository.UpdateWithTransaction(c.Context(), tx, run); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError,
				fmt.Sprintf("unable to update run %q: %s", params.ID, err))
		}
	}

	return c.JSON(fiber.Map{
		"id":     params.ID,
		"status": "OK",
	})
}

func ArchiveBatch(c *fiber.Ctx) error {
	var ids []string
	if err := c.BodyParser(&ids); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	// TODO this code should move to service
	runRepo := repositories.NewRunRepository(database.DB)
	if c.Query("archive") == "true" {
		if err := runRepo.ArchiveBatch(c.Context(), ids); err != nil {
			return err
		}
	} else {
		if err := runRepo.RestoreBatch(c.Context(), ids); err != nil {
			return err
		}
	}

	return c.JSON(fiber.Map{
		"status": "OK",
	})
}

func DeleteBatch(c *fiber.Ctx) error {
	var ids []string
	if err := c.BodyParser(&ids); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	// TODO this code should move to service
	runRepo := repositories.NewRunRepository(database.DB)
	if err := runRepo.DeleteBatch(c.Context(), ids); err != nil {
		return err
	}
	return c.JSON(fiber.Map{
		"status": "OK",
	})
}

func toNumpy(values []float64) fiber.Map {
	buf := bytes.NewBuffer(make([]byte, 0, len(values)*8))
	for _, v := range values {
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
