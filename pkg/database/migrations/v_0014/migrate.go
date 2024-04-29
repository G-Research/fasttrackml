package v_0014

import (
	"gorm.io/gorm"
)

const Version = "20240429012448"

func Migrate(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {

                // rename the existing Value column to ValueString
                if err := tx.Migrator().RenameColumn(&Param{}, "value", "value_str"); err != nil {
                        return err
                }

                // add the new Value columns and remove not null constraint.
                if err := tx.Migrator().AutoMigrate(&Param{}); err != nil {
                        return err
                }

		// Update the schema version
		return tx.Model(&SchemaVersion{}).
			Where("1 = 1").
			Update("Version", Version).
			Error
	})
}
