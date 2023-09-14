package namespace

import (
	"regexp"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
)

const namespaceValidationMessage = "namespace code is invalid -- must be 2-12 letters, numbers, dash, or underscore"

// validation rule for namespace code
var validNamespaceCode = regexp.MustCompile(`^[\w\d-_]{2,12}$`)
		
// ValidateNamespace validates namespace code
func ValidateNamespace(code string) error {
	if !validNamespaceCode.Match([]byte(code)) {
		return api.NewInvalidParameterValueError(namespaceValidationMessage)
	}
	return nil
}
