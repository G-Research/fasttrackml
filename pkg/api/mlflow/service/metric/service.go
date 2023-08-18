package metric

import (
	"context"
	"database/sql"

	"github.com/G-Research/fasttrackml/pkg/common/dao/models"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/repositories"
)

// Service provides service layer to work with `metric` business logic.
type Service struct {
	metricRepository repositories.MetricRepositoryProvider
}

// NewService creates new Service instance.
func NewService(metricRepository repositories.MetricRepositoryProvider) *Service {
	return &Service{
		metricRepository: metricRepository,
	}
}

func (s Service) GetMetricHistory(
	ctx context.Context, req *request.GetMetricHistoryRequest,
) ([]models.Metric, error) {
	if err := ValidateGetMetricHistoryRequest(req); err != nil {
		return nil, err
	}

	metrics, err := s.metricRepository.GetMetricHistoryByRunIDAndKey(ctx, req.GetRunID(), req.MetricKey)
	if err != nil {
		return nil, api.NewInternalError(
			"unable to get metric history for metric '%s' of run '%s'", req.MetricKey, req.GetRunID(),
		)
	}

	return metrics, nil
}

func (s Service) GetMetricHistoryBulk(
	ctx context.Context, req *request.GetMetricHistoryBulkRequest,
) ([]models.Metric, error) {
	if err := ValidateGetMetricHistoryBulkRequest(req); err != nil {
		return nil, err
	}
	metrics, err := s.metricRepository.GetMetricHistoryBulk(ctx, req.RunIDs, req.MetricKey, req.MaxResults)
	if err != nil {
		return nil, api.NewInternalError(
			"unable to get metric history in bulk for metric %q of runs %q", req.MetricKey, req.RunIDs,
		)
	}
	return metrics, nil
}

func (s Service) GetMetricHistories(
	ctx context.Context, req *request.GetMetricHistoriesRequest,
) (*sql.Rows, func(*sql.Rows, interface{}) error, error) {
	if err := ValidateGetMetricHistoriesRequest(req); err != nil {
		return nil, nil, err
	}

	rows, iterator, err := s.metricRepository.GetMetricHistories(
		ctx, req.ExperimentIDs, req.RunIDs, req.MetricKeys, req.ViewType, req.MaxResults,
	)
	if err != nil {
		return nil, nil, api.NewInternalError("Unable to search runs: %s", err)
	}

	return rows, iterator, nil
}
