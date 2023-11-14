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
		s.T().Run(tt.name, func(T *testing.T) {
			defer func() {
				require.Nil(s.T(), s.AppFixtures.UnloadFixtures())
			}()

			namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
				ID:                  1,
				Code:                "default",
				DefaultExperimentID: common.GetPointer(int32(0)),
			})
			require.Nil(s.T(), err)

			apps, err := s.AppFixtures.CreateApps(context.Background(), namespace, tt.expectedAppCount)
			require.Nil(s.T(), err)

			var resp []response.App
			require.Nil(s.T(), s.AIMClient.WithResponse(&resp).DoRequest("/apps"))
			assert.Equal(s.T(), tt.expectedAppCount, len(resp))
			for idx := 0; idx < tt.expectedAppCount; idx++ {
				assert.Equal(s.T(), apps[idx].ID.String(), resp[idx].ID)
				assert.Equal(s.T(), apps[idx].Type, resp[idx].Type)
				assert.Equal(s.T(), apps[idx].State, database.AppState(resp[idx].State))
				// TODO these timestamps are not populated by the endpoint -- should they be?
				// assert.NotEmpty(s.T(), resp[idx].CreatedAt)
				// assert.NotEmpty(s.T(), resp[idx].UpdatedAt)
			}
		})
	}
}
