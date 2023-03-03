package aim

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"strconv"

	"github.com/G-Resarch/fasttrack/api/aim/encoding"
	"github.com/G-Resarch/fasttrack/api/aim/query"
	"github.com/G-Resarch/fasttrack/database"

	"github.com/go-python/gpython/parser"
	"github.com/go-python/gpython/py"
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GetRunInfo(c *fiber.Ctx) error {
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
		Joins("Experiment", database.DB.Select("ID", "Name")).
		Preload("Params")

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

	tx.First(&r)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return fiber.ErrNotFound
		}
		return fmt.Errorf("error retrieving run %q: %w", p.ID, tx.Error)
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
	params := make(map[string]string, len(r.Params))
	for _, p := range r.Params {
		params[p.Key] = p.Value
	}

	if len(r.LatestMetrics) != 0 {
		metrics := make([]fiber.Map, len(r.LatestMetrics))
		for i, m := range r.LatestMetrics {
			metrics[i] = fiber.Map{
				"name":       m.Key,
				"last_value": 0.1,
				"context":    fiber.Map{},
			}
		}
		traces["metric"] = metrics
	}

	return c.JSON(fiber.Map{
		"params": params,
		"traces": traces,
		"props":  props,
	})
}

func GetRunMetricBatch(c *fiber.Ctx) error {
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
	if tx := database.DB.
		Select("ID").
		Preload("Metrics", func(db *gorm.DB) *gorm.DB {
			return db.
				Where("key IN ?", metricKeys).
				Order("key").
				Order("step").
				Order("timestamp")
		}).
		First(&r); tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return fiber.ErrNotFound
		}
		return fmt.Errorf("unable to find run %q: %w", p.ID, tx.Error)
	}

	metrics := make(map[string]struct {
		values []float64
		iters  []int
	}, len(metricKeys))
	for _, m := range r.Metrics {
		var i int
		k := metrics[m.Key]
		if len(k.iters) != 0 {
			i = k.iters[len(k.iters)-1] + 1
		}

		v := m.Value
		if m.IsNan {
			v = math.NaN()
		}

		k.values = append(k.values, v)
		k.iters = append(k.iters, i)
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
		Joins("Experiment", database.DB.Select("ID", "Name")).
		Preload("LatestMetrics").
		Find(&runs); tx.Error != nil {
		return fmt.Errorf("error retrieving active runs: %w", tx.Error)
	}

	c.Set("Content-Type", "application/octet-stream")
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		var err error
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
						"dtype": "float",
						// TODO would need to add this to LatestMetrics
						"first_step": 0,
						"last_step":  m.Step,
						"last":       v,
						"version":    2,
					},
				}
			}

			err = encoding.EncodeTree(w, fiber.Map{
				r.ID: fiber.Map{
					"props": props,
					"traces": fiber.Map{
						"metric": metrics,
					},
				},
			})
			if err != nil {
				break
			}

			if q.ReportProgress {
				err = encoding.EncodeTree(w, fiber.Map{
					fmt.Sprintf("progress_%d", i): []int{i + 1, len(runs)},
				})
				if err != nil {
					break
				}
			}

			err = w.Flush()
			if err != nil {
				break
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

		if err != nil {
			log.Errorf("Error encountered in %s %s: error streaming active runs: %s", c.Method(), c.Path(), err)
		}
	})

	return nil
}

