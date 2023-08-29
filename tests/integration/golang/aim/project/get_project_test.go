//go:build integration

package run

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetProjectTestSuite struct {
	suite.Suite
	client          *helpers.HttpClient
	projectFixtures fixtures.ProjectFixtures
	project         response.GetProject
}

func TestGetProjectTestSuite(t *testing.T) {
	suite.Run(t, new(GetProjectTestSuite))
}

func (s *GetProjectTestSuite) SetupTest() {
	s.client = helpers.NewAimApiClient(helpers.GetServiceUri())

	projectFixtures, err := fixtures.NewProjectFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.projectFixtures = *projectFixtures

	project := projectFixtures.GetProject(context.Background())
	s.project = *project
}

func (s *GetProjectTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.projectFixtures.UnloadFixtures())
	}()
	var resp response.GetProject
	err := s.client.DoGetRequest(
		"/projects",
		&resp,
	)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), s.project.Name, resp.Name)
	assert.Equal(s.T(), s.project.Path, resp.Path)
	assert.Equal(s.T(), s.project.Description, resp.Description)
	assert.Equal(s.T(), s.project.TelemetryEnabled, resp.TelemetryEnabled)
}
