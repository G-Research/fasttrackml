package v_cc1f77228345

import (
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Migrator().AlterColumn(&Param{}, "value"); err != nil {
			return err
		}
		return tx.Model(&AlembicVersion{}).
			Where("1 = 1").
			Update("Version", "cc1f77228345").
			Error
	})
}
