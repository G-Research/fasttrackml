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
			name: "SearchMetricLastOperationGreater",
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
			name: "SearchMetricLastOperationGreaterOrEqual",
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
			name: "SearchMetricLastStepOperationGreater",
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
			name: "SearchMetricLastStepOperationGreaterOrEqual",
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
			name: "SearchMetricNameAndRunDurationOperationGreater",
			request: request.SearchMetricRequest{
				Query: `(metric.name == "TestMetric1" and run.duration > 0)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationGreaterOrEquals",
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
			name: "SearchMetricNameAndRunDurationOperationNotEqualsAndGreater",
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
			name: "SearchMetricNameAndRunDurationOperationNotEqualsAndGreaterOrEquals",
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
			name: "SearchMetricNameAndRunDurationOperationStartsWithAndGreater",
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
			name: "SearchMetricNameAndRunDurationOperationStartsWithAndGreaterOrEquals",
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
			name: "SearchMetricNameAndRunDurationOperationEndsWithAndGreater",
			request: request.SearchMetricRequest{
				Query: `(metric.name.endswith("Metric2") and run.duration > 0)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationEndsWithAndGreaterOrEquals",
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
		// Search Metric Name (equals operation) and Run CreatedAt
		{
			name: "SearchMetricNameEqualsAndRunCreatedAtOperationGreater",
			request: request.SearchMetricRequest{
				Query: `(metric.name == "TestMetric1" and run.created_at > 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameEqualsAndRunCreatedAtOperationGreaterOrEqual",
			request: request.SearchMetricRequest{
				Query: `(metric.name == "TestMetric1" and run.created_at >= 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameEqualsAndRunCreatedAtOperationLess",
			request: request.SearchMetricRequest{
				Query: `(metric.name == "TestMetric1" and run.created_at < 222222222)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
			},
		},
		{
			name: "SearchMetricNameEqualsAndRunCreatedAtOperationLessOrEqual",
			request: request.SearchMetricRequest{
				Query: `(metric.name == "TestMetric1" and run.created_at <= 222222222)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameEqualsAndRunCreatedAtOperationEquals",
			request: request.SearchMetricRequest{
				Query: `(metric.name == "TestMetric1" and run.created_at == 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricNameEqualsAndRunCreatedAtOperationNotEquals",
			request: request.SearchMetricRequest{
				Query: `(metric.name == "TestMetric1" and run.created_at != 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
			},
		},
		// Search Metric Name (not equals operation) and Run CreatedAt
		{
			name: "SearchMetricNameNotEqualsAndRunCreatedAtOperationGreater",
			request: request.SearchMetricRequest{
				Query: `(metric.name != "TestMetric1" and run.created_at > 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameNotEqualsAndRunCreatedAtOperationGreaterOrEqual",
			request: request.SearchMetricRequest{
				Query: `(metric.name != "TestMetric1" and run.created_at >= 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameNotEqualsAndRunCreatedAtOperationLess",
			request: request.SearchMetricRequest{
				Query: `(metric.name != "TestMetric1" and run.created_at < 222222222)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricNameNotEqualsAndRunCreatedAtOperationLessOrEqual",
			request: request.SearchMetricRequest{
				Query: `(metric.name != "TestMetric1" and run.created_at <= 222222222)`,
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
			name: "SearchMetricNameNotEqualsAndRunCreatedAtOperationEquals",
			request: request.SearchMetricRequest{
				Query: `(metric.name != "TestMetric1" and run.created_at == 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricNameNotEqualsAndRunCreatedAtOperationNotEquals",
			request: request.SearchMetricRequest{
				Query: `(metric.name != "TestMetric1" and run.created_at != 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
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

		// Search Metric Last (equal operation) and Run Name
		{
			name: "SearchMetricLastAndRunNameOperationRegexpMatchFunction",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last == 1.6) and re.match("TestRun3", run.name))`,
			},

			metrics: []*models.LatestMetric{
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationRegexpSearchFunction",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last == 1.6) and re.search("TestRun3", run.name))`,
			},

			metrics: []*models.LatestMetric{
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last == 1.6) and run.name == "TestRun3")`,
			},

			metrics: []*models.LatestMetric{
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationNotEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last == 1.6) and run.name != "TestRun3")`,
			},

			metrics: []*models.LatestMetric{},
		},
		{
			name: "SearchMetricLastAndRunNameOperationStartsWith",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last == 1.6) and run.name.startswith("Test"))`,
			},

			metrics: []*models.LatestMetric{
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationEndsWith",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last == 1.6) and run.name.endswith("Run2"))`,
			},

			metrics: []*models.LatestMetric{},
		},

		// Search Metric Last (not equal operation) and Run Name
		{
			name: "SearchMetricLastAndRunNameOperationNotEqualsAndRegexpMatchFunction",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last != 1.6) and re.match("TestRun3", run.name))`,
			},

			metrics: []*models.LatestMetric{
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationNotEqualsAndRegexpSearchFunction",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last != 1.6) and re.search("TestRun3", run.name))`,
			},

			metrics: []*models.LatestMetric{
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationNotEqualsAndEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.name == "TestRun3")`,
			},

			metrics: []*models.LatestMetric{
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationNotEqualsAndNotEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.name != "TestRun3")`,
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
			name: "SearchMetricLastAndRunNameOperationNotEqualsAndStartsWith",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.name.startswith("Test"))`,
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
			name: "SearchMetricLastAndRunNameOperationNotEqualsAndEndsWith",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.name.endswith("Run3"))`,
			},

			metrics: []*models.LatestMetric{
				metric1Run3,
			},
		},

		// Search Metric Last (greater operation) and Run Name
		{
			name: "SearchMetricLastAndRunNameOperationGreaterAndRegexpMatchFunction",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last > 1.6) and re.match("TestRun1", run.name))`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationGreaterAndRegexpSearchFunction",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last > 1.6) and re.search("TestRun1", run.name))`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationGreaterAndEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.name == "TestRun1")`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationGreaterAndNotEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.name != "TestRun1")`,
			},

			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationGreaterAndStartsWith",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.name.startswith("Test"))`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationGreaterAndEndsWith",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.name.endswith("Run2"))`,
			},

			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
			},
		},

		// Search Metric Last (greater or equals operation) and Run Name
		{
			name: "SearchMetricLastAndRunNameOperationGreaterOrEqualsAndRegexpMatchFunction",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and re.match("TestRun1", run.name))`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationGreaterOrEqualsAndRegexpSearchFunction",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and re.search("TestRun1", run.name))`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationGreaterOrEqualsAndEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.name == "TestRun1")`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationGreaterOrEqualsAndNotEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.name != "TestRun1")`,
			},

			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationGreaterOrEqualsAndStartsWith",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.name.startswith("Test"))`,
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
			name: "SearchMetricLastAndRunNameOperationGreaterOrEqualsAndEndsWith",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.name.endswith("Run2"))`,
			},

			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
			},
		},
		// Search Metric Last (less operation) and Run Name
		{
			name: "SearchMetricLastAndRunNameOperationLessAndRegexpMatchFunction",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last < 1.6) and re.match("TestRun1", run.name))`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationLessAndRegexpSearchFunction",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last < 1.6) and re.search("TestRun1", run.name))`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationLessAndEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.name == "TestRun1")`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationLessAndNotEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.name != "TestRun1")`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationLessAndStartsWith",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.name.startswith("Test"))`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationLessAndEndsWith",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.name.endswith("Run2"))`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
			},
		},
		// Search Metric Last (less or equals operation) and Run Name
		{
			name: "SearchMetricLastAndRunNameOperationLessOrEqualsAndRegexpMatchFunction",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and re.match("TestRun1", run.name))`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationLessOrEqualsAndRegexpSearchFunction",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and re.search("TestRun1", run.name))`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationLessOrEqualsAndEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.name == "TestRun1")`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationLessOrEqualsAndNotEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.name != "TestRun1")`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationLessOrEqualsAndStartsWith",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.name.startswith("Test"))`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationLessOrEqualsAndEndsWith",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.name.endswith("Run2"))`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
			},
		},
		// Search Metric Last (equal operation) and Run Duration
		{
			name: "SearchMetricLastAndRunDurationOperationEqualsAndEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last == 1.6) and run.duration == 222222)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationEqualsAndNotEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last == 1.6) and run.duration != 222222)`,
			},

			metrics: []*models.LatestMetric{},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationEqualsAndGreater",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last == 1.6) and run.duration > 0)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationEqualsAndGreaterOrEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last == 1.6) and run.duration >= 0)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationEqualsAndLess",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last == 1.6) and run.duration < 222222)`,
			},

			metrics: []*models.LatestMetric{},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationEqualsAndLessOrEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last == 1.6) and run.duration <= 222222)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run3,
			},
		},
		// Search Metric Last (not equal operation) and Run Duration
		{
			name: "SearchMetricLastAndRunDurationOperationNotEqualsAndEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.duration == 222222)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationNotEqualsAndNotEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.duration != 222222)`,
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
			name: "SearchMetricLastAndRunDurationOperationNotEqualsAndGreater",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.duration > 0)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationNotEqualsAndGreaterOrEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.duration >= 0)`,
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
			name: "SearchMetricLastAndRunDurationOperationNotEqualsAndLess",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.duration < 222222)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationNotEqualsAndLessOrEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.duration <= 222222)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run3,
			},
		},

		// Search Metric Last (greater operation) and Run Duration
		{
			name: "SearchMetricLastAndRunDurationOperationGreaterAndEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.duration == 222222)`,
			},

			metrics: []*models.LatestMetric{},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationGreaterAndNotEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.duration != 222222)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationGreaterAndGreater",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.duration > 0)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationGreaterAndGreaterOrEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.duration >= 0)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationGreaterAndLess",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.duration < 333333)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationGreaterAndLessOrEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.duration <= 333333)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
			},
		},
		// Search Metric Last (greater or equals operation) and Run Duration
		{
			name: "SearchMetricLastAndRunDurationOperationGreaterOrEqualsAndEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.duration == 222222)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationGreaterOrEqualsAndNotEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.duration != 222222)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationGreaterOrEqualsAndGreater",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.duration > 0)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationGreaterOrEqualsAndGreaterOrEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.duration >= 0)`,
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
			name: "SearchMetricLastAndRunDurationOperationGreaterOrEqualsAndLess",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.duration < 333333)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationGreaterOrEqualsAndLessOrEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.duration <= 333333)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		// Search Metric Last (less operation) and Run Duration
		{
			name: "SearchMetricLastAndRunDurationOperationLessAndEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.duration == 222222)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationLessAndNotEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.duration != 222222)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationLessAndGreater",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.duration > 0)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationLessAndGreaterOrEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.duration >= 0)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationLessAndLess",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.duration < 333333)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationLessAndLessOrEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.duration <= 333333)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
			},
		},
		// Search Metric Last (less or equals operation) and Run Duration
		{
			name: "SearchMetricLastAndRunDurationOperationLessOrEqualsAndEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.duration == 222222)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationLessOrEqualsAndNotEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.duration != 222222)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationLessOrEqualsAndGreater",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.duration > 0)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationLessOrEqualsAndGreaterOrEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.duration >= 0)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationLessOrEqualsAndLess",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.duration < 333333)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationLessOrEqualsAndLessOrEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.duration <= 333333)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		// Search Metric Last (equal operation) and Run Hash
		{
			name: "SearchMetricLastAndRunHashOperationEqualsAndEquals",
			request: request.SearchMetricRequest{
				Query: fmt.Sprintf(`(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last == 1.1) and run.hash == "%s")`, run1.ID),
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunHashOperationEqualsAndNotEquals",
			request: request.SearchMetricRequest{
				Query: fmt.Sprintf(`(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last == 1.1) and run.hash != "%s")`, run1.ID),
			},

			metrics: []*models.LatestMetric{},
		},
		// Search Metric Last (not equal operation) and Run Hash
		{
			name: "SearchMetricLastAndRunHashOperationNotEqualsAndEquals",
			request: request.SearchMetricRequest{
				Query: fmt.Sprintf(`(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last != 1.1) and run.hash == "%s")`, run1.ID),
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunHashOperationNotEqualsAndNotEquals",
			request: request.SearchMetricRequest{
				Query: fmt.Sprintf(`(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last != 1.1) and run.hash != "%s")`, run1.ID),
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		// Search Metric Last (greater operation) and Run Hash
		{
			name: "SearchMetricLastAndRunHashOperationGreaterAndEquals",
			request: request.SearchMetricRequest{
				Query: fmt.Sprintf(`(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last > 1.1) and run.hash == "%s")`, run1.ID),
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunHashOperationGreaterAndNotEquals",
			request: request.SearchMetricRequest{
				Query: fmt.Sprintf(`(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last > 1.1) and run.hash != "%s")`, run1.ID),
			},

			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		// Search Metric Last (greater or equals operation) and Run Hash
		{
			name: "SearchMetricLastAndRunHashOperationGreaterOrEqualsAndEquals",
			request: request.SearchMetricRequest{
				Query: fmt.Sprintf(`(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last >= 1.1) and run.hash == "%s")`, run1.ID),
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunHashOperationGreaterOrEqualsAndNotEquals",
			request: request.SearchMetricRequest{
				Query: fmt.Sprintf(`(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last >= 1.1) and run.hash != "%s")`, run1.ID),
			},

			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		// Search Metric Last (less operation) and Run Hash
		{
			name: "SearchMetricLastAndRunHashOperationLessAndEquals",
			request: request.SearchMetricRequest{
				Query: fmt.Sprintf(`(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last < 3.1) and run.hash == "%s")`, run1.ID),
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunHashOperationLessAndNotEquals",
			request: request.SearchMetricRequest{
				Query: fmt.Sprintf(`(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last < 3.1) and run.hash != "%s")`, run1.ID),
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		// Search Metric Last (less or equals operation) and Run Hash
		{
			name: "SearchMetricLastAndRunHashOperationLessOrEqualsAndEquals",
			request: request.SearchMetricRequest{
				Query: fmt.Sprintf(`(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last <= 3.1) and run.hash == "%s")`, run1.ID),
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunHashOperationLessOrEqualsAndNotEquals",
			request: request.SearchMetricRequest{
				Query: fmt.Sprintf(`(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last <= 3.1) and run.hash != "%s")`, run1.ID),
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		// Search Metric Last (equal operation) and Run FinalizedAt
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationEqualsAndEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last == 1.1) and run.finalized_at == 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationEqualsAndNotEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last == 1.6) and run.finalized_at != 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationEqualsAndGreater",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last == 1.1) and run.finalized_at > 123456789)`,
			},

			metrics: []*models.LatestMetric{},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationEqualsAndGreaterOrEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last == 1.1) and run.finalized_at >= 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationEqualsAndLess",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last == 1.6) and run.finalized_at < 444444444)`,
			},

			metrics: []*models.LatestMetric{},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationEqualsAndLessOrEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last == 1.6) and run.finalized_at <= 444444444)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run3,
			},
		},
		// Search Metric Last (not equal operation) and Run FinalizedAt
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationNotEqualsAndEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.finalized_at == 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationNotEqualsAndNotEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.finalized_at != 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationNotEqualsAndGreater",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.finalized_at > 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationNotEqualsAndGreaterOrEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.finalized_at >= 123456789)`,
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
			name: "SearchMetricLastAndRunFinalizedAtOperationNotEqualsAndLess",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.finalized_at < 444444444)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationNotEqualsAndLessOrEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.finalized_at <= 444444444)`,
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

		// Search Metric Last (greater operation) and Run FinalizedAt
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationGreaterAndEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.finalized_at == 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationGreaterAndNotEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.finalized_at != 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationGreaterAndGreater",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.finalized_at > 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationGreaterAndGreaterOrEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.finalized_at >= 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationGreaterAndLess",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.finalized_at < 444444444)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationGreaterAndLessOrEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.finalized_at <= 444444444)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
			},
		},
		// Search Metric Last (greater or equals operation) and Run FinalizedAt
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationGreaterOrEqualsAndEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.finalized_at == 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationGreaterOrEqualsAndNotEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.finalized_at != 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationGreaterOrEqualsAndGreater",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.finalized_at > 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationGreaterOrEqualsAndGreaterOrEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.finalized_at >= 123456789)`,
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
			name: "SearchMetricLastAndRunFinalizedAtOperationGreaterOrEqualsAndLess",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.finalized_at < 444444444)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationGreaterOrEqualsAndLessOrEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.finalized_at <= 444444444)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		// Search Metric Last (less operation) and Run FinalizedAt
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationLessAndEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.finalized_at == 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationLessAndNotEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.finalized_at != 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationLessAndGreater",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.finalized_at > 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationLessAndGreaterOrEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.finalized_at >= 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationLessAndLess",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.finalized_at < 444444444)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationLessAndLessOrEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.finalized_at <= 444444444)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
			},
		},
		// Search Metric Last (less or equals operation) and Run FinalizedAt
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationLessOrEqualsAndEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.finalized_at == 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationLessOrEqualsAndNotEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.finalized_at != 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationLessOrEqualsAndGreater",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.finalized_at > 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationLessOrEqualsAndGreaterOrEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.finalized_at >= 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationLessOrEqualsAndLess",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.finalized_at < 444444444)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationLessOrEqualsAndLessOrEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.finalized_at <= 444444444)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		// Search Metric Last (equal operation) and Run CreatedAt
		{
			name: "SearchMetricLastAndRunCreatedAtOperationEqualsAndEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last == 1.1) and run.created_at == 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationEqualsAndNotEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last == 1.6) and run.created_at != 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationEqualsAndGreater",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last == 1.1) and run.created_at > 123456789)`,
			},

			metrics: []*models.LatestMetric{},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationEqualsAndGreaterOrEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last == 1.1) and run.created_at >= 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationEqualsAndLess",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last == 1.6) and run.created_at < 222222222)`,
			},

			metrics: []*models.LatestMetric{},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationEqualsAndLessOrEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last == 1.6) and run.created_at <= 222222222)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run3,
			},
		},
		// Search Metric Last (not equal operation) and Run CreatedAt
		{
			name: "SearchMetricLastAndRunCreatedAtOperationNotEqualsAndEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.created_at == 111111111)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationNotEqualsAndNotEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.created_at != 111111111)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationNotEqualsAndGreater",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.created_at > 111111111)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationNotEqualsAndGreaterOrEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.created_at >= 111111111)`,
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
			name: "SearchMetricLastAndRunCreatedAtOperationNotEqualsAndLess",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.created_at < 222222222)`,
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
			name: "SearchMetricLastAndRunCreatedAtOperationNotEqualsAndLessOrEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.created_at <= 222222222)`,
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

		// Search Metric Last (greater operation) and Run CreatedAt
		{
			name: "SearchMetricLastAndRunCreatedAtOperationGreaterAndEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.created_at == 111111111)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationGreaterAndNotEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.created_at != 111111111)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationGreaterAndGreater",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.created_at > 111111111)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationGreaterAndGreaterOrEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.created_at >= 111111111)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationGreaterAndLess",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.created_at < 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationGreaterAndLessOrEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.created_at <= 123456789)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
			},
		},
		// Search Metric Last (greater or equals operation) and Run CreatedAt
		{
			name: "SearchMetricLastAndRunCreatedAtOperationGreaterOrEqualsAndEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.created_at == 111111111)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationGreaterOrEqualsAndNotEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.created_at != 111111111)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationGreaterOrEqualsAndGreater",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.created_at > 111111111)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationGreaterOrEqualsAndGreaterOrEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.created_at >= 111111111)`,
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
			name: "SearchMetricLastAndRunCreatedAtOperationGreaterOrEqualsAndLess",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.created_at < 222222222)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationGreaterOrEqualsAndLessOrEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.created_at <= 222222222)`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		// Search Metric Last (less operation) and Run CreatedAt
		{
			name: "SearchMetricLastAndRunCreatedAtOperationLessAndEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.created_at == 111111111)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationLessAndNotEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.created_at != 111111111)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationLessAndGreater",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.created_at > 111111111)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationLessAndGreaterOrEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.created_at >= 111111111)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationLessAndLess",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.created_at < 222222222)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationLessAndLessOrEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.created_at <= 222222222)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
			},
		},
		// Search Metric Last (less or equals operation) and Run CreatedAt
		{
			name: "SearchMetricLastAndRunCreatedAtOperationLessOrEqualsAndEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.created_at == 111111111)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationLessOrEqualsAndNotEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.created_at != 111111111)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationLessOrEqualsAndGreater",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.created_at > 111111111)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationLessOrEqualsAndGreaterOrEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.created_at >= 111111111)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationLessOrEqualsAndLess",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.created_at < 222222222)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationLessOrEqualsAndLessOrEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.created_at <= 222222222)`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		// Search Metric Last Step (equal operation) and Run Name
		{
			name: "SearchMetricLastStepAndRunNameOperationRegexpMatchFunction",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step == 2) and re.match("TestRun2", run.name))`,
			},

			metrics: []*models.LatestMetric{
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationRegexpSearchFunction",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step == 2) and re.search("TestRun2", run.name))`,
			},

			metrics: []*models.LatestMetric{
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step == 2) and run.name == "TestRun2")`,
			},

			metrics: []*models.LatestMetric{
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationNotEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step == 2) and run.name != "TestRun2")`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationStartsWith",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step == 2) and run.name.startswith("Test"))`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationEndsWith",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step == 2) and run.name.endswith("Run2"))`,
			},

			metrics: []*models.LatestMetric{
				metric2Run2,
			},
		},
		// Search Metric Last Step (not equal operation) and Run Name
		{
			name: "SearchMetricLastStepAndRunNameOperationNotEqualsAndRegexpMatchFunction",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step != 2) and re.match("TestRun2", run.name))`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationNotEqualsAndRegexpSearchFunction",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step != 2) and re.search("TestRun2", run.name))`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationNotEqualsAndEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step != 2) and run.name == "TestRun2")`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationNotEqualsAndNotEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step != 2) and run.name != "TestRun2")`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric3Run1,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationNotEqualsAndStartsWith",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step != 2) and run.name.startswith("Test"))`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric3Run1,
				metric1Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationNotEqualsAndEndsWith",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step != 2) and run.name.endswith("Run2"))`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric3Run2,
			},
		},
		// Search Metric Last Step (greater operation) and Run Name
		{
			name: "SearchMetricLastStepAndRunNameOperationGreaterAndRegexpMatchFunction",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step > 2) and re.match("TestRun2", run.name))`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationGreaterAndRegexpSearchFunction",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step > 2) and re.search("TestRun2", run.name))`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationGreaterAndEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step > 2) and run.name == "TestRun2")`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationGreaterAndNotEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step > 2) and run.name != "TestRun2")`,
			},

			metrics: []*models.LatestMetric{
				metric3Run1,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationGreaterAndStartsWith",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step > 2) and run.name.startswith("Test"))`,
			},

			metrics: []*models.LatestMetric{
				metric3Run1,
				metric1Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationGreaterAndEndsWith",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step > 2) and run.name.endswith("Run2"))`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric3Run2,
			},
		},
		// Search Metric Last Step (greater or equal operation) and Run Name
		{
			name: "SearchMetricLastStepAndRunNameOperationGreaterOrEqualsAndRegexpMatchFunction",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step >= 2) and re.match("TestRun2", run.name))`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationGreaterOrEqualsAndRegexpSearchFunction",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step >= 2) and re.search("TestRun2", run.name))`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationGreaterOrEqualsAndEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step >= 2) and run.name == "TestRun2")`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationGreaterOrEqualsAndNotEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step >= 2) and run.name != "TestRun2")`,
			},

			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationGreaterOrEqualsAndStartsWith",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step >= 2) and run.name.startswith("Test"))`,
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
			name: "SearchMetricLastStepAndRunNameOperationGreaterOrEqualsAndEndsWith",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step >= 2) and run.name.endswith("Run2"))`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
			},
		},
		// Search Metric Last Step (less operation) and Run Name
		{
			name: "SearchMetricLastStepAndRunNameOperationLessAndRegexpMatchFunction",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step < 3) and re.match("TestRun2", run.name))`,
			},

			metrics: []*models.LatestMetric{
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationLessAndRegexpSearchFunction",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step < 3) and re.search("TestRun2", run.name))`,
			},

			metrics: []*models.LatestMetric{
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationLessAndEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step < 3) and run.name == "TestRun2")`,
			},

			metrics: []*models.LatestMetric{
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationLessAndNotEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step < 3) and run.name != "TestRun2")`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationLessAndStartsWith",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step < 3) and run.name.startswith("Test"))`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationLessAndEndsWith",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step < 3) and run.name.endswith("Run2"))`,
			},

			metrics: []*models.LatestMetric{
				metric2Run2,
			},
		},
		// Search Metric Last Step (less or equal operation) and Run Name
		{
			name: "SearchMetricLastStepAndRunNameOperationLessOrEqualsAndRegexpMatchFunction",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step <= 3) and re.match("TestRun2", run.name))`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationLessOrEqualsAndRegexpSearchFunction",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step <= 3) and re.search("TestRun2", run.name))`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationLessOrEqualsAndEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step <= 3) and run.name == "TestRun2")`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationLessOrEqualsAndNotEquals",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step <= 3) and run.name != "TestRun2")`,
			},

			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationLessrOrEqualsAndStartsWith",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step <= 3) and run.name.startswith("Test"))`,
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
			name: "SearchMetricLastStepAndRunNameOperationLessOrEqualsAndEndsWith",
			request: request.SearchMetricRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or (metric.name == "TestMetric3")) and (metric.last_step <= 3) and run.name.endswith("Run2"))`,
			},

			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
			},
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
			name: "SearchMetricLastRunNameOperationGreaterAndEqual",
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
			name: "SearchMetricLastRunNameOperationGreaterOrEqualAndEqual",
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
			name: "SearchMetricLastRunNameOperationGreaterAndNotEqual",
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
			name: "SearchMetricLastRunNameOperationGreaterOrEqualAndNotEqual",
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
