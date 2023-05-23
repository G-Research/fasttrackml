package controller

import (
	"encoding/json"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/response"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
)

// CreateExperiment handles `POST /experiments/create` endpoint.
func (c Controller) CreateExperiment(ctx *fiber.Ctx) error {
	var req request.CreateExperimentRequest
	if err := ctx.BodyParser(&req); err != nil {
		if err, ok := err.(*json.UnmarshalTypeError); ok {
			// TODO:DSuhinin do we really need this? or maybe we can always just return a common error message like in other cases?
			return api.NewInvalidParameterValueError("Invalid value for parameter '%s' supplied. Hint: Value was of type '%s'. See the API docs for more information about request parameters.", err.Field, err.Value)
		}
		return api.NewBadRequestError("Unable to decode request body: %s", err)
	}
	log.Debugf("CreateExperiment request: %#v", req)
	experiment, err := c.experimentService.CreateExperiment(ctx.Context(), &req)
	if err != nil {
		return err
	}

	resp := response.NewCreateExperimentResponse(experiment)
	log.Debugf("CreateExperiment response: %#v", resp)

	return ctx.JSON(resp)
}

// UpdateExperiment handles `POST /experiments/update` endpoint.
func (c Controller) UpdateExperiment(ctx *fiber.Ctx) error {
	var req request.UpdateExperimentRequest
	if err := ctx.BodyParser(&req); err != nil {
		return api.NewBadRequestError("Unable to decode request body: %s", err)
	}

	if err := c.experimentService.UpdateExperiment(ctx.Context(), &req); err != nil {
		return err
	}
	log.Debugf("UpdateExperiment request: %#v", req)

	return ctx.JSON(fiber.Map{})
}

// GetExperiment handles `GET /experiments/get` endpoint.
func (c Controller) GetExperiment(ctx *fiber.Ctx) error {
	var req request.GetExperimentRequest
	if err := ctx.QueryParser(&req); err != nil {
		return api.NewBadRequestError(err.Error())
	}
	log.Debugf("GetExperiment request: %#v", req)

	experiment, err := c.experimentService.GetExperiment(ctx.Context(), &req)
	if err != nil {
		return err
	}
	resp := response.NewExperimentResponse(experiment)
	log.Debugf("GetExperiment response: %#v", resp)
	return ctx.JSON(resp)
}

// GetExperimentByName handles `GET /experiments/get-by-name` endpoint.
func (c Controller) GetExperimentByName(ctx *fiber.Ctx) error {
	var req request.GetExperimentRequest
	if err := ctx.QueryParser(&req); err != nil {
		return api.NewBadRequestError(err.Error())
	}
	log.Debugf("GetExperimentByName request: %#v", req)

	experiment, err := c.experimentService.GetExperimentByName(ctx.Context(), &req)
	if err != nil {
		return err
	}
	resp := response.NewExperimentResponse(experiment)
	log.Debugf("GetExperimentByName response: %#v", resp)
	return ctx.JSON(resp)
}

// DeleteExperiment handles `POST /experiments/delete` endpoint.
func (c Controller) DeleteExperiment(ctx *fiber.Ctx) error {
	var req request.DeleteExperimentRequest
	if err := ctx.BodyParser(&req); err != nil {
		return api.NewBadRequestError("Unable to decode request body: %s", err)
	}
	log.Debugf("DeleteExperiment request: %#v", req)
	if err := c.experimentService.DeleteExperiment(ctx.Context(), &req); err != nil {
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
	log.Debugf("RestoreExperiment request: %#v", req)
	if err := c.experimentService.RestoreExperiment(ctx.Context(), &req); err != nil {
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
	log.Debugf("SetExperimentTag request: %#v", req)
	if err := c.experimentService.SetExperimentTag(ctx.Context(), &req); err != nil {
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
	log.Debugf("SearchExperiments request: %#v", req)
	experiments, limit, offset, err := c.experimentService.SearchExperiments(ctx.Context(), &req)
	if err != nil {
		return err
	}

	resp, err := response.NewSearchExperimentsResponse(experiments, limit, offset)
	if err != nil {
		return api.NewInternalError("unable to build next_page_token: %s", err)
	}
	return ctx.JSON(resp)
}
