package run

import (
	"context"
	"database/sql"

	"github.com/G-Research/fasttrackml/pkg/api/aim2/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/dto"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/models"
	aimModels "github.com/G-Research/fasttrackml/pkg/api/aim2/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/repositories"
	mlflowModels "github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/common/api"
)

// Service provides service layer to work with `run` business logic.
type Service struct {
	runRepository    repositories.RunRepositoryProvider
	metricRepository repositories.MetricRepositoryProvider
}

// NewService creates new Service instance.
func NewService(runRepository repositories.RunRepositoryProvider,
	metricRepository repositories.MetricRepositoryProvider) *Service {
	return &Service{
		runRepository:    runRepository,
		metricRepository: metricRepository,
	}
}

// GetRunInfo returns run info.
func (s Service) GetRunInfo(
	ctx context.Context, namespace *mlflowModels.Namespace, req *request.GetRunInfoRequest,
) (*aimModels.Run, error) {
	req = NormaliseGetRunInfoRequest(req)
	if err := ValidateGetRunInfoRequest(req); err != nil {
		return nil, err
	}

	runInfo, err := s.runRepository.GetRunInfo(ctx, namespace.ID, req)
	if err != nil {
		return nil, api.NewInternalError("unable to find run by id %s: %s", req.ID, err)
	}
	if runInfo == nil {
		return nil, api.NewResourceDoesNotExistError("run '%s' not found", req.ID)
	}

	return runInfo, nil
}

// GetRunMetrics returns run metrics.
func (s Service) GetRunMetrics(
	ctx context.Context, namespace *mlflowModels.Namespace, runID string, req *request.GetRunMetricsRequest,
) ([]models.Metric, dto.MetricKeysMapDTO, error) {
	run, err := s.runRepository.GetRunByNamespaceIDAndRunID(ctx, namespace.ID, runID)
	if err != nil {
		return nil, nil, api.NewInternalError("error getting run by id %s: %s", runID, err)
	}

	if run == nil {
		return nil, nil, api.NewResourceDoesNotExistError("run '%s' not found", runID)
	}

	metricKeysMap, err := ConvertRunMetricsRequestToMetricKeysMapDTO(req)
	if err != nil {
		return nil, nil, api.NewBadRequestError("unable to convert request: %s", err)
	}
	metrics, err := s.runRepository.GetRunMetrics(ctx, runID, metricKeysMap)
	if err != nil {
		return nil, nil, api.NewInternalError("error getting run metrics by id %s: %s", runID, err)
	}

	return metrics, metricKeysMap, nil
}

// GetRunsActive returns the active runs.
func (s Service) GetRunsActive(
	ctx context.Context, namespace *mlflowModels.Namespace, req *request.GetRunsActiveRequest,
) ([]models.Run, error) {
	runs, err := s.runRepository.GetByNamespaceIDAndStatus(ctx, namespace.ID, aimModels.StatusRunning)
	if err != nil {
		return nil, api.NewInternalError("error ative runs: %s", err)
	}
	return runs, nil
}

// SearchRuns returns the list of runs by provided search criteria.
func (s Service) SearchRuns(
	ctx context.Context, req request.SearchRunsRequest,
) ([]models.Run, int64, error) {
	runs, total, err := s.runRepository.SearchRuns(ctx, req)
	if err != nil {
		return nil, 0, api.NewInternalError("error searching runs: %s", err)
	}
	return runs, total, nil
}

// SearchMetrics returns the list of metrics by provided search criteria.
func (s Service) SearchMetrics(
	ctx context.Context, req request.SearchMetricsRequest,
) (*sql.Rows, int64, error) {
	rows, total, err := s.metricRepository.SearchMetrics(ctx, req)
	if err != nil {
		return nil, 0, api.NewInternalError("error searching runs: %s", err)
	}
	return rows, total, nil
}
