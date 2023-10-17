package namespace

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
)

func TestValidateUpdateRunRequest_Ok(t *testing.T) {
	err := ValidateNamespace("legit-123_ns")
	assert.Nil(t, err)
}

func TestValidateUpdateRunRequest_Error(t *testing.T) {
	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request string
	}{
		{
			name:    "TooShort",
			error:   api.NewInvalidParameterValueError(namespaceValidationMessage),
			request: "1",
		},
		{
			name:    "TooLong",
			error:   api.NewInvalidParameterValueError(namespaceValidationMessage),
			request: "123456789101112",
		},
		{
			name:    "NoSpaces",
			error:   api.NewInvalidParameterValueError(namespaceValidationMessage),
			request: "1234 567",
		},
		{
			name:    "NoExtraChars",
			error:   api.NewInvalidParameterValueError(namespaceValidationMessage),
			request: "1234+(*&^56",
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNamespace(tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}
