package metric

import (
	"context"
	"database/sql"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/repositories"
)

// Service provides service layer to work with `metric` business logic.
type Service struct {
	runRepository    repositories.RunRepositoryProvider
	metricRepository repositories.MetricRepositoryProvider
}

// NewService creates new Service instance.
func NewService(
	runRepository repositories.RunRepositoryProvider,
	metricRepository repositories.MetricRepositoryProvider,
) *Service {
	return &Service{
		runRepository:    runRepository,
		metricRepository: metricRepository,
	}
}

func (s Service) GetMetricHistory(
	ctx context.Context, namespace *models.Namespace, req *request.GetMetricHistoryRequest,
) ([]models.Metric, error) {
	if err := ValidateGetMetricHistoryRequest(req); err != nil {
		return nil, err
	}

	run, err := s.runRepository.GetByNamespaceIDAndRunID(ctx, namespace.ID, req.GetRunID())
	if err != nil {
		return nil, api.NewInternalError("unable to find run '%s': %s", req.GetRunID(), err)
	}
	if run == nil {
		return nil, api.NewResourceDoesNotExistError("unable to find run '%s'", req.GetRunID())
	}

	metrics, err := s.metricRepository.GetMetricHistoryByRunIDAndKey(ctx, run.ID, req.MetricKey)
	if err != nil {
		return nil, api.NewInternalError(
			"unable to get metric history for metric '%s' of run '%s'", req.MetricKey, req.GetRunID(),
		)
	}

	return metrics, nil
}

func (s Service) GetMetricHistoryBulk(
	ctx context.Context, namespace *models.Namespace, req *request.GetMetricHistoryBulkRequest,
) ([]models.Metric, error) {
	if err := ValidateGetMetricHistoryBulkRequest(req); err != nil {
		return nil, err
	}
	metrics, err := s.metricRepository.GetMetricHistoryBulk(
		ctx,
		namespace.ID,
		req.RunIDs,
		req.MetricKey,
		req.MaxResults,
	)
	if err != nil {
		return nil, api.NewInternalError(
			"unable to get metric history in bulk for metric %q of runs %q", req.MetricKey, req.RunIDs,
		)
	}
	return metrics, nil
}

func (s Service) GetMetricHistories(
	ctx context.Context, namespace *models.Namespace, req *request.GetMetricHistoriesRequest,
) (*sql.Rows, func(*sql.Rows, interface{}) error, error) {
	if err := ValidateGetMetricHistoriesRequest(req); err != nil {
		return nil, nil, err
	}

	rows, iterator, err := s.metricRepository.GetMetricHistories(
		ctx,
		namespace.ID,
		req.ExperimentIDs,
		req.RunIDs,
		req.MetricKeys,
		req.ViewType,
		req.MaxResults,
	)
	if err != nil {
		return nil, nil, api.NewInternalError("Unable to search runs: %s", err)
	}

	return rows, iterator, nil
}
