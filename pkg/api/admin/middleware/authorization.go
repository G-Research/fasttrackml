package middleware

import (
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/config/auth"
	"github.com/G-Research/fasttrackml/pkg/common/api"
)

// NewAdminUserMiddleware creates new User based Middleware instance.
func NewAdminUserMiddleware(userPermissions *auth.UserPermissions) fiber.Handler {
	return func(ctx *fiber.Ctx) (err error) {
		// check that user has permissions to access to the requested namespace.
		authToken := strings.Replace(ctx.Get(fiber.HeaderAuthorization), "Basic ", "", 1)
		if !userPermissions.HasAdminAccess(authToken) {
			return ctx.Status(
				http.StatusNotFound,
			).JSON(
				api.NewResourceDoesNotExistError("unable to find requested resource"),
			)
		}

		return ctx.Next()
	}
}
