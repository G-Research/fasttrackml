package response

// ExperimentTagPartialResponse is a partial response object for different responses.
type ExperimentTagPartialResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ExperimentPartialResponse is a partial response object for different responses.
type ExperimentPartialResponse struct {
	ID               string                         `json:"experiment_id"`
	Name             string                         `json:"name"`
	ArtifactLocation string                         `json:"artifact_location"`
	LifecycleStage   string                         `json:"lifecycle_stage"`
	LastUpdateTime   int64                          `json:"last_update_time"`
	CreationTime     int64                          `json:"creation_time"`
	Tags             []ExperimentTagPartialResponse `json:"tags"`
}

// CreateExperimentResponse is a response object for `POST mlflow/create` endpoint.
type CreateExperimentResponse struct {
	ID string `json:"experiment_id"`
}

// GetExperimentResponse is a response object for `GET mlflow/get` endpoint.
type GetExperimentResponse struct {
	Experiment ExperimentPartialResponse `json:"experiment"`
}

// SearchExperimentsResponse is a response object for `GET mlflow/search` endpoint.
type SearchExperimentsResponse struct {
	Experiments   []ExperimentPartialResponse `json:"experiments"`
	NextPageToken string                      `json:"next_page_token,omitempty"`
}
