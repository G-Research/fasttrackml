package artifact

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/api/request"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type CreateRunArtifactTestSuite struct {
	helpers.BaseTestSuite
}

func TestCreateRunArtifactTestSuite(t *testing.T) {
	suite.Run(t, new(CreateRunArtifactTestSuite))
}

func (s *CreateRunArtifactTestSuite) Test_Ok() {
	run, err := s.RunFixtures.CreateExampleRun(context.Background(), s.DefaultExperiment)
	s.Require().Nil(err)

	tests := []struct {
		name        string
		requestBody request.CreateRunArtifactRequest
	}{
		{
			name: "CreateRunArtifact",
			requestBody: request.CreateRunArtifactRequest{
				Iter:    1,
				Step:    1,
				Caption: "caption",
				Index:   1,
				Width:   100,
				Height:  200,
				Format:  "test",
				BlobURI: "/path/to/artifact",
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Require().Nil(
				s.AIMClient().WithMethod(
					http.MethodPost,
				).WithRequest(
					tt.requestBody,
				).DoRequest(
					"/runs/%s/artifact", run.ID,
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
