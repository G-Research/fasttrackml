package response

// RunTagPartialResponse is a partial response object for different responses.
type RunTagPartialResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// RunParamPartialResponse is a partial response object for different responses.
type RunParamPartialResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// RunMetricPartialResponse is a partial response object for different responses.
type RunMetricPartialResponse struct {
	Key       string `json:"key"`
	Value     any    `json:"value"`
	Timestamp int64  `json:"timestamp"`
	Step      int64  `json:"step"`
}

// RunDataPartialResponse is a partial response object for different responses.
type RunDataPartialResponse struct {
	Metrics []RunMetricPartialResponse `json:"metrics,omitempty"`
	Params  []RunParamPartialResponse  `json:"params,omitempty"`
	Tags    []RunTagPartialResponse    `json:"tags,omitempty"`
}

// RunInfoPartialResponse is a partial response object for different responses.
type RunInfoPartialResponse struct {
	ID             string `json:"run_id"`
	UUID           string `json:"run_uuid"`
	Name           string `json:"run_name"`
	ExperimentID   string `json:"experiment_id"`
	UserID         string `json:"user_id,omitempty"`
	Status         string `json:"status"`
	StartTime      int64  `json:"start_time"`
	EndTime        int64  `json:"end_time,omitempty"`
	ArtifactURI    string `json:"artifact_uri,omitempty"`
	LifecycleStage string `json:"lifecycle_stage"`
}

// RunPartialResponse is a partial response object for different responses.
type RunPartialResponse struct {
	Info RunInfoPartialResponse `json:"info"`
	Data RunDataPartialResponse `json:"data"`
}

// CreateRunResponse is a response object for `POST mlflow/runs/create` endpoint.
type CreateRunResponse struct {
	Run RunPartialResponse `json:"run"`
}

// UpdateRunResponse is a response object for `POST mlflow/runs/update` endpoint.
type UpdateRunResponse struct {
	RunInfo RunInfoPartialResponse `json:"run_info"`
}

// GetRunResponse is a response object for `GET mlflow/runs/get` endpoint.
type GetRunResponse struct {
	Run RunPartialResponse `json:"run"`
}

// SearchRunsResponse is a response object for `POST mlflow/runs/search` endpoint.
type SearchRunsResponse struct {
	Runs          []RunPartialResponse `json:"runs"`
	NextPageToken string               `json:"next_page_token,omitempty"`
}
