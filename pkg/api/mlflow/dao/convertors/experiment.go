package convertors

import (
	"database/sql"
	"time"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

// ConvertCreateExperimentToDBModel converts request.CreateExperimentRequest into actual models.Experiment model.
func ConvertCreateExperimentToDBModel(req *request.CreateExperimentRequest) *models.Experiment {
	ts := time.Now().UTC().UnixMilli()
	experiment := models.Experiment{
		Name:           req.Name,
		LifecycleStage: models.LifecycleStageActive,
		CreationTime: sql.NullInt64{
			Int64: ts,
			Valid: true,
		},
		LastUpdateTime: sql.NullInt64{
			Int64: ts,
			Valid: true,
		},
		Tags:             make([]models.ExperimentTag, len(req.Tags)),
		ArtifactLocation: req.ArtifactLocation,
	}

	for n, tag := range req.Tags {
		experiment.Tags[n] = models.ExperimentTag{
			Key:   tag.Key,
			Value: tag.Value,
		}
	}

	return &experiment
}

// ConvertUpdateExperimentToDBModel converts request.UpdateExperimentRequest into actual models.Experiment model.
func ConvertUpdateExperimentToDBModel(
	experiment *models.Experiment, req *request.UpdateExperimentRequest,
) *models.Experiment {
	experiment.Name = req.Name
	experiment.LastUpdateTime = sql.NullInt64{
		Int64: time.Now().UTC().UnixMilli(),
		Valid: true,
	}
	return experiment
}
