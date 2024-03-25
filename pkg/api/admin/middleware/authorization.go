package middleware

import (
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/pkg/common/db/models"
)

// NewAdminUserMiddleware creates new User based Middleware instance.
func NewAdminUserMiddleware(userPermissions *models.UserPermissions) fiber.Handler {
	return func(ctx *fiber.Ctx) (err error) {
		// check that user has permissions to access to the requested namespace.
		authToken := strings.Replace(ctx.Get(fiber.HeaderAuthorization), "Basic ", "", 1)
		authTokenInfo, isValid := userPermissions.ValidateAuthToken(authToken)
		if !isValid || !authTokenInfo.HasAdminAccess() {
			return ctx.Status(
				http.StatusNotFound,
			).JSON(
				api.NewResourceDoesNotExistError("unable to find requested resource"),
			)
		}

		return ctx.Next()
	}
}
