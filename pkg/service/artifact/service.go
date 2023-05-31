package artifact

import (
	"context"
	"github.com/G-Research/fasttrackml/pkg/repositories"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
)

// Service provides service layer to work with `artifact` business logic.
type Service struct {
	runRepository repositories.RunRepositoryProvider
}

// NewService creates new Service instance.
func NewService(runRepository repositories.RunRepositoryProvider) *Service {
	return &Service{
		runRepository: runRepository,
	}
}

func (s Service) ListArtifacts(ctx context.Context, req *request.ListArtifactsRequest) error {
	if err := ValidateListArtifactsRequest(req); err != nil {
		return err
	}

	_, err := s.runRepository.GetByID(ctx, req.GetRunID())
	if err != nil {
		return api.NewInternalError("unable to get artifact URI for run '%s'", req.GetRunID())
	}

	return nil
}
