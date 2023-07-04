package response

import "github.com/G-Research/fasttrackml/pkg/api/mlflow/service/artifact/storage"

// FilePartialResponse is a partial response object for different responses.
type FilePartialResponse struct {
	Path     string `json:"path"`
	IsDir    bool   `json:"is_dir"`
	FileSize int64  `json:"file_size"`
}

// ListArtifactsResponse is a response object for `GET mlflow/artifacts/list` endpoint.
type ListArtifactsResponse struct {
	Files         []FilePartialResponse `json:"files"`
	RootURI       string                `json:"root_uri"`
	NextPageToken string                `json:"next_page_token,omitempty"`
}

// NewListArtifactsResponse creates new instance of ListArtifactsResponse.
func NewListArtifactsResponse(
	nextPageToken, rootURI string, artifacts []storage.ArtifactObject,
) *ListArtifactsResponse {
	response := ListArtifactsResponse{
		Files:         make([]FilePartialResponse, len(artifacts)),
		RootURI:       rootURI,
		NextPageToken: nextPageToken,
	}

	for i, artifact := range artifacts {
		response.Files[i] = FilePartialResponse{
			Path:     artifact.GetPath(),
			FileSize: artifact.GetSize(),
			IsDir:    artifact.IsDirectory(),
		}
	}

	return &response
}
