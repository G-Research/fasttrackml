package v_0006

import (
	"gorm.io/gorm"

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
			if err := tx.Migrator().AddColumn(&Experiment{}, "NamespaceID"); err != nil {
				return err
			}
			if err := tx.Migrator().CreateConstraint(&Namespace{}, "Experiments"); err != nil {
				return err
			}
			if err := tx.Migrator().AlterColumn(&Experiment{}, "Name"); err != nil {
				return err
			}
			if err := tx.Migrator().CreateIndex(&Experiment{}, "idx_namespace_name"); err != nil {
				return err
			}
			return tx.Model(&SchemaVersion{}).
				Where("1 = 1").
				Update("Version", Version).
				Error
		})
	})
}
