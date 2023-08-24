package database

import (
	"fmt"
	"net/url"
	"time"

	"github.com/rotisserie/eris"
)

// MakeDbProvider will create a DbInstance from the parameters
func MakeDbProvider(
	dsn string, slowThreshold time.Duration, poolMax int, reset bool, migrate bool, artifactRoot string,
) (db DbProvider, err error) {
	dsnURL, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("invalid database URL: %w", err)
	}
	switch dsnURL.Scheme {
	case "sqlite":
		db, err = NewSqliteDbInstance(
			*dsnURL,
			slowThreshold,
			poolMax,
			reset,
		)
		if err != nil {
			return nil, eris.Wrap(err, "error creating sqlite provider")
		}
	case "postgres", "postgresql":
		db, err = NewPostgresDbInstance(
			*dsnURL,
			slowThreshold,
			poolMax,
			reset,
		)
		if err != nil {
			return nil, eris.Wrap(err, "error creating postgres provider")
		}
	default:
		{
			return nil, eris.New("unsupported database type")
		}
	}

	if reset {
		if err := db.Reset(); err != nil {
			db.Close()
			return nil, err
		}
	}

	if err := checkAndMigrate(migrate, db); err != nil {
		db.Close()
		return nil, err
	}

	if err := createDefaultExperiment(artifactRoot, db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
