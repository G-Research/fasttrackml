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
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/encoding"
	"github.com/G-Research/fasttrackml/pkg/api/aim/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type SearchAlignedMetricsTestSuite struct {
	helpers.BaseTestSuite
}

func TestSearchAlignedMetricsTestSuite(t *testing.T) {
	suite.Run(t, new(SearchAlignedMetricsTestSuite))
}

func (s *SearchAlignedMetricsTestSuite) Test_Ok() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()
	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	s.Require().Nil(err)

	// create test experiments.
	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		LifecycleStage: models.LifecycleStageActive,
		NamespaceID:    namespace.ID,
	})
	s.Require().Nil(err)

	experiment1, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		LifecycleStage: models.LifecycleStageActive,
		NamespaceID:    namespace.ID,
	})
	s.Require().Nil(err)

	// create different test runs and attach, metrics.
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
		ExperimentID:   *experiment.ID,
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
	_, err = s.MetricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "TestMetric2",
		Value:     2.1,
		Timestamp: 123456789,
		Step:      5,
		IsNan:     false,
		RunID:     run1.ID,
		Iter:      1,
	})
	s.Require().Nil(err)
	metric2Run1, err := s.MetricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       "TestMetric2",
		Value:     2.1,
		Timestamp: 123456789,
		Step:      5,
		IsNan:     false,
		RunID:     run1.ID,
		LastIter:  1,
	})
	s.Require().Nil(err)
	_, err = s.MetricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "TestMetric3",
		Value:     3.1,
		Timestamp: 123456789,
		Step:      10,
		IsNan:     false,
		RunID:     run1.ID,
		Iter:      1,
	})
	s.Require().Nil(err)
	metric3Run1, err := s.MetricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       "TestMetric3",
		Value:     3.1,
		Timestamp: 123456789,
		Step:      10,
		IsNan:     false,
		RunID:     run1.ID,
		LastIter:  1,
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
		ExperimentID:   *experiment.ID,
		ArtifactURI:    "artifact_uri2",
		LifecycleStage: models.LifecycleStageActive,
	})
	s.Require().Nil(err)
	_, err = s.MetricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "TestMetric1",
		Value:     0.5,
		Timestamp: 111111111,
		Step:      4,
		IsNan:     false,
		RunID:     run2.ID,
		Iter:      1,
	})
	s.Require().Nil(err)
	metric1Run2, err := s.MetricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       "TestMetric1",
		Value:     0.5,
		Timestamp: 111111111,
		Step:      4,
		IsNan:     false,
		RunID:     run2.ID,
		LastIter:  1,
	})
	s.Require().Nil(err)
	_, err = s.MetricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "TestMetric2",
		Value:     2.1,
		Timestamp: 222222222,
		Step:      5,
		IsNan:     false,
		RunID:     run2.ID,
		Iter:      1,
	})
	s.Require().Nil(err)
	metric2Run2, err := s.MetricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       "TestMetric2",
		Value:     2.1,
		Timestamp: 222222222,
		Step:      5,
		IsNan:     false,
		RunID:     run2.ID,
		LastIter:  1,
	})
	s.Require().Nil(err)
	_, err = s.MetricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "TestMetric3",
		Value:     3.1,
		Timestamp: 333333333,
		Step:      10,
		IsNan:     false,
		RunID:     run2.ID,
		Iter:      1,
	})
	s.Require().Nil(err)
	metric3Run2, err := s.MetricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       "TestMetric3",
		Value:     3.1,
		Timestamp: 333333333,
		Step:      10,
		IsNan:     false,
		RunID:     run2.ID,
		LastIter:  1,
	})
	s.Require().Nil(err)
	run3, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
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
		ExperimentID:   *experiment1.ID,
		ArtifactURI:    "artifact_uri3",
		LifecycleStage: models.LifecycleStageActive,
	})
	s.Require().Nil(err)
	_, err = s.MetricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "TestMetric1",
		Value:     1.2,
		Timestamp: 1511111111,
		Step:      6,
		IsNan:     false,
		RunID:     run3.ID,
		Iter:      1,
	})
	s.Require().Nil(err)
	metric1Run3, err := s.MetricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       "TestMetric1",
		Value:     1.2,
		Timestamp: 1511111111,
		Step:      6,
		IsNan:     false,
		RunID:     run3.ID,
		LastIter:  1,
	})
	s.Require().Nil(err)
	_, err = s.MetricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "TestMetric2",
		Value:     1.6,
		Timestamp: 2522222222,
		Step:      1,
		IsNan:     false,
		RunID:     run3.ID,
		Iter:      1,
	})
	s.Require().Nil(err)
	metric2Run3, err := s.MetricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       "TestMetric2",
		Value:     1.6,
		Timestamp: 2522222222,
		Step:      2,
		IsNan:     false,
		RunID:     run3.ID,
		LastIter:  1,
	})
	s.Require().Nil(err)
	_, err = s.MetricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "TestMetric3",
		Value:     2.6,
		Timestamp: 2522222222,
		Step:      2,
		IsNan:     false,
		RunID:     run3.ID,
		Iter:      1,
	})
	s.Require().Nil(err)
	metric3Run3, err := s.MetricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       "TestMetric3",
		Value:     2.6,
		Timestamp: 2522222222,
		Step:      2,
		IsNan:     false,
		RunID:     run3.ID,
		LastIter:  1,
	})
	s.Require().Nil(err)

	runs := []*models.Run{run1, run2, run3}

	tests := []struct {
		name     string
		request  *request.GetAlignedMetricRequest
		response []float64
	}{
		{
			name: "TestSearchAlignedByMetric1",
			request: &request.GetAlignedMetricRequest{
				AlignBy: "TestMetric1",
				Runs: []request.AlignedMetricRunRequest{
					{
						ID: run1.ID,
						Traces: []request.AlignedMetricTraceRequest{
							{
								Name:  "TestMetric1",
								Slice: []int{0, 0, 500},
							},
							{
								Name:  "TestMetric2",
								Slice: []int{0, 0, 500},
							},
							{
								Name:  "TestMetric3",
								Slice: []int{0, 0, 500},
							},
						},
					},
					{
						ID: run2.ID,
						Traces: []request.AlignedMetricTraceRequest{
							{
								Name:  "TestMetric1",
								Slice: []int{0, 0, 500},
							},
							{
								Name:  "TestMetric2",
								Slice: []int{0, 0, 500},
							},
							{
								Name:  "TestMetric3",
								Slice: []int{0, 0, 500},
							},
						},
					},
					{
						ID: run3.ID,
						Traces: []request.AlignedMetricTraceRequest{
							{
								Name:  "TestMetric1",
								Slice: []int{0, 0, 500},
							},
							{
								Name:  "TestMetric2",
								Slice: []int{0, 0, 500},
							},
							{
								Name:  "TestMetric3",
								Slice: []int{0, 0, 500},
							},
						},
					},
				},
			},
			response: []float64{metric1Run1.Value, metric1Run2.Value, metric1Run3.Value},
		},
		{
			name: "TestSearchAlignedByMetric2",
			request: &request.GetAlignedMetricRequest{
				AlignBy: "TestMetric2",
				Runs: []request.AlignedMetricRunRequest{
					{
						ID: run1.ID,
						Traces: []request.AlignedMetricTraceRequest{
							{
								Name:  "TestMetric1",
								Slice: []int{0, 0, 500},
							},
							{
								Name:  "TestMetric2",
								Slice: []int{0, 0, 500},
							},
							{
								Name:  "TestMetric3",
								Slice: []int{0, 0, 500},
							},
						},
					},
					{
						ID: run2.ID,
						Traces: []request.AlignedMetricTraceRequest{
							{
								Name:  "TestMetric1",
								Slice: []int{0, 0, 500},
							},
							{
								Name:  "TestMetric2",
								Slice: []int{0, 0, 500},
							},
							{
								Name:  "TestMetric3",
								Slice: []int{0, 0, 500},
							},
						},
					},
					{
						ID: run3.ID,
						Traces: []request.AlignedMetricTraceRequest{
							{
								Name:  "TestMetric1",
								Slice: []int{0, 0, 500},
							},
							{
								Name:  "TestMetric2",
								Slice: []int{0, 0, 500},
							},
							{
								Name:  "TestMetric3",
								Slice: []int{0, 0, 500},
							},
						},
					},
				},
			},
			response: []float64{metric2Run1.Value, metric2Run2.Value, metric2Run3.Value},
		},
		{
			name: "TestSearchAlignedByMetric3",
			request: &request.GetAlignedMetricRequest{
				AlignBy: "TestMetric3",
				Runs: []request.AlignedMetricRunRequest{
					{
						ID: run1.ID,
						Traces: []request.AlignedMetricTraceRequest{
							{
								Name:  "TestMetric1",
								Slice: []int{0, 0, 500},
							},
							{
								Name:  "TestMetric2",
								Slice: []int{0, 0, 500},
							},
							{
								Name:  "TestMetric3",
								Slice: []int{0, 0, 500},
							},
						},
					},
					{
						ID: run2.ID,
						Traces: []request.AlignedMetricTraceRequest{
							{
								Name:  "TestMetric1",
								Slice: []int{0, 0, 500},
							},
							{
								Name:  "TestMetric2",
								Slice: []int{0, 0, 500},
							},
							{
								Name:  "TestMetric3",
								Slice: []int{0, 0, 500},
							},
						},
					},
					{
						ID: run3.ID,
						Traces: []request.AlignedMetricTraceRequest{
							{
								Name:  "TestMetric1",
								Slice: []int{0, 0, 500},
							},
							{
								Name:  "TestMetric2",
								Slice: []int{0, 0, 500},
							},
							{
								Name:  "TestMetric3",
								Slice: []int{0, 0, 500},
							},
						},
					},
				},
			},
			response: []float64{metric3Run1.Value, metric3Run2.Value, metric3Run3.Value},
		},
		{
			name: "TestSearchMetric1Metric2AlignedByMetric1",
			request: &request.GetAlignedMetricRequest{
				AlignBy: "TestMetric1",
				Runs: []request.AlignedMetricRunRequest{
					{
						ID: run1.ID,
						Traces: []request.AlignedMetricTraceRequest{
							{
								Name:  "TestMetric1",
								Slice: []int{0, 0, 500},
							},
							{
								Name:  "TestMetric2",
								Slice: []int{0, 0, 500},
							},
						},
					},
					{
						ID: run2.ID,
						Traces: []request.AlignedMetricTraceRequest{
							{
								Name:  "TestMetric1",
								Slice: []int{0, 0, 500},
							},
							{
								Name:  "TestMetric2",
								Slice: []int{0, 0, 500},
							},
						},
					},
					{
						ID: run3.ID,
						Traces: []request.AlignedMetricTraceRequest{
							{
								Name:  "TestMetric1",
								Slice: []int{0, 0, 500},
							},
							{
								Name:  "TestMetric2",
								Slice: []int{0, 0, 500},
							},
						},
					},
				},
			},
			response: []float64{metric1Run1.Value, metric1Run2.Value, metric1Run3.Value},
		},
		{
			name: "TestSearchMetric1Metric2AlignedByMetric2",
			request: &request.GetAlignedMetricRequest{
				AlignBy: "TestMetric2",
				Runs: []request.AlignedMetricRunRequest{
					{
						ID: run1.ID,
						Traces: []request.AlignedMetricTraceRequest{
							{
								Name:  "TestMetric1",
								Slice: []int{0, 0, 500},
							},
							{
								Name:  "TestMetric2",
								Slice: []int{0, 0, 500},
							},
						},
					},
					{
						ID: run2.ID,
						Traces: []request.AlignedMetricTraceRequest{
							{
								Name:  "TestMetric1",
								Slice: []int{0, 0, 500},
							},
							{
								Name:  "TestMetric2",
								Slice: []int{0, 0, 500},
							},
						},
					},
					{
						ID: run3.ID,
						Traces: []request.AlignedMetricTraceRequest{
							{
								Name:  "TestMetric1",
								Slice: []int{0, 0, 500},
							},
							{
								Name:  "TestMetric2",
								Slice: []int{0, 0, 500},
							},
						},
					},
				},
			},
			response: []float64{metric2Run1.Value, metric2Run2.Value, metric2Run3.Value},
		},
		{
			name: "TestSearchMetric1Metric2AlignedByMetric3",
			request: &request.GetAlignedMetricRequest{
				AlignBy: "TestMetric3",
				Runs: []request.AlignedMetricRunRequest{
					{
						ID: run1.ID,
						Traces: []request.AlignedMetricTraceRequest{
							{
								Name:  "TestMetric1",
								Slice: []int{0, 0, 500},
							},
							{
								Name:  "TestMetric2",
								Slice: []int{0, 0, 500},
							},
						},
					},
					{
						ID: run2.ID,
						Traces: []request.AlignedMetricTraceRequest{
							{
								Name:  "TestMetric1",
								Slice: []int{0, 0, 500},
							},
							{
								Name:  "TestMetric2",
								Slice: []int{0, 0, 500},
							},
						},
					},
					{
						ID: run3.ID,
						Traces: []request.AlignedMetricTraceRequest{
							{
								Name:  "TestMetric1",
								Slice: []int{0, 0, 500},
							},
							{
								Name:  "TestMetric2",
								Slice: []int{0, 0, 500},
							},
						},
					},
				},
			},
			response: []float64{metric3Run1.Value, metric3Run2.Value, metric3Run3.Value},
		},
		{
			name: "TestSearchMetric1AlignedByMetric1",
			request: &request.GetAlignedMetricRequest{
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
					{
						ID: run2.ID,
						Traces: []request.AlignedMetricTraceRequest{
							{
								Name:  "TestMetric1",
								Slice: []int{0, 0, 500},
							},
						},
					},
					{
						ID: run3.ID,
						Traces: []request.AlignedMetricTraceRequest{
							{
								Name:  "TestMetric1",
								Slice: []int{0, 0, 500},
							},
						},
					},
				},
			},
			response: []float64{metric1Run1.Value, metric1Run2.Value, metric1Run3.Value},
		},
		{
			name: "TestSearchMetric1AlignedByMetric2",
			request: &request.GetAlignedMetricRequest{
				AlignBy: "TestMetric2",
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
					{
						ID: run2.ID,
						Traces: []request.AlignedMetricTraceRequest{
							{
								Name:  "TestMetric1",
								Slice: []int{0, 0, 500},
							},
						},
					},
					{
						ID: run3.ID,
						Traces: []request.AlignedMetricTraceRequest{
							{
								Name:  "TestMetric1",
								Slice: []int{0, 0, 500},
							},
						},
					},
				},
			},
			response: []float64{metric2Run1.Value, metric2Run2.Value, metric2Run3.Value},
		},
		{
			name: "TestSearchMetric1AlignedByMetric3",
			request: &request.GetAlignedMetricRequest{
				AlignBy: "TestMetric3",
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
					{
						ID: run2.ID,
						Traces: []request.AlignedMetricTraceRequest{
							{
								Name:  "TestMetric1",
								Slice: []int{0, 0, 500},
							},
						},
					},
					{
						ID: run3.ID,
						Traces: []request.AlignedMetricTraceRequest{
							{
								Name:  "TestMetric1",
								Slice: []int{0, 0, 500},
							},
						},
					},
				},
			},
			response: []float64{metric3Run1.Value, metric3Run2.Value, metric3Run3.Value},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			resp := new(bytes.Buffer)
			s.Require().Nil(s.AIMClient().WithMethod(
				http.MethodPost,
			).WithRequest(
				tt.request,
			).WithResponse(
				resp,
			).WithResponseType(
				helpers.ResponseTypeBuffer,
			).DoRequest(
				"/runs/search/metric/align/",
			))

			decodedData, err := encoding.Decode(resp)
			s.Require().Nil(err)

			xValues := make(map[int][]float64)

			for _, run := range runs {
				metricCount := 0
				for decodedData[fmt.Sprintf("%v.%d.name", run.ID, metricCount)] != nil {
					valueKey := fmt.Sprintf("%v.%d.x_axis_values.blob", run.ID, metricCount)
					xValues[metricCount] = append(xValues[metricCount], decodedData[valueKey].([]float64)[0])

					metricCount++
				}
			}

			// Check if the received values for each metric match the expected ones
			for _, metricValues := range xValues {
				s.Equal(tt.response, metricValues)
			}
		})
	}
}
