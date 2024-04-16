package request

// UpdateExperimentRequest is a request struct for `PUT /experiments/:id` endpoint.
type UpdateExperimentRequest struct {
	ID          int32   `params:"id"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Archived    *bool   `json:"archived"`
}

// GetExperimentRequest is a request object for `GET /aim/experiments/:id` endpoint.
type GetExperimentRequest struct {
	ID int32 `params:"id"`
}

// GetExperimentRunsRequest is a request object for `GET /aim/experiments/:id/runs` endpoint.
type GetExperimentRunsRequest struct {
	ID     int32  `params:"id"`
	Limit  int    `query:"limit"`
	Offset string `query:"offset"`
}

// GetExperimentActivityRequest is a request object for `GET /aim/experiments/:id/activity/` endpoint.
type GetExperimentActivityRequest struct {
	ID int32 `params:"id"`
}

// DeleteExperimentRequest is a request object for `DELETE /aim/experiments/:id` endpoint.
type DeleteExperimentRequest struct {
	ID int32 `params:"id"`
}
