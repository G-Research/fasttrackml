package response

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/dto"
)

// ProjectActivityResponse represents the response json for the `GET aim/projects/activity` endpoint.
type ProjectActivityResponse struct {
	NumRuns         int64          `json:"num_runs"`
	NumActiveRuns   int64          `json:"num_active_runs"`
	NumExperiments  int64          `json:"num_experiments"`
	NumArchivedRuns int64          `json:"num_archived_runs"`
	ActivityMap     map[string]int `json:"activity_map"`
}

// NewProjectActivityResponse creates new response object for `GET /projects/activity` endpoint.
func NewProjectActivityResponse(activity *dto.ProjectActivity) *ProjectActivityResponse {
	return &ProjectActivityResponse{
		NumRuns:         activity.NumRuns,
		NumActiveRuns:   activity.NumActiveRuns,
		NumExperiments:  activity.NumExperiments,
		NumArchivedRuns: activity.NumArchivedRuns,
		ActivityMap:     activity.ActivityMap,
	}
}

// GetProjectResponse represents the response json for the `GET aim/projects` endpoint.
type GetProjectResponse struct {
	Name             string `json:"name"`
	Path             string `json:"path"`
	Description      string `json:"description"`
	TelemetryEnabled int    `json:"telemetry_enabled"`
}

// NewGetProjectResponse creates new response object for `GET /projects` endpoint.
func NewGetProjectResponse(name, dialector string) *GetProjectResponse {
	return &GetProjectResponse{
		Name: name,
		Path: dialector,
	}
}

// ProjectParamsResponse is a response object for `GET /projects/params` endpoint.
type ProjectParamsResponse struct {
	Metric        map[string][]fiber.Map `json:"metric"`
	Params        fiber.Map              `json:"params"`
	Texts         fiber.Map              `json:"texts"`
	Audios        fiber.Map              `json:"audios"`
	Images        fiber.Map              `json:"images"`
	Figures       fiber.Map              `json:"figures"`
	Distributions fiber.Map              `json:"distributions"`
}

// NewProjectParamsResponse creates new response object for `GET /projects/params` endpoint.
func NewProjectParamsResponse(projectParams *dto.ProjectParams) (*ProjectParamsResponse, error) {
	// process params and tags
	params := make(map[string]any, len(projectParams.ParamKeys)+1)
	for _, paramKey := range projectParams.ParamKeys {
		params[paramKey] = map[string]string{
			"__example_type__": "<class 'str'>",
		}
	}

	tags := make(map[string]map[string]string, len(projectParams.TagKeys))
	for _, tagKey := range projectParams.TagKeys {
		tags[tagKey] = map[string]string{
			"__example_type__": "<class 'str'>",
		}
	}
	params["tags"] = tags

	// process metrics
	metrics, mapped := make(
		map[string][]fiber.Map, len(projectParams.Metrics),
	), make(map[string]map[string]fiber.Map, len(projectParams.Metrics))
	for _, metric := range projectParams.Metrics {
		if mapped[metric.Key] == nil {
			mapped[metric.Key] = map[string]fiber.Map{}
		}
		if _, ok := mapped[metric.Key][metric.Context.GetJsonHash()]; !ok {
			// to be properly decoded by AIM UI, json should be represented as a key:value object.
			context := fiber.Map{}
			if err := json.Unmarshal(metric.Context.Json, &context); err != nil {
				return nil, eris.Wrap(err, "error unmarshalling `context` json to `fiber.Map` object")
			}
			mapped[metric.Key][metric.Context.GetJsonHash()] = context
			metrics[metric.Key] = append(metrics[metric.Key], context)
		}
	}

	return &ProjectParamsResponse{
		Metric:        metrics,
		Params:        params,
		Texts:         fiber.Map{},
		Audios:        fiber.Map{},
		Images:        fiber.Map{},
		Figures:       fiber.Map{},
		Distributions: fiber.Map{},
	}, nil
}
