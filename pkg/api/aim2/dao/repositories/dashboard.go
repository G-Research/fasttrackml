package repositories

import (
	"context"
	"errors"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/models"
	"github.com/G-Research/fasttrackml/pkg/database"
)

// DashboardRepositoryProvider provides an interface to work with `dashboard` entity.
type DashboardRepositoryProvider interface {
	// Update updates existing models.Dashboard object.
	Update(cxt context.Context, dashboard *models.Dashboard) error
	// Create creates new models.Dashboard object.
	Create(ctx context.Context, dashboard *models.Dashboard) error
	// Delete deletes existing models.Dashboard object.
	Delete(ctx context.Context, dashboard *models.Dashboard) error
	// GetByNamespaceIDAndDashboardID returns models.Dashboard by Dashboard ID.
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
			d.db.Select(
				"ID", "Type",
			).Where(
				&models.App{
					NamespaceID: namespaceID,
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
		return nil, eris.Wrapf(err, "error fetching dashboards")
	}
	return dashboards, nil
}

// GetByNamespaceIDAndDashboardID returns models.Dashboard by Namespace and Dashboard ID.
func (d DashboardRepository) GetByNamespaceIDAndDashboardID(ctx context.Context, namespaceID uint, dashboardID string) (*models.Dashboard, error) {
	var dashboard models.Dashboard
	if err := d.db.WithContext(ctx).
		InnerJoins("App").
		Where(
			"NOT dashboards.is_archived",
		).Where(
		"dashboards.id = ?", dashboardID,
	).Where(
		"app.namespace_id = ?", namespaceID,
	).First(&dashboard).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, eris.Wrapf(err, "error getting dashboard by id: %s", dashboardID)
	}
	return &dashboard, nil
}

// Create creates new models.Dashboard object.
func (d DashboardRepository) Create(ctx context.Context, dashboard *models.Dashboard) error {
	if err := d.db.WithContext(ctx).Create(&dashboard).Error; err != nil {
		return eris.Wrap(err, "error creating dashboard entity")
	}
	return nil
}

// Update updates existing models.Dashboard object.
func (d DashboardRepository) Update(ctx context.Context, dashboard *models.Dashboard) error {
	if err := database.DB.
		Omit("App").
		Model(&dashboard).
		Updates(database.Dashboard{
			Name:        dashboard.Name,
			Description: dashboard.Description,
		}).
		Error; err != nil {
		return eris.Wrap(err, "error updating dashboard entity")
	}
	return nil
}

// Delete deletes a models.Dashboard object.
func (d DashboardRepository) Delete(ctx context.Context, dashboard *models.Dashboard) error {
	if err := database.DB.
		Omit("App").
		Model(&dashboard).
		Update("IsArchived", true).
		Error; err != nil {
		return eris.Wrap(err, "error deleting dashboard entity")
	}
	return nil
}
