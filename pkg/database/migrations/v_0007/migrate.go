package v_0007

import (
	"gorm.io/gorm"
)

const Version = "2c2299e4e061"

func Migrate(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Migrator().AddColumn(&Metric{}, "Context"); err != nil {
			return err
		}
		if err := tx.Migrator().AddColumn(&LatestMetric{}, "Context"); err != nil {
			return err
		}
		return tx.Model(&SchemaVersion{}).
			Where("1 = 1").
			Update("Version", Version).
			Error
	})
}
