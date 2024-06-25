package response

import (
	"github.com/google/uuid"

	"github.com/G-Research/fasttrackml/pkg/api/aim/dao/models"
)

// TagResponse represents a run tag.
type TagResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Color       string    `json:"color"`
	Description string    `json:"description"`
	RunCount    int       `json:"run_count"`
	Archived    bool      `json:"archived"`
}

// GetTagsResponse represents a list of run tags.
type GetTagsResponse []TagResponse

// GetRunsTaggedResponse represents list of runs for a tag.
type GetRunsTaggedResponse struct {
	ID   uuid.UUID                `json:"id"`
	Runs []GetRunInfoPropsPartial `json:"runs"`
}

// NewGetTagsResponse will convert the []model.SharedTag to GetTagsResponse
func NewGetTagsResponse(tags []models.SharedTag) GetTagsResponse {
	tagResponses := GetTagsResponse{}
	for _, tag := range tags {
		tagResponses = append(tagResponses, TagResponse{
			ID:          tag.ID,
			Name:        tag.Name,
			Color:       tag.Color,
			Description: tag.Description,
			Archived:    tag.IsArchived,
			RunCount:    len(tag.Runs),
		})
	}
	return tagResponses
}

// NewCreateTagResponse creates new response object for `POST /dashboards` endpoint.
func NewCreateTagResponse(tag *models.SharedTag) TagResponse {
	return TagResponse{
		ID:          tag.ID,
		Name:        tag.Name,
		Color:       tag.Color,
		Description: tag.Description,
		Archived:    tag.IsArchived,
		RunCount:    len(tag.Runs),
	}
}

func NewGetRunsTaggedResponse(tag *models.SharedTag) GetRunsTaggedResponse {
	resp := GetRunsTaggedResponse{
		ID: tag.ID,
	}
	for _, run := range tag.Runs {
		resp.Runs = append(resp.Runs, GetRunInfoPropsPartial{
			ID:           run.ID,
			RunID:        run.ID,
			Name:         run.Name,
			Experiment:   GetRunInfoExperimentPartial{Name: run.Experiment.Name},
			CreationTime: float64(run.StartTime.Int64),
			EndTime:      float64(run.EndTime.Int64),
		})
	}
	return resp
}

// NewGetTagResponse creates new response object for `GET /apps/:id` endpoint.
var NewGetTagResponse = NewCreateTagResponse

// NewUpdateTagResponse creates new response object for `PUT /apps/:id` endpoint.
var NewUpdateTagResponse = NewCreateTagResponse

// ConvertTagsToMaps converts tags for streaming repsonses.
func ConvertTagsToMaps(tags []models.SharedTag) []map[string]string {
	sharedTags := []map[string]string{}
	for _, tag := range tags {
		sharedTags = append(sharedTags, map[string]string{
			"id":    tag.ID.String(),
			"name":  tag.Name,
			"color": tag.Color,
		})
	}
	return sharedTags
}
