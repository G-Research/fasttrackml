package convertors

// TODO:DSuhinin not fully sure about naming of this file. Any suggestions?

import (
	"encoding/json"
	"math"

	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
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
	contexts []*models.Context,
) ([]models.Metric, []models.Param, []models.Tag, error) {
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

	metrics := make([]models.Metric, len(req.Metrics))
	for n, metric := range req.Metrics {
		m := models.Metric{
			Key:       metric.Key,
			Timestamp: metric.Timestamp,
			Step:      metric.Step,
			RunID:     runID,
		}
		if contexts[n] != nil {
			m.Context = contexts[n]
			m.ContextID = &contexts[n].ID
		}
		if v, ok := metric.Value.(float64); ok {
			m.Value = v
		} else if v, ok := metric.Value.(string); ok {
			switch v {
			case common.NANValue:
				m.Value = 0
				m.IsNan = true
			case common.NANPositiveInfinity:
				m.Value = math.MaxFloat64
			case common.NANNegativeInfinity:
				m.Value = -math.MaxFloat64
			default:
				return nil, nil, nil, eris.Errorf("invalid metric value '%s'", v)
			}
		} else {
			return nil, nil, nil, eris.Errorf("invalid metric value '%s'", v)
		}
		metrics[n] = m
	}
	return metrics, params, tags, nil
}

// ConvertLogBatchRequestToContextDBModel converts request.LogBatchRequest into []*models.Context model.
func ConvertLogBatchRequestToContextDBModel(req *request.LogBatchRequest) ([]*models.Context, error) {
	contexts := make([]*models.Context, len(req.Metrics))
	for n, metric := range req.Metrics {
		if metric.Context != nil {
			contextJSON, err := json.Marshal(metric.Context)
			if err != nil {
				return nil, err
			}
			contexts[n] = &models.Context{
				Json: contextJSON,
			}
		} else {
			contexts[n] = nil
		}
	}
	return contexts, nil
}
