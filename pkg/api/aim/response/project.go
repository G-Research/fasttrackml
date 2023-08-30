package response

// ProjectActivity represents the response json for the `GET aim/projects/activity` endpoint.
type ProjectActivity struct {
	NumExperiments  float64        `json:"num_experiments"`
	NumRuns         float64        `json:"num_runs"`
	NumActiveRuns   float64        `json:"num_active_runs"`
	NumArchivedRuns float64        `json:"num_archived_runs"`
	ActivityMap     map[string]int `json:"activity_map"`
}

// GetProject represents the response json for the `GET aim/projects` endpoint.
type GetProject struct {
	Name             string  `json:"name"`
	Path             string  `json:"path"`
	Description      string  `json:"description"`
	TelemetryEnabled float64 `json:"telementry_enabled"`
}

// ProjectParamsResponse is a response object for `GET aim/projects/params` endpoint.
type ProjectParamsResponse struct {
	Metric map[string][]struct{}  `json:"metric"`
	Params map[string]interface{} `json:"params"`
}
