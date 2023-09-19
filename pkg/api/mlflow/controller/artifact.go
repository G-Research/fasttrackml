package controller

import (
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/response"
	"github.com/G-Research/fasttrackml/pkg/common/middleware/namespace"
)

// ListArtifacts handles `GET /artifacts/list` endpoint.
func (c Controller) ListArtifacts(ctx *fiber.Ctx) error {
	req := request.ListArtifactsRequest{}
	if err := ctx.QueryParser(&req); err != nil {
		return api.NewBadRequestError(err.Error())
	}
	log.Debugf("listArtifacts request: %#v", req)

	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("listArtifacts namespace: %s", ns.Code)

	rootURI, artifacts, err := c.artifactService.ListArtifacts(ctx.Context(), ns, &req)
	if err != nil {
		return err
	}

	resp := response.NewListArtifactsResponse(rootURI, artifacts)
	log.Debugf("artifactList response: %#v", resp)
	return ctx.JSON(resp)
}


// GetArtifact handles `GET /get-artifact` endpoint.
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
	log.Debugf("GetArtifact namespace: %s", ns.Code)

	artifactURI, err := c.artifactService.GetArtifact(ctx.Context(), ns, &req)
	if err != nil {
		return err
	}

	log.Debugf("artifactURI redirect: %#v", artifactURI)
	return ctx.Redirect(artifactURI.String(), fiber.StatusMovedPermanently)
}
