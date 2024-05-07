package project

import (
	"slices"

	"github.com/G-Research/fasttrackml/pkg/api/aim/api/request"
	"github.com/G-Research/fasttrackml/pkg/common/api"
)

// SupportedSequences list of supported Sequences for `GET /projects/params` request.
var SupportedSequences = []string{
	"metric",
	"images",
	"texts",
	"figures",
	"distributions",
	"audios",
}

// ValidateGetProjectsRequest validates `GET /projects/params` request.
func ValidateGetProjectsRequest(req *request.GetProjectParamsRequest) error {
	for _, sequence := range req.Sequences {
		if !slices.Contains(SupportedSequences, sequence) {
			return api.NewInvalidParameterValueError("%q is not a valid Sequence", sequence)
		}
	}
	return nil
}
