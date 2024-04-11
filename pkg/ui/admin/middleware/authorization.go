package middleware

import (
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/common/client/oidc"
	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
)

// NewAdminUserMiddleware creates new User based Middleware instance.
func NewAdminUserMiddleware(userPermissions *models.UserPermissions) fiber.Handler {
	return func(ctx *fiber.Ctx) (err error) {
		authToken := userPermissions.ValidateAuthToken(
			strings.Replace(ctx.Get(fiber.HeaderAuthorization), "Basic ", "", 1),
		)
		if authToken == nil || !authToken.HasAdminAccess() {
			return ctx.Redirect("/errors/not-found", http.StatusMovedPermanently)
		}

		return ctx.Next()
	}
}

// NewOIDCMiddleware creates new OIDC based Middleware instance.
func NewOIDCMiddleware(client oidc.ClientProvider) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		authToken := ctx.Query("authToken", "")
		if authToken == "" {
			return ctx.Redirect("/login", http.StatusMovedPermanently)
		}

		user, err := client.Verify(ctx.Context(), authToken)
		if err != nil {
			return ctx.Redirect("/login", http.StatusMovedPermanently)
		}

		log.Debugf("user has roles: %v accociated", user.Roles())
		if !user.IsAdmin() {
			return ctx.Redirect("/login", http.StatusMovedPermanently)
		}
		return ctx.Next()
	}
}
