package database

import (
	"fmt"
	glog "log"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
)

// postgresDbFactory will make Postgres DbInstance.
type postgresDbFactory struct {
	baseDbFactory
}

// makeDbInstance will construct a Postgres DbInstance.
func (f postgresDbFactory) makeDbInstance() (*DbInstance, error) {
	var sourceConn gorm.Dialector
	var replicaConn gorm.Dialector

	db := DbInstance{dsn: f.dsnURL.String()}
	sourceConn = postgres.Open(f.dsnURL.String())

	logURL := f.dsnURL
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
				SlowThreshold:             f.slowThreshold,
				LogLevel:                  dbLogLevel,
				IgnoreRecordNotFoundError: true,
			},
		),
	})
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	db.DB = gormDB

	if replicaConn != nil {
		db.Use(
			dbresolver.Register(dbresolver.Config{
				Replicas: []gorm.Dialector{
					replicaConn,
				},
			}),
		)
	}
	return &db, nil
}
