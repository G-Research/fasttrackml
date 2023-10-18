package convertors

import (
	"math"

	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

// ConvertMetricParamRequestToDBModel converts request.LogMetricRequest into actual models.Metric model.
func ConvertMetricParamRequestToDBModel(runID string, req *request.LogMetricRequest) (*models.Metric, error) {
	metric := models.Metric{
		Key:       req.Key,
		Timestamp: req.Timestamp,
		Step:      req.Step,
		RunID:     runID,
		Context:   req.Context,
	}
	if v, ok := req.Value.(float64); ok {
		metric.Value = v
	} else if v, ok := req.Value.(string); ok {
		switch v {
		case "NaN":
			metric.Value = 0
			metric.IsNan = true
		case "Infinity":
			metric.Value = math.MaxFloat64
		case "-Infinity":
			metric.Value = -math.MaxFloat64
		default:
			return nil, eris.Errorf("invalid metric value '%s'", v)
		}
	} else {
		return nil, eris.Errorf("invalid metric value '%s'", v)
	}
	return &metric, nil
}
