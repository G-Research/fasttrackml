package response

import (
	"github.com/gofiber/fiber/v2"

	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/dto"
)

// ProjectActivityResponse represents the response json for the `GET aim/projects/activity` endpoint.
type ProjectActivityResponse struct {
	NumRuns         int64          `json:"num_runs"`
	NumActiveRuns   int64          `json:"num_active_runs"`
	NumExperiments  int64          `json:"num_experiments"`
	NumArchivedRuns int64          `json:"num_archived_runs"`
	ActivityMap     map[string]int `json:"activity_map"`
}

// NewProjectActivityResponse creates new response object for `GET /projects/activity` endpoint.
func NewProjectActivityResponse(activity *dto.ProjectActivity) *ProjectActivityResponse {
	return &ProjectActivityResponse{
		NumRuns:         activity.NumRuns,
		NumActiveRuns:   activity.NumActiveRuns,
		NumExperiments:  activity.NumExperiments,
		NumArchivedRuns: activity.NumArchivedRuns,
		ActivityMap:     activity.ActivityMap,
	}
}

// GetProjectResponse represents the response json for the `GET aim/projects` endpoint.
type GetProjectResponse struct {
	Name             string `json:"name"`
	Path             string `json:"path"`
	Description      string `json:"description"`
	TelemetryEnabled int    `json:"telemetry_enabled"`
}

// NewGetProjectResponse creates new response object for `GET /projects` endpoint.
func NewGetProjectResponse(name, dialector string) *GetProjectResponse {
	return &GetProjectResponse{
		Name: name,
		Path: dialector,
	}
}

// ProjectParamsResponse is a response object for `GET aim/projects/params` endpoint.
type ProjectParamsResponse struct {
	Metric map[string][]fiber.Map `json:"metric"`
	Params map[string]interface{} `json:"params"`
}
