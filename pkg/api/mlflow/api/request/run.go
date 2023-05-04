package request

// RunTagPartialRequest is a partial request object for different requests.
type RunTagPartialRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// CreateRunRequest is a request object for `POST mlflow/run/create` endpoint.
type CreateRunRequest struct {
	ExperimentID string                 `json:"experiment_id"`
	UserID       string                 `json:"user_id"`
	Name         string                 `json:"run_name"`
	StartTime    int64                  `json:"start_time"`
	Tags         []RunTagPartialRequest `json:"tags"`
}

// UpdateRunRequest is a request object for `POST mlflow/run/update` endpoint.
type UpdateRunRequest struct {
	ID      string `json:"run_id"`
	UUID    string `json:"run_uuid"`
	Name    string `json:"run_name"`
	Status  string `json:"status"`
	EndTime int64  `json:"end_time"`
}

// SearchRunsRequest is a request object for `POST mlflow/run/search` endpoint.
type SearchRunsRequest struct {
	ExperimentIDs []string `json:"experiment_ids"`
	Filter        string   `json:"filter"`
	ViewType      string   `json:"run_view_type"`
	MaxResults    int32    `json:"max_results"`
	OrderBy       []string `json:"order_by"`
	PageToken     string   `json:"page_token"`
}

// RestoreRunRequest is a request object for `POST mlflow/run/restore` endpoint.
type RestoreRunRequest struct {
	ID   string `json:"run_id"`
	UUID string `json:"run_uuid"`
}

// DeleteRunRequest is a request object for `POST mlflow/run/delete` endpoint.
type DeleteRunRequest struct {
	ID   string `json:"run_id"`
	UUID string `json:"run_uuid"`
}

// SetRunTagRequest is a request object for `POST mlflow/run/set-tag` endpoint.
type SetRunTagRequest struct {
	ID    string `json:"run_id"`
	UUID  string `json:"run_uuid"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

// DeleteRunTagRequest is a request object for `POST mlflow/run/delete-tag` endpoint.
type DeleteRunTagRequest struct {
	ID  string `json:"run_id"`
	Key string `json:"key"`
}
