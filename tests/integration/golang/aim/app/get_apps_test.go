//go:build integration

package run

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/database"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetAppsTestSuite struct {
	helpers.BaseTestSuite
}

func TestGetAppsTestSuite(t *testing.T) {
	suite.Run(t, new(GetAppsTestSuite))
}

func (s *GetAppsTestSuite) Test_Ok() {
	tests := []struct {
		name             string
		expectedAppCount int
	}{
		{
			name:             "GetAppsWithExistingRows",
			expectedAppCount: 2,
		},
		{
			name:             "GetAppsWithNoRows",
			expectedAppCount: 0,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			defer func() {
				s.Require().Nil(s.AppFixtures.UnloadFixtures())
			}()

			namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
				ID:                  1,
				Code:                "default",
				DefaultExperimentID: common.GetPointer(int32(0)),
			})
			s.Require().Nil(err)

			apps, err := s.AppFixtures.CreateApps(context.Background(), namespace, tt.expectedAppCount)
			s.Require().Nil(err)

			var resp []response.App
			s.Require().Nil(s.AIMClient().WithResponse(&resp).DoRequest("/apps"))
			s.Equal(tt.expectedAppCount, len(resp))
			for idx := 0; idx < tt.expectedAppCount; idx++ {
				s.Equal(apps[idx].ID.String(), resp[idx].ID)
				s.Equal(apps[idx].Type, resp[idx].Type)
				s.Equal(apps[idx].State, database.AppState(resp[idx].State))
				s.NotEmpty(resp[idx].CreatedAt)
				s.NotEmpty(resp[idx].UpdatedAt)
			}
		})
	}
}
