//go:build integration

package namespace

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type NamespaceTestSuite struct {
	helpers.BaseTestSuite
}

func TestNamespaceTestSuite(t *testing.T) {
	suite.Run(t, new(NamespaceTestSuite))
}

func (s *NamespaceTestSuite) TearDownTest() {
	require.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
}

func (s *NamespaceTestSuite) Test_Ok() {}

func (s *NamespaceTestSuite) Test_Error() {}
