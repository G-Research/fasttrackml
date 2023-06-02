package controller

import (
	"fmt"

	"github.com/G-Research/fasttrackml/pkg/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (ctlr *Controller) GetApps(c *fiber.Ctx) error {
	apps, err := ctlr.appService.GetApps()
	if err != nil {
		return err
	}

	return c.JSON(apps)
}

func (ctlr *Controller) CreateApp(c *fiber.Ctx) error {
	var a struct {
		Type  string
		State models.AppState
	}

	if err := c.BodyParser(&a); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	app := models.App{
		Type:  a.Type,
		State: a.State,
	}

	if err := ctlr.appService.CreateApp(&app); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("error inserting app: %s", err))
	}

	return c.Status(fiber.StatusCreated).JSON(app)
}

func (ctlr *Controller) GetApp(c *fiber.Ctx) error {
	p := struct {
		ID uuid.UUID `params:"id"`
	}{}

	if err := c.ParamsParser(&p); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	app, err := ctlr.appService.GetAppByID(p.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fiber.ErrNotFound
		}
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("unable to find app %q: %s", p.ID, err))
	}

	return c.JSON(app)
}

func (ctlr *Controller) UpdateApp(c *fiber.Ctx) error {
	p := struct {
		ID uuid.UUID `params:"id"`
	}{}

	if err := c.ParamsParser(&p); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	var a struct {
		Type  string
		State models.AppState
	}

	if err := c.BodyParser(&a); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	app, err := ctlr.appService.GetAppByID(p.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fiber.ErrNotFound
		}
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("unable to find app %q: %s", p.ID, err))
	}

	updateData := &models.App{
		Type:  a.Type,
		State: a.State,
	}

	if err := ctlr.appService.UpdateApp(app, updateData); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("error updating app %q: %s", p.ID, err))
	}

	return c.JSON(app)
}

func (ctlr *Controller) DeleteApp(c *fiber.Ctx) error {
	p := struct {
		ID uuid.UUID `params:"id"`
	}{}

	if err := c.ParamsParser(&p); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	app, err := ctlr.appService.GetAppByID(p.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fiber.ErrNotFound
		}
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("unable to find app %q: %s", p.ID, err))
	}

	if err := ctlr.appService.DeleteApp(app); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("unable to delete app %q: %s", p.ID, err))
	}

	return c.Status(fiber.StatusOK).JSON(nil)
}
