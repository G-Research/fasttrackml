package controller

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/aim2/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/api/response"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/convertors"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/repositories"
	"github.com/G-Research/fasttrackml/pkg/common"
	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/pkg/common/middleware/namespace"
	"github.com/G-Research/fasttrackml/pkg/database"
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

	// validate that requested experiment exists and is not a default experiment.
	experiment := database.Experiment{}
	if err := database.DB.Select(
		"ID", "Name",
	).Where(
		"experiments.experiment_id = ?", req.ID,
	).Where(
		"experiments.namespace_id = ?", ns.ID,
	).First(&experiment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.ErrNotFound
		}
		return fmt.Errorf("unable to find experiment %q: %w", req.ID, err)
	}

	if experiment.IsDefault(ns) {
		return fiber.NewError(fiber.StatusBadRequest, "unable to delete default experiment")
	}

	// TODO this code should move to service with injected repository
	experimentRepo := repositories.NewExperimentRepository(database.DB)
	if err = experimentRepo.Delete(ctx.Context(), &models.Experiment{
		ID: common.GetPointer(req.ID),
	}); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError,
			fmt.Sprintf("unable to delete experiment %q: %s", req.ID, err))
	}

	return ctx.JSON(fiber.Map{
		"id":     req.ID,
		"status": "OK",
	})
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

	experimentRepository := repositories.NewExperimentRepository(database.DB)
	experiment, err := experimentRepository.GetByNamespaceIDAndExperimentID(ctx.Context(), ns.ID, req.ID)
	if err != nil {
		return fiber.NewError(
			fiber.StatusInternalServerError, fmt.Sprintf("unable to find experiment '%d': %s", req.ID, err),
		)
	}
	if experiment == nil {
		return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("unable to find experiment '%d'", req.ID))
	}

	experiment = convertors.ConvertUpdateExperimentToDBModel(&req, experiment)
	if req.Archived != nil || req.Name != nil {
		if err := database.DB.Transaction(func(tx *gorm.DB) error {
			if err := experimentRepository.UpdateWithTransaction(ctx.Context(), tx, experiment); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError,
				fmt.Sprintf("unable to update experiment %q: %s", req.ID, err))
		}
	}
	if req.Description != nil {
		tagRepository := repositories.NewTagRepository(database.DB)
		if err := tagRepository.CreateExperimentTag(ctx.Context(), &models.ExperimentTag{
			Key:          common.DescriptionTagKey,
			Value:        *req.Description,
			ExperimentID: *experiment.ID,
		}); err != nil {
			return err
		}
	}

	return ctx.JSON(fiber.Map{
		"id":     req.ID,
		"status": "OK",
	})
}
