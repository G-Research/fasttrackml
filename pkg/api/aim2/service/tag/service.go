package tag

import (
	"context"

	aimModels "github.com/G-Research/fasttrackml/pkg/api/aim2/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/repositories"
	mlflowModels "github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/common/api"
)

// Service provides service layer to work with `tag` business logic.
type Service struct {
	tagRepository repositories.TagRepositoryProvider
}

// NewService creates new Service instance.
func NewService(tagRepository repositories.TagRepositoryProvider) *Service {
	return &Service{
		tagRepository: tagRepository,
	}
}

// GetTags returns the list of active apps.
// TODO this is not implemented
func (s Service) GetTags(ctx context.Context, namespace *mlflowModels.Namespace) ([]aimModels.Tag, error) {
	tags, err := s.tagRepository.GetTagsByNamespace(ctx, namespace.ID)
	if err != nil {
		return nil, api.NewInternalError("unable to get active apps: %v", err)
	}
	// TODO remove stub data
	tags = []aimModels.Tag{}
	return tags, nil
}
