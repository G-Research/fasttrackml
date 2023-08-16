package artifact

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/repositories"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/service/artifact/storage"
)

func TestService_ListArtifacts_Ok(t *testing.T) {
	artifactStorage := storage.MockProvider{}
	artifactStorage.On(
		"List", "/artifact/uri", "",
	).Return(
		"/root/uri/",
		[]storage.ArtifactObject{
			{
				Path:  "/artifact/path1",
				Size:  1234567890,
				IsDir: false,
			},
			{
				Path:  "/artifact/path2",
				Size:  123456788,
				IsDir: true,
			},
		}, nil,
	)

	// init repository mocks.
	runRepository := repositories.MockRunRepositoryProvider{}
	runRepository.On(
		"GetByID",
		mock.AnythingOfType("*context.emptyCtx"),
		"id",
	).Return(&models.Run{
		ID:          "id",
		ArtifactURI: "/artifact/uri",
	}, nil)

	// call service under testing.
	service := NewService(&artifactStorage, &runRepository)
	rootURI, artifacts, err := service.ListArtifacts(
		context.TODO(),
		&request.ListArtifactsRequest{
			RunID: "id",
		},
	)

	assert.Nil(t, err)
	assert.Equal(t, "/root/uri/", rootURI)
	assert.Equal(t, []storage.ArtifactObject{
		{
			Path:  "/artifact/path1",
			Size:  1234567890,
			IsDir: false,
		},
		{
			Path:  "/artifact/path2",
			Size:  123456788,
			IsDir: true,
		},
	}, artifacts)
}

func TestService_ListArtifacts_Error(t *testing.T) {
	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.ListArtifactsRequest
		service func() *Service
	}{
		{
			name:    "EmptyOrIncorrectRunID",
			error:   api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'"),
			request: &request.ListArtifactsRequest{},
			service: func() *Service {
				return NewService(
					&storage.MockProvider{},
					&repositories.MockRunRepositoryProvider{},
				)
			},
		},
		{
			name:  "PathIsRelativeAndContains2Dots",
			error: api.NewInvalidParameterValueError("provided 'path' parameter is invalid"),
			request: &request.ListArtifactsRequest{
				RunID: "id",
				Path:  "../",
			},
			service: func() *Service {
				return NewService(
					&storage.MockProvider{},
					&repositories.MockRunRepositoryProvider{},
				)
			},
		},
		{
			name:  "RunNotFoundDatabaseError",
			error: api.NewInternalError("unable to get artifact URI for run 'id'"),
			request: &request.ListArtifactsRequest{
				RunID: "id",
			},
			service: func() *Service {
				runRepository := repositories.MockRunRepositoryProvider{}
				runRepository.On(
					"GetByID",
					mock.AnythingOfType("*context.emptyCtx"),
					"id",
				).Return(nil, errors.New("database error"))
				return NewService(
					&storage.MockProvider{},
					&runRepository,
				)
			},
		},
		{
			name:  "StorageError",
			error: api.NewInternalError("error getting artifact list from storage"),
			request: &request.ListArtifactsRequest{
				RunID: "id",
			},
			service: func() *Service {
				artifactStorage := storage.MockProvider{}
				artifactStorage.On(
					"List", "/artifact/uri", "",
				).Return(
					"", nil, errors.New("storage error"),
				)

				runRepository := repositories.MockRunRepositoryProvider{}
				runRepository.On(
					"GetByID",
					mock.AnythingOfType("*context.emptyCtx"),
					"id",
				).Return(&models.Run{
					ID:          "id",
					ArtifactURI: "/artifact/uri",
				}, nil)
				return NewService(
					&artifactStorage,
					&runRepository,
				)
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			// call service under testing.
			_, _, err := tt.service().ListArtifacts(context.TODO(), tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}
