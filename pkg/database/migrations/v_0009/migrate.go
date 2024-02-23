package v_0009

import (
	"errors"
	"fmt"
	"strings"

	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"
	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/database/migrations"
)

const Version = "2c2299e4e061"

func Migrate(db *gorm.DB) error {
	return migrations.RunWithoutForeignKeyIfNeeded(db, func() error {
		return db.Transaction(func(tx *gorm.DB) error {
			switch tx.Dialector.Name() {
			case sqlite.Dialector{}.Name():
				err := sqliteMigrate(tx)
				if err != nil {
					return err
				}
			case postgres.Dialector{}.Name():
				err := postgresMigrate(tx)
				if err != nil {
					return err
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

func sqliteMigrate(tx *gorm.DB) error {
	tablesIndexes := map[string][]string{
		"metrics":        {"idx_metrics_run_id", "idx_metrics_iter"},
		"latest_metrics": {"idx_latest_metrics_run_id"},
	}
	for table, indexes := range tablesIndexes {
		for _, index := range indexes {
			if err := dropIndex(tx, table, index); err != nil {
				return eris.Wrapf(err, "error dropping %s", index)
			}
		}
		if err := tx.Migrator().RenameTable(table, backupName(table)); err != nil {
			return eris.Wrapf(err, "error renaming %s", table)
		}
	}

	if err := tx.Migrator().AutoMigrate(&Context{}, &Metric{}, &LatestMetric{}); err != nil {
		return eris.Wrap(err, "error automigrating new tables")
	}

	defaultContext, err := createDefaultMetricContext(tx)
	if err != nil {
		return eris.Wrap(err, "error creating default metric context")
	}

	for table := range tablesIndexes {
		if err := tx.Exec(fmt.Sprintf("INSERT INTO %s SELECT *, %d FROM %s",
			table,
			defaultContext.ID,
			backupName(table))).Error; err != nil {
			return eris.Wrapf(err, "error copying data for %s", table)
		}

		var oldRowCount, newRowCount int64
		if err := tx.Table(table).Count(&newRowCount).Error; err != nil {
			return eris.Wrapf(err, "error counting rows for %s", table)
		}
		if err := tx.Table(backupName(table)).Count(&oldRowCount).Error; err != nil {
			return eris.Wrapf(err, "error counting rows for %s", backupName(table))
		}
		if oldRowCount != newRowCount {
			return eris.Errorf("rowcount incorrect for %s (old: %d, new: %d)",
				table, oldRowCount, newRowCount)
		}

		if err := tx.Exec("DROP TABLE " + backupName(table)).Error; err != nil {
			return eris.Wrapf(err, "error dropping %s", backupName(table))
		}
	}
	return nil
}

func postgresMigrate(tx *gorm.DB) error {
	if err := tx.Migrator().AutoMigrate(&Context{}); err != nil {
		return eris.Wrap(err, "error auto migrating context")
	}
	_, err := createDefaultMetricContext(tx)
	if err != nil {
		return eris.Wrap(err, "error creating default metric context")
	}
	if err := tx.Debug().Migrator().AutoMigrate(&Metric{}, &LatestMetric{}); err != nil {
		return eris.Wrap(err, "error auto migrating metrics and latest_metrics")
	}

	tablesKeyCols := map[string][]string{
		"metrics":        {"key", "value", "timestamp", "run_uuid", "step", "is_nan", "context_id"},
		"latest_metrics": {"key", "run_uuid", "context_id"},
	}

	for table, pkCols := range tablesKeyCols {
		pk := fmt.Sprintf("%s_pkey", table)
		if tx.Migrator().HasConstraint(table, pk) {
			if err := tx.Migrator().DropConstraint(table, pk); err != nil {
				return eris.Wrap(err, "error dropping primary key")
			}
		}
		sql := fmt.Sprintf("ALTER TABLE %s ADD PRIMARY KEY (%s)", table, strings.Join(pkCols, ","))
		if err := tx.Exec(sql).Error; err != nil {
			return eris.Wrapf(err, "error creating pk for %s", table)
		}
	}
	return nil
}

func backupName(name string) string {
	return fmt.Sprintf("%s_old", name)
}

func dropIndex(tx *gorm.DB, table, index string) error {
	if tx.Migrator().HasIndex(table, index) {
		if err := tx.Migrator().DropIndex(table, index); err != nil {
			return eris.Wrapf(err, "error dropping %s", index)
		}
	}
	return nil
}

// createDefaultMetricContext creates the default metric context if it doesn't exist.
func createDefaultMetricContext(db *gorm.DB) (*Context, error) {
	defaultContext := Context{Json: datatypes.JSON("{}")}
	if err := db.First(&defaultContext).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Info("Creating default context")
			if err := db.Create(&defaultContext).Error; err != nil {
				return nil, fmt.Errorf("error creating default context: %s", err)
			}
		} else {
			return nil, fmt.Errorf("unable to find default context: %s", err)
		}
	}
	log.Debugf("default metric context: %v", defaultContext)
	return &defaultContext, nil
}
