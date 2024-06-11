package project

import "github.com/G-Research/fasttrackml/pkg/api/aim/api/request"

// NormaliseGetProjectParamsRequest normalizes request object for `GET /projects/params` endpoint.
func NormaliseGetProjectParamsRequest(req *request.GetProjectParamsRequest) *request.GetProjectParamsRequest {
	if len(req.Sequences) == 0 {
		req.Sequences = SupportedSequences
	}
	return req
}
