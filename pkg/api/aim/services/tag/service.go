package tag

import (
	"context"

	"github.com/G-Research/fasttrackml/pkg/api/aim/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/aim/dao/repositories"
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

// GetTags returns the list of tags.
func (s Service) GetTags(ctx context.Context, namespaceID uint) ([]models.Tag, error) {
	tags, err := s.tagRepository.GetTagsByNamespace(ctx, namespaceID)
	if err != nil {
		return nil, api.NewInternalError("unable to get active apps: %v", err)
	}
	return tags, nil
}
