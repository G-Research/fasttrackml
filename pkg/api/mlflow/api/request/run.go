package request

// RunTagPartialRequest is a partial request object for different requests.
type RunTagPartialRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// GetRunRequest is a request object for `GET /mlflow/runs/get` endpoint.
type GetRunRequest struct {
	RunID   string `query:"run_id"`
	RunUUID string `query:"run_uuid"`
}

// GetRunID returns Run RunID.
func (r GetRunRequest) GetRunID() string {
	if r.RunID != "" {
		return r.RunID
	}
	return r.RunUUID
}

// CreateRunRequest is a request object for `POST /mlflow/runs/create` endpoint.
type CreateRunRequest struct {
	ExperimentID string                 `json:"experiment_id"`
	UserID       string                 `json:"user_id"`
	Name         string                 `json:"run_name"`
	StartTime    int64                  `json:"start_time"`
	Tags         []RunTagPartialRequest `json:"tags"`
}

// UpdateRunRequest is a request object for `POST /mlflow/runs/update` endpoint.
type UpdateRunRequest struct {
	RunID   string `json:"run_id"`
	RunUUID string `json:"run_uuid"`
	Name    string `json:"run_name"`
	Status  string `json:"status"`
	EndTime int64  `json:"end_time"`
}

// GetRunID returns Run RunID.
func (r UpdateRunRequest) GetRunID() string {
	if r.RunID != "" {
		return r.RunID
	}
	return r.RunUUID
}

// SearchRunsRequest is a request object for `POST /mlflow/runs/search` endpoint.
type SearchRunsRequest struct {
	ExperimentIDs []string `json:"experiment_ids"`
	Filter        string   `json:"filter"`
	ViewType      ViewType `json:"run_view_type"`
	MaxResults    int32    `json:"max_results"`
	OrderBy       []string `json:"order_by"`
	PageToken     string   `json:"page_token"`
}

// RestoreRunRequest is a request object for `POST /mlflow/runs/restore` endpoint.
type RestoreRunRequest struct {
	RunID string `json:"run_id"`
}

// DeleteRunRequest is a request object for `POST /mlflow/runs/delete` endpoint.
type DeleteRunRequest struct {
	RunID string `json:"run_id"`
}

// SetRunTagRequest is a request object for `POST /mlflow/runs/set-tag` endpoint.
type SetRunTagRequest struct {
	RunID   string `json:"run_id"`
	RunUUID string `json:"run_uuid"`
	Key     string `json:"key"`
	Value   string `json:"value"`
}

// GetRunID returns Run RunID.
func (r SetRunTagRequest) GetRunID() string {
	if r.RunID != "" {
		return r.RunID
	}
	return r.RunUUID
}

// DeleteRunTagRequest is a request object for `POST /mlflow/runs/delete-tag` endpoint.
type DeleteRunTagRequest struct {
	RunID string `json:"run_id"`
	Key   string `json:"key"`
}
