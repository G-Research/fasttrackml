package dto

// ProjectActivity represents object to store and transfer project activity.
type ProjectActivity struct {
	NumRuns         int64          `json:"num_runs"`
	ActivityMap     map[string]int `json:"activity_map"`
	NumActiveRuns   int64          `json:"num_active_runs"`
	NumExperiments  int64          `json:"num_experiments"`
	NumArchivedRuns int64          `json:"num_archived_runs"`
}
