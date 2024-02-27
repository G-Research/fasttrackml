package aim2

import (
	"errors"

	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
)

func ErrorHandler(c *fiber.Ctx, err error) error {
	var e *api.ErrorResponse
	var f *fiber.Error

	switch {
	case errors.As(err, &f):
		e = &api.ErrorResponse{
			StatusCode: f.Code,
			Message:    f.Message,
		}
	case errors.As(err, &e):
	default:
		e = &api.ErrorResponse{
			StatusCode: fiber.StatusInternalServerError,
			Message:    err.Error(),
		}
	}

	fn := log.Errorf

	switch e.StatusCode {
	case fiber.StatusNotFound:
		fn = log.Debugf
	case fiber.StatusInternalServerError:
	default:
		fn = log.Warnf
	}

	fn("Error encountered in %s %s: %s", c.Method(), c.Path(), err)

	return c.Status(e.StatusCode).JSON(e)
}
