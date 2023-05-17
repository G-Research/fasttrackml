package convertors

import (
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

// ConvertSetRunTagRequestToDBModel converts request.SetRunTagRequest into actual models.Tag model.
func ConvertSetRunTagRequestToDBModel(runID string, req *request.SetRunTagRequest) *models.Tag {
	return &models.Tag{
		Key:   req.Key,
		Value: req.Value,
		RunID: runID,
	}
}
