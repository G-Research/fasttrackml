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

type ListArtifactS3TestSuite struct {
	suite.Suite
	helpers.BaseTestSuite
	s3Client *s3.Client
}

func TestListArtifactS3TestSuite(t *testing.T) {
	suite.Run(t, new(ListArtifactS3TestSuite))
}

func (s *ListArtifactS3TestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest(s.T())

	s3Client, err := helpers.NewS3Client(helpers.GetS3EndpointUri())
	assert.Nil(s.T(), err)

	//nolint:errcheck
	helpers.CreateBuckets(s3Client)
	s.s3Client = s3Client
}

func (s *ListArtifactS3TestSuite) TearDownTest() {
	err := helpers.RemoveBuckets(s.s3Client)
	assert.Nil(s.T(), err)
}

func (s *ListArtifactS3TestSuite) Test_Ok() {
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
				ArtifactLocation: fmt.Sprintf("s3://%s/1", tt.bucket),
			})
			assert.Nil(s.T(), err)

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
			assert.Nil(s.T(), err)

			// 3. upload artifact objects to S3.
			_, err = s.s3Client.PutObject(context.Background(), &s3.PutObjectInput{
				Key:    aws.String(fmt.Sprintf("1/%s/artifacts/artifact.file1", runID)),
				Body:   strings.NewReader("contentX"),
				Bucket: aws.String(tt.bucket),
			})
			assert.Nil(s.T(), err)
			_, err = s.s3Client.PutObject(context.Background(), &s3.PutObjectInput{
				Key:    aws.String(fmt.Sprintf("1/%s/artifacts/artifact.dir/artifact.file2", runID)),
				Body:   strings.NewReader("contentXX"),
				Bucket: aws.String(tt.bucket),
			})
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
			assert.Nil(s.T(), err)

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

func (s *ListArtifactS3TestSuite) Test_Error() {
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
			assert.Nil(
				s.T(),
				s.MlflowClient.WithQuery(
					tt.request,
				).WithResponse(
					&resp,
				).DoRequest(
					fmt.Sprintf("%s%s", mlflow.ArtifactsRoutePrefix, mlflow.ArtifactsListRoute),
				),
			)
			assert.Nil(t, err)
			assert.Equal(s.T(), tt.error.Error(), resp.Error())
		})
	}
}
