package controller

import (
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/api/aim2/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/api/response"
	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/pkg/common/middleware/namespace"
)

// GetDashboards handles `GET /dashboards` endpoint.
func (c Controller) GetDashboards(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("getDashboards namespace: %s", ns.Code)
	dashboards, err := c.dashboardService.GetDashboards(ctx.Context(), ns)
	if err != nil {
		return convertError(err)
	}

	resp := response.NewGetDashboardsResponse(dashboards)
	log.Debugf("getDashboards response %#v", resp)
	return ctx.JSON(resp)
}

// CreateDashboard handles `POST /dashboards` endpoint.
func (c Controller) CreateDashboard(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("createDashboard namespace: %s", ns.Code)
	req := request.CreateDashboardRequest{}
	if err := ctx.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}
	dash, err := c.dashboardService.Create(ctx.Context(), ns, &req)
	if err != nil {
		return convertError(err)
	}

	resp := response.NewCreateDashboardResponse(dash)
	log.Debugf("createDashboard response %#v", resp)
	return ctx.JSON(resp)
}

// GetDashboard handles `GET /dashboard/:id` endpoint.
func (c Controller) GetDashboard(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("getDashboard namespace: %s", ns.Code)

	req := request.GetDashboardRequest{}
	if err := ctx.ParamsParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}
	dashboard, err := c.dashboardService.Get(ctx.Context(), ns, &req)
	if err != nil {
		return convertError(err)
	}

	resp := response.NewGetDashboardResponse(dashboard)
	log.Debugf("getDashboard response %#v", resp)
	return ctx.JSON(resp)
}

// UpdateDashboard handles `PUT /dashboard/:id` endpoint.
func (c Controller) UpdateDashboard(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("updateDashboard namespace: %s", ns.Code)

	req := request.UpdateDashboardRequest{}
	if err := ctx.ParamsParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}
	if err := ctx.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}
	dash, err := c.dashboardService.Update(ctx.Context(), ns, &req)
	if err != nil {
		return convertError(err)
	}

	resp := response.NewUpdateDashboardResponse(dash)
	log.Debugf("updateDashboard response %#v", resp)
	return ctx.JSON(resp)
}

// DeleteDashboard handles `DELETE /dashboards/:id` endpoint.
func (c Controller) DeleteDashboard(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("deleteDashboard namespace: %s", ns.Code)

	req := request.DeleteDashboardRequest{}
	if err := ctx.ParamsParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}
	err = c.dashboardService.Delete(ctx.Context(), ns, &req)
	if err != nil {
		return convertError(err)
	}
	return ctx.Status(200).JSON(nil)
}
