package main

import (
	"fasttrack/api"
	"flag"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := flag.String("db", "host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable", "Postgres DSN")
	addr := flag.String("listen", ":5000", "Address to listen to")
	levelString := flag.String("level", "info", "Log level")
	flag.Parse()

	level, err := log.ParseLevel(*levelString)
	if err != nil {
		log.Fatal("unable to parse log level", err)
	}
	log.SetLevel(level)

	db, err := gorm.Open(postgres.Open(*dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to db", err)
	}

	// log.Info("Migrating DB")
	// if err = db.AutoMigrate(
	// 	&model.Experiment{},
	// 	&model.ExperimentTag{},
	// 	&model.Run{},
	// 	&model.Param{},
	// 	&model.Tag{},
	// 	&model.Metric{},
	// 	&model.LatestMetric{},
	// ); err != nil {
	// 	log.Fatal("error migrating db", err)
	// }

	apiHandler := api.NewServeMux()
	apiHandler.HandleFunc("/artifacts/list", api.ArtifactList(db))
	apiHandler.HandleFunc("/experiments/create", api.ExperimentCreate(db))
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
