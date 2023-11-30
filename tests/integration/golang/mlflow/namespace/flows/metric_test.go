//go:build integration

package flows

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type MetricFlowTestSuite struct {
	helpers.BaseTestSuite
}

// TestMetricFlowTestSuite tests the full `metric` flow connected to namespace functionality.
// Flow contains next endpoints:
// - `GET /metrics/get-history`
// - `GET /metrics/get-history-bulk`
// - `POST /metrics/get-histories` - TODO:dsuhinin we need firstly to create proper decoder.
func TestMetricFlowTestSuite(t *testing.T) {
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

			experiment2, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
				Name:             "Experiment2",
				ArtifactLocation: "/artifact/location",
				LifecycleStage:   models.LifecycleStageActive,
				NamespaceID:      namespace2.ID,
			})
			s.Require().Nil(err)

			// 2. run actual flow test over the test data.
			s.testRunMetricFlow(tt.namespace1Code, tt.namespace2Code, experiment1, experiment2)
		})
	}
}

func (s *MetricFlowTestSuite) testRunMetricFlow(
	namespace1Code, namespace2Code string, experiment1, experiment2 *models.Experiment,
) {
	run1ID := s.createRun(namespace1Code, &request.CreateRunRequest{
		Name:         "Run1",
		ExperimentID: fmt.Sprintf("%d", *experiment1.ID),
	})

	run2ID := s.createRun(namespace2Code, &request.CreateRunRequest{
		Name:         "Run2",
		ExperimentID: fmt.Sprintf("%d", *experiment2.ID),
	})

	// test `GET /runs/get` endpoint.
	// check that runs were created in scope of difference experiment namespaces.
	s.getRunAndCompare(
		namespace1Code,
		request.GetRunRequest{
			RunID: run1ID,
		},
		&response.GetRunResponse{
			Run: &response.RunPartialResponse{
				Info: response.RunInfoPartialResponse{
					ID:             run1ID,
					Name:           "Run1",
					Status:         string(models.StatusRunning),
					ArtifactURI:    fmt.Sprintf("/artifact/location/%s/artifacts", run1ID),
					ExperimentID:   fmt.Sprintf("%d", *experiment1.ID),
					LifecycleStage: string(models.LifecycleStageActive),
				},
				Data: response.RunDataPartialResponse{},
			},
		},
	)
	s.getRunAndCompare(
		namespace2Code,
		request.GetRunRequest{
			RunID: run2ID,
		},
		&response.GetRunResponse{
			Run: &response.RunPartialResponse{
				Info: response.RunInfoPartialResponse{
					ID:             run2ID,
					Name:           "Run2",
					Status:         string(models.StatusRunning),
					ArtifactURI:    fmt.Sprintf("/artifact/location/%s/artifacts", run2ID),
					ExperimentID:   fmt.Sprintf("%d", *experiment2.ID),
					LifecycleStage: string(models.LifecycleStageActive),
				},
				Data: response.RunDataPartialResponse{},
			},
		},
	)

	// create Run metrics in different runs in scope of different namespaces.
	s.logRunMetric(namespace1Code, &request.LogMetricRequest{
		RunID:     run1ID,
		Key:       "key1",
		Value:     1.1,
		Timestamp: 123456789,
		Step:      1,
	})
	s.logRunMetric(namespace1Code, &request.LogMetricRequest{
		RunID:     run1ID,
		Key:       "key2",
		Value:     2.2,
		Timestamp: 123456789,
		Step:      1,
	})
	s.logRunMetric(namespace2Code, &request.LogMetricRequest{
		RunID:     run2ID,
		Key:       "key3",
		Value:     3.3,
		Timestamp: 123456789,
		Step:      1,
	})
	s.logRunMetric(namespace2Code, &request.LogMetricRequest{
		RunID:     run2ID,
		Key:       "key4",
		Value:     4.4,
		Timestamp: 123456789,
		Step:      1,
	})

	// test `GET /metrics/get-history-bulk` endpoint.
	// try to get metrics for runs which belong to own namespaces.
	s.getMetricHistoryBulkAndCompare(namespace1Code, request.GetMetricHistoryBulkRequest{
		RunIDs:    []string{run1ID},
		MetricKey: "key1",
	}, response.GetMetricHistoryResponse{
		Metrics: []response.MetricPartialResponse{
			{
				RunID:     run1ID,
				Key:       "key1",
				Step:      1,
				Value:     1.1,
				Timestamp: 123456789,
			},
		},
	})
	s.getMetricHistoryBulkAndCompare(namespace2Code, request.GetMetricHistoryBulkRequest{
		RunIDs:    []string{run2ID},
		MetricKey: "key3",
	}, response.GetMetricHistoryResponse{
		Metrics: []response.MetricPartialResponse{
			{
				RunID:     run2ID,
				Key:       "key3",
				Step:      1,
				Value:     3.3,
				Timestamp: 123456789,
			},
		},
	})

	// test `GET /metrics/get-history-bulk` endpoint.
	// try to get metrics for runs which do not belong to own namespaces.
	s.getMetricHistoryBulkAndCompare(namespace1Code, request.GetMetricHistoryBulkRequest{
		RunIDs:    []string{run2ID},
		MetricKey: "key3",
	}, response.GetMetricHistoryResponse{
		Metrics: []response.MetricPartialResponse{},
	})
	s.getMetricHistoryBulkAndCompare(namespace2Code, request.GetMetricHistoryBulkRequest{
		RunIDs:    []string{run1ID},
		MetricKey: "key1",
	}, response.GetMetricHistoryResponse{
		Metrics: []response.MetricPartialResponse{},
	})

	// test `GET /metrics/get-history-bulk` endpoint.
	// try to get metrics for mixed runs.
	s.getMetricHistoryBulkAndCompare(namespace1Code, request.GetMetricHistoryBulkRequest{
		RunIDs:    []string{run1ID, run2ID},
		MetricKey: "key1",
	}, response.GetMetricHistoryResponse{
		Metrics: []response.MetricPartialResponse{
			{
				RunID:     run1ID,
				Key:       "key1",
				Step:      1,
				Value:     1.1,
				Timestamp: 123456789,
			},
		},
	})
	s.getMetricHistoryBulkAndCompare(namespace2Code, request.GetMetricHistoryBulkRequest{
		RunIDs:    []string{run2ID, run1ID},
		MetricKey: "key3",
	}, response.GetMetricHistoryResponse{
		Metrics: []response.MetricPartialResponse{
			{
				RunID:     run2ID,
				Key:       "key3",
				Step:      1,
				Value:     3.3,
				Timestamp: 123456789,
			},
		},
	})

	// test `GET /metrics/get-history` endpoint.
	// try to get metrics for runs in their own scopes.
	s.getMetricHistoryAndCompare(namespace1Code, request.GetMetricHistoryRequest{
		RunID:     run1ID,
		MetricKey: "key1",
	}, response.GetMetricHistoryResponse{
		Metrics: []response.MetricPartialResponse{
			{
				Key:       "key1",
				Step:      1,
				Value:     1.1,
				Timestamp: 123456789,
			},
		},
	})
	s.getMetricHistoryAndCompare(namespace2Code, request.GetMetricHistoryRequest{
		RunID:     run2ID,
		MetricKey: "key3",
	}, response.GetMetricHistoryResponse{
		Metrics: []response.MetricPartialResponse{
			{
				Key:       "key3",
				Step:      1,
				Value:     3.3,
				Timestamp: 123456789,
			},
		},
	})

	// test `GET /metrics/get-history` endpoint.
	// try to get metrics for runs which do not belong to own namespaces.
	s.getMetricHistoryAndCompare(namespace1Code, request.GetMetricHistoryRequest{
		RunID:     run2ID,
		MetricKey: "key3",
	}, response.GetMetricHistoryResponse{})
	s.getMetricHistoryAndCompare(namespace2Code, request.GetMetricHistoryRequest{
		RunID:     run1ID,
		MetricKey: "key1",
	}, response.GetMetricHistoryResponse{})
}

