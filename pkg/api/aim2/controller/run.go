package controller

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
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

// GetRunMetrics handles `GET /runs/active` endpoint.
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

	runs, err := c.runService.GetRunsActive(ctx.Context(), ns, &req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return ActiveRunsStreamResponse(ctx, runs, req.ReportProgress)
}

// SearchRuns handles `GET /runs/search` endpoint.
func (c Controller) SearchRuns(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("searchRuns namespace: %s", ns.Code)

	tzOffset, err := strconv.Atoi(ctx.Get("x-timezone-offset", "0"))
	if err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "x-timezone-offset header is not a valid integer")
	}

	// Complete the request
	req := request.SearchRunsRequest{}
	if err = ctx.QueryParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}
	if ctx.Query("report_progress") == "" {
		req.ReportProgress = true
	}
	req.TimeZoneOffset = tzOffset
	req.NamespaceID = ns.ID

	// Search runs
	runs, total, err := c.runService.SearchRuns(ctx.Context(), ns, req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	log.Debugf("found %d runs", len(runs))

	// Choose response
	switch req.Action {
	case "export":
		RunsSearchAsCSVResponse(ctx, runs, req.ExcludeTraces, req.ExcludeParams)
	default:
		RunsSearchAsStreamResponse(ctx, runs, total, req.ExcludeTraces, req.ExcludeParams, req.ReportProgress)
	}

	return nil
}

// SearchMetrics handles `GET /runs/search/metric` endpoint.
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
	req.TimeZoneOffset = tzOffset
	req.NamespaceID = ns.ID

	var XAxis bool
	if req.XAxis != "" {
		XAxis = true
	}

	rows, totalRuns, err := c.runService.SearchMetrics(ctx, req)

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
