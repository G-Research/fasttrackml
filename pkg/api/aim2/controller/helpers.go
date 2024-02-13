package controller

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"

	"github.com/G-Research/fasttrackml/pkg/api/aim/encoding"
	"github.com/G-Research/fasttrackml/pkg/database"
)

// RunsSearchAsCSVResponse formats and sends Runs search response as a CSV file.
//
//nolint:gocyclo
func RunsSearchAsCSVResponse(ctx *fiber.Ctx, runs []database.Run, excludeTraces, excludeParams bool) {
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
							paramData[param.Key][run.ID] = param.Value
						} else {
							paramKeys = append(paramKeys, param.Key)
							paramData[param.Key] = map[string]string{run.ID: param.Value}
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

// RunsSearchAsStreamResponse formats and sends Runs search response as a stream.
//
//nolint:gocyclo
func RunsSearchAsStreamResponse(
	ctx *fiber.Ctx, runs []database.Run, total int64, excludeTraces, excludeParams, reportProgress bool,
) {
	ctx.Set("Content-Type", "application/octet-stream")
	ctx.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		start := time.Now()
		if err := func() error {
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
						params[p.Key] = p.Value
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

// CompareJson compares two json objects.
func CompareJson(json1, json2 []byte) bool {
	var j, j2 interface{}
	if err := json.Unmarshal(json1, &j); err != nil {
		return false
	}
	if err := json.Unmarshal(json2, &j2); err != nil {
		return false
	}
	return reflect.DeepEqual(j2, j)
}
