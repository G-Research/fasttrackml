package request

// ListArtifactsRequest is a request object for `GET /mlflow/artifacts/list` endpoint.
type ListArtifactsRequest struct {
	Path    string `query:"path"`
	Token   string `query:"token"`
	RunID   string `query:"run_id"`
	RunUUID string `query:"run_uuid"`
}

// GetRunID returns Run ID.
func (r ListArtifactsRequest) GetRunID() string {
	if r.RunID == "" {
		return r.RunID
	}
	return r.RunUUID
}
