// TODO move this beneath package 'api' when available
package request

// UpdateRun represents the  data to archive or update a Run
type UpdateRun struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Archived    *bool   `json:"archived"`
}
