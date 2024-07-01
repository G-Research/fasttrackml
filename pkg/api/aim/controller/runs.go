package controller

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/api/aim/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim/api/response"
	"github.com/G-Research/fasttrackml/pkg/api/aim/services/run"
	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/pkg/common/middleware"
)

// GetRunInfo handles `GET /runs/:id/info` endpoint.
func (c Controller) GetRunInfo(ctx *fiber.Ctx) error {
	ns, err := middleware.GetNamespaceFromContext(ctx.Context())
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

	runInfo, err := c.runService.GetRunInfo(ctx.Context(), ns.ID, &req)
	if err != nil {
		return err
	}

	resp := response.NewGetRunInfoResponse(runInfo)
	log.Debugf("getRunInfo response: %#v", resp)
	return ctx.JSON(resp)
}

// GetRunMetrics handles `GET /runs/:id/metric/get-batch` endpoint.
func (c Controller) GetRunMetrics(ctx *fiber.Ctx) error {
	ns, err := middleware.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("getRunMetrics namespace: %s", ns.Code)

	req := request.GetRunMetricsRequest{}
	if err := ctx.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	metrics, metricKeysMap, err := c.runService.GetRunMetrics(ctx.Context(), ns.ID, ctx.Params("id"), &req)
	if err != nil {
		return err
	}

	resp := response.NewGetRunMetricsResponse(metrics, metricKeysMap)
	log.Debugf("getRunMetrics response: %#v", resp)
	return ctx.JSON(resp)
}

