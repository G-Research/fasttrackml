package response

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/dto"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/encoding"
	"github.com/G-Research/fasttrackml/pkg/common"
)

// GetRunInfoTracesMetricPartial is a partial response object for GetRunInfoTracesPartial.
type GetRunInfoTracesMetricPartial struct {
	Name      string          `json:"name"`
	Context   json.RawMessage `json:"context"`
	LastValue float64         `json:"last_value"`
}

// GetRunInfoParamsPartial is a partial response object for GetRunInfoResponse.
type GetRunInfoParamsPartial map[string]any

// GetRunInfoTracesPartial is a partial response object for GetRunInfoResponse.
type GetRunInfoTracesPartial struct {
	Tags          map[string]string               `json:"tags"`
	Logs          map[string]string               `json:"logs"`
	Texts         map[string]string               `json:"texts"`
	Audios        map[string]string               `json:"audios"`
	Metric        []GetRunInfoTracesMetricPartial `json:"metric"`
	Images        map[string]string               `json:"images"`
	Figures       map[string]string               `json:"figures"`
	LogRecords    map[string]string               `json:"log_records"`
	Distributions map[string]string               `json:"distributions"`
}

// GetRunInfoExperimentPartial experiment properties
type GetRunInfoExperimentPartial struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GetRunInfoPropsPartial is a partial response object for GetRunInfoResponse.
type GetRunInfoPropsPartial struct {
	Name         string                      `json:"name"`
	Description  string                      `json:"description"`
	Experiment   GetRunInfoExperimentPartial `json:"experiment"`
	Tags         []string                    `json:"tags"`
	CreationTime float64                     `json:"creation_time"`
	EndTime      float64                     `json:"end_time"`
	Archived     bool                        `json:"archived"`
	Active       bool                        `json:"active"`
}

// GetRunInfoResponse represents the response struct for GetRunInfoResponse endpoint
type GetRunInfoResponse struct {
	Params GetRunInfoParamsPartial `json:"params"`
	Traces GetRunInfoTracesPartial `json:"traces"`
	Props  GetRunInfoPropsPartial  `json:"props"`
}

// NewGetRunInfoResponse creates new response object for `GER runs/:id/info` endpoint.
func NewGetRunInfoResponse(run *models.Run) *GetRunInfoResponse {
	metrics := make([]GetRunInfoTracesMetricPartial, len(run.LatestMetrics))
	for i, metric := range run.LatestMetrics {
		metrics[i] = GetRunInfoTracesMetricPartial{
			Name:      metric.Key,
			Context:   json.RawMessage(metric.Context.Json),
			LastValue: 0.1,
		}
	}

	params := make(GetRunInfoParamsPartial, len(run.Params)+1)
	for _, p := range run.Params {
		params[p.Key] = p.Value
	}
	tags := make(GetRunInfoParamsPartial, len(run.Tags))
	for _, t := range run.Tags {
		tags[t.Key] = t.Value
	}
	params["tags"] = tags

	return &GetRunInfoResponse{
		Params: params,
		Traces: GetRunInfoTracesPartial{
			Tags:          map[string]string{},
			Logs:          map[string]string{},
			Texts:         map[string]string{},
			Audios:        map[string]string{},
			Metric:        metrics,
			Images:        map[string]string{},
			Figures:       map[string]string{},
			LogRecords:    map[string]string{},
			Distributions: map[string]string{},
		},
		Props: GetRunInfoPropsPartial{
			Name: run.Name,
			Experiment: GetRunInfoExperimentPartial{
				ID:   fmt.Sprintf("%d", *run.Experiment.ID),
				Name: run.Experiment.Name,
			},
			Tags:         []string{},
			CreationTime: float64(run.StartTime.Int64) / 1000,
			EndTime:      float64(run.EndTime.Int64) / 1000,
			Archived:     run.LifecycleStage == models.LifecycleStageDeleted,
			Active:       run.Status == models.StatusRunning,
		},
	}
}

// GetRunMetricsResponse is a response object to hold response data for `GET /runs/:id/metric/get-batch` endpoint.
type GetRunMetricsResponse struct {
	Name    string          `json:"name"`
	Iters   []int           `json:"iters"`
	Values  []*float64      `json:"values"`
	Context json.RawMessage `json:"context"`
}

