//go:build integration

package artifact

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetArtifactGSTestSuite struct {
	helpers.BaseTestSuite
	gsClient    *storage.Client
	testBuckets []string
}

func TestGetArtifactGSTestSuite(t *testing.T) {
	suite.Run(t, &GetArtifactGSTestSuite{
		testBuckets: []string{"bucket1", "bucket2"},
	})
}

func (s *GetArtifactGSTestSuite) SetupSuite() {
	gsClient, err := helpers.NewGSClient(helpers.GetGSEndpointUri())
	require.Nil(s.T(), err)
	s.gsClient = gsClient
}

func (s *GetArtifactGSTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest()
	require.Nil(s.T(), helpers.CreateGSBuckets(s.gsClient, s.testBuckets))
}

func (s *GetArtifactGSTestSuite) TearDownTest() {
	require.Nil(s.T(), helpers.DeleteGSBuckets(s.gsClient, s.testBuckets))
}

func (s *GetArtifactGSTestSuite) Test_Ok() {
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
		s.T().Run(tt.name, func(t *testing.T) {
			// create test experiment
			experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
				Name:             fmt.Sprintf("Test Experiment In Bucket %s", tt.bucket),
				NamespaceID:      namespace.ID,
				LifecycleStage:   models.LifecycleStageActive,
				ArtifactLocation: fmt.Sprintf("gs://%s/1", tt.bucket),
			})
			require.Nil(s.T(), err)

			// create test run
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

			// upload artifact root object to GS
			writer := s.gsClient.Bucket(tt.bucket).Object(
				fmt.Sprintf("/1/%s/artifacts/artifact.txt", runID),
			).NewWriter(context.Background())
			_, err = writer.Write([]byte("content"))
			require.Nil(s.T(), err)
			require.Nil(s.T(), writer.Close())

			// upload artifact subdir object to GS
			writer = s.gsClient.Bucket(tt.bucket).Object(
				fmt.Sprintf("/1/%s/artifacts/artifact/artifact.txt", runID),
			).NewWriter(context.Background())
			_, err = writer.Write([]byte("subdir-object-content"))
			require.Nil(s.T(), err)
			require.Nil(s.T(), writer.Close())

			// make API call for root object
			query := request.GetArtifactRequest{
				RunID: run.ID,
				Path:  "artifact.txt",
			}

			resp := new(bytes.Buffer)
			require.Nil(s.T(), s.MlflowClient().WithQuery(
				query,
			).WithResponseType(
				helpers.ResponseTypeBuffer,
			).WithResponse(
				resp,
			).DoRequest(
				fmt.Sprintf("%s%s", mlflow.ArtifactsRoutePrefix, mlflow.ArtifactsGetRoute),
			))
			assert.Equal(s.T(), "content", resp.String())

			// make API call for subdir object
			query = request.GetArtifactRequest{
				RunID: run.ID,
				Path:  "artifact/artifact.txt",
			}

			resp = new(bytes.Buffer)
			require.Nil(s.T(), s.MlflowClient().WithQuery(
				query,
			).WithResponseType(
				helpers.ResponseTypeBuffer,
			).WithResponse(
				resp,
			).DoRequest(
				fmt.Sprintf("%s%s", mlflow.ArtifactsRoutePrefix, mlflow.ArtifactsGetRoute),
			))
			assert.Equal(s.T(), "subdir-object-content", resp.String())
		})
	}
}

func (s *GetArtifactGSTestSuite) Test_Error() {
	defer func() {
		require.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	require.Nil(s.T(), err)

	// create test experiment
	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:             "Test Experiment In Bucket bucket1",
		NamespaceID:      namespace.ID,
		LifecycleStage:   models.LifecycleStageActive,
		ArtifactLocation: "gs://bucket1/1",
	})
	require.Nil(s.T(), err)

	// create test run
	runID := strings.ReplaceAll(uuid.New().String(), "-", "")
	_, err = s.RunFixtures.CreateRun(context.Background(), &models.Run{
		ID:             runID,
		Status:         models.StatusRunning,
		SourceType:     "JOB",
		ExperimentID:   *experiment.ID,
		ArtifactURI:    fmt.Sprintf("%s/%s/artifacts", experiment.ArtifactLocation, runID),
		LifecycleStage: models.LifecycleStageActive,
	})
	require.Nil(s.T(), err)

	// upload artifact subdir object to GS
	require.Nil(s.T(), err)
	writer := s.gsClient.Bucket("bucket1").Object(
		fmt.Sprintf("1/%s/artifacts/artifact/artifact.file", runID),
	).NewWriter(context.Background())
	_, err = writer.Write([]byte("content"))
	require.Nil(s.T(), err)
	require.Nil(s.T(), writer.Close())

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
			name: "GSIncompletePath",
			error: api.NewResourceDoesNotExistError(
				fmt.Sprintf("error getting artifact object for URI: gs:/bucket1/1/%s/artifacts/artifact", runID),
			),
			request: request.GetArtifactRequest{
				RunID: runID,
				Path:  "artifact",
			},
		},
		{
			name: "NonExistentFile",
			error: api.NewResourceDoesNotExistError(
				fmt.Sprintf("error getting artifact object for URI: gs:/bucket1/1/%s/artifacts/non-existent-file", runID),
			),
			request: request.GetArtifactRequest{
				RunID: runID,
				Path:  "non-existent-file",
			},
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			resp := api.ErrorResponse{}
			require.Nil(t, s.MlflowClient().WithQuery(
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
