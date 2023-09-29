package artifact

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
)

func TestValidateListArtifactsRequest_Ok(t *testing.T) {
	tests := []struct {
		name    string
		request request.ListArtifactsRequest
	}{
		{
			name: "NotEmptyPathCase1",
			request: request.ListArtifactsRequest{
				RunID: "run_id",
				Path:  "foo/..asd../",
			},
		},
		{
			name: "NotEmptyPathCase2",
			request: request.ListArtifactsRequest{
				RunID: "run_id",
				Path:  "./foo",
			},
		},
		{
			name: "NotEmptyPathCase3",
			request: request.ListArtifactsRequest{
				RunID: "run_id",
				Path:  "./foo/",
			},
		},
		{
			name: "NotEmptyPathCase4",
			request: request.ListArtifactsRequest{
				RunID: "run_id",
				Path:  ".foo",
			},
		},
		{
			name: "NotEmptyPathCase5",
			request: request.ListArtifactsRequest{
				RunID: "run_id",
				Path:  "foo.bar",
			},
		},
		{
			name: "NotEmptyPathCase6",
			request: request.ListArtifactsRequest{
				RunID: "run_id",
				Path:  "foo..bar",
			},
		},
		{
			name: "NotEmptyPathCase7",
			request: request.ListArtifactsRequest{
				RunID: "run_id",
				Path:  "foo../bar",
			},
		},
		{
			name: "NotEmptyPathCase8",
			request: request.ListArtifactsRequest{
				RunID: "run_id",
				Path:  "foo/..bar",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Nil(t, ValidateListArtifactsRequest(&tt.request))
		})
	}
}

func TestValidateListArtifactsRequest_Error(t *testing.T) {
	tests := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.ListArtifactsRequest
	}{
		{
			name:    "EmptyRunIDAndRunUUID",
			error:   api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'"),
			request: &request.ListArtifactsRequest{},
		},
		{
			name:  "IncorrectPathProvidedCase1",
			error: api.NewInvalidParameterValueError("provided 'path' parameter is invalid"),
			request: &request.ListArtifactsRequest{
				RunID: "run_id",
				Path:  "..",
			},
		},
		{
			name:  "IncorrectPathProvidedCase2",
			error: api.NewInvalidParameterValueError("provided 'path' parameter is invalid"),
			request: &request.ListArtifactsRequest{
				RunID: "run_id",
				Path:  "./..",
			},
		},
		{
			name:  "IncorrectPathProvidedCase3",
			error: api.NewInvalidParameterValueError("provided 'path' parameter is invalid"),
			request: &request.ListArtifactsRequest{
				RunID: "run_id",
				Path:  "./../",
			},
		},
		{
			name:  "IncorrectPathProvidedCase4",
			error: api.NewInvalidParameterValueError("provided 'path' parameter is invalid"),
			request: &request.ListArtifactsRequest{
				RunID: "run_id",
				Path:  "foo/../bar",
			},
		},
		{
			name:  "IncorrectPathProvidedCase5",
			error: api.NewInvalidParameterValueError("provided 'path' parameter is invalid"),
			request: &request.ListArtifactsRequest{
				RunID: "run_id",
				Path:  "/foo/../bar",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateListArtifactsRequest(tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}

func TestValidateGetArtifactRequest_Ok(t *testing.T) {
	tests := []struct {
		name    string
		request request.GetArtifactRequest
	}{
		{
			name: "NotEmptyPathCase1",
			request: request.GetArtifactRequest{
				RunID: "run_id",
				Path:  "foo/..asd../",
			},
		},
		{
			name: "NotEmptyPathCase2",
			request: request.GetArtifactRequest{
				RunID: "run_id",
				Path:  "./foo",
			},
		},
		{
			name: "NotEmptyPathCase3",
			request: request.GetArtifactRequest{
				RunID: "run_id",
				Path:  "./foo/",
			},
		},
		{
			name: "NotEmptyPathCase4",
			request: request.GetArtifactRequest{
				RunID: "run_id",
				Path:  ".foo",
			},
		},
		{
			name: "NotEmptyPathCase5",
			request: request.GetArtifactRequest{
				RunID: "run_id",
				Path:  "foo.bar",
			},
		},
		{
			name: "NotEmptyPathCase6",
			request: request.GetArtifactRequest{
				RunID: "run_id",
				Path:  "foo..bar",
			},
		},
		{
			name: "NotEmptyPathCase7",
			request: request.GetArtifactRequest{
				RunID: "run_id",
				Path:  "foo../bar",
			},
		},
		{
			name: "NotEmptyPathCase8",
			request: request.GetArtifactRequest{
				RunID: "run_id",
				Path:  "foo/..bar",
			},
		},
		{
			name: "RunUUIDAndNotEmptyPath",
			request: request.GetArtifactRequest{
				RunUUID: "6766d949-07e8-4fff-b3d2-b5f820007cbe",
				Path:    "foo/..bar",
			},
		},
		{
			name: "FilenameNoSlash",
			request: request.GetArtifactRequest{
				RunID: "run_id",
				Path:  "foo.txt",
			},
		},
		{
			name: "Dirname/Filename",
			request: request.GetArtifactRequest{
				RunID: "directory/run_id",
				Path:  "foo.txt",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Nil(t, ValidateGetArtifactRequest(&tt.request))
		})
	}
}

func TestValidateGetArtifactRequest_Error(t *testing.T) {
	tests := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.GetArtifactRequest
	}{
		{
			name:    "EmptyRunIDAndRunUUID",
			error:   api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'"),
			request: &request.GetArtifactRequest{},
		},
		{
			name:  "IncorrectPathProvidedCase1",
			error: api.NewInvalidParameterValueError("provided 'path' parameter is invalid"),
			request: &request.GetArtifactRequest{
				RunID: "run_id",
				Path:  "..",
			},
		},
		{
			name:  "IncorrectPathProvidedCase2",
			error: api.NewInvalidParameterValueError("provided 'path' parameter is invalid"),
			request: &request.GetArtifactRequest{
				RunID: "run_id",
				Path:  "./..",
			},
		},
		{
			name:  "IncorrectPathProvidedCase3",
			error: api.NewInvalidParameterValueError("provided 'path' parameter is invalid"),
			request: &request.GetArtifactRequest{
				RunID: "run_id",
				Path:  "./../",
			},
		},
		{
			name:  "IncorrectPathProvidedCase4",
			error: api.NewInvalidParameterValueError("provided 'path' parameter is invalid"),
			request: &request.GetArtifactRequest{
				RunID: "run_id",
				Path:  "foo/../bar",
			},
		},
		{
			name:  "IncorrectPathProvidedCase5",
			error: api.NewInvalidParameterValueError("provided 'path' parameter is invalid"),
			request: &request.GetArtifactRequest{
				RunID: "run_id",
				Path:  "/foo/../bar",
			},
		},
		{
			name:  "IncorrectLeadingSlash",
			error: api.NewInvalidParameterValueError("provided 'path' parameter is invalid"),
			request: &request.GetArtifactRequest{
				RunID: "run_id",
				Path:  "/foo.bar",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGetArtifactRequest(tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}
