package artifact

import (
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/response"
	"github.com/G-Research/fasttrackml/pkg/database"
)

func ListArtifacts(c *fiber.Ctx) error {
	req := request.ListArtifactsRequest{}
	if err := c.QueryParser(&req); err != nil {
		return api.NewBadRequestError(err.Error())
	}

	log.Debugf("ListArtifacts request: %#v", req)

	if err := ValidateListArtifactsRequest(&req); err != nil {
		return err
	}

	run := database.Run{ID: req.GetRunID()}
	if tx := database.DB.Select("artifact_uri").First(&run); tx.Error != nil {
		return api.NewInternalError("Unable to get artifact URI for run '%s'", req.GetRunID())
	}

	// TODO grab list of artifacts from S3
	resp := &response.ListArtifactsResponse{
		Files: make([]response.FilePartialResponse, 0),
	}

	log.Debugf("ArtifactList response: %#v", resp)

	return c.JSON(resp)
}
