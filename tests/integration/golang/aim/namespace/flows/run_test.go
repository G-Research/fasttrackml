//go:build integration

package flows

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type RunTestSuite struct {
	helpers.BaseTestSuite
}

func TestRunTestSuite(t *testing.T) {
	suite.Run(t, new(RunTestSuite))
}

func (s *RunTestSuite) TearDownTest() {
	require.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
}

func (s *RunTestSuite) Test_Ok() {}

func (s *RunTestSuite) Test_Error() {}
