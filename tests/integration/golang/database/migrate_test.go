//go:build integration

package database

import (
	"database/sql"
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/G-Research/fasttrackml/pkg/database"
)

func TestMigrate(t *testing.T) {
	testMigrateWithSchema(t, "mlflow-c48cb773bb87-v1.16.0.sql")
	testMigrateWithSchema(t, "mlflow-7f2a7d5fae7d-v2.8.0.sql")
}

func testMigrateWithSchema(t *testing.T, schema string) {
	// setup sqlite MLFlow database from the schema
	mlflowDBPath := path.Join(t.TempDir(), "mlflow.db")
	mlflowDB, err := sql.Open("sqlite3", mlflowDBPath)
	assert.Nil(t, err)

	mlflowSql, err := os.ReadFile(schema)
	assert.Nil(t, err)

	_, err = mlflowDB.Exec(string(mlflowSql))
	assert.Nil(t, err)
	assert.Nil(t, mlflowDB.Close())

	// make DbProvider using our package
	db, err := database.NewDBProvider(
		fmt.Sprintf("sqlite://%s", mlflowDBPath),
		1*time.Second,
		20,
	)
	assert.Nil(t, err)

	// run migrations
	assert.Nil(t, database.CheckAndMigrateDB(true, db.GormDB()))
}
