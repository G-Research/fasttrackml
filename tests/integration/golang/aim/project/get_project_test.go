//go:build integration

package run

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetProjectTestSuite struct {
	helpers.BaseTestSuite
}

func TestGetProjectTestSuite(t *testing.T) {
	suite.Run(t, new(GetProjectTestSuite))
}

func (s *GetProjectTestSuite) Test_Ok() {
	defer func() {
		require.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	_, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	require.Nil(s.T(), err)

	var resp response.GetProjectResponse
	require.Nil(s.T(), s.AIMClient().WithResponse(&resp).DoRequest("/projects"))
	assert.Equal(s.T(), "FastTrackML", resp.Name)
	// assert.Equal(s.T(), "", resp.Path)
	assert.Equal(s.T(), "", resp.Description)
	assert.Equal(s.T(), 0, resp.TelemetryEnabled)
}