// NewGetRunMetricsResponse creates new response object for `GET /runs/:id/metric/get-batch` endpoint.
func NewGetRunMetricsResponse(metrics []models.Metric, metricKeysMap dto.MetricKeysMapDTO) []GetRunMetricsResponse {
	data := make(map[dto.MetricKeysItemDTO]struct {
		iters  []int
		values []*float64
	}, len(metricKeysMap))

	for _, item := range metrics {
		v := common.GetPointer(item.Value)
		if item.IsNan {
			v = nil
		}
		key := dto.MetricKeysItemDTO{
			Name:    item.Key,
			Context: string(item.Context.Json),
		}
		m := data[key]
		m.iters = append(m.iters, int(item.Iter))
		m.values = append(m.values, v)
		data[key] = m
	}

	resp := make([]GetRunMetricsResponse, 0, len(metrics))
	for key, m := range data {
		resp = append(resp, GetRunMetricsResponse{
			Name:    key.Name,
			Iters:   m.iters,
			Values:  m.values,
			Context: json.RawMessage(key.Context),
		})
	}
	return resp
}

// SearchAlignedMetricsResponse  is a response object to hold response data for
// `GET /runs/search/metric/align` endpoint.
type SearchAlignedMetricsResponse struct {
	Name        string    `json:"name"`
	Context     fiber.Map `json:"context"`
	XAxisValues fiber.Map `json:"x_axis_values"`
	XAxisIters  fiber.Map `json:"x_axis_iters"`
}

// NewSearchAlignedMetricsResponse creates new response object for `GET /runs/search/metric/align` endpoint.
func NewSearchAlignedMetricsResponse(
	ctx *fiber.Ctx, rows *sql.Rows, next func(*sql.Rows) (*models.AlignedMetric, error), capacity int,
) {
	ctx.Set("Content-Type", "application/octet-stream")
	ctx.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		//nolint:errcheck
		defer rows.Close()

		flushMetrics := func(id string, metrics []SearchAlignedMetricsResponse) error {
			if len(metrics) == 0 {
				return nil
			}
			if err := encoding.EncodeTree(w, fiber.Map{
				id: metrics,
			}); err != nil {
				return eris.Wrap(err, "error encoding metrics")
			}
			return w.Flush()
		}

		start := time.Now()
		if err := func() error {
			var id string
			var key string
			var context fiber.Map
			var contextID uint

			iters := make([]float64, 0, capacity)
			metrics, values := make([]SearchAlignedMetricsResponse, 0), make([]float64, 0, capacity)
			addMetrics := func() {
				if key != "" {
					metrics = append(metrics, SearchAlignedMetricsResponse{
						Name:        key,
						Context:     context,
						XAxisValues: toNumpy(values),
						XAxisIters:  toNumpy(iters),
					})
				}
			}

			for rows.Next() {
				metric, err := next(rows)
				if err != nil {
					return eris.Wrap(err, "error getting next result")
				}

				// New series of metrics
				if metric.Key != key || metric.RunID != id || metric.ContextID != contextID {
					addMetrics()
					if metric.RunID != id {
						if err := flushMetrics(id, metrics); err != nil {
							return eris.Wrap(err, "error flushing metrics")
						}
						id, metrics = metric.RunID, metrics[:0]
					}
					key, values, iters, context = metric.Key, values[:0], iters[:0], fiber.Map{}
				}

				v := metric.Value
				if metric.IsNan {
					v = math.NaN()
				}

				iters, values = append(iters, float64(metric.Iter)), append(values, v)
				if metric.Context != nil {
					// to be properly decoded by AIM UI, json should be represented as a key:value object.
					if err := json.Unmarshal(metric.Context, &context); err != nil {
						return eris.Wrap(err, "error unmarshalling `context` json to `fiber.Map` object")
					}
					contextID = metric.ContextID
				}
			}

			addMetrics()
			if err := flushMetrics(id, metrics); err != nil {
				return eris.Wrap(err, "error flushing metrics")
			}

			return nil
		}(); err != nil {
			log.Errorf("error encountered in %s %s: error streaming metrics: %s", ctx.Method(), ctx.Path(), err)
		}
		log.Infof("body - %s %s %s", time.Since(start), ctx.Method(), ctx.Path())
	})
}
