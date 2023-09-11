package helpers

import (
	"time"

	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
)

// TransformRunsToActivityMap transform a slice of runs into a map of experiments activity to match the GetExperimentsActivity endpoint response.
func TransformRunsToActivityMap(runs []*models.Run) map[string]int {
	activity := map[string]int{}
	for _, r := range runs {
		key := time.UnixMilli(r.StartTime.Int64).Format("2006-01-02T15:00:00")
		activity[key] += 1
	}
	return activity
}
