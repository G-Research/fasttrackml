package cmd

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/G-Resarch/fasttrack/api/mlflow"
	"github.com/G-Resarch/fasttrack/database"
	"github.com/G-Resarch/fasttrack/ui"
	"github.com/G-Resarch/fasttrack/version"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts the tracking server",
	RunE:  serverCmd,
}

func serverCmd(cmd *cobra.Command, args []string) error {
	database.ConnectDB(
		viper.GetString("database-uri"),
		viper.GetDuration("database-slow-threshold"),
		viper.GetInt("database-pool-max"),
		viper.GetBool("database-init"),
		viper.GetBool("database-migrate"),
		viper.GetString("artifact-root"),
	)
	defer database.DB.Close()

	server := fiber.New(fiber.Config{
		ServerHeader:          fmt.Sprintf("fasttrack/%s", version.Version),
		DisableStartupMessage: true,
	})

	server.Mount("/mlflow/", mlflow.NewApp(viper.GetString("auth-username"), viper.GetString("auth-password")))

	// Somehow mounting/using ChooserFS as a filesystem handler _sometimes_ results in 404 status code for /mlflow/
	// This is working around it in a non-intellectually-satisfying but effective way
	server.Get("/", etag.New(), func(c *fiber.Ctx) error {
		file, _ := ui.ChooserFS.Open("index.html")
		stat, _ := file.Stat()
		c.Set("Content-Type", "text/html; charset=utf-8")
		c.Response().SetBodyStream(file, int(stat.Size()))
		return nil
	})
	server.Get("/simple.min.css", etag.New(), func(c *fiber.Ctx) error {
		file, _ := ui.ChooserFS.Open("simple.min.css")
		stat, _ := file.Stat()
		c.Set("Content-Type", "text/css")
		c.Response().SetBodyStream(file, int(stat.Size()))
		return nil
	})

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
	ServerCmd.Flags().Bool("database-init", false, "(Re-)Initialize database - WARNING all data will be lost!")
	ServerCmd.Flags().MarkHidden("database-init")
	viper.BindEnv("auth-username", "MLFLOW_TRACKING_USERNAME")
	viper.BindEnv("auth-password", "MLFLOW_TRACKING_PASSWORD")
}