func (s *MetricFlowTestSuite) createRun(
	namespace string, req *request.CreateRunRequest,
) string {
	resp := response.CreateRunResponse{}
	s.Require().Nil(
		s.MlflowClient().WithMethod(
			http.MethodPost,
		).WithNamespace(
			namespace,
		).WithRequest(
			req,
		).WithResponse(
			&resp,
		).DoRequest(
			"%s%s", mlflow.RunsRoutePrefix, mlflow.RunsCreateRoute,
		),
	)
	return resp.Run.Info.ID
}

func (s *MetricFlowTestSuite) getRunAndCompare(
	namespace string, req request.GetRunRequest, expectedResponse *response.GetRunResponse,
) {
	resp := response.GetRunResponse{}
	s.Require().Nil(
		s.MlflowClient().WithNamespace(
			namespace,
		).WithQuery(
			req,
		).WithResponse(
			&resp,
		).DoRequest(
			"%s%s", mlflow.RunsRoutePrefix, mlflow.RunsGetRoute,
		),
	)
	s.Equal(expectedResponse.Run.Info.ID, resp.Run.Info.ID)
	s.Equal(expectedResponse.Run.Info.Name, resp.Run.Info.Name)
	s.Equal(expectedResponse.Run.Info.Status, resp.Run.Info.Status)
	s.Equal(expectedResponse.Run.Info.ArtifactURI, resp.Run.Info.ArtifactURI)
	s.Equal(expectedResponse.Run.Info.ExperimentID, resp.Run.Info.ExperimentID)
	s.Equal(expectedResponse.Run.Info.LifecycleStage, resp.Run.Info.LifecycleStage)
	if expectedResponse.Run.Data.Tags != nil {
		s.Equal(expectedResponse.Run.Data.Tags, resp.Run.Data.Tags)
	}
	if expectedResponse.Run.Data.Params != nil {
		s.Equal(expectedResponse.Run.Data.Params, resp.Run.Data.Params)
	}
	if expectedResponse.Run.Data.Metrics != nil {
		s.Equal(expectedResponse.Run.Data.Metrics, resp.Run.Data.Metrics)
	}
}

