package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRebuildCmd(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "GeneratedMigrationsFileIsCreated",
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
			//nolint:gosec
			assert.Nil(t, os.Mkdir(filepath.Join(migrationsTmpDir, "v_0002"), 0o755))

			// Set the command flags as needed
			cmd.SetArgs([]string{"migrations", "rebuild", "-d", databaseTmpDir, "-m", migrationsTmpDir})

			// Exec command
			if err := cmd.Execute(); err != nil {
				t.Errorf("NewMigrationCmd() error = %v, output: %v", err, b.String())
			}

			// Verify
			//nolint:gosec
			bytes, err := os.ReadFile(filepath.Join(databaseTmpDir, "migrate_generated.go"))
			assert.Nil(t, err)
			assert.Contains(t, string(bytes), "return v_0002.Version")
			assert.Contains(t, string(bytes), "case \"\":")
			assert.Contains(t, string(bytes), "case v_0001.Version:")
			assert.NotContains(t, string(bytes), "case v_0002.Version:")
		})
	}
}
