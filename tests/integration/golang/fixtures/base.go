package fixtures

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

// BaseFixtures represents base fixtures object.
type baseFixtures struct {
	db *gorm.DB
}

// UnloadFixtures cleans database from the old data.
func (f baseFixtures) UnloadFixtures() error {
	for _, table := range []schema.Tabler{
		models.ExperimentTag{},
		models.Experiment{},
	} {
		if err := f.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(table).Error; err != nil {
			return errors.Wrapf(err, "error deleting data from %s table", table.TableName())
		}
	}
	return nil
}
