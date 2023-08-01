//go:build integration

package run

import (
	"context"
	"fmt"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetProjectTestSuite struct {
	suite.Suite
	client          *helpers.HttpClient
	projectFixtures *fixtures.ProjectFixtures
	project         *fiber.Map
}

func TestGetProjectTestSuite(t *testing.T) {
	suite.Run(t, new(GetProjectTestSuite))
}

func (s *GetProjectTestSuite) SetupTest() {
	s.client = helpers.NewAimApiClient(helpers.GetServiceUri())

	projectFixtures, err := fixtures.NewProjectFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.projectFixtures = projectFixtures

	project := projectFixtures.GetProject(context.Background())
	s.project = project
}

func (s *GetProjectTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.projectFixtures.UnloadFixtures())
	}()
	var resp fiber.Map
	err := s.client.DoGetRequest(
		fmt.Sprintf("/projects"),
		&resp,
	)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), (*s.project)["name"], resp["name"])
	assert.Equal(s.T(), (*s.project)["path"], resp["path"])
	assert.Equal(s.T(), (*s.project)["description"], resp["description"])
	assert.Equal(s.T(), (*s.project)["telemetry_enabled"], resp["telemetry_enabled"])
}
