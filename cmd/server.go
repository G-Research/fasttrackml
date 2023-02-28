package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/G-Resarch/fasttrack/api"
	"github.com/G-Resarch/fasttrack/database"
	"github.com/G-Resarch/fasttrack/ui"

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

	apiHandler := api.NewServeMux()
	apiHandler.HandleFunc("/artifacts/list", api.ArtifactList())
	apiHandler.HandleFunc("/experiments/create", api.ExperimentCreate())
	apiHandler.HandleFunc("/experiments/delete", api.ExperimentDelete())
	apiHandler.HandleFunc("/experiments/get", api.ExperimentGet())
	apiHandler.HandleFunc("/experiments/get-by-name", api.ExperimentGetByName())
	apiHandler.HandleFunc("/experiments/restore", api.ExperimentRestore())
	apiHandler.HandleFunc("/experiments/list", api.ExperimentSearch())
	apiHandler.HandleFunc("/experiments/search", api.ExperimentSearch())
	apiHandler.HandleFunc("/experiments/set-experiment-tag", api.ExperimentSetTag())
	apiHandler.HandleFunc("/experiments/update", api.ExperimentUpdate())
	apiHandler.HandleFunc("/metrics/get-history", api.MetricGetHistory())
	apiHandler.HandleFunc("/metrics/get-histories", api.MetricsGetHistories())
	apiHandler.HandleFunc("/runs/create", api.RunCreate())
	apiHandler.HandleFunc("/runs/delete", api.RunDelete())
	apiHandler.HandleFunc("/runs/delete-tag", api.RunDeleteTag())
	apiHandler.HandleFunc("/runs/get", api.RunGet())
	apiHandler.HandleFunc("/runs/log-batch", api.RunLogBatch())
	apiHandler.HandleFunc("/runs/log-metric", api.RunLogMetric())
	apiHandler.HandleFunc("/runs/log-parameter", api.RunLogParam())
	apiHandler.HandleFunc("/runs/restore", api.RunRestore())
	apiHandler.HandleFunc("/runs/search", api.RunSearch())
	apiHandler.HandleFunc("/runs/set-tag", api.RunSetTag())
	apiHandler.HandleFunc("/runs/update", api.RunUpdate())
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
		"/mlflow/api/2.0/mlflow/",
		"/mlflow/ajax-api/2.0/mlflow/",
		"/mlflow/api/2.0/preview/mlflow/",
		"/mlflow/ajax-api/2.0/preview/mlflow/",
	} {
		handler.Handle(path, api.BasicAuth(http.StripPrefix(strings.TrimRight(path, "/"), apiHandler)))
	}

	handler.HandleFunc("/mlflow/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK")
	})

	handler.HandleFunc("/mlflow/version", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, version)
	})

	handler.Handle("/mlflow/static-files/", http.StripPrefix("/mlflow/static-files/", http.FileServer(http.FS(ui.MlflowFS))))
	handler.HandleFunc("/mlflow/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mlflow/" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		f, _ := ui.MlflowFS.Open("index.html")
		defer f.Close()
		io.Copy(w, f)
	})

	handler.Handle("/", http.FileServer(http.FS(ui.ChooserFS)))

	server := &http.Server{
		Addr:    viper.GetString("listen-address"),
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
