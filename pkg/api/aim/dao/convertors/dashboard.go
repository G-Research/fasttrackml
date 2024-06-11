package convertors

import (
	"github.com/google/uuid"

	"github.com/G-Research/fasttrackml/pkg/api/aim/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim/dao/models"
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
