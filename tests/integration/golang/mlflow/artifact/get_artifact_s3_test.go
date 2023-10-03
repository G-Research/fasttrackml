//go:build integration

package artifact

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/hetiansu5/urlquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetArtifactS3TestSuite struct {
	suite.Suite
	helpers.BaseTestSuite
	s3Client *s3.Client
}

func TestGetArtifactS3TestSuite(t *testing.T) {
	suite.Run(t, new(GetArtifactS3TestSuite))
}

func (s *GetArtifactS3TestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest(s.T())

	s3Client, err := helpers.NewS3Client(helpers.GetS3EndpointUri())
	assert.Nil(s.T(), err)

	s.s3Client = s3Client
}

func (s *GetArtifactS3TestSuite) Test_Ok() {
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
				ArtifactLocation: fmt.Sprintf("s3://%s/1", tt.bucket),
			})
			assert.Nil(s.T(), err)

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
			assert.Nil(s.T(), err)

			// upload artifact root object to S3
			putObjReq := &s3.PutObjectInput{
				Key:    aws.String(fmt.Sprintf("1/%s/artifacts/artifact.file", runID)),
				Body:   strings.NewReader("content"),
				Bucket: aws.String(tt.bucket),
			}
			_, err = s.s3Client.PutObject(context.Background(), putObjReq)
			assert.Nil(s.T(), err)

			// upload artifact subdir object to S3
			putObjReq = &s3.PutObjectInput{
				Key: aws.String(fmt.Sprintf(
					"1/%s/artifacts/artifact.subdir/artifact.file",
					runID),
				),
				Body:   strings.NewReader("subdir-object-content"),
				Bucket: aws.String(tt.bucket),
			}
			_, err = s.s3Client.PutObject(context.Background(), putObjReq)
			assert.Nil(s.T(), err)

			// make API call for root object
			query, err := urlquery.Marshal(request.GetArtifactRequest{
				RunID: run.ID,
				Path:  "artifact.file",
			})
			assert.Nil(s.T(), err)

			resp, err := s.MlflowClient.DoStreamRequest(
				http.MethodGet,
				fmt.Sprintf("%s%s?%s", mlflow.ArtifactsRoutePrefix, mlflow.ArtifactsGetRoute, query),
				nil,
			)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), "content", string(resp))

			// make API call for subdir object
			query, err = urlquery.Marshal(request.GetArtifactRequest{
				RunID: run.ID,
				Path:  "artifact.subdir/artifact.file",
			})
			assert.Nil(s.T(), err)

			resp, err = s.MlflowClient.DoStreamRequest(
				http.MethodGet,
				fmt.Sprintf("%s%s?%s", mlflow.ArtifactsRoutePrefix, mlflow.ArtifactsGetRoute, query),
				nil,
			)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), "subdir-object-content", string(resp))
		})
	}
}

func (s *GetArtifactS3TestSuite) Test_Error() {
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
	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:             "Test Experiment In Bucket bucket1",
		NamespaceID:      namespace.ID,
		LifecycleStage:   models.LifecycleStageActive,
		ArtifactLocation: "s3://bucket1/1",
	})
	assert.Nil(s.T(), err)

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
	assert.Nil(s.T(), err)

	// upload artifact subdir object to S3
	putObjReq := &s3.PutObjectInput{
		Key:    aws.String(fmt.Sprintf("1/%s/artifacts/artifact.subdir/artifact.file", runID)),
		Body:   strings.NewReader("content"),
		Bucket: aws.String("bucket1"),
	}
	_, err = s.s3Client.PutObject(context.Background(), putObjReq)
	assert.Nil(s.T(), err)

	tests := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.GetArtifactRequest
	}{
		{
			name:    "EmptyOrIncorrectRunIDOrRunUUID",
			error:   api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'"),
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
		{
			name: "S3IncompletePath",
			error: api.NewResourceDoesNotExistError(
				fmt.Sprintf("error getting artifact object for URI: s3:/bucket1/1/%s/artifacts/artifact.subdir", runID),
			),
			request: &request.GetArtifactRequest{
				RunID: runID,
				Path:  "artifact.subdir",
			},
		},
		{
			name: "NonExistentFile",
			error: api.NewResourceDoesNotExistError(
				fmt.Sprintf("error getting artifact object for URI: s3:/bucket1/1/%s/artifacts/non-existent-file", runID),
			),
			request: &request.GetArtifactRequest{
				RunID: runID,
				Path:  "non-existent-file",
			},
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			query, err := urlquery.Marshal(tt.request)
			assert.Nil(s.T(), err)
			resp := api.ErrorResponse{}
			err = s.MlflowClient.DoGetRequest(
				fmt.Sprintf("%s%s?%s", mlflow.ArtifactsRoutePrefix, mlflow.ArtifactsGetRoute, query),
				&resp,
			)
			assert.Nil(t, err)
			assert.Equal(s.T(), tt.error.Error(), resp.Error())
		})
	}
}
