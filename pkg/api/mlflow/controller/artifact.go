package controller

import (
	"bufio"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
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

	filename := filepath.Base(req.Path)
	ctx.Set("Content-Type", common.GetContentType(filename))
	ctx.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	ctx.Set("X-Content-Type-Options", "nosniff")
	ctx.Context().Response.SetBodyStreamWriter(func(w *bufio.Writer) {
		defer artifact.Close()

		start := time.Now()
		if err := func() error {
			bytesWritten, err := io.CopyBuffer(w, artifact, make([]byte, 4096))
			if err != nil {
				return fmt.Errorf("unable to copy artifact Reader to output stream: %w", err)
			}
			if err := w.Flush(); err != nil {
				return fmt.Errorf("GetArtifact error encountered flushing stream: %w", err)
			}
			log.Debugf("GetArtifact wrote bytes to output stream: %d", bytesWritten)
			return nil
		}(); err != nil {
			log.Errorf(
				"error encountered in %s %s: error streaming artifact: %s",
				ctx.Method(),
				ctx.Path(),
				err,
			)
		}
		log.Infof("body - %s %s %s", time.Since(start), ctx.Method(), ctx.Path())
	})
	return nil
}
