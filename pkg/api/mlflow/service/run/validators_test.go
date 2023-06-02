package run

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
)

func TestValidateUpdateRunRequest_Ok(t *testing.T) {
	err := ValidateUpdateRunRequest(&request.UpdateRunRequest{
		RunID:   "id",
		RunUUID: "uuid",
	})
	assert.Nil(t, err)
}
func TestValidateUpdateRunRequest_Error(t *testing.T) {
	var testData = []struct {
		name    string
		error   *api.ErrorResponse
		request *request.UpdateRunRequest
	}{
		{
			name:    "EmptyRunIDAndRunUUID",
			error:   api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'"),
			request: &request.UpdateRunRequest{},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUpdateRunRequest(tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}

func TestValidateGetRunRequest_Ok(t *testing.T) {
	err := ValidateGetRunRequest(&request.GetRunRequest{
		RunID: "id",
	})
	assert.Nil(t, err)
}
func TestValidateGetRunRequest_Error(t *testing.T) {
	var testData = []struct {
		name    string
		error   *api.ErrorResponse
		request *request.GetRunRequest
	}{
		{
			name:    "EmptyRunIDAndRunUUID",
			error:   api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'"),
			request: &request.GetRunRequest{},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGetRunRequest(tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}

func TestValidateDeleteRunRequest_Ok(t *testing.T) {
	err := ValidateDeleteRunRequest(&request.DeleteRunRequest{
		RunID: "id",
	})
	assert.Nil(t, err)
}
func TestValidateDeleteRunRequest_Error(t *testing.T) {
	var testData = []struct {
		name    string
		error   *api.ErrorResponse
		request *request.DeleteRunRequest
	}{
		{
			name:    "EmptyRunID",
			error:   api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'"),
			request: &request.DeleteRunRequest{},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDeleteRunRequest(tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}

func TestValidateRestoreRunRequest_Ok(t *testing.T) {
	err := ValidateRestoreRunRequest(&request.RestoreRunRequest{
		RunID: "id",
	})
	assert.Nil(t, err)
}
func TestValidateRestoreRunRequest_Error(t *testing.T) {
	var testData = []struct {
		name    string
		error   *api.ErrorResponse
		request *request.RestoreRunRequest
	}{
		{
			name:    "EmptyRunID",
			error:   api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'"),
			request: &request.RestoreRunRequest{},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRestoreRunRequest(tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}

func TestValidateLogMetricRequest_Ok(t *testing.T) {
	err := ValidateLogMetricRequest(&request.LogMetricRequest{
		RunID:     "id",
		RunUUID:   "uuid",
		Key:       "key",
		Timestamp: 123,
	})
	assert.Nil(t, err)
}
func TestValidateLogMetricRequest_Error(t *testing.T) {
	var testData = []struct {
		name    string
		error   *api.ErrorResponse
		request *request.LogMetricRequest
	}{
		{
			name:    "EmptyRunIDAndRunUUID",
			error:   api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'"),
			request: &request.LogMetricRequest{},
		},
		{
			name:  "EmptyKey",
			error: api.NewInvalidParameterValueError("Missing value for required parameter 'key'"),
			request: &request.LogMetricRequest{
				RunID: "id",
			},
		},
		{
			name:  "EmptyTimestamp",
			error: api.NewInvalidParameterValueError("Missing value for required parameter 'timestamp'"),
			request: &request.LogMetricRequest{
				RunID: "id",
				Key:   "key",
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLogMetricRequest(tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}

func TestValidateLogParamRequest_Ok(t *testing.T) {
	err := ValidateLogParamRequest(&request.LogParamRequest{
		RunID:   "id",
		RunUUID: "uuid",
		Key:     "key",
	})
	assert.Nil(t, err)
}
func TestValidateLogParamRequest_Error(t *testing.T) {
	var testData = []struct {
		name    string
		error   *api.ErrorResponse
		request *request.LogParamRequest
	}{
		{
			name:    "EmptyRunIDAndRunUUID",
			error:   api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'"),
			request: &request.LogParamRequest{},
		},
		{
			name:  "EmptyKey",
			error: api.NewInvalidParameterValueError("Missing value for required parameter 'key'"),
			request: &request.LogParamRequest{
				RunID: "id",
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLogParamRequest(tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}

func TestValidateSetRunTagRequest_Ok(t *testing.T) {
	err := ValidateSetRunTagRequest(&request.SetRunTagRequest{
		RunID:   "id",
		RunUUID: "uuid",
		Key:     "key",
	})
	assert.Nil(t, err)
}
func TestValidateSetRunTagRequest_Error(t *testing.T) {
	var testData = []struct {
		name    string
		error   *api.ErrorResponse
		request *request.SetRunTagRequest
	}{
		{
			name:    "EmptyRunIDAndRunUUID",
			error:   api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'"),
			request: &request.SetRunTagRequest{},
		},
		{
			name:  "EmptyKey",
			error: api.NewInvalidParameterValueError("Missing value for required parameter 'key'"),
			request: &request.SetRunTagRequest{
				RunID: "id",
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSetRunTagRequest(tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}

func TestValidateDeleteRunTagRequest_Ok(t *testing.T) {
	err := ValidateDeleteRunTagRequest(&request.DeleteRunTagRequest{
		RunID: "id",
	})
	assert.Nil(t, err)
}
func TestValidateDeleteRunTagRequest_Error(t *testing.T) {
	var testData = []struct {
		name    string
		error   *api.ErrorResponse
		request *request.DeleteRunTagRequest
	}{
		{
			name:    "EmptyRunID",
			error:   api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'"),
			request: &request.DeleteRunTagRequest{},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDeleteRunTagRequest(tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}

func TestValidateLogBatchRequest_Ok(t *testing.T) {
	err := ValidateLogBatchRequest(&request.LogBatchRequest{
		RunID: "id",
	})
	assert.Nil(t, err)
}
func TestValidateLogBatchRequest_Error(t *testing.T) {
	var testData = []struct {
		name    string
		error   *api.ErrorResponse
		request *request.LogBatchRequest
	}{
		{
			name:    "EmptyRunID",
			error:   api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'"),
			request: &request.LogBatchRequest{},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLogBatchRequest(tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}

func TestValidateSearchRunsRequest_Ok(t *testing.T) {
	err := ValidateSearchRunsRequest(&request.SearchRunsRequest{
		ViewType:   request.ViewTypeAll,
		MaxResults: 10,
	})
	assert.Nil(t, err)
}
func TestValidateSearchRunsRequest_Error(t *testing.T) {

	var testData = []struct {
		name    string
		error   *api.ErrorResponse
		request *request.SearchRunsRequest
	}{
		{
			name:  "NotAllowedViewTypeProperty",
			error: api.NewInvalidParameterValueError("Invalid run_view_type 'not-allowed-view-type'"),
			request: &request.SearchRunsRequest{
				ViewType: request.ViewType(`not-allowed-view-type`),
			},
		},
		{
			name:  "IncorrectMaxResultsProperty",
			error: api.NewInvalidParameterValueError("Invalid value for parameter 'max_results' supplied."),
			request: &request.SearchRunsRequest{
				ViewType:   request.ViewTypeAll,
				MaxResults: MaxResultsPerPage + 1,
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSearchRunsRequest(tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}
