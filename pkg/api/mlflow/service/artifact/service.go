package artifact

import (
	"context"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/repositories"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/service/artifact/storage"
)

// Service provides service layer to work with `artifact` business logic.
type Service struct {
	runRepository   repositories.RunRepositoryProvider
	artifactStorage storage.Provider
}

// NewService creates new Service instance.
func NewService(artifactStorage storage.Provider, runRepository repositories.RunRepositoryProvider) *Service {
	return &Service{
		runRepository:   runRepository,
		artifactStorage: artifactStorage,
	}
}

// ListArtifacts handles business logic of `GET /artifacts/list` endpoint.
func (s Service) ListArtifacts(
	ctx context.Context, req *request.ListArtifactsRequest,
) (string, []storage.ArtifactObject, error) {
	if err := ValidateListArtifactsRequest(req); err != nil {
		return "", nil, err
	}

	run, err := s.runRepository.GetByID(ctx, req.GetRunID())
	if err != nil {
		return "", nil, api.NewInternalError("unable to get artifact URI for run '%s'", req.GetRunID())
	}

	rootURI, artifacts, err := s.artifactStorage.List(
		run.ArtifactURI, req.Path,
	)
	if err != nil {
		return "", nil, api.NewInternalError("error getting artifact list from storage")
	}

	return rootURI, artifacts, nil
}
