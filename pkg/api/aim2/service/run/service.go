package run

import (
	"context"

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
	runRepository repositories.RunRepositoryProvider
}

// NewService creates new Service instance.
func NewService(runRepository repositories.RunRepositoryProvider) *Service {
	return &Service{
		runRepository: runRepository,
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
