package controller

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/api/aim/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim/api/response"
	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/pkg/common/middleware"
)

// GetApps handles `GET /apps` endpoint.
func (c Controller) GetApps(ctx *fiber.Ctx) error {
	ns, err := middleware.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("getApps namespace: %s", ns.Code)

	apps, err := c.appService.GetApps(ctx.Context(), ns.ID)
	if err != nil {
		return err
	}

	resp := response.NewGetAppsResponse(apps)
	log.Debugf("getApps response: %#v", resp)

	return ctx.JSON(resp)
}

// CreateApp handles `POST /apps` endpoint.
func (c Controller) CreateApp(ctx *fiber.Ctx) error {
	ns, err := middleware.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("createApp namespace: %s", ns.Code)

	req := request.CreateAppRequest{}
	if err := ctx.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	app, err := c.appService.Create(ctx.Context(), ns.ID, &req)
	if err != nil {
		return err
	}

	resp := response.NewCreateAppResponse(app)
	log.Debugf("createApp response: %#v", resp)

	return ctx.Status(fiber.StatusCreated).JSON(app)
}

// GetApp handles `GET /apps/:id` endpoint.
func (c Controller) GetApp(ctx *fiber.Ctx) error {
	ns, err := middleware.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("getApp namespace: %s", ns.Code)

	req := request.GetAppRequest{}
	if err := ctx.ParamsParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	app, err := c.appService.Get(ctx.Context(), ns.ID, &req)
	if err != nil {
		return convertError(err)
	}

	resp := response.NewGetAppResponse(app)
	log.Debugf("getApp response: %#v", resp)

	return ctx.JSON(resp)
}

// UpdateApp handles `PUT /apps/:id` endpoint.
func (c Controller) UpdateApp(ctx *fiber.Ctx) error {
	ns, err := middleware.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("updateApp namespace: %s", ns.Code)

	req := request.UpdateAppRequest{}
	if err := ctx.ParamsParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	if err := ctx.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	app, err := c.appService.Update(ctx.Context(), ns.ID, &req)
	if err != nil {
		return convertError(err)
	}

	resp := response.NewUpdateAppResponse(app)
	log.Debugf("updateApp response: %#v", resp)

	return ctx.JSON(resp)
}

// DeleteApp handles `DELETE /apps/:id` endpoint.
func (c Controller) DeleteApp(ctx *fiber.Ctx) error {
	ns, err := middleware.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("deleteApp namespace: %s", ns.Code)

	req := request.DeleteAppRequest{}
	if err := ctx.ParamsParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	if err := c.appService.Delete(ctx.Context(), ns.ID, &req); err != nil {
		return convertError(err)
	}

	return ctx.Status(http.StatusOK).JSON(nil)
}
