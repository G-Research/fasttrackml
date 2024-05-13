package compatibility

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	aimRequest "github.com/G-Research/fasttrackml/pkg/api/aim/api/request"
	aimResponse "github.com/G-Research/fasttrackml/pkg/api/aim/api/response"
	"github.com/G-Research/fasttrackml/pkg/common/config"
	"github.com/G-Research/fasttrackml/pkg/database"
	"github.com/G-Research/fasttrackml/pkg/server"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type AimTestSuite struct {
	suite.Suite
	server              server.Server
	aimClient           func() *helpers.HttpClient
	mlflowClient        func() *helpers.HttpClient
	runsFixtures        *fixtures.RunFixtures
	metricsFixtures     *fixtures.MetricFixtures
	experimentsFixtures *fixtures.ExperimentFixtures
}

func TestAIMCompatibilityTestSuite(t *testing.T) {
	suite.Run(t, new(AimTestSuite))
}

func (s *AimTestSuite) SetupSuite() {
	db, err := database.NewDBProvider(
		helpers.GetPostgresUri(),
		1*time.Second,
		20,
	)

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
	s.Nil(err)
	s.server = srv

	s.aimClient = func() *helpers.HttpClient {
		return helpers.NewAimApiClient(s.server)
	}
	s.mlflowClient = func() *helpers.HttpClient {
		return helpers.NewMlflowApiClient(s.server)
	}
}

func (s *MLflowTestSuite) Test_Aim_Ok() {
	experiments, err := s.experimentsFixtures.GetExperiments(context.Background())
	s.Nil(err)
	s.Equal(2, len(experiments))
	for _, experiment := range experiments {
		var resp aimResponse.ExperimentRuns
		s.Require().Nil(
			s.aimClient().WithResponse(
				&resp,
			).DoRequest(
				"/experiments/%d/runs", *experiment.ID,
			),
		)
		s.Equal(fmt.Sprintf("%d", *experiment.ID), resp.ID)

		for _, run := range resp.Runs {
			existingRun, err := s.runsFixtures.GetRun(context.Background(), run.ID)
			s.Nil(err)
			s.Equal(run.ID, existingRun.ID)
			s.Equal(run.Name, existingRun.Name)

			var resp []aimResponse.GetRunMetricsResponse
			s.Require().Nil(
				s.aimClient().WithMethod(
					http.MethodPost,
				).WithRequest(
					aimRequest.GetRunMetricsRequest{
						{
							Name: run.Name,
						},
					},
				).WithResponse(
					&resp,
				).DoRequest(
					"/runs/%s/metric/get-batch", run.ID,
				),
			)
			s.Equal(1, len(resp))
		}
	}
}
