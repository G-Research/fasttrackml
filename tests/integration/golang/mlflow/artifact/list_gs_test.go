//go:build integration

package artifact

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type ListArtifactGSTestSuite struct {
	helpers.BaseTestSuite
	gsClient    *storage.Client
	testBuckets []string
}

func TestListArtifactGSTestSuite(t *testing.T) {
	suite.Run(t, &ListArtifactGSTestSuite{
		testBuckets: []string{"bucket1", "bucket2"},
	})
}

func (s *ListArtifactGSTestSuite) SetupSuite() {
	gsClient, err := helpers.NewGSClient(helpers.GetGSEndpointUri())
	require.Nil(s.T(), err)
	s.gsClient = gsClient
}

func (s *ListArtifactGSTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest()
	require.Nil(s.T(), helpers.CreateGSBuckets(s.gsClient, s.testBuckets))
}

func (s *ListArtifactGSTestSuite) TearDownTest() {
	require.Nil(s.T(), helpers.DeleteGSBuckets(s.gsClient, s.testBuckets))
}

func (s *ListArtifactGSTestSuite) Test_Ok() {
	defer func() {
		require.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	require.Nil(s.T(), err)

	tests := []struct {
		name   string
		bucket string
	}{
		{
			name:   "TestWithBucket1",
			bucket: "bucket1",
		},
		{
			name:   "TestWithBucket2",
			bucket: "bucket2",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// 1. create test experiment.
			experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
				Name: fmt.Sprintf("Test Experiment In Bucket %s", tt.bucket),
				Tags: []models.ExperimentTag{
					{
						Key:   "key1",
						Value: "value1",
					},
				},
				NamespaceID: namespace.ID,
				CreationTime: sql.NullInt64{
					Int64: time.Now().UTC().UnixMilli(),
					Valid: true,
				},
				LastUpdateTime: sql.NullInt64{
					Int64: time.Now().UTC().UnixMilli(),
					Valid: true,
				},
				LifecycleStage:   models.LifecycleStageActive,
				ArtifactLocation: fmt.Sprintf("gs://%s/1", tt.bucket),
			})
			require.Nil(s.T(), err)

			// 2. create test run.
			runID := strings.ReplaceAll(uuid.New().String(), "-", "")
			run, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
				ID:             runID,
				Status:         models.StatusRunning,
				SourceType:     "JOB",
				ExperimentID:   *experiment.ID,
				ArtifactURI:    fmt.Sprintf("%s/%s/artifacts", experiment.ArtifactLocation, runID),
				LifecycleStage: models.LifecycleStageActive,
			})
			require.Nil(s.T(), err)

			// 3. upload artifact objects to GS.
			writer := s.gsClient.Bucket(
				tt.bucket,
			).Object(
				fmt.Sprintf("1/%s/artifacts/artifact.txt", runID),
			).NewWriter(
				context.Background(),
			)
			_, err = writer.Write([]byte("contentX"))
			require.Nil(s.T(), err)
			require.Nil(s.T(), writer.Close())

			writer = s.gsClient.Bucket(
				tt.bucket,
			).Object(
				fmt.Sprintf("1/%s/artifacts/artifact/artifact.txt", runID),
			).NewWriter(
				context.Background(),
			)
			_, err = writer.Write([]byte("contentXX"))
			require.Nil(s.T(), err)
			require.Nil(s.T(), writer.Close())

			// 4. make actual API call for root dir.
			rootDirQuery := request.ListArtifactsRequest{
				RunID: run.ID,
			}

			rootDirResp := response.ListArtifactsResponse{}
			require.Nil(
				s.T(),
				s.MlflowClient().WithQuery(
					rootDirQuery,
				).WithResponse(
					&rootDirResp,
				).DoRequest(
					"%s%s", mlflow.ArtifactsRoutePrefix, mlflow.ArtifactsListRoute,
				),
			)

			assert.Equal(s.T(), run.ArtifactURI, rootDirResp.RootURI)
			assert.Equal(s.T(), 2, len(rootDirResp.Files))
			assert.Equal(s.T(), []response.FilePartialResponse{
				{
					Path:     "artifact",
					IsDir:    true,
					FileSize: 0,
				},
				{
					Path:     "artifact.txt",
					IsDir:    false,
					FileSize: 8,
				},
			}, rootDirResp.Files)
			require.Nil(s.T(), err)

			// 5. make actual API call for sub dir.
			subDirQuery := request.ListArtifactsRequest{
				RunID: run.ID,
				Path:  "artifact",
			}

			subDirResp := response.ListArtifactsResponse{}
			require.Nil(
				s.T(),
				s.MlflowClient().WithQuery(
					subDirQuery,
				).WithResponse(
					&subDirResp,
				).DoRequest(
					"%s%s", mlflow.ArtifactsRoutePrefix, mlflow.ArtifactsListRoute,
				),
			)

			assert.Equal(s.T(), run.ArtifactURI, subDirResp.RootURI)
			assert.Equal(s.T(), 1, len(subDirResp.Files))
			assert.Equal(s.T(), response.FilePartialResponse{
				Path:     "artifact/artifact.txt",
				IsDir:    false,
				FileSize: 9,
			}, subDirResp.Files[0])
			require.Nil(s.T(), err)

			// 6. make actual API call for non-existing dir.
			nonExistingDirQuery := request.ListArtifactsRequest{
				RunID: run.ID,
				Path:  "non-existing-dir",
			}
			require.Nil(s.T(), err)

			nonExistingDirResp := response.ListArtifactsResponse{}
			require.Nil(
				s.T(),
				s.MlflowClient().WithQuery(
					nonExistingDirQuery,
				).WithResponse(
					&nonExistingDirResp,
				).DoRequest(
					"%s%s", mlflow.ArtifactsRoutePrefix, mlflow.ArtifactsListRoute,
				),
			)

			assert.Equal(s.T(), run.ArtifactURI, nonExistingDirResp.RootURI)
			assert.Equal(s.T(), 0, len(nonExistingDirResp.Files))
			require.Nil(s.T(), err)
		})
	}
}

