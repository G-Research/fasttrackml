package controller

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/pkg/common/middleware/namespace"
	"github.com/G-Research/fasttrackml/pkg/database"
)

func (c Controller) GetApps(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("getApps namespace: %s", ns.Code)

	var apps []database.App
	if err := database.DB.
		Where("NOT is_archived").
		Where("namespace_id = ?", ns.ID).
		Find(&apps).
		Error; err != nil {
		return fmt.Errorf("error fetching apps: %w", err)
	}

	return ctx.JSON(apps)
}

func (c Controller) CreateApp(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("createApp namespace: %s", ns.Code)

	var a struct {
		Type  string
		State database.AppState
	}

	if err := ctx.BodyParser(&a); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	app := database.App{
		Type:        a.Type,
		State:       a.State,
		NamespaceID: ns.ID,
	}

	if err := database.DB.
		Create(&app).
		Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("error inserting app: %s", err))
	}

	return ctx.Status(fiber.StatusCreated).JSON(app)
}

func (c Controller) GetApp(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("getApp namespace: %s", ns.Code)

	p := struct {
		ID uuid.UUID `params:"id"`
	}{}

	if err := ctx.ParamsParser(&p); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	app := database.App{
		Base: database.Base{
			ID: p.ID,
		},
		NamespaceID: ns.ID,
	}
	if err := database.DB.
		Where("NOT is_archived").
		Where("namespace_id = ?", ns.ID).
		First(&app).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.ErrNotFound
		}
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("unable to find app %q: %s", p.ID, err))
	}

	return ctx.JSON(app)
}

func (c Controller) UpdateApp(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("updateApp namespace: %s", ns.Code)

	p := struct {
		ID uuid.UUID `params:"id"`
	}{}

	if err := ctx.ParamsParser(&p); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	var a struct {
		Type  string
		State database.AppState
	}

	if err := ctx.BodyParser(&a); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	app := database.App{
		Base: database.Base{
			ID: p.ID,
		},
		NamespaceID: ns.ID,
	}
	if err := database.DB.
		Where("NOT is_archived").
		Where("namespace_id = ?", ns.ID).
		First(&app).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.ErrNotFound
		}
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("unable to find app %q: %s", p.ID, err))
	}

	if err := database.DB.
		Model(&app).
		Updates(database.App{
			Type:  a.Type,
			State: a.State,
		}).
		Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("error updating app %q: %s", p.ID, err))
	}

	return ctx.JSON(app)
}

func (c Controller) DeleteApp(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("deleteApp namespace: %s", ns.Code)

	p := struct {
		ID uuid.UUID `params:"id"`
	}{}

	if err := ctx.ParamsParser(&p); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	app := database.App{
		Base: database.Base{
			ID: p.ID,
		},
		NamespaceID: ns.ID,
	}
	if err := database.DB.
		Select("ID").
		Where("NOT is_archived").
		Where("namespace_id = ?", ns.ID).
		First(&app).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.ErrNotFound
		}
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("unable to find app %q: %s", p.ID, err))
	}

	if err := database.DB.
		Model(&app).
		Update("IsArchived", true).
		Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("unable to delete app %q: %s", p.ID, err))
	}

	return ctx.Status(http.StatusOK).JSON(nil)
}
