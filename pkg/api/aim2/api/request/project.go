package request

// GetProjectParamsRequest is a request object for `GET /projects/params` endpoint.
type GetProjectParamsRequest struct {
	Sequences     []string `query:"sequence"`
	Experiments   []int    `query:"experiments"`
	ExcludeParams bool     `query:"exclude_params"`
}
