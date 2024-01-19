package helpers

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rotisserie/eris"
)

func GenerateDatabaseURI(t *testing.T, backend string) (string, error) {
	switch backend {
	case "sqlite":
		return fmt.Sprintf("sqlite://%s/test.db", t.TempDir()), nil
	case "sqlcipher":
		return fmt.Sprintf("sqlite://%s/test.db?_key=passphrase", t.TempDir()), nil
	case "postgres":
		return getPostgresDatabase(t, GetPostgresUri(),
			strings.ToLower(
				strings.ReplaceAll(t.TempDir(), "/", "_"),
			),
		)
	default:
		return "", fmt.Errorf("unknown backend: %s", backend)
	}
}

func getPostgresDatabase(t *testing.T, dsn string, name string) (string, error) {
	uri, err := url.Parse(dsn)
	if err != nil {
		return "", eris.Wrapf(err, "failed to parse dsn %q", dsn)
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return "", eris.Wrapf(err, "failed to open database %q", dsn)
	}

	_, err = db.Exec("CREATE DATABASE " + name)
	if err != nil {
		return "", eris.Wrapf(err, "failed to create database %q on %q", name, dsn)
	}

	t.Cleanup(func() {
		// nolint:errcheck
		defer db.Close()
		_, err := db.Exec("DROP DATABASE " + name + " WITH (FORCE)")
		if err != nil {
			t.Errorf("failed to drop database %q on %q: %v", name, dsn, err)
		}
	})

	uri.Path = name
	return uri.String(), nil
}
