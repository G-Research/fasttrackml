package response

import (
	"fmt"
	"strconv"

	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/models"
)

// Experiment represents the response object to hold models.ExperimentExtended data.
type Experiment struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Archived     bool    `json:"archived"`
	RunCount     int     `json:"run_count"`
	CreationTime float64 `json:"creation_time"`
}

// NewGetExperimentResponse creates new response object for `GET /experiments/:id` endpoint.
func NewGetExperimentResponse(experiment *models.ExperimentExtended) Experiment {
	return Experiment{
		ID:           strconv.Itoa(int(*experiment.ID)),
		Name:         experiment.Name,
		Description:  experiment.Description,
		Archived:     experiment.LifecycleStage == models.LifecycleStageDeleted,
		RunCount:     experiment.RunCount,
		CreationTime: float64(experiment.CreationTime.Int64) / 1000,
	}
}

// NewGetExperimentsResponse creates new response object for `GET /experiments` endpoint.
func NewGetExperimentsResponse(experiments []models.ExperimentExtended) []Experiment {
	resp := make([]Experiment, len(experiments))
	for i, experiment := range experiments {
		//nolint:gosec
		resp[i] = NewGetExperimentResponse(&experiment)
	}
	return resp
}

// ExperimentRunPartial represents partial object of ExperimentRuns.
type ExperimentRunPartial struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	CreationTime int64  `json:"creationTime"`
	EndTime      int64  `json:"endTime"`
	Archived     bool   `json:"archived"`
}

// ExperimentRuns represents the response object to hold models.Runs data.
type ExperimentRuns struct {
	ID   string                 `json:"id"`
	Runs []ExperimentRunPartial `json:"runs"`
}

// NewGetExperimentRunsResponse creates new response object for `GET /experiments/:id/runs` endpoint.
func NewGetExperimentRunsResponse(experimentID int32, runs []models.Run) *ExperimentRuns {
	experimentRuns := make([]ExperimentRunPartial, len(runs))
	for i, run := range runs {
		experimentRuns[i] = ExperimentRunPartial{
			ID:           run.ID,
			Name:         run.Name,
			CreationTime: int64(float64(run.StartTime.Int64) / 1000),
			EndTime:      int64(float64(run.EndTime.Int64) / 1000),
			Archived:     run.LifecycleStage == models.LifecycleStageDeleted,
		}
	}
	return &ExperimentRuns{
		ID:   fmt.Sprintf("%d", experimentID),
		Runs: experimentRuns,
	}
}

// ExperimentActivity represents the response object to hold models.Experiment activity data.
type ExperimentActivity struct {
	NumRuns         int            `json:"num_runs"`
	ActivityMap     map[string]int `json:"activity_map"`
	NumActiveRuns   int            `json:"num_active_runs"`
	NumArchivedRuns int            `json:"num_archived_runs"`
}

// NewGetExperimentActivityResponse creates new response object for `GET /experiments/:id/activity` endpoint.
func NewGetExperimentActivityResponse(activity *models.ExperimentActivity) *ExperimentActivity {
	return &ExperimentActivity{
		NumRuns:         activity.NumRuns,
		ActivityMap:     activity.ActivityMap,
		NumActiveRuns:   activity.NumActiveRuns,
		NumArchivedRuns: activity.NumArchivedRuns,
	}
}

// UpdateExperimentResponse is a response object to hold response data for `PUT experiments/:id` endpoint.
type UpdateExperimentResponse struct {
	ID     string `json:"ID"`
	Status string `json:"status"`
}

// NewUpdateExperimentResponse creates new response object for `PUT experiments/:id` endpoint.
func NewUpdateExperimentResponse(id int32, status string) *UpdateExperimentResponse {
	return &UpdateExperimentResponse{
		ID:     fmt.Sprintf("%d", id),
		Status: status,
	}
}

// DeleteExperimentResponse is a response object to hold response data for `DELETE experiments/:id` endpoint.
type DeleteExperimentResponse struct {
	ID     string `json:"ID"`
	Status string `json:"status"`
}

// NewDeleteExperimentResponse creates new response object for `DELETE experiments/:id` endpoint.
func NewDeleteExperimentResponse(id int32, status string) *DeleteExperimentResponse {
	return &DeleteExperimentResponse{
		ID:     fmt.Sprintf("%d", id),
		Status: status,
	}
}
