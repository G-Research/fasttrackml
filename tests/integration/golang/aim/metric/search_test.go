//go:build integration

package run

import (
	"fmt"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type SearchMetricsTestSuite struct {
	suite.Suite
	client             *helpers.HttpClient
	runFixtures        *fixtures.RunFixtures
	experimentFixtures *fixtures.ExperimentFixtures
	runs               []*models.Run
}

func TestSearchMetricsTestSuite(t *testing.T) {
	suite.Run(t, new(SearchMetricsTestSuite))
}

func (s *SearchMetricsTestSuite) SetupTest() {
	s.client = helpers.NewAimApiClient(os.Getenv("SERVICE_BASE_URL"))
	/*
		runFixtures, err := fixtures.NewRunFixtures(os.Getenv("DATABASE_DSN"))
		assert.Nil(s.T(), err)
		s.runFixtures = runFixtures
		expFixtures, err := fixtures.NewExperimentFixtures(os.Getenv("DATABASE_DSN"))
		assert.Nil(s.T(), err)
		s.experimentFixtures = expFixtures

		experiment := &models.Experiment{
			Name:           uuid.New().String(),
			LifecycleStage: models.LifecycleStageActive,
		}
		_, err = s.experimentFixtures.CreateTestExperiment(context.Background(), experiment)
		assert.Nil(s.T(), err)
		_, err = s.runFixtures.CreateTestRuns(context.Background(), experiment, 1)
		assert.Nil(s.T(), err)
	*/
}

func (s *SearchMetricsTestSuite) Test_Ok() {
	tests := []struct {
		name  string
		query string
	}{
		{
			name:  "TestContainsFunction",
			query: `q=(run.name.contains("Run"))&p=500&report_progress=false`,
		},
		{
			name:  "TestStartWithFunction",
			query: `q=(run.name.startwith("Test"))&p=500&report_progress=false`,
		},
		{
			name:  "TestEndWithFunction",
			query: `q=(run.name.endwith("Run_1"))&p=500&report_progress=false`,
		},
	}
	for _, tt := range tests {
		resp := fiber.Map{}
		s.T().Run(tt.name, func(T *testing.T) {
			err := s.client.DoGetRequest(
				fmt.Sprintf("/runs/search/metric?%s", tt.query),
				&resp,
			)
			assert.Nil(s.T(), err)
			fmt.Println(resp)
		})
	}
}

func (s *SearchMetricsTestSuite) Test_Error() {}
