package v_0009

import (
	"github.com/rotisserie/eris"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/database/migrations"
)

const Version = "2c2299e4e061"

func Migrate(db *gorm.DB) error {
	return migrations.RunWithoutForeignKeyIfNeeded(db, func() error {
		return db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Migrator().AutoMigrate(&Context{}); err != nil {
				return err
			}

			for _, model := range []interface{}{
				&Metric{},
				&LatestMetric{},
			} {
				if err := tx.Migrator().AddColumn(model, "ContextID"); err != nil {
					return eris.Wrapf(err, "error altering column context_id for %t", model)
				}
			}

			return tx.Model(&SchemaVersion{}).
				Where("1 = 1").
				Update("Version", Version).
				Error
		})
	})
}
