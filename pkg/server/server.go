package server

import (
	"context"
	"fmt"
	"net/http"
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

	"github.com/G-Research/fasttrackml/pkg/api/admin/service/namespace"
	aimAPI "github.com/G-Research/fasttrackml/pkg/api/aim"
	mlflowAPI "github.com/G-Research/fasttrackml/pkg/api/mlflow"
	mlflowConfig "github.com/G-Research/fasttrackml/pkg/api/mlflow/config"
	mlflowController "github.com/G-Research/fasttrackml/pkg/api/mlflow/controller"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/repositories"
	mlflowRepositories "github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/repositories"
	mlflowService "github.com/G-Research/fasttrackml/pkg/api/mlflow/service"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/service/artifact"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/service/artifact/storage"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/service/experiment"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/service/metric"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/service/model"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/service/run"
	namespaceMiddleware "github.com/G-Research/fasttrackml/pkg/common/middleware/namespace"
	"github.com/G-Research/fasttrackml/pkg/database"
	adminUI "github.com/G-Research/fasttrackml/pkg/ui/admin"
	adminUIController "github.com/G-Research/fasttrackml/pkg/ui/admin/controller"
	aimUI "github.com/G-Research/fasttrackml/pkg/ui/aim"
	"github.com/G-Research/fasttrackml/pkg/ui/chooser"
	chooserController "github.com/G-Research/fasttrackml/pkg/ui/chooser/controller"
	mlflowUI "github.com/G-Research/fasttrackml/pkg/ui/mlflow"
	"github.com/G-Research/fasttrackml/pkg/version"
)

type Server interface {
	Listen(address string) error
	ShutdownWithTimeout(timeout time.Duration) error
	Test(req *http.Request, msTimeout ...int) (*http.Response, error)
}

type server struct {
	*fiber.App
}

func NewServer(ctx context.Context, config *mlflowConfig.ServiceConfig) (Server, error) {
	// init database connection.
	db, err := initDB(config)
	if err != nil {
		return nil, err
	}

	var namespaceRepository repositories.NamespaceRepositoryProvider
	var artifactStorageFactory storage.ArtifactStorageFactoryProvider
	if err := func() error {
		// create namespace notification listener.
		namespaceListener, err := dao.NewNamespaceListener(ctx, db.GormDB())
		if err != nil {
			return eris.Wrap(err, "error creating namespace notification listener")
		}

		// create cached namespace repository.
		namespaceRepository, err = repositories.NewNamespaceCachedRepository(
			db.GormDB(), namespaceListener, repositories.NewNamespaceRepository(db.GormDB()),
		)
		if err != nil {
			return eris.Wrap(err, "error creating cached namespace repository")
		}

		// create artifact storage factory.
		artifactStorageFactory, err = storage.NewArtifactStorageFactory(config)
		if err != nil {
			return eris.Wrap(err, "error creating artifact storage factory")
		}

		return nil
	}(); err != nil {
		//nolint:errcheck,gosec
		db.Close()
		return nil, err
	}

	// init main HTTP server.
	//nolint:contextcheck
	server := initServer(config, db, artifactStorageFactory, namespaceRepository)

	return server, nil
}

// initDB init DB connection.
func initDB(config *mlflowConfig.ServiceConfig) (database.DBProvider, error) {
	db, err := database.NewDBProvider(
		config.DatabaseURI,
		config.DatabaseSlowThreshold,
		config.DatabasePoolMax,
	)
	if err != nil {
		return nil, fmt.Errorf("error connecting to DB: %w", err)
	}

	if config.DatabaseReset {
		if err := db.Reset(); err != nil {
			return nil, eris.Wrap(err, "error resetting database")
		}
	}

	if err := database.CheckAndMigrateDB(config.DatabaseMigrate, db.GormDB()); err != nil {
		return nil, eris.Wrap(err, "error running database migration")
	}

	if err := database.CreateDefaultNamespace(db.GormDB()); err != nil {
		return nil, eris.Wrap(err, "error creating default namespace")
	}

	if err := database.CreateDefaultExperiment(db.GormDB(), config.DefaultArtifactRoot); err != nil {
		return nil, eris.Wrap(err, "error creating default experiment")
	}

	// cache a global reference to the gorm.DB
	database.DB = db.GormDB()
	return db, nil
}

// initServer init HTTP server with base configuration.
func initServer(
	config *mlflowConfig.ServiceConfig,
	db database.DBProvider,
	artifactStorageFactory storage.ArtifactStorageFactoryProvider,
	namespaceRepository repositories.NamespaceRepositoryProvider,
) Server {
	app := fiber.New(fiber.Config{
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

	app.Hooks().OnShutdown(func() error {
		log.Info("Shutting down database connection")
		return db.Close()
	})

	if config.DevMode {
		log.Info("Development mode - enabling CORS")
		app.Use(cors.New())
	}

	if config.AuthUsername != "" && config.AuthPassword != "" {
		app.Use(basicauth.New(basicauth.Config{
			Users: map[string]string{
				config.AuthUsername: config.AuthPassword,
			},
		}))
	}

	app.Use(compress.New(compress.Config{
		Next: func(c *fiber.Ctx) bool {
			// This is a little brittle, maybe there is a better way?
			// Do not compress metric histories as urllib3 did not support file-like compressed reads until 2.0.0a1
			return strings.HasSuffix(c.Path(), "/metrics/get-histories")
		},
	}))

	app.Use(recover.New(recover.Config{EnableStackTrace: true}))
	app.Use(logger.New(logger.Config{
		Format: "${status} - ${latency} ${method} ${path}\n",
		Output: log.StandardLogger().Writer(),
	}))

	app.Use(namespaceMiddleware.New(namespaceRepository))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})
	app.Get("/version", func(c *fiber.Ctx) error {
		return c.SendString(version.Version)
	})

	// init `aim` api and ui routes.
	router := app.Group("/aim/api/")
	aimAPI.AddRoutes(router)
	aimUI.AddRoutes(app)

	// init `mlflow` api and ui routes.
	// TODO:DSuhinin right now it might look scary. we prettify it a bit later.
	mlflowAPI.NewRouter(
		mlflowController.NewController(
			run.NewService(
				mlflowRepositories.NewTagRepository(db.GormDB()),
				mlflowRepositories.NewRunRepository(db.GormDB()),
				mlflowRepositories.NewParamRepository(db.GormDB()),
				mlflowRepositories.NewMetricRepository(db.GormDB()),
				mlflowRepositories.NewExperimentRepository(db.GormDB()),
			),
			model.NewService(),
			metric.NewService(
				mlflowRepositories.NewRunRepository(db.GormDB()),
				mlflowRepositories.NewMetricRepository(db.GormDB()),
			),
			artifact.NewService(
				mlflowRepositories.NewRunRepository(db.GormDB()),
				artifactStorageFactory,
			),
			experiment.NewService(
				config,
				mlflowRepositories.NewTagRepository(db.GormDB()),
				mlflowRepositories.NewExperimentRepository(db.GormDB()),
			),
		),
	).Init(app)
	mlflowUI.AddRoutes(app)

	// init `admin` UI routes.
	adminUI.NewRouter(
		adminUIController.NewController(
			namespace.NewService(namespaceRepository),
		),
	).Init(app)

	// init `chooser` ui routes.
	chooser.NewRouter(
		chooserController.NewController(
			namespace.NewService(namespaceRepository),
		),
	).AddRoutes(app)

	return &server{app}
}
