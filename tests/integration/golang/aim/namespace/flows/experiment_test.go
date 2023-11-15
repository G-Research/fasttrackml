//go:build integration

package flows

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type ExperimentTestSuite struct {
	helpers.BaseTestSuite
}

func TestNamespaceTestSuite(t *testing.T) {
	suite.Run(t, new(ExperimentTestSuite))
}

func (s *ExperimentTestSuite) TearDownTest() {
	require.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
}

func (s *ExperimentTestSuite) Test_Ok() {}

func (s *ExperimentTestSuite) Test_Error() {}
