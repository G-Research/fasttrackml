package service

import (
	"github.com/G-Research/fasttrackml/pkg/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AppService struct {
	DB *gorm.DB
}

func NewAppService(db *gorm.DB) *AppService {
	return &AppService{
		DB: db,
	}
}

func (svc *AppService) GetApps() ([]models.App, error) {
	apps := []models.App{}
	err := svc.DB.Find(&apps).Error
	return apps, err
}

func (svc *AppService) CreateApp(a *models.App) error {
	err := svc.DB.Create(a).Error
	return err
}

func (svc *AppService) GetAppByID(id uuid.UUID) (*models.App, error) {
	app := &models.App{}
	err := svc.DB.Where("NOT is_archived").First(app, id).Error
	return app, err
}

func (svc *AppService) UpdateApp(app *models.App, updateData *models.App) error {
	err := svc.DB.Model(app).Updates(updateData).Error
	return err
}

func (svc *AppService) DeleteApp(app *models.App) error {
	err := svc.DB.Model(app).Update("IsArchived", true).Error
	return err
}
