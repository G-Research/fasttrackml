package artifact

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"slices"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/repositories"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/service/artifact/storage"
)

// Service provides service layer to work with `artifact` business logic.
type Service struct {
	runRepository          repositories.RunRepositoryProvider
	artifactStorageFactory storage.ArtifactStorageFactoryProvider
}

// NewService creates new Service instance.
func NewService(
	runRepository repositories.RunRepositoryProvider,
	artifactStorageFactory storage.ArtifactStorageFactoryProvider,
) *Service {
	return &Service{
		runRepository:          runRepository,
		artifactStorageFactory: artifactStorageFactory,
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

	artifactStorage, err := s.artifactStorageFactory.GetStorage(run.ArtifactURI)
	if err != nil {
		return "", nil, api.NewInternalError("run with id '%s' has unsupported artifact storage", run.ID)
	}

	artifacts, err := artifactStorage.List(run.ArtifactURI, req.Path)
	if err != nil {
		return "", nil, api.NewInternalError("error getting artifact list from storage")
	}

	// sort artifacts by path
	slices.SortFunc(artifacts, func(a, b storage.ArtifactObject) int {
		return cmp.Compare(a.Path, b.Path)
	})

	return run.ArtifactURI, artifacts, nil
}

// GetArtifact handles business logic of `GET /artifacts/get` endpoint.
func (s Service) GetArtifact(
	ctx context.Context, req *request.GetArtifactRequest,
) (io.ReadCloser, error) {
	if err := ValidateGetArtifactRequest(req); err != nil {
		return nil, err
	}

	run, err := s.runRepository.GetByID(ctx, req.GetRunID())
	if err != nil {
		return nil, api.NewInternalError("unable to find run '%s': %s", req.GetRunID(), err)
	}
	if run == nil {
		return nil, api.NewResourceDoesNotExistError("unable to find run '%s'", req.GetRunID())
	}
	artifactStorage, err := s.artifactStorageFactory.GetStorage(run.ArtifactURI)
	if err != nil {
		return nil, api.NewInternalError("run with id '%s' has unsupported artifact storage", run.ID)
	}

	artifactReader, err := artifactStorage.Get(
		run.ArtifactURI, req.Path,
	)
	if err != nil {
		msg := fmt.Sprintf(
			"error getting artifact object for URI: %s",
			filepath.Join(run.ArtifactURI, req.Path),
		)
		var s3NotFound *types.NoSuchKey
		if errors.Is(err, fs.ErrNotExist) || errors.As(err, &s3NotFound) {
			return nil, api.NewResourceDoesNotExistError(msg)
		} else {
			return nil, api.NewInternalError(msg)
		}
	}
	return artifactReader, nil
}
