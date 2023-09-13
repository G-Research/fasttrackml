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

// NewPostgresDBInstance constructs a Postgres DbInstance.
func NewPostgresDBInstance(
	dsnURL url.URL, slowThreshold time.Duration, poolMax int, reset bool,
) (*PostgresDBInstance, error) {
	logURL := dsnURL
	q := logURL.Query()
	if q.Has("_key") {
		q.Set("_key", "xxxxx")
	}
	logURL.RawQuery = q.Encode()
	log.Infof("using database %s", logURL.Redacted())

	dbLogLevel := logger.Warn
	if log.GetLevel() == log.DebugLevel {
		dbLogLevel = logger.Info
	}
	gormDB, err := gorm.Open(postgres.Open(dsnURL.String()), &gorm.Config{
		Logger: logger.New(
			glog.New(
				log.StandardLogger().WriterLevel(log.WarnLevel),
				"",
				0,
			),
			logger.Config{
				LogLevel:                  dbLogLevel,
				SlowThreshold:             slowThreshold,
				IgnoreRecordNotFoundError: true,
			},
		),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &PostgresDBInstance{
		DBInstance: DBInstance{
			db:  gormDB,
			dsn: dsnURL.String(),
		},
	}, nil
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
