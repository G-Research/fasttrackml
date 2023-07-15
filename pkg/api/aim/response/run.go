package response

// GetRunInfo represents the response struct for GetRunInfo endpoint
type GetRunInfo struct {
	Params GetRunInfoParams `json:"params"`
	Traces GetRunInfoTraces `json:"traces"`
	Props  GetRunInfoProps  `json:"props"`
}

// GetRunInfoParams params
type GetRunInfoParams struct {
	Tags map[string]string `json:"tags"`
}

// GetRunInfoTraces traces
type GetRunInfoTraces struct {
	Tags map[string]string `json:"tags"`
}

// GetRunInfoProps run properties
type GetRunInfoProps struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Experiment  GetRunInfoExperiment `json:"experiment"`
	Tags        []string             `json:"tags"`
	StartTime   float64              `json:"creation_time"`
	EndTime     float64              `json:"end_time"`
	Archived    bool                 `json:"archived"`
	Active      bool                 `json:"active"`
}

// GetRunInfoExperiment experiment properties
type GetRunInfoExperiment struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