// GetRunsActive handles `GET /runs/active` endpoint.
func (c Controller) GetRunsActive(ctx *fiber.Ctx) error {
	ns, err := middleware.GetNamespaceFromContext(ctx.Context())
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

	runs, err := c.runService.GetRunsActive(ctx.Context(), ns.ID, &req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return response.NewActiveRunsStreamResponse(ctx, runs, req.ReportProgress)
}

// SearchRuns handles `GET /runs/search` endpoint.
func (c Controller) SearchRuns(ctx *fiber.Ctx) error {
	ns, err := middleware.GetNamespaceFromContext(ctx.Context())
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

	// Search runs
	runs, total, err := c.runService.SearchRuns(ctx.Context(), ns.ID, tzOffset, req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	log.Debugf("found %d runs", len(runs))

	// Choose response
	switch req.Action {
	case "export":
		response.NewRunsSearchCSVResponse(ctx, runs, req.ExcludeTraces, req.ExcludeParams)
	default:
		response.NewRunsSearchStreamResponse(ctx, runs, total, req.ExcludeTraces, req.ExcludeParams, req.ReportProgress)
	}

	return nil
}

// SearchMetrics handles `POST /runs/search/metric` endpoint.
func (c Controller) SearchMetrics(ctx *fiber.Ctx) error {
	ns, err := middleware.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("searchMetrics namespace: %s", ns.Code)

	req := request.SearchMetricsRequest{}
	if err = ctx.BodyParser(&req); err != nil {
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

	//nolint:rowserrcheck
	rows, totalRuns, result, err := c.runService.SearchMetrics(ctx.Context(), ns.ID, tzOffset, req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	response.NewStreamMetricsResponse(ctx, rows, totalRuns, result, req)
	return nil
}

// SearchAlignedMetrics handles `POST /runs/search/metric/align` endpoint.
func (c Controller) SearchAlignedMetrics(ctx *fiber.Ctx) error {
	ns, err := middleware.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("searchAlignedMetrics namespace: %s", ns.Code)

	req := request.SearchAlignedMetricsRequest{}
	if err := ctx.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	//nolint:rowserrcheck
	rows, next, capacity, err := c.runService.SearchAlignedMetrics(ctx.Context(), ns.ID, &req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	response.NewSearchAlignedMetricsResponse(ctx, rows, next, capacity)
	return nil
}

// SearchMetrics handles `POST /runs/search/image` endpoint.
func (c Controller) SearchImages(ctx *fiber.Ctx) error {
	ns, err := middleware.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("searchMetrics namespace: %s", ns.Code)

	req := request.SearchArtifactsRequest{}
	if err = ctx.QueryParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}
	if ctx.Query("report_progress") == "" {
		req.ReportProgress = true
	}

	tzOffset, err := strconv.Atoi(ctx.Get("x-timezone-offset", "0"))
	if err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "x-timezone-offset header is not a valid integer")
	}

	//nolint:rowserrcheck
	rows, totalRuns, result, err := c.runService.SearchArtifacts(ctx.Context(), ns.ID, tzOffset, req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	response.NewStreamArtifactsResponse(ctx, rows, totalRuns, result, req)
	return nil
}

// DeleteRun handles `DELETE /runs/:id` endpoint.
func (c Controller) DeleteRun(ctx *fiber.Ctx) error {
	ns, err := middleware.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("deleteRun namespace: %s", ns.Code)

	req := request.DeleteRunRequest{}
	if err = ctx.ParamsParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	if err := c.runService.DeleteRun(ctx.Context(), ns.ID, &req); err != nil {
		return err
	}

	return ctx.JSON(response.NewDeleteRunResponse(req.ID, "OK"))
}

// UpdateRun handles `PUT /runs/:id` endpoint.
func (c Controller) UpdateRun(ctx *fiber.Ctx) error {
	ns, err := middleware.GetNamespaceFromContext(ctx.Context())
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

	if err := c.runService.UpdateRun(ctx.Context(), ns.ID, &req); err != nil {
		return err
	}

	return ctx.JSON(response.NewUpdateRunResponse(req.ID, "OK"))
}

// GetRunLogs handles `GET /runs/:id/logs` endpoint.
func (c Controller) GetRunLogs(ctx *fiber.Ctx) error {
	ns, err := middleware.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("GetRunLogs namespace: %s", ns.Code)

	req := request.GetRunLogsRequest{}
	if err = ctx.ParamsParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	//nolint:rowserrcheck
	rows, next, err := c.runService.GetRunLogs(ctx.Context(), ns.ID, &req)
	if err != nil {
		return err
	}

	response.NewGetRunLogsResponse(ctx, rows, next)
	return nil
}

// ArchiveBatch handles `POST /runs/archive-batch` endpoint.
func (c Controller) ArchiveBatch(ctx *fiber.Ctx) error {
	ns, err := middleware.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("archiveBatch namespace: %s", ns.Code)

	req := request.ArchiveBatchRequest{}
	if err := ctx.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	action := run.BatchActionRestore
	if ctx.Query("archive") == "true" {
		action = run.BatchActionArchive
	}

	if err := c.runService.ProcessBatch(ctx.Context(), ns.ID, action, req); err != nil {
		return err
	}

	return ctx.JSON(response.NewArchiveBatchResponse("OK"))
}

// DeleteBatch handles `DELETE /runs/delete-batch` endpoint.
func (c Controller) DeleteBatch(ctx *fiber.Ctx) error {
	ns, err := middleware.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("deleteBatch namespace: %s", ns.Code)

	req := request.DeleteBatchRequest{}
	if err := ctx.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	if err := c.runService.ProcessBatch(ctx.Context(), ns.ID, run.BatchActionDelete, req); err != nil {
		return err
	}

	return ctx.JSON(response.NewArchiveBatchResponse("OK"))
}

// AddRunTag handles `POST /runs/:id/tags/new` endpoint.
func (c Controller) AddRunTag(ctx *fiber.Ctx) error {
	ns, err := middleware.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("addRunTag namespace: %s", ns.Code)

	req := request.AddRunTagRequest{}
	if err = ctx.ParamsParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}
	if err = ctx.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	if err := c.runService.AddRunTag(ctx.Context(), ns.ID, &req); err != nil {
		return err
	}

	return ctx.SendStatus(fiber.StatusCreated)
}

// AddRunTag handles `DELETE /runs/:id/tags/:tagID` endpoint.
func (c Controller) DeleteRunTag(ctx *fiber.Ctx) error {
	ns, err := middleware.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("deleteRunTag namespace: %s", ns.Code)

	req := request.DeleteRunTagRequest{}
	if err = ctx.ParamsParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	if err := c.runService.DeleteRunTag(ctx.Context(), ns.ID, &req); err != nil {
		return err
	}

	return ctx.SendStatus(fiber.StatusOK)
}
