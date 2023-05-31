package convertors

import (
	"github.com/G-Research/fasttrackml/pkg/models"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
)

// ConvertSetRunTagRequestToDBModel converts request.SetRunTagRequest into actual models.Tag model.
func ConvertSetRunTagRequestToDBModel(runID string, req *request.SetRunTagRequest) *models.Tag {
	return &models.Tag{
		Key:   req.Key,
		Value: req.Value,
		RunID: runID,
	}
}

// ConvertSetExperimentTagRequestToDBModel converts request.SetExperimentTagRequest into actual models.ExperimentTag model.
func ConvertSetExperimentTagRequestToDBModel(
	experimentID int32, req *request.SetExperimentTagRequest,
) *models.ExperimentTag {
	return &models.ExperimentTag{
		Key:          req.Key,
		Value:        req.Value,
		ExperimentID: experimentID,
	}
}
