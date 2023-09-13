package fixtures

import (
	"context"

	"github.com/G-Research/fasttrackml/pkg/common/dao/repositories"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
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
