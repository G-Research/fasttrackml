//go:build integration

package run

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/encoding"
	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type SearchMetricsTestSuite struct {
	suite.Suite
	runs []*models.Run
	helpers.BaseTestSuite
}

func TestSearchMetricsTestSuite(t *testing.T) {
	suite.Run(t, new(SearchMetricsTestSuite))
}

func (s *SearchMetricsTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest(s.T())

	// 1. create test `namespace` and connect test `run`.
	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  0,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	assert.Nil(s.T(), err)

	// 2. create test `experiment` and connect test `run`.
	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		NamespaceID:    namespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	assert.Nil(s.T(), err)

	run, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
		ID:             "id",
		Name:           "chill-run",
		Status:         models.StatusScheduled,
		SourceType:     "JOB",
		LifecycleStage: models.LifecycleStageActive,
		ExperimentID:   *experiment.ID,
	})
	assert.Nil(s.T(), err)

	// 3. create test `metric` and test `the latest metric` and connect to run.
	metric, err := s.MetricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "key1",
		Value:     123.1,
		Timestamp: 1234567890,
		RunID:     run.ID,
		Step:      1,
		IsNan:     false,
		Iter:      1,
	})
	assert.Nil(s.T(), err)
	_, err = s.MetricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       metric.Key,
		Value:     123.1,
		Timestamp: 1234567890,
		Step:      1,
		IsNan:     false,
		RunID:     run.ID,
		LastIter:  1,
	})
	assert.Nil(s.T(), err)
}

func (s *SearchMetricsTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()
	tests := []struct {
		name  string
		query string
	}{
		{
			name:  "TestStartWithFunction",
			query: `q=(metric.name=='key1' and run.name.startswith("chill"))&p=500&report_progress=false`,
		},
		{
			name:  "TestEndWithFunction",
			query: `q=(metric.name=='key1' and run.name.endswith("run"))&p=500&report_progress=false`,
		},
		{
			name:  "TestRegexpMatchFunction",
			query: `q=(metric.name=='key1' and re.match("chill", run.name))&p=500&report_progress=false`,
		},
		{
			name:  "TestRegexpSearchFunction",
			query: `q=(metric.name=='key1' and re.search("run", run.name))&p=500&report_progress=false`,
		},
		{
			name:  "TestInFunction",
			query: `q=(metric.name=='key1' and 'chill' in run.name)&p=500&report_progress=false`,
		},
		{
			name:  "TestNotInFunction",
			query: `q=(metric.name=='key1' and 'grill' not in run.name)&p=500&report_progress=false`,
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			data, err := s.AIMClient.DoStreamRequest(
				http.MethodGet,
				fmt.Sprintf("/runs/search/metric?%s", tt.query),
				nil,
			)
			decodedData, err := encoding.Decode(bytes.NewBuffer(data))
			assert.Nil(s.T(), err)
			value, ok := decodedData["id.props.name"]
			assert.True(s.T(), ok)
			assert.Equal(s.T(), "chill-run", value)
		})
	}
}

func (s *SearchMetricsTestSuite) Test_Error() {
	defer func() {
		assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()
}
