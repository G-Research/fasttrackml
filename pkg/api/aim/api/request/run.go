package request

// UpdateRunRequest represents the data to archive or update a Run
type UpdateRunRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Archived    *bool   `json:"archived"`
}
