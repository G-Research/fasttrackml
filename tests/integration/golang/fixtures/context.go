package fixtures

import (
	"context"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

// ContextFixtures represents data fixtures object.
type ContextFixtures struct {
	baseFixtures
}

// NewContextFixtures creates new instance of ContextFixtures.
func NewContextFixtures(db *gorm.DB) (*ContextFixtures, error) {
	return &ContextFixtures{
		baseFixtures: baseFixtures{db: db},
	}, nil
}

// CreateContext creates new test Context.
func (f ContextFixtures) CreateContext(ctx context.Context, context *models.Context) (*models.Context, error) {
	if err := f.baseFixtures.db.WithContext(ctx).Create(context).Error; err != nil {
		return nil, eris.Wrap(err, "error creating test context")
	}
	return context, nil
}

// GetContext returns the Context with the given JSON.
func (f ContextFixtures) GetContext(ctx context.Context, json string) (*models.Context, error) {
	var context models.Context
	if err := f.db.WithContext(ctx).Where("json = ?", json).First(&context).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, eris.Wrapf(err, "error getting context with json: %s", json)
	}
	return &context, nil
}
