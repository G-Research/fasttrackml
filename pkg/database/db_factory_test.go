package database

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMakeDBProvider(t *testing.T) {
	tests := []struct {
		name              string
		dsn               string
		expectedDialector string
	}{
		{
			name:              "WithSqliteURI",
			dsn:               "sqlite://" + filepath.Join(t.TempDir(), "fasttrackml.db"),
			expectedDialector: SQLiteDialectorName,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DB = nil
			db, err := NewDBProvider(
				tt.dsn,
				time.Second*2,
				2,
			)
			require.Nil(t, err)
			assert.NotNil(t, db)
			assert.Equal(t, tt.expectedDialector, db.GormDB().Dialector.Name())

			// expecting the global 'DB' not to be set
			assert.Nil(t, DB)
			require.Nil(t, db.Close())
		})
	}
}
