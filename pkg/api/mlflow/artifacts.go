package mlflow

import (
	"github.com/G-Research/fasttrack/pkg/database"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
)

func ListArtifacts(c *fiber.Ctx) error {
	q := struct {
		RunID string `query:"run_id"`
		Path  string `query:"path"`
		Token string `query:"token"`
	}{}

	if err := c.QueryParser(&q); err != nil {
		return NewError(ErrorCodeBadRequest, err.Error())
	}

	if q.RunID == "" {
		q.RunID = c.Query("run_uuid")
	}

	log.Debugf("ListArtifacts request: %#v", q)

	if q.RunID == "" {
		return NewError(ErrorCodeInvalidParameterValue, "Missing value for required parameter 'run_id'")
	}

	run := database.Run{
		ID: q.RunID,
	}

	if tx := database.DB.Select("artifact_uri").First(&run); tx.Error != nil {
		return NewError(ErrorCodeInternalError, "Unable to get artifact URI for run '%s'", q.RunID)
	}

	// TODO grab list of artifacts from S3
	resp := &ListArtifactsResponse{
		Files: make([]File, 0),
	}

	log.Debugf("ArtifactList response: %#v", resp)

	return c.JSON(resp)
}
