package cmd

import (
	"github.com/spf13/cobra"

	"github.com/G-Research/fasttrackml/pkg/cmd/migrations"
)

var MigrationsCmd = &cobra.Command{
	Use:    "migrations",
	Short:  "Top-level command for migrations",
}

func init() {
	RootCmd.AddCommand(MigrationsCmd)
	MigrationsCmd.AddCommand(migrations.CreateCmd, migrations.RebuildCmd)
	MigrationsCmd.PersistentFlags().StringP(migrations.DatabaseSourcesFlag,
		"d", "./pkg/database", "Location for database package sources")
	MigrationsCmd.PersistentFlags().StringP(migrations.MigrationsSourcesFlag,
		"m", "./pkg/database/migrations", "Location for migration sources")

}
