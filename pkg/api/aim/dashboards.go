package aim

import (
	"fmt"

	"github.com/G-Research/fasttrack/pkg/database"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/gofiber/fiber/v2"
)

func GetDashboards(c *fiber.Ctx) error {
	var dashboards []database.Dashboard
	if err := database.DB.
		Where("NOT dashboards.is_archived").
		Joins("App", database.DB.Select("ID", "Type", "IsArchived")).
		Order("dashboards.updated_at").
		Find(&dashboards).
		Error; err != nil {
		return fmt.Errorf("error fetching dashboards: %w", err)
	}

	return c.JSON(dashboards)
}

func CreateDashboard(c *fiber.Ctx) error {
	var d struct {
		AppID       uuid.UUID `json:"app_id"`
		Name        string
		Description string
	}

	if err := c.BodyParser(&d); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	app := database.App{
		Base: database.Base{
			ID: d.AppID,
		},
	}
	if err := database.DB.
		Select("ID", "Type").
		First(&app).
		Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fiber.ErrNotFound
		}
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("unable to find app %q: %s", d.AppID, err))
	}

	dash := database.Dashboard{
		AppID:       &d.AppID,
		App:         app,
		Name:        d.Name,
		Description: d.Description,
	}

	if err := database.DB.
		Omit("App").
		Create(&dash).
		Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("error inserting dashboard: %s", err))
	}

	return c.Status(fiber.StatusCreated).JSON(dash)
}

func GetDashboard(c *fiber.Ctx) error {
	p := struct {
		ID uuid.UUID `params:"id"`
	}{}

	if err := c.ParamsParser(&p); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	dash := database.Dashboard{
		Base: database.Base{
			ID: p.ID,
		},
	}
	if err := database.DB.
		Where("NOT dashboards.is_archived").
		Joins("App", database.DB.Select("ID", "Type", "IsArchived")).
		First(&dash).
		Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fiber.ErrNotFound
		}
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("unable to find dashboard %q: %s", p.ID, err))
	}

	return c.JSON(dash)
}

func UpdateDashboard(c *fiber.Ctx) error {
	p := struct {
		ID uuid.UUID `params:"id"`
	}{}

	if err := c.ParamsParser(&p); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	var d struct {
		Name        string
		Description string
	}

	if err := c.BodyParser(&d); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	dash := database.Dashboard{
		Base: database.Base{
			ID: p.ID,
		},
	}
	if err := database.DB.
		Where("NOT is_archived").
		First(&dash).
		Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fiber.ErrNotFound
		}
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("unable to find dashboard %q: %s", p.ID, err))
	}

	if err := database.DB.
		Model(&dash).
		Updates(database.Dashboard{
			Name:        d.Name,
			Description: d.Description,
		}).
		Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("error updating dashboard %q: %s", p.ID, err))
	}

	return c.JSON(dash)
}

func DeleteDashboard(c *fiber.Ctx) error {
	p := struct {
		ID uuid.UUID `params:"id"`
	}{}

	if err := c.ParamsParser(&p); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	dash := database.Dashboard{
		Base: database.Base{
			ID: p.ID,
		},
	}
	if err := database.DB.
		Select("ID").
		Where("NOT is_archived").
		First(&dash).
		Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fiber.ErrNotFound
		}
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("unable to find app %q: %s", p.ID, err))
	}

	if err := database.DB.
		Model(&dash).
		Update("IsArchived", true).
		Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("unable to delete app %q: %s", p.ID, err))
	}

	return c.Status(200).JSON(nil)
}
