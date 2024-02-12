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

	var dashboards []database.Dashboard
	if err := database.DB.
		InnerJoins(
			"App",
			database.DB.Select(
				"ID", "Type",
			).Where(
				&database.App{
					NamespaceID: ns.ID,
				},
				"NamespaceID",
			),
		).
		Where("NOT dashboards.is_archived").
		Order(clause.OrderByColumn{
			Column: clause.Column{
				Table: "App",
				Name:  "updated_at",
			},
			Desc: true,
		}).
		Find(&dashboards).
		Error; err != nil {
		return fmt.Errorf("error fetching dashboards: %w", err)
	}

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

	app := database.App{
		Base: database.Base{
			ID: req.AppID,
		},
		NamespaceID: ns.ID,
	}
	if err := database.DB.
		Select("ID", "Type").
		Where("NOT is_archived").
		First(&app).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.ErrNotFound
		}
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("unable to find app %q: %s", req.AppID, err))
	}

	dash := database.Dashboard{
		AppID:       &req.AppID,
		App:         app,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := database.DB.
		Omit("App").
		Create(&dash).
		Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("error inserting dashboard: %s", err))
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
	if err := ctx.ParamsParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	dashboard := database.Dashboard{
		Base: database.Base{
			ID: req.ID,
		},
	}
	if err := database.DB.
		InnerJoins(
			"App",
			database.DB.Select(
				"ID", "Type",
			).Where(
				&database.App{
					NamespaceID: ns.ID,
				},
				"NamespaceID",
			),
		).
		Where("NOT dashboards.is_archived").
		First(&dashboard).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.ErrNotFound
		}
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("unable to find dashboard %q: %s", req.ID, err))
	}

	return ctx.JSON(dashboard)
}

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

	dash := database.Dashboard{
		Base: database.Base{
			ID: req.ID,
		},
	}
	if err := database.DB.
		InnerJoins(
			"App",
			database.DB.Select(
				"ID", "Type",
			).Where(
				&database.App{
					NamespaceID: ns.ID,
				},
				"NamespaceID",
			),
		).
		Where("NOT dashboards.is_archived").
		First(&dash).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.ErrNotFound
		}
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("unable to find dashboard %q: %s", req.ID, err))
	}

	if err := database.DB.
		Omit("App").
		Model(&dash).
		Updates(database.Dashboard{
			Name:        req.Name,
			Description: req.Description,
		}).
		Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("error updating dashboard %q: %s", req.ID, err))
	}

	return ctx.JSON(dash)
}

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

	dash := database.Dashboard{
		Base: database.Base{
			ID: req.ID,
		},
	}
	if err := database.DB.
		Select("dashboards.id").
		InnerJoins(
			"App",
			database.DB.Select(
				"ID", "Type",
			).Where(
				&database.App{
					NamespaceID: ns.ID,
				},
				"NamespaceID",
			),
		).
		Where("NOT dashboards.is_archived").
		First(&dash).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.ErrNotFound
		}
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("unable to find app %q: %s", req.ID, err))
	}

	if err := database.DB.
		Omit("App").
		Model(&dash).
		Update("IsArchived", true).
		Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("unable to delete app %q: %s", req.ID, err))
	}

	return ctx.Status(200).JSON(nil)
}
