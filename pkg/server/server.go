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

	aimAPI "github.com/G-Research/fasttrackml/pkg/api/aim"
	aimController "github.com/G-Research/fasttrackml/pkg/api/aim/controller"
	aimRepositories "github.com/G-Research/fasttrackml/pkg/api/aim/dao/repositories"
	aimAppService "github.com/G-Research/fasttrackml/pkg/api/aim/services/app"
	aimDashboardService "github.com/G-Research/fasttrackml/pkg/api/aim/services/dashboard"
	aimExperimentService "github.com/G-Research/fasttrackml/pkg/api/aim/services/experiment"
	aimProjectService "github.com/G-Research/fasttrackml/pkg/api/aim/services/project"
	aimRunService "github.com/G-Research/fasttrackml/pkg/api/aim/services/run"
	aimTagService "github.com/G-Research/fasttrackml/pkg/api/aim/services/tag"
	mlflowAPI "github.com/G-Research/fasttrackml/pkg/api/mlflow"
	mlflowController "github.com/G-Research/fasttrackml/pkg/api/mlflow/controller"
	mlflowRepositories "github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/repositories"
	mlflowService "github.com/G-Research/fasttrackml/pkg/api/mlflow/services"
	mlflowArtifactService "github.com/G-Research/fasttrackml/pkg/api/mlflow/services/artifact"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/services/artifact/storage"
	mlflowExperimentService "github.com/G-Research/fasttrackml/pkg/api/mlflow/services/experiment"
	mlflowMetricService "github.com/G-Research/fasttrackml/pkg/api/mlflow/services/metric"
	mlflowModelService "github.com/G-Research/fasttrackml/pkg/api/mlflow/services/model"
	mlflowRunService "github.com/G-Research/fasttrackml/pkg/api/mlflow/services/run"
	"github.com/G-Research/fasttrackml/pkg/common/auth/oidc"
	"github.com/G-Research/fasttrackml/pkg/common/config"
	"github.com/G-Research/fasttrackml/pkg/common/dao"
	"github.com/G-Research/fasttrackml/pkg/common/dao/repositories"
	"github.com/G-Research/fasttrackml/pkg/common/middleware"
	"github.com/G-Research/fasttrackml/pkg/database"
	adminUI "github.com/G-Research/fasttrackml/pkg/ui/admin"
	adminUIController "github.com/G-Research/fasttrackml/pkg/ui/admin/controller"
	adminUINamespaceService "github.com/G-Research/fasttrackml/pkg/ui/admin/service/namespace"
	aimUI "github.com/G-Research/fasttrackml/pkg/ui/aim"
	"github.com/G-Research/fasttrackml/pkg/ui/chooser"
	chooserController "github.com/G-Research/fasttrackml/pkg/ui/chooser/controller"
	chooserNamespaceService "github.com/G-Research/fasttrackml/pkg/ui/chooser/service/namespace"
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

// NewServer creates a new server instance.
func NewServer(ctx context.Context, config *config.Config) (Server, error) {
	// create database provider.
	db, err := createDBProvider(ctx, config)
	if err != nil {
		return nil, err
	}

	// create artifact storage factory.
	artifactStorageFactory, err := storage.NewArtifactStorageFactory(config)
	if err != nil {
		return nil, eris.Wrap(err, "error creating artifact storage factory")
	}

	// create fiber app.
	//nolint:contextcheck
	app, err := createApp(ctx, config, db, artifactStorageFactory)
	if err != nil {
		return nil, eris.Wrapf(err, "error creating application")
	}

	return server{app}, nil
}

// createDBProvider creates a new DB provider.
func createDBProvider(ctx context.Context, config *config.Config) (database.DBProvider, error) {
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

	gormDBWithContext := db.GormDB().WithContext(ctx)
	if err := database.CheckAndMigrateDB(config.DatabaseMigrate, gormDBWithContext); err != nil {
		return nil, eris.Wrap(err, "error running database migration")
	}

	if err := database.CreateDefaultNamespace(gormDBWithContext); err != nil {
		return nil, eris.Wrap(err, "error creating default namespace")
	}

	if err := database.CreateDefaultExperiment(gormDBWithContext, config.DefaultArtifactRoot); err != nil {
		return nil, eris.Wrap(err, "error creating default experiment")
	}

	if err := database.CreateDefaultMetricContext(gormDBWithContext); err != nil {
		return nil, eris.Wrap(err, "error creating default context")
	}

	// cache a global reference to the gorm.DB
	database.DB = db.GormDB()
	return db, nil
}

