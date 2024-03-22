package v_0012

import (
	"fmt"

	"gorm.io/gorm"
)

const Version = "20240322124259"

func Migrate(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(
			&RegisteredModel{},
			&RegisteredModelTag{},
			&RegisteredModelAlias{},
			&ModelVersion{},
			&ModelVersionTag{},
		); err != nil {
			return fmt.Errorf("error initializing database: %w", err)
		}

		// Update the schema version
		return tx.Model(&SchemaVersion{}).
			Where("1 = 1").
			Update("Version", Version).
			Error
	})
}
