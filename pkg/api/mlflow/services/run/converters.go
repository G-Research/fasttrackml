package run

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

// adjustSearchRunsRequestForNamespace preprocesses the SearchRunRequest for the given namespace.
func adjustSearchRunsRequestForNamespace(ns *models.Namespace, req *request.SearchRunsRequest) {
	for i, expID := range req.ExperimentIDs {
		if expID == "0" {
			req.ExperimentIDs[i] = fmt.Sprintf("%d", *ns.DefaultExperimentID)
		}
	}
}

// adjustCreateRunRequestForNamespace preprocesses the CreateRunRequest for the given namespace.
func adjustCreateRunRequestForNamespace(ns *models.Namespace, req *request.CreateRunRequest) {
	if req.ExperimentID == "0" {
		req.ExperimentID = fmt.Sprintf("%d", *ns.DefaultExperimentID)
	}
}

// ConvertCreateRunArtifactRequestToModel  converts request of
// `POST /runs/:id/artifact` endpoint to an internal Model object.
func ConvertCreateRunArtifactRequestToModel(
	namespaceID uint, req *request.LogArtifactRequest,
) *models.Artifact {
	return &models.Artifact{
		ID:      uuid.New(),
		Iter:    req.Iter,
		Step:    req.Step,
		RunID:   req.RunID,
		Index:   req.Index,
		Width:   req.Width,
		Height:  req.Height,
		Format:  req.Format,
		Caption: req.Caption,
		BlobURI: req.BlobURI,
	}
}
