package mlflow

import (
	"bufio"

	"github.com/G-Resarch/fasttrack/pkg/database"

	"github.com/apache/arrow/go/v11/arrow"
	"github.com/apache/arrow/go/v11/arrow/array"
	"github.com/apache/arrow/go/v11/arrow/ipc"
	"github.com/apache/arrow/go/v11/arrow/memory"
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
)

func GetMetricHistory(c *fiber.Ctx) error {
	id := c.Query("run_id", c.Query("run_uuid"))
	key := c.Query("metric_key")

	log.Debugf("GetMetricHistory request: run_id=%q, metric_key=%q", id, key)

	if id == "" {
		return NewError(ErrorCodeInvalidParameterValue, "Missing value for required parameter 'run_id'")
	}
	if key == "" {
		return NewError(ErrorCodeInvalidParameterValue, "Missing value for required parameter 'metric_key'")
	}

	var metrics []database.Metric
	if tx := database.DB.Where("run_uuid = ?", id).Where("key = ?", key).Find(&metrics); tx.Error != nil {
		return NewError(ErrorCodeInternalError, "Unable to get metric history for metric %q of run %q", key, id)
	}

	resp := &GetMetricHistoryResponse{
		Metrics: make([]Metric, len(metrics)),
	}
	for n, m := range metrics {

		resp.Metrics[n] = Metric{
			Key:       m.Key,
			Value:     m.Value,
			Timestamp: m.Timestamp,
			Step:      m.Step,
		}
		if m.IsNan {
			resp.Metrics[n].Value = "NaN"
		}
	}

	log.Debugf("GetMetricHistory response: %#v", resp)

	return c.JSON(resp)
}

func GetMetricHistoryBulk(c *fiber.Ctx) error {
	q := struct {
		RunIDs     []string `query:"run_id"`
		MetricKey  string   `query:"metric_key"`
		MaxResults int      `query:"max_results"`
	}{}

	if err := c.QueryParser(&q); err != nil {
		return NewError(ErrorCodeBadRequest, err.Error())
	}

	if len(q.RunIDs) == 0 {
		return NewError(ErrorCodeInvalidParameterValue, "GetMetricHistoryBulk request must specify at least one run_id.")
	}

	if len(q.RunIDs) > 200 {
		return NewError(ErrorCodeInvalidParameterValue, "GetMetricHistoryBulk request cannot specify more than 200 run_ids. Received %d run_ids.", len(q.RunIDs))
	}

	if q.MetricKey == "" {
		return NewError(ErrorCodeInvalidParameterValue, "GetMetricHistoryBulk request must specify a metric_key.")
	}

	if q.MaxResults == 0 {
		q.MaxResults = 25000
	}

	log.Debugf("GetMetricHistoryBulk request: %#v", q)

	var dbMetrics []database.Metric
	if tx := database.DB.
		Where("run_uuid IN ?", q.RunIDs).
		Where("key = ?", q.MetricKey).
		Order("run_uuid").
		Order("timestamp").
		Order("step").
		Order("value").
		Limit(q.MaxResults).
		Find(&dbMetrics); tx.Error != nil {
		return NewError(ErrorCodeInternalError, "Unable to get metric history in bulk for metric %q of runs %q", q.MetricKey, q.RunIDs)
	}

	metrics := make([]fiber.Map, len(dbMetrics))
	for n, m := range dbMetrics {
		metrics[n] = fiber.Map{
			"run_id":    m.RunID,
			"key":       m.Key,
			"step":      m.Step,
			"timestamp": m.Timestamp,
			"value":     m.Value,
		}
		if m.IsNan {
			metrics[n]["value"] = "NaN"
		}
	}

	resp := fiber.Map{
		"metrics": metrics,
	}

	log.Debugf("GetMetricHistoryBulk response: %#v", resp)

	return c.JSON(resp)
}

