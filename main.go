package main

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fasttrack/api"
	"fasttrack/model"
	"flag"
	"fmt"
	"io"
	"io/fs"
	glog "log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
)

const version = "0.0.2"

//go:embed js/build
var content embed.FS

func main() {
	dsn := flag.String("db", "sqlite://fasttrack.db", "Database URL")
	addr := flag.String("listen", ":5000", "Address to listen to")
	levelString := flag.String("level", "info", "Log level")
	slowThreshold := flag.Duration("slow-threshold", 500*time.Millisecond, "Slow SQL warning threshold")
	init := flag.Bool("init", false, "(Re-)Initialize database - WARNING all data will be lost!")
	migrate := flag.Bool("migrate", true, "Run database migrations")
	artifactRoot := flag.String("artifact-root", "s3://fasttrack", "Artifact root")
	showVersion := flag.Bool("version", false, "Show version number")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

	logLevel, err := log.ParseLevel(*levelString)
	if err != nil {
		log.Fatalf("Unable to parse log level: %s", err)
	}
	log.SetLevel(logLevel)

	var sourceConn gorm.Dialector
	var replicaConn gorm.Dialector
	u, err := url.Parse(*dsn)
	if err != nil {
		log.Fatalf("Invalid database URL: %s", err)
	}
	switch u.Scheme {
	case "postgres":
		sourceConn = postgres.Open(u.String())
	case "sqlite":
		q := u.Query()
		q.Set("_case_sensitive_like", "true")
		q.Set("_mutex", "no")
		if q.Get("mode") != "memory" && !(q.Has("_journal") || q.Has("_journal_mode")) {
			q.Set("_journal", "WAL")
		}
		u.RawQuery = q.Encode()

		s, err := sql.Open(sqlite.DriverName, strings.Replace(u.String(), "sqlite://", "file:", 1))
		if err != nil {
			log.Fatalf("Failed to connect to database: %s", err)
		}
		defer s.Close()
		s.SetMaxIdleConns(1)
		s.SetMaxOpenConns(4)
		s.SetConnMaxIdleTime(0)
		s.SetConnMaxLifetime(0)
		sourceConn = sqlite.Dialector{
			Conn: s,
		}

		q.Set("_query_only", "true")
		u.RawQuery = q.Encode()
		r, err := sql.Open(sqlite.DriverName, strings.Replace(u.String(), "sqlite://", "file:", 1))
		if err != nil {
			log.Fatalf("Failed to connect to database: %s", err)
		}
		defer r.Close()
		replicaConn = sqlite.Dialector{
			Conn: r,
		}
	default:
		log.Fatalf("Unsupported database scheme %s", u.Scheme)
	}

	log.Infof("Using database %s", *dsn)

	dbLogLevel := logger.Warn
	if logLevel == log.DebugLevel {
		dbLogLevel = logger.Info
	}
	db, err := gorm.Open(sourceConn, &gorm.Config{
		Logger: logger.New(
			glog.New(
				log.StandardLogger().WriterLevel(log.WarnLevel),
				"",
				0,
			),
			logger.Config{
				SlowThreshold:             *slowThreshold,
				LogLevel:                  dbLogLevel,
				IgnoreRecordNotFoundError: true,
			},
		),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %s", err)
	}

	if replicaConn != nil {
		db.Use(
			dbresolver.Register(dbresolver.Config{
				Replicas: []gorm.Dialector{
					replicaConn,
				},
			}),
		)
	}

	if u.Scheme != "sqlite" {
		sqlDB, _ := db.DB()
		sqlDB.SetConnMaxIdleTime(time.Minute)
		sqlDB.SetMaxIdleConns(100)
		sqlDB.SetMaxOpenConns(100)
	}

	if *init {
		log.Info("Initializing database")
		db.Exec("drop schema public cascade")
		db.Exec("create schema public")
	}

	var schemaVersion model.AlembicVersion
	db.Session(&gorm.Session{
		Logger: logger.Discard,
	}).First(&schemaVersion)

	if schemaVersion.Version != "97727af70f4d" {
		if !*migrate {
			log.Fatalf("Unsupported database schema version %s", schemaVersion.Version)
		}

		switch schemaVersion.Version {
		case "":
			log.Info("Migrating database to 97727af70f4d")
			tx := db.Begin()
			if err = tx.AutoMigrate(
				&model.Experiment{},
				&model.ExperimentTag{},
				&model.Run{},
				&model.Param{},
				&model.Tag{},
				&model.Metric{},
				&model.LatestMetric{},
				&model.AlembicVersion{},
			); err != nil {
				log.Fatalf("Error migrating database to 97727af70f4d: %s", err)
			}
			tx.Create(&model.AlembicVersion{
				Version: "97727af70f4d",
			})
			tx.Commit()
			if tx.Error != nil {
				log.Fatalf("Error setting database schema version: %s", tx.Error)
			}

		case "c48cb773bb87":
			log.Info("Migrating database to bd07f7e963c5")
			tx := db.Begin()
			for _, table := range []interface{}{
				&model.Param{},
				&model.Metric{},
				&model.LatestMetric{},
				&model.Tag{},
			} {
				if err := tx.Migrator().CreateIndex(table, "RunID"); err != nil {
					log.Fatalf("Error migrating database to bd07f7e963c5: %s", err)
				}
			}
			tx.Model(&model.AlembicVersion{}).Where("1 = 1").Update("Version", "bd07f7e963c5")
			tx.Commit()
			if tx.Error != nil {
				log.Fatalf("Error setting database schema version to bd07f7e963c5: %s", err)
			}
			fallthrough

		case "bd07f7e963c5":
			log.Info("Migrating database to 0c779009ac13")
			tx := db.Begin()
			if err := tx.Migrator().AddColumn(&model.Run{}, "DeletedTime"); err != nil {
				log.Fatalf("Error migrating database to 0c779009ac13: %s", err)
			}
			tx.Model(&model.AlembicVersion{}).Where("1 = 1").Update("Version", "0c779009ac13")
			tx.Commit()
			if tx.Error != nil {
				log.Fatalf("Error setting database schema version to 0c779009ac13: %s", err)
			}
			fallthrough

		case "0c779009ac13":
			log.Info("Migrating database to cc1f77228345")
			tx := db.Begin()
			if err := tx.Migrator().AlterColumn(&model.Param{}, "value"); err != nil {
				log.Fatalf("Error migrating database to cc1f77228345: %s", err)
			}
			tx.Model(&model.AlembicVersion{}).Where("1 = 1").Update("Version", "cc1f77228345")
			tx.Commit()
			if tx.Error != nil {
				log.Fatalf("Error setting database schema version to cc1f77228345: %s", err)
			}
			fallthrough

		case "cc1f77228345":
			log.Info("Migrating database to 97727af70f4d")
			tx := db.Begin()
			for _, column := range []string{
				"CreationTime",
				"LastUpdateTime",
			} {
				if err := tx.Migrator().AddColumn(&model.Experiment{}, column); err != nil {
					log.Fatalf("Error migrating database to 97727af70f4d: %s", err)
				}
			}
			tx.Model(&model.AlembicVersion{}).Where("1 = 1").Update("Version", "97727af70f4d")
			tx.Commit()
			if tx.Error != nil {
				log.Fatalf("Error setting database schema version to 97727af70f4d: %s", err)
			}

		default:
			log.Fatalf("Unsupported database schema version %s", schemaVersion.Version)
		}
	}

	if tx := db.First(&model.Experiment{}, 0); tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			log.Info("Creating default experiment")
			var id int32 = 0
			ts := time.Now().UTC().UnixMilli()
			exp := model.Experiment{
				ID:             &id,
				Name:           "Default",
				LifecycleStage: model.LifecycleStageActive,
				CreationTime: sql.NullInt64{
					Int64: ts,
					Valid: true,
				},
				LastUpdateTime: sql.NullInt64{
					Int64: ts,
					Valid: true,
				},
			}
			if tx := db.Create(&exp); tx.Error != nil {
				log.Fatalf("Error creating default experiment: %s", tx.Error)
			}

			exp.ArtifactLocation = fmt.Sprintf("%s/%d", strings.TrimRight(*artifactRoot, "/"), *exp.ID)
			if tx := db.Model(&exp).Update("ArtifactLocation", exp.ArtifactLocation); tx.Error != nil {
				log.Fatalf("Error updating artifact_location for experiment '%s': %s", exp.Name, tx.Error)
			}
		} else {
			log.Fatalf("Unable to find default experiment: %s", tx.Error)
		}
	}

	apiHandler := api.NewServeMux()
	apiHandler.HandleFunc("/artifacts/list", api.ArtifactList(db))
	apiHandler.HandleFunc("/experiments/create", api.ExperimentCreate(db, *artifactRoot))
	apiHandler.HandleFunc("/experiments/delete", api.ExperimentDelete(db))
	apiHandler.HandleFunc("/experiments/get", api.ExperimentGet(db))
	apiHandler.HandleFunc("/experiments/get-by-name", api.ExperimentGetByName(db))
	apiHandler.HandleFunc("/experiments/restore", api.ExperimentRestore(db))
	apiHandler.HandleFunc("/experiments/list", api.ExperimentSearch(db))
	apiHandler.HandleFunc("/experiments/search", api.ExperimentSearch(db))
	apiHandler.HandleFunc("/experiments/set-experiment-tag", api.ExperimentSetTag(db))
	apiHandler.HandleFunc("/experiments/update", api.ExperimentUpdate(db))
	apiHandler.HandleFunc("/metrics/get-history", api.MetricGetHistory(db))
	apiHandler.HandleFunc("/runs/create", api.RunCreate(db))
	apiHandler.HandleFunc("/runs/delete", api.RunDelete(db))
	apiHandler.HandleFunc("/runs/delete-tag", api.RunDeleteTag(db))
	apiHandler.HandleFunc("/runs/get", api.RunGet(db))
	apiHandler.HandleFunc("/runs/log-batch", api.RunLogBatch(db))
	apiHandler.HandleFunc("/runs/log-metric", api.RunLogMetric(db))
	apiHandler.HandleFunc("/runs/log-parameter", api.RunLogParam(db))
	apiHandler.HandleFunc("/runs/restore", api.RunRestore(db))
	apiHandler.HandleFunc("/runs/search", api.RunSearch(db))
	apiHandler.HandleFunc("/runs/set-tag", api.RunSetTag(db))
	apiHandler.HandleFunc("/runs/update", api.RunUpdate(db))
	apiHandler.HandleFunc("/model-versions/search", func(w http.ResponseWriter, r *http.Request) any {
		return struct {
			ModelVersions []struct{} `json:"model_versions"`
		}{
			ModelVersions: make([]struct{}, 0),
		}
	})
	apiHandler.HandleFunc("/registered-models/search", func(w http.ResponseWriter, r *http.Request) any {
		return struct {
			RegisteredModels []struct{} `json:"registered_models"`
		}{
			RegisteredModels: make([]struct{}, 0),
		}
	})

	handler := http.NewServeMux()
	for _, path := range []string{
		"/api/2.0/mlflow/",
		"/ajax-api/2.0/mlflow/",
		"/api/2.0/preview/mlflow/",
		"/ajax-api/2.0/preview/mlflow/",
	} {
		handler.Handle(path, api.BasicAuth(http.StripPrefix(strings.TrimRight(path, "/"), apiHandler)))
	}

	handler.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	staticFiles, _ := fs.Sub(content, "js/build")
	handler.Handle("/static-files/", http.StripPrefix("/static-files/", http.FileServer(http.FS(staticFiles))))
	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		f, _ := staticFiles.Open("index.html")
		defer f.Close()
		io.Copy(w, f)
	})

	server := &http.Server{
		Addr:    *addr,
		Handler: handler,
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		log.Infof("Shutting down")
		if err := server.Shutdown(context.Background()); err != nil {
			log.Infof("Error shutting down server: %v", err)
		}
		close(idleConnsClosed)
	}()

	log.Infof("Listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Error listening: %v", err)
	}

	<-idleConnsClosed
}
