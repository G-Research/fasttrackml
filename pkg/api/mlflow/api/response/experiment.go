package response

import (
	"fmt"

	"github.com/G-Research/fasttrackml/pkg/database"
)

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

// CreateExperimentResponse is a response object for `POST /mlflow/experiments/create` endpoint.
type CreateExperimentResponse struct {
	ID string `json:"experiment_id"`
}

// NewCreateExperimentResponse creates new CreateExperimentResponse object.
func NewCreateExperimentResponse(experiment *database.Experiment) *CreateExperimentResponse {
	return &CreateExperimentResponse{
		ID: fmt.Sprint(*experiment.ID),
	}
}

// GetExperimentResponse is a response object for `GET /mlflow/experiments/get` endpoint.
type GetExperimentResponse struct {
	Experiment ExperimentPartialResponse `json:"experiment"`
}

// NewExperimentResponse creates new GetExperimentResponse object.
func NewExperimentResponse(experiment *database.Experiment) *GetExperimentResponse {
	response := GetExperimentResponse{
		Experiment: ExperimentPartialResponse{
			// TODO:DSuhinin - we have to check that value is not null before use it. Ideally get rid of pointer.
			ID:               fmt.Sprint(*experiment.ID),
			Name:             experiment.Name,
			ArtifactLocation: experiment.ArtifactLocation,
			LifecycleStage:   string(experiment.LifecycleStage),
			LastUpdateTime:   experiment.LastUpdateTime.Int64,
			CreationTime:     experiment.CreationTime.Int64,
			Tags:             make([]ExperimentTagPartialResponse, len(experiment.Tags)),
		},
	}

	for n, t := range experiment.Tags {
		response.Experiment.Tags[n] = ExperimentTagPartialResponse{
			Key:   t.Key,
			Value: t.Value,
		}
	}
	return &response
}

// SearchExperimentsResponse is a response object for `GET /mlflow/experiments/search` endpoint.
type SearchExperimentsResponse struct {
	Experiments   []ExperimentPartialResponse `json:"experiments"`
	NextPageToken string                      `json:"next_page_token,omitempty"`
}
