package mlflow

import (
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrack/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrack/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrack/pkg/api/mlflow/api/response"
	"github.com/G-Research/fasttrack/pkg/database"
)

func ListArtifacts(c *fiber.Ctx) error {
	query := request.ListArtifactsRequest{}
	if err := c.QueryParser(&query); err != nil {
		return api.NewBadRequestError(err.Error())
	}

	if query.RunID == "" {
		query.RunID = c.Query("run_uuid")
	}

	log.Debugf("ListArtifacts request: %#v", query)

	if query.RunID == "" {
		return api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'")
	}

	run := database.Run{
		ID: query.RunID,
	}

	if tx := database.DB.Select("artifact_uri").First(&run); tx.Error != nil {
		return api.NewInternalError("Unable to get artifact URI for run '%s'", query.RunID)
	}

	// TODO grab list of artifacts from S3
	resp := &response.ListArtifactsResponse{
		Files: make([]response.FilePartialResponse, 0),
	}

	log.Debugf("ArtifactList response: %#v", resp)

	return c.JSON(resp)
}
