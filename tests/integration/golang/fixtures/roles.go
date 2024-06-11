package fixtures

import (
	"context"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

// RoleFixtures represents data fixtures object.
type RoleFixtures struct {
	baseFixtures
}

// NewRoleFixtures creates new instance of RoleFixtures.
func NewRoleFixtures(db *gorm.DB) (*RoleFixtures, error) {
	return &RoleFixtures{
		baseFixtures: baseFixtures{db: db},
	}, nil
}

// CreateRole creates a new test Role.
func (f RoleFixtures) CreateRole(ctx context.Context, role *models.Role) error {
	if err := f.db.WithContext(ctx).Create(role).Error; err != nil {
		return eris.Wrap(err, "error creating role entity")
	}
	return nil
}

// AttachNamespaceToRole attaches a Role to provided Namespace.
func (f RoleFixtures) AttachNamespaceToRole(
	ctx context.Context, role *models.Role, namespace *models.Namespace,
) error {
	if err := f.db.WithContext(ctx).Create(&models.RoleNamespace{
		RoleID:      role.ID,
		NamespaceID: namespace.ID,
	}).Error; err != nil {
		return eris.Wrap(err, "error attaching namespace to role ")
	}
	return nil
}
