package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/G-Research/fasttrackml/pkg/version"
)

const (
	envPrefix = "FML"
)

var RootCmd = &cobra.Command{
	Use:   filepath.Base(os.Args[0]),
	Short: "Experiment tracking server focused on speed and scalability",
	Long: `FastTrackML is an experiment tracking server focused on speed and scalability.
It aims at being 100% compatible with the MLFlow client library and should be a
drop-in replacement.`,
	Version:           version.Version,
	PersistentPreRunE: initCmd,
	SilenceUsage:      true,
	SilenceErrors:     true,
}

func initCmd(cmd *cobra.Command, args []string) error {
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}

	level, err := log.ParseLevel(viper.GetString("log-level"))
	if err != nil {
		return fmt.Errorf(`invalid log level "%s"`, viper.GetString("log-level"))
	}
	log.SetLevel(level)
	if log.IsLevelEnabled(log.DebugLevel) {
		log.SetReportCaller(true)
	}
	return nil
}

func init() {
	RootCmd.PersistentFlags().StringP("log-level", "l", "info", "Log level")
	RootCmd.SetVersionTemplate("FastTrackML version {{.Version}}\n")

	viper.SetEnvPrefix(envPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
}
