package repositories

import (
	"context"
	"encoding/json"
	"slices"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/common/dao"
	"github.com/G-Research/fasttrackml/pkg/common/events"
)

// RoleRepositoryProvider provides an interface to work with `role` entity.
type RoleRepositoryProvider interface {
	// ValidateRolesAccessToNamespace makes validation that requested roles has access to requested namespace.
	ValidateRolesAccessToNamespace(ctx context.Context, roles []string, namespaceCode string) (bool, error)
}

// RoleCachedRepository cached repository to work with `role` entity.
type RoleCachedRepository struct {
	db                     *gorm.DB
	cache                  *lru.Cache[string, []string]
	namespaceEventListener dao.EventListenerProvider
}

// NewRoleCachedRepository creates new instance of cached repository to work with `role` entity.
func NewRoleCachedRepository(
	ctx context.Context, db *gorm.DB, namespaceEventListener dao.EventListenerProvider,
) (*RoleCachedRepository, error) {
	cache, err := lru.New[string, []string](1000)
	if err != nil {
		return nil, eris.Wrap(err, "error creating lru cache for roles entities")
	}

	repository := RoleCachedRepository{
		db:                     db,
		cache:                  cache,
		namespaceEventListener: namespaceEventListener,
	}

	ch := make(chan string)
	go func() {
		defer close(ch)
		for {
			select {
			case <-ctx.Done():
				return
			case data := <-ch:
				if err := repository.processEvent(data); err != nil {
					log.Errorf(`error processing incoming event: %s, error: %+v`, data, err)
				}
			}
		}
	}()

	// subscribe to incoming events.
	namespaceEventListener.Subscribe(ch)

	return &repository, nil
}

// ValidateRolesAccessToNamespace makes validation that requested roles has access to requested namespace.
func (r RoleCachedRepository) ValidateRolesAccessToNamespace(
	ctx context.Context, requestedRoles []string, requestedNamespaceCode string,
) (bool, error) {
	// if namespace already exists in cache, check permissions immediately.
	namespaceRoles, ok := r.cache.Get(requestedNamespaceCode)
	if ok {
		for _, requestedRole := range requestedRoles {
			if slices.Contains(namespaceRoles, requestedRole) {
				return true, nil
			}
		}
		return false, nil
	}

	// otherwise check database and store result in cache.
	var data []models.RoleNamespace
	if err := r.db.WithContext(ctx).Model(
		&models.RoleNamespace{},
	).Joins(
		"Name",
		r.db.Select("role"),
	).InnerJoins(
		"Namespace",
		r.db.Select(
			"code",
		).Where(
			&models.Namespace{Code: requestedNamespaceCode},
		),
	).Find(&data).Error; err != nil {
		return false, eris.Wrapf(err, "error getting roles for namespace with code: %s", requestedNamespaceCode)
	}

	namespaceRoles = make([]string, len(data))
	for i, namespaceRole := range data {
		namespaceRoles[i] = namespaceRole.Role.Name
	}

	// save into cache.
	r.cache.Add(requestedNamespaceCode, namespaceRoles)

	// check permissions from database.
	for _, requestedRole := range requestedRoles {
		if slices.Contains(namespaceRoles, requestedRole) {
			return true, nil
		}
	}

	return false, nil
}

// processEvent process incoming event from database.
func (r RoleCachedRepository) processEvent(data string) error {
	log.Debugf("got incoming namespace event: %s", data)
	event := events.NamespaceEvent{}
	if err := json.Unmarshal([]byte(data), &event); err != nil {
		return eris.Wrap(err, "error unmarshaling incoming database event")
	}
	switch event.Action {
	case events.NamespaceEventActionDeleted:
		r.cache.Remove(event.Namespace.Code)
	}
	log.Debugf("namespace keys in local cache: %+v", r.cache.Keys())
	return nil
}
