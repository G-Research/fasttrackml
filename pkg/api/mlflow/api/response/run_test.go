package response

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

func TestNewRunPartialResponse(t *testing.T) {
	testData := []struct {
		name             string
		run              *models.Run
		expectedResponse *RunPartialResponse
	}{
		{
			name: "WithNaNValue",
			run: &models.Run{
				ID:             "ID",
				Name:           "Name",
				SourceType:     "SourceType",
				SourceName:     "SourceName",
				EntryPointName: "EntryPointName",
				UserID:         "UserID",
				Status:         "Status",
				StartTime: sql.NullInt64{
					Valid: true,
					Int64: 1234567890,
				},
				EndTime: sql.NullInt64{
					Valid: true,
					Int64: 1234567890,
				},
				SourceVersion:  "SourceVersion",
				LifecycleStage: "LifecycleStage",
				ArtifactURI:    "ArtifactURI",
				ExperimentID:   1,
				RowNum:         1,
				Params: []models.Param{
					{
						Key:      "Key",
						ValueStr: common.GetPointer[string]("Value"),
						RunID:    "RunID",
					},
				},
				Tags: []models.Tag{
					{
						Key:   "Key",
						Value: "Value",
						RunID: "RunID",
					},
				},
				LatestMetrics: []models.LatestMetric{
					{
						Key:       "Key",
						Value:     1,
						Timestamp: 1234567890,
						Step:      1,
						IsNan:     true,
						RunID:     "",
						LastIter:  0,
					},
				},
			},
			expectedResponse: &RunPartialResponse{
				Info: RunInfoPartialResponse{
					ID:             "ID",
					UUID:           "ID",
					Name:           "Name",
					ExperimentID:   "1",
					UserID:         "UserID",
					Status:         "Status",
					StartTime:      1234567890,
					EndTime:        1234567890,
					ArtifactURI:    "ArtifactURI",
					LifecycleStage: "LifecycleStage",
				},
				Data: RunDataPartialResponse{
					Metrics: []RunMetricPartialResponse{
						{
							Key:       "Key",
							Value:     common.NANValue,
							Timestamp: 1234567890,
							Step:      1,
						},
					},
					Params: []RunParamPartialResponse{{
						Key:   "Key",
						Value: "Value",
					}},
					Tags: []RunTagPartialResponse{{
						Key:   "Key",
						Value: "Value",
					}},
				},
			},
		},
		{
			name: "WithNotNaNValue",
			run: &models.Run{
				ID:             "ID",
				Name:           "Name",
				SourceType:     "SourceType",
				SourceName:     "SourceName",
				EntryPointName: "EntryPointName",
				UserID:         "UserID",
				Status:         "Status",
				StartTime: sql.NullInt64{
					Valid: true,
					Int64: 1234567890,
				},
				EndTime: sql.NullInt64{
					Valid: true,
					Int64: 1234567890,
				},
				SourceVersion:  "SourceVersion",
				LifecycleStage: "LifecycleStage",
				ArtifactURI:    "ArtifactURI",
				ExperimentID:   1,
				RowNum:         1,
				Params: []models.Param{
					{
						Key:      "Key",
						ValueStr: common.GetPointer[string]("Value"),
						RunID:    "RunID",
					},
				},
				Tags: []models.Tag{
					{
						Key:   "Key",
						Value: "Value",
						RunID: "RunID",
					},
				},
				LatestMetrics: []models.LatestMetric{
					{
						Key:       "Key",
						Value:     123,
						Timestamp: 1234567890,
						Step:      1,
						IsNan:     false,
						RunID:     "",
						LastIter:  0,
					},
				},
			},
			expectedResponse: &RunPartialResponse{
				Info: RunInfoPartialResponse{
					ID:             "ID",
					UUID:           "ID",
					Name:           "Name",
					ExperimentID:   "1",
					UserID:         "UserID",
					Status:         "Status",
					StartTime:      1234567890,
					EndTime:        1234567890,
					ArtifactURI:    "ArtifactURI",
					LifecycleStage: "LifecycleStage",
				},
				Data: RunDataPartialResponse{
					Metrics: []RunMetricPartialResponse{
						{
							Key:       "Key",
							Value:     float64(123),
							Timestamp: 1234567890,
							Step:      1,
						},
					},
					Params: []RunParamPartialResponse{{
						Key:   "Key",
						Value: "Value",
					}},
					Tags: []RunTagPartialResponse{{
						Key:   "Key",
						Value: "Value",
					}},
				},
			},
		},
		{
			name: "WithTagKeyRunName",
			run: &models.Run{
				Params: []models.Param{},
				Tags: []models.Tag{
					{
						Key:   "mlflow.runName",
						Value: "Value",
						RunID: "RunID",
					},
				},
				LatestMetrics: []models.LatestMetric{},
			},
			expectedResponse: &RunPartialResponse{
				Info: RunInfoPartialResponse{
					Name:         "Value",
					ExperimentID: "0",
				},
				Data: RunDataPartialResponse{
					Tags: []RunTagPartialResponse{{
						Key:   "mlflow.runName",
						Value: "Value",
					}},
					Params:  []RunParamPartialResponse{},
					Metrics: []RunMetricPartialResponse{},
				},
			},
		},
		{
			name: "WithTagKeyName",
			run: &models.Run{
				Params: []models.Param{},
				Tags: []models.Tag{
					{
						Key:   "mlflow.user",
						Value: "Value",
						RunID: "RunID",
					},
				},
				LatestMetrics: []models.LatestMetric{},
			},
			expectedResponse: &RunPartialResponse{
				Info: RunInfoPartialResponse{
					UserID:       "Value",
					ExperimentID: "0",
				},
				Data: RunDataPartialResponse{
					Tags: []RunTagPartialResponse{{
						Key:   "mlflow.user",
						Value: "Value",
					}},
					Params:  []RunParamPartialResponse{},
					Metrics: []RunMetricPartialResponse{},
				},
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			actualResponse := NewRunPartialResponse(tt.run)
			assert.Equal(t, tt.expectedResponse, actualResponse)
		})
	}
}
