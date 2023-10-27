package metric

import (
	"fmt"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

// adjustGetMetricHistoriesRequestForNamespace preprocesses the GetMetricHistoriesRequest for the given namespace.
func adjustGetMetricHistoriesRequestForNamespace(ns *models.Namespace, gmhr *request.GetMetricHistoriesRequest) {
	for i, expID := range gmhr.ExperimentIDs {
		if expID == "0" {
			gmhr.ExperimentIDs[i] = fmt.Sprintf("%d", *ns.DefaultExperimentID)
		}
	}
}
