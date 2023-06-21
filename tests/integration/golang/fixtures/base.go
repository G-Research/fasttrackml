package fixtures

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

// BaseFixtures represents base fixtures object.
type baseFixtures struct {
	db *gorm.DB
}

// UnloadFixtures cleans database from the old data.
func (f baseFixtures) UnloadFixtures() error {
	for _, table := range []interface{}{
		models.Tag{},
		models.Param{},
		models.LatestMetric{},
		models.Metric{},
		models.Run{},
		models.ExperimentTag{},
		models.Experiment{},
	} {
		if err := f.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(table).Error; err != nil {
			return errors.Wrap(err, "error deleting data")
		}
	}
	return nil
}
