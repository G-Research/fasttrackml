//go:build integration

package artifact

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type ListArtifactLocalTestSuite struct {
	suite.Suite
	helpers.BaseTestSuite
}

func TestListArtifactLocalTestSuite(t *testing.T) {
	suite.Run(t, new(ListArtifactLocalTestSuite))
}

func (s *ListArtifactLocalTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest(s.T())
}

func (s *ListArtifactLocalTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	assert.Nil(s.T(), err)

	tests := []struct {
		name   string
		prefix string
	}{
		{
			name:   "TestWithFilePrefix",
			prefix: "file://",
		},
		{
			name:   "TestWithoutPrefix",
			prefix: "",
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			// 1. create test experiment.
			experimentArtifactDir := t.TempDir()
			experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
				Name: fmt.Sprintf("Test Experiment In Path %s", experimentArtifactDir),
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
				ArtifactLocation: fmt.Sprintf("%s%s", tt.prefix, experimentArtifactDir),
			})
			assert.Nil(s.T(), err)

			// 2. create test run.
			runID := strings.ReplaceAll(uuid.New().String(), "-", "")
			runArtifactDir := filepath.Join(experimentArtifactDir, runID, "artifacts")
			run, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
				ID:             runID,
				Status:         models.StatusRunning,
				SourceType:     "JOB",
				ExperimentID:   *experiment.ID,
				ArtifactURI:    fmt.Sprintf("%s%s", tt.prefix, runArtifactDir),
				LifecycleStage: models.LifecycleStageActive,
			})
			assert.Nil(s.T(), err)

			// 3. create artifacts.
			err = os.MkdirAll(runArtifactDir, fs.ModePerm)
			assert.Nil(s.T(), err)
			err = os.WriteFile(filepath.Join(runArtifactDir, "artifact.file1"), []byte("contentX"), fs.ModePerm)
			assert.Nil(s.T(), err)
			err = os.Mkdir(filepath.Join(runArtifactDir, "artifact.dir"), fs.ModePerm)
			assert.Nil(s.T(), err)
			err = os.WriteFile(
				filepath.Join(runArtifactDir, "artifact.dir", "artifact.file2"), []byte("contentXX"), fs.ModePerm,
			)
			assert.Nil(s.T(), err)

			// 4. make actual API call for root dir.
			rootDirQuery := request.ListArtifactsRequest{
				RunID: run.ID,
			}

			rootDirResp := response.ListArtifactsResponse{}
			assert.Nil(
				s.T(),
				s.MlflowClient.WithQuery(
					rootDirQuery,
				).WithResponse(
					&rootDirResp,
				).DoRequest(
					fmt.Sprintf("%s%s", mlflow.ArtifactsRoutePrefix, mlflow.ArtifactsListRoute),
				),
			)

			assert.Equal(s.T(), run.ArtifactURI, rootDirResp.RootURI)
			assert.Equal(s.T(), 2, len(rootDirResp.Files))
			assert.Equal(s.T(), []response.FilePartialResponse{
				{
					Path:     "artifact.dir",
					IsDir:    true,
					FileSize: 0,
				},
				{
					Path:     "artifact.file1",
					IsDir:    false,
					FileSize: 8,
				},
			}, rootDirResp.Files)
			assert.Nil(s.T(), err)

			// 5. make actual API call for sub dir.
			subDirQuery := request.ListArtifactsRequest{
				RunID: run.ID,
				Path:  "artifact.dir",
			}

			subDirResp := response.ListArtifactsResponse{}
			assert.Nil(
				s.T(),
				s.MlflowClient.WithQuery(
					subDirQuery,
				).WithResponse(
					&subDirResp,
				).DoRequest(
					fmt.Sprintf("%s%s", mlflow.ArtifactsRoutePrefix, mlflow.ArtifactsListRoute),
				),
			)

			assert.Equal(s.T(), run.ArtifactURI, subDirResp.RootURI)
			assert.Equal(s.T(), 1, len(subDirResp.Files))
			assert.Equal(s.T(), response.FilePartialResponse{
				Path:     "artifact.dir/artifact.file2",
				IsDir:    false,
				FileSize: 9,
			}, subDirResp.Files[0])
			assert.Nil(s.T(), err)

			// 6. make actual API call for non-existing dir.
			nonExistingDirQuery := request.ListArtifactsRequest{
				RunID: run.ID,
				Path:  "non-existing-dir",
			}

			nonExistingDirResp := response.ListArtifactsResponse{}
			assert.Nil(
				s.T(),
				s.MlflowClient.WithQuery(
					nonExistingDirQuery,
				).WithResponse(
					&nonExistingDirResp,
				).DoRequest(
					fmt.Sprintf("%s%s", mlflow.ArtifactsRoutePrefix, mlflow.ArtifactsListRoute),
				),
			)

			assert.Equal(s.T(), run.ArtifactURI, nonExistingDirResp.RootURI)
			assert.Equal(s.T(), 0, len(nonExistingDirResp.Files))
			assert.Nil(s.T(), err)
		})
	}
}

func (s *ListArtifactLocalTestSuite) Test_Error() {
	defer func() {
		assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	_, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	assert.Nil(s.T(), err)

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
		s.T().Run(tt.name, func(t *testing.T) {
			resp := api.ErrorResponse{}
			assert.Nil(s.T(), s.MlflowClient.WithQuery(
				tt.request,
			).WithResponse(
				&resp,
			).DoRequest(
				fmt.Sprintf("%s%s", mlflow.ArtifactsRoutePrefix, mlflow.ArtifactsListRoute)),
			)
			assert.Equal(s.T(), tt.error.Error(), resp.Error())
		})
	}
}
