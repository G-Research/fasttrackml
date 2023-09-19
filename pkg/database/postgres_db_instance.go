package database

import (
	"fmt"
	glog "log"
	"net/url"
	"time"

	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// PostgresDBInstance is the Postgres-specific DbInstance variant.
type PostgresDBInstance struct {
	DBInstance
}

// Reset implementation for this type.
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

// NewPostgresDBInstance will construct a Postgres DbInstance.
func NewPostgresDBInstance(
	dsnURL url.URL, slowThreshold time.Duration, poolMax int, reset bool,
) (*PostgresDBInstance, error) {
	pgdb := PostgresDBInstance{
		DBInstance: DBInstance{dsn: dsnURL.String()},
	}

	conn := postgres.Open(dsnURL.String())

	log.Infof("Using database %s", dsnURL.Redacted())

	dbLogLevel := logger.Warn
	if log.GetLevel() == log.DebugLevel {
		dbLogLevel = logger.Info
	}
	gormDB, err := gorm.Open(conn, &gorm.Config{
		Logger: logger.New(
			glog.New(
				log.StandardLogger().WriterLevel(log.WarnLevel),
				"",
				0,
			),
			logger.Config{
				SlowThreshold:             slowThreshold,
				LogLevel:                  dbLogLevel,
				IgnoreRecordNotFoundError: true,
			},
		),
	})
	if err != nil {
		pgdb.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	pgdb.DB = gormDB

	return &pgdb, nil
}
