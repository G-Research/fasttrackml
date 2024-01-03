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
			for _, model := range []interface{}{
				&Context{},
				&Metric{},
				&LatestMetric{},
			} {
				if err := tx.Migrator().AutoMigrate(model); err != nil {
					return eris.Wrapf(err, "error altering table %t", model)
				}
			}

			return tx.Model(&SchemaVersion{}).
				Where("1 = 1").
				Update("Version", Version).
				Error
		})
	})
}
