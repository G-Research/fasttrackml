package mlflow

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
)

func ErrorHandler(c *fiber.Ctx, err error) error {
	var e *ErrorResponse
	if !errors.As(err, &e) {
		var code ErrorCode = ErrorCodeInternalError

		var f *fiber.Error
		if errors.As(err, &f) {
			switch f.Code {
			case fiber.StatusBadRequest:
				code = ErrorCodeBadRequest
			case fiber.StatusServiceUnavailable:
				code = ErrorCodeTemporarilyUnavailable
			case fiber.StatusNotFound:
				code = ErrorCodeEndpointNotFound
			}
		}

		e = &ErrorResponse{
			ErrorCode: code,
			Message:   err.Error(),
		}
	}

	var code int
	var fn func(format string, args ...any)

	switch e.ErrorCode {
	case ErrorCodeBadRequest, ErrorCodeInvalidParameterValue, ErrorCodeResourceAlreadyExists:
		code = fiber.StatusBadRequest
		fn = log.Infof
	case ErrorCodeTemporarilyUnavailable:
		code = fiber.StatusServiceUnavailable
		fn = log.Warnf
	case ErrorCodeEndpointNotFound, ErrorCodeResourceDoesNotExist:
		code = fiber.StatusNotFound
		fn = log.Debugf
	default:
		code = fiber.StatusInternalServerError
		fn = log.Errorf
	}

	fn("Error encountered in %s %s: %s", c.Method(), c.Path(), err)

	return c.Status(code).JSON(e)
}
