package aim

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/common/middleware/namespace"
	"github.com/G-Research/fasttrackml/pkg/database"
)

func GetDashboards(c *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(c.Context())
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
				&database.App{NamespaceID: ns.ID}, "NamespaceID", "IsArchived"),
		).
		Where("NOT dashboards.is_archived").
		Order("dashboards.updated_at").
		Find(&dashboards).
		Error; err != nil {
		return fmt.Errorf("error fetching dashboards: %w", err)
	}

	return c.JSON(dashboards)
}

func CreateDashboard(c *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(c.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("createDashboard namespace: %s", ns.Code)

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
		NamespaceID: ns.ID,
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
	ns, err := namespace.GetNamespaceFromContext(c.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("getDashboard namespace: %s", ns.Code)

	p := struct {
		ID uuid.UUID `params:"id"`
	}{}

	if err := c.ParamsParser(&p); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	dashboard := database.Dashboard{
		Base: database.Base{
			ID: p.ID,
		},
	}
	if err := database.DB.
		InnerJoins(
			"App",
			database.DB.Select(
				"ID", "Type",
			).Where(
				&database.App{NamespaceID: ns.ID}, "NamespaceID", "IsArchived"),
		).
		Where("NOT dashboards.is_archived").
		First(&dashboard).
		Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fiber.ErrNotFound
		}
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("unable to find dashboard %q: %s", p.ID, err))
	}

	return c.JSON(dashboard)
}

func UpdateDashboard(c *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(c.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("updateDashboard namespace: %s", ns.Code)

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
		InnerJoins(
			"App",
			database.DB.Select(
				"ID", "Type",
			).Where(
				&database.App{NamespaceID: ns.ID}, "NamespaceID", "IsArchived"),
		).
		Where("NOT dashboards.is_archived").
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
	ns, err := namespace.GetNamespaceFromContext(c.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("deleteDashboard namespace: %s", ns.Code)

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
		Select("dashboards.id").
		InnerJoins(
			"App",
			database.DB.Select(
				"ID", "Type",
			).Where(
				&database.App{NamespaceID: ns.ID}, "NamespaceID", "IsArchived"),
		).
		Where("NOT dashboards.is_archived").
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
