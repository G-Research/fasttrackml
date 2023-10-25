package v_1ce8669664d2

import (
	"fmt"

	"github.com/G-Research/fasttrackml/pkg/database/migrations"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		constraints := []string{"Params", "Tags", "Metrics", "LatestMetrics"}
		for _, constraint := range constraints {
			switch tx.Dialector.Name() {
			case migrations.SQLiteDialectorName:
				// SQLite tables need to be recreated to add or remove constraints.
				// By not dropping the constraint, we can avoid having to recreate the table twice.
			case migrations.PostgresDialectorName:
				// Existing MLFlow Postgres databases have foreign key constraints
				// with their own names. We need to drop them before we can add our own.
				table := tx.NamingStrategy.TableName(constraint)
				fk := fmt.Sprintf("%s_run_uuid_fkey", table)
				if tx.Migrator().HasConstraint(table, fk) {
					if err := tx.Migrator().DropConstraint(table, fk); err != nil {
						return err
					}
				} else {
					if err := tx.Migrator().DropConstraint(&Run{}, constraint); err != nil {
						return err
					}
				}
			default:
				return fmt.Errorf("unsupported database dialect %s", tx.Dialector.Name())
			}

			if err := tx.Migrator().CreateConstraint(&Run{}, constraint); err != nil {
				return err
			}
		}
		return tx.Model(&SchemaVersion{}).
			Where("1 = 1").
			Update("Version", "1ce8669664d2").
			Error
	})
}
