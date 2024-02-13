package repositories

import (
	"context"
	"errors"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/models"
)

// AppRepositoryProvider provides an interface to work with `app` entity.
type AppRepositoryProvider interface {
	// Update updates existing database.App object.
	Update(cxt context.Context, app *models.App) error
	// Create creates new database.App object.
	Create(ctx context.Context, app *models.App) error
	// Delete deletes existing database.App object.
	Delete(ctx context.Context, app *models.App) error
	// GetByNamespaceIDAndAppID returns database.App by Namespace and App ID.
	GetByNamespaceIDAndAppID(ctx context.Context, namespaceID uint, appID string) (*models.App, error)
	// GetActiveAppsByNamespace returns the list of active database.App by provided Namespace ID.
	GetActiveAppsByNamespace(ctx context.Context, namespaceID uint) ([]models.App, error)
}

// AppRepository repository to work with `app` entity.
type AppRepository struct {
	db *gorm.DB
}

// NewAppRepository creates repository to work with `app` entity.
func NewAppRepository(db *gorm.DB) *AppRepository {
	return &AppRepository{
		db: db,
	}
}

// Update updates existing database.App object.
func (r AppRepository) Update(ctx context.Context, app *models.App) error {
	if err := r.db.WithContext(ctx).Model(&app).Updates(app).Error; err != nil {
		return eris.Wrapf(err, "error updating app with id: %s", app.ID)
	}
	return nil
}

// Create creates new app object.
func (r AppRepository) Create(ctx context.Context, app *models.App) error {
	if err := r.db.WithContext(ctx).Create(&app).Error; err != nil {
		return eris.Wrap(err, "error creating app entity")
	}
	return nil
}

// GetByNamespaceIDAndAppID returns database.App by Namespace and App ID.
func (r AppRepository) GetByNamespaceIDAndAppID(
	ctx context.Context, namespaceID uint, appID string,
) (*models.App, error) {
	var app models.App
	if err := r.db.WithContext(ctx).Where(
		"NOT is_archived",
	).Where(
		"id = ?", appID,
	).Where(
		"namespace_id = ?", namespaceID,
	).First(&app).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, eris.Wrapf(err, "error getting app by id: %s", appID)
	}
	return &app, nil
}

// GetActiveAppsByNamespace returns the list of active apps by provided Namespace ID.
func (r AppRepository) GetActiveAppsByNamespace(ctx context.Context, namespaceID uint) ([]models.App, error) {
	var apps []models.App
	if err := r.db.WithContext(ctx).Where(
		"NOT is_archived",
	).Where(
		"namespace_id = ?", namespaceID,
	).Find(&apps).Error; err != nil {
		return nil, eris.Wrapf(err, "error getting active apps by namespace id: %d", namespaceID)
	}
	return apps, nil
}

// Delete deletes existing database.App object.
func (r AppRepository) Delete(ctx context.Context, app *models.App) error {
	if err := r.db.WithContext(ctx).Model(app).Update("IsArchived", true).Error; err != nil {
		return eris.Wrapf(err, "error deleting app by id: %s", app.ID)
	}
	return nil
}
