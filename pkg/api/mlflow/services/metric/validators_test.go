package metric

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/common/api"
)

func TestValidateGetMetricHistoryRequest_Ok(t *testing.T) {
	err := ValidateGetMetricHistoryRequest(&request.GetMetricHistoryRequest{
		RunID:     "id",
		RunUUID:   "uuid",
		MetricKey: "key",
	})
	require.Nil(t, err)
}

func TestValidateGetMetricHistoryRequest_Error(t *testing.T) {
	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.GetMetricHistoryRequest
	}{
		{
			name:    "EmptyRunIDAndRunUUID",
			error:   api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'"),
			request: &request.GetMetricHistoryRequest{},
		},
		{
			name:  "EmptyMetricKeyProperty",
			error: api.NewInvalidParameterValueError("Missing value for required parameter 'metric_key'"),
			request: &request.GetMetricHistoryRequest{
				RunID: "id",
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGetMetricHistoryRequest(tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}

func TestValidateGetMetricHistoryBulkRequest_Ok(t *testing.T) {
	err := ValidateGetMetricHistoryBulkRequest(&request.GetMetricHistoryBulkRequest{
		RunIDs:     []string{"id1", "id2"},
		MetricKey:  "key",
		MaxResults: 10,
	})
	require.Nil(t, err)
}

func TestValidateGetMetricHistoryBulkRequest_Error(t *testing.T) {
	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.GetMetricHistoryBulkRequest
	}{
		{
			name:    "EmptyRunIDsProperty",
			error:   api.NewInvalidParameterValueError("GetMetricHistoryBulk request must specify at least one run_id."),
			request: &request.GetMetricHistoryBulkRequest{},
		},
		{
			name: "IncorrectSizeOfRunIDsProperty",
			error: api.NewInvalidParameterValueError(
				"GetMetricHistoryBulk request cannot specify more than 200 run_ids. Received 201 run_ids.",
			),
			request: &request.GetMetricHistoryBulkRequest{
				RunIDs: make([]string, 201),
			},
		},
		{
			name:  "EmptyMetricKeyProperty",
			error: api.NewInvalidParameterValueError("GetMetricHistoryBulk request must specify a metric_key."),
			request: &request.GetMetricHistoryBulkRequest{
				RunIDs: []string{"id1"},
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGetMetricHistoryBulkRequest(tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}

func TestValidateGetMetricHistoriesRequest_Ok(t *testing.T) {
	err := ValidateGetMetricHistoriesRequest(&request.GetMetricHistoriesRequest{
		RunIDs:     []string{"id1"},
		ViewType:   "",
		MaxResults: 10,
	})
	require.Nil(t, err)
}

func TestValidateGetMetricHistoriesRequest_Error(t *testing.T) {
	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.GetMetricHistoriesRequest
	}{
		{
			name: "RunIDsAndExperimentIDsPopulatedAtTheSameTime",
			error: api.NewInvalidParameterValueError(
				"experiment_ids and run_ids cannot both be specified at the same time",
			),
			request: &request.GetMetricHistoriesRequest{
				RunIDs:        []string{"id1"},
				ExperimentIDs: []string{"id1"},
			},
		},
		{
			name:  "IncorrectViewTypeProperty",
			error: api.NewInvalidParameterValueError("Invalid run_view_type 'incorrect_value'"),
			request: &request.GetMetricHistoriesRequest{
				RunIDs:   []string{"id1"},
				ViewType: "incorrect_value",
			},
		},
		{
			name:  "IncorrectMaxResultsProperty",
			error: api.NewInvalidParameterValueError("Invalid value for parameter 'max_results' supplied."),
			request: &request.GetMetricHistoriesRequest{
				RunIDs:     []string{"id1"},
				MaxResults: MaxResultsForMetricHistoriesRequest + 1,
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGetMetricHistoriesRequest(tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}
