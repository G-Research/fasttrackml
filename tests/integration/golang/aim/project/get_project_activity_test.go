package run

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetProjectActivityTestSuite struct {
	helpers.BaseTestSuite
}

func TestGetProjectActivityTestSuite(t *testing.T) {
	suite.Run(t, new(GetProjectActivityTestSuite))
}

func (s *GetProjectActivityTestSuite) Test_Ok() {
	runs, err := s.RunFixtures.CreateExampleRuns(context.Background(), s.DefaultExperiment, 10)
	s.Require().Nil(err)

	archivedRunsIds := []string{runs[0].ID, runs[1].ID}
	s.Require().Nil(s.RunFixtures.ArchiveRuns(context.Background(), s.DefaultNamespace.ID, archivedRunsIds))

	var resp response.ProjectActivityResponse
	s.Require().Nil(s.AIMClient().WithResponse(&resp).DoRequest("/projects/activity"))

	s.Equal(8, resp.NumActiveRuns)
	s.Equal(2, resp.NumArchivedRuns)
	s.Equal(1, resp.NumExperiments)
	s.Equal(10, resp.NumRuns)
	s.Equal(1, len(resp.ActivityMap))
	for _, v := range resp.ActivityMap {
		s.Equal(10, v)
	}
}
