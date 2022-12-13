package api

import (
	"encoding/base64"
	"encoding/json"
	"fasttrack/model"
	"fmt"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func MetricGetHistory(db *gorm.DB) HandlerFunc {
	return EnsureMethod(func(w http.ResponseWriter, r *http.Request) any {
		id := r.URL.Query().Get("run_id")
		if id == "" {
			id = r.URL.Query().Get("run_uuid")
		}
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

func MetricsGetHistories(db *gorm.DB) HandlerFunc {
	return EnsureJson(EnsureMethod(func(w http.ResponseWriter, r *http.Request) any {
		var req MetricsGetHistoriesRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return NewError(ErrorCodeBadRequest, "Unable to decode request body: %s", err)
		}

		log.Debugf("MetricsGetHistories request: %#v", req)

		if len(req.ExperimentIDs) > 0 && len(req.RunIDs) > 0 {
			return NewError(ErrorCodeInvalidParameterValue, "experiment_ids and run_ids cannot both be specified at the same time")
		}

		// MaxResults
		limit := int(req.MaxResults)
		if limit == 0 {
			limit = 1000
		} else if limit > 1000000 {
			return NewError(ErrorCodeInvalidParameterValue, "Invalid value for parameter 'max_results' supplied.")
		}
		tx := db.Limit(limit + 1)

		// PageToken
		var offset int
		if req.PageToken != "" {
			var token PageToken
			if err := json.NewDecoder(
				base64.NewDecoder(
					base64.StdEncoding,
					strings.NewReader(req.PageToken),
				),
			).Decode(&token); err != nil {
				return NewError(ErrorCodeInvalidParameterValue, "Invalid page_token '%s': %s", req.PageToken, err)

			}
			offset = int(token.Offset)
		}
		tx.Offset(offset)

		// Default order
		tx.Order("runs.start_time DESC")
		tx.Order("runs.run_uuid")

		// Filter by experiments
		if len(req.ExperimentIDs) > 0 {
			tx.Where("experiment_id IN ?", req.ExperimentIDs)

			// ViewType
			var lifecyleStages []model.LifecycleStage
			switch req.ViewType {
			case ViewTypeActiveOnly, "":
				lifecyleStages = []model.LifecycleStage{
					model.LifecycleStageActive,
				}
			case ViewTypeDeletedOnly:
				lifecyleStages = []model.LifecycleStage{
					model.LifecycleStageDeleted,
				}
			case ViewTypeAll:
				lifecyleStages = []model.LifecycleStage{
					model.LifecycleStageActive,
					model.LifecycleStageDeleted,
				}
			default:
				return NewError(ErrorCodeInvalidParameterValue, "Invalid run_view_type '%s'", req.ViewType)
			}
			tx.Where("lifecycle_stage IN ?", lifecyleStages)
		}

		// Filter by runs
		if len(req.RunIDs) > 0 {
			tx.Where("run_uuid IN ?", req.RunIDs)
		}

		// Filter by metric keys
		order := func(db *gorm.DB) *gorm.DB {
			return db.Order("metrics.key").
				Order("metrics.step").
				Order("metrics.timestamp").
				Order("metrics.value")

		}
		if len(req.MetricKeys) > 0 {
			tx.Preload("Metrics", "key IN ?", req.MetricKeys, order)
		} else {
			tx.Preload("Metrics", order)
		}

		// Actual query
		runs := []model.Run{}
		tx.Find(&runs)
		if tx.Error != nil {
			return NewError(ErrorCodeInternalError, "Unable to search runs: %s", tx.Error)
		}

		resp := &MetricsGetHistoriesResponse{}

		// NextPageToken
		if len(runs) > limit {
			runs = runs[:limit]
			var token strings.Builder
			b64 := base64.NewEncoder(base64.StdEncoding, &token)
			if err := json.NewEncoder(b64).Encode(PageToken{
				Offset: int32(offset + limit),
			}); err != nil {
				return NewError(ErrorCodeInternalError, "Unable to build next_page_token: %s", err)
			}
			b64.Close()
			resp.NextPageToken = token.String()
		}

		resp.Runs = make([]Run, len(runs))
		for i, r := range runs {
			run := Run{
				Info: RunInfo{
					ID:             r.ID,
					UUID:           r.ID,
					Name:           r.Name,
					ExperimentID:   fmt.Sprint(r.ExperimentID),
					UserID:         r.UserID,
					Status:         RunStatus(r.Status),
					StartTime:      r.StartTime.Int64,
					EndTime:        r.EndTime.Int64,
					ArtifactURI:    r.ArtifactURI,
					LifecycleStage: LifecycleStage(r.LifecycleStage),
				},
				Data: RunData{
					Metrics: make([]Metric, len(r.Metrics)),
				},
			}
			for j, m := range r.Metrics {
				metric := Metric{
					Key:       m.Key,
					Value:     m.Value,
					Timestamp: m.Timestamp,
					Step:      m.Step,
				}
				if m.IsNan {
					metric.Value = "NaN"
				}
				run.Data.Metrics[j] = metric
			}
			resp.Runs[i] = run
		}

		log.Debugf("MetricsGetHistories response: %#v", resp)

		return resp
	},
		http.MethodPost,
	))
}
