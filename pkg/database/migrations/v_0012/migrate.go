package v_0012

import (
	"gorm.io/gorm"
)

const Version = "20240228084640"

func Migrate(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// Add new tables
		if err := tx.Migrator().AutoMigrate(&TagData{}); err != nil {
			return err
		}

		// Update the schema version
		return tx.Model(&SchemaVersion{}).
			Where("1 = 1").
			Update("Version", Version).
			Error
	})
}
