package v_0012

import (
	"gorm.io/gorm"
)

const Version = "20240403100850"

func Migrate(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&Role{}); err != nil {
			return err
		}
		if err := tx.AutoMigrate(&RoleNamespace{}); err != nil {
			return err
		}
		return tx.Model(&SchemaVersion{}).
			Where("1 = 1").
			Update("Version", Version).
			Error
	})
}
