package fixtures

import (
	"context"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
)

// ParamFixtures represents data fixtures object.
type ParamFixtures struct {
	baseFixtures
}

// NewParamFixtures creates new instance of ParamFixtures.
func NewParamFixtures(db *gorm.DB) (*ParamFixtures, error) {
	return &ParamFixtures{
		baseFixtures: baseFixtures{db: db},
	}, nil
}

// CreateParam creates new test Param.
func (f ParamFixtures) CreateParam(ctx context.Context, param *models.Param) (*models.Param, error) {
	if err := f.baseFixtures.db.WithContext(ctx).Create(param).Error; err != nil {
		return nil, eris.Wrap(err, "error creating test param")
	}
	return param, nil
}
