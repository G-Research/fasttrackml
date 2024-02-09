package migrations

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_createNewMigration(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "MigrationModuleIsGenerated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directories for the command to use
			databaseTmpDir := t.TempDir()
			migrationsTmpDir := t.TempDir()
			//nolint:gosec
			assert.Nil(t, os.Mkdir(filepath.Join(migrationsTmpDir, "v_0001"), 0o755))

			sourceModel := filepath.Join(databaseTmpDir, "model.go")
			originalModelContent := []byte("the original model.go bytes")
			//nolint:gosec
			assert.Nil(t, os.WriteFile(sourceModel, originalModelContent, 0o664))

			// Exec command
			require.Nil(t, createNewMigration(migrationsTmpDir, databaseTmpDir))

			// Verify
			//nolint:gosec
			bytes, err := os.ReadFile(filepath.Join(migrationsTmpDir, "v_0002", "model.go"))
			assert.Nil(t, err)
			assert.Equal(t, originalModelContent, bytes)

			//nolint:gosec
			bytes, err = os.ReadFile(filepath.Join(migrationsTmpDir, "v_0002", "migrate.go"))
			assert.Nil(t, err)
			assert.Contains(t, string(bytes), "package v_0002")
			assert.Contains(t, string(bytes), "Version = \""+time.Now().Format("20060102030405"))
		})
	}
}
