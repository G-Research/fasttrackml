package v_0006

import (
	"errors"
	"fmt"

	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/common"
	"github.com/G-Research/fasttrackml/pkg/database/migrations"
)

const Version = "e0d125c68d9a"

func Migrate(db *gorm.DB) error {
	// We need to run this migration without foreign key constraints to avoid
	// the cascading delete to kick in and delete all the runs.
	return migrations.RunWithoutForeignKeyIfNeeded(db, func() error {
		return db.Transaction(func(tx *gorm.DB) error {
			if err := tx.AutoMigrate(&Namespace{}); err != nil {
				return err
			}
			if err := tx.Migrator().AddColumn(&App{}, "NamespaceID"); err != nil {
				return err
			}
			if err := tx.Migrator().CreateConstraint(&Namespace{}, "Apps"); err != nil {
				return err
			}

			switch tx.Dialector.Name() {
			case sqlite.Dialector{}.Name():
				if err := sqliteMigrate(tx); err != nil {
					return err
				}
			case postgres.Dialector{}.Name():
				if err := postgresMigrate(tx); err != nil {
					return err
				}
			}
			if err := tx.Migrator().AlterColumn(&Experiment{}, "Name"); err != nil {
				return err
			}
			if err := tx.Migrator().CreateIndex(&Experiment{}, "Name"); err != nil {
				return err
			}
			return tx.Model(&SchemaVersion{}).
				Where("1 = 1").
				Update("Version", Version).
				Error
		})
	})
}

func sqliteMigrate(tx *gorm.DB) error {
	if err := tx.Migrator().AddColumn(&Experiment{}, "NamespaceID"); err != nil {
		return err
	}
	if err := tx.Migrator().CreateConstraint(&Namespace{}, "Experiments"); err != nil {
		return err
	}
	return nil
}

func postgresMigrate(tx *gorm.DB) error {
	defaultNamespace, err := createDefaultNamespace(tx)
	if err != nil {
		return err
	}
	if err := tx.Exec(
		fmt.Sprintf(`
			ALTER TABLE experiments ADD COLUMN namespace_id BIGINT NOT NULL DEFAULT %d, 
			ADD CONSTRAINT fk_namespaces_experiments FOREIGN KEY (namespace_id) REFERENCES namespaces(id)`,
			defaultNamespace.ID,
		),
	).Error; err != nil {
		return eris.Wrap(err, "error adding namespace_id column for experiments")
	}

	if err := tx.Exec(`ALTER TABLE experiments ALTER COLUMN namespace_id DROP DEFAULT`).Error; err != nil {
		return eris.Wrapf(err, "error dropping default value for `namespace_id` field in `experiments` table")
	}

	return nil
}

// createDefaultMetricContext creates the default metric context if it doesn't exist.
func createDefaultNamespace(db *gorm.DB) (*Namespace, error) {
	defaultNamespace := Namespace{
		Code:                "default",
		Description:         "Default namespace",
		DefaultExperimentID: common.GetPointer(int32(0)),
	}
	if err := db.First(&defaultNamespace).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Info("Creating default context")
			if err := db.Create(&defaultNamespace).Error; err != nil {
				return nil, fmt.Errorf("error creating default context: %s", err)
			}
		} else {
			return nil, fmt.Errorf("unable to find default context: %s", err)
		}
	}
	log.Debugf("default namespace: %v", defaultNamespace)
	return &defaultNamespace, nil
}
