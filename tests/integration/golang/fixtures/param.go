package fixtures

import (
	"context"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

// ParamFixtures represents data fixtures object.
type ParamFixtures struct {
	baseFixtures
}

// NewParamFixtures creates a new instance of ParamFixtures.
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

// GetParamsByRunID returns all params for a given run.
func (f ParamFixtures) GetParamsByRunID(ctx context.Context, runID string) ([]models.Param, error) {
	var params []models.Param
	if err := f.baseFixtures.db.WithContext(ctx).Where("run_uuid = ?", runID).Find(&params).Error; err != nil {
		return nil, eris.Wrap(err, "error getting params by run id")
	}
	return params, nil
}
