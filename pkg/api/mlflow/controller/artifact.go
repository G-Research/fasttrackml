package controller

import (
	"bufio"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/common/middleware/namespace"
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

	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("getArtifact namespace: %s", ns.Code)

	artifact, err := c.artifactService.GetArtifact(ctx.Context(), ns, &req)
	if err != nil {
		return err
	}

	filename := filepath.Base(req.Path)
	ctx.Set("Content-Type", common.GetContentType(filename))
	ctx.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	ctx.Set("X-Content-Type-Options", "nosniff")
	ctx.Context().Response.SetBodyStreamWriter(func(w *bufio.Writer) {
		//nolint:errcheck
		defer artifact.Close()

		start := time.Now()
		if err := func() error {
			bytesWritten, err := io.CopyBuffer(w, artifact, make([]byte, 4096))
			if err != nil {
				return eris.Wrap(err, "error copying artifact Reader to output stream")
			}
			if err := w.Flush(); err != nil {
				return eris.Wrap(err, "error flushing output stream")
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
