package response

import (
	"time"

	"github.com/google/uuid"

	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/models"
)

// App represents the response json in App endpoints
type App struct {
	ID        uuid.UUID `json:"id"`
	Type      string    `json:"type"`
	State     AppState  `json:"state"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AppState represents key/value state data
type AppState map[string]any

// NewGetAppsResponse creates new response object for `GET /apps` endpoint.
func NewGetAppsResponse(apps []models.App) []App {
	resp := make([]App, len(apps))
	for i, app := range apps {
		//nolint:gosec
		resp[i] = NewCreateAppResponse(&app)
	}
	return resp
}

// NewCreateAppResponse creates new response object for `POST /apps` endpoint.
func NewCreateAppResponse(app *models.App) App {
	return App{
		ID:        app.ID,
		Type:      app.Type,
		State:     map[string]any(app.State),
		CreatedAt: app.CreatedAt,
		UpdatedAt: app.UpdatedAt,
	}
}

// NewGetAppResponse creates new response object for `GET /apps/:id` endpoint.
var NewGetAppResponse = NewCreateAppResponse

// NewUpdateAppResponse creates new response object for `PUT /apps/:id` endpoint.
var NewUpdateAppResponse = NewCreateAppResponse
