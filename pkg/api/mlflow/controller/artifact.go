package controller

import (
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/response"
)

// ListArtifacts handles `GET /artifacts/list` endpoint.
func (c Controller) ListArtifacts(ctx *fiber.Ctx) error {
	req := request.ListArtifactsRequest{}
	if err := ctx.QueryParser(&req); err != nil {
		return api.NewBadRequestError(err.Error())
	}
	log.Debugf("listArtifacts request: %#v", req)

	if err := c.artifactService.ListArtifacts(ctx.Context(), &req); err != nil {
		return err
	}

	// TODO grab list of artifacts from S3
	resp := &response.ListArtifactsResponse{
		Files: make([]response.FilePartialResponse, 0),
	}

	log.Debugf("artifactList response: %#v", resp)
	return ctx.JSON(resp)
}
