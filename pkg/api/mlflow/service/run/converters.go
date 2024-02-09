package run

import (
	"fmt"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
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
