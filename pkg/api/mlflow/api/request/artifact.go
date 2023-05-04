package request

// ListArtifactsRequest is a request object for `GET mlflow/artifacts/list` endpoint.
type ListArtifactsRequest struct {
	RunID string `query:"run_id"`
	Path  string `query:"path"`
	Token string `query:"token"`
}
