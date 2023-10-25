package v_e0d125c68d9a

import (
	"gorm.io/gorm"
)

func MigrateWithWrapper(db *gorm.DB, wrapper func(fn func() error) error) error {
	if err := wrapper(func() error {
		return Migrate(db)
	}); err != nil {
		return err
	}
	return nil
}

func Migrate(db *gorm.DB) error {
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
			Update("Version", "e0d125c68d9a").
			Error
	})
}
