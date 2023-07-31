package fixtures

import (
	"context"
	"time"

	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/database"
)

// ParamFixtures represents data fixtures object.
type ParamFixtures struct {
	baseFixtures
}

// NewParamFixtures creates new instance of ParamFixtures.
func NewParamFixtures(databaseDSN string) (*ParamFixtures, error) {
	db, err := database.ConnectDB(
		databaseDSN,
		1*time.Second,
		20,
		false,
		false,
		"",
	)
	if err != nil {
		return nil, eris.Wrap(err, "error connection to database")
	}
	return &ParamFixtures{
		baseFixtures: baseFixtures{db: db.DB},
	}, nil
}

// CreateParam creates new test Param.
func (f ParamFixtures) CreateParam(ctx context.Context, param *models.Param) (*models.Param, error) {
	if err := f.baseFixtures.db.WithContext(ctx).Create(param).Error; err != nil {
		return nil, eris.Wrap(err, "error creating test param")
	}
	return param, nil
}
