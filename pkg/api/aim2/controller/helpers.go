package controller

import (
	"github.com/gofiber/fiber/v2"

	"github.com/G-Research/fasttrackml/pkg/common/api"
)

// convertError converts api.ErrorResponse to fiber error.
func convertError(err error) error {
	switch v := err.(type) {
	case *api.ErrorResponse:
		if v.ErrorCode == api.ErrorCodeResourceDoesNotExist {
			return fiber.ErrNotFound
		}
	}
	return err
}
