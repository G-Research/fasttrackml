package v_0005

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/database/migrations"
)

const Version = "5d042539be4f"

func Migrate(db *gorm.DB) error {
	// We need to run this migration without foreign key constraints to avoid
	// the cascading delete to kick in and delete all the run data.
	return migrations.RunWithoutForeignKeyIfNeeded(db, func() error {
		return db.Transaction(func(tx *gorm.DB) error {
			elems := []struct {
				Table      string
				Constraint string
			}{
				{"experiment_tags", "Tags"},
				{"runs", "Runs"},
			}
			for _, e := range elems {
				switch tx.Dialector.Name() {
				case sqlite.Dialector{}.Name():
					// SQLite tables need to be recreated to add or remove constraints.
					// By not dropping the constraint, we can avoid having to recreate the table twice.
				case postgres.Dialector{}.Name():
					// Existing MLFlow Postgres databases have foreign key constraints with their own names.
					// We need to drop them before we can add our own.
					fk := fmt.Sprintf("%s_experiment_id_fkey", e.Table)
					if tx.Migrator().HasConstraint(e.Table, fk) {
						if err := tx.Migrator().DropConstraint(e.Table, fk); err != nil {
							return err
						}
					} else {
						if err := tx.Migrator().DropConstraint(&Experiment{}, e.Constraint); err != nil {
							return err
						}
					}
				default:
					return fmt.Errorf("unsupported database dialect %s", tx.Dialector.Name())
				}

				if err := tx.Migrator().CreateConstraint(&Experiment{}, e.Constraint); err != nil {
					return err
				}
			}
			return tx.Model(&SchemaVersion{}).
				Where("1 = 1").
				Update("Version", Version).
				Error
		})
	})
}
