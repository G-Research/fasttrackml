package convertors

import (
	"github.com/G-Research/fasttrackml/pkg/api/aim2/api/request"
	aimModels "github.com/G-Research/fasttrackml/pkg/api/aim2/dao/models"
	mlflowModels "github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

// ConvertCreateAppRequestToDBModel converts request.CreateAppRequest into actual models.App model.
func ConvertCreateAppRequestToDBModel(
	namespace *mlflowModels.Namespace, req *request.CreateAppRequest,
) *aimModels.App {
	return &aimModels.App{
		Type:        req.Type,
		State:       aimModels.AppState(req.State),
		NamespaceID: namespace.ID,
	}
}
