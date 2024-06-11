//go:build compatibility

package compatibility

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	mlflowRequest "github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	mlflowResponse "github.com/G-Research/fasttrackml/pkg/api/mlflow/api/response"
	"github.com/G-Research/fasttrackml/pkg/common/config"
	"github.com/G-Research/fasttrackml/pkg/database"
	"github.com/G-Research/fasttrackml/pkg/server"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type MLflowTestSuite struct {
	suite.Suite
	server              server.Server
	aimClient           func() *helpers.HttpClient
	mlflowClient        func() *helpers.HttpClient
	runsFixtures        *fixtures.RunFixtures
	metricsFixtures     *fixtures.MetricFixtures
	experimentsFixtures *fixtures.ExperimentFixtures
}

func TestMLflowCompatibilityTestSuite(t *testing.T) {
	suite.Run(t, new(MLflowTestSuite))
}

func (s *MLflowTestSuite) SetupSuite() {
	db, err := database.NewDBProvider(
		helpers.GetPostgresUri(),
		1*time.Second,
		20,
	)
	s.Nil(err)

	runsFixtures, err := fixtures.NewRunFixtures(db.GormDB())
	s.Require().Nil(err)
	s.runsFixtures = runsFixtures

	metricsFixtures, err := fixtures.NewMetricFixtures(db.GormDB())
	s.Require().Nil(err)
	s.metricsFixtures = metricsFixtures

	experimentFixtures, err := fixtures.NewExperimentFixtures(db.GormDB())
	s.Require().Nil(err)
	s.experimentsFixtures = experimentFixtures

	srv, err := server.NewServer(context.Background(), &config.Config{
		DatabaseURI:           db.Dsn(),
		DatabasePoolMax:       10,
		DatabaseSlowThreshold: 1 * time.Second,
		DatabaseMigrate:       true,
		DefaultArtifactRoot:   s.T().TempDir(),
		S3EndpointURI:         helpers.GetS3EndpointUri(),
		GSEndpointURI:         helpers.GetGSEndpointUri(),
	})
	s.Require().Nil(err)
	s.server = srv

	s.aimClient = func() *helpers.HttpClient {
		return helpers.NewAimApiClient(s.server)
	}
	s.mlflowClient = func() *helpers.HttpClient {
		return helpers.NewMlflowApiClient(s.server)
	}
}

func (s *MLflowTestSuite) Test_MLflow_Ok() {
	experiments, err := s.experimentsFixtures.GetExperiments(context.Background())
	s.Nil(err)
	s.Equal(2, len(experiments))
	for _, experiment := range experiments {
		// test few `experiment` endpoints.
		getExperimentResponse := mlflowResponse.GetExperimentResponse{}
		s.Require().Nil(
			s.mlflowClient().WithQuery(
				mlflowRequest.GetExperimentRequest{
					ID: fmt.Sprintf("%d", *experiment.ID),
				},
			).WithResponse(
				&getExperimentResponse,
			).DoRequest(
				"%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsGetRoute,
			),
		)
		s.Equal(fmt.Sprintf("%d", *experiment.ID), getExperimentResponse.Experiment.ID)
		s.Equal(experiment.Name, getExperimentResponse.Experiment.Name)
		s.Equal(string(experiment.LifecycleStage), getExperimentResponse.Experiment.LifecycleStage)
		s.Equal(experiment.ArtifactLocation, getExperimentResponse.Experiment.ArtifactLocation)
		s.Equal(experiment.CreationTime.Int64, getExperimentResponse.Experiment.CreationTime)
		s.Equal(experiment.LastUpdateTime.Int64, getExperimentResponse.Experiment.LastUpdateTime)
		s.Require().Equal(len(experiment.Tags), len(getExperimentResponse.Experiment.Tags))
		for i, tag := range experiment.Tags {
			s.Equal(tag.Key, getExperimentResponse.Experiment.Tags[i].Key)
			s.Equal(tag.Value, getExperimentResponse.Experiment.Tags[i].Value)
		}

		getExperimentResponse = mlflowResponse.GetExperimentResponse{}
		s.Require().Nil(
			s.mlflowClient().WithQuery(
				mlflowRequest.GetExperimentRequest{
					Name: experiment.Name,
				},
			).WithResponse(
				&getExperimentResponse,
			).DoRequest(
				"%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsGetByNameRoute,
			),
		)
		s.Equal(fmt.Sprintf("%d", *experiment.ID), getExperimentResponse.Experiment.ID)
		s.Equal(experiment.Name, getExperimentResponse.Experiment.Name)
		s.Equal(string(experiment.LifecycleStage), getExperimentResponse.Experiment.LifecycleStage)
		s.Equal(experiment.ArtifactLocation, getExperimentResponse.Experiment.ArtifactLocation)
		s.Equal(experiment.CreationTime.Int64, getExperimentResponse.Experiment.CreationTime)
		s.Equal(experiment.LastUpdateTime.Int64, getExperimentResponse.Experiment.LastUpdateTime)
		s.Require().Equal(len(experiment.Tags), len(getExperimentResponse.Experiment.Tags))
		for i, tag := range experiment.Tags {
			s.Equal(tag.Key, getExperimentResponse.Experiment.Tags[i].Key)
			s.Equal(tag.Value, getExperimentResponse.Experiment.Tags[i].Value)
		}

		// test few `run` endpoints.
		runs, err := s.runsFixtures.GetRuns(context.Background(), *experiment.ID)
		s.Nil(err)
		for _, run := range runs {
			getRunResponse := mlflowResponse.GetRunResponse{}
			s.Require().Nil(
				s.mlflowClient().WithQuery(
					mlflowRequest.GetRunRequest{
						RunID: run.ID,
					},
				).WithResponse(
					&getRunResponse,
				).DoRequest(
					"%s%s", mlflow.RunsRoutePrefix, mlflow.RunsGetRoute,
				),
			)
			s.Equal(run.ID, getRunResponse.Run.Info.ID)
			s.Equal(run.Name, getRunResponse.Run.Info.Name)
		}

		searchRunsResponse := &mlflowResponse.SearchRunsResponse{}
		s.Require().Nil(
			s.mlflowClient().WithMethod(
				http.MethodPost,
			).WithRequest(
				mlflowRequest.SearchRunsRequest{
					ViewType:      mlflowRequest.ViewTypeAll,
					ExperimentIDs: []string{fmt.Sprintf("%d", *experiment.ID)},
				},
			).WithResponse(
				&searchRunsResponse,
			).DoRequest(
				"%s%s", mlflow.RunsRoutePrefix, mlflow.RunsSearchRoute,
			),
		)
		s.Equal(len(runs), len(searchRunsResponse.Runs))

		// test few `metric` endpoints.
		runs, err = s.runsFixtures.GetRuns(context.Background(), *experiment.ID)
		s.Nil(err)
		for _, run := range runs {
			metrics, err := s.metricsFixtures.GetMetricsByRunID(context.Background(), run.ID)
			assert.Nil(s.T(), err)
			for _, metric := range metrics {
				getMetricHistoryResponse := mlflowResponse.GetMetricHistoryResponse{}
				s.Require().Nil(
					s.mlflowClient().WithQuery(
						mlflowRequest.GetMetricHistoryRequest{
							RunID:     run.ID,
							MetricKey: metric.Key,
						},
					).WithResponse(
						&getMetricHistoryResponse,
					).DoRequest(
						"%s%s", mlflow.MetricsRoutePrefix, mlflow.MetricsGetHistoryRoute,
					),
				)
				s.Equal(len(metrics), len(getMetricHistoryResponse.Metrics))
			}
		}
	}
}
