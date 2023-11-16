//go:build integration

package run

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetProjectStatusTestSuite struct {
	helpers.BaseTestSuite
}

func TestGetProjectStatusTestSuite(t *testing.T) {
	suite.Run(t, new(GetProjectStatusTestSuite))
}

func (s *GetProjectStatusTestSuite) Test_Ok() {
	defer func() {
		require.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	_, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	require.Nil(s.T(), err)

	var resp string
	require.Nil(s.T(), s.AIMClient().WithResponse(&resp).DoRequest("/projects/status"))
	assert.Equal(s.T(), "up-to-date", resp)
}
