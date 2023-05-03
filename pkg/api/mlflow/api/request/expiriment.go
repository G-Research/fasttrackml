package request

// ExperimentTagPartialRequest is a partial request object for different requests.
type ExperimentTagPartialRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ExperimentPartialRequest is a partial request object for different requests.
type ExperimentPartialRequest struct {
	ID               string                        `json:"experiment_id"`
	Name             string                        `json:"name"`
	ArtifactLocation string                        `json:"artifact_location"`
	LifecycleStage   string                        `json:"lifecycle_stage"`
	LastUpdateTime   int64                         `json:"last_update_time"`
	CreationTime     int64                         `json:"creation_time"`
	Tags             []ExperimentTagPartialRequest `json:"tags"`
}

// CreateExperimentRequest is a request object for `POST mlflow/create` endpoint.
type CreateExperimentRequest struct {
	Name             string                        `json:"name"`
	Tags             []ExperimentTagPartialRequest `json:"tags"`
	ArtifactLocation string                        `json:"artifact_location"`
}

// UpdateExperimentRequest is a request object for `POST mlflow/update` endpoint.
type UpdateExperimentRequest struct {
	ID   string `json:"experiment_id"`
	Name string `json:"new_name"`
}

// DeleteExperimentRequest is a request object for `POST mlflow/delete` endpoint.
type DeleteExperimentRequest struct {
	ID string `json:"experiment_id"`
}

// RestoreExperimentRequest is a request object for `POST mlflow/restore` endpoint.
type RestoreExperimentRequest struct {
	ID string `json:"experiment_id"`
}

// SetExperimentTagRequest is a request object for `POST mlflow/set-experiment-tag` endpoint.
type SetExperimentTagRequest struct {
	ID    string `json:"experiment_id"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

// SearchExperimentsRequest is a request object for `POST mlflow/list` or `POST mlflow/search` or `GET mlflow/search` endpoint.
type SearchExperimentsRequest struct {
	MaxResults int64    `json:"max_results" query:"max_results"`
	PageToken  string   `json:"page_token"  query:"page_token"`
	Filter     string   `json:"filter"      query:"filter"`
	OrderBy    []string `json:"order_by"    query:"order_by"`
	ViewType   string   `json:"view_type"   query:"view_type"`
}
