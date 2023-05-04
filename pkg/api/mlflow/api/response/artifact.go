package response

// FilePartialResponse is a partial response object for different responses.
type FilePartialResponse struct {
	Path     string `json:"path"`
	IsDir    bool   `json:"is_dir"`
	FileSize int64  `json:"file_size"`
}

// ListArtifactsResponse is a response object for `GET mlflow/artifacts/list` endpoint.
type ListArtifactsResponse struct {
	RootURI       string                `json:"root_uri"`
	Files         []FilePartialResponse `json:"files"`
	NextPageToken string                `json:"next_page_token,omitempty"`
}
