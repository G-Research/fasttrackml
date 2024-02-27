package controller

import (
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/api/aim2/api/response"
	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/pkg/common/middleware/namespace"
)

// GetTags fetches run tags for the current namespace.
func (c Controller) GetTags(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("getTags namespace: %s", ns.Code)

	tags, err := c.tagService.GetTags(ctx.Context(), ns.ID)
	if err != nil {
		return err
	}

	resp := response.NewGetTagsResponse(tags)
	log.Debugf("getTags response: %#v", resp)

	return ctx.JSON(resp)
}
