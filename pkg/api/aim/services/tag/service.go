package tag

import (
	"context"

	"github.com/G-Research/fasttrackml/pkg/api/aim/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim/dao/convertors"
	"github.com/G-Research/fasttrackml/pkg/api/aim/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/aim/dao/repositories"
	"github.com/G-Research/fasttrackml/pkg/common/api"
)

// Service provides service layer to work with `tag` business logic.
type Service struct {
	sharedTagRepository repositories.SharedTagRepositoryProvider
}

// NewService creates new Service instance.
func NewService(sharedTagRepository repositories.SharedTagRepositoryProvider) *Service {
	return &Service{
		sharedTagRepository: sharedTagRepository,
	}
}

// GetTags returns the list of tags.
func (s Service) GetTags(ctx context.Context, namespaceID uint) ([]models.SharedTag, error) {
	tags, err := s.sharedTagRepository.GetTagsByNamespace(ctx, namespaceID)
	if err != nil {
		return nil, api.NewInternalError("unable to get tags: %v", err)
	}
	return tags, nil
}

// Get returns tag object.
func (s Service) Get(
	ctx context.Context, namespaceID uint, req *request.GetTagRequest,
) (*models.SharedTag, error) {
	tag, err := s.sharedTagRepository.GetByNamespaceIDAndTagID(ctx, namespaceID, req.ID.String())
	if err != nil {
		return nil, api.NewInternalError("unable to find tag by id %q: %s", req.ID, err)
	}
	if tag == nil {
		return nil, api.NewResourceDoesNotExistError("tag '%s' not found", req.ID)
	}
	return tag, nil
}

// Create creates new tag object.
func (s Service) Create(
	ctx context.Context, namespaceID uint, req *request.CreateTagRequest,
) (*models.SharedTag, error) {
	if err := ValidateCreateTagRequest(req); err != nil {
		return nil, err
	}
	tag := convertors.ConvertCreateTagRequestToDBModel(*req, namespaceID)
	if err := s.sharedTagRepository.Create(ctx, &tag); err != nil {
		return nil, api.NewInternalError("unable to create tag: %v", err)
	}
	return &tag, nil
}

// Update updates existing tag object.
func (s Service) Update(
	ctx context.Context, namespaceID uint, req *request.UpdateTagRequest,
) (*models.SharedTag, error) {
	tag, err := s.sharedTagRepository.GetByNamespaceIDAndTagID(ctx, namespaceID, req.ID.String())
	if err != nil {
		return nil, api.NewInternalError("unable to find tag by id %s: %s", req.ID, err)
	}
	if tag == nil {
		return nil, api.NewResourceDoesNotExistError("tag with id '%s' not found", req.ID)
	}

	tag.Name = req.Name
	tag.Description = req.Description
	tag.Color = req.Color
	tag.IsArchived = req.IsArchived

	if err := s.sharedTagRepository.Update(ctx, tag); err != nil {
		return nil, api.NewInternalError("unable to update tag '%s': %s", tag.ID, err)
	}
	return tag, nil
}

// Delete deletes existing object.
func (s Service) Delete(ctx context.Context, namespaceID uint, req *request.DeleteTagRequest) error {
	tag, err := s.sharedTagRepository.GetByNamespaceIDAndTagID(ctx, namespaceID, req.ID.String())
	if err != nil {
		return api.NewInternalError("error trying to find tag by id %s: %s", req.ID, err)
	}
	if tag == nil {
		return api.NewResourceDoesNotExistError("tag with id '%s' not found", req.ID)
	}
	if err := s.sharedTagRepository.Delete(ctx, tag); err != nil {
		return api.NewInternalError("unable to delete tag by id %s: %s", req.ID, err)
	}
	return nil
}
