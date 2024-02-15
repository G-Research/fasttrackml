package controller

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/api/aim2/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/api/response"
	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/pkg/common/middleware/namespace"
)

// GetExperiments handles `GET /experiments` endpoint.
func (c Controller) GetExperiments(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("getExperiments namespace: %s", ns.Code)

	experiments, err := c.experimentService.GetExperiments(ctx.Context(), ns)
	if err != nil {
		return err
	}

	resp := response.NewGetExperimentsResponse(experiments)
	log.Debugf("getExperiments response: %#v", resp)

	return ctx.JSON(resp)
}

// GetExperiment handles `GET /experiments/:id` endpoint.
func (c Controller) GetExperiment(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("getExperiment namespace: %s", ns.Code)

	req := request.GetExperimentRequest{}
	if err = ctx.ParamsParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	experiment, err := c.experimentService.GetExperiment(ctx.Context(), ns, &req)
	if err != nil {
		return err
	}

	resp := response.NewGetExperimentResponse(experiment)
	log.Debugf("getExperiment response: %#v", resp)

	return ctx.JSON(resp)
}

// GetExperimentRuns handles `GET /experiments/:id/runs` endpoint.
func (c Controller) GetExperimentRuns(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("getExperimentRuns namespace: %s", ns.Code)

	req := request.GetExperimentRunsRequest{}
	if err = ctx.QueryParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	if err = ctx.ParamsParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	runs, err := c.experimentService.GetExperimentRuns(ctx.Context(), ns, &req)
	if err != nil {
		return err
	}

	resp := response.NewGetExperimentRunsResponse(req.ID, runs)
	log.Debugf("getExperimentRuns response: %#v", resp)

	return ctx.JSON(resp)
}

// GetExperimentActivity handles `GET /experiments/:id/activity` endpoint.
func (c Controller) GetExperimentActivity(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("getExperimentActivity namespace: %s", ns.Code)

	tzOffset, err := strconv.Atoi(ctx.Get("x-timezone-offset", "0"))
	if err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "x-timezone-offset header is not a valid integer")
	}

	req := request.GetExperimentActivityRequest{}
	if err = ctx.ParamsParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	activity, err := c.experimentService.GetExperimentActivity(ctx.Context(), ns, &req, tzOffset)
	if err != nil {
		return err
	}

	resp := response.NewGetExperimentActivityResponse(activity)
	log.Debugf("getExperimentActivity response: %#v", resp)

	return ctx.JSON(resp)
}

// DeleteExperiment handles `DELETE /experiments/:id` endpoint.
func (c Controller) DeleteExperiment(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("deleteExperiment namespace: %s", ns.Code)

	req := request.DeleteExperimentRequest{}
	if err = ctx.ParamsParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	if err := c.experimentService.DeleteExperiment(ctx.Context(), ns, &req); err != nil {
		return err
	}

	return ctx.JSON(response.NewDeleteExperimentResponse(req.ID, "OK"))
}

// UpdateExperiment handles `PUT /experiments/:id` endpoint.
func (c Controller) UpdateExperiment(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("updateExperiment namespace: %s", ns.Code)

	req := request.UpdateExperimentRequest{}
	if err = ctx.ParamsParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	if err = ctx.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	if err := c.experimentService.UpdateExperiment(ctx.Context(), ns, &req); err != nil {
		return err
	}

	return ctx.JSON(response.NewUpdateExperimentResponse(req.ID, "OK"))
}
