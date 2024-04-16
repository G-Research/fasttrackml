package response

import (
	"time"

	"github.com/google/uuid"

	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/models"
)

// Dashboard represents the response json in Dashboard endpoints
type Dashboard struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	AppID       uuid.UUID `json:"app_id"`
	AppType     string    `json:"app_type"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewGetDashboardsResponse creates new response object for `GET /dashboards` endpoint.
func NewGetDashboardsResponse(dashboards []models.Dashboard) []Dashboard {
	resp := make([]Dashboard, len(dashboards))
	for i, app := range dashboards {
		//nolint:gosec
		resp[i] = NewCreateDashboardResponse(&app)
	}
	return resp
}

// NewCreateDashboardResponse creates new response object for `POST /dashboards` endpoint.
func NewCreateDashboardResponse(dashboard *models.Dashboard) Dashboard {
	return Dashboard{
		ID:          dashboard.ID,
		Name:        dashboard.Name,
		Description: dashboard.Description,
		AppID:       *dashboard.AppID,
		AppType:     dashboard.App.Type,
		CreatedAt:   dashboard.CreatedAt,
		UpdatedAt:   dashboard.UpdatedAt,
	}
}

// NewGetDashboardResponse creates new response object for `GET /apps/:id` endpoint.
var NewGetDashboardResponse = NewCreateDashboardResponse

// NewUpdateDashboardResponse creates new response object for `PUT /apps/:id` endpoint.
var NewUpdateDashboardResponse = NewCreateDashboardResponse
