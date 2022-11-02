package api

import (
	"fasttrack/model"
	"net/http"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func MetricGetHistory(db *gorm.DB) HandlerFunc {
	return EnsureMethod(func(w http.ResponseWriter, r *http.Request) any {
		id := r.URL.Query().Get("run_id")
		key := r.URL.Query().Get("metric_key")

		log.Debugf("MetricGetHistory request: run_id='%s', metric_key='%s'", id, key)

		if id == "" {
			return NewError(ErrorCodeInvalidParameterValue, "Missing value for required parameter 'run_id'")
		}
		if key == "" {
			return NewError(ErrorCodeInvalidParameterValue, "Missing value for required parameter 'metric_key'")
		}

		var metrics []model.Metric
		if tx := db.Where("run_uuid = ?", id).Where("key = ?", key).Find(&metrics); tx.Error != nil {
			return NewError(ErrorCodeInternalError, "Unable to get metric history for metric '%s' of run '%s'", key, id)
		}

		resp := &MetricsGetHistoryResponse{
			Metrics: make([]Metric, len(metrics)),
		}
		for n, m := range metrics {

			resp.Metrics[n] = Metric{
				Key:       m.Key,
				Value:     m.Value,
				Timestamp: m.Timestamp,
				Step:      m.Step,
			}
			if m.IsNan {
				resp.Metrics[n].Value = "NaN"
			}
		}

		log.Debugf("MetricGetHistory response: %#v", resp)

		return resp
	},
		http.MethodGet,
	)
}
