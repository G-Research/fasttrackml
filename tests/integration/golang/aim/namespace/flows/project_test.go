//go:build integration

package flows

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type ProjectTestSuite struct {
	helpers.BaseTestSuite
}

func TestProjectTestSuite(t *testing.T) {
	suite.Run(t, new(ProjectTestSuite))
}

func (s *ProjectTestSuite) TearDownTest() {
	require.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
}

func (s *ProjectTestSuite) Test_Ok() {}

func (s *ProjectTestSuite) Test_Error() {}
