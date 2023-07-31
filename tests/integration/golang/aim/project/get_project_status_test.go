//go:build integration

package run

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetProjectStatusTestSuite struct {
	suite.Suite
	client *helpers.HttpClient
}

func TestGetProjectStatusTestSuite(t *testing.T) {
	suite.Run(t, new(GetProjectStatusTestSuite))
}

func (s *GetProjectStatusTestSuite) SetupTest() {
	s.client = helpers.NewAimApiClient(helpers.GetServiceUri())
}

func (s *GetProjectStatusTestSuite) Test_Ok() {
	var resp string
	err := s.client.DoGetRequest(
		fmt.Sprintf("/projects/status"),
		&resp,
	)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), resp, "up-to-date")
}
