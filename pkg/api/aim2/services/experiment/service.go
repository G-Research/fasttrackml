package experiment

import (
	"context"

	"github.com/G-Research/fasttrackml/pkg/api/aim2/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/common"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/convertors"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/repositories"
	"github.com/G-Research/fasttrackml/pkg/common/api"
)

// Service provides service layer to work with `experiment` business logic.
type Service struct {
	tagRepository        repositories.TagRepositoryProvider
	experimentRepository repositories.ExperimentRepositoryProvider
}

// NewService creates new Service instance.
func NewService(
	tagRepository repositories.TagRepositoryProvider,
	experimentRepository repositories.ExperimentRepositoryProvider,
) *Service {
	return &Service{
		tagRepository:        tagRepository,
		experimentRepository: experimentRepository,
	}
}

// GetExperiment returns requested experiment.
func (s Service) GetExperiment(
	ctx context.Context, namespaceID uint, req *request.GetExperimentRequest,
) (*models.ExperimentExtended, error) {
	experiment, err := s.experimentRepository.GetExtendedExperimentByNamespaceIDAndExperimentID(
		ctx, namespaceID, req.ID,
	)
	if err != nil {
		return nil, api.NewInternalError("unable to find experiment by id %d: %s", req.ID, err)
	}
	if experiment == nil {
		return nil, api.NewResourceDoesNotExistError("experiment '%d' not found", req.ID)
	}
	return experiment, nil
}

// GetExperiments returns the list of experiments.
func (s Service) GetExperiments(ctx context.Context, namespaceID uint) ([]models.ExperimentExtended, error) {
	experiments, err := s.experimentRepository.GetExperiments(ctx, namespaceID)
	if err != nil {
		return nil, api.NewInternalError("unable to find experiments: %s", err)
	}
	return experiments, nil
}

// GetExperimentActivity returns experiment activity.
func (s Service) GetExperimentActivity(
	ctx context.Context, namespaceID uint, req *request.GetExperimentActivityRequest, tzOffset int,
) (*models.ExperimentActivity, error) {
	experiment, err := s.experimentRepository.GetExperimentByNamespaceIDAndExperimentID(ctx, namespaceID, req.ID)
	if err != nil {
		return nil, api.NewInternalError("unable to find experiment by id %d: %s", req.ID, err)
	}
	if experiment == nil {
		return nil, api.NewResourceDoesNotExistError("experiment '%d' not found", req.ID)
	}

	activity, err := s.experimentRepository.GetExperimentActivity(ctx, namespaceID, *experiment.ID, tzOffset)
	if err != nil {
		return nil, api.NewInternalError("unable to get experiment activity: %s", err)
	}
	return activity, nil
}

// GetExperimentRuns returns list of runs related to requested experiment.
func (s Service) GetExperimentRuns(
	ctx context.Context, namespaceID uint, req *request.GetExperimentRunsRequest,
) ([]models.Run, error) {
	experiment, err := s.experimentRepository.GetExperimentByNamespaceIDAndExperimentID(ctx, namespaceID, req.ID)
	if err != nil {
		return nil, api.NewInternalError("unable to find experiment by id %d: %s", req.ID, err)
	}
	if experiment == nil {
		return nil, api.NewResourceDoesNotExistError("experiment '%d' not found", req.ID)
	}
	runs, err := s.experimentRepository.GetExperimentRuns(ctx, req)
	if err != nil {
		return nil, api.NewInternalError("unable to find experiment runs")
	}
	return runs, nil
}

// UpdateExperiment updates existing experiment.
func (s Service) UpdateExperiment(
	ctx context.Context, namespaceID uint, req *request.UpdateExperimentRequest,
) error {
	experiment, err := s.experimentRepository.GetExperimentByNamespaceIDAndExperimentID(ctx, namespaceID, req.ID)
	if err != nil {
		return api.NewInternalError("unable to find experiment by id %d: %s", req.ID, err)
	}
	if experiment == nil {
		return api.NewResourceDoesNotExistError("experiment '%d' not found", req.ID)
	}

	experiment = convertors.ConvertUpdateExperimentToDBModel(req, experiment)
	if req.Archived != nil || req.Name != nil {
		if err := s.experimentRepository.Update(ctx, experiment); err != nil {
			return api.NewInternalError("unable to update experiment %q: %s", req.ID, err)
		}
	}
	if req.Description != nil {
		if err := s.tagRepository.CreateExperimentTag(ctx, &models.ExperimentTag{
			Key:          common.DescriptionTagKey,
			Value:        *req.Description,
			ExperimentID: *experiment.ID,
		}); err != nil {
			return api.NewInternalError("unable to create experiment tag: %s", err)
		}
	}
	return nil
}

// DeleteExperiment deletes existing experiment.
func (s Service) DeleteExperiment(
	ctx context.Context, namespaceID uint, namespaceDefaultExperimentID *int32, req *request.DeleteExperimentRequest,
) error {
	experiment, err := s.experimentRepository.GetExperimentByNamespaceIDAndExperimentID(ctx, namespaceID, req.ID)
	if err != nil {
		return api.NewInternalError("unable to find experiment by id %d: %s", req.ID, err)
	}
	if experiment == nil {
		return api.NewResourceDoesNotExistError("experiment '%d' not found", req.ID)
	}

	if experiment.IsDefault(namespaceDefaultExperimentID) {
		return api.NewBadRequestError("unable to delete default experiment")
	}

	if err := s.experimentRepository.Delete(ctx, experiment); err != nil {
		return api.NewInternalError("unable to delete experiment by id %d: %s", req.ID, err)
	}

	return nil
}
