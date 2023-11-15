package migrations

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// RunWithoutForeignKeyIfNeeded disables foreign keys if needed for the migration
func RunWithoutForeignKeyIfNeeded(db *gorm.DB, fn func() error) error {
	switch db.Dialector.Name() {
	case sqlite.Dialector{}.Name():
		//nolint:errcheck
		migrator := db.Migrator().(sqlite.Migrator)
		return migrator.RunWithoutForeignKey(fn)
	}
	return fn()
}
