package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var NewMigrationCmd = &cobra.Command{
	Use:   "new-migration",
	Short: "Creates a blank migration at the next available number",
	Long: `The new-migration command will create a new, blank database
               migration using the next available migration number.`,
	RunE: newMigrationCmd,
}

func newMigrationCmd(cmd *cobra.Command, args []string) error {
	module, uniqueID, err := getNextModuleAndUniqueID()
	if err != nil {
		return err
	}
	fmt.Printf("next migration module is: %s\n", module)
	fmt.Printf("next uniqueID is: %s\n", uniqueID)

	// if err := createNewMigration(module, uniqueID); err != nil {
	// 	return err
	// }

	if err := rebuildMigrations(); err != nil {
		return err
	}

	return nil
}

func getNextModuleAndUniqueID() (module string, uniqueID string, err error) {
	// find next migration number
	files, err := os.ReadDir("pkg/database/migrations")
	if err != nil {
		return
	}

	max := ""
	for _, file := range files {
		if file.IsDir() && file.Name() > max {
			max = file.Name()
		}
	}

	maxModule := strings.Split(max, "_")[1]
	number, err := strconv.Atoi(maxModule)
	if err != nil {
		return
	}
	module = fmt.Sprintf("v_%04d", number+1)
	uniqueID = time.Now().Format("20060102150405")
	return
}

// nolint:errcheck,gosec
func init() {
	RootCmd.AddCommand(NewMigrationCmd)
}
