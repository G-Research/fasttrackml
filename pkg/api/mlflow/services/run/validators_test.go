package run

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/common/api"
)

func TestValidateUpdateRunRequest_Ok(t *testing.T) {
	err := ValidateUpdateRunRequest(&request.UpdateRunRequest{
		RunID:   "id",
		RunUUID: "uuid",
	})
	require.Nil(t, err)
}

func TestValidateUpdateRunRequest_Error(t *testing.T) {
	testData := []struct {
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
	require.Nil(t, err)
}

func TestValidateGetRunRequest_Error(t *testing.T) {
	testData := []struct {
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
	require.Nil(t, err)
}

func TestValidateDeleteRunRequest_Error(t *testing.T) {
	testData := []struct {
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
	require.Nil(t, err)
}

func TestValidateRestoreRunRequest_Error(t *testing.T) {
	testData := []struct {
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
		Timestamp: 123456789,
	})
	require.Nil(t, err)
}

func TestValidateLogMetricRequest_Error(t *testing.T) {
	testData := []struct {
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
	require.Nil(t, err)
}

func TestValidateLogParamRequest_Error(t *testing.T) {
	testData := []struct {
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
	require.Nil(t, err)
}

func TestValidateSetRunTagRequest_Error(t *testing.T) {
	testData := []struct {
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
	require.Nil(t, err)
}

func TestValidateDeleteRunTagRequest_Error(t *testing.T) {
	testData := []struct {
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
	require.Nil(t, err)
}

func TestValidateLogBatchRequest_Error(t *testing.T) {
	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.LogBatchRequest
	}{
		{
			name:    "EmptyRunID",
			error:   api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'"),
			request: &request.LogBatchRequest{},
		},
		{
			name:  "EmptyMetricsKey",
			error: api.NewInvalidParameterValueError("Invalid value for parameter 'metrics' supplied"),
			request: &request.LogBatchRequest{
				RunID: "id",
				Metrics: []request.MetricPartialRequest{
					{
						Key:       "key1",
						Timestamp: 123456789,
						Value:     1.0,
					},
					{
						Timestamp: 123456789,
						Value:     1.0,
					},
				},
			},
		},
		{
			name:  "EmptyMetricsTimestamp",
			error: api.NewInvalidParameterValueError("Invalid value for parameter 'metrics' supplied"),
			request: &request.LogBatchRequest{
				RunID: "id",
				Metrics: []request.MetricPartialRequest{
					{
						Key:       "key1",
						Timestamp: 123456789,
						Value:     1.0,
					},
					{
						Key:   "key2",
						Value: 1.0,
					},
				},
			},
		},
		{
			name:  "EmptyParamsKey",
			error: api.NewInvalidParameterValueError("Invalid value for parameter 'params' supplied"),
			request: &request.LogBatchRequest{
				RunID: "id",
				Params: []request.ParamPartialRequest{
					{
						Key:      "key1",
						ValueStr: common.GetPointer("value1"),
					},
					{
						ValueStr: common.GetPointer("value2"),
					},
				},
			},
		},
		{
			name:  "EmptyTagsKey",
			error: api.NewInvalidParameterValueError("Invalid value for parameter 'tags' supplied"),
			request: &request.LogBatchRequest{
				RunID: "id",
				Tags: []request.TagPartialRequest{
					{
						Key:   "key1",
						Value: "value1",
					},
					{
						Value: "value2",
					},
				},
			},
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
	require.Nil(t, err)
}

func TestValidateSearchRunsRequest_Error(t *testing.T) {
	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.SearchRunsRequest
	}{
		{
			name:  "NotAllowedViewTypeProperty",
			error: api.NewInvalidParameterValueError("Invalid run_view_type 'not-allowed-view-type'"),
			request: &request.SearchRunsRequest{
				ViewType: request.ViewType("not-allowed-view-type"),
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

func TestValidateLogOutputRequest_Ok(t *testing.T) {
	err := ValidateLogOutputRequest(&request.LogOutputRequest{
		RunID: "id",
		Data:  "some log row",
	})
	require.Nil(t, err)
}

func TestValidateLogOutputRequest_Error(t *testing.T) {
	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.LogOutputRequest
	}{
		{
			name:  "EmptyRunID",
			error: api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'"),
			request: &request.LogOutputRequest{
				Data: "Some log data",
			},
		},
		{
			name:  "EmptyData",
			error: api.NewInvalidParameterValueError("Missing value for required parameter 'data'"),
			request: &request.LogOutputRequest{
				RunID: "id",
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLogOutputRequest(tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}
