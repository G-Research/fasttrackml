package api

type DetailedError interface {
	error
	Detail() any
	Message() string
	Code() int
}

type ErrorResponse struct {
	Message string `json:"message"`
	Detail  any    `json:"detail"`
	Code    int    `json:"-"`
}

func (e *ErrorResponse) Error() string {
	return e.Message
}
