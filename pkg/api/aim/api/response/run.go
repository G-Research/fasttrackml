package response

import (
	"bufio"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"
	"gorm.io/datatypes"

	"github.com/G-Research/fasttrackml/pkg/api/aim/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/aim/dao/repositories"
	"github.com/G-Research/fasttrackml/pkg/api/aim/encoding"
	mlflowCommon "github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/common"
	"github.com/G-Research/fasttrackml/pkg/common/services/artifact/storage"
	"github.com/G-Research/fasttrackml/pkg/database"
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
	Texts         []GetRunInfoTracesMetricPartial `json:"texts"`
	Audios        map[string]string               `json:"audios"`
	Metric        []GetRunInfoTracesMetricPartial `json:"metric"`
	Images        []GetRunInfoTracesMetricPartial `json:"images"`
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
	ID           string                      `json:"id,omitempty"`
	RunID        string                      `json:"run_id,omitempty"`
	Name         string                      `json:"name"`
	Description  string                      `json:"description"`
	Experiment   GetRunInfoExperimentPartial `json:"experiment"`
	Tags         []map[string]string         `json:"tags"`
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
func NewGetRunInfoResponse(run *models.Run, artifacts []storage.ArtifactObject) *GetRunInfoResponse {
	metrics := make([]GetRunInfoTracesMetricPartial, len(run.LatestMetrics))
	for i, metric := range run.LatestMetrics {
		metrics[i] = GetRunInfoTracesMetricPartial{
			Name:      metric.Key,
			Context:   json.RawMessage(metric.Context.Json),
			LastValue: 0.1,
		}
	}

	imagesCounter := 0
	textsCounter := 0
	const imageMimeType = "image/"
	const textMimeType = "text/"
	for _, artifact := range artifacts {
		filename := filepath.Base(artifact.Path)
		mime := mlflowCommon.GetContentType(filename)
		if strings.HasPrefix(mime, imageMimeType) {
			imagesCounter++
		} else if strings.HasPrefix(mime, textMimeType) {
			textsCounter++
		}
	}

	images := make([]GetRunInfoTracesMetricPartial, imagesCounter)
	texts := make([]GetRunInfoTracesMetricPartial, textsCounter)
	imagesCounter = 0
	textsCounter = 0
	for _, artifact := range artifacts {
		filename := filepath.Base(artifact.Path)
		mime := mlflowCommon.GetContentType(filename)
		if strings.HasPrefix(mime, imageMimeType) {
			images[imagesCounter] = GetRunInfoTracesMetricPartial{
				Name:      artifact.Path,
				Context:   nil,
				LastValue: 0,
			}
			imagesCounter++
		} else if strings.HasPrefix(mime, textMimeType) {
			texts[textsCounter] = GetRunInfoTracesMetricPartial{
				Name:      artifact.Path,
				Context:   nil,
				LastValue: 0,
			}
			textsCounter++
		}
	}
	params := make(GetRunInfoParamsPartial, len(run.Params)+1)
	for _, p := range run.Params {
		params[p.Key] = p.ValueAny()
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
			Texts:         texts,
			Audios:        map[string]string{},
			Metric:        metrics,
			Images:        images,
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
			Tags:         ConvertTagsToMaps(run.SharedTags),
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

// NewGetRunMetricsResponse creates a new response object for `GET /runs/:id/metric/get-batch` endpoint.
func NewGetRunMetricsResponse(metrics []models.Metric, metricKeysMap models.MetricKeysMap) []GetRunMetricsResponse {
	data := make(map[models.MetricKeysItem]struct {
		iters  []int
		values []*float64
	}, len(metricKeysMap))

	for _, item := range metrics {
		v := common.GetPointer(item.Value)
		if item.IsNan {
			v = nil
		}
		key := models.MetricKeysItem{
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

// NewSearchAlignedMetricsResponse creates a new response object for `GET /runs/search/metric/align` endpoint.
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

// DeleteRunResponse is a response object to hold response data for `DELETE /runs/:id` endpoint.
type DeleteRunResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// NewDeleteRunResponse creates new response object for `DELETE /runs/:id` endpoint.
func NewDeleteRunResponse(id string, status string) *DeleteRunResponse {
	return &DeleteRunResponse{
		ID:     id,
		Status: status,
	}
}

// UpdateRunResponse is a response object to hold response data for `PUT /runs/:id` endpoint.
type UpdateRunResponse struct {
	ID     string `json:"ID"`
	Status string `json:"status"`
}

// NewUpdateRunResponse creates new response object for `PUT /runs/:id` endpoint.
func NewUpdateRunResponse(id string, status string) *UpdateRunResponse {
	return &UpdateRunResponse{
		ID:     id,
		Status: status,
	}
}

// ArchiveBatchResponse is a response object to hold response data for `POST /runs/archive-batch` endpoint.
type ArchiveBatchResponse struct {
	Status string `json:"status"`
}

// NewArchiveBatchResponse creates new response object for `POST /runs/archive-batch` endpoint.
func NewArchiveBatchResponse(status string) *ArchiveBatchResponse {
	return &ArchiveBatchResponse{
		Status: status,
	}
}

// DeleteBatchResponse is a response object to hold response data for `DELETE /runs/delete-batch` endpoint.
type DeleteBatchResponse struct {
	Status string `json:"status"`
}

// NewDeleteBatchResponse creates new response object for `POST /runs/archive-batch` endpoint.
func NewDeleteBatchResponse(status string) *DeleteBatchResponse {
	return &DeleteBatchResponse{
		Status: status,
	}
}

// NewStreamMetricsResponse streams the provided sql.Rows to the fiber context.
//
//nolint:gocyclo
func NewStreamMetricsResponse(ctx *fiber.Ctx, rows *sql.Rows, totalRuns int64,
	result repositories.SearchResultMap, req request.SearchMetricsRequest,
) {
	ctx.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		//nolint:errcheck
		defer rows.Close()

		start := time.Now()

		var xAxis bool
		if req.XAxis != "" {
			xAxis = true
		}

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
}

// NewStreamArtifactsResponse streams the provided sql.Rows to the fiber context.
//
//nolint:gocyclo
func NewStreamArtifactsResponse(ctx *fiber.Ctx, rows *sql.Rows, runs map[string]models.Run,
	summary repositories.ArtifactSearchSummary, req request.SearchArtifactsRequest,
) {
	ctx.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		//nolint:errcheck
		defer rows.Close()

		start := time.Now()

		if err := func() error {
			var (
				runID     string
				runData   fiber.Map
				tracesMap map[string]fiber.Map
				cur       int64
			)
			reportProgress := func() error {
				if !req.ReportProgress {
					return nil
				}
				err := encoding.EncodeTree(w, fiber.Map{
					fmt.Sprintf("progress_%d", cur): []int64{cur, int64(len(runs))},
				})
				if err != nil {
					return err
				}
				cur++
				return w.Flush()
			}
			addImage := func(img models.Artifact, run models.Run) {
				maxIndex := summary.MaxIndex(img.RunID, img.Name)
				maxStep := summary.MaxStep(img.RunID, img.Name)
				if runData == nil {
					runData = fiber.Map{
						"ranges": fiber.Map{
							"record_range_total": []int{0, maxStep},
							"record_range_used":  []int{req.RecordRangeMin(), req.RecordRangeMax(maxStep)},
							"index_range_total":  []int{0, maxIndex},
							"index_range_used":   []int{req.IndexRangeMin(), req.IndexRangeMax(maxIndex)},
						},
						"params": fiber.Map{
							"images_per_step": maxIndex,
						},
						"props": renderProps(run),
					}
					tracesMap = map[string]fiber.Map{}
				}
				trace, ok := tracesMap[img.Name]
				if !ok {
					trace = fiber.Map{
						"name":    img.Name,
						"context": fiber.Map{},
						"caption": img.Caption,
					}
					tracesMap[img.Name] = trace
				}
				traceValues, ok := trace["values"].([][]fiber.Map)
				if !ok {
					stepsSlice := make([][]fiber.Map, maxStep+1)
					traceValues = stepsSlice
				}

				iters, ok := trace["iters"].([]int64)
				if !ok {
					iters = make([]int64, maxStep+1)
				}
				value := fiber.Map{
					"blob_uri": img.BlobURI,
					"caption":  img.Caption,
					"height":   img.Height,
					"width":    img.Width,
					"format":   img.Format,
					"iter":     img.Iter,
					"index":    img.Index,
					"step":     img.Step,
				}

				stepImages := traceValues[img.Step]
				if stepImages == nil {
					stepImages = []fiber.Map{}
				}
				stepImages = append(stepImages, value)
				traceValues[img.Step] = stepImages
				iters[img.Step] = img.Iter // TODO maybe not correct
				trace["values"] = traceValues
				trace["iters"] = iters
				tracesMap[img.Name] = trace
			}
			selectTraces := func() {
				// collect the traces for this run, limiting to RecordDensity and IndexDensity.
				selectIndices := func(trace fiber.Map) fiber.Map {
					// limit steps slice to len of RecordDensity.
					stepCount := req.StepCount()
					imgCount := req.ItemsPerStep()
					steps, ok := trace["values"].([][]fiber.Map)
					if !ok {
						return trace
					}
					iters, ok := trace["iters"].([]int64)
					if !ok {
						return trace
					}
					filteredSteps := [][]fiber.Map{}
					filteredIters := []int64{}
					stepInterval := len(steps) / stepCount
					for stepIndex := 0; stepIndex < len(steps); stepIndex++ {
						if stepCount == -1 ||
							len(steps) <= stepCount ||
							stepIndex%stepInterval == 0 {
							step := steps[stepIndex]
							newStep := []fiber.Map{}
							imgInterval := len(step) / imgCount
							for imgIndex := 0; imgIndex < len(step); imgIndex++ {
								if imgCount == -1 ||
									len(step) <= imgCount ||
									imgIndex%imgInterval == 0 {
									newStep = append(newStep, step[imgIndex])
								}
							}
							filteredSteps = append(filteredSteps, newStep)
							filteredIters = append(filteredIters, iters[stepIndex])
						}
					}
					trace["values"] = filteredSteps
					trace["iters"] = filteredIters
					return trace
				}

				traces := make([]fiber.Map, len(tracesMap))
				i := 0
				for _, trace := range tracesMap {
					traces[i] = selectIndices(trace)
					i++
				}
				runData["traces"] = traces
			}
			flushImages := func() error {
				if runID == "" {
					return nil
				}
				selectTraces()
				if err := encoding.EncodeTree(w, fiber.Map{
					runID: runData,
				}); err != nil {
					return err
				}
				if err := reportProgress(); err != nil {
					return err
				}
				return w.Flush()
			}
			hasRows := false
			for rows.Next() {
				var image models.Artifact
				if err := database.DB.ScanRows(rows, &image); err != nil {
					return err
				}
				// flush after each change in runID
				// (assumes order by runID)
				if image.RunID != runID {
					if err := flushImages(); err != nil {
						return err
					}
					runID = image.RunID
					runData = nil
				}
				addImage(image, runs[image.RunID])
				hasRows = true
			}

			if hasRows {
				if err := flushImages(); err != nil {
					return err
				}
				if err := reportProgress(); err != nil {
					return err
				}
			}

			return nil
		}(); err != nil {
			log.Errorf("Error encountered in %s %s: error streaming images: %s", ctx.Method(), ctx.Path(), err)
		}

		log.Infof("body - %s %s %s", time.Since(start), ctx.Method(), ctx.Path())
	})
}

// NewRunsSearchCSVResponse formats and sends Runs search response as a CSV file.
//
//nolint:gocyclo
func NewRunsSearchCSVResponse(ctx *fiber.Ctx, runs []models.Run, excludeTraces, excludeParams bool) {
	ctx.Set("Transfer-Encoding", "chunked")
	ctx.Set("Content-Type", "text/csv")
	ctx.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="runs-reports-%d.csv"`, time.Now().Unix()))

	ctx.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		//nolint:errcheck
		start := time.Now()
		if err := func() error {
			records := make([][]string, len(runs))
			tagData, tagKeys := map[string]map[string]string{}, []string{}
			paramData, paramKeys := map[string]map[string]string{}, []string{}
			metricData, metricKeys := map[string]map[string]float64{}, []string{}
			for i, run := range runs {
				// group metrics information for further usage.
				if !excludeTraces {
					for _, metric := range run.LatestMetrics {
						v := metric.Value
						if metric.IsNan {
							v = math.NaN()
						}
						key := fmt.Sprintf("%s %s", metric.Key, string(metric.Context.Json))
						if _, ok := metricData[key]; ok {
							metricData[key][run.ID] = v
						} else {
							metricKeys = append(metricKeys, key)
							metricData[key] = map[string]float64{run.ID: v}
						}
					}
				}
				// group params and tags information for further usage.
				if !excludeParams {
					for _, param := range run.Params {
						if _, ok := paramData[param.Key]; ok {
							paramData[param.Key][run.ID] = param.ValueString()
						} else {
							paramKeys = append(paramKeys, param.Key)
							paramData[param.Key] = map[string]string{run.ID: param.ValueString()}
						}
					}
					for _, tag := range run.Tags {
						if _, ok := tagData[tag.Key]; ok {
							tagData[tag.Key][run.ID] = tag.Value
						} else {
							tagKeys = append(tagKeys, tag.Key)
							tagData[tag.Key] = map[string]string{run.ID: tag.Value}
						}
					}
				}

				records[i] = []string{
					run.Name,
					run.Experiment.Name,
					"-",
					time.Unix(run.StartTime.Int64/1000, 0).Format("15:04:05 2006-01-02"),
					fmt.Sprintf("%dms", run.EndTime.Int64-run.StartTime.Int64),
				}
			}

			// process headers.
			headers := []string{
				"run",
				"experiment",
				"experiment_description",
				"date",
				"duration",
			}
			// add metrics as headers.
			slices.Sort(metricKeys)
			headers = append(headers, metricKeys...)

			// add params as headers.
			slices.Sort(paramKeys)
			for _, paramKey := range paramKeys {
				headers = append(headers, fmt.Sprintf("params[%s]", paramKey))
			}
			// add tags as headers.
			slices.Sort(tagKeys)
			for _, tagKey := range tagKeys {
				headers = append(headers, fmt.Sprintf("tags[%s]", tagKey))
			}
			writer := csv.NewWriter(w)
			if err := writer.Write(headers); err != nil {
				return err
			}

			// process data.
			chunkSize, recordCounter := 500, 0
			for i, run := range runs {
				record := records[i]
				// add metrics data.
				for _, metricKey := range metricKeys {
					if metricValue, ok := metricData[metricKey][run.ID]; ok {
						record = append(record, fmt.Sprintf("%f", metricValue))
					} else {
						record = append(record, "-")
					}
				}

				// add params data.
				for _, paramKey := range paramKeys {
					if paramValue, ok := paramData[paramKey][run.ID]; ok {
						record = append(record, paramValue)
					} else {
						record = append(record, "-")
					}
				}

				// add tags data.
				for _, tagKey := range tagKeys {
					if tagValue, ok := tagData[tagKey][run.ID]; ok {
						record = append(record, tagValue)
					} else {
						record = append(record, "-")
					}
				}

				if err := writer.Write(record); err != nil {
					return err
				}

				// divide data by chunks.
				if recordCounter >= chunkSize {
					if err := w.Flush(); err != nil {
						return err
					}
					recordCounter = 0
				} else {
					recordCounter++
				}
			}

			if err := w.Flush(); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			log.Errorf("error encountered in %s %s: error streaming runs export: %s", ctx.Method(), ctx.Path(), err)
		}

		log.Infof("body - %s %s %s", time.Since(start), ctx.Method(), ctx.Path())
	})
}

// NewRunsSearchStreamResponse formats and sends Runs search response as a stream.
//
//nolint:gocyclo
func NewRunsSearchStreamResponse(
	ctx *fiber.Ctx, runs []models.Run, total int64, excludeTraces, excludeParams, reportProgress bool,
) {
	ctx.Set("Content-Type", "application/octet-stream")
	ctx.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		start := time.Now()
		if err := func() error {
			for i, r := range runs {
				run := fiber.Map{
					"props": renderProps(r),
				}

				if !excludeTraces {
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
							},
							"context": fiber.Map{},
						}
						// to be properly decoded by AIM UI, json should be represented as a key:value object.
						context := fiber.Map{}
						if err := json.Unmarshal(m.Context.Json, &context); err != nil {
							return eris.Wrap(err, "error unmarshalling `context` json to `fiber.Map` object")
						}
						data["context"] = context
						metrics[i] = data
					}
					run["traces"] = fiber.Map{
						"metric": metrics,
					}
				}

				if !excludeParams {
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
				}

				if err := encoding.EncodeTree(w, fiber.Map{
					r.ID: run,
				}); err != nil {
					return err
				}

				if reportProgress {
					if err := encoding.EncodeTree(w, fiber.Map{
						fmt.Sprintf("progress_%d", i): []int64{total - int64(r.RowNum), total},
					}); err != nil {
						return err
					}
				}

				if err := w.Flush(); err != nil {
					return err
				}
			}

			if reportProgress {
				if err := encoding.EncodeTree(w, fiber.Map{
					fmt.Sprintf("progress_%d", len(runs)): []int64{total, total},
				}); err != nil {
					if err = w.Flush(); err != nil {
						return err
					}
				}
			}
			return nil
		}(); err != nil {
			log.Errorf("error encountered in %s %s: error streaming runs: %s", ctx.Method(), ctx.Path(), err)
		}
		log.Infof("body - %s %s %s", time.Since(start), ctx.Method(), ctx.Path())
	})
}

// NewActiveRunsStreamResponse streams the provided []models.Run to the fiber context.
func NewActiveRunsStreamResponse(ctx *fiber.Ctx, runs []models.Run, reportProgress bool) error {
	ctx.Set("Content-Type", "application/octet-stream")
	ctx.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		start := time.Now()
		if err := func() error {
			for i, r := range runs {
				props := renderProps(r)
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

				if reportProgress {
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
			return nil
		}(); err != nil {
			log.Errorf("Error encountered in %s %s: error streaming active runs: %s", ctx.Method(), ctx.Path(), err)
		}

		log.Infof("body - %s %s %s", time.Since(start), ctx.Method(), ctx.Path())
	})
	return nil
}

// renderProps makes the "props" map for a run.
func renderProps(r models.Run) fiber.Map {
	m := fiber.Map{
		"name":        r.Name,
		"description": nil,
		"experiment": fiber.Map{
			"id":                fmt.Sprintf("%d", r.ExperimentID),
			"name":              r.Experiment.Name,
			"artifact_location": r.Experiment.ArtifactLocation,
		},
		"tags":          ConvertTagsToMaps(r.SharedTags),
		"creation_time": float64(r.StartTime.Int64) / 1000,
		"end_time":      float64(r.EndTime.Int64) / 1000,
		"archived":      r.LifecycleStage == models.LifecycleStageDeleted,
		"active":        r.Status == models.StatusRunning,
	}
	return m
}

// NewRunImagesStreamResponse streams the provided images to the fiber context.
func NewRunImagesStreamResponse(ctx *fiber.Ctx, images []models.Image) error {
	ctx.Set("Content-Type", "application/octet-stream")
	ctx.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		start := time.Now()
		if err := func() error {
			var values [][]map[string]interface{}
			var valuesResult []map[string]interface{}

			for _, image := range images {
				for _, valueArray := range image.Values {
					for _, val := range valueArray {
						valMap := map[string]interface{}{
							"blob_uri": val.BlobURI,
							"caption":  val.Caption,
							"context":  val.Context,
							"format":   val.Format,
							"height":   val.Height,
							"index":    val.Index,
							"key":      val.Key,
							"seqKey":   val.SeqKey,
							"name":     val.Name,
							"run":      val.Run,
							"step":     val.Step,
							"width":    val.Width,
						}
						valuesResult = append(valuesResult, valMap)
					}
				}

				values = append(values, valuesResult)
				imgMap := map[string]interface{}{
					"record_range": image.RecordRange,
					"index_range":  image.IndexRange,
					"name":         image.Name,
					"context":      image.Context,
					"values":       values,
					"iters":        image.Iters,
				}

				if err := encoding.EncodeTree(w, imgMap); err != nil {
					return err
				}
			}

			return nil
		}(); err != nil {
			log.Errorf("Error encountered in %s %s: error streaming active runs: %s", ctx.Method(), ctx.Path(), err)
		}

		log.Infof("body - %s %s %s", time.Since(start), ctx.Method(), ctx.Path())
	})
	return nil
}

// NewRunImagesBatchStreamResponse streams the provided images to the fiber context.
func NewRunImagesBatchStreamResponse(ctx *fiber.Ctx, imagesMap map[string]any) error {
	ctx.Context().Response.SetBodyStreamWriter(func(w *bufio.Writer) {
		start := time.Now()
		if err := func() error {
			if err := encoding.EncodeTree(w, imagesMap); err != nil {
				return err
			}
			if err := w.Flush(); err != nil {
				return eris.Wrap(err, "error flushing output stream")
			}
			return nil
		}(); err != nil {
			log.Errorf(
				"error encountered in %s %s: error streaming artifact: %s",
				ctx.Method(),
				ctx.Path(),
				err,
			)
		}
		log.Infof("body - %s %s %s", time.Since(start), ctx.Method(), ctx.Path())
	})
	return nil
}
