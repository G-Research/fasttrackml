package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/pkg/common/auth/oidc"
	"github.com/G-Research/fasttrackml/pkg/common/dao/repositories"
)

// nolint:gosec
const (
	oidcUserContextKey = "oidc_user"
)

// OIDCMiddleware represents OIDC middleware.
type OIDCMiddleware struct {
	client          oidc.ClientProvider
	rolesRepository repositories.RoleRepositoryProvider
}

// NewOIDCMiddleware creates new OIDC middleware logic.
func NewOIDCMiddleware(
	client oidc.ClientProvider,
	rolesRepository repositories.RoleRepositoryProvider,
) fiber.Handler {
	return OIDCMiddleware{
		client:          client,
		rolesRepository: rolesRepository,
	}.Handle()
}

// Handle handles OIDC middleware logic.
func (m OIDCMiddleware) Handle() fiber.Handler {
	return func(ctx *fiber.Ctx) (err error) {
		path := ctx.Path()
		// if requested resource related to something static, then we don't need to apply auth.
		if !strings.Contains(path, "static") {
			switch {
			case AdminPrefixRegexp.MatchString(path):
				return m.handleAdminResourceRequest(ctx)
			case ChooserPrefixRegexp.MatchString(path):
				return m.handleChooserResourceRequest(ctx)
			case MlflowAimPrefixRegexp.MatchString(path):
				return m.handleAimMlflowResourceRequest(ctx)
			}
		}

		return ctx.Next()
	}
}

// handleAdminResourceRequest applies OIDC check for Admin resources.
func (m OIDCMiddleware) handleAdminResourceRequest(ctx *fiber.Ctx) error {
	user, err := m.client.Verify(ctx.Context(), ctx.Cookies("access_token", ""))
	if err != nil {
		log.Errorf("error verifying access token: %+v", err)
		ctx.Response().Header.Add("Cache-Control", "no-store")
		return ctx.Redirect("/login", http.StatusMovedPermanently)
	}

	log.Debugf("user has roles: %v associated", user.GetRoles())
	if !user.IsAdmin() {
		return ctx.Redirect("/errors/not-found", http.StatusMovedPermanently)
	}
	return ctx.Next()
}

// handleChooserResourceRequest applies OIDC check for Chooser resources.
func (m OIDCMiddleware) handleChooserResourceRequest(ctx *fiber.Ctx) error {
	namespace, err := GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return ctx.Redirect("/errors/not-found", http.StatusMovedPermanently)
	}
	log.Debugf("checking access permission to %s namespace", namespace.Code)
	user, err := m.client.Verify(ctx.Context(), ctx.Cookies("access_token", ""))
	if err != nil {
		log.Errorf("error verifying access token: %+v", err)
		ctx.Response().Header.Add("Cache-Control", "no-store")
		return ctx.Redirect("/login", http.StatusMovedPermanently)
	}
	log.Debugf("user has roles: %v associated", user.GetRoles())
	ctx.Locals(oidcUserContextKey, user)
	return ctx.Next()
}

// handleAimMlflowResourceRequest applies OIDC check for Aim or Mlflow resources.
func (m OIDCMiddleware) handleAimMlflowResourceRequest(ctx *fiber.Ctx) error {
	namespace, err := GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("checking access permission to %s namespace", namespace.Code)

	user, err := m.client.Verify(ctx.Context(), ctx.Cookies("access_token", ""))
	if err != nil {
		return ctx.Status(
			http.StatusUnauthorized,
		).JSON(
			api.NewResourceDoesNotExistError("unable to find namespace with code: %s", namespace.Code),
		)
	}
	log.Debugf("user has roles: %v associated", user.GetRoles())

	if user.IsAdmin() {
		return ctx.Next()
	}

	isValid, err := m.rolesRepository.ValidateRolesAccessToNamespace(ctx.Context(), user.GetRoles(), namespace.Code)
	if err != nil {
		log.Errorf("error validating access to requested namespace with code: %s, %+v", namespace.Code, err)
		return api.NewInternalError(
			"error validating access to requested namespace with code: %s", namespace.Code,
		)
	}
	if !isValid {
		return ctx.Status(
			http.StatusForbidden,
		).JSON(
			api.NewResourceDoesNotExistError("unable to find namespace with code: %s", namespace.Code),
		)
	}
	return ctx.Next()
}

// GetOIDCUserFromContext returns OIDC User object from the context.
func GetOIDCUserFromContext(ctx context.Context) (*oidc.User, error) {
	user, ok := ctx.Value(oidcUserContextKey).(*oidc.User)
	if !ok {
		return nil, eris.New("error getting oidc user object from context")
	}
	return user, nil
}
