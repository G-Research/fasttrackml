package request

// UpdateRunRequest is a request struct for `PUT /runs/:id` endpoint.
type UpdateRunRequest struct {
	RunID       *string `json:"run_id"`
	RunUUID     *string `json:"run_uuid"`
	Name        *string `json:"run_name"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	EndTime     *int64  `json:"end_time"`
	Archived    *bool   `json:"archived"`
}

// GetRunMetrics is a request struct for `POST /runs/:id/metric/get-batch`
type GetRunMetrics []GetRunMetric

// GetRunMetric is one element of GetRunMetricsRequest
type GetRunMetric struct {
	Context map[string]string `json:"context"`
	Name    string            `json:"name"`
}
