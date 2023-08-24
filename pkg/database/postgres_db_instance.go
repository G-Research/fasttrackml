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

type PostgresDbInstance struct {
	DbInstance
}

// Reset will provide type-specific reset
func (pgdb PostgresDbInstance) Reset() error {
	log.Info("Resetting database schema")
	pgdb.Db().Exec("drop schema public cascade")
	pgdb.Db().Exec("create schema public")
	return nil
}

// makeDbInstance will construct a Postgres DbInstance.
func NewPostgresDbInstance(
	dsnURL url.URL, slowThreshold time.Duration, poolMax int, reset bool,
) (*PostgresDbInstance, error) {
	pgdb := PostgresDbInstance{
		DbInstance: DbInstance{dsn: dsnURL.String()},
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
