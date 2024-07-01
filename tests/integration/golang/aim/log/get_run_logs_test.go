package log

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/encoding"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetRunLogsTestSuite struct {
	helpers.BaseTestSuite
}

func TestGetRunLogsTestSuite(t *testing.T) {
	suite.Run(t, new(GetRunLogsTestSuite))
}

func (s *GetRunLogsTestSuite) Test_Ok() {
	run, err := s.RunFixtures.CreateExampleRun(context.Background(), s.DefaultExperiment)
	s.Require().Nil(err)

	for i := 1; i <= 10; i++ {
		_, err := s.LogFixtures.CreateLog(context.Background(), &models.Log{
			Timestamp: time.Now().Unix(),
			Value:     fmt.Sprintf("value_%d", i),
			RunID:     run.ID,
		})
		s.Require().Nil(err)
	}

	resp := new(bytes.Buffer)
	s.Require().Nil(
		s.AIMClient().WithResponseType(
			helpers.ResponseTypeBuffer,
		).WithResponse(
			resp,
		).DoRequest("/runs/%s/logs", run.ID),
	)

	decodedData, err := encoding.NewDecoder(resp).Decode()
	s.Require().Nil(err)
	for i := 1; i <= 10; i++ {
		value, ok := decodedData[fmt.Sprintf("%d", i)]
		s.Require().True(ok)
		s.Require().Equal(fmt.Sprintf("value_%d", i), value)
	}
}
