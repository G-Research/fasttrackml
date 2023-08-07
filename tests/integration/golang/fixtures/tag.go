package fixtures

import (
	"context"

	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

// TagFixtures represents data fixtures object.
type TagFixtures struct {
	BaseFixtures
}

// NewTagFixtures creates new instance of TagFixtures.
func NewTagFixtures(databaseDSN string) (*TagFixtures, error) {
	db, err := CreateDB(databaseDSN)
	if err != nil {
		return nil, err
	}
	return &TagFixtures{
		BaseFixtures: BaseFixtures{db: db.DB},
	}, nil
}

// CreateTag creates new test Tag.
func (f TagFixtures) CreateTag(ctx context.Context, tag *models.Tag) (*models.Tag, error) {
	if err := f.BaseFixtures.db.WithContext(ctx).Create(tag).Error; err != nil {
		return nil, eris.Wrap(err, "error creating test tag")
	}
	return tag, nil
}
