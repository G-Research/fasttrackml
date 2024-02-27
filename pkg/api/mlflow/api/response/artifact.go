package response

import "github.com/G-Research/fasttrackml/pkg/api/mlflow/services/artifact/storage"

// FilePartialResponse is a partial response object for different responses.
type FilePartialResponse struct {
	Path     string `json:"path"`
	IsDir    bool   `json:"is_dir"`
	FileSize int64  `json:"file_size"`
}

// ListArtifactsResponse is a response object for `GET mlflow/artifacts/list` endpoint.
type ListArtifactsResponse struct {
	Files   []FilePartialResponse `json:"files"`
	RootURI string                `json:"root_uri"`
}

// NewListArtifactsResponse creates new instance of ListArtifactsResponse.
func NewListArtifactsResponse(
	rootURI string, artifacts []storage.ArtifactObject,
) *ListArtifactsResponse {
	response := ListArtifactsResponse{
		Files:   make([]FilePartialResponse, len(artifacts)),
		RootURI: rootURI,
	}

	for i, artifact := range artifacts {
		response.Files[i] = FilePartialResponse{
			Path:     artifact.GetPath(),
			IsDir:    artifact.IsDirectory(),
			FileSize: artifact.GetSize(),
		}
	}

	return &response
}
