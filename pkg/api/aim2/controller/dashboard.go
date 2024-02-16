package controller

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/G-Research/fasttrackml/pkg/api/aim2/api/request"
	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/pkg/common/middleware/namespace"
	"github.com/G-Research/fasttrackml/pkg/database"
)

func (c Controller) GetDashboards(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("getDashboards namespace: %s", ns.Code)


	return ctx.JSON(dashboards)
}

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


	return ctx.Status(fiber.StatusCreated).JSON(dash)
}

func (c Controller) GetDashboard(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("getDashboard namespace: %s", ns.Code)

	req := request.GetDashboardRequest{}

	return ctx.JSON(dashboard)
}

func (c Controller) UpdateDashboard(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("updateDashboard namespace: %s", ns.Code)

	req := request.UpdateDashboardRequest{}

	return ctx.JSON(dash)
}

func (c Controller) DeleteDashboard(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("deleteDashboard namespace: %s", ns.Code)

	req := request.DeleteDashboardRequest{}

	return ctx.Status(200).JSON(nil)
}
