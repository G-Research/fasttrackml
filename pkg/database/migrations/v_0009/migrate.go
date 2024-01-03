package v_0009

import (
	"fmt"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/database/migrations"
)

const Version = "2c2299e4e061"

func Migrate(db *gorm.DB) error {
	return migrations.RunWithoutForeignKeyIfNeeded(db, func() error {
		return db.Transaction(func(tx *gorm.DB) error {
			// Rename the existing tables and indexes
			tables := []string{"metrics", "latest_metrics"}
			for _, table := range tables {
				index := fmt.Sprintf("idx_%s_run_id", table)
				if err := tx.Migrator().DropIndex(table, index); err != nil {
					return eris.Wrapf(err, "error renaming %s", index)
				}
				if err := tx.Migrator().RenameTable(table, backupName(table)); err != nil {
					return eris.Wrapf(err, "error renaming %s", table)
				}
			}
			index := "idx_metrics_iter"
			if err := tx.Migrator().DropIndex(backupName("metrics"), index); err != nil {
				return eris.Wrapf(err, "error renaming %s", index)
			}

			// Auto-migrate the new tables
			if err := tx.Migrator().AutoMigrate(&Context{}, &Metric{}, &LatestMetric{}); err != nil {
				return eris.Wrap(err, "error automigrating new tables")
			}

			// Copy the data from the old tables to the new ones
			for _, table := range tables {
				if err := tx.Exec(fmt.Sprintf("INSERT INTO %s SELECT *, '0' FROM %s", table, backupName(table))).Error; err != nil {
					return eris.Wrapf(err, "error copying data for %s", table)
				}

				// Drop the backup tables
				if err := tx.Exec("DROP TABLE " + backupName(table)).Error; err != nil {
					return eris.Wrapf(err, "error dropping %s", backupName(table))
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

func backupName(name string) string {
	return fmt.Sprintf("%s_old", name)
}