func GetRunsSearch(c *fiber.Ctx) error {
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

	if err := c.QueryParser(&q); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	if c.Query("report_progress") == "" {
		q.ReportProgress = true
	}

	tzOffset, err := strconv.Atoi(c.Get("x-timezone-offset", "0"))
	if err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "x-timezone-offset header is not a valid integer")
	}

	pq := query.QueryParser{
		Tables: map[string]query.Table{
			"run": {
				"created_at": clause.Column{
					// Table: "runs",
					Name: "start_time",
				},
				"finalized_at": clause.Column{
					// Table: "runs",
					Name: "end_time",
				},
				"hash": clause.Column{
					// Table: "runs",
					Name: "run_uuid",
				},
				"name": clause.Column{
					// Table: "runs",
					Name: "name",
				},
				"experiment": clause.Column{
					Table: "Experiment",
					Name:  "name",
				},
				"archived": clause.Eq{
					Column: clause.Column{
						// Table: "runs",
						Name: "lifecycle_stage",
					},
					Value: database.LifecycleStageDeleted,
				},
				"active": clause.Eq{
					Column: clause.Column{
						Table: "runs",
						Name:  "status",
					},
					Value: database.StatusRunning,
				},
				"duration": clause.Column{
					Name: "runs.end_time - runs.start_time",
					Raw:  true,
				},
				// "tags":
				// "metrics":
			},
		},
		TzOffset: tzOffset,
	}
	qp, err := pq.Parse(q.Query)
	if err != nil {
		// TODO should have a type for syntax errors
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("invalid query: %s", err))
	}

	tx := database.DB.
		Model(&database.Run{}).
		Joins("Experiment")

	var total int64
	qp.Filter(tx).Count(&total)
	if tx.Error != nil {
		return fmt.Errorf("unable to count total runs: %w", tx.Error)
	}

	log.Debugf("Total runs: %d", total)

	tx = database.DB.
		Joins("Experiment", database.DB.Select("ID", "Name")).
		Order("start_time DESC").
		Order("run_uuid")

	if q.Limit > 0 {
		tx.Limit(q.Limit)
	}

	var offset int64
	if q.Offset != "" {
		r := &database.Run{
			ID: q.Offset,
		}
		if tx := database.DB.Select("start_time").First(&r); tx.Error != nil && tx.Error != gorm.ErrRecordNotFound {
			return fmt.Errorf("unable to find search runs offset %q: %w", q.Offset, tx.Error)
		}
		var count int64

		expr := clause.Or(
			clause.Gt{
				Column: "start_time",
				Value:  r.StartTime,
			},
			clause.And(
				clause.Eq{
					Column: "start_time",
					Value:  r.StartTime,
				},
				clause.Lt{
					Column: "run_uuid",
					Value:  r.ID,
				},
			),
		)

		// TODO how do we deal with this???
		// if where != nil {
		// 	expr = clause.And(where, expr)
		// }
		if tx := qp.Filter(
			database.DB.
				Model(&database.Run{}).
				Joins("Experiment").
				Where(expr)).
			Count(&count); tx.Error != nil {
			return fmt.Errorf("unable to compute search runs offset %q: %w", q.Offset, tx.Error)
		}
		offset = count + 1

		log.Debugf("Runs offset: %d", offset)
	}
	tx.Offset(int(offset))

	if !q.ExcludeParams {
		tx.Preload("Params")
	}

	if !q.ExcludeTraces {
		tx.Preload("LatestMetrics")
	}

	var runs []database.Run
	qp.Filter(tx).Find(&runs)
	if tx.Error != nil {
		return fmt.Errorf("error searching runs: %w", tx.Error)
	}

	log.Debugf("Found %d runs", len(runs))

	c.Set("Content-Type", "application/octet-stream")
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		var err error
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
							"dtype": "float",
							// TODO would need to add this to LatestMetrics
							"first_step": 0,
							"last_step":  m.Step,
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
				params := make(fiber.Map, len(r.Params))
				for _, p := range r.Params {
					params[p.Key] = p.Value
				}
				run["params"] = params
			}

			err = encoding.EncodeTree(w, fiber.Map{
				r.ID: run,
			})
			if err != nil {
				break
			}

			if q.ReportProgress {
				err = encoding.EncodeTree(w, fiber.Map{
					fmt.Sprintf("progress_%d", i): []int64{offset + int64(i) + 1, total},
				})
				if err != nil {
					break
				}
			}

			err = w.Flush()
			if err != nil {
				break
			}
		}

		if q.ReportProgress && err == nil {
			err = encoding.EncodeTree(w, fiber.Map{
				fmt.Sprintf("progress_%d", len(runs)): []int64{offset + int64(len(runs)), total},
			})
			if err != nil {
				err = w.Flush()
			}
		}

		if err != nil {
			log.Errorf("Error encountered in %s %s: error streaming active runs: %s", c.Method(), c.Path(), err)
		}
	})

	return nil
}

