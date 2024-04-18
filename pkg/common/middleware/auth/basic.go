package auth

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
	"github.com/G-Research/fasttrackml/pkg/common/middleware"
)

// nolint:gosec
const (
	basicAuthTokenContextKey = "basic_auth_token"
)

// BasicAuthMiddleware represents Basic Auth middleware.
type BasicAuthMiddleware struct {
	userPermissions *models.UserPermissions
}

// NewBasicAuthMiddleware creates new Basic Auth middleware logic.
func NewBasicAuthMiddleware(userPermissions *models.UserPermissions) fiber.Handler {
	return BasicAuthMiddleware{
		userPermissions: userPermissions,
	}.Handle()
}

// Handle handles OIDC middleware logic.
func (m BasicAuthMiddleware) Handle() fiber.Handler {
	return func(ctx *fiber.Ctx) (err error) {
		authToken := m.userPermissions.ValidateAuthToken(ctx.Get(fiber.HeaderAuthorization)[6:])
		switch {
		case AdminPrefixRegexp.MatchString(ctx.Path()):
			return m.handleAdminResourceRequest(ctx, authToken)
		case ChooserPrefixRegexp.MatchString(ctx.Path()):
			return m.handleChooserResourceRequest(ctx, authToken)
		case MlflowAimPrefixRegexp.MatchString(ctx.Path()):
			return m.handleAimMlflowResourceRequest(ctx, authToken)
		}
		return ctx.Next()
	}
}

// handleAdminResourceRequest applies Basic Auth check for Admin resources.
func (m BasicAuthMiddleware) handleAdminResourceRequest(ctx *fiber.Ctx, authToken *models.BasicAuthToken) error {
	if authToken == nil || !authToken.HasAdminAccess() {
		return ctx.Redirect("/errors/not-found", http.StatusMovedPermanently)
	}
	return ctx.Next()
}

// handleChooserResourceRequest applies Basic Auth check for Chooser resources.
func (m BasicAuthMiddleware) handleChooserResourceRequest(ctx *fiber.Ctx, authToken *models.BasicAuthToken) error {
	namespace, err := middleware.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return ctx.Redirect("/errors/not-found", http.StatusMovedPermanently)
	}
	log.Debugf("checking access permission to %s namespace", namespace.Code)
	if authToken == nil {
		return ctx.Redirect("/errors/not-found", http.StatusMovedPermanently)
	}
	if authToken.HasAdminAccess() {
		ctx.Locals(basicAuthTokenContextKey, authToken)
		return ctx.Next()
	}
	if !authToken.HasUserAccess(namespace.Code) {
		return ctx.Redirect("/errors/not-found", http.StatusMovedPermanently)
	}
	ctx.Locals(basicAuthTokenContextKey, authToken)
	return ctx.Next()
}

// handleAimMlflowResourceRequest applies Basic Auth check for Aim or Mlflow resources.
func (m BasicAuthMiddleware) handleAimMlflowResourceRequest(ctx *fiber.Ctx, authToken *models.BasicAuthToken) error {
	namespace, err := middleware.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("checking access permission to %s namespace", namespace.Code)
	if authToken == nil {
		return ctx.Status(
			http.StatusNotFound,
		).JSON(
			api.NewResourceDoesNotExistError("unable to find namespace with code: %s", namespace.Code),
		)
	}
	if !authToken.HasUserAccess(namespace.Code) && !authToken.HasAdminAccess() {
		return ctx.Status(
			http.StatusNotFound,
		).JSON(
			api.NewResourceDoesNotExistError("unable to find namespace with code: %s", namespace.Code),
		)
	}
	return ctx.Next()
}

// GetBasicAuthTokenFromContext returns Basic Auth Token from the context.
func GetBasicAuthTokenFromContext(ctx context.Context) (*models.BasicAuthToken, error) {
	authToken, ok := ctx.Value(basicAuthTokenContextKey).(*models.BasicAuthToken)
	if !ok {
		return nil, eris.New("error getting auth token from context")
	}
	return authToken, nil
}
