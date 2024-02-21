package project

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/api/aim2/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/dto"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/repositories"
	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/pkg/database"
)

// Service provides service layer to work with `project` business logic.
type Service struct {
	tagRepository        repositories.TagRepositoryProvider
	runRepository        repositories.RunRepositoryProvider
	paramRepository      repositories.ParamRepositoryProvider
	experimentRepository repositories.ExperimentRepositoryProvider
}

// NewService creates new Service instance.
func NewService(
	tagRepository repositories.TagRepositoryProvider,
	runRepository repositories.RunRepositoryProvider,
	paramRepository repositories.ParamRepositoryProvider,
	experimentRepository repositories.ExperimentRepositoryProvider,
) *Service {
	return &Service{
		tagRepository:        tagRepository,
		runRepository:        runRepository,
		paramRepository:      paramRepository,
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

func (s Service) GetProjectParams(ctx context.Context, namespaceID uint, req *request.GetProjectParamsRequest) error {
	resp := fiber.Map{}
	if !req.ExcludeParams {
		paramKeys, err := s.paramRepository.GetParamKeysByParameters(ctx, namespaceID, req.Experiments)
		if err != nil {
			return api.NewInternalError("error getting param keys: %s", err)
		}

		params := make(map[string]any, len(paramKeys)+1)
		for _, p := range paramKeys {
			params[p] = map[string]string{
				"__example_type__": "<class 'str'>",
			}
		}

		tagKeys, err := s.tagRepository.GetParamKeysByParameters(ctx, namespaceID, req.Experiments)
		if err != nil {
			return api.NewInternalError("error getting tag keys: %s", err)
		}

		tags := make(map[string]map[string]string, len(tagKeys))
		for _, t := range tagKeys {
			tags[t] = map[string]string{
				"__example_type__": "<class 'str'>",
			}
		}

		params["tags"] = tags
		resp["params"] = params
	}

	if len(req.Sequences) == 0 {
		req.Sequences = []string{
			"metric",
			"images",
			"texts",
			"figures",
			"distributions",
			"audios",
		}
	}

	for _, s := range req.Sequences {
		switch s {
		case "images", "texts", "figures", "distributions", "audios":
			resp[s] = fiber.Map{}
		case "metric":
			query := database.DB.Distinct().Model(
				&database.LatestMetric{},
			).Joins(
				"JOIN runs USING(run_uuid)",
			).Joins(
				"INNER JOIN experiments ON experiments.experiment_id = runs.experiment_id AND experiments.namespace_id = ?",
				ns.ID,
			).Joins(
				"Context",
			).Where(
				"runs.lifecycle_stage = ?", database.LifecycleStageActive,
			)
			if len(req.Experiments) != 0 {
				query.Where("experiments.experiment_id IN ?", req.Experiments)
			}
			var metrics []database.LatestMetric
			if err = query.Find(&metrics).Error; err != nil {
				return fmt.Errorf("error retrieving metric keys: %w", err)
			}

			data, mapped := make(map[string][]fiber.Map, len(metrics)), make(map[string]map[string]fiber.Map, len(metrics))
			for _, metric := range metrics {
				if mapped[metric.Key] == nil {
					mapped[metric.Key] = map[string]fiber.Map{}
				}
				if _, ok := mapped[metric.Key][metric.Context.GetJsonHash()]; !ok {
					// to be properly decoded by AIM UI, json should be represented as a key:value object.
					context := fiber.Map{}
					if err := json.Unmarshal(metric.Context.Json, &context); err != nil {
						return eris.Wrap(err, "error unmarshalling `context` json to `fiber.Map` object")
					}
					mapped[metric.Key][metric.Context.GetJsonHash()] = context
					data[metric.Key] = append(data[metric.Key], context)
				}
			}
			resp[s] = data
		default:
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("%q is not a valid Sequence", s))
		}
	}
}
