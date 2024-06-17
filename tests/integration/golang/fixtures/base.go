package fixtures

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"

	aimModels "github.com/G-Research/fasttrackml/pkg/api/aim/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

// baseFixtures represents base fixtures object.
type baseFixtures struct {
	db *gorm.DB
}

// TruncateTables cleans database from the old data.
func (f baseFixtures) TruncateTables() error {
	if err := f.db.Session(
		&gorm.Session{AllowGlobalUpdate: true},
	).Exec("DELETE from run_shared_tags").Error; err != nil {
		return errors.Wrap(err, "error deleting from many2many table")
	}
	for _, table := range []interface{}{
		aimModels.Dashboard{},
		aimModels.App{},
		aimModels.SharedTag{},
		models.Tag{},
		models.Param{},
		models.LatestMetric{},
		models.Metric{},
		models.Context{},
		models.Run{},
		models.ExperimentTag{},
		models.Experiment{},
		models.Namespace{},
		models.RoleNamespace{},
		models.Role{},
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
