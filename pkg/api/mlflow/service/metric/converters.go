package metric

import (
	"fmt"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
)

// adjustGetMetricHistoriesRequestForNamespace preprocesses the GetMetricHistoriesRequest for the given namespace.
func adjustGetMetricHistoriesRequestForNamespace(ns *models.Namespace, req *request.GetMetricHistoriesRequest) {
	for i, expID := range req.ExperimentIDs {
		if expID == "0" {
			req.ExperimentIDs[i] = fmt.Sprintf("%d", *ns.DefaultExperimentID)
		}
	}
}
