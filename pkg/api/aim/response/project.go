package response

// ProjectActivity represents the response json in Project activity endpoints
type ProjectActivity struct {
	NumExperiments  float64        `json:"num_experiments"`
	NumRuns         float64        `json:"num_runs"`
	NumActiveRuns   float64        `json:"num_active_runs"`
	NumArchivedRuns float64        `json:"num_archived_runs"`
	ActivityMap     map[string]int `json:"activity_map"`
}
