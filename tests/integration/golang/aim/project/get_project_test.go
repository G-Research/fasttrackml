//go:build integration

package run

import (
	"context"
	"testing"

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
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()

	_, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	s.Require().Nil(err)

	var resp response.GetProjectResponse
	s.Require().Nil(s.AIMClient().WithResponse(&resp).DoRequest("/projects"))
	s.Equal("FastTrackML", resp.Name)
	// s.Equal( "", resp.Path)
	s.Equal("", resp.Description)
	s.Equal(0, resp.TelemetryEnabled)
}