func (s *ListArtifactGSTestSuite) Test_Error() {
	defer func() {
		require.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	_, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	require.Nil(s.T(), err)

	tests := []struct {
		name    string
		error   *api.ErrorResponse
		request request.ListArtifactsRequest
	}{
		{
			name:    "EmptyOrIncorrectRunIDOrRunUUID",
			error:   api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'"),
			request: request.ListArtifactsRequest{},
		},
		{
			name:  "IncorrectPathProvidedCase1",
			error: api.NewInvalidParameterValueError("provided 'path' parameter is invalid"),
			request: request.ListArtifactsRequest{
				RunID: "run_id",
				Path:  "..",
			},
		},
		{
			name:  "IncorrectPathProvidedCase2",
			error: api.NewInvalidParameterValueError("provided 'path' parameter is invalid"),
			request: request.ListArtifactsRequest{
				RunID: "run_id",
				Path:  "./..",
			},
		},
		{
			name:  "IncorrectPathProvidedCase3",
			error: api.NewInvalidParameterValueError("provided 'path' parameter is invalid"),
			request: request.ListArtifactsRequest{
				RunID: "run_id",
				Path:  "./../",
			},
		},
		{
			name:  "IncorrectPathProvidedCase4",
			error: api.NewInvalidParameterValueError("provided 'path' parameter is invalid"),
			request: request.ListArtifactsRequest{
				RunID: "run_id",
				Path:  "foo/../bar",
			},
		},
		{
			name:  "IncorrectPathProvidedCase5",
			error: api.NewInvalidParameterValueError("provided 'path' parameter is invalid"),
			request: request.ListArtifactsRequest{
				RunID: "run_id",
				Path:  "/foo/../bar",
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			resp := api.ErrorResponse{}
			require.Nil(
				s.T(),
				s.MlflowClient().WithQuery(
					tt.request,
				).WithResponse(
					&resp,
				).DoRequest(
					"%s%s", mlflow.ArtifactsRoutePrefix, mlflow.ArtifactsListRoute,
				),
			)
			require.Nil(s.T(), err)
			assert.Equal(s.T(), tt.error.Error(), resp.Error())
		})
	}
}
