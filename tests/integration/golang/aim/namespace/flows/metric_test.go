//go:build integration

package flows

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/encoding"
	"github.com/G-Research/fasttrackml/pkg/api/aim/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type MetricFlowTestSuite struct {
	helpers.BaseTestSuite
}

// TestMetricTestSuite tests the full `metric` flow connected to namespace functionality.
// Flow contains next endpoints:
// - `GET /runs/search/metric`
// - `GET /runs/search/metric/align`
func TestMetricTestSuite(t *testing.T) {
	suite.Run(t, new(MetricFlowTestSuite))
}

func (s *MetricFlowTestSuite) TearDownTest() {
	s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
}

func (s *MetricFlowTestSuite) Test_Ok() {
	tests := []struct {
		name           string
		setup          func() (*models.Namespace, *models.Namespace)
		namespace1Code string
		namespace2Code string
	}{
		{
			name: "TestCustomNamespaces",
			setup: func() (*models.Namespace, *models.Namespace) {
				return &models.Namespace{
						Code:                "namespace-1",
						DefaultExperimentID: common.GetPointer(int32(0)),
					}, &models.Namespace{
						Code:                "namespace-2",
						DefaultExperimentID: common.GetPointer(int32(0)),
					}
			},
			namespace1Code: "namespace-1",
			namespace2Code: "namespace-2",
		},
		{
			name: "TestExplicitDefaultAndCustomNamespaces",
			setup: func() (*models.Namespace, *models.Namespace) {
				return &models.Namespace{
						Code:                "default",
						DefaultExperimentID: common.GetPointer(int32(0)),
					}, &models.Namespace{
						Code:                "namespace-1",
						DefaultExperimentID: common.GetPointer(int32(0)),
					}
			},
			namespace1Code: "default",
			namespace2Code: "namespace-1",
		},
		{
			name: "TestImplicitDefaultAndCustomNamespaces",
			setup: func() (*models.Namespace, *models.Namespace) {
				return &models.Namespace{
						Code:                "default",
						DefaultExperimentID: common.GetPointer(int32(0)),
					}, &models.Namespace{
						Code:                "namespace-1",
						DefaultExperimentID: common.GetPointer(int32(0)),
					}
			},
			namespace1Code: "",
			namespace2Code: "namespace-1",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			defer s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())

			// 1. setup data under the test.
			namespace1, namespace2 := tt.setup()
			namespace1, err := s.NamespaceFixtures.CreateNamespace(context.Background(), namespace1)
			s.Require().Nil(err)
			namespace2, err = s.NamespaceFixtures.CreateNamespace(context.Background(), namespace2)
			s.Require().Nil(err)

			experiment1, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
				Name:             "Experiment1",
				ArtifactLocation: "/artifact/location",
				LifecycleStage:   models.LifecycleStageActive,
				NamespaceID:      namespace1.ID,
			})
			s.Require().Nil(err)

			// create different test runs and attach tags, metrics, params, etc.
			run1, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
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
				ExperimentID:   *experiment1.ID,
				ArtifactURI:    "artifact_uri1",
				LifecycleStage: models.LifecycleStageActive,
			})
			s.Require().Nil(err)
			_, err = s.MetricFixtures.CreateMetric(context.Background(), &models.Metric{
				Key:       "TestMetric1",
				Value:     1.1,
				Timestamp: 123456789,
				Step:      1,
				IsNan:     false,
				RunID:     run1.ID,
				Iter:      1,
			})
			s.Require().Nil(err)
			metric1Run1, err := s.MetricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
				Key:       "TestMetric1",
				Value:     1.1,
				Timestamp: 123456789,
				Step:      1,
				IsNan:     false,
				RunID:     run1.ID,
				LastIter:  1,
			})
			s.Require().Nil(err)
			_, err = s.ParamFixtures.CreateParam(context.Background(), &models.Param{
				Key:   "param1",
				Value: "value1",
				RunID: run1.ID,
			})
			s.Require().Nil(err)
			_, err = s.TagFixtures.CreateTag(context.Background(), &models.Tag{
				Key:   "mlflow.runName",
				Value: "TestRunTag1",
				RunID: run1.ID,
			})
			s.Require().Nil(err)

			experiment2, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
				Name:             "Experiment2",
				ArtifactLocation: "/artifact/location",
				LifecycleStage:   models.LifecycleStageActive,
				NamespaceID:      namespace2.ID,
			})
			s.Require().Nil(err)

			run2, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
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
				ExperimentID:   *experiment2.ID,
				ArtifactURI:    "artifact_uri2",
				LifecycleStage: models.LifecycleStageActive,
			})
			s.Require().Nil(err)
			_, err = s.MetricFixtures.CreateMetric(context.Background(), &models.Metric{
				Key:       "TestMetric2",
				Value:     0.5,
				Timestamp: 111111111,
				Step:      4,
				IsNan:     false,
				RunID:     run2.ID,
				Iter:      3,
			})
			s.Require().Nil(err)
			metric1Run2, err := s.MetricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
				Key:       "TestMetric2",
				Value:     0.5,
				Timestamp: 111111111,
				Step:      4,
				IsNan:     false,
				RunID:     run2.ID,
				LastIter:  3,
			})
			s.Require().Nil(err)
			_, err = s.ParamFixtures.CreateParam(context.Background(), &models.Param{
				Key:   "param2",
				Value: "value2",
				RunID: run2.ID,
			})
			s.Require().Nil(err)
			_, err = s.TagFixtures.CreateTag(context.Background(), &models.Tag{
				Key:   "mlflow.runName",
				Value: "TestRunTag2",
				RunID: run2.ID,
			})
			s.Require().Nil(err)

			// 2. run actual flow test over the test data.
			s.testRunFlow(
				tt.namespace1Code, tt.namespace2Code, run1, run2, metric1Run1, metric1Run2,
			)
		})
	}
}

