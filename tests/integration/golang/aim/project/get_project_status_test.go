//go:build pipeline

package run

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetProjectStatusTestSuite struct {
	helpers.BaseTestSuite
}

func TestGetProjectStatusTestSuite(t *testing.T) {
	suite.Run(t, new(GetProjectStatusTestSuite))
}

func (s *GetProjectStatusTestSuite) Test_Ok() {
	var resp string
	s.Require().Nil(s.AIMClient().WithResponse(&resp).DoRequest("/projects/status"))
	s.Equal("up-to-date", resp)
}
