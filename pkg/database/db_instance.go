package database

import (
	"io"

	"gorm.io/gorm"
)

// DBProvider is the interface to access the DB.
type DBProvider interface {
	GormDB() *gorm.DB
	Dsn() string
	Close() error
	Reset() error
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

// Dsn returns the dsn string.
func (db *DBInstance) Dsn() string {
	return db.dsn
}

// GormDB returns the gorm DB.
func (db *DBInstance) GormDB() *gorm.DB {
	return db.DB
}
