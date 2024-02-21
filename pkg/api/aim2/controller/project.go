package controller

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/api/aim2/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/api/response"
	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/pkg/common/middleware/namespace"
	"github.com/G-Research/fasttrackml/pkg/database"
)

// GetProject handles `GET /projects` endpoint.
func (c Controller) GetProject(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("getProjectActivity namespace: %s", ns.Code)

	name, dialector := c.projectService.GetProjectInformation()

	return ctx.JSON(response.NewGetProjectResponse(name, dialector))
}

// GetProjectActivity handles `GET /projects/activity` endpoint.
func (c Controller) GetProjectActivity(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("getProjectActivity namespace: %s", ns.Code)

	tzOffset, err := strconv.Atoi(ctx.Get("x-timezone-offset", "0"))
	if err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "x-timezone-offset header is not a valid integer")
	}

	activity, err := c.projectService.GetProjectActivity(ctx.Context(), ns.ID, tzOffset)
	if err != nil {
		return err
	}

	resp := response.NewProjectActivityResponse(activity)
	log.Debugf("getProjectActivity response: %#v", resp)

	return ctx.JSON(resp)
}

// TODO
func (c Controller) GetProjectPinnedSequences(ctx *fiber.Ctx) error {
	return ctx.JSON(fiber.Map{
		"sequences": []string{},
	})
}

// TODO
func (c Controller) UpdateProjectPinnedSequences(ctx *fiber.Ctx) error {
	return ctx.JSON(fiber.Map{
		"sequences": []string{},
	})
}

func (c Controller) GetProjectParams(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("getProjectParams namespace: %s", ns.Code)

	req := request.GetProjectParamsRequest{}
	if err := ctx.QueryParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	resp := fiber.Map{}

	if !req.ExcludeParams {
		// fetch and process params.
		query := database.DB.Distinct().Model(
			&database.Param{},
		).Joins(
			"JOIN runs USING(run_uuid)",
		).Joins(
			"INNER JOIN experiments ON experiments.experiment_id = runs.experiment_id AND experiments.namespace_id = ?",
			ns.ID,
		).Where(
			"runs.lifecycle_stage = ?", database.LifecycleStageActive,
		)
		if len(req.Experiments) != 0 {
			query.Where("experiments.experiment_id IN ?", req.Experiments)
		}
		var paramKeys []string
		if err = query.Pluck("Key", &paramKeys).Error; err != nil {
			return fmt.Errorf("error retrieving param keys: %w", err)
		}

		params := make(map[string]any, len(paramKeys)+1)
		for _, p := range paramKeys {
			params[p] = map[string]string{
				"__example_type__": "<class 'str'>",
			}
		}

		// fetch and process tags.
		query = database.DB.Distinct().Model(
			&database.Tag{},
		).Joins(
			"JOIN runs USING(run_uuid)",
		).Joins(
			"INNER JOIN experiments ON experiments.experiment_id = runs.experiment_id AND experiments.namespace_id = ?",
			ns.ID,
		).Where(
			"runs.lifecycle_stage = ?", database.LifecycleStageActive,
		)
		if len(req.Experiments) != 0 {
			query.Where("experiments.experiment_id IN ?", req.Experiments)
		}
		var tagKeys []string
		if err = query.Pluck("Key", &tagKeys).Error; err != nil {
			return fmt.Errorf("error retrieving tag keys: %w", err)
		}

		tags := make(map[string]map[string]string, len(tagKeys))
		for _, t := range tagKeys {
			tags[t] = map[string]string{
				"__example_type__": "<class 'str'>",
			}
		}

		params["tags"] = tags
		resp["params"] = params
	}

	if len(req.Sequences) == 0 {
		req.Sequences = []string{
			"metric",
			"images",
			"texts",
			"figures",
			"distributions",
			"audios",
		}
	}

	for _, s := range req.Sequences {
		switch s {
		case "images", "texts", "figures", "distributions", "audios":
			resp[s] = fiber.Map{}
		case "metric":
			query := database.DB.Distinct().Model(
				&database.LatestMetric{},
			).Joins(
				"JOIN runs USING(run_uuid)",
			).Joins(
				"INNER JOIN experiments ON experiments.experiment_id = runs.experiment_id AND experiments.namespace_id = ?",
				ns.ID,
			).Joins(
				"Context",
			).Where(
				"runs.lifecycle_stage = ?", database.LifecycleStageActive,
			)
			if len(req.Experiments) != 0 {
				query.Where("experiments.experiment_id IN ?", req.Experiments)
			}
			var metrics []database.LatestMetric
			if err = query.Find(&metrics).Error; err != nil {
				return fmt.Errorf("error retrieving metric keys: %w", err)
			}

			data, mapped := make(map[string][]fiber.Map, len(metrics)), make(map[string]map[string]fiber.Map, len(metrics))
			for _, metric := range metrics {
				if mapped[metric.Key] == nil {
					mapped[metric.Key] = map[string]fiber.Map{}
				}
				if _, ok := mapped[metric.Key][metric.Context.GetJsonHash()]; !ok {
					// to be properly decoded by AIM UI, json should be represented as a key:value object.
					context := fiber.Map{}
					if err := json.Unmarshal(metric.Context.Json, &context); err != nil {
						return eris.Wrap(err, "error unmarshalling `context` json to `fiber.Map` object")
					}
					mapped[metric.Key][metric.Context.GetJsonHash()] = context
					data[metric.Key] = append(data[metric.Key], context)
				}
			}
			resp[s] = data
		default:
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("%q is not a valid Sequence", s))
		}
	}

	return ctx.JSON(resp)
}

func (c Controller) GetProjectStatus(ctx *fiber.Ctx) error {
	return ctx.JSON("up-to-date")
}
