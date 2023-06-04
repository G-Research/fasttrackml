//go:build integration

package experiment

/*
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
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type DeleteExperimentTestSuite struct {
	suite.Suite
	client   *helpers.HttpClient
	fixtures *fixtures.ExperimentFixtures
}

func TestDeleteExperimentTestSuite(t *testing.T) {
	suite.Run(t, new(DeleteExperimentTestSuite))
}

func (s *DeleteExperimentTestSuite) SetupTest() {
	s.client = helpers.NewHttpClient(os.Getenv("SERVICE_BASE_URL"))
	fixtures, err := fixtures.NewExperimentFixtures(os.Getenv("DATABASE_DSN"))
	assert.Nil(s.T(), err)
	s.fixtures = fixtures
}

func (s *DeleteExperimentTestSuite) Test_Ok() {
	// 1. prepare database with test data.
	experiment, err := s.fixtures.CreateTestExperiment(context.Background(), &models.Experiment{
		Name: "Test Experiment",
		Tags: []models.ExperimentTag{
			{
				Key:   "key1",
				Value: "value1",
			},
		},
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
	defer func() {
		assert.Nil(s.T(), s.fixtures.UnloadFixtures())
	}()
	// 2. make actual API call.
	req := request.DeleteExperimentRequest{
		ID: fmt.Sprintf("%d", *experiment.ID),
	}
	resp := response.CreateExperimentResponse{}
	err := s.client.DoPostRequest(
		fmt.Sprintf("%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsDeleteRoute),
		req,
		&resp,
	)
	assert.Nil(s.T(), err)
	// 3. check actual API response.
}

func (s *DeleteExperimentTestSuite) Test_Error() {

}
*/
