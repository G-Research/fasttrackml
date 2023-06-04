package service

import (
	"github.com/G-Research/fasttrackml/pkg/models"
	"github.com/G-Research/fasttrackml/pkg/repositories"

	"github.com/google/uuid"
)

type AppService struct {
	appRepository repositories.AppRepositoryProvider
}

func NewAppService(appRepo repositories.AppRepositoryProvider) *AppService {
	return &AppService{
		appRepository: appRepo,
	}
}

func (svc *AppService) GetApps() ([]models.App, error) {
	return svc.appRepository.List()
}

func (svc *AppService) CreateApp(a *models.App) error {
	return svc.appRepository.Create(a)
}

func (svc *AppService) GetAppByID(id uuid.UUID) (*models.App, error) {
	return svc.appRepository.GetByID(id)
}

func (svc *AppService) UpdateApp(app *models.App, updateData *models.App) error {
	return svc.appRepository.Update(app, updateData)
}

func (svc *AppService) DeleteApp(app *models.App) error {
	return svc.appRepository.Delete(app)
}
