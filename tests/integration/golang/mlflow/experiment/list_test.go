//go:build integration

package experiment

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hetiansu5/urlquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type ListExperimentsTestSuite struct {
	suite.Suite
	client   *helpers.HttpClient
	fixtures *fixtures.ExperimentFixtures
}

func TestListExperimentsTestSuite(t *testing.T) {
	suite.Run(t, new(ListExperimentsTestSuite))
}

func (s *ListExperimentsTestSuite) SetupTest() {
	s.client = helpers.NewHttpClient(os.Getenv("SERVICE_BASE_URL"))
	fixtures, err := fixtures.NewExperimentFixtures(os.Getenv("DATABASE_DSN"))
	assert.Nil(s.T(), err)
	s.fixtures = fixtures
}

func (s *ListExperimentsTestSuite) Test_Ok() {
	// 1. prepare database with test data.
	names := []string{"Test Experiment 1", "Test Experiment 2", "Test Experiment 3"}
	for _, name := range names {
		_, err := s.fixtures.CreateTestExperiment(context.Background(), &models.Experiment{
			Name: name,
			CreationTime: sql.NullInt64{
				Int64: time.Now().UTC().UnixMilli(),
				Valid: true,
			},
			LastUpdateTime: sql.NullInt64{
				Int64: time.Now().UTC().UnixMilli(),
				Valid: true,
			},
			LifecycleStage:   models.LifecycleStageActive,
			ArtifactLocation: "/artifact/location",
		})
		assert.Nil(s.T(), err)
	}
	defer func() {
		assert.Nil(s.T(), s.fixtures.UnloadFixtures())
	}()

	query, err := urlquery.Marshal(request.SearchExperimentsRequest{})
	assert.Nil(s.T(), err)

	resp := response.SearchExperimentsResponse{}
	err = s.client.DoGetRequest(
		fmt.Sprintf(
			"%s%s?%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsSearchRoute, query,
		),
		&resp,
	)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 4, len(resp.Experiments))
}

func (s *ListExperimentsTestSuite) Test_Error() {
}
