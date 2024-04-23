package v_0013

import (
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/database/migrations"
)

const Version = "20240423055442"

func Migrate(db *gorm.DB) error {
	return migrations.RunWithoutForeignKeyIfNeeded(db, func() error {
		return db.Transaction(func(tx *gorm.DB) error {
			if err := tx.AutoMigrate(&Role{}); err != nil {
				return err
			}
			if err := tx.AutoMigrate(&RoleNamespace{}); err != nil {
				return err
			}
			// Update the schema version
			return tx.Model(&SchemaVersion{}).
				Where("1 = 1").
				Update("Version", Version).
				Error
		})
	})
}
