package database

import (
	"net/url"
	"time"

	"github.com/rotisserie/eris"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// PostgresDBInstance is the Postgres-specific DbInstance variant.
type PostgresDBInstance struct {
	DBInstance
}

// NewPostgresDBInstance constructs a Postgres DbInstance.
func NewPostgresDBInstance(
	dsnURL url.URL, slowThreshold time.Duration, poolMax int, reset bool,
) (*PostgresDBInstance, error) {
	db := PostgresDBInstance{
		DBInstance: DBInstance{dsn: dsnURL.String()},
	}

	conn := postgres.Open(dsnURL.String())

	log.Infof("Using database %s", dsnURL.Redacted())

	gormDB, err := gorm.Open(conn, &gorm.Config{
		Logger: NewLoggerAdaptor(log.StandardLogger(), LoggerAdaptorConfig{
			SlowThreshold:             slowThreshold,
			IgnoreRecordNotFoundError: true,
		}),
	})
	if err != nil {
		return nil, eris.Wrap(err, "failed to connect to database")
	}
	db.DB = gormDB

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, eris.Wrap(err, "failed to get underlying database connection pool")
	}
	sqlDB.SetConnMaxIdleTime(time.Minute)
	sqlDB.SetMaxIdleConns(poolMax)
	sqlDB.SetMaxOpenConns(poolMax)

	return &db, nil
}

// Reset resets database.
func (pgdb PostgresDBInstance) Reset() error {
	log.Info("Resetting database schema")
	if err := pgdb.GormDB().Exec("drop schema public cascade").Error; err != nil {
		return eris.Wrap(err, "error attempting to drop schema")
	}
	if err := pgdb.GormDB().Exec("create schema public").Error; err != nil {
		return eris.Wrap(err, "error attempting to create schema")
	}
	return nil
}
