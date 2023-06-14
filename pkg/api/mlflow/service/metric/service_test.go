package metric

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/repositories"
)

func TestService_GetMetricHistory_Ok(t *testing.T) {
	// init repository mocks.
	metricRepository := repositories.MockMetricRepositoryProvider{}
	metricRepository.On(
		"GetMetricHistoryByRunIDAndKey",
		mock.AnythingOfType("*context.emptyCtx"),
		"1",
		"key",
	).Return([]models.Metric{
		{
			Key:       "key",
			Step:      1,
			Value:     1.1,
			Timestamp: 1234567890,
		},
	}, nil)

	// call service under testing.
	service := NewService(&metricRepository)
	metrics, err := service.GetMetricHistory(context.TODO(), &request.GetMetricHistoryRequest{
		RunID:     "1",
		MetricKey: "key",
	})

	// compare results.
	assert.Nil(t, err)
	assert.Equal(t, []models.Metric{
		{
			Key:       "key",
			Step:      1,
			Value:     1.1,
			Timestamp: 1234567890,
		},
	}, metrics)
}
func TestService_GetMetricHistory_Error(t *testing.T) {
	var testData = []struct {
		name    string
		error   *api.ErrorResponse
		request *request.GetMetricHistoryRequest
		service func() *Service
	}{
		{
			name:    "EmptyOrIncorrectRunID",
			error:   api.NewInvalidParameterValueError(`Missing value for required parameter 'run_id'`),
			request: &request.GetMetricHistoryRequest{},
			service: func() *Service {
				metricRepository := repositories.MockMetricRepositoryProvider{}
				return NewService(&metricRepository)
			},
		},
		{
			name:  "EmptyOrIncorrectMetricKey",
			error: api.NewInvalidParameterValueError(`Missing value for required parameter 'metric_key'`),
			request: &request.GetMetricHistoryRequest{
				RunID: "1",
			},
			service: func() *Service {
				metricRepository := repositories.MockMetricRepositoryProvider{}
				return NewService(&metricRepository)
			},
		},
		{
			name:  "GetMetricHistoryDatabaseError",
			error: api.NewInternalError(`unable to get metric history for metric "key" of run "1"`),
			request: &request.GetMetricHistoryRequest{
				RunID:     "1",
				MetricKey: "key",
			},
			service: func() *Service {
				metricRepository := repositories.MockMetricRepositoryProvider{}
				metricRepository.On(
					"GetMetricHistoryByRunIDAndKey",
					mock.AnythingOfType("*context.emptyCtx"),
					"1",
					"key",
				).Return(nil, errors.New("database error"))
				return NewService(&metricRepository)
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			// call service under testing.
			_, err := tt.service().GetMetricHistory(context.TODO(), tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}

func TestService_GetMetricHistoryBulk_Ok(t *testing.T) {
	// init repository mocks.
	metricRepository := repositories.MockMetricRepositoryProvider{}
	metricRepository.On(
		"GetMetricHistoryBulk",
		mock.AnythingOfType("*context.emptyCtx"),
		[]string{"1", "2"},
		"key",
		10,
	).Return([]models.Metric{
		{
			Key:       "key",
			Step:      1,
			Value:     1.1,
			Timestamp: 1234567890,
		},
	}, nil)

	// call service under testing.
	service := NewService(&metricRepository)
	metrics, err := service.GetMetricHistoryBulk(context.TODO(), &request.GetMetricHistoryBulkRequest{
		RunIDs:     []string{"1", "2"},
		MetricKey:  "key",
		MaxResults: 10,
	})

	// compare results.
	assert.Nil(t, err)
	assert.Equal(t, []models.Metric{
		{
			Key:       "key",
			Step:      1,
			Value:     1.1,
			Timestamp: 1234567890,
		},
	}, metrics)
}
func TestService_GetMetricHistoryBulk_Error(t *testing.T) {
	var testData = []struct {
		name    string
		error   *api.ErrorResponse
		request *request.GetMetricHistoryBulkRequest
		service func() *Service
	}{
		{
			name: "EmptyOrIncorrectRunIDs",
			error: api.NewInvalidParameterValueError(
				`GetMetricHistoryBulk request must specify at least one run_id.`,
			),
			request: &request.GetMetricHistoryBulkRequest{},
			service: func() *Service {
				metricRepository := repositories.MockMetricRepositoryProvider{}
				return NewService(&metricRepository)
			},
		},
		{
			name: "NumberOfRunIDsMoreThenAllowed",
			error: api.NewInvalidParameterValueError(
				`GetMetricHistoryBulk request cannot specify more than 200 run_ids. Received 201 run_ids.`,
			),
			request: &request.GetMetricHistoryBulkRequest{
				RunIDs: make([]string, MaxRunIDsForMetricHistoryBulkRequest+1),
			},
			service: func() *Service {
				metricRepository := repositories.MockMetricRepositoryProvider{}
				return NewService(&metricRepository)
			},
		},
		{
			name: "EmptyOrIncorrectMetricKey",
			error: api.NewInvalidParameterValueError(
				`GetMetricHistoryBulk request must specify a metric_key.`,
			),
			request: &request.GetMetricHistoryBulkRequest{
				RunIDs: []string{"1"},
			},
			service: func() *Service {
				metricRepository := repositories.MockMetricRepositoryProvider{}
				return NewService(&metricRepository)
			},
		},
		{
			name:  "GetMetricHistoryBulkDatabaseError",
			error: api.NewInternalError(`unable to get metric history in bulk for metric "key" of runs ["1"]`),
			request: &request.GetMetricHistoryBulkRequest{
				RunIDs:     []string{"1"},
				MetricKey:  "key",
				MaxResults: 10,
			},
			service: func() *Service {
				metricRepository := repositories.MockMetricRepositoryProvider{}
				metricRepository.On(
					"GetMetricHistoryBulk",
					mock.AnythingOfType("*context.emptyCtx"),
					[]string{"1"},
					"key",
					10,
				).Return(nil, errors.New("database error"))
				return NewService(&metricRepository)
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			// call service under testing.
			_, err := tt.service().GetMetricHistoryBulk(context.TODO(), tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}

func TestNewService_GetMetricHistories_Ok(t *testing.T) {
	// init repository mocks.
	metricRepository := repositories.MockMetricRepositoryProvider{}
	metricRepository.On(
		"GetMetricHistories",
		mock.AnythingOfType("*context.emptyCtx"),
		[]string{"1", "2"},
		mock.Anything,
		[]string{"key1", "key2"},
		request.ViewTypeActiveOnly,
		int32(1),
	).Return(
		&sql.Rows{},
		func(*sql.Rows, interface{}) error {
			return nil
		},
		nil,
	)

	// call service under testing.
	service := NewService(&metricRepository)
	rows, iterator, err := service.GetMetricHistories(context.TODO(), &request.GetMetricHistoriesRequest{
		ExperimentIDs: []string{"1", "2"},
		MetricKeys:    []string{"key1", "key2"},
		ViewType:      request.ViewTypeActiveOnly,
		MaxResults:    1,
	})
	assert.Nil(t, err)
	assert.NotNil(t, rows)
	assert.NotNil(t, iterator)
}
func TestNewService_GetMetricHistories_Error(t *testing.T) {
	var testData = []struct {
		name    string
		error   *api.ErrorResponse
		request *request.GetMetricHistoriesRequest
		service func() *Service
	}{
		{
			name: "HasToBeProvidedExperimentDdsOrRunIdsProperty",
			error: api.NewInvalidParameterValueError(
				`experiment_ids and run_ids cannot both be specified at the same time`,
			),
			request: &request.GetMetricHistoriesRequest{
				RunIDs:        []string{"1"},
				ExperimentIDs: []string{"2"},
			},
			service: func() *Service {
				metricRepository := repositories.MockMetricRepositoryProvider{}
				return NewService(&metricRepository)
			},
		},
		{
			name:  "UnsupportedViewType",
			error: api.NewInvalidParameterValueError(`Invalid run_view_type 'unsupported'`),
			request: &request.GetMetricHistoriesRequest{
				RunIDs:   []string{"1"},
				ViewType: "unsupported",
			},
			service: func() *Service {
				metricRepository := repositories.MockMetricRepositoryProvider{}
				return NewService(&metricRepository)
			},
		},
		{
			name:  "EmptyOrIncorrectMaxResults",
			error: api.NewInvalidParameterValueError(`Invalid value for parameter 'max_results' supplied.`),
			request: &request.GetMetricHistoriesRequest{
				RunIDs:     []string{"1"},
				ViewType:   request.ViewTypeAll,
				MaxResults: MaxResultsPerPage + 1,
			},
			service: func() *Service {
				metricRepository := repositories.MockMetricRepositoryProvider{}
				return NewService(&metricRepository)
			},
		},
		{
			name:  "GetGetMetricHistoriesDatabaseError",
			error: api.NewInternalError(`Unable to search runs: database error`),
			request: &request.GetMetricHistoriesRequest{
				RunIDs:     []string{"1"},
				ViewType:   request.ViewTypeAll,
				MetricKeys: []string{"key1", "key2"},
				MaxResults: 1,
			},
			service: func() *Service {
				metricRepository := repositories.MockMetricRepositoryProvider{}
				metricRepository.On(
					"GetMetricHistories",
					mock.AnythingOfType("*context.emptyCtx"),
					mock.Anything,
					[]string{"1"},
					[]string{"key1", "key2"},
					request.ViewTypeAll,
					int32(1),
				).Return(
					nil,
					nil,
					errors.New("database error"),
				)
				return NewService(&metricRepository)
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			// call service under testing.
			_, _, err := tt.service().GetMetricHistories(context.TODO(), tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}
