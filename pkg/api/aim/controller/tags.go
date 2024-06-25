package controller

import (
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/api/aim/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim/api/response"
	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/pkg/common/middleware"
)

// GetTags fetches run tags for the current namespace.
func (c Controller) GetTags(ctx *fiber.Ctx) error {
	ns, err := middleware.GetNamespaceFromContext(ctx.Context())
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

// CreateTag handles `POST /tags` endpoint.
func (c Controller) CreateTag(ctx *fiber.Ctx) error {
	ns, err := middleware.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("createTag namespace: %s", ns.Code)
	req := request.CreateTagRequest{}
	if err := ctx.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}
	tag, err := c.tagService.Create(ctx.Context(), ns.ID, &req)
	if err != nil {
		return convertError(err)
	}

	resp := response.NewCreateTagResponse(tag)
	log.Debugf("createTag response %#v", resp)
	return ctx.Status(fiber.StatusCreated).JSON(resp)
}

// GetTag handles `GET /tags/:id` endpoint.
func (c Controller) GetTag(ctx *fiber.Ctx) error {
	ns, err := middleware.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("getTag namespace: %s", ns.Code)

	req := request.GetTagRequest{}
	if err := ctx.ParamsParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}
	tag, err := c.tagService.Get(ctx.Context(), ns.ID, &req)
	if err != nil {
		return convertError(err)
	}

	resp := response.NewGetTagResponse(tag)
	log.Debugf("getTag response %#v", resp)
	return ctx.JSON(resp)
}

// UpdateTag handles `PUT /tags/:id` endpoint.
func (c Controller) UpdateTag(ctx *fiber.Ctx) error {
	ns, err := middleware.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("updateTag namespace: %s", ns.Code)

	req := request.UpdateTagRequest{}
	if err := ctx.ParamsParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}
	if err := ctx.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}
	tag, err := c.tagService.Update(ctx.Context(), ns.ID, &req)
	if err != nil {
		return convertError(err)
	}

	resp := response.NewUpdateTagResponse(tag)
	log.Debugf("updateTag response %#v", resp)
	return ctx.JSON(resp)
}

// DeleteTag handles `DELETE /tags/:id` endpoint.
func (c Controller) DeleteTag(ctx *fiber.Ctx) error {
	ns, err := middleware.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("deleteTag namespace: %s", ns.Code)

	req := request.DeleteTagRequest{}
	if err := ctx.ParamsParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}
	err = c.tagService.Delete(ctx.Context(), ns.ID, &req)
	if err != nil {
		return convertError(err)
	}
	log.Debugf("deleteTag response: %#v", fiber.StatusOK)
	return ctx.SendStatus(fiber.StatusOK)
}

func (c Controller) GetRunsTagged(ctx *fiber.Ctx) error {
	ns, err := middleware.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("deleteTag namespace: %s", ns.Code)

	req := request.GetRunsTaggedRequest{}
	if err := ctx.ParamsParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}
	tag, err := c.tagService.Get(ctx.Context(), ns.ID, &req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	resp := response.NewGetRunsTaggedResponse(tag)
	log.Debugf("getRunsTagged response: %#v", resp)
	return ctx.Status(fiber.StatusOK).JSON(resp)
}
