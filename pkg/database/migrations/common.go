package common

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DisableForeignKeysIfNeeded disables foreign keys if needed for the migration
func DisableForeignKeysIfNeeded(db *gorm.DB, fn func() error) error {
	switch db.Dialector.Name() {
	case SQLiteDialectorName:
		//nolint:errcheck
		migrator := db.Migrator().(sqlite.Migrator)
		return migrator.RunWithoutForeignKey(fn)
	}
	return fn()
}
