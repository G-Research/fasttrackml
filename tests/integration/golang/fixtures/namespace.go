package fixtures

import (
	"context"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/common/dao/repositories"
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

// GetDefaultNamespace returns default namespace.
func (f NamespaceFixtures) GetDefaultNamespace(ctx context.Context) (*models.Namespace, error) {
	namespace, err := f.namespaceRepository.GetByCode(ctx, "default")
	if err != nil {
		return nil, eris.Wrap(err, "error getting default namespace")
	}
	return namespace, nil
}
