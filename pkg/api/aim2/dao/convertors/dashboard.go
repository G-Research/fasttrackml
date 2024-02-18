package convertors

import (
	"github.com/G-Research/fasttrackml/pkg/api/aim2/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/models"
	"github.com/google/uuid"
)

// ConvertCreateDashboardRequestToDBModel translates the request to a model.
func ConvertCreateDashboardRequestToDBModel(req request.CreateDashboardRequest) models.Dashboard {
	return models.Dashboard{
		Base:        models.Base{ID: uuid.New()},
		AppID:       &req.AppID,
		Name:        req.Name,
		Description: req.Description,
	}
}
