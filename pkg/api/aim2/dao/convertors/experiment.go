package convertors

import (
	"github.com/G-Research/fasttrackml/pkg/api/aim2/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

// ConvertUpdateExperimentToDBModel converts request.CreateExperimentRequest into actual models.Experiment model.
func ConvertUpdateExperimentToDBModel(
	req *request.UpdateExperimentRequest, experiment *models.Experiment,
) *models.Experiment {
	if req.Archived != nil {
		if *req.Archived {
			experiment.LifecycleStage = models.LifecycleStageDeleted
		} else {
			experiment.LifecycleStage = models.LifecycleStageActive
		}
	}
	if req.Name != nil {
		experiment.Name = *req.Name
	}
	return experiment
}