// TODO proper streaming (à la api.mlflow.GetMetricHistories)
// though maybe not necessary once Query and Steps are implemented
func GetRunsMetricsSearch(c *fiber.Ctx) error {
	q := struct {
		Query string `query:"q"`
		Steps int    `query:"p"`
		XAxis string `query:"x_axis"`
		// TODO skip_system is unused - should we keep it?
		SkipSystem     bool `query:"skip_system"`
		ReportProgress bool `query:"report_progress"`
	}{}

	if err := c.QueryParser(&q); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	if c.Query("report_progress") == "" {
		q.ReportProgress = true
	}

	if c.Query("p") == "" {
		q.Steps = 50
	}

	if q.Query != "" {
		query := fmt.Sprintf("((%s))", q.Query)
		if _, err := parser.ParseString(query, py.EvalMode); err != nil {
			if err, ok := err.(*py.Exception); ok && err.Base.Name == py.SyntaxError.Name {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"message": "SyntaxError",
					"detail": fiber.Map{
						"statement":  query,
						"line":       err.Dict["lineno"],
						"offset":     err.Dict["offset"],
						"end_offset": 0,
					},
				})
			}
			return fmt.Errorf("error parsing run search metrics query %q: %w", q.Query, err)
		}
	}

	// x-timezone-offset seems to be ignored in streams
	// tzOffset, err := strconv.Atoi(c.Get("x-timezone-offset", "0"))
	// if err != nil {
	// 	return fiber.NewError(fiber.StatusUnprocessableEntity, "x-timezone-offset header is not a valid integer")
	// }

	// TODO do something with the query!

	// TODO use q.Steps as a limit

	// TODO use q.XAxis -- the tricky bit may be hashing the steps properly

	var runs []database.Run
	if tx := database.DB.
		Joins("Experiment", database.DB.Select("ID", "Name")).
		Preload("Params").
		Preload("Metrics", func(db *gorm.DB) *gorm.DB {
			return db.
				Order("Key").
				Order("Step").
				Order("Timestamp").
				Order("Value").
				// TODO this is temporary
				Where("Key = ?", "metric1")
		}).
		Order("start_time DESC").
		Order("run_uuid").
		// TODO this is temporary
		// Limit(1).
		Find(&runs); tx.Error != nil {
		return fmt.Errorf("error searching run metrics: %w", tx.Error)
	}

	c.Set("Content-Type", "application/octet-stream")
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		var err error
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

			metrics := map[string]struct {
				values     []float64
				iters      []float64
				epochs     []float64
				timestamps []float64
			}{}
			for _, m := range r.Metrics {
				var i int
				metric, ok := metrics[m.Key]
				if ok {
					i = len(metric.iters)
				}

				v := m.Value
				if m.IsNan {
					v = math.NaN()
				}

				metric.values = append(metric.values, v)
				metric.iters = append(metric.iters, float64(i))
				metric.epochs = append(metric.epochs, float64(m.Step))
				metric.timestamps = append(metric.timestamps, float64(m.Timestamp)/1000)
				metrics[m.Key] = metric
			}

			fm := make([]fiber.Map, len(metrics))
			var n int
			for k, m := range metrics {
				fm[n] = fiber.Map{
					"name":          k,
					"context":       fiber.Map{},
					"slice":         []int{0, 0, q.Steps},
					"values":        toNumpy(m.values),
					"iters":         toNumpy(m.iters),
					"epochs":        toNumpy(m.epochs),
					"timestamps":    toNumpy(m.timestamps),
					"x_axis_values": nil,
					"x_axis_iters":  nil,
				}
				n++
			}
			run["traces"] = fm

			params := make(fiber.Map, len(r.Params))
			for _, p := range r.Params {
				params[p.Key] = p.Value
			}
			run["params"] = params

			err = encoding.EncodeTree(w, fiber.Map{
				r.ID: run,
			})
			if err != nil {
				break
			}

			if q.ReportProgress {
				err = encoding.EncodeTree(w, fiber.Map{
					fmt.Sprintf("progress_%d", i): []int{i + 1, len(runs)},
				})
				if err != nil {
					break
				}
			}

			err = w.Flush()
			if err != nil {
				break
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

		if err != nil {
			log.Errorf("Error encountered in %s %s: error streaming active runs: %s", c.Method(), c.Path(), err)
		}
	})

	return nil
}

func toNumpy(values []float64) fiber.Map {
	buf := bytes.NewBuffer(make([]byte, 0, len(values)*8))
	for _, v := range values {
		binary.Write(buf, binary.LittleEndian, v)
	}
	return fiber.Map{
		"type":  "numpy",
		"dtype": "float64",
		"shape": len(values),
		"blob":  buf.Bytes(),
	}
}
