package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	jsonContentType *regexp.Regexp = regexp.MustCompile("^application/json;?")
)

func NewError(e ErrorCode, msg string, args ...interface{}) *ErrorResponse {
	return &ErrorResponse{
		ErrorCode: e,
		Message:   fmt.Sprintf(msg, args...),
	}
}

func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode, e.Message)
}

type HandlerFunc func(http.ResponseWriter, *http.Request) any

func EnsureJson(f HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) any {
		if !jsonContentType.MatchString(r.Header.Get("Content-Type")) {
			return NewError(ErrorCodeBadRequest, "Invalid Content-Type '%s'", r.Header.Get("Content-Type"))
		}
		return f(w, r)
	}
}

func EnsureMethod(f HandlerFunc, methods ...string) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) any {
		valid := false
		for _, m := range methods {
			if r.Method == m {
				valid = true
			}
		}
		if !valid {
			return NewError(ErrorCodeBadRequest, "Invalid method '%s'", r.Method)
		}
		return f(w, r)
	}
}

func ReturnJson(f HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		v := f(w, r)
		w.Header().Set("Content-Type", "application/json")
		if v, ok := v.(error); ok {
			if _, ok := v.(*ErrorResponse); !ok {
				v = &ErrorResponse{
					ErrorCode: ErrorCodeInternalError,
					Message:   v.Error(),
				}
			}
			code := http.StatusInternalServerError
			switch v.(*ErrorResponse).ErrorCode {
			case ErrorCodeInternalError, ErrorCodeInvalidState:
				code = http.StatusInternalServerError
			case ErrorCodeBadRequest, ErrorCodeInvalidParameterValue, ErrorCodeResourceAlreadyExists:
				code = http.StatusBadRequest
			case ErrorCodeTemporarilyUnavailable:
				code = http.StatusServiceUnavailable
			case ErrorCodeEndpointNotFound, ErrorCodeResourceDoesNotExist:
				code = http.StatusNotFound
			}
			log.Error(v)
			w.WriteHeader(code)
		}
		if v == nil {
			v = struct{}{}
		}
		if err := json.NewEncoder(w).Encode(v); err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		elapsed := time.Since(start)
		log.Infof("Elapsed time: %s", elapsed)
	}
}

type ServeMux struct {
	*http.ServeMux
}

func NewServeMux() *ServeMux {
	return &ServeMux{http.NewServeMux()}
}

func (m *ServeMux) HandleFunc(p string, h HandlerFunc) {
	m.ServeMux.HandleFunc(p, ReturnJson(h))
}
