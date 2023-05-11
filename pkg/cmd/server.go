package cmd

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	aimAPI "github.com/G-Research/fasttrack/pkg/api/aim"
	mlflowAPI "github.com/G-Research/fasttrack/pkg/api/mlflow"
	mlflowService "github.com/G-Research/fasttrack/pkg/api/mlflow/service"
	"github.com/G-Research/fasttrack/pkg/database"
	aimUI "github.com/G-Research/fasttrack/pkg/ui/aim"
	"github.com/G-Research/fasttrack/pkg/ui/chooser"
	mlflowUI "github.com/G-Research/fasttrack/pkg/ui/mlflow"
	"github.com/G-Research/fasttrack/pkg/version"
)

var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts the tracking server",
	RunE:  serverCmd,
}

func serverCmd(cmd *cobra.Command, args []string) error {
	if err := database.ConnectDB(
		viper.GetString("database-uri"),
		viper.GetDuration("database-slow-threshold"),
		viper.GetInt("database-pool-max"),
		viper.GetBool("database-reset"),
		viper.GetBool("database-migrate"),
		viper.GetString("artifact-root"),
	); err != nil {
		database.DB.Close()
		return fmt.Errorf("error connecting to DB: %w", err)
	}
	defer database.DB.Close()

	server := fiber.New(fiber.Config{
		ReadBufferSize:        16384,
		ReadTimeout:           5 * time.Second,
		WriteTimeout:          600 * time.Second,
		IdleTimeout:           120 * time.Second,
		ServerHeader:          fmt.Sprintf("fasttrack/%s", version.Version),
		DisableStartupMessage: true,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			p := string(c.Request().URI().Path())
			switch {
			case strings.HasPrefix(p, "/aim/api/"):
				return aimAPI.ErrorHandler(c, err)

			case strings.HasPrefix(p, "/api/2.0/mlflow/") ||
				strings.HasPrefix(p, "/ajax-api/2.0/mlflow/") ||
				strings.HasPrefix(p, "/mlflow/ajax-api/2.0/mlflow/"):
				return mlflowService.ErrorHandler(c, err)

			default:
				return fiber.DefaultErrorHandler(c, err)
			}
		},
	})

	authUsername := viper.GetString("auth-username")
	authPassword := viper.GetString("auth-password")
	if authUsername != "" && authPassword != "" {
		log.Infof(`BasicAuth enabled with user "%s"`, authUsername)
		server.Use(basicauth.New(basicauth.Config{
			Users: map[string]string{
				authUsername: authPassword,
			},
		}))
	}

	server.Use(compress.New(compress.Config{
		Next: func(c *fiber.Ctx) bool {
			// This is a little bit brittle, maybe there is a better way?
			// Do not compress metric histories as urllib3 did not support file-like compressed reads until 2.0.0a1
			return strings.HasSuffix(c.Path(), "/metrics/get-histories")
		},
	}))

	server.Use(recover.New(recover.Config{EnableStackTrace: true}))

	server.Use(logger.New(logger.Config{
		Format: "${status} - ${latency} ${method} ${path}\n",
		Output: log.StandardLogger().Writer(),
	}))

	aimAPI.AddRoutes(server.Group("/aim/api/"))
	aimUI.AddRoutes(server.Group("/aim/"))

	mlflowAPI.AddRoutes(server.Group("/api/2.0/mlflow/"))
	mlflowAPI.AddRoutes(server.Group("/ajax-api/2.0/mlflow/"))
	mlflowAPI.AddRoutes(server.Group("/mlflow/ajax-api/2.0/mlflow/"))
	mlflowUI.AddRoutes(server.Group("/mlflow/"))

	server.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})
	server.Get("/version", func(c *fiber.Ctx) error {
		return c.SendString(version.Version)
	})

	chooser.AddRoutes(server.Group("/"))

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		log.Infof("Shutting down")
		if err := server.Shutdown(); err != nil {
			log.Infof("Error shutting down server: %v", err)
		}
		close(idleConnsClosed)
	}()

	addr := viper.GetString("listen-address")
	log.Infof("Listening on %s", addr)
	if err := server.Listen(addr); err != http.ErrServerClosed {
		return fmt.Errorf("error listening: %v", err)
	}

	<-idleConnsClosed

	return nil
}

func init() {
	RootCmd.AddCommand(ServerCmd)

	ServerCmd.Flags().StringP("listen-address", "a", "localhost:5000", "Address (host:post) to listen to")
	ServerCmd.Flags().String("artifact-root", "s3://fasttrack", "Artifact root")
	ServerCmd.Flags().String("auth-username", "", "BasicAuth username")
	ServerCmd.Flags().String("auth-password", "", "BasicAuth password")
	ServerCmd.Flags().StringP("database-uri", "d", "sqlite://fasttrack.db", "Database URI")
	ServerCmd.Flags().Int("database-pool-max", 20, "Maximum number of database connections in the pool")
	ServerCmd.Flags().Duration("database-slow-threshold", 1*time.Second, "Slow SQL warning threshold")
	ServerCmd.Flags().Bool("database-migrate", true, "Run database migrations")
	ServerCmd.Flags().Bool("database-reset", false, "Reinitialize database - WARNING all data will be lost!")
	ServerCmd.Flags().MarkHidden("database-reset")
	viper.BindEnv("auth-username", "MLFLOW_TRACKING_USERNAME")
	viper.BindEnv("auth-password", "MLFLOW_TRACKING_PASSWORD")
}
