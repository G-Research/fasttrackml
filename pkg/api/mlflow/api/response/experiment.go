package response

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"

	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
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
func NewCreateExperimentResponse(experiment *models.Experiment) *CreateExperimentResponse {
	return &CreateExperimentResponse{
		ID: fmt.Sprint(*experiment.ID),
	}
}

// GetExperimentResponse is a response object for `GET /mlflow/experiments/get` endpoint.
type GetExperimentResponse struct {
	Experiment *ExperimentPartialResponse `json:"experiment"`
}

// NewExperimentResponse creates new GetExperimentResponse object.
func NewExperimentResponse(experiment *models.Experiment) *GetExperimentResponse {
	return &GetExperimentResponse{
		Experiment: NewExperimentPartialResponse(experiment),
	}
}

// SearchExperimentsResponse is a response object for `GET /mlflow/experiments/search` endpoint.
type SearchExperimentsResponse struct {
	Experiments   []*ExperimentPartialResponse `json:"experiments"`
	NextPageToken string                       `json:"next_page_token,omitempty"`
}

// NewSearchExperimentsResponse  creates new SearchExperimentsResponse object.
func NewSearchExperimentsResponse(
	experiments []models.Experiment, limit, offset int,
) (*SearchExperimentsResponse, error) {
	// encode `nextPageToken` value.
	var token strings.Builder
	if len(experiments) > limit {
		experiments = experiments[:limit]
		if err := json.NewEncoder(
			base64.NewEncoder(base64.StdEncoding, &token),
		).Encode(request.PageToken{
			Offset: int32(offset + limit),
		}); err != nil {
			return nil, eris.Wrap(err, "error encoding 'nextPageToken' value")
		}
	}

	resp := SearchExperimentsResponse{
		NextPageToken: token.String(),
	}
	// transform each models.Experiment entity.
	for _, experiment := range experiments {
		resp.Experiments = append(resp.Experiments, NewExperimentPartialResponse(&experiment))
	}

	return &resp, nil
}

// NewExperimentPartialResponse is a helper function for NewExperimentResponse and NewSearchExperimentsResponse functions,
// because the use almost the same response structure.
func NewExperimentPartialResponse(experiment *models.Experiment) *ExperimentPartialResponse {
	tags := make([]ExperimentTagPartialResponse, len(experiment.Tags))
	for n, t := range experiment.Tags {
		tags[n] = ExperimentTagPartialResponse{
			Key:   t.Key,
			Value: t.Value,
		}
	}

	return &ExperimentPartialResponse{
		ID:               fmt.Sprint(*experiment.ID),
		Name:             experiment.Name,
		ArtifactLocation: experiment.ArtifactLocation,
		LifecycleStage:   string(experiment.LifecycleStage),
		LastUpdateTime:   experiment.LastUpdateTime.Int64,
		CreationTime:     experiment.CreationTime.Int64,
		Tags:             tags,
	}
}
