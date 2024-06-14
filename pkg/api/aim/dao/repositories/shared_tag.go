package repositories

import (
	"context"
	"errors"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/aim/dao/models"
	"github.com/G-Research/fasttrackml/pkg/common/dao/repositories"
)

// SharedTagRepositoryProvider provides an interface to work with models.SharedTag entity.
type SharedTagRepositoryProvider interface {
	repositories.BaseRepositoryProvider
	// GetTagsByNamespace returns the list of tags.
	GetTagsByNamespace(ctx context.Context, namespaceID uint) ([]models.SharedTag, error)
	// GetByNamespaceIDAndTagID returns a single SharedTag.
	GetByNamespaceIDAndTagID(ctx context.Context, namespaceID uint, tagID string) (*models.SharedTag, error)
	// GetByNamespaceIDAndTagName returns a single SharedTag.
	GetByNamespaceIDAndTagName(ctx context.Context, namespaceID uint, tagName string) (*models.SharedTag, error)
	// Create a SharedTag.
	Create(context.Context, *models.SharedTag) error
	// Update an existing SharedTag.
	Update(context.Context, *models.SharedTag) error
	// Delete an existing SharedTag.
	Delete(context.Context, *models.SharedTag) error
	// AddAssociation adds an existing SharedTag/Run association.
	AddAssociation(context.Context, *models.SharedTag, *models.Run) error
	// DeleteAssociation removes an existing SharedTag/Run association.
	DeleteAssociation(context.Context, *models.SharedTag, *models.Run) error
}

// SharedTagRepository repository to work with models.Tag entity.
type SharedTagRepository struct {
	repositories.BaseRepositoryProvider
}

// NewSharedTagRepository creates repository to work with models.Tag entity.
func NewSharedTagRepository(db *gorm.DB) *SharedTagRepository {
	return &SharedTagRepository{
		repositories.NewBaseRepository(db),
	}
}

// GetTagsByNamespace returns the list of SharedTag, with virtual rows populated from the Tag table.
func (r SharedTagRepository) GetTagsByNamespace(ctx context.Context, namespaceID uint) ([]models.SharedTag, error) {
	var tags []models.SharedTag
	if err := r.GetDB().WithContext(ctx).
		Preload("Runs").
		Preload("Runs.Experiment").
		Find(&tags).Error; err != nil {
		return nil, eris.Wrap(err, "unable to fetch tags")
	}
	return tags, nil
}

// GetByNamespaceIDAndTagID returns models.Tag by Namespace and Tag ID.
func (d SharedTagRepository) GetByNamespaceIDAndTagID(ctx context.Context,
	namespaceID uint, tagID string,
) (*models.SharedTag, error) {
	var tag models.SharedTag
	if err := d.GetDB().WithContext(ctx).
		Where("NOT is_archived").
		Where("namespace_id = ?", namespaceID).
		Where("id = ?", tagID).
		Preload("Runs").
		Preload("Runs.Experiment").
		First(&tag).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, eris.Wrapf(err, "error getting tag by id: %s", tagID)
	}
	return &tag, nil
}

// GetByNamespaceIDAndTagName returns models.Tag by Namespace and Tag ID.
func (d SharedTagRepository) GetByNamespaceIDAndTagName(ctx context.Context,
	namespaceID uint, tagName string,
) (*models.SharedTag, error) {
	var tag models.SharedTag
	if err := d.GetDB().WithContext(ctx).
		Where("NOT is_archived").
		Where("namespace_id = ?", namespaceID).
		Where("name = ?", tagName).
		Preload("Runs").
		Preload("Runs.Experiment").
		First(&tag).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, eris.Wrapf(err, "error getting tag by name: %s", tagName)
	}
	return &tag, nil
}

// Create creates new models.SharedTag object.
func (d SharedTagRepository) Create(ctx context.Context, tag *models.SharedTag) error {
	if err := d.GetDB().WithContext(ctx).Create(&tag).Error; err != nil {
		return eris.Wrap(err, "error creating tag entity")
	}
	return nil
}

// Update updates existing models.Tag object.
func (d SharedTagRepository) Update(ctx context.Context, tag *models.SharedTag) error {
	if err := d.GetDB().WithContext(ctx).
		Model(&tag).
		Updates(models.SharedTag{
			Name:        tag.Name,
			Description: tag.Description,
			Color:       tag.Color,
		}).
		Error; err != nil {
		return eris.Wrap(err, "error updating tag entity")
	}
	return nil
}

// Delete deletes a models.Tag object.
func (d SharedTagRepository) Delete(ctx context.Context, tag *models.SharedTag) error {
	if err := d.GetDB().WithContext(ctx).
		Model(&tag).
		Update("IsArchived", true).
		Error; err != nil {
		return eris.Wrap(err, "error deleting tag entity")
	}
	return nil
}

// AddAssociation will add the association between SharedTag and Run.
func (d SharedTagRepository) AddAssociation(ctx context.Context, tag *models.SharedTag, run *models.Run) error {
	if err := d.GetDB().WithContext(ctx).
		Exec("INSERT INTO run_shared_tags VALUES(?, ?) ON CONFLICT DO NOTHING", tag.ID, run.ID).
		Error; err != nil {
		return eris.Wrap(err, "error adding tag/run association")
	}
	return nil
}

// DeleteAssociation will remove the association between SharedTag and Run.
func (d SharedTagRepository) DeleteAssociation(ctx context.Context, tag *models.SharedTag, run *models.Run) error {
	if err := d.GetDB().WithContext(ctx).
		Model(&tag).
		Association("Runs").
		Delete(run); err != nil {
		return eris.Wrap(err, "error removing tag/run association")
	}
	return nil
}
