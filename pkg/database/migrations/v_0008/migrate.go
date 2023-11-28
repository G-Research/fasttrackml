package v_0008

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/database/migrations"
)

const Version = "cbc41c0f4fc5"

func Migrate(db *gorm.DB) error {
	// We need to run this migration without foreign key constraints to avoid
	// the cascading delete to kick in and delete all the runs.
	return migrations.RunWithoutForeignKeyIfNeeded(db, func() error {
		return db.Transaction(func(tx *gorm.DB) error {
			switch tx.Dialector.Name() {
			case sqlite.Dialector{}.Name():
				// SQLite no action needed
			case postgres.Dialector{}.Name():
				// Postgres needs to remove this constraint
				constraint := "experiments_name_key"
				if tx.Migrator().HasConstraint("experiments", constraint) {
					if err := tx.Migrator().DropConstraint("experiments", constraint); err != nil {
						return err
					}
				}
			default:
				return fmt.Errorf("unsupported database dialect %s", tx.Dialector.Name())
			}

			return tx.Model(&SchemaVersion{}).
				Where("1 = 1").
				Update("Version", Version).
				Error
		})
	})
}
