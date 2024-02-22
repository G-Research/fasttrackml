package project

import (
	"context"
	"slices"
	"time"

	"github.com/G-Research/fasttrackml/pkg/api/aim2/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/dto"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/repositories"
	"github.com/G-Research/fasttrackml/pkg/common/api"
)

// Service provides service layer to work with `project` business logic.
type Service struct {
	tagRepository        repositories.TagRepositoryProvider
	runRepository        repositories.RunRepositoryProvider
	paramRepository      repositories.ParamRepositoryProvider
	metricRepository     repositories.MetricRepositoryProvider
	experimentRepository repositories.ExperimentRepositoryProvider
}

// NewService creates new Service instance.
func NewService(
	tagRepository repositories.TagRepositoryProvider,
	runRepository repositories.RunRepositoryProvider,
	paramRepository repositories.ParamRepositoryProvider,
	metricRepository repositories.MetricRepositoryProvider,
	experimentRepository repositories.ExperimentRepositoryProvider,
) *Service {
	return &Service{
		tagRepository:        tagRepository,
		runRepository:        runRepository,
		paramRepository:      paramRepository,
		metricRepository:     metricRepository,
		experimentRepository: experimentRepository,
	}
}

// GetProjectInformation returns project information.
func (s Service) GetProjectInformation() (string, string) {
	return "FastTrackML", s.runRepository.GetDB().Dialector.Name()
}

// GetProjectActivity returns project activity.
func (s Service) GetProjectActivity(
	ctx context.Context, namespaceID uint, tzOffset int,
) (*dto.ProjectActivity, error) {
	runs, err := s.runRepository.GetByNamespaceID(ctx, namespaceID)
	if err != nil {
		return nil, api.NewInternalError("error getting runs: %s", err)
	}
	activity, numActiveRuns, numArchivedRuns := map[string]int{}, int64(0), int64(0)
	for _, run := range runs {
		switch {
		case run.LifecycleStage == models.LifecycleStageDeleted:
			numArchivedRuns += 1
		case run.Status == models.StatusRunning:
			numActiveRuns += 1
		}
		key := time.UnixMilli(run.StartTime.Int64).Add(time.Duration(-tzOffset) * time.Minute).Format("2006-01-02T15:00:00")
		activity[key] += 1
	}

	numActiveExperiments, err := s.experimentRepository.GetCountOfActiveExperiments(ctx, namespaceID)
	if err != nil {
		return nil, api.NewInternalError("error getting number of active experiments: %s", err)
	}

	return &dto.ProjectActivity{
		NumRuns:         int64(len(runs)),
		ActivityMap:     activity,
		NumActiveRuns:   numActiveRuns,
		NumExperiments:  numActiveExperiments,
		NumArchivedRuns: numArchivedRuns,
	}, nil
}

// GetProjectParams returns project params.
func (s Service) GetProjectParams(
	ctx context.Context, namespaceID uint, req *request.GetProjectParamsRequest,
) (*dto.ProjectParams, error) {
	req = NormaliseGetProjectParamsRequest(req)
	if err := ValidateGetProjectsRequest(req); err != nil {
		return nil, err
	}

	projectParams := dto.ProjectParams{}
	if !req.ExcludeParams {
		paramKeys, err := s.paramRepository.GetParamKeysByParameters(ctx, namespaceID, req.Experiments)
		if err != nil {
			return nil, api.NewInternalError("error getting param keys: %s", err)
		}
		projectParams.ParamKeys = paramKeys

		tagKeys, err := s.tagRepository.GetParamKeysByParameters(ctx, namespaceID, req.Experiments)
		if err != nil {
			return nil, api.NewInternalError("error getting tag keys: %s", err)
		}
		projectParams.TagKeys = tagKeys
	}

	if slices.Contains(req.Sequences, "metric") {
		metrics, err := s.metricRepository.GetLatestMetricsByExperiments(ctx, namespaceID, req.Experiments)
		if err != nil {
			return nil, api.NewInternalError("error getting metrics: %s", err)
		}
		projectParams.Metrics = metrics
	}
	return &projectParams, nil
}
