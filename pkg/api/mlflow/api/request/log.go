package request

// ParamPartialRequest is a partial request object for different requests.
type ParamPartialRequest struct {
	Key        string   `json:"key"`
	ValueInt   *int64   `json:"value_int"`
	ValueFloat *float64 `json:"value_float"`
	ValueStr   *string  `json:"value"`
}

// TagPartialRequest is a partial request object for different requests.
type TagPartialRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// MetricPartialRequest is a partial request object for different requests.
type MetricPartialRequest struct {
	Key       string         `json:"key"`
	Value     any            `json:"value"`
	Timestamp int64          `json:"timestamp"`
	Step      int64          `json:"step"`
	Context   map[string]any `json:"context"`
}

// LogParamRequest is a request object for `POST mlflow/runs/log-parameter` endpoint.
type LogParamRequest struct {
	RunID      string   `json:"run_id"`
	RunUUID    string   `json:"run_uuid"`
	Key        string   `json:"key"`
	ValueInt   *int64   `json:"value_int"`
	ValueFloat *float64 `json:"value_float"`
	ValueStr   *string  `json:"value"`
}

// GetRunID returns Run ID.
func (r LogParamRequest) GetRunID() string {
	if r.RunID != "" {
		return r.RunID
	}
	return r.RunUUID
}

// LogMetricRequest is a request object for `POST mlflow/runs/log-metric` endpoint.
type LogMetricRequest struct {
	RunID     string         `json:"run_id"`
	RunUUID   string         `json:"run_uuid"`
	Key       string         `json:"key"`
	Value     any            `json:"value"`
	Timestamp int64          `json:"timestamp"`
	Step      int64          `json:"step"`
	Context   map[string]any `json:"context"`
}

// GetRunID returns Run ID.
func (r LogMetricRequest) GetRunID() string {
	if r.RunID != "" {
		return r.RunID
	}
	return r.RunUUID
}

// LogBatchRequest is a request object for `POST mlflow/runs/log-batch` endpoint.
type LogBatchRequest struct {
	RunID   string                 `json:"run_id"`
	Tags    []TagPartialRequest    `json:"tags,omitempty"`
	Params  []ParamPartialRequest  `json:"params,omitempty"`
	Metrics []MetricPartialRequest `json:"metrics,omitempty"`
}

// LogOutputRequest is a request object for `POST mlflow/runs/log-output` endpoint.
type LogOutputRequest struct {
	Data  string `json:"data"`
	RunID string `json:"run_id"`
}

// LogArtifactRequest is a request object for `POST mlflow/runs/log-artifact` endpoint.
type LogArtifactRequest struct {
	Iter    int64  `json:"iter"`
	Step    int64  `json:"step"`
	Caption string `json:"caption"`
	Index   int64  `json:"index"`
	Width   int64  `json:"width"`
	RunID   string `json:"run_id"`
	Height  int64  `json:"height"`
	Format  string `json:"format"`
	BlobURI string `json:"blob_uri"`
}
