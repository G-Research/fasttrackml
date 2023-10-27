package metric

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

func Test_adjustGetMetricHistoriesRequestForNamespace_Ok(t *testing.T) {
	testData := []struct {
		name           string
		ns             *models.Namespace
		srr            *request.GetMetricHistoriesRequest
		expectedResult *request.GetMetricHistoriesRequest
	}{
		{
			name: "DefaultExperimentIDsProvided",
			ns: &models.Namespace{
				DefaultExperimentID: common.GetPointer(int32(123)),
			},
			srr: &request.GetMetricHistoriesRequest{
				ExperimentIDs: []string{"0", "456"},
			},
			expectedResult: &request.GetMetricHistoriesRequest{
				ExperimentIDs: []string{"123", "456"},
			},
		},
		{
			name: "DefaultExperimentIDsNotProvided",
			ns: &models.Namespace{
				DefaultExperimentID: common.GetPointer(int32(123)),
			},
			srr: &request.GetMetricHistoriesRequest{
				ExperimentIDs: []string{"789", "456"},
			},
			expectedResult: &request.GetMetricHistoriesRequest{
				ExperimentIDs: []string{"789", "456"},
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			adjustGetMetricHistoriesRequestForNamespace(tt.ns, tt.srr)
			assert.Equal(t, tt.expectedResult, tt.srr)
		})
	}
}
