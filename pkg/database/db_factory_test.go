package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMakeDBProvider(t *testing.T) {
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
			db, err := MakeDBProvider(
				tt.dsn,
				time.Second*2,
				2,
				false,
			)
			assert.Nil(t, err)
			assert.NotNil(t, db)
			assert.Equal(t, tt.expectedDialector, db.GormDB().Dialector.Name())

			// expecting the global 'DB' not to be set
			assert.Nil(t, DB)

			err = db.Close()
			assert.Nil(t, err)
		})
	}
}
