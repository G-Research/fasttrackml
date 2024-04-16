package middleware

import (
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/pkg/common/db/models"
)

// NewUserMiddleware creates new User based Middleware instance.
func NewUserMiddleware(userPermissions *models.UserPermissions) fiber.Handler {
	return func(ctx *fiber.Ctx) (err error) {
		namespace, err := GetNamespaceFromContext(ctx.Context())
		if err != nil {
			return api.NewInternalError("error getting namespace from context")
		}
		log.Debugf("checking access permission to %s namespace", namespace.Code)

		// check that user has permissions to access to the requested namespace.
		authToken := userPermissions.ValidateAuthToken(
			strings.Replace(ctx.Get(fiber.HeaderAuthorization), "Basic ", "", 1),
		)
		if authToken != nil && authToken.HasAdminAccess() {
			return ctx.Next()
		}
		if authToken == nil || !authToken.HasUserAccess(namespace.Code) {
			return ctx.Status(
				http.StatusNotFound,
			).JSON(
				api.NewResourceDoesNotExistError("unable to find namespace with code: %s", namespace.Code),
			)
		}

		return ctx.Next()
	}
}
