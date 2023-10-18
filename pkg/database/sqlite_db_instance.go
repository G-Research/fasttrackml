package database

import (
	"database/sql"
	"errors"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/mattn/go-sqlite3"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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

	query := dsnURL.Query()
	query.Set("_case_sensitive_like", "true")
	query.Set("_mutex", "no")
	if query.Get("mode") != "memory" && !(query.Has("_journal") || query.Has("_journal_mode")) {
		query.Set("_journal", "WAL")
	}
	sourceURL := dsnURL
	sourceURL.RawQuery = query.Encode()

	if reset && query.Get("mode") != "memory" {
		file := sourceURL.Host
		if file == "" {
			file = sourceURL.Path
		}
		log.Infof("Removing database file %s", file)
		if err := os.Remove(file); err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, eris.Wrap(err, "failed to remove database file")
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

	sourceDB, err := sql.Open(SQLiteCustomDriverName, strings.Replace(sourceURL.String(), "sqlite://", "file:", 1))
	if err != nil {
		return nil, eris.Wrap(err, "failed to connect to database")
	}
	db.closers = append(db.closers, sourceDB)
	sourceDB.SetMaxIdleConns(1)
	sourceDB.SetMaxOpenConns(1)
	sourceDB.SetConnMaxIdleTime(0)
	sourceDB.SetConnMaxLifetime(0)
	sourceConn = sqlite.Dialector{
		Conn: sourceDB,
	}

	query.Set("_query_only", "true")
	replicaURL := dsnURL
	replicaURL.RawQuery = query.Encode()
	replicaDB, err := sql.Open(SQLiteCustomDriverName, strings.Replace(replicaURL.String(), "sqlite://", "file:", 1))
	if err != nil {
		//nolint:errcheck,gosec
		db.Close()
		return nil, eris.Wrap(err, "failed to connect to database")
	}
	db.closers = append(db.closers, replicaDB)
	replicaConn = sqlite.Dialector{
		Conn: replicaDB,
	}

	logURL := dsnURL
	query = logURL.Query()
	if query.Has("_key") {
		query.Set("_key", "xxxxx")
	}
	logURL.RawQuery = query.Encode()
	log.Infof("Using database %s", logURL.Redacted())

	db.DB, err = gorm.Open(sourceConn, &gorm.Config{
		Logger: NewLoggerAdaptor(log.StandardLogger(), LoggerAdaptorConfig{
			SlowThreshold:             slowThreshold,
			IgnoreRecordNotFoundError: true,
		}),
	})
	if err != nil {
		//nolint:errcheck,gosec
		db.Close()
		return nil, eris.Wrap(err, "failed to connect to database")
	}

	if err := db.Use(
		dbresolver.Register(dbresolver.Config{
			Replicas: []gorm.Dialector{
				replicaConn,
			},
		}),
	); err != nil {
		return nil, eris.Wrap(err, "error attaching plugin")
	}

	return &db, nil
}

// Reset resets database.
func (f SqliteDBInstance) Reset() error {
	return eris.New("reset for sqlite database not supported")
}