// createApp creates a new fiber app with base configuration.
//
//nolint:contextcheck
func createApp(
	ctx context.Context,
	config *config.Config,
	db database.DBProvider,
	artifactStorageFactory storage.ArtifactStorageFactoryProvider,
) (*fiber.App, error) {
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
			case strings.HasPrefix(p, "/aim"):
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

	// create namespace notification listener.
	namespaceEventListener, err := dao.NewNamespaceListener(ctx, db.GormDB())
	if err != nil {
		return nil, eris.Wrap(err, "error creating namespace notification listener")
	}

	namespaceCachedRepository, err := mlflowRepositories.NewNamespaceCachedRepository(
		ctx, mlflowRepositories.NewNamespaceRepository(db.GormDB()), namespaceEventListener,
	)
	if err != nil {
		return nil, eris.Wrap(err, "error creating namespace repository")
	}
	rolesCachedRepository, err := repositories.NewRoleCachedRepository(
		ctx, db.GormDB(), namespaceEventListener,
	)
	if err != nil {
		return nil, eris.Wrap(err, "error creating roles repository")
	}

	namespaceEventListener.Listen()

	// attach global middlewares.
	if config.Auth.AuthUsername != "" && config.Auth.AuthPassword != "" {
		log.Info("Auth - enabling Basic Auth")
		app.Use(basicauth.New(basicauth.Config{
			Users: map[string]string{
				config.Auth.AuthUsername: config.Auth.AuthPassword,
			},
		}))
	}
	app.Use(middleware.NewNamespaceMiddleware(namespaceCachedRepository))

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

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})
	app.Get("/version", func(c *fiber.Ctx) error {
		return c.SendString(version.Version)
	})

	// based on Auth configuration, attach global OIDC or Basic Auth middleware.
	switch {
	case config.Auth.IsAuthTypeOIDC():
		oidcClient, err := oidc.NewClient(ctx, config)
		if err != nil {
			return nil, eris.Wrap(err, "error creating oidc client")
		}
		app.Get("/auth/oidc", func(ctx *fiber.Ctx) error {
			oauth2Token, err := oidcClient.Exchange(ctx.Context(), ctx.Query("code"))
			if err != nil {
				log.Errorf("error exchanging code to oauth2 token: %+v", err)
				return ctx.Redirect("/errors/internal-server", http.StatusMovedPermanently)
			}
			rawIDToken, ok := oauth2Token.Extra("id_token").(string)
			if !ok {
				log.Error("id_token is missing")
				return ctx.Redirect("/errors/internal-server", http.StatusMovedPermanently)
			}
			ctx.Cookie(&fiber.Cookie{
				Name:  "access_token",
				Value: rawIDToken,
			})
			ctx.Response().Header.Add("Cache-Control", "no-store")
			return ctx.Redirect("/", http.StatusMovedPermanently)
		})
		app.Get("/logout", func(ctx *fiber.Ctx) error {
			ctx.Cookie(&fiber.Cookie{
				Name:    "access_token",
				Expires: time.Now().Add(-5 * time.Second),
			})
			ctx.Response().Header.Add("Cache-Control", "no-store")
			return ctx.Redirect("/", http.StatusMovedPermanently)
		})
		app.Use(middleware.NewOIDCMiddleware(oidcClient, rolesCachedRepository))
	case config.Auth.IsAuthTypeUser():
		app.Use(middleware.NewBasicAuthMiddleware(config.Auth.AuthParsedUserPermissions))
	}

	app.Use(compress.New(compress.Config{
		Next: func(c *fiber.Ctx) bool {
			// This is a little brittle, maybe there is a better way?
			// Do not compress metric histories as urllib3 did not support file-like compressed reads until 2.0.0a1
			return strings.HasSuffix(c.Path(), "/metrics/get-histories")
		},
	}))

	// init `aim` api routes.
	aimAPI.NewRouter(
		aimController.NewController(
			aimTagService.NewService(
				aimRepositories.NewSharedTagRepository(db.GormDB()),
			),
			aimAppService.NewService(
				aimRepositories.NewAppRepository(db.GormDB()),
			),
			aimRunService.NewService(
				aimRepositories.NewRunRepository(db.GormDB()),
				aimRepositories.NewLogRepository(db.GormDB()),
				aimRepositories.NewMetricRepository(db.GormDB()),
				aimRepositories.NewTagRepository(db.GormDB()),
				aimRepositories.NewSharedTagRepository(db.GormDB()),
				aimRepositories.NewArtifactRepository(db.GormDB()),
			),
			aimProjectService.NewService(
				aimRepositories.NewTagRepository(db.GormDB()),
				aimRepositories.NewRunRepository(db.GormDB()),
				aimRepositories.NewParamRepository(db.GormDB()),
				aimRepositories.NewMetricRepository(db.GormDB()),
				aimRepositories.NewExperimentRepository(db.GormDB()),
				config.LiveUpdatesEnabled,
			),
			aimDashboardService.NewService(
				aimRepositories.NewDashboardRepository(db.GormDB()),
				aimRepositories.NewAppRepository(db.GormDB()),
			),
			aimExperimentService.NewService(
				aimRepositories.NewTagRepository(db.GormDB()),
				aimRepositories.NewExperimentRepository(db.GormDB()),
			),
		),
	).Init(app)

	// init `mlflow` api and ui routes.
	// TODO:refactoring right now it might look scary. we prettify it a bit later.
	mlflowAPI.NewRouter(
		mlflowController.NewController(
			mlflowRunService.NewService(
				mlflowRepositories.NewTagRepository(db.GormDB()),
				mlflowRepositories.NewRunRepository(db.GormDB()),
				mlflowRepositories.NewParamRepository(db.GormDB()),
				mlflowRepositories.NewMetricRepository(db.GormDB()),
				mlflowRepositories.NewExperimentRepository(db.GormDB()),
				mlflowRepositories.NewLogRepository(db.GormDB(), config.RunLogOutputMax),
				mlflowRepositories.NewArtifactRepository(db.GormDB()),
			),
			mlflowModelService.NewService(),
			mlflowMetricService.NewService(
				mlflowRepositories.NewRunRepository(db.GormDB()),
				mlflowRepositories.NewMetricRepository(db.GormDB()),
			),
			mlflowArtifactService.NewService(
				mlflowRepositories.NewRunRepository(db.GormDB()),
				artifactStorageFactory,
			),
			mlflowExperimentService.NewService(
				config,
				mlflowRepositories.NewTagRepository(db.GormDB()),
				mlflowRepositories.NewExperimentRepository(db.GormDB()),
			),
		),
	).Init(app)

	// run a log cleaner background job.
	mlflowRunService.NewLogCleaner(
		ctx,
		config,
		mlflowRepositories.NewLogRepository(db.GormDB(), config.RunLogOutputMax),
	).Run()

	mlflowUI.AddRoutes(app)
	aimUI.AddRoutes(app)

	// init `admin` UI routes.
	if err := adminUI.NewRouter(
		adminUIController.NewController(
			adminUINamespaceService.NewService(
				config,
				namespaceCachedRepository,
				mlflowRepositories.NewExperimentRepository(db.GormDB()),
			),
		),
	).Init(app); err != nil {
		return nil, eris.Wrap(err, "error initializing admin routes")
	}

	// init `chooser` ui routes.
	controller := chooserController.NewController(
		config,
		chooserNamespaceService.NewService(
			config,
			namespaceCachedRepository,
		),
	)
	if config.Auth.IsAuthTypeOIDC() {
		oidcClient, err := oidc.NewClient(ctx, config)
		if err != nil {
			return nil, eris.Wrap(err, "error creating oidc client")
		}
		controller.SetOIDCClient(oidcClient)
	}
	if err := chooser.NewRouter(controller).Init(app); err != nil {
		return nil, eris.Wrap(err, "error initializing chooser routes")
	}

	return app, nil
}
