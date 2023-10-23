//go:build integration

package database

import (
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/G-Research/fasttrackml/pkg/database"
)

func TestMigrate(t *testing.T) {
	// copy the original mlflow db to a temp file
	mlflowDbPath := t.TempDir()
	src, err := os.Open("mlflow.db")
	assert.Nil(t, err)
	defer func() {
		err := src.Close()
		assert.Nil(t, err)
	}()

	dst, err := os.Create(fmt.Sprintf("%s/mlflow.db", mlflowDbPath))
	assert.Nil(t, err)
	defer func() {
		err := dst.Close()
		assert.Nil(t, err)
	}()

	_, err = io.Copy(dst, src)
	assert.Nil(t, err)

	// make DbProvider using the temp copy
	db, err := database.NewDBProvider(
		fmt.Sprintf("sqlite://%s/mlflow.db", mlflowDbPath),
		1*time.Second,
		20,
	)
	assert.Nil(t, err)

	// run migrations
	assert.Nil(t, database.CheckAndMigrateDB(true, db.GormDB()))
}
