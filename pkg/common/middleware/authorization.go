package middleware

import (
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/config/auth"
	"github.com/G-Research/fasttrackml/pkg/common/api"
)

// NewRBACMiddleware creates new RBAC based Middleware instance.
func NewRBACMiddleware(permissions *auth.Permissions) fiber.Handler {
	return func(ctx *fiber.Ctx) (err error) {
		namespace, err := GetNamespaceFromContext(ctx.Context())
		if err != nil {
			return api.NewInternalError("error getting namespace from context")
		}
		log.Debugf("checking access permission to %s namespace", namespace.Code)

		// check that user has permissions to access to the requested namespace.
		authToken := strings.Replace(ctx.Get(fiber.HeaderAuthorization), "Basic ", "", 1)
		if !permissions.HasPermissions(namespace.Code, authToken) {
			return ctx.Status(
				http.StatusForbidden,
			).JSON(
				api.NewResourceAccessForbiddenError("access to %s namespace is forbidden", namespace.Code),
			)
		}

		return ctx.Next()
	}
}

// NewOIDCMiddleware creates new OIDC based Middleware instance.
func NewOIDCMiddleware() fiber.Handler {
	return func(ctx *fiber.Ctx) (err error) {
		namespace, err := GetNamespaceFromContext(ctx.Context())
		if err != nil {
			return api.NewInternalError("error getting namespace from context")
		}
		log.Debugf("checking access permission to %s namespace", namespace.Code)

		authToken := strings.Replace(ctx.Get(fiber.HeaderAuthorization), "Bearer ", "", 1)
		if authToken == "" {
			return ctx.Status(
				http.StatusBadRequest,
			).JSON(
				api.NewBadRequestError("authorization header is empty or incorrect"),
			)
		}

		return ctx.Next()
	}
}
