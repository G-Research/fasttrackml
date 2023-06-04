package repositories

import (
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/models"
	"github.com/google/uuid"
)

type AppRepositoryProvider interface {
	List() ([]models.App, error)
	Create(a *models.App) error
	GetByID(id uuid.UUID) (*models.App, error)
	Update(app *models.App, updateData *models.App) error
	Delete(app *models.App) error
}

type AppRepository struct {
	db *gorm.DB
}

func NewAppRepository(db *gorm.DB) *AppRepository {
	return &AppRepository{
		db: db,
	}
}

func (svc *AppRepository) List() ([]models.App, error) {
	apps := []models.App{}
	err := svc.db.Find(&apps).Error
	return apps, err
}

func (svc *AppRepository) Create(a *models.App) error {
	err := svc.db.Create(a).Error
	return err
}

func (svc *AppRepository) GetByID(id uuid.UUID) (*models.App, error) {
	app := &models.App{}
	err := svc.db.Where("NOT is_archived").First(app, id).Error
	return app, err
}

func (svc *AppRepository) Update(app *models.App, updateData *models.App) error {
	err := svc.db.Model(app).Updates(updateData).Error
	return err
}

func (svc *AppRepository) Delete(app *models.App) error {
	err := svc.db.Model(app).Update("IsArchived", true).Error
	return err
}
