package experiment

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
)

func TestValidateCreateExperimentRequest_Ok(t *testing.T) {
	err := ValidateCreateExperimentRequest(&request.CreateExperimentRequest{
		Name:             "name",
		ArtifactLocation: "location.com",
	})
	assert.Nil(t, err)
}

func TestValidateCreateExperimentRequest_Error(t *testing.T) {
	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.CreateExperimentRequest
	}{
		{
			name:    "EmptyNameProperty",
			error:   api.NewInvalidParameterValueError("Missing value for required parameter 'name'"),
			request: &request.CreateExperimentRequest{},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCreateExperimentRequest(tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}

func TestValidateUpdateExperimentRequest_Ok(t *testing.T) {
	err := ValidateUpdateExperimentRequest(&request.UpdateExperimentRequest{
		ID:   "id",
		Name: "name",
	})
	assert.Nil(t, err)
}

func TestValidateUpdateExperimentRequest_Error(t *testing.T) {
	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.UpdateExperimentRequest
	}{
		{
			name:    "EmptyIdProperty",
			error:   api.NewInvalidParameterValueError("Missing value for required parameter 'experiment_id'"),
			request: &request.UpdateExperimentRequest{},
		},
		{
			name:  "EmptyNameProperty",
			error: api.NewInvalidParameterValueError("Missing value for required parameter 'new_name'"),
			request: &request.UpdateExperimentRequest{
				ID: "id",
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUpdateExperimentRequest(tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}

func TestValidateGetExperimentByIDRequest_Ok(t *testing.T) {
	err := ValidateGetExperimentByIDRequest(&request.GetExperimentRequest{
		ID: "id",
	})
	assert.Nil(t, err)
}

func TestValidateGetExperimentByIDRequest_Error(t *testing.T) {
	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.UpdateExperimentRequest
	}{
		{
			name:    "EmptyIdProperty",
			error:   api.NewInvalidParameterValueError("Missing value for required parameter 'experiment_id'"),
			request: &request.UpdateExperimentRequest{},
		},
		{
			name:  "EmptyNameProperty",
			error: api.NewInvalidParameterValueError("Missing value for required parameter 'new_name'"),
			request: &request.UpdateExperimentRequest{
				ID: "id",
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUpdateExperimentRequest(tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}

func TestValidateGetExperimentByNameRequest_Ok(t *testing.T) {
	err := ValidateGetExperimentByNameRequest(&request.GetExperimentRequest{
		Name: "name",
	})
	assert.Nil(t, err)
}

func TestValidateGetExperimentByNameRequest_Error(t *testing.T) {
	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.GetExperimentRequest
	}{
		{
			name:    "EmptyNameProperty",
			error:   api.NewInvalidParameterValueError("Missing value for required parameter 'experiment_name'"),
			request: &request.GetExperimentRequest{},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGetExperimentByNameRequest(tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}

func TestValidateDeleteExperimentRequest_Ok(t *testing.T) {
	err := ValidateDeleteExperimentRequest(&request.DeleteExperimentRequest{
		ID: "id",
	})
	assert.Nil(t, err)
}

func TestValidateDeleteExperimentRequest_Error(t *testing.T) {
	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.DeleteExperimentRequest
	}{
		{
			name:    "EmptyIDProperty",
			error:   api.NewInvalidParameterValueError("Missing value for required parameter 'experiment_id'"),
			request: &request.DeleteExperimentRequest{},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDeleteExperimentRequest(tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}

func TestValidateRestoreExperimentRequest_Ok(t *testing.T) {
	err := ValidateRestoreExperimentRequest(&request.RestoreExperimentRequest{
		ID: "id",
	})
	assert.Nil(t, err)
}

func TestValidateRestoreExperimentRequest_Error(t *testing.T) {
	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.RestoreExperimentRequest
	}{
		{
			name:    "EmptyIDProperty",
			error:   api.NewInvalidParameterValueError("Missing value for required parameter 'experiment_id'"),
			request: &request.RestoreExperimentRequest{},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRestoreExperimentRequest(tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}

func TestValidateSearchExperimentsRequest_Ok(t *testing.T) {
	err := ValidateSearchExperimentsRequest(&request.SearchExperimentsRequest{
		MaxResults: 10,
		ViewType:   request.ViewTypeAll,
	})
	assert.Nil(t, err)
}

func TestValidateSearchExperimentsRequest_Error(t *testing.T) {
	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.SearchExperimentsRequest
	}{
		{
			name:  "IncorrectViewTypeProperty",
			error: api.NewInvalidParameterValueError("Invalid view_type 'incorrect-view-type'"),
			request: &request.SearchExperimentsRequest{
				ViewType: request.ViewType("incorrect-view-type"),
			},
		},
		{
			name:  "IncorrectMaxResultsProperty",
			error: api.NewInvalidParameterValueError("Invalid value for parameter 'max_results' supplied."),
			request: &request.SearchExperimentsRequest{
				ViewType:   request.ViewTypeAll,
				MaxResults: MaxResultsPerPage + 1,
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSearchExperimentsRequest(tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}

func TestValidateSetExperimentTagRequest_Ok(t *testing.T) {
	err := ValidateSetExperimentTagRequest(&request.SetExperimentTagRequest{
		ID:  "id",
		Key: "key",
	})
	assert.Nil(t, err)
}

func TestValidateSetExperimentTagRequest_Error(t *testing.T) {
	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.SetExperimentTagRequest
	}{
		{
			name:    "EmptyIDProperty",
			error:   api.NewInvalidParameterValueError("Missing value for required parameter 'experiment_id'"),
			request: &request.SetExperimentTagRequest{},
		},
		{
			name:  "EmptyKeyProperty",
			error: api.NewInvalidParameterValueError("Missing value for required parameter 'key'"),
			request: &request.SetExperimentTagRequest{
				ID: "id",
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSetExperimentTagRequest(tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}
