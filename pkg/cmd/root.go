package cmd

import (
	"fmt"
	"strings"

	"github.com/G-Research/fasttrackml/pkg/version"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	envPrefix = "FML"
)

var RootCmd = &cobra.Command{
	Use:   "fml",
	Short: "A fast experiment tracking server compatible with MLFlow",
	Long: `FastTrackML is a rewrite of the MLFlow tracking server with a focus on scalability.
It aims at being 100% compatible with the MLFlow client library and should be a
drop-in replacement. It can even use existing SQLite/SQLCipher/PostgreSQL
databases created by MLFlow 1.21+.`,
	Version:           version.Version,
	PersistentPreRunE: initCmd,
	SilenceUsage:      true,
	SilenceErrors:     true,
}

func initCmd(cmd *cobra.Command, args []string) error {
	viper.BindPFlags(cmd.Flags())

	level, err := log.ParseLevel(viper.GetString("log-level"))
	if err != nil {
		return fmt.Errorf(`invalid log level "%s"`, viper.GetString("log-level"))
	}
	log.SetLevel(level)
	return nil
}

func init() {
	RootCmd.PersistentFlags().StringP("log-level", "l", "info", "Log level")

	viper.SetEnvPrefix(envPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
}
