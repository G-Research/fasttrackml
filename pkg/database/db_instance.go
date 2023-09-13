package database

import (
	"io"

	"gorm.io/gorm"
)

// DBProvider is the interface to access the DB.
type DBProvider interface {
	DSN() string
	Close() error
	Reset() error
	GormDB() *gorm.DB
}

// DB is a global gorm.DB reference
var DB *gorm.DB

// DBInstance is the base concrete type for DbProvider.
type DBInstance struct {
	*gorm.DB
	dsn     string
	closers []io.Closer
}

// Close invokes the closers.
func (db *DBInstance) Close() error {
	for _, c := range db.closers {
		err := c.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// DSN returns the dsn string.
func (db *DBInstance) DSN() string {
	return db.dsn
}

// GormDB returns the gorm DB.
func (db *DBInstance) GormDB() *gorm.DB {
	return db.DB
}
