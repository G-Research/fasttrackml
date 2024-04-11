package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/pkg/common/client/oidc"
	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
	"github.com/G-Research/fasttrackml/pkg/common/dao/repositories"
	"github.com/G-Research/fasttrackml/pkg/common/middleware"
)

// nolint:gosec
const (
	basicAuthTokenContextKey = "basic_auth_token"
)

// NewUserMiddleware creates new User based Middleware instance.
func NewUserMiddleware(userPermissions *models.UserPermissions) fiber.Handler {
	return func(ctx *fiber.Ctx) (err error) {
		namespace, err := middleware.GetNamespaceFromContext(ctx.Context())
		if err != nil {
			log.Errorf("error getting namespace from context: %+v", err)
			return ctx.Redirect("/errors/not-found", http.StatusMovedPermanently)
		}
		log.Debugf("checking access permission to %s namespace", namespace.Code)

		authToken := userPermissions.ValidateAuthToken(ctx.Get(fiber.HeaderAuthorization)[6:])
		if authToken != nil && authToken.HasAdminAccess() {
			ctx.Locals(basicAuthTokenContextKey, authToken)
			return ctx.Next()
		}
		if authToken == nil || !authToken.HasUserAccess(namespace.Code) {
			return ctx.Redirect("/errors/not-found", http.StatusMovedPermanently)
		}

		ctx.Locals(basicAuthTokenContextKey, authToken)
		return ctx.Next()
	}
}

// NewOIDCMiddleware creates new OIDC based Middleware instance.
func NewOIDCMiddleware(
	client oidc.ClientProvider, rolesRepository repositories.RoleRepositoryProvider,
) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if path := ctx.Path(); path != "/login" && !strings.Contains(path, "chooser/static") {
			authToken := ctx.Query("authToken", "")
			if authToken == "" {
				return ctx.Redirect("/login", http.StatusMovedPermanently)
			}

			user, err := client.Verify(ctx.Context(), authToken)
			if err != nil {
				return ctx.Redirect("/login", http.StatusMovedPermanently)
			}

			log.Debugf("user has roles: %v accociated", user.Roles())
			if user.IsAdmin() {
				return ctx.Next()
			}

			namespace, err := middleware.GetNamespaceFromContext(ctx.Context())
			if err != nil {
				return api.NewInternalError("error getting namespace from context")
			}
			log.Debugf("checking access permission to %s namespace", namespace.Code)

			isValid, err := rolesRepository.ValidateRolesAccessToNamespace(ctx.Context(), user.Roles(), namespace.Code)
			if err != nil {
				log.Errorf("error validating access to requested namespace with code: %s, %+v", namespace.Code, err)
				return ctx.Redirect("/errors/not-found", http.StatusMovedPermanently)
			}
			if !isValid {
				return ctx.Redirect("/errors/not-found", http.StatusMovedPermanently)
			}
		}

		return ctx.Next()
	}
}

// GetBasicAuthTokenFromContext returns Basic Auth Token from the context.
func GetBasicAuthTokenFromContext(ctx context.Context) (*models.BasicAuthToken, error) {
	authToken, ok := ctx.Value(basicAuthTokenContextKey).(*models.BasicAuthToken)
	if !ok {
		return nil, eris.New("error getting auth token from context")
	}
	return authToken, nil
}
