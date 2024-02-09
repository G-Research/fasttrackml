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
		name          string
		ns            *models.Namespace
		inputRequest  *request.SearchRunsRequest
		resultRequest *request.SearchRunsRequest
	}{
		{
			name: "DefaultExperimentIDsProvided",
			ns: &models.Namespace{
				DefaultExperimentID: common.GetPointer(int32(123)),
			},
			inputRequest: &request.SearchRunsRequest{
				ExperimentIDs: []string{"0", "456"},
			},
			resultRequest: &request.SearchRunsRequest{
				ExperimentIDs: []string{"123", "456"},
			},
		},
		{
			name: "DefaultExperimentIDsNotProvided",
			ns: &models.Namespace{
				DefaultExperimentID: common.GetPointer(int32(123)),
			},
			inputRequest: &request.SearchRunsRequest{
				ExperimentIDs: []string{"789", "456"},
			},
			resultRequest: &request.SearchRunsRequest{
				ExperimentIDs: []string{"789", "456"},
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			adjustSearchRunsRequestForNamespace(tt.ns, tt.inputRequest)
			assert.Equal(t, tt.resultRequest, tt.inputRequest)
		})
	}
}

func Test_adjustCreateRunRequestForNamespace_Ok(t *testing.T) {
	testData := []struct {
		name          string
		ns            *models.Namespace
		inputRequest  *request.CreateRunRequest
		resultRequest *request.CreateRunRequest
	}{
		{
			name: "DefaultExperimentIDProvided",
			ns: &models.Namespace{
				DefaultExperimentID: common.GetPointer(int32(123)),
			},
			inputRequest: &request.CreateRunRequest{
				ExperimentID: "0",
			},
			resultRequest: &request.CreateRunRequest{
				ExperimentID: "123",
			},
		},
		{
			name: "DefaultExperimentIDNotProvided",
			ns: &models.Namespace{
				DefaultExperimentID: common.GetPointer(int32(123)),
			},
			inputRequest: &request.CreateRunRequest{
				ExperimentID: "456",
			},
			resultRequest: &request.CreateRunRequest{
				ExperimentID: "456",
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			adjustCreateRunRequestForNamespace(tt.ns, tt.inputRequest)
			assert.Equal(t, tt.resultRequest, tt.inputRequest)
		})
	}
}
