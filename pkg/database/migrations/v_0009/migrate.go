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
			defaultContext := &Context{
				Json: []byte(`{}`),
			}
			if err := tx.Migrator().AutoMigrate(defaultContext); err != nil {
				return err
			}
			if err := tx.FirstOrCreate(defaultContext).Error; err != nil {
				return eris.Wrapf(err, "error creating default context")
			}

			for _, model := range []interface{}{
				&Metric{},
				&LatestMetric{},
			} {
				if err := tx.Migrator().AddColumn(model, "ContextID"); err != nil {
					return eris.Wrapf(err, "error altering column context_id for %t", model)
				}

				if err := tx.Model(model).
					Where("context_id IS NULL").
					Update("context_id", defaultContext.ID).
					Error; err != nil {
					return eris.Wrapf(err, "error updating %t", model)
				}
			}

			return tx.Model(&SchemaVersion{}).
				Where("1 = 1").
				Update("Version", Version).
				Error
		})
	})
}
