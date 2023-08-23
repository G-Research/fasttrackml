package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConnectDB(t *testing.T) {
	tests := []struct {
		name              string
		dsn               string
		expectedDialector string
	}{
		{
			name:              "WithSqliteURI",
			dsn:               "sqlite:///tmp/fasttrack-tmp.db",
			expectedDialector: "sqlite",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DB = nil
			db, err := ConnectDB(
				tt.dsn,
				time.Second*2,
				2,
				false,
				false,
				"s3://somewhere",
			)
			assert.Nil(t, err)
			assert.NotNil(t, db)
			assert.Equal(t, tt.expectedDialector, db.DB.Dialector.Name())

			// expecting the global 'DB' to be set to the same pointer
			assert.Equal(t, db, DB)

			err = db.Close()
			assert.Nil(t, err)
		})
	}
}

func TestMakeDBInstance(t *testing.T) {
	tests := []struct {
		name              string
		dsn               string
		expectedDialector string
	}{
		{
			name:              "WithSqliteURI",
			dsn:               "sqlite:///tmp/fasttrack.db",
			expectedDialector: "sqlite",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DB = nil
			db, err := MakeDBInstance(
				tt.dsn,
				time.Second*2,
				2,
				false,
				false,
				"s3://somewhere",
			)
			assert.Nil(t, err)
			assert.NotNil(t, db)
			assert.Equal(t, tt.expectedDialector, db.DB.Dialector.Name())

			// expecting the global 'DB' not to be set
			assert.Nil(t, DB)

			err = db.Close()
			assert.Nil(t, err)
		})
	}
}

func Test_newDbInstanceFactory(t *testing.T) {
	tests := []struct {
		name            string
		dsn             string
		expectedFactory dbFactory
	}{
		{
			name:            "WithSqliteURI",
			dsn:             "sqlite:///tmp/fasttrack.db",
			expectedFactory: sqliteDbFactory{},
		},
		{
			name:            "WithPostgresURI",
			dsn:             "postgres://pg:psw@db",
			expectedFactory: postgresDbFactory{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbFactory, err := newDbInstanceFactory(
				tt.dsn,
				time.Second*2,
				2,
				false,
			)
			assert.Nil(t, err)
			assert.NotNil(t, dbFactory)
			assert.IsType(t, tt.expectedFactory, dbFactory)
		})
	}
}
