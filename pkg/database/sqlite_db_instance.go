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
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
)

const (
	SQLiteCustomDriverName = "sqlite3_custom_driver"
)

// SqliteDBInstance is the sqlite specific variant of DbInstance.
type SqliteDBInstance struct {
	DBInstance
}

// NewSqliteDBInstance creates a SqliteDBInstance.
func NewSqliteDBInstance(
	dsnURL url.URL, slowThreshold time.Duration, poolMax int, reset bool,
) (*SqliteDBInstance, error) {
	db := SqliteDBInstance{
		DBInstance: DBInstance{dsn: dsnURL.String()},
	}
	var sourceConn gorm.Dialector
	var replicaConn gorm.Dialector

	q := dsnURL.Query()
	q.Set("_case_sensitive_like", "true")
	q.Set("_mutex", "no")
	if q.Get("mode") != "memory" && !(q.Has("_journal") || q.Has("_journal_mode")) {
		q.Set("_journal", "WAL")
	}
	dsnURL.RawQuery = q.Encode()

	if reset && q.Get("mode") != "memory" {
		file := dsnURL.Host
		if file == "" {
			file = dsnURL.Path
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

	s, err := sql.Open(SQLiteCustomDriverName, strings.Replace(dsnURL.String(), "sqlite://", "file:", 1))
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
	dsnURL.RawQuery = q.Encode()
	r, err := sql.Open(SQLiteCustomDriverName, strings.Replace(dsnURL.String(), "sqlite://", "file:", 1))
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	db.closers = append(db.closers, r)
	replicaConn = sqlite.Dialector{
		Conn: r,
	}

	logURL := dsnURL
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
				SlowThreshold:             slowThreshold,
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

	return &db, nil
}

// Reset resets database.
func (f SqliteDBInstance) Reset() error {
	return eris.New("reset for sqlite database not supported")
}
