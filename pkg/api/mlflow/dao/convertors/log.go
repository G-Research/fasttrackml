package convertors

// TODO:DSuhinin not fully sure about naming of this file. Any suggestions?

import (
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/models"
)

// ConvertLogParamRequestToDBModel converts request.LogParamRequest into actual models.Param model.
func ConvertLogParamRequestToDBModel(runID string, req *request.LogParamRequest) *models.Param {
	return &models.Param{
		Key:   req.Key,
		Value: req.Value,
		RunID: runID,
	}
}

// ConvertLogBatchRequestToDBModel converts request.LogBatchRequest into actual []models.Param, []models.Tag models.
func ConvertLogBatchRequestToDBModel(
	runID string, req *request.LogBatchRequest,
) ([]models.Param, []models.Tag) {
	params := make([]models.Param, len(req.Params))
	for i, param := range req.Params {
		params[i] = models.Param{
			Key:   param.Key,
			Value: param.Value,
			RunID: runID,
		}
	}

	tags := make([]models.Tag, len(req.Tags))
	for i, tag := range req.Tags {
		tags[i] = models.Tag{
			Key:   tag.Key,
			Value: tag.Value,
			RunID: runID,
		}
	}
	return params, tags
}
