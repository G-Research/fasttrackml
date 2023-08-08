package response

// GetExperiment represents the response json fot the GetExperimnt endpoint.
type GetExperiment struct {
	ID           int32   `json:"id"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Archived     bool    `json:"archived"`
	RunCount     int     `json:"run_count"`
	CreationTime float64 `json:"creation_time"`
}

// GetExperiment represents the response json fot the GetExperimntActivity endpoint.
type GetExperimentActivity struct {
	NumRuns         int            `json:"num_runs"`
	NumArchivedRuns int            `json:"num_archived_runs"`
	NumActiveRuns   int            `json:"num_active_runs"`
	ActivityMap     map[string]int `json:"activity_map"`
}

// GetExperiment represents the response json fot the GetExperimntRuns endpoint.
type GetExperimentRuns struct {
	ID   string          `json:"id"`
	Runs []ExperimentRun `json:"runs"`
}

// ExperimentRun represents a run of an experiment.
type ExperimentRun struct {
	ID           string  `json:"run_id"`
	Name         string  `json:"name"`
	CreationTime float64 `json:"creation_time"`
	EndTime      float64 `json:"end_time"`
	Archived     bool    `json:"archived"`
}

// DeleteExperiment represents the response json fot the DeleteExperimnt endpoint.
type DeleteExperiment struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// Experiments is the response struct for the GetExperiments endpoint (slice of GetExperiment).
type Experiments []GetExperiment
