//go:build integration

package artifact

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/hetiansu5/urlquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

<<<<<<<< HEAD:tests/integration/golang/mlflow/artifact/get_artifact_test.go
type GetArtifactTestSuite struct {
========
type ListArtifactS3TestSuite struct {
>>>>>>>> main:tests/integration/golang/mlflow/artifact/list_s3_test.go
	suite.Suite
	s3Client           *s3.Client
	runFixtures        *fixtures.RunFixtures
	serviceClient      *helpers.HttpClient
	experimentFixtures *fixtures.ExperimentFixtures
}

<<<<<<<< HEAD:tests/integration/golang/mlflow/artifact/get_artifact_test.go
func TestGetArtifactTestSuite(t *testing.T) {
	suite.Run(t, new(GetArtifactTestSuite))
}

func (s *GetArtifactTestSuite) SetupTest() {
========
func TestListArtifactS3TestSuite(t *testing.T) {
	suite.Run(t, new(ListArtifactS3TestSuite))
}

func (s *ListArtifactS3TestSuite) SetupTest() {
>>>>>>>> main:tests/integration/golang/mlflow/artifact/list_s3_test.go
	s3Client, err := helpers.NewS3Client(helpers.GetS3EndpointUri())
	assert.Nil(s.T(), err)
	s.s3Client = s3Client

	s.serviceClient = helpers.NewMlflowApiClient(helpers.GetServiceUri())

	experimentFixtures, err := fixtures.NewExperimentFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.experimentFixtures = experimentFixtures

	runFixtures, err := fixtures.NewRunFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.runFixtures = runFixtures
}

<<<<<<<< HEAD:tests/integration/golang/mlflow/artifact/get_artifact_test.go
func (s *GetArtifactTestSuite) Test_Ok() {
========
func (s *ListArtifactS3TestSuite) Test_Ok() {
>>>>>>>> main:tests/integration/golang/mlflow/artifact/list_s3_test.go
	defer func() {
		assert.Nil(s.T(), s.experimentFixtures.UnloadFixtures())
	}()

	testData := []struct {
		name   string
		bucket string
	}{
		{
			name:   "TestWithBucket1",
			bucket: "bucket3",
		},
		{
			name:   "TestWithBucket2",
			bucket: "bucket4",
		},
	}

	for _, tt := range testData {
		s.T().Run(tt.name, func(t *testing.T) {
			// 1. create test experiment.
			experiment, err := s.experimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
				Name: fmt.Sprintf("Test Experiment In Bucket %s", tt.bucket),
				Tags: []models.ExperimentTag{
					{
						Key:   "key1",
						Value: "value1",
					},
				},
				CreationTime: sql.NullInt64{
					Int64: time.Now().UTC().UnixMilli(),
					Valid: true,
				},
				LastUpdateTime: sql.NullInt64{
					Int64: time.Now().UTC().UnixMilli(),
					Valid: true,
				},
				LifecycleStage:   models.LifecycleStageActive,
				ArtifactLocation: fmt.Sprintf("s3://%s/1", tt.bucket),
			})
			assert.Nil(s.T(), err)

			// 2. create test run.
			runID := strings.ReplaceAll(uuid.New().String(), "-", "")
			run, err := s.runFixtures.CreateRun(context.Background(), &models.Run{
				ID:             runID,
				Status:         models.StatusRunning,
				SourceType:     "JOB",
				ExperimentID:   *experiment.ID,
				ArtifactURI:    fmt.Sprintf("%s/%s/artifacts", experiment.ArtifactLocation, runID),
				LifecycleStage: models.LifecycleStageActive,
			})
			assert.Nil(s.T(), err)

			// 3. upload artifact objects to S3.
			_, err = s.s3Client.PutObject(context.Background(), &s3.PutObjectInput{
				Key:    aws.String(fmt.Sprintf("1/%s/artifacts/artifact.file1", runID)),
				Body:   strings.NewReader(`contentX`),
				Bucket: aws.String(tt.bucket),
			})
			assert.Nil(s.T(), err)
			_, err = s.s3Client.PutObject(context.Background(), &s3.PutObjectInput{
				Key:    aws.String(fmt.Sprintf("1/%s/artifacts/artifact.dir/artifact.file2", runID)),
				Body:   strings.NewReader(`contentXX`),
				Bucket: aws.String(tt.bucket),
			})
			assert.Nil(s.T(), err)

<<<<<<<< HEAD:tests/integration/golang/mlflow/artifact/get_artifact_test.go
			// 4. make actual API call.
			query, err := urlquery.Marshal(request.GetArtifactRequest{
========
			// 4. make actual API call for root dir.
			rootDirQuery, err := urlquery.Marshal(request.ListArtifactsRequest{
>>>>>>>> main:tests/integration/golang/mlflow/artifact/list_s3_test.go
				RunID: run.ID,
				Path:  "artifact.file",
			})
			assert.Nil(s.T(), err)

<<<<<<<< HEAD:tests/integration/golang/mlflow/artifact/get_artifact_test.go
			resp, err := s.serviceClient.DoGetRequestNoUnmarshalling(
				fmt.Sprintf("%s%s?%s", mlflow.ArtifactsRoutePrefix, mlflow.ArtifactsGetRoute, query),
			)
			assert.Nil(s.T(), err)

			assert.Equal(s.T(), "content", string(resp))
========
			rootDirResp := response.ListArtifactsResponse{}
			err = s.serviceClient.DoGetRequest(
				fmt.Sprintf("%s%s?%s", mlflow.ArtifactsRoutePrefix, mlflow.ArtifactsListRoute, rootDirQuery),
				&rootDirResp,
			)
			assert.Nil(s.T(), err)

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
			subDirQuery, err := urlquery.Marshal(request.ListArtifactsRequest{
				RunID: run.ID,
				Path:  "artifact.dir",
			})
			assert.Nil(s.T(), err)

			subDirResp := response.ListArtifactsResponse{}
			err = s.serviceClient.DoGetRequest(
				fmt.Sprintf("%s%s?%s", mlflow.ArtifactsRoutePrefix, mlflow.ArtifactsListRoute, subDirQuery),
				&subDirResp,
			)
			assert.Nil(s.T(), err)

			assert.Equal(s.T(), run.ArtifactURI, subDirResp.RootURI)
			assert.Equal(s.T(), 1, len(subDirResp.Files))
			assert.Equal(s.T(), response.FilePartialResponse{
				Path:     "artifact.dir/artifact.file2",
				IsDir:    false,
				FileSize: 9,
			}, subDirResp.Files[0])
			assert.Nil(s.T(), err)

			// 6. make actual API call for non-existing dir.
			nonExistingDirQuery, err := urlquery.Marshal(request.ListArtifactsRequest{
				RunID: run.ID,
				Path:  "non-existing-dir",
			})
			assert.Nil(s.T(), err)

			nonExistingDirResp := response.ListArtifactsResponse{}
			err = s.serviceClient.DoGetRequest(
				fmt.Sprintf("%s%s?%s", mlflow.ArtifactsRoutePrefix, mlflow.ArtifactsListRoute, nonExistingDirQuery),
				&nonExistingDirResp,
			)
			assert.Nil(s.T(), err)

			assert.Equal(s.T(), run.ArtifactURI, nonExistingDirResp.RootURI)
			assert.Equal(s.T(), 0, len(nonExistingDirResp.Files))
			assert.Nil(s.T(), err)
>>>>>>>> main:tests/integration/golang/mlflow/artifact/list_s3_test.go
		})
	}
}

<<<<<<<< HEAD:tests/integration/golang/mlflow/artifact/get_artifact_test.go
func (s *GetArtifactTestSuite) Test_Error() {
	defer func() {
		assert.Nil(s.T(), s.experimentFixtures.UnloadFixtures())
	}()

========
func (s *ListArtifactS3TestSuite) Test_Error() {
>>>>>>>> main:tests/integration/golang/mlflow/artifact/list_s3_test.go
	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.GetArtifactRequest
	}{
		{
			name:    "EmptyOrIncorrectRunIDOrRunUUID",
			error:   api.NewInvalidParameterValueError(`Missing value for required parameter 'run_id'`),
			request: &request.GetArtifactRequest{},
		},
		{
			name:  "IncorrectPathProvidedCase1",
			error: api.NewInvalidParameterValueError("provided 'path' parameter is invalid"),
			request: &request.GetArtifactRequest{
				RunID: "run_id",
				Path:  "..",
			},
		},
		{
			name:  "IncorrectPathProvidedCase2",
			error: api.NewInvalidParameterValueError("provided 'path' parameter is invalid"),
			request: &request.GetArtifactRequest{
				RunID: "run_id",
				Path:  "./..",
			},
		},
		{
			name:  "IncorrectPathProvidedCase3",
			error: api.NewInvalidParameterValueError("provided 'path' parameter is invalid"),
			request: &request.GetArtifactRequest{
				RunID: "run_id",
				Path:  "./../",
			},
		},
		{
			name:  "IncorrectPathProvidedCase4",
			error: api.NewInvalidParameterValueError("provided 'path' parameter is invalid"),
			request: &request.GetArtifactRequest{
				RunID: "run_id",
				Path:  "foo/../bar",
			},
		},
		{
			name:  "IncorrectPathProvidedCase5",
			error: api.NewInvalidParameterValueError("provided 'path' parameter is invalid"),
			request: &request.GetArtifactRequest{
				RunID: "run_id",
				Path:  "/foo/../bar",
			},
		},
	}

	for _, tt := range testData {
		s.T().Run(tt.name, func(t *testing.T) {
			query, err := urlquery.Marshal(tt.request)
			assert.Nil(s.T(), err)
			resp := api.ErrorResponse{}
			err = s.serviceClient.DoGetRequest(
				fmt.Sprintf("%s%s?%s", mlflow.ArtifactsRoutePrefix, mlflow.ArtifactsGetRoute, query),
				&resp,
			)
			assert.Nil(t, err)
			assert.Equal(s.T(), tt.error.Error(), resp.Error())
		})
	}
}
