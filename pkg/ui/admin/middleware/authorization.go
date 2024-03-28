package middleware

import (
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/G-Research/fasttrackml/pkg/common/db/models"
)

// NewAdminUserMiddleware creates new User based Middleware instance.
func NewAdminUserMiddleware(userPermissions *models.UserPermissions) fiber.Handler {
	return func(ctx *fiber.Ctx) (err error) {
		authToken := userPermissions.ValidateAuthToken(
			strings.Replace(ctx.Get(fiber.HeaderAuthorization), "Basic ", "", 1),
		)
		if authToken == nil || !authToken.HasAdminAccess() {
			return ctx.Redirect("/admin/errors/not-found", http.StatusMovedPermanently)
		}

		return ctx.Next()
	}
}
