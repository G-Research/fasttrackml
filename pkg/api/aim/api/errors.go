package api

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
)

func ErrorHandler(c *fiber.Ctx, err error) error {
	var e *ErrorResponse
	var f *fiber.Error
	var d DetailedError

	switch {
	case errors.As(err, &e):
	case errors.As(err, &f):
		e = &ErrorResponse{
			Code:    f.Code,
			Message: f.Message,
			Detail:  "",
		}
	case errors.As(err, &d):
		e = &ErrorResponse{
			Code:    d.Code(),
			Message: d.Message(),
			Detail:  d.Detail(),
		}
	default:
		e = &ErrorResponse{
			Code:    fiber.StatusInternalServerError,
			Message: err.Error(),
			Detail:  "",
		}
	}

	fn := log.Errorf

	switch e.Code {
	case fiber.StatusNotFound:
		fn = log.Debugf
	case fiber.StatusInternalServerError:
	default:
		fn = log.Warnf
	}

	fn("Error encountered in %s %s: %s", c.Method(), c.Path(), err)

	return c.Status(e.Code).JSON(e)
}
