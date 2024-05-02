package convertors

import (
	"github.com/google/uuid"

	"github.com/G-Research/fasttrackml/pkg/api/aim/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim/dao/models"
)

// ConvertCreateAppRequestToDBModel converts request.CreateAppRequest into actual models.App model.
func ConvertCreateAppRequestToDBModel(
	namespaceID uint, req *request.CreateAppRequest,
) *models.App {
	return &models.App{
		Base:        models.Base{ID: uuid.New()},
		Type:        req.Type,
		State:       models.AppState(req.State),
		NamespaceID: namespaceID,
	}
}
