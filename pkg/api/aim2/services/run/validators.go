package run

import (
	"slices"

	"github.com/G-Research/fasttrackml/pkg/api/aim2/api/request"
	"github.com/G-Research/fasttrackml/pkg/common/api"
)

// SupportedSequences list of supported Sequences for `GET /runs/:id/info` request.
var SupportedSequences = []string{
	"audios",
	"distributions",
	"figures",
	"images",
	"log_records",
	"logs",
	"texts",
	"metric",
}

// ValidateGetRunInfoRequest validates `GET /runs/:id/info` request.
func ValidateGetRunInfoRequest(req *request.GetRunInfoRequest) error {
	for _, sequence := range req.Sequences {
		if !slices.Contains(SupportedSequences, sequence) {
			return api.NewInvalidParameterValueError("%q is not a valid Sequence", sequence)
		}
	}
	return nil
}

func ValidateGetSearchRunsRequest(req *request.SearchRunsRequest) error {
	if len(req.ExperimentNames) == 0 {
		return api.NewBadRequestError("required `experiment_names` is missing")
	}
	return nil
}
