package response

type GetExperiment struct {
	ID       int32  `json:"id"`
	Name     string `json:"name"`
	Archived bool   `json:"archived"`
	RunCount int    `json:"run_count"`
}

type GetExperimentActivity struct {
	NumRuns         int            `json:"num_runs"`
	NumArchivedRuns int            `json:"num_archived_runs"`
	NumActiveRuns   int            `json:"num_active_runs"`
	ActivityMap     map[string]int `json:"activity_map"`
}

type GetExperimentRuns struct {
	ID   string          `json:"id"`
	Runs []ExperimentRun `json:"runs"`
}

type ExperimentRun struct {
	ID           string  `json:"run_id"`
	Name         string  `json:"name"`
	CreationTime float64 `json:"creation_time"`
	EndTime      float64 `json:"end_time"`
	Archived     bool    `json:"archived"`
}

type DeleteExperiment struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}
