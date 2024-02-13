package convertors

import (
	"github.com/G-Research/fasttrackml/pkg/api/aim2/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/database"
)

// ConvertCreateAppRequestToDBModel converts request.CreateAppRequest into actual models.App model.
func ConvertCreateAppRequestToDBModel(
	namespace *models.Namespace, req *request.CreateAppRequest,
) *database.App {
	return &database.App{
		Type:        req.Type,
		State:       database.AppState(req.State),
		NamespaceID: namespace.ID,
	}
}
