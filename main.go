package main

import (
	"database/sql"
	"errors"
	"fasttrack/api"
	"fasttrack/model"
	"flag"
	"fmt"
	glog "log"
	"net/http"
	"net/url"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	dsn := flag.String("db", "sqlite://:memory:?cache=shared", "Database URL")
	addr := flag.String("listen", ":5000", "Address to listen to")
	levelString := flag.String("level", "info", "Log level")
	migrate := flag.Bool("migrate", true, "Run database migrations")
	artifactRoot := flag.String("artifact-root", "s3://fasttrack", "Artifact root")
	flag.Parse()

	level, err := log.ParseLevel(*levelString)
	if err != nil {
		log.Fatalf("Unable to parse log level: %s", err)
	}
	log.SetLevel(level)

	var dialector gorm.Dialector
	u, err := url.Parse(*dsn)
	if err != nil {
		log.Fatalf("Invalid database URL: %s", err)
	}
	switch u.Scheme {
	case "postgres":
		dialector = postgres.Open(u.String())
	case "sqlite":
		q := u.Query()
		q.Set("_case_sensitive_like", "true")
		dialector = sqlite.Open(u.Host + u.Path + "?" + q.Encode())
	}
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.New(
			glog.New(
				log.StandardLogger().WriterLevel(log.WarnLevel),
				"\r\n",
				glog.LstdFlags,
			),
			logger.Config{
				SlowThreshold:             200 * time.Millisecond,
				LogLevel:                  logger.Warn,
				IgnoreRecordNotFoundError: true,
			},
		),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %s", err)
	}

	if *migrate {
		log.Info("Migrating DB")
		if err = db.AutoMigrate(
			&model.Experiment{},
			&model.ExperimentTag{},
			&model.Run{},
			&model.Param{},
			&model.Tag{},
			&model.Metric{},
			&model.LatestMetric{},
		); err != nil {
			log.Fatalf("Error migrating database: %s", err)
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
	}

	apiHandler := api.NewServeMux()
	apiHandler.HandleFunc("/artifacts/list", api.ArtifactList(db))
	apiHandler.HandleFunc("/experiments/create", api.ExperimentCreate(db, *artifactRoot))
	apiHandler.HandleFunc("/experiments/delete", api.ExperimentDelete(db))
	apiHandler.HandleFunc("/experiments/get", api.ExperimentGet(db))
	apiHandler.HandleFunc("/experiments/get-by-name", api.ExperimentGetByName(db))
	apiHandler.HandleFunc("/experiments/restore", api.ExperimentRestore(db))
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

	handler := http.NewServeMux()
	for _, path := range []string{
		"/api/2.0/mlflow/",
		"/ajax-api/2.0/mlflow/",
		"/api/2.0/preview/mlflow/",
		"/ajax-api/2.0/preview/mlflow/",
	} {
		handler.Handle(path, http.StripPrefix(strings.TrimRight(path, "/"), apiHandler))
	}

	server := &http.Server{
		Addr:    *addr,
		Handler: handler,
	}

	log.Infof("Listening on %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
