package run

import (
	"github.com/G-Research/fasttrackml/pkg/api/aim/api/request"
)

// NormaliseGetRunInfoRequest normalizes request object for `GET /runs/:id/info` endpoint.
func NormaliseGetRunInfoRequest(req *request.GetRunInfoRequest) *request.GetRunInfoRequest {
	if len(req.Sequences) == 0 {
		req.Sequences = SupportedSequences
	}
	return req
}
