package helpers

import (
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

// GetDescriptionFromexperiment returns the description of a given experiment.
func GetDescriptionFromExperiment(experiment models.Experiment) string {
	for _, tag := range experiment.Tags {
		if tag.Key == common.DescriptionTagKey {
			return tag.Value
		}
	}
	return ""
}
