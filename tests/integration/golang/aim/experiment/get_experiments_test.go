//go:build integration

package experiment

import (
	"context"
	"fmt"
	"testing"

	"github.com/G-Research/fasttrackml/pkg/common/dao/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetExperimentsTestSuite struct {
	suite.Suite
	client   *helpers.HttpClient
	fixtures *fixtures.ExperimentFixtures
}

func TestGetExperimentsTestSuite(t *testing.T) {
	suite.Run(t, new(GetExperimentTestSuite))
}

func (s *GetExperimentsTestSuite) SetupTest() {
	s.client = helpers.NewAimApiClient(helpers.GetServiceUri())

	fixtures, err := fixtures.NewExperimentFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.fixtures = fixtures
}

func (s *GetExperimentsTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.fixtures.UnloadFixtures())
	}()
	experiments, err := s.fixtures.CreateExperiments(context.Background(), 5)
	assert.Nil(s.T(), err)
	var resp response.Experiments

	err = s.client.DoGetRequest(
		fmt.Sprintf(
			"/experiments/",
		),
		&resp,
	)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), len(experiments), len(resp))
	for idx := 0; idx < len(experiments); idx++ {
		assert.Equal(s.T(), *experiments[idx].ID, resp[idx].ID)
		assert.Equal(s.T(), experiments[idx].Name, resp[idx].Name)
		assert.Equal(s.T(), "", resp[idx].Description)
		assert.Equal(s.T(), float64(experiments[idx].CreationTime.Int64)/1000, resp[idx].CreationTime)
		assert.Equal(s.T(), experiments[idx].LifecycleStage == models.LifecycleStageDeleted, resp[idx].Archived)
		assert.Equal(s.T(), len(experiments[idx].Runs), resp[idx].RunCount)
	}
}