func (s *MetricFlowTestSuite) logRunMetric(namespace string, req *request.LogMetricRequest) {
	s.Require().Nil(
		s.MlflowClient().WithMethod(
			http.MethodPost,
		).WithNamespace(
			namespace,
		).WithRequest(
			req,
		).DoRequest(
			"%s%s", mlflow.RunsRoutePrefix, mlflow.RunsLogMetricRoute,
		),
	)
}

func (s *MetricFlowTestSuite) getMetricHistoryBulkAndCompare(
	namespace string, req request.GetMetricHistoryBulkRequest, expectedResponse response.GetMetricHistoryResponse,
) {
	actualResponse := response.GetMetricHistoryResponse{}
	s.Require().Nil(
		s.MlflowClient().WithNamespace(
			namespace,
		).WithQuery(
			req,
		).WithResponse(
			&actualResponse,
		).DoRequest(
			"%s%s", mlflow.MetricsRoutePrefix, mlflow.MetricsGetHistoryBulkRoute,
		),
	)
	s.Equal(expectedResponse, actualResponse)
}

func (s *MetricFlowTestSuite) getMetricHistoryAndCompare(
	namespace string, req request.GetMetricHistoryRequest, expectedResponse response.GetMetricHistoryResponse,
) {
	actualResponse := response.GetMetricHistoryResponse{}
	s.Require().Nil(
		s.MlflowClient().WithNamespace(
			namespace,
		).WithQuery(
			req,
		).WithResponse(
			&actualResponse,
		).DoRequest(
			"%s%s", mlflow.MetricsRoutePrefix, mlflow.MetricsGetHistoryRoute,
		),
	)
	s.Equal(expectedResponse, actualResponse)
}
