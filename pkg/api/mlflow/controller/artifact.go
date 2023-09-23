package controller

import (
	"io"

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

	rootURI, artifacts, err := c.artifactService.ListArtifacts(ctx.Context(), &req)
	if err != nil {
		return err
	}

	resp := response.NewListArtifactsResponse(rootURI, artifacts)
	log.Debugf("artifactList response: %#v", resp)
	return ctx.JSON(resp)
}

// GetArtifact handles `GET /artifacts/get` endpoint.
func (c Controller) GetArtifact(ctx *fiber.Ctx) error {
	req := request.GetArtifactRequest{}
	if err := ctx.QueryParser(&req); err != nil {
		return api.NewBadRequestError(err.Error())
	}
	log.Debugf("GetArtifact request: %#v", req)

	artifact, err := c.artifactService.GetArtifact(ctx.Context(), &req)
	if err != nil {
		return err
	}
	defer artifact.Close()

	bytesWritten, err := io.CopyBuffer(ctx, artifact, make([]byte, 4096))
	log.Debugf("GetArtifact wrote bytes to output stream: %d", bytesWritten)
	return err
}
