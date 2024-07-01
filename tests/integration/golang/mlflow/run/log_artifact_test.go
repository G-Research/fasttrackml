package run

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type LogArtifactTestSuite struct {
	helpers.BaseTestSuite
}

func TestLogArtifactTestSuite(t *testing.T) {
	suite.Run(t, new(LogArtifactTestSuite))
}

func (s *LogArtifactTestSuite) Test_Ok() {
	run, err := s.RunFixtures.CreateExampleRun(context.Background(), s.DefaultExperiment)
	s.Require().Nil(err)

	tests := []struct {
		name        string
		requestBody request.LogArtifactRequest
	}{
		{
			name: "CreateRunArtifact",
			requestBody: request.LogArtifactRequest{
				Iter:    1,
				Step:    1,
				Index:   1,
				RunID:   run.ID,
				Width:   100,
				Height:  200,
				Format:  "test",
				BlobURI: "/path/to/artifact",
				Caption: "caption",
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Require().Nil(
				s.MlflowClient().WithMethod(
					http.MethodPost,
				).WithRequest(
					tt.requestBody,
				).DoRequest(
					"%s%s", mlflow.RunsRoutePrefix, mlflow.RunsLogArtifactRoute,
				),
			)
		})
	}

	artifact, err := s.ArtifactFixtures.GetArtifactByRunID(context.Background(), run.ID)
	s.Require().Nil(err)
	s.Equal(int64(1), artifact.Iter)
	s.Equal(int64(1), artifact.Step)
	s.Equal("caption", artifact.Caption)
	s.Equal(int64(1), artifact.Index)
	s.Equal(int64(100), artifact.Width)
	s.Equal(int64(200), artifact.Height)
	s.Equal("test", artifact.Format)
	s.Equal("/path/to/artifact", artifact.BlobURI)
}
