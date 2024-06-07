package convertors

// TODO:DSuhinin not fully sure about naming of this file. Any suggestions?

import (
	"encoding/json"
	"math"
	"time"

	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

// ConvertLogParamRequestToDBModel converts request.LogParamRequest into actual models.Param model.
func ConvertLogParamRequestToDBModel(runID string, req *request.LogParamRequest) *models.Param {
	return &models.Param{
		Key:        req.Key,
		RunID:      runID,
		ValueInt:   req.ValueInt,
		ValueFloat: req.ValueFloat,
		ValueStr:   req.ValueStr,
	}
}

// ConvertLogOutputRequestToDBModel converts request.LogOutRequest into actual models.Log model.
func ConvertLogOutputRequestToDBModel(runID string, req *request.LogOutputRequest) *models.Log {
	return &models.Log{
		RunID:     runID,
		Value:     req.Data,
		Timestamp: time.Now().Unix(),
	}
}

// ConvertLogBatchRequestToDBModel converts request.LogBatchRequest into actual []models.Param, []models.Tag models.
func ConvertLogBatchRequestToDBModel(
	runID string, req *request.LogBatchRequest,
) ([]models.Metric, []models.Param, []models.Tag, error) {
	params := make([]models.Param, len(req.Params))
	for i, param := range req.Params {
		params[i] = models.Param{
			Key:        param.Key,
			RunID:      runID,
			ValueInt:   param.ValueInt,
			ValueFloat: param.ValueFloat,
			ValueStr:   param.ValueStr,
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
			return nil, nil, nil, eris.Errorf("invalid metric value '%v'", metric.Value)
		}
		if metric.Context == nil || len(metric.Context) == 0 {
			m.Context = models.DefaultContext
		} else {
			contextJSON, err := json.Marshal(metric.Context)
			if err != nil {
				return nil, nil, nil, eris.Wrap(err, "error marshalling context")
			}
			m.Context = models.Context{
				Json: contextJSON,
			}
		}
		metrics[n] = m
	}
	return metrics, params, tags, nil
}
