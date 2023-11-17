package convertors

import (
	"encoding/json"
	"math"

	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

// ConvertMetricParamRequestToDBModel converts request.LogMetricRequest into actual models.Metric model.
func ConvertMetricParamRequestToDBModel(runID string, context *models.Context, req *request.LogMetricRequest) (*models.Metric, error) {
	metric := models.Metric{
		Key:       req.Key,
		Timestamp: req.Timestamp,
		Step:      req.Step,
		RunID:     runID,
		Context:   context,
		ContentID: &context.ID,
	}
	if v, ok := req.Value.(float64); ok {
		metric.Value = v
	} else if v, ok := req.Value.(string); ok {
		switch v {
		case common.NANValue:
			metric.Value = 0
			metric.IsNan = true
		case common.NANPositiveInfinity:
			metric.Value = math.MaxFloat64
		case common.NANNegativeInfinity:
			metric.Value = -math.MaxFloat64
		default:
			return nil, eris.Errorf("invalid metric value '%s'", v)
		}
	} else {
		return nil, eris.Errorf("invalid metric value '%s'", v)
	}
	return &metric, nil
}

// ConvertMetricContextParamRequestToDBModel converts request.LogMetricRequest into actual models.Context model.
func ConvertMetricContextParamRequestToDBModel(req *request.LogMetricRequest) (*models.Context, error) {
	contextJSON, err := json.Marshal(req.Context)
	if err != nil {
		return nil, err
	}

	return &models.Context{
		Json: contextJSON,
	}, nil
}
