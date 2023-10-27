package run

import (
	"fmt"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

// AdjustSearchRunsRequestForNamespace preprocesses the SearchRunRequest for the given namespace.
func AdjustSearchRunsRequestForNamespace(ns *models.Namespace, srr *request.SearchRunsRequest) {
	for i, expID := range srr.ExperimentIDs {
		if expID == "0" {
			srr.ExperimentIDs[i] = fmt.Sprintf("%d", *ns.DefaultExperimentID)
		}
	}
}

// AdjustCreateRunRequestForNamespace preprocesses the CreateRunRequest for the given namespace.
func AdjustCreateRunRequestForNamespace(ns *models.Namespace, crr *request.CreateRunRequest) {
	if crr.ExperimentID == "0" {
		crr.ExperimentID = fmt.Sprintf("%d", *ns.DefaultExperimentID)
	}
}
