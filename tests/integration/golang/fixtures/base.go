package fixtures

import (
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/database"
)

// baseFixtures represents base fixtures object.
type baseFixtures struct {
	db *gorm.DB
}

// UnloadFixtures cleans database from the old data.
func (f baseFixtures) UnloadFixtures() error {
	for _, table := range []interface{}{
		database.Dashboard{}, // TODO update to models when available
		database.App{},       // TODO update to models when available
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

// CreateDB will convert the a DSN input into a database connection
func CreateDB(databaseDSN string) (db database.DbProvider, err error) {
	db, err = database.MakeDbProvider(
		databaseDSN,
		1*time.Second,
		20,
		false,
		false,
		"",
	)
	database.DB = db.Db()
	return
}
