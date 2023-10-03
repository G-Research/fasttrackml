//go:build integration

package artifact

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetArtifactLocalTestSuite struct {
	suite.Suite
	helpers.BaseTestSuite
}

func TestGetArtifactLocalTestSuite(t *testing.T) {
	suite.Run(t, new(GetArtifactLocalTestSuite))
}

func (s *GetArtifactLocalTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest(s.T())
}

func (s *GetArtifactLocalTestSuite) Test_Ok() {
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
				Name:             fmt.Sprintf("Test Experiment In Path %s", experimentArtifactDir),
				NamespaceID:      namespace.ID,
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
			err = os.WriteFile(filepath.Join(runArtifactDir, "artifact.dir", "artifact.file2"), []byte("contentXX"), fs.ModePerm)
			assert.Nil(s.T(), err)

			// 4. make actual API call for root dir file
			rootFileQuery := request.GetArtifactRequest{
				RunID: run.ID,
				Path:  "artifact.file1",
			}

			resp := new(bytes.Buffer)
			assert.Nil(s.T(), s.MlflowClient.WithQuery(
				rootFileQuery,
			).WithResponseType(
				helpers.ResponseTypeStream,
			).WithResponse(
				resp,
			).DoRequest(
				fmt.Sprintf("%s%s", mlflow.ArtifactsRoutePrefix, mlflow.ArtifactsGetRoute),
			))
			assert.Equal(s.T(), "contentX", resp.String())

			// 5. make actual API call for sub dir file
			subDirQuery := request.GetArtifactRequest{
				RunID: run.ID,
				Path:  "artifact.dir/artifact.file2",
			}

			resp = new(bytes.Buffer)
			assert.Nil(s.T(), s.MlflowClient.WithQuery(
				subDirQuery,
			).WithResponseType(
				helpers.ResponseTypeStream,
			).WithResponse(
				resp,
			).DoRequest(
				fmt.Sprintf("%s%s", mlflow.ArtifactsRoutePrefix, mlflow.ArtifactsGetRoute),
			))
			assert.Equal(s.T(), "contentXX", resp.String())
		})
	}
}

func (s *GetArtifactLocalTestSuite) Test_Error() {
	defer func() {
		assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	assert.Nil(s.T(), err)

	// create test experiment
	experimentArtifactDir := s.T().TempDir()
	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:             fmt.Sprintf("Test Experiment In Path %s", experimentArtifactDir),
		NamespaceID:      namespace.ID,
		LifecycleStage:   models.LifecycleStageActive,
		ArtifactLocation: experimentArtifactDir,
	})
	assert.Nil(s.T(), err)

	// create test run
	runID := strings.ReplaceAll(uuid.New().String(), "-", "")
	runArtifactDir := filepath.Join(experimentArtifactDir, runID, "artifacts")
	_, err = s.RunFixtures.CreateRun(context.Background(), &models.Run{
		ID:             runID,
		Status:         models.StatusRunning,
		SourceType:     "JOB",
		ExperimentID:   *experiment.ID,
		ArtifactURI:    runArtifactDir,
		LifecycleStage: models.LifecycleStageActive,
	})
	assert.Nil(s.T(), err)

	err = os.MkdirAll(filepath.Join(runArtifactDir, "subdir"), fs.ModePerm)
	assert.Nil(s.T(), err)

	tests := []struct {
		name    string
		error   *api.ErrorResponse
		request request.GetArtifactRequest
	}{
		{
			name:    "EmptyOrIncorrectRunIDOrRunUUID",
			error:   api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'"),
			request: request.GetArtifactRequest{},
		},
		{
			name:  "IncorrectPathProvidedCase1",
			error: api.NewInvalidParameterValueError("provided 'path' parameter is invalid"),
			request: request.GetArtifactRequest{
				RunID: "run_id",
				Path:  "..",
			},
		},
		{
			name:  "IncorrectPathProvidedCase2",
			error: api.NewInvalidParameterValueError("provided 'path' parameter is invalid"),
			request: request.GetArtifactRequest{
				RunID: "run_id",
				Path:  "./..",
			},
		},
		{
			name:  "IncorrectPathProvidedCase3",
			error: api.NewInvalidParameterValueError("provided 'path' parameter is invalid"),
			request: request.GetArtifactRequest{
				RunID: "run_id",
				Path:  "./../",
			},
		},
		{
			name:  "IncorrectPathProvidedCase4",
			error: api.NewInvalidParameterValueError("provided 'path' parameter is invalid"),
			request: request.GetArtifactRequest{
				RunID: "run_id",
				Path:  "foo/../bar",
			},
		},
		{
			name:  "IncorrectPathProvidedCase5",
			error: api.NewInvalidParameterValueError("provided 'path' parameter is invalid"),
			request: request.GetArtifactRequest{
				RunID: "run_id",
				Path:  "/foo/../bar",
			},
		},
		{
			name: "NonExistentPathProvided",
			error: api.NewResourceDoesNotExistError(
				fmt.Sprintf(
					"error getting artifact object for URI: %s/%s/artifacts/non-existent-file",
					experimentArtifactDir,
					runID,
				),
			),
			request: request.GetArtifactRequest{
				RunID: runID,
				Path:  "non-existent-file",
			},
		},
		{
			name: "ExistingDirectoryProvided",
			error: api.NewResourceDoesNotExistError(
				fmt.Sprintf(
					"error getting artifact object for URI: %s/%s/artifacts/subdir",
					experimentArtifactDir,
					runID,
				),
			),
			request: request.GetArtifactRequest{
				RunID: runID,
				Path:  "subdir",
			},
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			resp := api.ErrorResponse{}
			assert.Nil(t, s.MlflowClient.WithQuery(
				tt.request,
			).WithResponse(
				&resp,
			).DoRequest(
				fmt.Sprintf("%s%s", mlflow.ArtifactsRoutePrefix, mlflow.ArtifactsGetRoute),
			))
			assert.Equal(s.T(), tt.error.Error(), resp.Error())
		})
	}
}
