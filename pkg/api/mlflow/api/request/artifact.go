package request

// BaseArtifactRequest common fields for artifact requests
type BaseArtifactRequest struct {
	Path    string `query:"path"`
	RunID   string `query:"run_id"`
	RunUUID string `query:"run_uuid"`
}

// ListArtifactsRequest is a request object for `GET /mlflow/artifacts/list` endpoint.
type ListArtifactsRequest struct {
	BaseArtifactRequest
}

// GetArtifactRequest is for linking to individual artifact item
type GetArtifactRequest struct {
	BaseArtifactRequest
}

// GetRunID returns Run ID.
func (r BaseArtifactRequest) GetRunID() string {
	if r.RunID != "" {
		return r.RunID
	}
	return r.RunUUID
}
