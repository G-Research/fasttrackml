package request

// UpdateRunRequest is a request struct for `PUT /runs/:id` endpoint.
type UpdateRunRequest struct {
	RunID       *string `json:"run_id"`
	RunUUID     *string `json:"run_uuid"`
	Name        *string `json:"run_name"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	EndTime     *int64  `json:"end_time"`
	Archived    *bool   `json:"archived"`
}

// GetRunMetrics is a request struct for `POST /runs/:id/metric/get-batch`
type GetRunMetrics []GetRunMetric

// GetRunMetric is one element of GetRunMetricsRequest
type GetRunMetric struct {
	Context map[string]string `json:"context"`
	Name    string            `json:"name"`
}

type SearchRunsRequest struct {
	Query           string   `query:"q"`
	Limit           int      `query:"limit"`
	Offset          string   `query:"offset"`
	Action          string   `query:"action"`
	SkipSystem      bool     `query:"skip_system"`
	ReportProgress  bool     `query:"report_progress"`
	ExcludeParams   bool     `query:"exclude_params"`
	ExcludeTraces   bool     `query:"exclude_traces"`
	ExperimentNames []string `query:"experiment_names"`
}
