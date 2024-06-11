package dashboard

import (
	"context"

	"github.com/G-Research/fasttrackml/pkg/api/aim/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim/dao/convertors"
	"github.com/G-Research/fasttrackml/pkg/api/aim/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/aim/dao/repositories"
	"github.com/G-Research/fasttrackml/pkg/common/api"
)

// Service provides service layer to work with `dashboard` business logic.
type Service struct {
	appRepository       repositories.AppRepositoryProvider
	dashboardRepository repositories.DashboardRepositoryProvider
}

// NewService creates new Service instance.
func NewService(
	dashboardRepo repositories.DashboardRepositoryProvider,
	appRepo repositories.AppRepositoryProvider,
) *Service {
	return &Service{
		appRepository:       appRepo,
		dashboardRepository: dashboardRepo,
	}
}

// Get returns dashboard object.
func (s Service) Get(
	ctx context.Context, namespaceID uint, req *request.GetDashboardRequest,
) (*models.Dashboard, error) {
	dashboard, err := s.dashboardRepository.GetByNamespaceIDAndDashboardID(ctx, namespaceID, req.ID.String())
	if err != nil {
		return nil, api.NewInternalError("unable to find dashboard by id %q: %s", req.ID, err)
	}
	if dashboard == nil {
		return nil, api.NewResourceDoesNotExistError("dashboard '%s' not found", req.ID)
	}
	return dashboard, nil
}

// Create creates new dashboard object.
func (s Service) Create(
	ctx context.Context, namespaceID uint, req *request.CreateDashboardRequest,
) (*models.Dashboard, error) {
	app, err := s.appRepository.GetByNamespaceIDAndAppID(ctx, namespaceID, req.AppID.String())
	if err != nil {
		return nil, api.NewInternalError("unable to find app %q for dashboard: %s", req.AppID, err)
	}
	if app == nil || app.IsArchived {
		return nil, api.NewResourceDoesNotExistError("app with id '%s' not found", req.AppID)
	}
	dashboard := convertors.ConvertCreateDashboardRequestToDBModel(*req)
	dashboard.App = *app
	if err := s.dashboardRepository.Create(ctx, &dashboard); err != nil {
		return nil, api.NewInternalError("unable to create dashboard: %v", err)
	}
	return &dashboard, nil
}

// Update updates existing dashboard object.
func (s Service) Update(
	ctx context.Context, namespaceID uint, req *request.UpdateDashboardRequest,
) (*models.Dashboard, error) {
	dashboard, err := s.dashboardRepository.GetByNamespaceIDAndDashboardID(ctx, namespaceID, req.ID.String())
	if err != nil {
		return nil, api.NewInternalError("unable to find dashboard by id %s: %s", req.ID, err)
	}
	if dashboard == nil {
		return nil, api.NewResourceDoesNotExistError("dashboard with id '%s' not found", req.ID)
	}

	dashboard.Name = req.Name
	dashboard.Description = req.Description

	if err := s.dashboardRepository.Update(ctx, dashboard); err != nil {
		return nil, api.NewInternalError("unable to update dashboard '%s': %s", dashboard.ID, err)
	}
	return dashboard, nil
}

// GetDashboards returns the list of active dashboards.
func (s Service) GetDashboards(ctx context.Context, namespaceID uint) ([]models.Dashboard, error) {
	dashboards, err := s.dashboardRepository.GetDashboardsByNamespace(ctx, namespaceID)
	if err != nil {
		return nil, api.NewInternalError("unable to get active dashboards: %v", err)
	}
	return dashboards, nil
}

// Delete deletes existing object.
func (s Service) Delete(ctx context.Context, namespaceID uint, req *request.DeleteDashboardRequest) error {
	dashboard, err := s.dashboardRepository.GetByNamespaceIDAndDashboardID(ctx, namespaceID, req.ID.String())
	if err != nil {
		return api.NewInternalError("error trying to find dashboard by id %s: %s", req.ID, err)
	}
	if dashboard == nil {
		return api.NewResourceDoesNotExistError("dashboard with id '%s' not found", req.ID)
	}
	if err := s.dashboardRepository.Delete(ctx, dashboard); err != nil {
		return api.NewInternalError("unable to delete dashboard by id %s: %s", req.ID, err)
	}
	return nil
}
