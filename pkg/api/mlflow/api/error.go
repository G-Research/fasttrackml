package api

import (
	"fmt"
	"net/http"
)

// ErrorResponse represents error response object.
type ErrorResponse struct {
	Message       string    `json:"message"`
	ErrorCode     ErrorCode `json:"error_code"`
	StatusCode    int       `json:"-"`
	OriginalError error     `json:"-"`
}

// Error provides error interface to be compatible with std errors.
func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode, e.Message)
}

// WithError stores the original error which actually caused the problem.
func (e *ErrorResponse) WithError(err error) {
	e.OriginalError = err
}

type ErrorCode string

const (
	ErrorCodeInternalError          = "INTERNAL_ERROR"
	ErrorCodeTemporarilyUnavailable = "TEMPORARILY_UNAVAILABLE"
	ErrorCodeBadRequest             = "BAD_REQUEST"
	ErrorCodeInvalidParameterValue  = "INVALID_PARAMETER_VALUE"
	ErrorCodeEndpointNotFound       = "ENDPOINT_NOT_FOUND"
	ErrorCodeResourceAlreadyExists  = "RESOURCE_ALREADY_EXISTS"
	ErrorCodeResourceDoesNotExist   = "RESOURCE_DOES_NOT_EXIST"
)

// NewBadRequestError creates new Response object with ErrorCodeBadRequest.
func NewBadRequestError(msg string, args ...any) *ErrorResponse {
	return &ErrorResponse{
		Message:    fmt.Sprintf(msg, args...),
		ErrorCode:  ErrorCodeBadRequest,
		StatusCode: http.StatusBadRequest,
	}
}

// NewInternalError creates new Response object with ErrorCodeInternalError.
func NewInternalError(msg string, args ...any) *ErrorResponse {
	return &ErrorResponse{
		Message:    fmt.Sprintf(msg, args...),
		ErrorCode:  ErrorCodeInternalError,
		StatusCode: http.StatusInternalServerError,
	}
}

// NewInvalidParameterValueError creates new Response object with ErrorCodeInternalError.
func NewInvalidParameterValueError(msg string, args ...any) *ErrorResponse {
	return &ErrorResponse{
		Message:    fmt.Sprintf(msg, args...),
		ErrorCode:  ErrorCodeInvalidParameterValue,
		StatusCode: http.StatusBadRequest,
	}
}

// NewResourceDoesNotExistError creates new Response object with ErrorCodeResourceDoesNotExist.
func NewResourceDoesNotExistError(msg string, args ...any) *ErrorResponse {
	return &ErrorResponse{
		Message:    fmt.Sprintf(msg, args...),
		ErrorCode:  ErrorCodeResourceDoesNotExist,
		StatusCode: http.StatusBadRequest,
	}
}

// NewResourceAlreadyExistsError creates new Response object with ErrorCodeResourceAlreadyExists.
func NewResourceAlreadyExistsError(msg string, args ...any) *ErrorResponse {
	return &ErrorResponse{
		Message:    fmt.Sprintf(msg, args...),
		ErrorCode:  ErrorCodeResourceAlreadyExists,
		StatusCode: http.StatusBadRequest,
	}
}

// NewEndpointNotFound creates new Response object with ErrorCodeEndpointNotFound.
func NewEndpointNotFound(msg string, args ...any) *ErrorResponse {
	return &ErrorResponse{
		Message:    fmt.Sprintf(msg, args...),
		ErrorCode:  ErrorCodeEndpointNotFound,
		StatusCode: http.StatusNotFound,
	}
}
