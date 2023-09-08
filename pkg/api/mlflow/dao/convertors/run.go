package convertors

import (
	"database/sql"
	"net/url"

	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
	"github.com/G-Research/fasttrackml/pkg/database"
)

// supported tag keys.
const (
	TagKeyUser       = "mlflow.user"
	TagKeyRunName    = "mlflow.runName"
	TagKeySourceName = "mlflow.source.name"
	TagKeySourceType = "mlflow.source.type"
)

// ConvertCreateRunRequestToDBModel converts request.CreateRunRequest into actual models.Run model.
func ConvertCreateRunRequestToDBModel(
	experiment *models.Experiment, req *request.CreateRunRequest,
) (*models.Run, error) {
	runID := database.NewUUID()
	artifactURI, err := url.JoinPath(experiment.ArtifactLocation, runID, "artifacts")
	if err != nil {
		return nil, eris.Wrap(err, "error constructing artifact_uri")
	}
	run := models.Run{
		ID:     runID,
		Name:   req.Name,
		Tags:   make([]models.Tag, len(req.Tags)),
		UserID: req.UserID,
		Status: models.StatusRunning,
		StartTime: sql.NullInt64{
			Int64: req.StartTime,
			Valid: true,
		},
		ArtifactURI:    artifactURI,
		ExperimentID:   *experiment.ID,
		LifecycleStage: models.LifecycleStageActive,
	}

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
	return &run, nil
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
