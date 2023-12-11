package v_0008

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/database/migrations"
)

const Version = "9c134b0e72a3"

func Migrate(db *gorm.DB) error {
	// We need to run this migration without foreign key constraints to avoid
	// the cascading delete to kick in and delete all the runs and dashboards.
	return migrations.RunWithoutForeignKeyIfNeeded(db, func() error {
		return db.Transaction(func(tx *gorm.DB) error {
			ns := Namespace{
				Code:                "default",
				Description:         "Default namespace",
				DefaultExperimentID: common.GetPointer[int32](0),
			}
			if err := tx.Where(Namespace{Code: "default"}).FirstOrCreate(&ns).Error; err != nil {
				return fmt.Errorf("error creating default namespace: %s", err)
			}

			for _, model := range []interface{}{
				&Experiment{},
				&App{},
			} {
				if err := tx.Model(model).
					Where("namespace_id IS NULL").
					Update("namespace_id", ns.ID).
					Error; err != nil {
					return eris.Wrapf(err, "error updating %t", model)
				}
				if err := tx.Migrator().AlterColumn(model, "namespace_id"); err != nil {
					return eris.Wrapf(err, "error altering column namespace_id for %t", model)
				}
			}

			return tx.Model(&SchemaVersion{}).
				Where("1 = 1").
				Update("Version", Version).
				Error
		})
	})
}
