//go:build integration

package run

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/datatypes"

	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetRunInfoTestSuite struct {
	helpers.BaseTestSuite
	namespaceID uint
}

func TestGetRunInfoTestSuite(t *testing.T) {
	suite.Run(t, new(GetRunInfoTestSuite))
}

func (s *GetRunInfoTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest()
	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	s.Require().Nil(err)
	s.namespaceID = namespace.ID
}

func (s *GetRunInfoTestSuite) Test_Ok() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()

	// create test data.
	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		NamespaceID:    s.namespaceID,
		LifecycleStage: models.LifecycleStageActive,
	})
	s.Require().Nil(err)

	run, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
		ID:             "id",
		Name:           "TestRun",
		Status:         models.StatusScheduled,
		StartTime:      sql.NullInt64{Int64: 123456789, Valid: true},
		EndTime:        sql.NullInt64{Int64: 123456789, Valid: true},
		SourceType:     "JOB",
		ExperimentID:   *experiment.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	s.Require().Nil(err)

	metricContext, err := s.ContextFixtures.CreateContext(context.Background(), &models.Context{
		Json: datatypes.JSON(`{"key": "key", "value": "value"}`),
	})
	s.Require().Nil(err)

	_, err = s.MetricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       "key",
		Value:     123,
		Timestamp: 123456789,
		Step:      1,
		IsNan:     false,
		RunID:     run.ID,
		ContextID: common.GetPointer(metricContext.ID),
	})
	s.Require().Nil(err)

	// run tests over the test data.
	tests := []struct {
		name  string
		runID string
	}{
		{
			name:  "GetOneRun",
			runID: run.ID,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			var resp response.GetRunInfo
			s.Require().Nil(
				s.AIMClient().WithResponse(&resp).DoRequest("/runs/%s/info", tt.runID),
			)
			s.Equal(s.run.Name, resp.Props.Name)
			s.Equal(fmt.Sprintf("%v", s.run.ExperimentID), resp.Props.Experiment.ID)
			s.Equal(float64(s.run.StartTime.Int64)/1000, resp.Props.CreationTime)
			s.Equal(float64(s.run.EndTime.Int64)/1000, resp.Props.EndTime)
			s.Require().JSONEq(metricContext.Json.String(), string(resp.Traces.Metric[0].Context))
			expectedTags := make(map[string]string, len(s.run.Tags))
			for _, tag := range s.run.Tags {
				expectedTags[tag.Key] = tag.Value
			}
			s.Equal(expectedTags, resp.Params.Tags)
		})
	}
}

func (s *GetRunInfoTestSuite) Test_Error() {
	tests := []struct {
		name  string
		runID string
	}{
		{
			name:  "GetNonexistentRun",
			runID: uuid.NewString(),
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			var resp response.Error
			s.Require().Nil(s.AIMClient().WithResponse(&resp).DoRequest("/runs/%s/info", tt.runID))
			s.Equal("Not Found", resp.Message)
		})
	}
}
