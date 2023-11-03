package v_8073e7e037e5

import (
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(
			&Dashboard{},
			&App{},
		); err != nil {
			return err
		}
		return tx.Model(&SchemaVersion{}).
			Where("1 = 1").
			Update("Version", "8073e7e037e5").
			Error
	})
}
