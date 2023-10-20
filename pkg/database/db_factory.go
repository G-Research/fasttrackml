package database

import (
	"net/url"
	"time"

	"github.com/rotisserie/eris"
)

// NewDBProvider creates a DBProvider of the correct type from the parameters.
func NewDBProvider(
	dsn string, slowThreshold time.Duration, poolMax int,
) (db DBProvider, err error) {
	dsnURL, err := url.Parse(dsn)
	if err != nil {
		return nil, eris.Wrap(err, "invalid database URL")
	}
	switch dsnURL.Scheme {
	case SQLiteSchemaName:
		db, err = NewSqliteDBInstance(
			*dsnURL,
			slowThreshold,
			poolMax,
		)
		if err != nil {
			return nil, eris.Wrap(err, "error creating sqlite provider")
		}
	case PostgresSchemaName, PostgresQLSchemaName:
		db, err = NewPostgresDBInstance(
			*dsnURL,
			slowThreshold,
			poolMax,
		)
		if err != nil {
			return nil, eris.Wrap(err, "error creating postgres provider")
		}
	default:
		return nil, eris.New("unsupported database type")
	}

	return db, nil
}
