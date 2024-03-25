package middleware

import (
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/config/auth"
	"github.com/G-Research/fasttrackml/pkg/common/api"
)

// NewUserMiddleware creates new User based Middleware instance.
func NewUserMiddleware(userPermissions *auth.UserPermissions) fiber.Handler {
	return func(ctx *fiber.Ctx) (err error) {
		namespace, err := GetNamespaceFromContext(ctx.Context())
		if err != nil {
			return api.NewInternalError("error getting namespace from context")
		}
		log.Debugf("checking access permission to %s namespace", namespace.Code)

		// check that user has permissions to access to the requested namespace.
		authToken := strings.Replace(ctx.Get(fiber.HeaderAuthorization), "Basic ", "", 1)
		if userPermissions.HasAdminAccess(authToken) {
			return ctx.Next()
		}
		if !userPermissions.HasUserAccess(namespace.Code, authToken) {
			return ctx.Status(
				http.StatusNotFound,
			).JSON(
				api.NewResourceDoesNotExistError("unable to find namespace with code: %s", namespace.Code),
			)
		}

		return ctx.Next()
	}
}

// NewOIDCMiddleware creates new OIDC based Middleware instance.
func NewOIDCMiddleware() fiber.Handler {
	return func(ctx *fiber.Ctx) (err error) {
		return ctx.Next()
	}
}
