package aim

import (
	"fmt"

	"github.com/G-Resarch/fasttrack/pkg/database"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func GetApps(c *fiber.Ctx) error {
	var apps []database.App
	if err := database.DB.
		Where("NOT is_archived").
		Find(&apps).
		Error; err != nil {
		return fmt.Errorf("error fetching apps: %w", err)
	}

	return c.JSON(apps)
}

func CreateApp(c *fiber.Ctx) error {
	var a struct {
		Type  string
		State database.AppState
	}

	if err := c.BodyParser(&a); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	app := database.App{
		Type:  a.Type,
		State: a.State,
	}

	if err := database.DB.
		Create(&app).
		Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("error inserting app: %s", err))
	}

	return c.Status(fiber.StatusCreated).JSON(app)
}

func GetApp(c *fiber.Ctx) error {
	p := struct {
		ID uuid.UUID `params:"id"`
	}{}

	if err := c.ParamsParser(&p); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	app := database.App{
		Base: database.Base{
			ID: p.ID,
		},
	}
	if err := database.DB.
		Where("NOT is_archived").
		First(&app).
		Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fiber.ErrNotFound
		}
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("unable to find app %q: %s", p.ID, err))
	}

	return c.JSON(app)
}

func UpdateApp(c *fiber.Ctx) error {
	p := struct {
		ID uuid.UUID `params:"id"`
	}{}

	if err := c.ParamsParser(&p); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	var a struct {
		Type  string
		State database.AppState
	}

	if err := c.BodyParser(&a); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	app := database.App{
		Base: database.Base{
			ID: p.ID,
		},
	}
	if err := database.DB.
		Where("NOT is_archived").
		First(&app).
		Error; err != nil {
		if err == gorm.ErrRecordNotFound {
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

	return c.JSON(app)
}

func DeleteApp(c *fiber.Ctx) error {
	p := struct {
		ID uuid.UUID `params:"id"`
	}{}

	if err := c.ParamsParser(&p); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	app := database.App{
		Base: database.Base{
			ID: p.ID,
		},
	}
	if err := database.DB.
		Select("ID").
		Where("NOT is_archived").
		First(&app).
		Error; err != nil {
		if err == gorm.ErrRecordNotFound {
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

	return c.Status(200).JSON(nil)
}
