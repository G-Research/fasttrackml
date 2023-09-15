package convertors

// TODO:DSuhinin not fully sure about naming of this file. Any suggestions?

import (
	"math"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"

	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
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
		if v, ok := metric.Value.(float64); ok {
			m.Value = v
		} else if v, ok := metric.Value.(string); ok {
			switch v {
			case "NaN":
				m.Value = 0
				m.IsNan = true
			case "Infinity":
				m.Value = math.MaxFloat64
			case "-Infinity":
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
