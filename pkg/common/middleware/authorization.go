package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/common/api"
)

// NewRBACAuthorizationMiddleware creates new RBAC based Middleware instance.
func NewRBACAuthorizationMiddleware(config map[string]map[string]struct{}) fiber.Handler {
	return func(ctx *fiber.Ctx) (err error) {
		namespace, err := GetNamespaceFromContext(ctx.Context())
		if err != nil {
			return api.NewInternalError("error getting namespace from context")
		}
		log.Debugf("checking acess permission to %s namespace", namespace.Code)

		authorization := strings.Replace(ctx.Get(fiber.HeaderAuthorization), "Basic ", "", 1)
		if authorization == "" {
			return ctx.Status(
				http.StatusBadRequest,
			).JSON(
				api.NewBadRequestError("authorization header is empty or incorrect"),
			)
		}

		// check that requested Authorization token actually exists in user list and get the user roles.
		roles, ok := config[authorization]
		if !ok {
			return ctx.Status(
				http.StatusForbidden,
			).JSON(
				api.NewResourceAccessForbiddenError("access to %s namespace is forbidden", namespace.Code),
			)
		}

		// if role list contains `admin` then move forward immediately. `admin` has access to all namespaces.
		if _, ok := roles["admin"]; ok {
			return ctx.Next()
		}

		// if user is not admin, then check that this user has permission to access to the requested namespace.
		if _, ok := roles[fmt.Sprintf("ns:%s", namespace.Code)]; !ok {
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

		authorization := strings.Replace(ctx.Get(fiber.HeaderAuthorization), "Bearer ", "", 1)
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
