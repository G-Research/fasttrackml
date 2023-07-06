package controller

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

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/response"
	"github.com/G-Research/fasttrackml/pkg/database"
)

// GetMetricHistory handles `GET /metrics/get-history` endpoint.
func (c Controller) GetMetricHistory(ctx *fiber.Ctx) error {
	req := request.GetMetricHistoryRequest{}
	if err := ctx.QueryParser(&req); err != nil {
		return api.NewBadRequestError(err.Error())
	}
	log.Debugf("getMetricHistory request: %#v", req)
	metrics, err := c.metricService.GetMetricHistory(ctx.Context(), &req)
	if err != nil {
		return err
	}

	resp := response.NewMetricHistoryResponse(metrics)
	log.Debugf("getMetricHistory response: %#v", resp)

	return ctx.JSON(resp)
}

// GetMetricHistoryBulk handles `GET /metrics/get-history-bulk` endpoint.
func (c Controller) GetMetricHistoryBulk(ctx *fiber.Ctx) error {
	req := request.GetMetricHistoryBulkRequest{}
	if err := ctx.QueryParser(&req); err != nil {
		return api.NewBadRequestError(err.Error())
	}
	log.Debugf("getMetricHistoryBulk request: %#v", req)

	metrics, err := c.metricService.GetMetricHistoryBulk(ctx.Context(), &req)
	if err != nil {
		return err
	}

	resp := response.NewMetricHistoryBulkResponse(metrics)
	log.Debugf("getMetricHistoryBulk response: %#v", resp)

	return ctx.JSON(resp)
}

// GetMetricHistories handles `POST /metrics/get-histories` endpoint.
func (c Controller) GetMetricHistories(ctx *fiber.Ctx) error {
	var req request.GetMetricHistoriesRequest
	if err := ctx.BodyParser(&req); err != nil {
		return api.NewBadRequestError("unable to decode request body: %s", err)
	}
	log.Debugf("GetMetricHistories request: %#v", req)

	rows, iterator, err := c.metricService.GetMetricHistories(ctx.Context(), &req)
	if err != nil {
		return err
	}

	ctx.Set("Content-Type", "application/octet-stream")
	ctx.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
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
			writer := ipc.NewWriter(w, ipc.WithAllocator(pool), ipc.WithSchema(schema))
			defer writer.Close()

			b := array.NewRecordBuilder(pool, schema)
			defer b.Release()

			for i := 0; rows.Next(); i++ {
				var m database.Metric
				iterator(rows, &m)
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
					if err := WriteStreamingRecord(writer, b.NewRecord()); err != nil {
						return fmt.Errorf("unable to write Arrow record batch: %w", err)
					}
				}
			}
			if b.Field(0).Len() > 0 {
				if err := WriteStreamingRecord(writer, b.NewRecord()); err != nil {
					return fmt.Errorf("unable to write Arrow record batch: %w", err)
				}
			}

			return nil
		}(); err != nil {
			log.Errorf("error encountered in %s %s: error streaming metrics: %s", ctx.Method(), ctx.Path(), err)
		}
		log.Infof("body - %s %s %s", time.Since(start), ctx.Method(), ctx.Path())
	})
	return nil
}
