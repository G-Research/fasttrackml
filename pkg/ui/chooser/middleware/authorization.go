package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/pkg/common/db/models"
	"github.com/G-Research/fasttrackml/pkg/common/middleware"
)

// nolint:gosec
const (
	authTokenContextKey = "basic_auth_token"
)

// NewUserMiddleware creates new User based Middleware instance.
func NewUserMiddleware(userPermissions *models.UserPermissions) fiber.Handler {
	return func(ctx *fiber.Ctx) (err error) {
		namespace, err := middleware.GetNamespaceFromContext(ctx.Context())
		if err != nil {
			return api.NewInternalError("error getting namespace from context")
		}
		log.Debugf("checking access permission to %s namespace", namespace.Code)

		authToken := userPermissions.ValidateAuthToken(
			strings.Replace(ctx.Get(fiber.HeaderAuthorization), "Basic ", "", 1),
		)
		if authToken != nil && authToken.HasAdminAccess() {
			ctx.Locals(authTokenContextKey, authToken)
			return ctx.Next()
		}
		if authToken == nil || !authToken.HasUserAccess(namespace.Code) {
			return ctx.Redirect("/errors/not-found", http.StatusMovedPermanently)
		}

		ctx.Locals(authTokenContextKey, authToken)
		return ctx.Next()
	}
}

// GetAuthTokenFromContext returns Basic Auth Token from the context.
func GetAuthTokenFromContext(ctx context.Context) (*models.BasicAuthToken, error) {
	authToken, ok := ctx.Value(authTokenContextKey).(*models.BasicAuthToken)
	if !ok {
		return nil, eris.New("error getting auth token from context")
	}
	return authToken, nil
}
