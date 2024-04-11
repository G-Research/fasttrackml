package middleware

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/pkg/common/client/oidc"
	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
	"github.com/G-Research/fasttrackml/pkg/common/dao/repositories"
)

// regexps to detect requested API.
var (
	AdminPrefixRegexp     = regexp.MustCompile(`/admin`)
	ChooserPrefixRegexp   = regexp.MustCompile(`/chooser`)
	MlflowAimPrefixRegexp = regexp.MustCompile(`/aim/api|/ajax-api/2.0/mlflow|/api/2.0/mlflow`)
)

// nolint:gosec
const (
	basicAuthTokenContextKey = "basic_auth_token"
)

// NewUserMiddleware creates new User based Middleware instance.
func NewUserMiddleware(userPermissions *models.UserPermissions) fiber.Handler {
	return func(ctx *fiber.Ctx) (err error) {
		namespace, err := GetNamespaceFromContext(ctx.Context())
		if err != nil {
			return api.NewInternalError("error getting namespace from context")
		}
		log.Debugf("checking access permission to %s namespace", namespace.Code)

		authToken := userPermissions.ValidateAuthToken(ctx.Get(fiber.HeaderAuthorization)[6:])
		// based on requested resource check permissions.
		switch {
		case AdminPrefixRegexp.MatchString(ctx.Path()):
			if authToken == nil || !authToken.HasAdminAccess() {
				return ctx.Redirect("/errors/not-found", http.StatusMovedPermanently)
			}
		case ChooserPrefixRegexp.MatchString(ctx.Path()):
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
		case MlflowAimPrefixRegexp.MatchString(ctx.Path()):
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
		}
		return ctx.Next()
	}
}

// NewOIDCMiddleware creates new OIDC based Middleware instance.
func NewOIDCMiddleware(
	client oidc.ClientProvider, rolesRepository repositories.RoleRepositoryProvider,
) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		namespace, err := GetNamespaceFromContext(ctx.Context())
		if err != nil {
			return api.NewInternalError("error getting namespace from context")
		}
		log.Debugf("checking role access to %s namespace", namespace.Code)

		// based on requested resource check permissions.
		switch {
		case AdminPrefixRegexp.MatchString(ctx.Path()):
			user, err := client.Verify(ctx.Context(), ctx.Query("authToken", ""))
			if err != nil {
				return ctx.Redirect("/login", http.StatusMovedPermanently)
			}

			log.Debugf("user has roles: %v accociated", user.Roles())
			if !user.IsAdmin() {
				return ctx.Redirect("/login", http.StatusMovedPermanently)
			}
		case ChooserPrefixRegexp.MatchString(ctx.Path()):
			if path := ctx.Path(); path != "/login" && !strings.Contains(path, "/chooser/static") {
				user, err := client.Verify(ctx.Context(), ctx.Query("authToken", ""))
				if err != nil {
					return ctx.Redirect("/login", http.StatusMovedPermanently)
				}

				log.Debugf("user has roles: %v accociated", user.Roles())
				if user.IsAdmin() {
					return ctx.Next()
				}

				isValid, err := rolesRepository.ValidateRolesAccessToNamespace(
					ctx.Context(), user.Roles(), namespace.Code,
				)
				if err != nil {
					log.Errorf("error validating access to requested namespace with code: %s, %+v", namespace.Code, err)
					return ctx.Redirect("/errors/not-found", http.StatusMovedPermanently)
				}
				if !isValid {
					return ctx.Redirect("/errors/not-found", http.StatusMovedPermanently)
				}
			}
		case MlflowAimPrefixRegexp.MatchString(ctx.Path()):
			user, err := client.Verify(ctx.Context(), ctx.Get(fiber.HeaderAuthorization)[7:])
			if err != nil {
				return ctx.Status(
					http.StatusNotFound,
				).JSON(
					api.NewResourceDoesNotExistError("unable to find namespace with code: %s", namespace.Code),
				)
			}
			log.Debugf("user has roles: %v accociated", user.Roles())

			if user.IsAdmin() {
				return ctx.Next()
			}

			isValid, err := rolesRepository.ValidateRolesAccessToNamespace(ctx.Context(), user.Roles(), namespace.Code)
			if err != nil {
				log.Errorf("error validating access to requested namespace with code: %s, %+v", namespace.Code, err)
				return api.NewInternalError(
					"error validating access to requested namespace with code: %s", namespace.Code,
				)
			}
			if !isValid {
				return ctx.Status(
					http.StatusNotFound,
				).JSON(
					api.NewResourceDoesNotExistError("unable to find namespace with code: %s", namespace.Code),
				)
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
