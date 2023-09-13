package database

import (
	"net/url"
	"time"

	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"
)

// NewDBProvider creates a DBProvider of the correct type from the parameters.
func NewDBProvider(
	dsn string, slowThreshold time.Duration, poolMax int, reset bool,
) (db DBProvider, err error) {
	dsnURL, err := url.Parse(dsn)
	if err != nil {
		return nil, eris.Wrap(err, "invalid database URL")
	}
	switch dsnURL.Scheme {
	case "sqlite":
		db, err = NewSqliteDBInstance(
			*dsnURL,
			slowThreshold,
			poolMax,
			reset,
		)
		if err != nil {
			return nil, eris.Wrap(err, "error creating sqlite provider")
		}
	case "postgres", "postgresql":
		db, err = NewPostgresDBInstance(
			*dsnURL,
			slowThreshold,
			poolMax,
			reset,
		)
		if err != nil {
			return nil, eris.Wrap(err, "error creating postgres provider")
		}
	default:
		return nil, eris.New("unsupported database type")
	}

	// TODO:DSuhinin - it shouldn't be there. NewDBProvider has to only create an instance without any hidden logic.
	if reset {
		log.Infof("reseting database")
		if err := db.Reset(); err != nil {
			db.Close()
			return nil, eris.Wrap(err, "error resetting database")
		}
	}

	return db, nil
}
