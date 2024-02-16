package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/models"
	"github.com/G-Research/fasttrackml/pkg/database"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DashboardRepositoryProvider provides an interface to work with `dashboard` entity.
type DashboardRepositoryProvider interface {
	// Update updates existing models.Dashboard object.
	Update(cxt context.Context, dashboard *models.Dashboard) error
	// Create creates new models.Dashboard object.
	Create(ctx context.Context, dashboard *models.Dashboard) error
	// Delete deletes existing models.Dashboard object.
	Delete(ctx context.Context, dashboard *models.Dashboard) error
	// GetByNamespaceIDAndDashboardID returns models.Dashboard by Namespace and Dashboard ID.
	GetByNamespaceIDAndDashboardID(ctx context.Context, namespaceID uint, dashboardID string) (*models.Dashboard, error)
	// GetDashboardsByNamespace returns the list of active models.Dashboard by provided Namespace ID.
	GetDashboardsByNamespace(ctx context.Context, namespaceID uint) ([]models.Dashboard, error)
}

// DashboardRepository repository to work with `dashboard` entity.
type DashboardRepository struct {
	db *gorm.DB
}

// NewDashboardRepository creates repository to work with `dashboard` entity.
func NewDashboardRepository(db *gorm.DB) *DashboardRepository {
	return &DashboardRepository{
		db: db,
	}
}

// GetDashboardsByNamespace returns the list of active models.Dashboard by provided Namespace ID.
func (d DashboardRepository) GetDashboardsByNamespace(ctx context.Context, namespaceID uint) ([]models.Dashboard, error) {
	var dashboards []models.Dashboard
	if err := d.db.
		InnerJoins(
			"App",
			database.DB.Select(
				"ID", "Type",
			).Where(
				&models.App{
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
		return nil, fmt.Errorf("error fetching dashboards: %w", err)
	}
	return dashboards, nil
}

// GetByNamespaceIDAndDashboardID returns models.Dashboard by Namespace and Dashboard ID.
func (d DashboardRepository) GetByNamespaceIDAndDashboardID(ctx context.Context, nsID uint, dashboardID string) (models.Dashboard, error) {
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
}

// Create creates new models.Dashboard object.
func (d DashboardRepository) Create(ctx context.Context, dashboard *models.Dashboard) error {
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
}

// Update updates existing models.Dashboard object.
func (d DashboardRepository) Update(ctx context.Context, dashboard *models.Dashboard) error {
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
}

// Delete deletes a models.Dashboard object.
func (d DashboardRepository) Delete(ctx context.Context, dashboard *models.Dashboard) error {
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
}


