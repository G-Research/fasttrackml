package migrations

import (
	"bytes"
	"fmt"
	"go/format"
	"html/template"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rotisserie/eris"
	"github.com/spf13/cobra"
)

const (
	DatabaseSources   = "pkg/database"
	MigrationsSources = "pkg/database/migrations"
)

var newMigrateTemplate = `package {{ .module }}

import (
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/database/migrations"
)

const Version = "{{ .uniqueID }}"

func Migrate(db *gorm.DB) error {
	return migrations.RunWithoutForeignKeyIfNeeded(db, func() error {
		return db.Transaction(func(tx *gorm.DB) error {

                        // TODO add migration code as needed

			// Update the schema version
			return tx.Model(&SchemaVersion{}).
				Where("1 = 1").
				Update("Version", Version).
				Error
		})
	})
}
`

var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates a blank migration at the next available number",
	Long: `The create command generates a new database migration module using
               the next available migration number. The result is placed in the
               'pkg/database/migrations' folder`,
	RunE: newMigrationCmd,
}

func newMigrationCmd(cmd *cobra.Command, args []string) error {
	if err := createNewMigration(MigrationsSources, DatabaseSources); err != nil {
		return eris.Wrap(err, "error creating new migration")
	}

	if err := rebuildMigrations(MigrationsSources, DatabaseSources); err != nil {
		return eris.Wrap(err, "error rebuilding migrations")
	}

	return nil
}

func getNextModuleAndUniqueID(migrationsSources string) (string, string, error) {
	// find next migration number
	files, err := os.ReadDir(migrationsSources)
	if err != nil {
		return "", "", eris.Wrap(err, "error reading migration sources dir")
	}

	max := ""
	for _, file := range files {
		if file.IsDir() && file.Name() > max {
			max = file.Name()
		}
	}

	number := 0
	if len(max) > 0 {
		maxModule := strings.Split(max, "_")[1]
		number, err = strconv.Atoi(maxModule)
		if err != nil {
			return "", "",
				eris.Wrapf(err, "error parsing module name, should have pattern 'v_NNNN' but is '%s'", maxModule)
		}
	}
	module := fmt.Sprintf("v_%04d", number+1)
	uniqueID := time.Now().Format("20060102030405")
	return module, uniqueID, nil
}

func createNewMigration(migrationsSources, databaseSources string) error {
	module, uniqueID, err := getNextModuleAndUniqueID(migrationsSources)
	if err != nil {
		return eris.Wrap(err, "error finding next module name")
	}
	fmt.Printf("next migration module is: %s\n", module)
	fmt.Printf("next uniqueID is: %s\n", uniqueID)

	newModuleFolder := fmt.Sprintf("%s/%s",
		migrationsSources, module)
	//nolint:gosec
	if err := os.Mkdir(newModuleFolder, 0o755); err != nil {
		return eris.Wrap(err, "error creating director")
	}

	modelsBytes, err := os.ReadFile(fmt.Sprintf("%s/model.go",
		databaseSources))
	if err != nil {
		return eris.Wrap(err, "error reading the database/model.go file")
	}
	modelsBytes = []byte(strings.Replace(
		string(modelsBytes),
		"package database",
		fmt.Sprintf("package %s", module),
		1,
	))

	modelsFile := fmt.Sprintf("%s/model.go", newModuleFolder)
	//nolint:gosec
	if err := os.WriteFile(modelsFile, modelsBytes, 0o644); err != nil {
		return eris.Wrap(err, "error writing file")
	}

	tmpl, err := template.New("migrations").Parse(newMigrateTemplate)
	if err != nil {
		return eris.Wrap(err, "error parsing template")
	}
	data := map[string]any{
		"module":   module,
		"uniqueID": uniqueID,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return eris.Wrap(err, "error executing template")
	}

	src, err := format.Source(buf.Bytes())
	if err != nil {
		return eris.Wrap(err, "error formatting generated file")
	}
	newFile := fmt.Sprintf("%s/migrate.go", newModuleFolder)
	// nolint:gosec
	if err := os.WriteFile(newFile, src, 0o644); err != nil {
		return eris.Wrap(err, "error writing generated file")
	}
	return nil
}
