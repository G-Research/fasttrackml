//go:build integration

package database

import (
	"database/sql"
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/database"
)

type MigrateTestSuite struct {
	suite.Suite
}

func TestMigrateTestSuite(t *testing.T) {
	suite.Run(t, new(MigrateTestSuite))
}

func (s *MigrateTestSuite) TestMigrate() {
	tests := []struct {
		name   string
		schema string
	}{
		{
			name:   "MigrateFromMLFlow2.8.0",
			schema: "mlflow-7f2a7d5fae7d-v2.8.0.sql",
		},
		{
			name:   "MigrateFromMLFlow1.16.0",
			schema: "mlflow-c48cb773bb87-v1.16.0.sql",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			// setup sqlite MLFlow database from the schema
			mlflowDBPath := path.Join(s.T().TempDir(), "mlflow.db")
			mlflowDB, err := sql.Open("sqlite3", mlflowDBPath)
			require.Nil(s.T(), err)

			//nolint:gosec
			mlflowSql, err := os.ReadFile(tt.schema)
			require.Nil(s.T(), err)

			_, err = mlflowDB.Exec(string(mlflowSql))
			require.Nil(s.T(), err)
			require.Nil(s.T(), mlflowDB.Close())

			// make DbProvider using our package
			db, err := database.NewDBProvider(
				fmt.Sprintf("sqlite://%s", mlflowDBPath),
				1*time.Second,
				20,
			)
			require.Nil(s.T(), err)

			// run migrations
			require.Nil(s.T(), database.CheckAndMigrateDB(true, db.GormDB()))
		})
	}
}
