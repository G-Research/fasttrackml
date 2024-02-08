package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateCmd(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "MigrationModuleIsGenerated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := RootCmd

			// Redirect command output to a byte buffer
			b := &bytes.Buffer{}
			cmd.SetOut(b)
			cmd.SetErr(b)

			// Create a temporary directories for the command to use
			databaseTmpDir := t.TempDir()
			migrationsTmpDir := t.TempDir()
			//nolint:gosec
			assert.Nil(t, os.Mkdir(filepath.Join(migrationsTmpDir, "v_0001"), 0o755))

			sourceModel := filepath.Join(databaseTmpDir, "model.go")
			originalModelContent := []byte("the original model.go bytes")
			//nolint:gosec
			assert.Nil(t, os.WriteFile(sourceModel, originalModelContent, 0o664))

			// Set the command flags as needed
			cmd.SetArgs([]string{"migrations", "create", "-d", databaseTmpDir, "-m", migrationsTmpDir})

			// Exec command
			if err := cmd.Execute(); err != nil {
				t.Errorf("CreateCmd() error = %v, output: %v", err, b.String())
			}

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
