package v_0c779009ac13

import (
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Migrator().AddColumn(&Run{}, "DeletedTime"); err != nil {
			return err
		}
		return tx.Model(&AlembicVersion{}).
			Where("1 = 1").
			Update("Version", "0c779009ac13").
			Error
	})
}
