package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/config/auth"
	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/pkg/common/middleware"
)

const (
	authTokenContextKey = "auth_token"
)

// NewUserMiddleware creates new User based Middleware instance.
func NewUserMiddleware(userPermissions *auth.UserPermissions) fiber.Handler {
	return func(ctx *fiber.Ctx) (err error) {
		namespace, err := middleware.GetNamespaceFromContext(ctx.Context())
		if err != nil {
			return api.NewInternalError("error getting namespace from context")
		}
		log.Debugf("checking access permission to %s namespace", namespace.Code)

		// check that user has permissions to access to the requested namespace.
		authToken := strings.Replace(ctx.Get(fiber.HeaderAuthorization), "Basic ", "", 1)
		if !userPermissions.HasUserAccess(namespace.Code, authToken) {
			return ctx.Status(
				http.StatusNotFound,
			).JSON(
				api.NewResourceDoesNotExistError("unable to find requested resource"),
			)
		}

		ctx.Locals(authTokenContextKey, namespace)
		return ctx.Next()
	}
}

// GetAuthTokenFromContext returns Basic Auth Token from the context.
func GetAuthTokenFromContext(ctx context.Context) (string, error) {
	authToken, ok := ctx.Value(authTokenContextKey).(string)
	if !ok {
		return "", eris.New("error getting auth token from context")
	}
	return authToken, nil
}
