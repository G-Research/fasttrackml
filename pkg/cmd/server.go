package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/G-Research/fasttrackml/pkg/common/config"
	"github.com/G-Research/fasttrackml/pkg/server"
)

var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts the tracking server",
	RunE:  serverCmd,
}

func serverCmd(cmd *cobra.Command, args []string) error {
	// process config parameters.
	mlflowConfig := config.NewConfig()
	if err := mlflowConfig.Validate(); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	server, err := server.NewServer(ctx, mlflowConfig)
	if err != nil {
		return err
	}

	isRunning := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		log.Infof("Shutting down")
		if err := server.ShutdownWithTimeout(1 * time.Minute); err != nil {
			log.Infof("Error shutting down server: %v", err)
		}
		close(isRunning)
	}()

	log.Infof("Listening on %s", mlflowConfig.ListenAddress)
	if err := server.Listen(mlflowConfig.ListenAddress); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("error listening: %v", err)
	}

	<-isRunning

	return nil
}

// nolint:errcheck,gosec
func init() {
	RootCmd.AddCommand(ServerCmd)

	ServerCmd.Flags().StringP("listen-address", "a", "localhost:5000", "Address (host:post) to listen to")
	ServerCmd.Flags().String("default-artifact-root", "./artifacts", "Default artifact root")
	ServerCmd.Flags().String("s3-endpoint-uri", "", "S3 compatible storage base endpoint url")
	ServerCmd.Flags().String("gs-endpoint-uri", "", "Google Storage base endpoint url")
	ServerCmd.Flags().MarkHidden("gs-endpoint-uri")
	ServerCmd.Flags().String("auth-username", "", "BasicAuth username")
	ServerCmd.Flags().String("auth-password", "", "BasicAuth password")
	ServerCmd.Flags().String("auth-users-config", "", "Users configuration file")
	ServerCmd.Flags().String("auth-oidc-client-id", "", "OIDC auth client id")
	ServerCmd.Flags().String("auth-oidc-client-secret", "", "OIDC auth client secret")
	ServerCmd.Flags().String("auth-oidc-provider-endpoint", "", "OIDC auth provider endpoint")
	ServerCmd.Flags().String("auth-oidc-scopes", "", "OIDC requested scopes")
	ServerCmd.Flags().String("auth-oidc-admin-role", "", "OIDC admin role identifier")
	ServerCmd.Flags().String("auth-oidc-claim-roles", "", "OIDC claim to inspect for roles")
	ServerCmd.Flags().StringP("database-uri", "d", "sqlite://fasttrackml.db", "Database URI")
	ServerCmd.Flags().Int("database-pool-max", 20, "Maximum number of database connections in the pool")
	ServerCmd.Flags().Duration("database-slow-threshold", 1*time.Second, "Slow SQL warning threshold")
	ServerCmd.Flags().Bool("database-migrate", true, "Run database migrations")
	ServerCmd.Flags().Bool("database-reset", false, "Reinitialize database - WARNING all data will be lost!")
	ServerCmd.Flags().Bool("live-updates-enabled", false, "Enable 'live updates' in the Aim UI")
	ServerCmd.Flags().MarkHidden("database-reset")
	ServerCmd.Flags().Bool("dev-mode", false, "Development mode - enable CORS")
	ServerCmd.Flags().MarkHidden("dev-mode")
	ServerCmd.Flags().Int("log-output-max", 2000, "Maximum log rows per run to retain.")
	ServerCmd.Flags().Duration("log-output-retention", 7*24*time.Hour, "Run logs retention period")
	viper.BindEnv("auth-username", "MLFLOW_TRACKING_USERNAME")
	viper.BindEnv("auth-password", "MLFLOW_TRACKING_PASSWORD")
}
