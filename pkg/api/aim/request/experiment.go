package request

// UpdateExperimentRequest is a request struct for `PUT /experiments/:id` endpoint.
type UpdateExperimentRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Archived    *bool   `json:"archived"`
}
