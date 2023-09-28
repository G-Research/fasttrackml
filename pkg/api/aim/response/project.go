package response

// ProjectActivityResponse represents the response json for the `GET aim/projects/activity` endpoint.
type ProjectActivityResponse struct {
	NumExperiments  int64          `json:"num_experiments"`
	NumRuns         int64          `json:"num_runs"`
	NumActiveRuns   int64          `json:"num_active_runs"`
	NumArchivedRuns int64          `json:"num_archived_runs"`
	ActivityMap     map[string]int `json:"activity_map"`
}

// GetProjectResponse represents the response json for the `GET aim/projects` endpoint.
type GetProjectResponse struct {
	Name             string `json:"name"`
	Path             string `json:"path"`
	Description      string `json:"description"`
	TelemetryEnabled bool   `json:"telementry_enabled"`
}

// ProjectParamsResponse is a response object for `GET aim/projects/params` endpoint.
type ProjectParamsResponse struct {
	Metric map[string][]struct{}  `json:"metric"`
	Params map[string]interface{} `json:"params"`
}
