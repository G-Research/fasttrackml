package response

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

// RunTagPartialResponse is a partial response object for different responses.
type RunTagPartialResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// RunParamPartialResponse is a partial response object for different responses.
type RunParamPartialResponse struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
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

// NewCreateRunResponse creates a new instance of CreateRunResponse object.
func NewCreateRunResponse(run *models.Run) *CreateRunResponse {
	resp := CreateRunResponse{
		Run: RunPartialResponse{
			Info: RunInfoPartialResponse{
				ID:             run.ID,
				UUID:           run.ID,
				Name:           run.Name,
				ExperimentID:   fmt.Sprint(run.ExperimentID),
				UserID:         run.UserID,
				Status:         string(run.Status),
				StartTime:      run.StartTime.Int64,
				ArtifactURI:    run.ArtifactURI,
				LifecycleStage: string(run.LifecycleStage),
			},
			Data: RunDataPartialResponse{
				Tags: make([]RunTagPartialResponse, len(run.Tags)),
			},
		},
	}
	for n, tag := range run.Tags {
		resp.Run.Data.Tags[n] = RunTagPartialResponse{
			Key:   tag.Key,
			Value: tag.Value,
		}
	}
	return &resp
}

// UpdateRunResponse is a response object for `POST mlflow/runs/update` endpoint.
type UpdateRunResponse struct {
	RunInfo RunInfoPartialResponse `json:"run_info"`
}

// NewUpdateRunResponse creates a new UpdateRunResponse object.
func NewUpdateRunResponse(run *models.Run) *UpdateRunResponse {
	// TODO grab name and user from tags?
	return &UpdateRunResponse{
		RunInfo: RunInfoPartialResponse{
			ID:             run.ID,
			UUID:           run.ID,
			Name:           run.Name,
			ExperimentID:   fmt.Sprint(run.ExperimentID),
			UserID:         run.UserID,
			Status:         string(run.Status),
			StartTime:      run.StartTime.Int64,
			EndTime:        run.EndTime.Int64,
			ArtifactURI:    run.ArtifactURI,
			LifecycleStage: string(run.LifecycleStage),
		},
	}
}

// GetRunResponse is a response object for `GET mlflow/runs/get` endpoint.
type GetRunResponse struct {
	Run *RunPartialResponse `json:"run"`
}

// NewGetRunResponse creates a new GetRunResponse object.
func NewGetRunResponse(run *models.Run) *GetRunResponse {
	return &GetRunResponse{
		Run: NewRunPartialResponse(run),
	}
}

// SearchRunsResponse is a response object for `POST mlflow/runs/search` endpoint.
type SearchRunsResponse struct {
	Runs          []*RunPartialResponse `json:"runs"`
	NextPageToken string                `json:"next_page_token,omitempty"`
}

// NewSearchRunsResponse creates a new SearchRunsResponse object.
func NewSearchRunsResponse(runs []models.Run, limit, offset int) (*SearchRunsResponse, error) {
	resp := SearchRunsResponse{
		Runs: make([]*RunPartialResponse, len(runs)),
	}

	// transform each models.Run entity.
	for i, run := range runs {
		//nolint:gosec
		resp.Runs[i] = NewRunPartialResponse(&run)
	}

	// encode `nextPageToken` value.
	if len(runs) == limit {
		var token strings.Builder
		if err := json.NewEncoder(
			base64.NewEncoder(base64.StdEncoding, &token),
		).Encode(request.PageToken{
			Offset: int32(offset + limit),
		}); err != nil {
			return nil, eris.Wrap(err, "error encoding 'nextPageToken' value")
		}
		resp.NextPageToken = token.String()
	}

	return &resp, nil
}

// NewRunPartialResponse is a helper function for NewSearchRunsResponse and NewGetRunResponse functions,
// because they use almost the same response structure.
func NewRunPartialResponse(run *models.Run) *RunPartialResponse {
	metrics := make([]RunMetricPartialResponse, len(run.LatestMetrics))
	for n, m := range run.LatestMetrics {
		metrics[n] = RunMetricPartialResponse{
			Key:       m.Key,
			Value:     m.Value,
			Timestamp: m.Timestamp,
			Step:      m.Step,
		}
		if m.IsNan {
			metrics[n].Value = common.NANValue
		}
	}

	params := make([]RunParamPartialResponse, len(run.Params))
	for n, p := range run.Params {
		params[n] = RunParamPartialResponse{
			Key:   p.Key,
			Value: p.ValueAny(),
		}
	}

	tags := make([]RunTagPartialResponse, len(run.Tags))
	for n, t := range run.Tags {
		tags[n] = RunTagPartialResponse{
			Key:   t.Key,
			Value: t.Value,
		}
		switch t.Key {
		case "mlflow.runName":
			run.Name = t.Value
		case "mlflow.user":
			run.UserID = t.Value
		}
	}

	return &RunPartialResponse{
		Info: RunInfoPartialResponse{
			ID:             run.ID,
			UUID:           run.ID,
			Name:           run.Name,
			ExperimentID:   fmt.Sprint(run.ExperimentID),
			UserID:         run.UserID,
			Status:         string(run.Status),
			StartTime:      run.StartTime.Int64,
			EndTime:        run.EndTime.Int64,
			ArtifactURI:    run.ArtifactURI,
			LifecycleStage: string(run.LifecycleStage),
		},
		Data: RunDataPartialResponse{
			Metrics: metrics,
			Params:  params,
			Tags:    tags,
		},
	}
}
