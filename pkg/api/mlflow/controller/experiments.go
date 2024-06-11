package controller

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/response"
	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/pkg/common/middleware"
)

// CreateExperiment handles `POST /experiments/create` endpoint.
func (c Controller) CreateExperiment(ctx *fiber.Ctx) error {
	var req request.CreateExperimentRequest
	if err := ctx.BodyParser(&req); err != nil {
		if err, ok := err.(*json.UnmarshalTypeError); ok {
			return api.NewInvalidParameterValueError(
				`Invalid value for parameter '%s' supplied. Hint: Value was of type '%s'. `+
					`See the API docs for more information about request parameters.`,
				err.Field, err.Value,
			)
		}
		return api.NewBadRequestError("Unable to decode request body: %s", err)
	}
	log.Debugf("createExperiment request: %#v", req)
	ns, err := middleware.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("createExperiment namespace: %s", ns.Code)
	experiment, err := c.experimentService.CreateExperiment(ctx.Context(), ns, &req)
	if err != nil {
		return err
	}

	resp := response.NewCreateExperimentResponse(experiment)
	log.Debugf("createExperiment response: %#v", resp)

	return ctx.JSON(resp)
}

// UpdateExperiment handles `POST /experiments/update` endpoint.
func (c Controller) UpdateExperiment(ctx *fiber.Ctx) error {
	var req request.UpdateExperimentRequest
	if err := ctx.BodyParser(&req); err != nil {
		return api.NewBadRequestError("Unable to decode request body: %s", err)
	}
	log.Debugf("updateExperiment request: %#v", req)
	ns, err := middleware.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("createExperiment namespace: %s", ns.Code)
	if err := c.experimentService.UpdateExperiment(ctx.Context(), ns, &req); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{})
}

// GetExperiment handles `GET /experiments/get` endpoint.
func (c Controller) GetExperiment(ctx *fiber.Ctx) error {
	var req request.GetExperimentRequest
	if err := ctx.QueryParser(&req); err != nil {
		return api.NewBadRequestError(err.Error())
	}
	log.Debugf("getExperiment request: %#v", req)
	ns, err := middleware.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("getExperiment namespace: %s", ns.Code)

	experiment, err := c.experimentService.GetExperiment(ctx.Context(), ns, &req)
	if err != nil {
		return err
	}
	resp := response.NewExperimentResponse(experiment)
	log.Debugf("getExperiment response: %#v", resp)
	return ctx.JSON(resp)
}

// GetExperimentByName handles `GET /experiments/get-by-name` endpoint.
func (c Controller) GetExperimentByName(ctx *fiber.Ctx) error {
	var req request.GetExperimentRequest
	if err := ctx.QueryParser(&req); err != nil {
		return api.NewBadRequestError(err.Error())
	}
	log.Debugf("getExperimentByName request: %#v", req)
	ns, err := middleware.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("getExperimentByName namespace: %s", ns.Code)

	experiment, err := c.experimentService.GetExperimentByName(ctx.Context(), ns, &req)
	if err != nil {
		return err
	}
	resp := response.NewExperimentResponse(experiment)
	log.Debugf("getExperimentByName response: %#v", resp)
	return ctx.JSON(resp)
}

// DeleteExperiment handles `POST /experiments/delete` endpoint.
func (c Controller) DeleteExperiment(ctx *fiber.Ctx) error {
	var req request.DeleteExperimentRequest
	if err := ctx.BodyParser(&req); err != nil {
		return api.NewBadRequestError("Unable to decode request body: %s", err)
	}
	log.Debugf("deleteExperiment request: %#v", req)
	ns, err := middleware.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("deleteExperiment namespace: %s", ns.Code)
	if err := c.experimentService.DeleteExperiment(ctx.Context(), ns, &req); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{})
}

// RestoreExperiment handles `POST /experiments/restore` endpoint.
func (c Controller) RestoreExperiment(ctx *fiber.Ctx) error {
	var req request.RestoreExperimentRequest
	if err := ctx.BodyParser(&req); err != nil {
		return api.NewBadRequestError("Unable to decode request body: %s", err)
	}
	log.Debugf("restoreExperiment request: %#v", req)
	ns, err := middleware.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("restoreExperiment namespace: %s", ns.Code)
	if err := c.experimentService.RestoreExperiment(ctx.Context(), ns, &req); err != nil {
		return err
	}
	return ctx.JSON(fiber.Map{})
}

// SetExperimentTag handles `POST /experiments/set-experiment-tag` endpoint.
func (c Controller) SetExperimentTag(ctx *fiber.Ctx) error {
	var req request.SetExperimentTagRequest
	if err := ctx.BodyParser(&req); err != nil {
		return api.NewBadRequestError("Unable to decode request body: %s", err)
	}
	log.Debugf("setExperimentTag request: %#v", req)
	ns, err := middleware.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("setExperimentTag namespace: %s", ns.Code)
	if err := c.experimentService.SetExperimentTag(ctx.Context(), ns, &req); err != nil {
		return err
	}
	return ctx.JSON(fiber.Map{})
}

// SearchExperiments handles `GET /experiments/list`, `GET /experiments/search`, `POST /experiments/search` endpoints.
func (c Controller) SearchExperiments(ctx *fiber.Ctx) error {
	var req request.SearchExperimentsRequest
	switch ctx.Method() {
	case fiber.MethodPost:
		if err := ctx.BodyParser(&req); err != nil {
			return api.NewBadRequestError("Unable to decode request body: %s", err)
		}
	case fiber.MethodGet:
		if err := ctx.QueryParser(&req); err != nil {
			return api.NewBadRequestError(err.Error())
		}
	}
	log.Debugf("searchExperiments request: %#v", req)
	ns, err := middleware.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("searchExperiments namespace: %s", ns.Code)
	experiments, limit, offset, err := c.experimentService.SearchExperiments(ctx.Context(), ns, &req)
	if err != nil {
		return err
	}

	resp, err := response.NewSearchExperimentsResponse(experiments, limit, offset)
	if err != nil {
		return api.NewInternalError("unable to build next_page_token: %s", err)
	}
	log.Debugf("searchExperiments response: %#v", resp)
	return ctx.JSON(resp)
}
