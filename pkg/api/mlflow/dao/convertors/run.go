package convertors

import (
	"fmt"

	"database/sql"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/database"
	"github.com/G-Research/fasttrackml/pkg/models"
)

// supported tag keys.
const (
	TagKeyUser       = "mlflow.user"
	TagKeyRunName    = "mlflow.runName"
	TagKeySourceName = "mlflow.source.name"
	TagKeySourceType = "mlflow.source.type"
)

// ConvertCreateRunRequestToDBModel converts request.CreateRunRequest into actual models.Run model.
func ConvertCreateRunRequestToDBModel(experiment *models.Experiment, req *request.CreateRunRequest) *models.Run {
	run := models.Run{
		// TODO:Dsuhinin why sometimes we create ID like that and sometimes created it using DB?
		ID:           database.NewUUID(),
		Name:         req.Name,
		ExperimentID: *experiment.ID,
		UserID:       req.UserID,
		Status:       models.StatusRunning,
		StartTime: sql.NullInt64{
			Int64: req.StartTime,
			Valid: true,
		},
		LifecycleStage: models.LifecycleStageActive,
		Tags:           make([]models.Tag, len(req.Tags)),
	}

	run.ArtifactURI = fmt.Sprintf("%s/%s/artifacts", experiment.ArtifactLocation, run.ID)

	for n, tag := range req.Tags {
		switch tag.Key {
		case TagKeyUser:
			if run.UserID == "" {
				run.UserID = tag.Value
			}
		case TagKeySourceName:
			run.SourceName = tag.Value
		case TagKeySourceType:
			run.SourceType = tag.Value
		case TagKeyRunName:
			if run.Name == "" {
				run.Name = tag.Value
			}
		}
		run.Tags[n] = models.Tag{
			Key:   tag.Key,
			Value: tag.Value,
		}
	}

	if run.Name == "" {
		run.Name = GenerateRandomName()
		run.Tags = append(run.Tags, models.Tag{
			Key:   TagKeySourceName,
			Value: run.Name,
		})
	}

	if run.SourceType == "" {
		run.SourceType = "UNKNOWN"
	}
	return &run
}

// ConvertUpdateRunRequestToDBModel converts request.UpdateRunRequest into actual models.Run model.
func ConvertUpdateRunRequestToDBModel(run *models.Run, req *request.UpdateRunRequest) *models.Run {
	run.Name = req.Name
	run.Status = models.Status(req.Status)
	run.EndTime = sql.NullInt64{
		Int64: req.EndTime,
		Valid: true,
	}
	return run
}
