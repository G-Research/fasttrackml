package v_0009

import (
	"github.com/rotisserie/eris"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/database/migrations"
)

const Version = "2c2299e4e061"

func Migrate(db *gorm.DB) error {
	return migrations.RunWithoutForeignKeyIfNeeded(db, func() error {
		return db.Transaction(func(tx *gorm.DB) error {
			// Rename the existing tables for backup
			tables := []string{ "metrics", "latest_metrics"}
			for _, table := range tables {
				backup_name := fmt.Sprintf("%s_old", table)
				if err := tx.Migrator().RenameTable(table, backup_name); err != nil {
					return eris.Wrapf(err, "error renaming %s", table)
				}
			}

			// Auto-migrate the new tables
			if err := tx.Migrator().AutoMigrate(&Context{}, &Metric{}, &LatestMetric{}); err != nil {
				return eris.Wrap(err, "error automigrating new tables")
			}

			// Copy the data from the old tables to the new ones
			for _, table := tables {
				if err := tx.Exec("INSERT INTO metrics SELECT *, '0' FROM old_metrics").Error; err != nil {
					return eris.Wrapf(err, "error copying data for %s", table)
				}

				// Drop the backup tables
				if err := tx.Exec("DROP TABLE old_metrics").Error; err != nil {
					return eris.Wrapf(err, "error dropping %s", table)
				}
			}

			// Update the schema version
			return tx.Model(&SchemaVersion{}).
				Where("1 = 1").
				Update("Version", Version).
				Error
		})
	})
}
