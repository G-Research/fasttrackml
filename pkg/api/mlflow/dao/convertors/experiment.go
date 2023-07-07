package convertors

import (
	"database/sql"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

// ConvertCreateExperimentToDBModel converts request.CreateExperimentRequest into actual models.Experiment model.
func ConvertCreateExperimentToDBModel(req *request.CreateExperimentRequest) (*models.Experiment, error) {
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
		ArtifactLocation: strings.TrimRight(req.ArtifactLocation, "/"),
	}

	for n, tag := range req.Tags {
		experiment.Tags[n] = models.ExperimentTag{
			Key:   tag.Key,
			Value: tag.Value,
		}
	}

	u, err := url.Parse(req.ArtifactLocation)
	if err != nil {
		return nil, api.NewInvalidParameterValueError("Invalid value for parameter 'artifact_location': %s", err)
	}

	switch u.Scheme {
	case "s3":
		artifactLocation := fmt.Sprintf("%s://%s", u.Scheme, strings.Trim(u.Host, "/"))
		if path := strings.Trim(u.Path, "/"); len(path) > 0 {
			artifactLocation = fmt.Sprintf("%s/%s", artifactLocation, path)
		}
		experiment.ArtifactLocation = artifactLocation
	default:
		// TODO:DSuhinin - default case right now has to satisfy Python integration tests.
		p, err := filepath.Abs(u.Path)
		if err != nil {
			return nil, api.NewInvalidParameterValueError("Invalid value for parameter 'artifact_location': %s", err)
		}
		u.Path = p
		experiment.ArtifactLocation = u.String()
	}

	return &experiment, nil
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
