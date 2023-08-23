package database

import (
	"fmt"
	"net/url"
	"time"

	"github.com/rotisserie/eris"
)

const (
	SQLiteCustomDriverName = "sqlite3_custom_driver"
)

// dbFactory is the interface for all datasource creation.
type dbFactory interface {
	makeDbInstance() (*DbInstance, error)
}

// baseDbFactory are the common attributes for all datasources.
type baseDbFactory struct {
	dsnURL        url.URL
	slowThreshold time.Duration
	poolMax       int
	reset         bool
}

// ConnectDB will establish and return a DbInstance while also caching it in the global
// var database.DB.
func ConnectDB(dsn string, slowThreshold time.Duration, poolMax int, reset bool, migrate bool, artifactRoot string,
) (*DbInstance, error) {
	db, err := MakeDBInstance(dsn, slowThreshold, poolMax, reset, migrate, artifactRoot)
	if err != nil {
		return nil, err
	}
	// set the global DB
	DB = db
	return DB, nil
}

// MakeDbInstance will create a DbInstance from the parameters, without affecting the global var database.DB.
func MakeDBInstance(
	dsn string, slowThreshold time.Duration, poolMax int, reset bool, migrate bool, artifactRoot string,
) (*DbInstance, error) {
	dbFactory, err := newDbInstanceFactory(dsn, slowThreshold, poolMax, reset)
	if err != nil {
		return nil, err
	}
	db, err := dbFactory.makeDbInstance()
	if err != nil {
		return nil, err
	}

	if reset {
		if err := db.reset(); err != nil {
			db.Close()
			return nil, err
		}
	}

	if err := db.checkAndMigrate(migrate); err != nil {
		db.Close()
		return nil, err
	}

	if err := db.createDefaultExperiment(artifactRoot); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// newDbInstanceFactory will return the correct factory type for the datasource URI.
func newDbInstanceFactory(
	dsn string, slowThreshold time.Duration, poolMax int, reset bool,
) (dbFactory, error) {
	dsnURL, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("invalid database URL: %w", err)
	}
	switch dsnURL.Scheme {
	case "sqlite":
		return sqliteDbFactory{
			baseDbFactory: baseDbFactory{
				dsnURL:        *dsnURL,
				slowThreshold: slowThreshold,
				poolMax:       poolMax,
				reset:         reset,
			},
		}, nil
	case "postgres", "postgresql":
		return postgresDbFactory{
			baseDbFactory: baseDbFactory{
				dsnURL:        *dsnURL,
				slowThreshold: slowThreshold,
				poolMax:       poolMax,
				reset:         reset,
			},
		}, nil
	}
	return nil, eris.New("unsupported database type")
}
