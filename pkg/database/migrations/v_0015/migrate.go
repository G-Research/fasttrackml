package v_0015

import (
	"gorm.io/gorm"
)

const Version = "20240604123811"

func Migrate(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Migrator().AutoMigrate(&Log{}); err != nil {
			return err
		}

		// Update the schema version
		return tx.Model(&SchemaVersion{}).
			Where("1 = 1").
			Update("Version", Version).
			Error
	})
}
