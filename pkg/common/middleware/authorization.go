package middleware

import (
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/common/api"
)

const (
	AuthorizationHeader = "Authorization"
)

// NewRoleAuthorizationMiddleware creates new Role based Middleware instance.
func NewRoleAuthorizationMiddleware(users map[string]struct{}) fiber.Handler {
	return func(ctx *fiber.Ctx) (err error) {
		namespace, err := GetNamespaceFromContext(ctx.Context())
		if err != nil {
			return api.NewInternalError("error getting namespace from context")
		}
		log.Debugf("checking acess permission to %s namespace", namespace.Code)

		authorization := strings.Replace(ctx.Get(AuthorizationHeader), "Basic ", "", 1)
		if authorization == "" {
			return ctx.Status(
				http.StatusBadRequest,
			).JSON(
				api.NewBadRequestError("authorization header is empty or incorrect"),
			)
		}

		if _, ok := users[authorization]; !ok {
			return ctx.Status(
				http.StatusForbidden,
			).JSON(
				api.NewResourceAccessForbiddenError("access to %s namespace is forbidden", namespace.Code),
			)
		}

		return ctx.Next()
	}
}

// NewOIDCAuthorizationMiddleware creates new OIDC based Middleware instance.
func NewOIDCAuthorizationMiddleware() fiber.Handler {
	return func(ctx *fiber.Ctx) (err error) {
		namespace, err := GetNamespaceFromContext(ctx.Context())
		if err != nil {
			return api.NewInternalError("error getting namespace from context")
		}
		log.Debugf("checking acess permission to %s namespace", namespace.Code)

		authorization := strings.Replace(ctx.Get(AuthorizationHeader), "Bearer ", "", 1)
		if authorization == "" {
			return ctx.Status(
				http.StatusBadRequest,
			).JSON(
				api.NewBadRequestError("authorization header is empty or incorrect"),
			)
		}

		return ctx.Next()
	}
}
