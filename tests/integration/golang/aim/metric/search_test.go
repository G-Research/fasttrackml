//go:build integration

package run

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/hetiansu5/urlquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/encoding"
	"github.com/G-Research/fasttrackml/pkg/api/aim/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type SearchMetricsTestSuite struct {
	suite.Suite
	runs               []*models.Run
	client             *helpers.HttpClient
	runFixtures        *fixtures.RunFixtures
	paramFixtures      *fixtures.ParamFixtures
	metricFixtures     *fixtures.MetricFixtures
	experimentFixtures *fixtures.ExperimentFixtures
}

func TestSearchMetricsTestSuite(t *testing.T) {
	suite.Run(t, new(SearchMetricsTestSuite))
}

func (s *SearchMetricsTestSuite) SetupTest() {
	s.client = helpers.NewAimApiClient(helpers.GetServiceUri())
	runFixtures, err := fixtures.NewRunFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.runFixtures = runFixtures
	paramFixtures, err := fixtures.NewParamFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.paramFixtures = paramFixtures
	metricFixtures, err := fixtures.NewMetricFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.metricFixtures = metricFixtures
	expFixtures, err := fixtures.NewExperimentFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.experimentFixtures = expFixtures
}

func (s *SearchMetricsTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.experimentFixtures.UnloadFixtures())
	}()

	// create test experiments.
	experiment, err := s.experimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		LifecycleStage: models.LifecycleStageActive,
	})
	assert.Nil(s.T(), err)

	// create different test runs and attach tags, metrics, params, etc.
	run1, err := s.runFixtures.CreateRun(context.Background(), &models.Run{
		ID:         "id1",
		Name:       "TestRun1",
		UserID:     "1",
		Status:     models.StatusRunning,
		SourceType: "JOB",
		StartTime: sql.NullInt64{
			Int64: 123456789,
			Valid: true,
		},
		EndTime: sql.NullInt64{
			Int64: 123456789,
			Valid: true,
		},
		ExperimentID:   *experiment.ID,
		ArtifactURI:    "artifact_uri1",
		LifecycleStage: models.LifecycleStageActive,
	})
	assert.Nil(s.T(), err)
	_, err = s.metricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "TestMetric1",
		Value:     1.1,
		Timestamp: 123456789,
		Step:      1,
		IsNan:     false,
		RunID:     run1.ID,
		Iter:      1,
	})
	assert.Nil(s.T(), err)
	metric1Run1, err := s.metricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       "TestMetric1",
		Value:     1.1,
		Timestamp: 123456789,
		Step:      1,
		IsNan:     false,
		RunID:     run1.ID,
		LastIter:  1,
	})
	assert.Nil(s.T(), err)
	_, err = s.metricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "TestMetric2",
		Value:     2.1,
		Timestamp: 123456789,
		Step:      5,
		IsNan:     false,
		RunID:     run1.ID,
		Iter:      2,
	})
	assert.Nil(s.T(), err)
	metric2Run1, err := s.metricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       "TestMetric2",
		Value:     2.1,
		Timestamp: 123456789,
		Step:      5,
		IsNan:     false,
		RunID:     run1.ID,
		LastIter:  2,
	})
	assert.Nil(s.T(), err)
	_, err = s.metricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "TestMetric3",
		Value:     3.1,
		Timestamp: 123456789,
		Step:      10,
		IsNan:     false,
		RunID:     run1.ID,
		Iter:      3,
	})
	assert.Nil(s.T(), err)
	metric3Run1, err := s.metricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       "TestMetric3",
		Value:     3.1,
		Timestamp: 123456789,
		Step:      10,
		IsNan:     false,
		RunID:     run1.ID,
		LastIter:  3,
	})
	assert.Nil(s.T(), err)
	_, err = s.paramFixtures.CreateParam(context.Background(), &models.Param{
		Key:   "param1",
		Value: "value1",
		RunID: run1.ID,
	})
	assert.Nil(s.T(), err)

	run2, err := s.runFixtures.CreateRun(context.Background(), &models.Run{
		ID:         "id2",
		Name:       "TestRun2",
		UserID:     "2",
		Status:     models.StatusScheduled,
		SourceType: "JOB",
		StartTime: sql.NullInt64{
			Int64: 111111111,
			Valid: true,
		},
		EndTime: sql.NullInt64{
			Int64: 444444444,
			Valid: true,
		},
		ExperimentID:   *experiment.ID,
		ArtifactURI:    "artifact_uri2",
		LifecycleStage: models.LifecycleStageActive,
	})
	assert.Nil(s.T(), err)
	_, err = s.metricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "TestMetric1",
		Value:     0.5,
		Timestamp: 111111111,
		Step:      4,
		IsNan:     false,
		RunID:     run2.ID,
		Iter:      3,
	})
	assert.Nil(s.T(), err)
	metric1Run2, err := s.metricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       "TestMetric1",
		Value:     0.5,
		Timestamp: 111111111,
		Step:      4,
		IsNan:     false,
		RunID:     run2.ID,
		LastIter:  3,
	})
	assert.Nil(s.T(), err)
	_, err = s.metricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "TestMetric2",
		Value:     2.1,
		Timestamp: 222222222,
		Step:      5,
		IsNan:     false,
		RunID:     run2.ID,
		Iter:      2,
	})
	assert.Nil(s.T(), err)
	metric2Run2, err := s.metricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       "TestMetric2",
		Value:     2.1,
		Timestamp: 222222222,
		Step:      5,
		IsNan:     false,
		RunID:     run2.ID,
		LastIter:  2,
	})
	assert.Nil(s.T(), err)
	_, err = s.metricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "TestMetric3",
		Value:     3.1,
		Timestamp: 333333333,
		Step:      10,
		IsNan:     false,
		RunID:     run2.ID,
		Iter:      3,
	})
	assert.Nil(s.T(), err)
	metric3Run2, err := s.metricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       "TestMetric3",
		Value:     3.1,
		Timestamp: 333333333,
		Step:      10,
		IsNan:     false,
		RunID:     run2.ID,
		LastIter:  3,
	})
	assert.Nil(s.T(), err)
	_, err = s.paramFixtures.CreateParam(context.Background(), &models.Param{
		Key:   "param2",
		Value: "value2",
		RunID: run2.ID,
	})
	assert.Nil(s.T(), err)

	run3, err := s.runFixtures.CreateRun(context.Background(), &models.Run{
		ID:         "id3",
		Name:       "TestRun3",
		UserID:     "3",
		Status:     models.StatusScheduled,
		SourceType: "JOB",
		StartTime: sql.NullInt64{
			Int64: 222222222,
			Valid: true,
		},
		EndTime: sql.NullInt64{
			Int64: 444444444,
			Valid: true,
		},
		ExperimentID:   *experiment.ID,
		ArtifactURI:    "artifact_uri3",
		LifecycleStage: models.LifecycleStageActive,
	})
	assert.Nil(s.T(), err)
	_, err = s.metricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "TestMetric1",
		Value:     1.2,
		Timestamp: 1511111111,
		Step:      6,
		IsNan:     false,
		RunID:     run3.ID,
		Iter:      3,
	})
	assert.Nil(s.T(), err)
	metric1Run3, err := s.metricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       "TestMetric1",
		Value:     1.2,
		Timestamp: 1511111111,
		Step:      6,
		IsNan:     false,
		RunID:     run3.ID,
		LastIter:  3,
	})
	assert.Nil(s.T(), err)
	_, err = s.metricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "TestMetric2",
		Value:     1.6,
		Timestamp: 2522222222,
		Step:      2,
		IsNan:     false,
		RunID:     run3.ID,
		Iter:      4,
	})
	assert.Nil(s.T(), err)
	metric2Run3, err := s.metricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       "TestMetric2",
		Value:     1.6,
		Timestamp: 2522222222,
		Step:      2,
		IsNan:     false,
		RunID:     run3.ID,
		LastIter:  4,
	})
	assert.Nil(s.T(), err)

	_, err = s.paramFixtures.CreateParam(context.Background(), &models.Param{
		Key:   "param3",
		Value: "value3",
		RunID: run3.ID,
	})
	assert.Nil(s.T(), err)

	runs := []*models.Run{run1, run2, run3}
	tests := []struct {
		name    string
		request request.SearchMetricRequest
		metrics []*models.LatestMetric
	}{
		// Search Metric Name
		{
			name: "SearchMetricNameOperationEqual",
			request: request.SearchMetricRequest{
				Query: `(metric.name == "TestMetric1")`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameOperationNotEqual",
			request: request.SearchMetricRequest{
				Query: `(metric.name != "TestMetric3")`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric1Run2,
				metric2Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameOperationStartsWith",
			request: request.SearchMetricRequest{
				Query: `(metric.name.startswith("Test"))`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameOperationEndsWith",
			request: request.SearchMetricRequest{
				Query: `(metric.name.endswith("3"))`,
			},

			metrics: []*models.LatestMetric{
				metric3Run1,
				metric3Run2,
			},
		},
		// Search Metric Last
		{
			name: "SearchMetricLastOperationEquals",
			request: request.SearchMetricRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last == 1.6)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastOperationNotEquals",
			request: request.SearchMetricRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last != 1.6)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastOperationGrater",
			request: request.SearchMetricRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last > 1.6)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastOperationLess",
			request: request.SearchMetricRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last < 1.6)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastOperationGraterOrEqual",
			request: request.SearchMetricRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last >= 1.6)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastOperationLessOrEqual",
			request: request.SearchMetricRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last <= 1.6)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		// Search Metric Last Step
		{
			name: "SearchMetricLastStepOperationEquals",
			request: request.SearchMetricRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step == 1)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricLastStepOperationNotEquals",
			request: request.SearchMetricRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step != 1)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepOperationGrater",
			request: request.SearchMetricRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step > 1)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepOperationLess",
			request: request.SearchMetricRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step < 3)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepOperationGraterOrEqual",
			request: request.SearchMetricRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step >= 1)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepOperationLessOrEqual",
			request: request.SearchMetricRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step <= 3)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
			},
		},
		// Search Metric Name (equal operation) and Run Name
		{
			name: "SearchMetricNameAndRunNameOperationRegexpMatchFunction",
			request: request.SearchMetricRequest{
				Query: `(metric.name == "TestMetric1" and re.match("TestRun1", run.name))`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationRegexpSearchFunction",
			request: request.SearchMetricRequest{
				Query: `(metric.name == "TestMetric1" and re.search("TestRun1", run.name))`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationEquals",
			request: request.SearchMetricRequest{
				Query: `(metric.name == "TestMetric1" and run.name == "TestRun1")`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationNotEquals",
			request: request.SearchMetricRequest{
				Query: `(metric.name == "TestMetric1" and run.name != "TestRun1")`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationStartsWith",
			request: request.SearchMetricRequest{
				Query: `(metric.name == "TestMetric1" and run.name.startswith("Test"))`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationEndsWith",
			request: request.SearchMetricRequest{
				Query: `(metric.name == "TestMetric1" and run.name.endswith("Run2"))`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
			},
		},
		// Search Metric Name (not equal operation) and Run Name
		{
			name: "SearchMetricNameAndRunNameOperationNotEqualsAndRegexpMatchFunction",
			request: request.SearchMetricRequest{
				Query: `(metric.name != "TestMetric1" and re.match("TestRun1", run.name))`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationNotEqualsAndRegexpSearchFunction",
			request: request.SearchMetricRequest{
				Query: `(metric.name != "TestMetric1" and re.search("TestRun1", run.name))`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationNotEqualsAndEquals",
			request: request.SearchMetricRequest{
				Query: `(metric.name != "TestMetric1" and run.name == "TestRun1")`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationNotEqualsAndNotEquals",
			request: request.SearchMetricRequest{
				Query: `(metric.name != "TestMetric1" and run.name != "TestRun1")`,
			},

			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationNotEqualsAndStartsWith",
			request: request.SearchMetricRequest{
				Query: `(metric.name != "TestMetric1" and run.name.startswith("Test"))`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationNotEqualsAndEndsWith",
			request: request.SearchMetricRequest{
				Query: `(metric.name != "TestMetric1" and run.name.endswith("Run2"))`,
			},

			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
			},
		},
		// Search Metric Name (startswith operation) and Run Name
		{
			name: "SearchMetricNameAndRunNameOperationStartsWithAndRegexpMatchFunction",
			request: request.SearchMetricRequest{
				Query: `(metric.name.startswith("Test") and re.match("TestRun1", run.name))`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationStartsWithAndRegexpSearchFunction",
			request: request.SearchMetricRequest{
				Query: `(metric.name.startswith("Test") and re.search("TestRun1", run.name))`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationStartsWithAndEquals",
			request: request.SearchMetricRequest{
				Query: `(metric.name.startswith("Test") and run.name == "TestRun1")`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationStartsWithAndNotEquals",
			request: request.SearchMetricRequest{
				Query: `(metric.name.startswith("Test") and run.name != "TestRun1")`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationStartsWithAndStartsWith",
			request: request.SearchMetricRequest{
				Query: `(metric.name.startswith("Test") and run.name.startswith("Test"))`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationStartsWithAndEndsWith",
			request: request.SearchMetricRequest{
				Query: `(metric.name.startswith("Test") and run.name.endswith("Run2"))`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
			},
		},
		// Search Metric Name (endswith operation) and Run Name
		{
			name: "SearchMetricNameAndRunNameOperationEndsWithAndRegexpMatchFunction",
			request: request.SearchMetricRequest{
				Query: `(metric.name.endswith("Metric2") and re.match("TestRun1", run.name))`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationEndsWithAndRegexpSearchFunction",
			request: request.SearchMetricRequest{
				Query: `(metric.name.endswith("Metric2") and re.search("TestRun1", run.name))`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationEndsWithAndEquals",
			request: request.SearchMetricRequest{
				Query: `(metric.name.endswith("Metric2") and run.name == "TestRun1")`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationEndsWithAndNotEquals",
			request: request.SearchMetricRequest{
				Query: `(metric.name.endswith("Metric2") and run.name != "TestRun1")`,
			},

			metrics: []*models.LatestMetric{
				metric2Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationEndsWithAndStartsWith",
			request: request.SearchMetricRequest{
				Query: `(metric.name.endswith("Metric2") and run.name.startswith("Test"))`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric2Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationEndsWithAndEndsWith",
			request: request.SearchMetricRequest{
				Query: `(metric.name.endswith("Metric2") and run.name.endswith("Run2"))`,
			},

			metrics: []*models.LatestMetric{
				metric2Run2,
			},
		},
		// Search Metric Name (equal operation) and Run Duration
		{
			name: "SearchMetricNameAndRunDurationOperationEqual",
			request: request.SearchMetricRequest{
				Query: `(metric.name == "TestMetric1" and run.duration == 222222)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationNotEqual",
			request: request.SearchMetricRequest{
				Query: `(metric.name == "TestMetric1" and run.duration != 222222)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationGrater",
			request: request.SearchMetricRequest{
				Query: `(metric.name == "TestMetric1" and run.duration > 0)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationGraterOrEquals",
			request: request.SearchMetricRequest{
				Query: `(metric.name == "TestMetric1" and run.duration >= 0)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationLess",
			request: request.SearchMetricRequest{
				Query: `(metric.name == "TestMetric1" and run.duration < 333333)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationLessOrEquals",
			request: request.SearchMetricRequest{
				Query: `(metric.name == "TestMetric1" and run.duration <= 333333)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
			},
		},
		// Search Metric Name (not equals operation) and Run Duration
		{
			name: "SearchMetricNameAndRunDurationOperationNotEqualsAndEquals",
			request: request.SearchMetricRequest{
				Query: `(metric.name != "TestMetric1" and run.duration == 222222)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationNotEqualsAndNotEquals",
			request: request.SearchMetricRequest{
				Query: `(metric.name != "TestMetric1" and run.duration != 222222)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationNotEqualsAndGrater",
			request: request.SearchMetricRequest{
				Query: `(metric.name != "TestMetric1" and run.duration > 0)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationNotEqualsAndGraterOrEquals",
			request: request.SearchMetricRequest{
				Query: `(metric.name != "TestMetric1" and run.duration >= 0)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationNotEqualsAndLess",
			request: request.SearchMetricRequest{
				Query: `(metric.name != "TestMetric1" and run.duration < 333333)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationNotEqualsAndLessOrEquals",
			request: request.SearchMetricRequest{
				Query: `(metric.name != "TestMetric1" and run.duration <= 333333)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		// Search Metric Name (startswith operation) and Run Duration
		{
			name: "SearchMetricNameAndRunDurationOperationStartsWithAndEquals",
			request: request.SearchMetricRequest{
				Query: `(metric.name.startswith("Test") and run.duration == 222222)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationStartsWithAndNotEquals",
			request: request.SearchMetricRequest{
				Query: `(metric.name.startswith("Test") and run.duration != 222222)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationStartsWithAndGrater",
			request: request.SearchMetricRequest{
				Query: `(metric.name.startswith("Test") and run.duration > 0)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationStartsWithAndGraterOrEquals",
			request: request.SearchMetricRequest{
				Query: `(metric.name.startswith("Test") and run.duration >= 0)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationStartsWithAndLess",
			request: request.SearchMetricRequest{
				Query: `(metric.name.startswith("Test") and run.duration < 333333)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationStartsWithAndLessOrEquals",
			request: request.SearchMetricRequest{
				Query: `(metric.name.startswith("Test") and run.duration <= 333333)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		// Search Metric Name (endswith operation) and Run Duration
		{
			name: "SearchMetricNameAndRunDurationOperationEndsWithAndEquals",
			request: request.SearchMetricRequest{
				Query: `(metric.name.endswith("Metric2") and run.duration == 222222)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationEndsWithAndNotEquals",
			request: request.SearchMetricRequest{
				Query: `(metric.name.endswith("Metric2") and run.duration != 222222)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric2Run2,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationEndsWithAndGrater",
			request: request.SearchMetricRequest{
				Query: `(metric.name.endswith("Metric2") and run.duration > 0)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationEndsWithAndGraterOrEquals",
			request: request.SearchMetricRequest{
				Query: `(metric.name.endswith("Metric2") and run.duration >= 0)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric2Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationEndsWithAndLess",
			request: request.SearchMetricRequest{
				Query: `(metric.name.endswith("Metric2") and run.duration < 333333)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationEndsWithAndLessOrEquals",
			request: request.SearchMetricRequest{
				Query: `(metric.name.endswith("Metric2") and run.duration <= 333333)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric2Run2,
				metric2Run3,
			},
		},
		// Search Metric Name (equal operation) and Run Hash
		{
			name: "SearchMetricNameAndRunHashOperationEquals",
			request: request.SearchMetricRequest{
				Query: fmt.Sprintf(`(metric.name == "TestMetric1" and run.hash == "%s")`, run1.ID),
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunHashOperationNotEquals",
			request: request.SearchMetricRequest{
				Query: fmt.Sprintf(`(metric.name == "TestMetric1" and run.hash != "%s")`, run1.ID),
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
			},
		},
		// Search Metric Name (not equal operation) and Run Hash
		{
			name: "SearchMetricNameAndRunHashOperationNotEqualsAndEquals",
			request: request.SearchMetricRequest{
				Query: fmt.Sprintf(`(metric.name != "TestMetric1" and run.hash == "%s")`, run1.ID),
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunHashOperationNotEqualsAndNotEquals",
			request: request.SearchMetricRequest{
				Query: fmt.Sprintf(`(metric.name != "TestMetric1" and run.hash != "%s")`, run1.ID),
			},

			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		// Search Metric Name (startswith operation) and Run Hash
		{
			name: "SearchMetricNameAndRunHashOperationStartsWithAndEquals",
			request: request.SearchMetricRequest{
				Query: fmt.Sprintf(`(metric.name.startswith("Test") and run.hash == "%s")`, run1.ID),
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunHashOperationStartsWithAndNotEquals",
			request: request.SearchMetricRequest{
				Query: fmt.Sprintf(`(metric.name.startswith("Test") and run.hash != "%s")`, run1.ID),
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		// Search Metric Name (endswith operation) and Run Hash
		{
			name: "SearchMetricNameAndRunHashOperationEndsWithAndEquals",
			request: request.SearchMetricRequest{
				Query: fmt.Sprintf(`(metric.name.endswith("Metric2") and run.hash == "%s")`, run1.ID),
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunHashOperationEndsWithAndNotEqual",
			request: request.SearchMetricRequest{
				Query: fmt.Sprintf(`(metric.name.endswith("Metric2") and run.hash != "%s")`, run1.ID),
			},

			metrics: []*models.LatestMetric{
				metric2Run2,
				metric2Run3,
			},
		},
		// Search Metric Name (equals operation) and Run FinalizedAt
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationGreater",
			request: request.SearchMetricRequest{
				Query: `(metric.name == "TestMetric1" and run.finalized_at > 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationGreaterOrEqual",
			request: request.SearchMetricRequest{
				Query: `(metric.name == "TestMetric1" and run.finalized_at >= 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationLess",
			request: request.SearchMetricRequest{
				Query: `(metric.name == "TestMetric1" and run.finalized_at < 444444444)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationLessOrEqual",
			request: request.SearchMetricRequest{
				Query: `(metric.name == "TestMetric1" and run.finalized_at <= 444444444)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedOperationAtEquals",
			request: request.SearchMetricRequest{
				Query: `(metric.name == "TestMetric1" and run.finalized_at == 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationNotEquals",
			request: request.SearchMetricRequest{
				Query: `(metric.name == "TestMetric1" and run.finalized_at != 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
			},
		},
		// Search Metric Name (not equals operation) and Run FinalizedAt
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationNotEqualsAndGreater",
			request: request.SearchMetricRequest{
				Query: `(metric.name != "TestMetric1" and run.finalized_at > 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationNotEqualsAndGreaterOrEqual",
			request: request.SearchMetricRequest{
				Query: `(metric.name != "TestMetric1" and run.finalized_at >= 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationNotEqualsAndLess",
			request: request.SearchMetricRequest{
				Query: `(metric.name != "TestMetric1" and run.finalized_at < 444444444)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationNotEqualsAndLessOrEqual",
			request: request.SearchMetricRequest{
				Query: `(metric.name != "TestMetric1" and run.finalized_at <= 444444444)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationNotEqualsAndEquals",
			request: request.SearchMetricRequest{
				Query: `(metric.name != "TestMetric1" and run.finalized_at == 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationNotEqualsAndNotEquals",
			request: request.SearchMetricRequest{
				Query: `(metric.name != "TestMetric1" and run.finalized_at != 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		// Search Metric Name (startswith operation) and Run FinalizedAt
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationStartsWithAndGreater",
			request: request.SearchMetricRequest{
				Query: `(metric.name.startswith("Test") and run.finalized_at > 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationStartsWithAndGreaterOrEqual",
			request: request.SearchMetricRequest{
				Query: `(metric.name.startswith("Test") and run.finalized_at >= 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationStartsWithAndLess",
			request: request.SearchMetricRequest{
				Query: `(metric.name.startswith("Test") and run.finalized_at < 444444444)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricNamehAndRunFinalizedAtOperationStartsWithAndLessOrEqual",
			request: request.SearchMetricRequest{
				Query: `(metric.name.startswith("Test") and run.finalized_at <= 444444444)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationStartsWithAndEquals",
			request: request.SearchMetricRequest{
				Query: `(metric.name.startswith("Test") and run.finalized_at == 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationStartsWithAndNotEquals",
			request: request.SearchMetricRequest{
				Query: `(metric.name.startswith("Test") and run.finalized_at != 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		// Search Metric Name (startswith operation) and Run CreatedAt
		{
			name: "SearchMetricNameStartsWithAndRunCreatedAtOperationGreater",
			request: request.SearchMetricRequest{
				Query: `(metric.name.startswith("Test") and run.created_at > 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameStartsWithAndRunCreatedAtOperationGreaterOrEqual",
			request: request.SearchMetricRequest{
				Query: `(metric.name.startswith("Test") and run.created_at >= 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameStartsWithAndRunCreatedAtOperationLess",
			request: request.SearchMetricRequest{
				Query: `(metric.name.startswith("Test") and run.created_at < 222222222)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricNameStartsWithAndRunCreatedAtOperationLessOrEqual",
			request: request.SearchMetricRequest{
				Query: `(metric.name.startswith("Test") and run.created_at <= 222222222)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameStartsWithAndRunCreatedAtOperationEquals",
			request: request.SearchMetricRequest{
				Query: `(metric.name.startswith("Test") and run.created_at == 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricNameStartsWithAndRunCreatedAtOperationNotEquals",
			request: request.SearchMetricRequest{
				Query: `(metric.name.startswith("Test") and run.created_at != 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},

		// Search Metric Name (endswith operation) and Run FinalizedAt
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationEndsWithAndGreater",
			request: request.SearchMetricRequest{
				Query: `(metric.name.endswith("Metric2") and run.finalized_at > 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationEndsWithAndGreaterOrEqual",
			request: request.SearchMetricRequest{
				Query: `(metric.name.endswith("Metric2") and run.finalized_at >= 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric2Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationEndsWithAndLess",
			request: request.SearchMetricRequest{
				Query: `(metric.name.endswith("Metric2") and run.finalized_at < 444444444)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationEndsWithAndLessOrEqual",
			request: request.SearchMetricRequest{
				Query: `(metric.name.endswith("Metric2") and run.finalized_at <= 444444444)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric2Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationEndsWithAndEquals",
			request: request.SearchMetricRequest{
				Query: `(metric.name.endswith("Metric2") and run.finalized_at == 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationEndsWithAndNotEquals",
			request: request.SearchMetricRequest{
				Query: `(metric.name.endswith("Metric2") and run.finalized_at != 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run2,
				metric2Run3,
			},
		},
		// Search Metric Name (endswith operation) and Run CreatedAt
		{
			name: "SearchMetricNameEndsWithAndRunCreatedAtOperationNotEqualsAndGreater",
			request: request.SearchMetricRequest{
				Query: `(metric.name.endswith("Metric2") and run.created_at > 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameEndsWithAndRunCreatedAtOperationNotEqualsAndGreaterOrEqual",
			request: request.SearchMetricRequest{
				Query: `(metric.name.endswith("Metric2") and run.created_at >= 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameEndsWithAndRunCreatedAtOperationNotEqualsAndLess",
			request: request.SearchMetricRequest{
				Query: `(metric.name.endswith("Metric2") and run.created_at < 222222222)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric2Run2,
			},
		},
		{
			name: "SearchMetricNameEndsWithAndRunCreatedAtOperationNotEqualsAndLessOrEqual",
			request: request.SearchMetricRequest{
				Query: `(metric.name.endswith("Metric2") and run.created_at <= 222222222)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric2Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameEndsWithAndRunCreatedAtOperationNotEqualsAndEquals",
			request: request.SearchMetricRequest{
				Query: `(metric.name.endswith("Metric2") and run.created_at == 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
			},
		},
		{
			name: "SearchMetricNameEndsWithAndRunCreatedAtOperationNotEqualsAndNotEquals",
			request: request.SearchMetricRequest{
				Query: `(metric.name.endswith("Metric2") and run.created_at != 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run2,
				metric2Run3,
			},
		},
		// aa

		{
			name: "SearchMetricLastRunNameOperationEqualsAndEqual",
			request: request.SearchMetricRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and metric.last == 1.6 and run.name == "TestRun1"`,
			},

			metrics: []*models.LatestMetric{},
		},
		{
			name: "SearchMetricLastRunNameOperationNotEqualsAndEqual",
			request: request.SearchMetricRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and metric.last != 1.6 and run.name == "TestRun1"`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastRunNameOperationGraterAndEqual",
			request: request.SearchMetricRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and metric.last > 1.6 and run.name == "TestRun1"`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastRunNameOperationLessAndEqual",
			request: request.SearchMetricRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and metric.last < 1.6 and run.name == "TestRun1"`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricLastRunNameOperationGraterOrEqualAndEqual",
			request: request.SearchMetricRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and metric.last >= 1.6 and run.name == "TestRun1"`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastRunNameOperationLessOrEqualAndEqual",
			request: request.SearchMetricRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and metric.last <= 1.6 and run.name == "TestRun1"`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricLastRunNameOperationEqualsAndNotEqual",
			request: request.SearchMetricRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and metric.last == 1.6 and run.name != "TestRun1"`,
			},

			metrics: []*models.LatestMetric{
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastRunNameOperationNotEqualsAndNotEqual",
			request: request.SearchMetricRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and metric.last != 1.6 and run.name != "TestRun1"`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastRunNameOperationGraterAndNotEqual",
			request: request.SearchMetricRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and metric.last > 1.6 and run.name != "TestRun1"`,
			},

			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastRunNameOperationLessAndNotEqual",
			request: request.SearchMetricRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and metric.last < 1.6 and run.name != "TestRun1"`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastRunNameOperationGraterOrEqualAndNotEqual",
			request: request.SearchMetricRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and metric.last >= 1.6 and run.name != "TestRun1"`,
			},

			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastRunNameOperationLessOrEqualAndNotEqual",
			request: request.SearchMetricRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and metric.last <= 1.6 and run.name != "TestRun1"`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricComplexQuery",
			request: request.SearchMetricRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2")) and metric.last_step >= 1 and (run.name.endswith("2") or re.match("TestRun1", run.name)) and (metric.last < 1.6) and run.duration > 0`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
			},
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			var resp []byte
			query, err := urlquery.Marshal(tt.request)
			assert.Nil(s.T(), err)
			resp, err = s.client.DoStreamRequest(
				http.MethodGet,
				fmt.Sprintf(`/runs/search/metric?%s`, query),
				nil,
			)
			assert.Nil(s.T(), err)
			decodedData, err := encoding.Decode(bytes.NewBuffer(resp))
			assert.Nil(s.T(), err)

			decodedMetrics := []*models.LatestMetric{}
			for _, run := range runs {
				metricCount := 0
				for decodedData[fmt.Sprintf("%v.traces.%d.name", run.ID, metricCount)] != nil {
					epochsKey := fmt.Sprintf("%v.traces.%d.epochs.blob", run.ID, metricCount)
					itersKey := fmt.Sprintf("%v.traces.%d.iters.blob", run.ID, metricCount)
					nameKey := fmt.Sprintf("%v.traces.%d.name", run.ID, metricCount)
					timestampsKey := fmt.Sprintf("%v.traces.%d.timestamps.blob", run.ID, metricCount)
					valuesKey := fmt.Sprintf("%v.traces.%d.values.blob", run.ID, metricCount)

					m := models.LatestMetric{
						Key:       decodedData[nameKey].(string),
						Value:     decodedData[valuesKey].([]float64)[0],
						Timestamp: int64(decodedData[timestampsKey].([]float64)[0] * 1000),
						Step:      int64(decodedData[epochsKey].([]float64)[0]),
						IsNan:     false,
						RunID:     run.ID,
						LastIter:  int64(decodedData[itersKey].([]float64)[0]),
					}
					decodedMetrics = append(decodedMetrics, &m)
					metricCount++
				}
			}

			// Check if the received metrics match the expected ones
			assert.Equal(s.T(), tt.metrics, decodedMetrics)
		})
	}
}

func (s *SearchMetricsTestSuite) Test_Error() {
	defer func() {
		assert.Nil(s.T(), s.experimentFixtures.UnloadFixtures())
	}()
}
