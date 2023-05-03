package request

// RunTagPartialRequest is a partial request object for different requests.
type RunTagPartialRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// RunParamPartialRequest is a partial request object for different requests.
type RunParamPartialRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// RunMetricPartialRequest is a partial request object for different requests.
type RunMetricPartialRequest struct {
	Key       string `json:"key"`
	Value     any    `json:"value"`
	Timestamp int64  `json:"timestamp"`
	Step      int64  `json:"step"`
}

// RunDataPartialRequest is a partial request object for different requests.
type RunDataPartialRequest struct {
	Metrics []RunMetricPartialRequest `json:"metrics,omitempty"`
	Params  []RunParamPartialRequest  `json:"params,omitempty"`
	Tags    []RunTagPartialRequest    `json:"tags,omitempty"`
}

// RunPartialRequest is a partial request object for different requests.
type RunPartialRequest struct {
	Info RunInfoPartialRequest `json:"info"`
	Data RunDataPartialRequest `json:"data"`
}

// RunInfoPartialRequest is a partial request object for different requests.
type RunInfoPartialRequest struct {
	UUID           string `json:"run_uuid"`
	Name           string `json:"run_name"`
	ExperimentID   string `json:"experiment_id"`
	UserID         string `json:"user_id,omitempty"`
	Status         string `json:"status"`
	StartTime      int64  `json:"start_time"`
	EndTime        int64  `json:"end_time,omitempty"`
	ArtifactURI    string `json:"artifact_uri,omitempty"`
	LifecycleStage string `json:"lifecycle_stage"`
	ID             string `json:"run_id"`
}

// CreateRunRequest is a request object for `POST mlflow/create` endpoint.
type CreateRunRequest struct {
	ExperimentID string                 `json:"experiment_id"`
	UserID       string                 `json:"user_id"`
	Name         string                 `json:"run_name"`
	StartTime    int64                  `json:"start_time"`
	Tags         []RunTagPartialRequest `json:"tags"`
}

// UpdateRunRequest is a request object for `POST mlflow/update` endpoint.
type UpdateRunRequest struct {
	ID      string `json:"run_id"`
	UUID    string `json:"run_uuid"`
	Name    string `json:"run_name"`
	Status  string `json:"status"`
	EndTime int64  `json:"end_time"`
}

// SearchRunsRequest is a request object for `POST mlflow/search` endpoint.
type SearchRunsRequest struct {
	ExperimentIDs []string `json:"experiment_ids"`
	Filter        string   `json:"filter"`
	ViewType      string   `json:"run_view_type"`
	MaxResults    int32    `json:"max_results"`
	OrderBy       []string `json:"order_by"`
	PageToken     string   `json:"page_token"`
}

// RestoreRunRequest is a request object for `POST mlflow/restore` endpoint.
type RestoreRunRequest struct {
	ID   string `json:"run_id"`
	UUID string `json:"run_uuid"`
}

// DeleteRunRequest is a request object for `POST mlflow/delete` endpoint.
type DeleteRunRequest struct {
	ID   string `json:"run_id"`
	UUID string `json:"run_uuid"`
}

// SetRunTagRequest is a request object for `POST mlflow/set-tag` endpoint.
type SetRunTagRequest struct {
	ID    string `json:"run_id"`
	UUID  string `json:"run_uuid"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

// DeleteRunTagRequest is a request object for `POST mlflow/delete-tag` endpoint.
type DeleteRunTagRequest struct {
	ID  string `json:"run_id"`
	Key string `json:"key"`
}