func GetMetricHistories(c *fiber.Ctx) error {
	var req GetMetricHistoriesRequest
	if err := c.BodyParser(&req); err != nil {
		return NewError(ErrorCodeBadRequest, "Unable to decode request body: %s", err)
	}

	log.Debugf("GetMetricHistories request: %#v", req)

	if len(req.ExperimentIDs) > 0 && len(req.RunIDs) > 0 {
		return NewError(ErrorCodeInvalidParameterValue, "experiment_ids and run_ids cannot both be specified at the same time")
	}

	// Filter by experiments
	if len(req.ExperimentIDs) > 0 {
		tx := database.DB.Model(&database.Run{}).
			Where("experiment_id IN ?", req.ExperimentIDs)

		// ViewType
		var lifecyleStages []database.LifecycleStage
		switch req.ViewType {
		case ViewTypeActiveOnly, "":
			lifecyleStages = []database.LifecycleStage{
				database.LifecycleStageActive,
			}
		case ViewTypeDeletedOnly:
			lifecyleStages = []database.LifecycleStage{
				database.LifecycleStageDeleted,
			}
		case ViewTypeAll:
			lifecyleStages = []database.LifecycleStage{
				database.LifecycleStageActive,
				database.LifecycleStageDeleted,
			}
		default:
			return NewError(ErrorCodeInvalidParameterValue, "Invalid run_view_type %q", req.ViewType)
		}
		tx.Where("lifecycle_stage IN ?", lifecyleStages)

		tx.Pluck("run_uuid", &req.RunIDs)
		if tx.Error != nil {
			return NewError(ErrorCodeInternalError, "Unable to search runs: %s", tx.Error)
		}
	}

	tx := database.DB.Model(&database.Metric{})

	// Filter by runs
	tx.Where("metrics.run_uuid IN ?", req.RunIDs)

	// MaxResults
	limit := int(req.MaxResults)
	if limit == 0 {
		limit = 10000000
	} else if limit > 1000000000 {
		return NewError(ErrorCodeInvalidParameterValue, "Invalid value for parameter 'max_results' supplied.")
	}
	tx.Limit(limit)

	// Default order
	tx.Joins("JOIN runs on runs.run_uuid = metrics.run_uuid").
		Order("runs.start_time DESC").
		Order("metrics.run_uuid").
		Order("metrics.key").
		Order("metrics.step").
		Order("metrics.timestamp").
		Order("metrics.value")

	if len(req.MetricKeys) > 0 {
		tx.Where("metrics.key IN ?", req.MetricKeys)
	}

	// Actual query
	rows, err := tx.Rows()
	if err != nil {
		return NewError(ErrorCodeInternalError, "Unable to search runs: %s", err)
	}

	// Stream it as Arrow record batches
	c.Set("Content-Type", "application/octet-stream")
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer rows.Close()

		pool := memory.NewGoAllocator()

		schema := arrow.NewSchema(
			[]arrow.Field{
				{Name: "run_id", Type: arrow.BinaryTypes.String},
				{Name: "key", Type: arrow.BinaryTypes.String},
				{Name: "step", Type: arrow.PrimitiveTypes.Int64},
				{Name: "timestamp", Type: arrow.PrimitiveTypes.Int64},
				{Name: "value", Type: arrow.PrimitiveTypes.Float64},
			},
			nil,
		)
		ww := ipc.NewWriter(w, ipc.WithAllocator(pool), ipc.WithSchema(schema))
		defer ww.Close()

		b := array.NewRecordBuilder(pool, schema)
		defer b.Release()

		for i := 0; rows.Next(); i++ {
			var m database.Metric
			database.DB.ScanRows(rows, &m)
			b.Field(0).(*array.StringBuilder).Append(m.RunID)
			b.Field(1).(*array.StringBuilder).Append(m.Key)
			b.Field(2).(*array.Int64Builder).Append(m.Step)
			b.Field(3).(*array.Int64Builder).Append(m.Timestamp)
			if m.IsNan {
				b.Field(4).(*array.Float64Builder).AppendNull()
			} else {
				b.Field(4).(*array.Float64Builder).Append(m.Value)
			}
			if (i+1)%100000 == 0 {
				if err := WriteRecord(ww, b.NewRecord()); err != nil {
					log.Errorf("unable to write Arrow record batch: %s", err)
					return
				}
			}
		}
		if b.Field(0).Len() > 0 {
			if err := WriteRecord(ww, b.NewRecord()); err != nil {
				log.Errorf("unable to write Arrow record batch: %s", err)
			}
		}
	})

	return nil
}

func WriteRecord(w *ipc.Writer, r arrow.Record) error {
	defer r.Release()
	return w.Write(r)
}
