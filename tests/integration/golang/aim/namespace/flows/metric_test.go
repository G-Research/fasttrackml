//go:build integration

package flows

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type MetricTestSuite struct {
	helpers.BaseTestSuite
}

func TestMetricTestSuite(t *testing.T) {
	suite.Run(t, new(MetricTestSuite))
}

func (s *MetricTestSuite) TearDownTest() {
	require.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
}

func (s *MetricTestSuite) Test_Ok() {}

func (s *MetricTestSuite) Test_Error() {}
