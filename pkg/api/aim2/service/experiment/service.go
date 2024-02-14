package experiment

import (
	"context"

	"github.com/G-Research/fasttrackml/pkg/api/aim2/api/request"
	aimModels "github.com/G-Research/fasttrackml/pkg/api/aim2/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/repositories"
	mlflowModels "github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/common/api"
)

// Service provides service layer to work with `experiment` business logic.
type Service struct {
	experimentRepository repositories.ExperimentRepositoryProvider
}

// NewService creates new Service instance.
func NewService(experimentRepository repositories.ExperimentRepositoryProvider) *Service {
	return &Service{
		experimentRepository: experimentRepository,
	}
}

// GetExperiment returns requested experiment.
func (s Service) GetExperiment(
	ctx context.Context, namespace *mlflowModels.Namespace, req *request.GetExperimentRequest,
) (*aimModels.ExperimentExtended, error) {
	experiment, err := s.experimentRepository.GetExperimentByNamespaceIDAndExperimentID(
		ctx, namespace.ID, req.ID,
	)
	if err != nil {
		return nil, api.NewInternalError("unable to find experiment by id %q: %s", req.ID, err)
	}
	if experiment == nil {
		return nil, api.NewResourceDoesNotExistError("experiment '%s' not found", req.ID)
	}
	return experiment, nil
}

// GetExperiments returns the list of experiments.
func (s Service) GetExperiments(
	ctx context.Context, namespace *mlflowModels.Namespace,
) ([]aimModels.ExperimentExtended, error) {
	experiments, err := s.experimentRepository.GetExperiments(ctx, namespace.ID)
	if err != nil {
		return nil, api.NewInternalError("unable to find experiments: %s", err)
	}
	return experiments, nil
}

// GetExperimentActivity returns experiment activity.
func (s Service) GetExperimentActivity(
	ctx context.Context, namespace *mlflowModels.Namespace, req *request.GetExperimentActivityRequest, tzOffset int,
) (*aimModels.ExperimentActivity, error) {
	activity, err := s.experimentRepository.GetExperimentActivity(ctx, namespace.ID, req.ID, tzOffset)
	if err != nil {
		return nil, api.NewInternalError("unable to get experiment activity: %s", err)
	}
	return activity, nil
}

// GetExperimentRuns returns list of runs related to requested experiment.
func (s Service) GetExperimentRuns(
	ctx context.Context, namespace *mlflowModels.Namespace, req *request.GetExperimentRunsRequest,
) ([]aimModels.Run, error) {
	runs, err := s.experimentRepository.GetExperimentRuns(ctx, namespace.ID, req)
	if err != nil {
		return nil, api.NewInternalError("unable to find experiment runs")
	}
	return runs, nil
}
