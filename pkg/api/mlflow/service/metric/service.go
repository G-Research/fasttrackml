package metric

import (
	"bufio"
	"fmt"
	"time"

	"github.com/apache/arrow/go/v11/arrow"
	"github.com/apache/arrow/go/v11/arrow/array"
	"github.com/apache/arrow/go/v11/arrow/ipc"
	"github.com/apache/arrow/go/v11/arrow/memory"
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrack/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrack/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrack/pkg/api/mlflow/api/response"
	"github.com/G-Research/fasttrack/pkg/database"
)

func GetMetricHistory(c *fiber.Ctx) error {
	req := request.GetMetricHistoryRequest{}
	if err := c.QueryParser(&req); err != nil {
		return api.NewBadRequestError(err.Error())
	}

	log.Debugf("GetMetricHistory request: %#v", req)
	if err := ValidateGetMetricHistoryRequest(&req); err != nil {
		return err
	}

	var metrics []database.Metric
	if err := database.DB.Where(
		"run_uuid = ?", req.GetRunID(),
	).Where(
		"key = ?", req.MetricKey,
	).Find(&metrics).Error; err != nil {
		return api.NewInternalError(
			"Unable to get metric history for metric %q of run %q", req.MetricKey, req.GetRunID(),
		)
	}

	resp := response.NewMetricHistoryResponse(metrics)

	log.Debugf("GetMetricHistory response: %#v", resp)

	return c.JSON(resp)
}

func GetMetricHistoryBulk(c *fiber.Ctx) error {
	req := request.GetMetricHistoryBulkRequest{}
	if err := c.QueryParser(&req); err != nil {
		return api.NewBadRequestError(err.Error())
	}

	log.Debugf("GetMetricHistoryBulk request: %#v", req)
	if err := ValidateGetMetricHistoryBulkRequest(&req); err != nil {
		return err
	}

	var metrics []database.Metric
	if err := database.DB.
		Where("run_uuid IN ?", req.RunIDs).
		Where("key = ?", req.MetricKey).
		Order("run_uuid").
		Order("timestamp").
		Order("step").
		Order("value").
		Limit(req.MaxResults).
		Find(&metrics).Error; err != nil {
		return api.NewInternalError(
			"Unable to get metric history in bulk for metric %q of runs %q", req.MetricKey, req.RunIDs,
		)
	}

	resp := response.NewMetricHistoryBulkResponse(metrics)

	log.Debugf("GetMetricHistoryBulk response: %#v", resp)

	return c.JSON(resp)
}

func GetMetricHistories(c *fiber.Ctx) error {
	var req request.GetMetricHistoriesRequest
	if err := c.BodyParser(&req); err != nil {
		return api.NewBadRequestError("Unable to decode request body: %s", err)
	}

	log.Debugf("GetMetricHistories request: %#v", req)
	if err := ValidateGetMetricHistoriesRequest(&req); err != nil {
		return err
	}

	// Filter by experiments
	if len(req.ExperimentIDs) > 0 {
		tx := database.DB.Model(&database.Run{}).
			Where("experiment_id IN ?", req.ExperimentIDs)

		// ViewType
		var lifecyleStages []database.LifecycleStage
		switch req.ViewType {
		case string(request.ViewTypeActiveOnly), "":
			lifecyleStages = []database.LifecycleStage{
				database.LifecycleStageActive,
			}
		case string(request.ViewTypeDeletedOnly):
			lifecyleStages = []database.LifecycleStage{
				database.LifecycleStageDeleted,
			}
		case string(request.ViewTypeAll):
			lifecyleStages = []database.LifecycleStage{
				database.LifecycleStageActive,
				database.LifecycleStageDeleted,
			}
		default:
			return api.NewInvalidParameterValueError("Invalid run_view_type %q", req.ViewType)
		}
		tx.Where("lifecycle_stage IN ?", lifecyleStages)

		tx.Pluck("run_uuid", &req.RunIDs)
		if tx.Error != nil {
			return api.NewInternalError("Unable to search runs: %s", tx.Error)
		}
	}

	tx := database.DB.Model(&database.Metric{})

	// Filter by runs
	tx.Where("metrics.run_uuid IN ?", req.RunIDs)

	// MaxResults
	limit := int(req.MaxResults)
	if limit == 0 {
		limit = 10000000
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
		return api.NewInternalError("Unable to search runs: %s", err)
	}

	// Stream it as Arrow record batches
	c.Set("Content-Type", "application/octet-stream")
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer rows.Close()

		start := time.Now()
		if err := func() error {
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
						return fmt.Errorf("unable to write Arrow record batch: %w", err)
					}
				}
			}
			if b.Field(0).Len() > 0 {
				if err := WriteRecord(ww, b.NewRecord()); err != nil {
					return fmt.Errorf("unable to write Arrow record batch: %w", err)
				}
			}

			return nil
		}(); err != nil {
			log.Errorf("Error encountered in %s %s: error streaming metrics: %s", c.Method(), c.Path(), err)
		}
		log.Infof("body - %s %s %s", time.Since(start), c.Method(), c.Path())
	})

	return nil
}

func WriteRecord(w *ipc.Writer, r arrow.Record) error {
	defer r.Release()
	return w.Write(r)
}
