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
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	aimAPI "github.com/G-Research/fasttrackml/pkg/api/aim"
	mlflowAPI "github.com/G-Research/fasttrackml/pkg/api/mlflow"
	mlflowConfig "github.com/G-Research/fasttrackml/pkg/api/mlflow/config"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/controller"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/repositories"
	mlflowService "github.com/G-Research/fasttrackml/pkg/api/mlflow/service"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/service/artifact"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/service/artifact/storage"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/service/experiment"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/service/metric"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/service/model"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/service/run"
	"github.com/G-Research/fasttrackml/pkg/database"
	aimUI "github.com/G-Research/fasttrackml/pkg/ui/aim"
	"github.com/G-Research/fasttrackml/pkg/ui/chooser"
	mlflowUI "github.com/G-Research/fasttrackml/pkg/ui/mlflow"
	"github.com/G-Research/fasttrackml/pkg/version"
)

var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts the tracking server",
	RunE:  serverCmd,
}

func serverCmd(cmd *cobra.Command, args []string) error {
	// 1. process config parameters.
	mlflowConfig := mlflowConfig.NewServiceConfig()
	if err := mlflowConfig.Validate(); err != nil {
		return err
	}

	// 2. init database connection.
	db, err := initDB(mlflowConfig)
	if err != nil {
		return err
	}
	defer db.Close()

	// 3. init main HTTP server.
	server := initServer(mlflowConfig)

	// 4. init `aim` api and ui routes.
	aimAPI.AddRoutes(server.Group("/aim/api/"))
	aimUI.AddRoutes(server.Group("/aim/"))

	storage, err := storage.NewArtifactStorage(mlflowConfig)
	if err != nil {
		return eris.Wrap(err, "error initializing artifact storage")
	}

	// 5. init `mlflow` api and ui routes.
	// TODO:DSuhinin right now it might look scary. we prettify it a bit later.
	mlflowAPI.NewRouter(
		controller.NewController(
			run.NewService(
				repositories.NewTagRepository(db.GormDB()),
				repositories.NewRunRepository(db.GormDB()),
				repositories.NewParamRepository(db.GormDB()),
				repositories.NewMetricRepository(db.GormDB()),
				repositories.NewExperimentRepository(db.GormDB()),
			),
			model.NewService(),
			metric.NewService(
				repositories.NewMetricRepository(db.GormDB()),
			),
			artifact.NewService(
				storage,
				repositories.NewRunRepository(db.GormDB()),
			),
			experiment.NewService(
				mlflowConfig,
				repositories.NewTagRepository(db.GormDB()),
				repositories.NewExperimentRepository(db.GormDB()),
			),
		),
	).Init(server)
	mlflowUI.AddRoutes(server.Group("/mlflow/"))
	// TODO:DSuhinin we have to move it.
	chooser.AddRoutes(server.Group("/"))

	isRunning := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		log.Infof("Shutting down")
		if err := server.Shutdown(); err != nil {
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

// initDB init DB connection.
func initDB(config *mlflowConfig.ServiceConfig) (database.DBProvider, error) {
	db, err := database.MakeDBProvider(
		config.DatabaseURI,
		config.DatabaseSlowThreshold,
		config.DatabasePoolMax,
		config.DatabaseReset,
		config.DatabaseMigrate,
		config.ArtifactRoot,
	)
	if err != nil {
		return nil, fmt.Errorf("error connecting to DB: %w", err)
	}
	// cache a global reference to the gorm.DB
	database.DB = db.GormDB()
	return db, nil
}

// initServer init HTTP server with base configuration.
func initServer(config *mlflowConfig.ServiceConfig) *fiber.App {
	server := fiber.New(fiber.Config{
		BodyLimit:             16 * 1024 * 1024,
		ReadBufferSize:        16384,
		ReadTimeout:           5 * time.Second,
		WriteTimeout:          600 * time.Second,
		IdleTimeout:           120 * time.Second,
		ServerHeader:          fmt.Sprintf("FastTrackML/%s", version.Version),
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

	if config.DevMode {
		log.Info("Development mode - enabling CORS")
		server.Use(cors.New())
	}

	if config.AuthUsername != "" && config.AuthPassword != "" {
		server.Use(basicauth.New(basicauth.Config{
			Users: map[string]string{
				config.AuthUsername: config.AuthPassword,
			},
		}))
	}

	server.Use(compress.New(compress.Config{
		Next: func(c *fiber.Ctx) bool {
			// This is a little brittle, maybe there is a better way?
			// Do not compress metric histories as urllib3 did not support file-like compressed reads until 2.0.0a1
			return strings.HasSuffix(c.Path(), "/metrics/get-histories")
		},
	}))

	server.Use(recover.New(recover.Config{EnableStackTrace: true}))
	server.Use(logger.New(logger.Config{
		Format: "${status} - ${latency} ${method} ${path}\n",
		Output: log.StandardLogger().Writer(),
	}))

	server.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})
	server.Get("/version", func(c *fiber.Ctx) error {
		return c.SendString(version.Version)
	})

	return server
}

func init() {
	RootCmd.AddCommand(ServerCmd)

	ServerCmd.Flags().StringP("listen-address", "a", "localhost:5000", "Address (host:post) to listen to")
	ServerCmd.Flags().String("artifact-root", "./artifacts", "Artifact root")
	ServerCmd.Flags().String("s3-endpoint-uri", "", "S3 compatible storage base endpoint url")
	ServerCmd.Flags().String("auth-username", "", "BasicAuth username")
	ServerCmd.Flags().String("auth-password", "", "BasicAuth password")
	ServerCmd.Flags().StringP("database-uri", "d", "sqlite://fasttrackml.db", "Database URI")
	ServerCmd.Flags().Int("database-pool-max", 20, "Maximum number of database connections in the pool")
	ServerCmd.Flags().Duration("database-slow-threshold", 1*time.Second, "Slow SQL warning threshold")
	ServerCmd.Flags().Bool("database-migrate", true, "Run database migrations")
	ServerCmd.Flags().Bool("database-reset", false, "Reinitialize database - WARNING all data will be lost!")
	ServerCmd.Flags().MarkHidden("database-reset")
	ServerCmd.Flags().Bool("dev-mode", false, "Development mode - enable CORS")
	ServerCmd.Flags().MarkHidden("dev-mode")
	viper.BindEnv("auth-username", "MLFLOW_TRACKING_USERNAME")
	viper.BindEnv("auth-password", "MLFLOW_TRACKING_PASSWORD")
}
