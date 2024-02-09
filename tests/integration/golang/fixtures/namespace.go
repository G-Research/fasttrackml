package fixtures

import (
	"context"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/repositories"
)

// NamespaceFixtures represents data fixtures object.
type NamespaceFixtures struct {
	baseFixtures
	namespaceRepository repositories.NamespaceRepositoryProvider
}

// NewNamespaceFixtures creates new instance of NamespaceFixtures.
func NewNamespaceFixtures(db *gorm.DB) (*NamespaceFixtures, error) {
	return &NamespaceFixtures{
		baseFixtures:        baseFixtures{db: db},
		namespaceRepository: repositories.NewNamespaceRepository(db),
	}, nil
}

// CreateNamespace creates a new test Namespace.
func (f NamespaceFixtures) CreateNamespace(
	ctx context.Context, namespace *models.Namespace,
) (*models.Namespace, error) {
	if err := f.namespaceRepository.Create(ctx, namespace); err != nil {
		return nil, eris.Wrap(err, "error creating test namespace")
	}
	return namespace, nil
}

// UpsertNamespace creates a new test Namespace or updates existing.
func (f NamespaceFixtures) UpsertNamespace(
	ctx context.Context, namespace *models.Namespace,
) (*models.Namespace, error) {
	if err := f.db.
		Clauses(
			clause.OnConflict{
				Columns:   []clause.Column{{Name: "code"}},
				UpdateAll: true,
			}).
		Model(models.Namespace{}).
		Create(namespace).Error; err != nil {
		return nil, eris.Wrap(err, "error creating test namespace")
	}
	return namespace, nil
}

// GetNamespaces fetches all namespaces.
func (f NamespaceFixtures) GetNamespaces(
	ctx context.Context,
) ([]models.Namespace, error) {
	var namespaces []models.Namespace
	if err := f.db.WithContext(ctx).
		Find(&namespaces).Error; err != nil {
		return nil, eris.Wrapf(err, "error getting 'namespaces' entities")
	}
	return namespaces, nil
}

// GetNamespaceByID fetches a namespace by ID.
func (f NamespaceFixtures) GetNamespaceByID(
	ctx context.Context, id uint,
) (*models.Namespace, error) {
	var namespace models.Namespace
	if err := f.db.WithContext(ctx).
		Where("id = ?", id).
		First(&namespace).Error; err != nil {
		return nil, eris.Wrapf(err, "error getting namespace with ID %d", id)
	}
	return &namespace, nil
}

// UpdateNamespace updates an existing test Namespace.
func (f NamespaceFixtures) UpdateNamespace(
	ctx context.Context, namespace *models.Namespace,
) (*models.Namespace, error) {
	if err := f.namespaceRepository.Update(ctx, namespace); err != nil {
		return nil, eris.Wrap(err, "error updating test namespace")
	}
	return namespace, nil
}

// GetNamespaceByCode fetches a namespace by code.
func (f NamespaceFixtures) GetNamespaceByCode(
	ctx context.Context,
	code string,
) (*models.Namespace, error) {
	var namespace models.Namespace
	if err := f.db.WithContext(
		ctx,
	).Where(
		"code = ?", code,
	).First(
		&namespace,
	).Error; err != nil {
		return nil, eris.Wrap(err, "error getting default namespace")
	}
	return &namespace, nil
}
