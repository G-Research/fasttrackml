package v_0008

import (
	"gorm.io/gorm"
)

const Version = "2c2299e4e061"

func Migrate(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Migrator().AddColumn(&Metric{}, "ContextID"); err != nil {
			return err
		}
		if err := tx.Migrator().AddColumn(&LatestMetric{}, "ContextID"); err != nil {
			return err
		}
		if err := tx.Migrator().AutoMigrate(&Context{}); err != nil {
			return err
		}
		return tx.Model(&SchemaVersion{}).
			Where("1 = 1").
			Update("Version", Version).
			Error
	})
}
