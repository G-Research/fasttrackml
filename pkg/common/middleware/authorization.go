package middleware

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/pkg/common/client/oidc"
	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
	"github.com/G-Research/fasttrackml/pkg/common/dao/repositories"
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
		authToken := userPermissions.ValidateAuthToken(ctx.Get(fiber.HeaderAuthorization)[6:])
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
		user, err := client.Verify(ctx.Context(), ctx.Get(fiber.HeaderAuthorization)[7:])
		if err != nil {
			return ctx.Status(
				http.StatusNotFound,
			).JSON(
				api.NewResourceDoesNotExistError("unable to find namespace with code: %s", namespace.Code),
			)
		}
		log.Debugf("user has roles: %v accociated", user.Groups)
		if user.IsAdmin() {
			return ctx.Next()
		}

		isValid, err := rolesRepository.ValidateRolesAccessToNamespace(ctx.Context(), user.Groups, namespace.Code)
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

		return ctx.Next()
	}
}
