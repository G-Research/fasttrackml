package v_0011

import (
	"gorm.io/gorm"
)

const Version = "8x230yiog1gv"

func Migrate(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Migrator().AlterColumn(&Param{}, "Value"); err != nil {
			return err
		}
		return tx.Model(&SchemaVersion{}).
			Where("1 = 1").
			Update("Version", Version).
			Error
	})
}
