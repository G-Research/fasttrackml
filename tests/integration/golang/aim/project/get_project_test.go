//go:build integration

package run

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetProjectTestSuite struct {
	helpers.BaseTestSuite
}

func TestGetProjectTestSuite(t *testing.T) {
	suite.Run(t, new(GetProjectTestSuite))
}

func (s *GetProjectTestSuite) Test_Ok() {
	var resp response.GetProjectResponse
	s.Require().Nil(s.AIMClient().WithResponse(&resp).DoRequest("/projects"))
	s.Equal("FastTrackML", resp.Name)
	s.NotEmpty(resp.Path)
	s.Equal("", resp.Description)
	s.Equal(0, resp.TelemetryEnabled)
}
