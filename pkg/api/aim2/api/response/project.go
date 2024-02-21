package response

import "github.com/gofiber/fiber/v2"

// ProjectActivityResponse represents the response json for the `GET aim/projects/activity` endpoint.
type ProjectActivityResponse struct {
	NumExperiments  int            `json:"num_experiments"`
	NumRuns         int            `json:"num_runs"`
	NumActiveRuns   int            `json:"num_active_runs"`
	NumArchivedRuns int            `json:"num_archived_runs"`
	ActivityMap     map[string]int `json:"activity_map"`
}

// GetProjectResponse represents the response json for the `GET aim/projects` endpoint.
type GetProjectResponse struct {
	Name             string `json:"name"`
	Path             string `json:"path"`
	Description      string `json:"description"`
	TelemetryEnabled int    `json:"telementry_enabled"`
}

// ProjectParamsResponse is a response object for `GET aim/projects/params` endpoint.
type ProjectParamsResponse struct {
	Metric map[string][]fiber.Map `json:"metric"`
	Params map[string]interface{} `json:"params"`
}
