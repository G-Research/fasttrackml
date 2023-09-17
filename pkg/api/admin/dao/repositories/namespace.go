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
	// GetByCode returns namespace by its Code.
	GetByCode(ctx context.Context, code string) (*models.Namespace, error)
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
	if err := r.db.WithContext(ctx).Create(&namespace).Error; err != nil {
		return eris.Wrap(err, "error creating namespace entity")
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
