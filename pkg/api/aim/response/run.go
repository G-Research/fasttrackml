package response

import "encoding/json"

// GetRunInfo represents the response struct for GetRunInfo endpoint
type GetRunInfo struct {
	Params GetRunInfoParams `json:"params"`
	Traces GetRunInfoTraces `json:"traces"`
	Props  GetRunInfoProps  `json:"props"`
}

// GetRunsActive represents the response struct for GetRunsActive endpoint
type GetRunsActive map[string]GetRunInfo

// GetRunInfoParams is a partial response object for GetRunInfo.
type GetRunInfoParams struct {
	Tags map[string]string `json:"tags"`
}

// GetRunInfoTraces is a partial response object for GetRunInfo.
type GetRunInfoTraces struct {
	Tags   map[string]string        `json:"tags"`
	Metric []GetRunInfoTracesMetric `json:"metric"`
}

// GetRunInfoTracesMetric is a partial response object for GetRunInfoTraces.
type GetRunInfoTracesMetric struct {
	Name      string          `json:"name"`
	Context   json.RawMessage `json:"context"`
	LastValue float64         `json:"last_value"`
}

// GetRunInfoProps is a partial response object for GetRunInfo.
type GetRunInfoProps struct {
	Name         string               `json:"name"`
	Description  string               `json:"description"`
	Experiment   GetRunInfoExperiment `json:"experiment"`
	Tags         []string             `json:"tags"`
	CreationTime float64              `json:"creation_time"`
	EndTime      float64              `json:"end_time"`
	Archived     bool                 `json:"archived"`
	Active       bool                 `json:"active"`
}

// GetRunInfoExperiment experiment properties
type GetRunInfoExperiment struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GetRunMetrics is the reponse struct for GetRunMetrics endpoint (slice of RunMetric)
type GetRunMetrics []RunMetrics

// RunMetrics is one run metrics
type RunMetrics struct {
	Name    string          `json:"name"`
	Context json.RawMessage `json:"context"`
	Values  []float64       `json:"values"`
	Iters   []int64         `json:"iters"`
}