func (s *MetricFlowTestSuite) testRunFlow(
	namespace1Code, namespace2Code string, run1, run2 *models.Run, metric1Run1, metric1Run2 *models.LatestMetric,
) {
	// test `GET /runs/search/metric` endpoint.
	s.searchMetricsAndCompare(namespace1Code, request.SearchMetricsRequest{
		Query: `(metric.name == "TestMetric1")`,
	}, []*models.Run{run1}, []*models.LatestMetric{
		metric1Run1,
	})
	s.searchMetricsAndCompare(namespace2Code, request.SearchMetricsRequest{
		Query: `(metric.name == "TestMetric2")`,
	}, []*models.Run{run2}, []*models.LatestMetric{
		metric1Run2,
	})

	// test `GET /runs/search/metric/align` endpoint.
	s.searchAlignedMetricsAndCompare(namespace1Code, &request.GetAlignedMetricRequest{
		AlignBy: "TestMetric1",
		Runs: []request.AlignedMetricRunRequest{
			{
				ID: run1.ID,
				Traces: []request.AlignedMetricTraceRequest{
					{
						Name:  "TestMetric1",
						Slice: []int{0, 0, 500},
					},
				},
			},
		},
	}, []*models.Run{run1}, []float64{metric1Run1.Value})
	s.searchAlignedMetricsAndCompare(namespace2Code, &request.GetAlignedMetricRequest{
		AlignBy: "TestMetric2",
		Runs: []request.AlignedMetricRunRequest{
			{
				ID: run1.ID,
				Traces: []request.AlignedMetricTraceRequest{
					{
						Name:  "TestMetric2",
						Slice: []int{0, 0, 500},
					},
				},
			},
		},
	}, []*models.Run{run2}, []float64{metric1Run2.Value})
}

func (s *MetricFlowTestSuite) searchMetricsAndCompare(
	namespace string,
	request request.SearchMetricsRequest,
	expectedRuns []*models.Run,
	expectedMetrics []*models.LatestMetric,
) {
	resp := new(bytes.Buffer)
	s.Require().Nil(
		s.AIMClient().WithNamespace(
			namespace,
		).WithQuery(
			request,
		).WithResponseType(
			helpers.ResponseTypeBuffer,
		).WithResponse(
			resp,
		).DoRequest("/runs/search/metric"),
	)
	decodedData, err := encoding.Decode(resp)
	s.Require().Nil(err)

	var decodedMetrics []*models.LatestMetric
	for _, run := range expectedRuns {
		metricCount := 0
		for decodedData[fmt.Sprintf("%v.traces.%d.name", run.ID, metricCount)] != nil {
			prefix := fmt.Sprintf("%v.traces.%d", run.ID, metricCount)
			epochsKey := prefix + ".epochs.blob"
			itersKey := prefix + ".iters.blob"
			nameKey := prefix + ".name"
			timestampsKey := prefix + ".timestamps.blob"
			valuesKey := prefix + ".values.blob"

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
	s.Equal(expectedMetrics, decodedMetrics)
}

func (s *MetricFlowTestSuite) searchAlignedMetricsAndCompare(
	namespace string, request *request.GetAlignedMetricRequest, expectedRuns []*models.Run, expectedResponse []float64,
) {
	resp := new(bytes.Buffer)
	s.Require().Nil(s.AIMClient().WithMethod(
		http.MethodPost,
	).WithNamespace(
		namespace,
	).WithRequest(
		request,
	).WithResponse(
		resp,
	).WithResponseType(
		helpers.ResponseTypeBuffer,
	).DoRequest(
		"/runs/search/metric/align",
	))

	decodedData, err := encoding.Decode(resp)
	s.Require().Nil(err)

	xValues := make(map[int][]float64)

	for _, run := range expectedRuns {
		metricCount := 0
		for decodedData[fmt.Sprintf("%v.%d.name", run.ID, metricCount)] != nil {
			valueKey := fmt.Sprintf("%v.%d.x_axis_values.blob", run.ID, metricCount)
			xValues[metricCount] = append(xValues[metricCount], decodedData[valueKey].([]float64)[0])

			metricCount++
		}
	}

	// Check if the received values for each metric match the expected ones
	for _, metricValues := range xValues {
		s.Equal(expectedResponse, metricValues)
	}
}
