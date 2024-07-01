package run

import (
	"encoding/json"

	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/api/aim/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim/dao/models"
)

// ConvertRunMetricsRequestToMap converts request of
// `GET /runs/:id/metric/get-batch` endpoint to an internal DTO object.
func ConvertRunMetricsRequestToMap(req *request.GetRunMetricsRequest) (models.MetricKeysMap, error) {
	// collect unique metrics. uniqueness provides metricKeysMap + metric struct.
	metricKeysMap := make(map[models.MetricKeysItem]any, len(*req))
	for _, m := range *req {
		if m.Context != nil {
			serializedContext, err := json.Marshal(m.Context)
			if err != nil {
				return nil, eris.Wrap(err, "error marshaling metric context")
			}
			metricKeysMap[models.MetricKeysItem{
				Name:    m.Name,
				Context: string(serializedContext),
			}] = nil
		}
	}
	return metricKeysMap, nil
}
