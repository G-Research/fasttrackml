//go:build integration

package database

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/G-Research/fasttrackml/pkg/database"
)

func TestMigrate(t *testing.T) {
	// setup sqlite MLFlow database from the schema
	mlflowDBPath := fmt.Sprintf("%s/mlflow.db", t.TempDir())
	mlflowDB, err := sql.Open("sqlite3", mlflowDBPath)
	assert.Nil(t, err)

	mlflowSql, err := os.ReadFile("mlflow.sql")
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
