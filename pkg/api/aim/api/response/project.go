package response

// ProjectParamsResponse is a response object for `GET aim/projects/params` endpoint.
type ProjectParamsResponse struct {
	Metric map[string][]struct{}  `json:"metric"`
	Params map[string]interface{} `json:"params"`
}
