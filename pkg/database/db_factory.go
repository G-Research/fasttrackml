package database

import (
	"database/sql"
	"errors"
	"fmt"
	glog "log"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/rotisserie/eris"
	"golang.org/x/exp/slices"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
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

// postgresDbFactory will make Postgres DbInstance.
type postgresDbFactory struct {
	baseDbFactory
}

// sqliteDbFactory will make Sqlite DbInstance.
type sqliteDbFactory struct {
	baseDbFactory
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

// makeDbInstance will create a Sqlite DbInstance
func (f sqliteDbFactory) makeDbInstance() (*DbInstance, error) {
	var sourceConn gorm.Dialector
	var replicaConn gorm.Dialector
	db := DbInstance{dsn: f.dsnURL.String()}
	q := f.dsnURL.Query()
	q.Set("_case_sensitive_like", "true")
	q.Set("_mutex", "no")
	if q.Get("mode") != "memory" && !(q.Has("_journal") || q.Has("_journal_mode")) {
		q.Set("_journal", "WAL")
	}
	f.dsnURL.RawQuery = q.Encode()

	if f.reset && q.Get("mode") != "memory" {
		file := f.dsnURL.Host
		if file == "" {
			file = f.dsnURL.Path
		}
		log.Infof("Removing database file %s", file)
		if err := os.Remove(file); err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("failed to remove database file: %w", err)
		}
	}

	if !slices.Contains(sql.Drivers(), SQLiteCustomDriverName) {
		sql.Register(SQLiteCustomDriverName, &sqlite3.SQLiteDriver{
			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
				// create LRU cache to cache regexp statements and results.
				cache, err := lru.New[string, *regexp.Regexp](1000)
				if err != nil {
					return eris.Wrap(err, "error creating lru cache to cache regexp statements")
				}
				return conn.RegisterFunc("regexp", func(re, s string) bool {
					result, ok := cache.Get(re)
					if !ok {
						result, err = regexp.Compile(re)
						if err != nil {
							return false
						}
						cache.Add(re, result)
					}
					return result.MatchString(s)
				}, true)
			},
		})
	}

	s, err := sql.Open(SQLiteCustomDriverName, strings.Replace(f.dsnURL.String(), "sqlite://", "file:", 1))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	db.closers = append(db.closers, s)
	s.SetMaxIdleConns(1)
	s.SetMaxOpenConns(1)
	s.SetConnMaxIdleTime(0)
	s.SetConnMaxLifetime(0)
	sourceConn = sqlite.Dialector{
		Conn: s,
	}

	q.Set("_query_only", "true")
	f.dsnURL.RawQuery = q.Encode()
	r, err := sql.Open(SQLiteCustomDriverName, strings.Replace(f.dsnURL.String(), "sqlite://", "file:", 1))
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	db.closers = append(db.closers, r)
	replicaConn = sqlite.Dialector{
		Conn: r,
	}

	logURL := f.dsnURL
	q = logURL.Query()
	if q.Has("_key") {
		q.Set("_key", "xxxxx")
	}
	logURL.RawQuery = q.Encode()
	log.Infof("Using database %s", logURL.Redacted())

	dbLogLevel := logger.Warn
	if log.GetLevel() == log.DebugLevel {
		dbLogLevel = logger.Info
	}
	db.DB, err = gorm.Open(sourceConn, &gorm.Config{
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

	db.Use(
		dbresolver.Register(dbresolver.Config{
			Replicas: []gorm.Dialector{
				replicaConn,
			},
		}),
	)

	sqlDB, _ := db.DB.DB()
	sqlDB.SetConnMaxIdleTime(time.Minute)
	sqlDB.SetMaxIdleConns(f.poolMax)
	sqlDB.SetMaxOpenConns(f.poolMax)

	return &db, nil
}
