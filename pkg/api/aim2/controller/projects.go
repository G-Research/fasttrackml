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

// GetProjectPinnedSequences handles `GET /projects/pinned-sequences` endpoint.
func (c Controller) GetProjectPinnedSequences(ctx *fiber.Ctx) error {
	return ctx.JSON(fiber.Map{
		"sequences": []string{},
	})
}

// UpdateProjectPinnedSequences handles `PUT /projects/pinned-sequences` endpoint.
func (c Controller) UpdateProjectPinnedSequences(ctx *fiber.Ctx) error {
	return ctx.JSON(fiber.Map{
		"sequences": []string{},
	})
}

// GetProjectParams handles `GET /projects/params` endpoint.
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

	params, err := c.projectService.GetProjectParams(ctx.Context(), ns.ID, &req)
	if err != nil {
		return err
	}

	resp, err := response.NewProjectParamsResponse(params, req.ExcludeParams, req.Sequences)
	if err != nil {
		return api.NewInternalError("error creating response object: %s", err)
	}
	log.Debugf("getProjectParams response: %#v", resp)

	return ctx.JSON(resp)
}

// GetProjectStatus handles `PUT /projects/status` endpoint.
func (c Controller) GetProjectStatus(ctx *fiber.Ctx) error {
	return ctx.JSON("up-to-date")
}
