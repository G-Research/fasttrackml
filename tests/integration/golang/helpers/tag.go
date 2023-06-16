package helpers

import (
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

func CheckTagExists(tags []models.ExperimentTag, key, value string) bool {
	for _, tag := range tags {
		if tag.Key == key && tag.Value == value {
			return true
		}
	}
	return false
}
