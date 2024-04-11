package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/common/dao"
	"github.com/G-Research/fasttrackml/pkg/common/events"
	"github.com/G-Research/fasttrackml/pkg/database"
)

// NamespaceCachedRepository cached repository to work with `namespace` entity.
type NamespaceCachedRepository struct {
	cache                  *lru.Cache[string, models.Namespace]
	namespaceRepository    NamespaceRepositoryProvider
	namespaceEventListener dao.EventListenerProvider
}

// NewNamespaceCachedRepository creates new instance of cached repository to work with `namespace` entity.
func NewNamespaceCachedRepository(
	ctx context.Context,
	namespaceRepository NamespaceRepositoryProvider,
	namespaceEventListener dao.EventListenerProvider,
) (*NamespaceCachedRepository, error) {
	cache, err := lru.New[string, models.Namespace](1000)
	if err != nil {
		return nil, eris.Wrap(err, "error creating lru cache for namespace entities")
	}

	repository := NamespaceCachedRepository{
		cache:                  cache,
		namespaceRepository:    namespaceRepository,
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

// Create creates new models.Namespace entity.
func (r NamespaceCachedRepository) Create(ctx context.Context, namespace *models.Namespace) error {
	if err := r.namespaceRepository.Create(ctx, namespace); err != nil {
		return eris.Wrap(err, "error creating cached namespace entity")
	}

	// trigger database event to notify current instance and
	// other instances to create record in theirs local cache.
	if err := r.sendEvent(events.NamespaceEventActionCreated, namespace); err != nil {
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
	if err := r.sendEvent(events.NamespaceEventActionUpdated, namespace); err != nil {
		return eris.Wrap(err, "error sending database event")
	}
	return nil
}

// GetByCode returns namespace by its Code.
func (r NamespaceCachedRepository) GetByCode(
	ctx context.Context, code string,
) (*models.Namespace, error) {
	result, ok := r.cache.Get(code)
	if ok {
		return &result, nil
	}

	namespace, err := r.namespaceRepository.GetByCode(ctx, code)
	if err != nil {
		return nil, eris.Wrapf(err, "error getting cached namespace by code: %s", code)
	}
	if namespace == nil {
		return nil, nil
	}

	// trigger database event to notify current instance and
	// other instances to add record to theirs local cache.
	if err := r.sendEvent(events.NamespaceEventActionFetched, namespace); err != nil {
		return nil, eris.Wrap(err, "error sending database event")
	}
	return namespace, nil
}

// GetByID returns namespace by its ID.
func (r NamespaceCachedRepository) GetByID(ctx context.Context, id uint) (*models.Namespace, error) {
	return r.namespaceRepository.GetByID(ctx, id)
}

// Delete deletes existing models.Namespace entity.
func (r NamespaceCachedRepository) Delete(ctx context.Context, namespace *models.Namespace) error {
	if err := r.namespaceRepository.Delete(ctx, namespace); err != nil {
		return eris.Wrap(err, "error deleting cached namespace entity")
	}

	// trigger database event to notify current instance and
	// other instances to remove record from theirs local cache.
	if err := r.sendEvent(events.NamespaceEventActionDeleted, namespace); err != nil {
		return eris.Wrap(err, "error sending database event")
	}
	return nil
}

// List returns all namespaces.
func (r NamespaceCachedRepository) List(ctx context.Context) ([]models.Namespace, error) {
	return r.namespaceRepository.List(ctx)
}

// processEvent process incoming event from database.
func (r NamespaceCachedRepository) processEvent(data string) error {
	log.Debugf("got incoming namespace event: %s", data)
	event := events.NamespaceEvent{}
	if err := json.Unmarshal([]byte(data), &event); err != nil {
		return eris.Wrap(err, "error unmarshaling incoming database event")
	}
	switch event.Action {
	case events.NamespaceEventActionFetched:
		r.cache.Add(event.Namespace.Code, event.Namespace)
	case events.NamespaceEventActionCreated:
		r.cache.Add(event.Namespace.Code, event.Namespace)
	case events.NamespaceEventActionUpdated:
		r.cache.Add(event.Namespace.Code, event.Namespace)
	case events.NamespaceEventActionDeleted:
		r.cache.Remove(event.Namespace.Code)
	}
	log.Debugf("namespace keys in local cache: %+v", r.cache.Keys())
	return nil
}

// GetDB returns current DB instance.
func (r NamespaceCachedRepository) GetDB() *gorm.DB {
	return r.namespaceRepository.GetDB()
}

// sendEvent sends database event.
func (r NamespaceCachedRepository) sendEvent(action events.NamespaceEventAction, namespace *models.Namespace) error {
	// skip event processing if current database is not a `postgres`.
	if r.namespaceRepository.GetDB().Dialector.Name() != database.PostgresDialectorName {
		return nil
	}

	data, err := json.Marshal(events.NamespaceEvent{
		Action:    action,
		Namespace: *namespace,
	})
	if err != nil {
		return eris.Wrap(err, "error serializing NamespaceEvent event")
	}
	if err := r.namespaceRepository.GetDB().Exec(
		fmt.Sprintf(`SELECT pg_notify('%s', '%s')`, r.namespaceEventListener.GetChannelName(), data),
	).Error; err != nil {
		return eris.Wrap(err, "error triggering 'pg_notify'")
	}
	return nil
}
