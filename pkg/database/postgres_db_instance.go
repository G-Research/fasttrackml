package database

import (
	"fmt"
	glog "log"
	"net/url"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
)

// PostgresDBInstance is the Postgres-specific DbInstance variant.
type PostgresDBInstance struct {
	DBInstance
}

// Reset implementation for this type.
func (pgdb PostgresDBInstance) Reset() error {
	log.Info("Resetting database schema")
	pgdb.GormDB().Exec("drop schema public cascade")
	pgdb.GormDB().Exec("create schema public")
	return nil
}

// NewPostgresDBInstance will construct a Postgres DbInstance.
func NewPostgresDBInstance(
	dsnURL url.URL, slowThreshold time.Duration, poolMax int, reset bool,
) (*PostgresDBInstance, error) {
	pgdb := PostgresDBInstance{
		DBInstance: DBInstance{dsn: dsnURL.String()},
	}

	var sourceConn gorm.Dialector
	var replicaConn gorm.Dialector
	sourceConn = postgres.Open(dsnURL.String())

	logURL := dsnURL
	q := logURL.Query()
	if q.Has("_key") {
		q.Set("_key", "xxxxx")
	}
	logURL.RawQuery = q.Encode()
	log.Infof("Using database %s", logURL.Redacted())

	dbLogLevel := logger.Warn
	if log.GetLevel() == log.DebugLevel {
		dbLogLevel = logger.Info
	}
	gormDB, err := gorm.Open(sourceConn, &gorm.Config{
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

	if replicaConn != nil {
		pgdb.Use(
			dbresolver.Register(dbresolver.Config{
				Replicas: []gorm.Dialector{
					replicaConn,
				},
			}),
		)
	}
	return &pgdb, nil
}
