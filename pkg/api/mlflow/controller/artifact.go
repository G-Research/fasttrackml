package controller

import (
	"bufio"
	"fmt"
	"io"
	"mime"
	"slices"
	"strings"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/response"
)

var textTypes = []string{
	"txt",
	"log",
	"err",
	"cfg",
	"conf",
	"cnf",
	"cf",
	"ini",
	"properties",
	"prop",
	"hocon",
	"toml",
	"yaml",
	"yml",
	"xml",
	"json",
	"js",
	"py",
	"py3",
	"csv",
	"tsv",
	"md",
	"rst",
	"MLmodel",
	"mlproject",
}

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

	filename := getFilename(req.Path)
	ctx.Set("Content-Type", getContentType(filename))
	ctx.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	ctx.Set("X-Content-Type-Options", "nosniff")
	ctx.Context().Response.SetBodyStreamWriter(func(w *bufio.Writer) {
		defer artifact.Close()
		bytesWritten, err := io.CopyBuffer(w, artifact, make([]byte, 4096))
		if err != nil {
			log.Errorf("error encountered in %s %s: error streaming artifact: %s", ctx.Method(), ctx.Path(), err)
		}
		log.Debugf("GetArtifact wrote bytes to output stream: %d", bytesWritten)
		w.Flush()
	})

	return nil
}

// getFilename returns the final bit of the path (the filename).
func getFilename(path string) string {
	pathParts := strings.Split(path, "/")
	return pathParts[len(pathParts)-1]
}

// getContentType will determine the content type of the file.
func getContentType(filename string) string {
	fileParts := strings.Split(filename, ".")
	fileExt := fileParts[len(fileParts)-1]
	if slices.Contains(textTypes, fileExt) {
		return "text/plain"
	}
	mimeType := mime.TypeByExtension("." + fileExt)
	if mimeType != "" {
		return mimeType
	}
	return "application/octet-stream"
}
