package repositories

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2/log"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

// NamespaceEventAction represents Event action.
type NamespaceEventAction string

// Supported event actions.
const (
	NamespaceEventActionFetched = "fetched"
	NamespaceEventActionCreated = "created"
	NamespaceEventActionDeleted = "deleted"
	NamespaceEventActionUpdated = "updated"
)

// NamespaceEvent represents database event.
type NamespaceEvent struct {
	Action    NamespaceEventAction `json:"action"`
	Namespace *models.Namespace    `json:"namespace"`
}

// NamespaceCachedRepository cached repository to work with `namespace` entity.
type NamespaceCachedRepository struct {
	db                  *gorm.DB
	cache               *lru.Cache[string, *models.Namespace]
	listener            dao.EventListenerProvider
	namespaceRepository NamespaceRepositoryProvider
}

// NewNamespaceCachedRepository creates new instance of cached repository to work with `namespace` entity.
func NewNamespaceCachedRepository(
	db *gorm.DB, listener dao.EventListenerProvider, namespaceRepository NamespaceRepositoryProvider,
) (*NamespaceCachedRepository, error) {
	cache, err := lru.New[string, *models.Namespace](1000)
	if err != nil {
		return nil, eris.Wrap(err, "error creating lru cache for namespace entities")
	}

	// pre load all namespaces into cache.
	var namespaces []models.Namespace
	if err := db.Find(&namespaces).Error; err != nil {
		return nil, eris.Wrapf(err, "error getting namespaces")
	}
	for _, namespace := range namespaces {
		cache.Add(namespace.Code, &namespace)
	}

	repository := NamespaceCachedRepository{
		db:                  db,
		cache:               cache,
		listener:            listener,
		namespaceRepository: namespaceRepository,
	}

	go func() {
		for data := range listener.Listen() {
			if err := repository.processEvent(data); err != nil {
				log.Errorf(`error processing incoming event: %s, error: %+v`, data, err)
			}
		}
	}()
	return &repository, nil
}

// Create creates new models.Namespace entity.
func (r NamespaceCachedRepository) Create(ctx context.Context, namespace *models.Namespace) error {
	if err := r.namespaceRepository.Create(ctx, namespace); err != nil {
		return eris.Wrap(err, "error creating cached namespace entity")
	}

	// trigger database event to notify current instance and
	// other instances to create record in theirs local cache.
	if err := r.sendEvent(NamespaceEventActionCreated, namespace); err != nil {
		return eris.Wrap(err, "error sending database event")
	}
	return nil
}

// Update updates existing models.Namespace entity.
func (r NamespaceCachedRepository) Update(ctx context.Context, namespace *models.Namespace) error {
	if err := r.namespaceRepository.Update(ctx, namespace); err != nil {
		return eris.Wrap(err, "error updating cached namespace entity")
	}

	// trigger database event to notify current instance and
	// other instances to update record in theirs local cache.
	if err := r.sendEvent(NamespaceEventActionUpdated, namespace); err != nil {
		return eris.Wrap(err, "error sending database event")
	}
	return nil
}

// GetByCode returns namespace by its Code.
func (r NamespaceCachedRepository) GetByCode(ctx context.Context, code string) (*models.Namespace, error) {
	result, ok := r.cache.Get(code)
	if ok {
		return result, nil
	}

	namespace, err := r.namespaceRepository.GetByCode(ctx, code)
	if err != nil {
		return nil, eris.Wrapf(err, "error getting cached namespace by code: %s", code)
	}

	// trigger database event to notify current instance and
	// other instances to add record to theirs local cache.
	if err := r.sendEvent(NamespaceEventActionFetched, namespace); err != nil {
		return nil, eris.Wrap(err, "error sending database event")
	}
	return namespace, nil
}

// GetByID returns namespace by its ID.
func (r NamespaceCachedRepository) GetByID(ctx context.Context, id uint) (*models.Namespace, error) {
	return r.GetByID(ctx, id)
}

// Delete deletes existing models.Namespace entity.
func (r NamespaceCachedRepository) Delete(ctx context.Context, namespace *models.Namespace) error {
	if err := r.namespaceRepository.Delete(ctx, namespace); err != nil {
		return eris.Wrap(err, "error deleting cached namespace entity")
	}

	// trigger database event to notify current instance and
	// other instances to remove record from theirs local cache.
	if err := r.sendEvent(NamespaceEventActionDeleted, namespace); err != nil {
		return eris.Wrap(err, "error sending database event")
	}
	return nil
}

// List returns all namespaces.
func (r NamespaceCachedRepository) List(ctx context.Context) ([]models.Namespace, error) {
	return r.List(ctx)
}

// processEvent process incoming event from database.
func (r NamespaceCachedRepository) processEvent(data string) error {
	log.Debugf("got incoming namespace event: %s", data)
	event := NamespaceEvent{}
	if err := json.Unmarshal([]byte(data), &event); err != nil {
		return eris.Wrap(err, "error unmarshaling incoming database event")
	}
	switch event.Action {
	case NamespaceEventActionFetched:
		r.cache.Add(event.Namespace.Code, event.Namespace)
	case NamespaceEventActionCreated:
		r.cache.Add(event.Namespace.Code, event.Namespace)
	case NamespaceEventActionUpdated:
		r.cache.Add(event.Namespace.Code, event.Namespace)
	case NamespaceEventActionDeleted:
		r.cache.Remove(event.Namespace.Code)
	}
	log.Debugf("namespace keys in local cache: %+v", r.cache.Keys())
	return nil
}

// sendEvent sends database event.
func (r NamespaceCachedRepository) sendEvent(action NamespaceEventAction, namespace *models.Namespace) error {
	data, err := json.Marshal(NamespaceEvent{
		Action:    action,
		Namespace: namespace,
	})
	if err != nil {
		return eris.Wrap(err, "error serializing NamespaceEvent event")
	}
	if err := r.db.Exec(
		fmt.Sprintf(`SELECT pg_notify('%s', '%s')`, r.listener.GetChannelName(), data),
	).Error; err != nil {
		return eris.Wrap(err, "error triggering 'pg_notify'")
	}
	return nil
}
