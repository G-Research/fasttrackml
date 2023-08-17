package response

// ProjectActivity represents the response json in Project activity endpoints
type ProjectActivity struct {
	NumExperiments  float64        `json:"num_experiments"`
	NumRuns         float64        `json:"num_runs"`
	NumActiveRuns   float64        `json:"num_active_runs"`
	NumArchivedRuns float64        `json:"num_archived_runs"`
	ActivityMap     map[string]int `json:"activity_map"`
}

// GetProject reprsents the response json in Get Project endpoint
type GetProject struct {
	Name             string  `json:"name"`
	Path             string  `json:"path"`
	Description      string  `json:"description"`
	TelemetryEnabled float64 `json:"telementry_enabled"`
}
