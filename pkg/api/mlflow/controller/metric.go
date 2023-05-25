package controller

import (
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/response"
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

	metrics, err := c.metricService.GetMetricHistoryBulk(ctx, &req)
	if err != nil {
		return err
	}

	resp := response.NewMetricHistoryBulkResponse(metrics)
	log.Debugf("getMetricHistoryBulk response: %#v", resp)

	return ctx.JSON(resp)
}

// GetMetricHistories handles `POST /metrics/get-histories` endpoint.
func (c Controller) GetMetricHistories(ctx *fiber.Ctx) error {
	return nil
}
