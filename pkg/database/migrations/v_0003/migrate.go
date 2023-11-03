package v_0003

import (
	"gorm.io/gorm"
)

const Version = "ed364de02645"

func Migrate(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Migrator().CreateIndex(&Run{}, "RowNum"); err != nil {
			return err
		}
		if err := tx.Migrator().CreateIndex(&Metric{}, "Iter"); err != nil {
			return err
		}
		return tx.Model(&SchemaVersion{}).
			Where("1 = 1").
			Update("Version", Version).
			Error
	})
}
