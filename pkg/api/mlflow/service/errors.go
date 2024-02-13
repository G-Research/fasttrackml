package service

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/common/api"
)

func ErrorHandler(c *fiber.Ctx, err error) error {
	var e *api.ErrorResponse
	if !errors.As(err, &e) {
		var code api.ErrorCode = api.ErrorCodeInternalError

		var f *fiber.Error
		if errors.As(err, &f) {
			switch f.Code {
			case fiber.StatusBadRequest:
				code = api.ErrorCodeBadRequest
			case fiber.StatusServiceUnavailable:
				code = api.ErrorCodeTemporarilyUnavailable
			case fiber.StatusNotFound:
				code = api.ErrorCodeEndpointNotFound
			}
		}

		e = &api.ErrorResponse{
			ErrorCode: code,
			Message:   err.Error(),
		}
	}

	var code int
	var fn func(format string, args ...any)

	switch e.ErrorCode {
	case api.ErrorCodeBadRequest, api.ErrorCodeInvalidParameterValue, api.ErrorCodeResourceAlreadyExists:
		code = fiber.StatusBadRequest
		fn = log.Infof
	case api.ErrorCodeTemporarilyUnavailable:
		code = fiber.StatusServiceUnavailable
		fn = log.Warnf
	case api.ErrorCodeEndpointNotFound, api.ErrorCodeResourceDoesNotExist:
		code = fiber.StatusNotFound
		fn = log.Debugf
	default:
		code = fiber.StatusInternalServerError
		fn = log.Errorf
	}

	fn("Error encountered in %s %s: %s", c.Method(), c.Path(), err)

	return c.Status(code).JSON(e)
}
