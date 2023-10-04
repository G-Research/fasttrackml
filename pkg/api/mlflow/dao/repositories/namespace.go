package repositories

import (
	"context"
	"errors"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

// NamespaceRepositoryProvider provides an interface to work with `namespace` entity.
type NamespaceRepositoryProvider interface {
	// Create creates new models.Namespace entity.
	Create(ctx context.Context, namespace *models.Namespace) error
	// Update modifies the existing models.Namespace entity.
	Update(ctx context.Context, namespace *models.Namespace) error
	// Delete removes a namespace and it's associated experiments by its ID.
	Delete(ctx context.Context, id uint) error
	// GetByCode returns namespace by its Code.
	GetByCode(ctx context.Context, code string) (*models.Namespace, error)
	// GetByID returns namespace by its ID.
	GetByID(ctx context.Context, id uint) (*models.Namespace, error)
	// List returns all namespaces.
	List(ctx context.Context) ([]models.Namespace, error)
}

// NamespaceRepository repository to work with `namespace` entity.
type NamespaceRepository struct {
	db *gorm.DB
}

// NewNamespaceRepository creates repository to work with `namespace` entity.
func NewNamespaceRepository(db *gorm.DB) *NamespaceRepository {
	return &NamespaceRepository{
		db: db,
	}
}

// Create creates new models.Namespace entity.
func (r NamespaceRepository) Create(ctx context.Context, namespace *models.Namespace) error {
	if err := r.db.WithContext(ctx).Create(namespace).Error; err != nil {
		return eris.Wrap(err, "error creating namespace entity")
	}
	return nil
}

// Update persists modifications to models.Namespace entity.
func (r NamespaceRepository) Update(ctx context.Context, namespace *models.Namespace) error {
	if err := r.db.WithContext(ctx).Updates(namespace).Error; err != nil {
		return eris.Wrap(err, "error updating namespace entity")
	}
	return nil
}

// Delete removes the  models.Namespace entity.
func (r NamespaceRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&models.Namespace{}, id).Error; err != nil {
		return eris.Wrap(err, "error deleting namespace entity")
	}
	return nil
}

// GetByCode returns namespace by its Code.
func (r NamespaceRepository) GetByCode(ctx context.Context, code string) (*models.Namespace, error) {
	var namespace models.Namespace
	if err := r.db.WithContext(ctx).Where(
		"code = ?", code,
	).First(&namespace).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, eris.Wrapf(err, "error getting namespace by code: %s", code)
	}
	return &namespace, nil
}

// GetByID returns namespace by its ID.
func (r NamespaceRepository) GetByID(ctx context.Context, id uint) (*models.Namespace, error) {
	var namespace models.Namespace
	if err := r.db.WithContext(ctx).First(&namespace, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, eris.Wrapf(err, "error getting namespace by id: %d", id)
	}
	return &namespace, nil
}

// List returns all namespaces.
func (r NamespaceRepository) List(ctx context.Context) ([]models.Namespace, error) {
	var namespaces []models.Namespace
	if err := r.db.WithContext(ctx).Find(&namespaces).Error; err != nil {
		return nil, eris.Wrap(err, "error listing namespaces")
	}
	return namespaces, nil
}
