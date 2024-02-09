package fixtures

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
	"github.com/G-Research/fasttrackml/pkg/database"
)

// baseFixtures represents base fixtures object.
type baseFixtures struct {
	db *gorm.DB
}

// TruncateTables cleans database from the old data.
func (f baseFixtures) TruncateTables() error {
	for _, table := range []interface{}{
		database.Dashboard{}, // TODO update to models when available
		database.App{},       // TODO update to models when available
		models.Tag{},
		models.Param{},
		models.LatestMetric{},
		models.Metric{},
		models.Context{},
		models.Run{},
		models.ExperimentTag{},
		models.Experiment{},
		models.Namespace{},
	} {
		if err := f.db.Session(
			&gorm.Session{AllowGlobalUpdate: true},
		).Unscoped().Delete(
			table,
		).Error; err != nil {
			return errors.Wrap(err, "error deleting data")
		}
	}
	return nil
}
