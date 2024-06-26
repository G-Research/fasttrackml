package tag

import (
	"github.com/G-Research/fasttrackml/pkg/api/aim/api/request"
	"github.com/G-Research/fasttrackml/pkg/common/api"
)

// ValidateCreateTagRequest validates `POST /tags` request.
func ValidateCreateTagRequest(req *request.CreateTagRequest) error {
	if len(req.Name) == 0 {
		return api.NewInvalidParameterValueError("`%s` is not a valid tag name", req.Name)
	}
	return nil
}
