package run

import (
	"encoding/json"

	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/api/aim2/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/dto"
)

// ConvertRunMetricsRequestToMetricKeysMapDTO converts request of `GET /runs/:id/metric/get-batch`
// endpoint to internal DTO object.
func ConvertRunMetricsRequestToMetricKeysMapDTO(req *request.GetRunMetricsRequest) (dto.MetricKeysMapDTO, error) {
	// collect unique metrics. uniqueness provides metricKeysMap + metric struct.
	metricKeysMap := make(map[dto.MetricKeysItemDTO]any, len(*req))
	for _, m := range *req {
		if m.Context != nil {
			serializedContext, err := json.Marshal(m.Context)
			if err != nil {
				return nil, eris.Wrap(err, "error marshaling metric context")
			}
			metricKeysMap[dto.MetricKeysItemDTO{
				Name:    m.Name,
				Context: string(serializedContext),
			}] = nil
		}
	}
	return metricKeysMap, nil
}
