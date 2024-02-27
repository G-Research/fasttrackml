package response

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/models"
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
func NewProjectActivityResponse(activity *models.ProjectActivity) *ProjectActivityResponse {
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
	Metric        *map[string][]fiber.Map `json:"metric,omitempty"`
	Params        *map[string]any         `json:"params,omitempty"`
	Texts         *fiber.Map              `json:"texts,omitempty"`
	Audios        *fiber.Map              `json:"audios,omitempty"`
	Images        *fiber.Map              `json:"images,omitempty"`
	Figures       *fiber.Map              `json:"figures,omitempty"`
	Distributions *fiber.Map              `json:"distributions,omitempty"`
}

// NewProjectParamsResponse creates new response object for `GET /projects/params` endpoint.
func NewProjectParamsResponse(projectParams *models.ProjectParams,
	excludeParams bool, sequences []string,
) (*ProjectParamsResponse, error) {
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

	rsp := ProjectParamsResponse{}
	if !excludeParams {
		rsp.Params = &params
	}
	if len(sequences) == 0 {
		sequences = []string{
			"metric",
			"images",
			"texts",
			"figures",
			"distributions",
			"audios",
		}
	}
	for _, s := range sequences {
		switch s {
		case "images":
			rsp.Images = &fiber.Map{}
		case "texts":
			rsp.Texts = &fiber.Map{}
		case "figures":
			rsp.Figures = &fiber.Map{}
		case "distributions":
			rsp.Distributions = &fiber.Map{}
		case "audios":
			rsp.Audios = &fiber.Map{}
		case "metric":
			rsp.Metric = &metrics
		}
	}
	return &rsp, nil
}
