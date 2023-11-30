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
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetArtifactLocalTestSuite struct {
	helpers.BaseTestSuite
}

func TestGetArtifactLocalTestSuite(t *testing.T) {
	suite.Run(t, new(GetArtifactLocalTestSuite))
}

func (s *GetArtifactLocalTestSuite) Test_Ok() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()

	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	s.Require().Nil(err)

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
		s.Run(tt.name, func() {
			// 1. create test experiment.
			experimentArtifactDir := s.T().TempDir()
			experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
				Name:             fmt.Sprintf("Test Experiment In Path %s", experimentArtifactDir),
				NamespaceID:      namespace.ID,
				LifecycleStage:   models.LifecycleStageActive,
				ArtifactLocation: fmt.Sprintf("%s%s", tt.prefix, experimentArtifactDir),
			})
			s.Require().Nil(err)

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
			s.Require().Nil(err)

			// 3. create artifacts.
			err = os.MkdirAll(runArtifactDir, fs.ModePerm)
			s.Require().Nil(err)
			err = os.WriteFile(filepath.Join(runArtifactDir, "artifact.file1"), []byte("contentX"), fs.ModePerm)
			s.Require().Nil(err)
			err = os.Mkdir(filepath.Join(runArtifactDir, "artifact.dir"), fs.ModePerm)
			s.Require().Nil(err)
			err = os.WriteFile(filepath.Join(runArtifactDir, "artifact.dir", "artifact.file2"), []byte("contentXX"), fs.ModePerm)
			s.Require().Nil(err)

			// 4. make actual API call for root dir file
			rootFileQuery := request.GetArtifactRequest{
				RunID: run.ID,
				Path:  "artifact.file1",
			}

			resp := new(bytes.Buffer)
			s.Require().Nil(s.MlflowClient().WithQuery(
				rootFileQuery,
			).WithResponseType(
				helpers.ResponseTypeBuffer,
			).WithResponse(
				resp,
			).DoRequest(
				"%s%s", mlflow.ArtifactsRoutePrefix, mlflow.ArtifactsGetRoute,
			))
			s.Equal("contentX", resp.String())

			// 5. make actual API call for sub dir file
			subDirQuery := request.GetArtifactRequest{
				RunID: run.ID,
				Path:  "artifact.dir/artifact.file2",
			}

			resp = new(bytes.Buffer)
			s.Require().Nil(s.MlflowClient().WithQuery(
				subDirQuery,
			).WithResponseType(
				helpers.ResponseTypeBuffer,
			).WithResponse(
				resp,
			).DoRequest(
				"%s%s", mlflow.ArtifactsRoutePrefix, mlflow.ArtifactsGetRoute,
			))
			s.Equal("contentXX", resp.String())
		})
	}
}

func (s *GetArtifactLocalTestSuite) Test_Error() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()

	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	s.Require().Nil(err)

	// create test experiment
	experimentArtifactDir := s.T().TempDir()
	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:             fmt.Sprintf("Test Experiment In Path %s", experimentArtifactDir),
		NamespaceID:      namespace.ID,
		LifecycleStage:   models.LifecycleStageActive,
		ArtifactLocation: experimentArtifactDir,
	})
	s.Require().Nil(err)

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
	s.Require().Nil(err)

	err = os.MkdirAll(filepath.Join(runArtifactDir, "subdir"), fs.ModePerm)
	s.Require().Nil(err)

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
		s.Run(tt.name, func() {
			resp := api.ErrorResponse{}
			s.Require().Nil(s.MlflowClient().WithQuery(
				tt.request,
			).WithResponse(
				&resp,
			).DoRequest(
				"%s%s", mlflow.ArtifactsRoutePrefix, mlflow.ArtifactsGetRoute,
			))
			s.Equal(tt.error.Error(), resp.Error())
		})
	}
}
