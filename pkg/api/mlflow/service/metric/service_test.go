package metric

import (
	"context"
	"testing"

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
func TestService_GetMetricHistory_Error(t *testing.T) {}

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
func TestService_GetMetricHistoryBulk_Error(t *testing.T) {}
