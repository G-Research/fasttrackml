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
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type ListArtifactTestSuite struct {
	suite.Suite
	s3Client           *s3.Client
	runFixtures        *fixtures.RunFixtures
	serviceClient      *helpers.HttpClient
	experimentFixtures *fixtures.ExperimentFixtures
}

func TestListArtifactTestSuite(t *testing.T) {
	suite.Run(t, new(ListArtifactTestSuite))
}

func (s *ListArtifactTestSuite) SetupTest() {
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

func (s *ListArtifactTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.experimentFixtures.UnloadFixtures())
	}()

	testData := []struct {
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

			// 3. upload artifact object to S3.
			_, err = s.s3Client.PutObject(context.Background(), &s3.PutObjectInput{
				Key:    aws.String(fmt.Sprintf("1/%s/artifacts/artifact.file", runID)),
				Body:   strings.NewReader(`content`),
				Bucket: aws.String(tt.bucket),
			})
			assert.Nil(s.T(), err)

			// 4. make actual API call.
			query, err := urlquery.Marshal(request.ListArtifactsRequest{
				RunID: run.ID,
			})
			assert.Nil(s.T(), err)

			resp := response.ListArtifactsResponse{}
			err = s.serviceClient.DoGetRequest(
				fmt.Sprintf("%s%s?%s", mlflow.ArtifactsRoutePrefix, mlflow.ArtifactsListRoute, query),
				&resp,
			)
			assert.Nil(s.T(), err)

			assert.Equal(s.T(), 1, len(resp.Files))
			assert.Equal(s.T(), response.FilePartialResponse{
				Path:     fmt.Sprintf("1/%s/artifacts/artifact.file", runID),
				IsDir:    false,
				FileSize: 7,
			}, resp.Files[0])
			assert.Nil(s.T(), err)
		})
	}
}

func (s *ListArtifactTestSuite) Test_Error() {
	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.ListArtifactsRequest
	}{
		{
			name:    "EmptyOrIncorrectRunIDOrRunUUID",
			error:   api.NewInvalidParameterValueError(`Missing value for required parameter 'run_id'`),
			request: &request.ListArtifactsRequest{},
		},
		{
			name:  "PathIsRelativeAndContains1Dot",
			error: api.NewInvalidParameterValueError(`incorrect path has been provided. path has to be absolute`),
			request: &request.ListArtifactsRequest{
				RunID: "run_id",
				Path:  "./",
			},
		},
		{
			name:  "PathIsRelativeAndContains2Dots",
			error: api.NewInvalidParameterValueError(`incorrect path has been provided. path has to be absolute`),
			request: &request.ListArtifactsRequest{
				RunID: "run_id",
				Path:  "../",
			},
		},
	}

	for _, tt := range testData {
		s.T().Run(tt.name, func(t *testing.T) {
			query, err := urlquery.Marshal(tt.request)
			assert.Nil(s.T(), err)
			resp := api.ErrorResponse{}
			err = s.serviceClient.DoGetRequest(
				fmt.Sprintf("%s%s?%s", mlflow.ArtifactsRoutePrefix, mlflow.ArtifactsListRoute, query),
				&resp,
			)
			assert.Nil(t, err)
			assert.Equal(s.T(), tt.error.Error(), resp.Error())
		})
	}
}
