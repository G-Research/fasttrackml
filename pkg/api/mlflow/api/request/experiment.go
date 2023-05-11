package request

// ExperimentTagPartialRequest is a partial request object for different requests.
type ExperimentTagPartialRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// CreateExperimentRequest is a request object for `POST /mlflow/experiments/create` endpoint.
type CreateExperimentRequest struct {
	Name             string                        `json:"name"`
	Tags             []ExperimentTagPartialRequest `json:"tags"`
	ArtifactLocation string                        `json:"artifact_location"`
}

// UpdateExperimentRequest is a request object for `POST /mlflow/experiments/update` endpoint.
type UpdateExperimentRequest struct {
	ID   string `json:"experiment_id"`
	Name string `json:"new_name"`
}

// GetExperimentRequest is a request object for `POST /mlflow/experiments/update` endpoint.
type GetExperimentRequest struct {
	ID   string `query:"experiment_id"`
	Name string `query:"experiment_name"`
}

// DeleteExperimentRequest is a request object for `POST /mlflow/experiments/delete` endpoint.
type DeleteExperimentRequest struct {
	ID string `json:"experiment_id"`
}

// RestoreExperimentRequest is a request object for `POST /mlflow/experiments/restore` endpoint.
type RestoreExperimentRequest struct {
	ID string `json:"experiment_id"`
}

// SetExperimentTagRequest is a request object for `POST /mlflow/experiments/set-experiment-tag` endpoint.
type SetExperimentTagRequest struct {
	ID    string `json:"experiment_id"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

// SearchExperimentsRequest is a request object for
// `POST /mlflow/experiments/list` or `POST /mlflow/experiments/search` or `GET /mlflow/experiments/search` endpoints.
type SearchExperimentsRequest struct {
	MaxResults int64    `json:"max_results" query:"max_results"`
	PageToken  string   `json:"page_token"  query:"page_token"`
	Filter     string   `json:"filter"      query:"filter"`
	OrderBy    []string `json:"order_by"    query:"order_by"`
	ViewType   ViewType `json:"view_type"   query:"view_type"`
}
