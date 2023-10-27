package run

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

func Test_adjustSearchRunsRequestForNamespace_Ok(t *testing.T) {
	testData := []struct {
		name           string
		ns             *models.Namespace
		srr            *request.SearchRunsRequest
		expectedResult *request.SearchRunsRequest
	}{
		{
			name: "DefaultExperimentIDsProvided",
			ns: &models.Namespace{
				DefaultExperimentID: common.GetPointer(int32(123)),
			},
			srr: &request.SearchRunsRequest{
				ExperimentIDs: []string{"0", "456"},
			},
			expectedResult: &request.SearchRunsRequest{
				ExperimentIDs: []string{"123", "456"},
			},
		},
		{
			name: "DefaultExperimentIDsNotProvided",
			ns: &models.Namespace{
				DefaultExperimentID: common.GetPointer(int32(123)),
			},
			srr: &request.SearchRunsRequest{
				ExperimentIDs: []string{"789", "456"},
			},
			expectedResult: &request.SearchRunsRequest{
				ExperimentIDs: []string{"789", "456"},
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			adjustSearchRunsRequestForNamespace(tt.ns, tt.srr)
			assert.Equal(t, tt.expectedResult, tt.srr)
		})
	}
}

func Test_adjustCreateRunRequestForNamespace_Ok(t *testing.T) {
	testData := []struct {
		name           string
		ns             *models.Namespace
		crr            *request.CreateRunRequest
		expectedResult *request.CreateRunRequest
	}{
		{
			name: "DefaultExperimentIDProvided",
			ns: &models.Namespace{
				DefaultExperimentID: common.GetPointer(int32(123)),
			},
			crr: &request.CreateRunRequest{
				ExperimentID: "0",
			},
			expectedResult: &request.CreateRunRequest{
				ExperimentID: "123",
			},
		},
		{
			name: "DefaultExperimentIDNotProvided",
			ns: &models.Namespace{
				DefaultExperimentID: common.GetPointer(int32(123)),
			},
			crr: &request.CreateRunRequest{
				ExperimentID: "456",
			},
			expectedResult: &request.CreateRunRequest{
				ExperimentID: "456",
			},
		},
	}

	for _, tc := range testData {
		t.Run(tc.name, func(t *testing.T) {
			adjustCreateRunRequestForNamespace(tc.ns, tc.crr)
			assert.Equal(t, tc.expectedResult, tc.crr)
		})
	}
}
