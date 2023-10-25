package v_97727af70f4d

import (
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, column := range []string{
			"CreationTime",
			"LastUpdateTime",
		} {
			if err := tx.Migrator().AddColumn(&Experiment{}, column); err != nil {
				return err
			}
		}
		return tx.Model(&AlembicVersion{}).
			Where("1 = 1").
			Update("Version", "97727af70f4d").
			Error

	})
}
