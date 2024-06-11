package dao

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// EventListenerProvider provides an interface to work with database event listener.
type EventListenerProvider interface {
	// Listen listens for incoming database events.
	Listen()
	// Subscribe subscribe to particular channel.
	Subscribe(subscriber chan<- string)
	// GetChannelName returns channel name.
	GetChannelName() string
}

// EventListener represents database event listener.
type EventListener struct {
	mu            sync.Mutex
	ctx           context.Context
	channel       string
	connection    *stdlib.Conn
	subscriptions map[string][]chan<- string
}

// NewEventListener creates new database event listener.
func NewEventListener(ctx context.Context, db *gorm.DB, channel string) (*EventListener, error) {
	eventListener := EventListener{
		ctx:           ctx,
		channel:       channel,
		subscriptions: make(map[string][]chan<- string),
	}

	switch db.Dialector.Name() {
	case "postgres":
		sqlDB, err := db.DB()
		if err != nil {
			return nil, eris.Wrap(err, "error getting db instance")
		}
		driverConnection, err := sqlDB.Conn(ctx)
		if err != nil {
			return nil, eris.Wrap(err, "error getting database connection")
		}

		if err := driverConnection.Raw(func(driverConn any) error {
			var ok bool
			if eventListener.connection, ok = driverConn.(*stdlib.Conn); !ok {
				return eris.New(
					"error getting underlying driver connection. driver connection has no type *stdlib.Conn",
				)
			}
			return nil
		}); err != nil {
			return nil, eris.Wrap(err, "error getting underlying driver connection")
		}

		if _, err := eventListener.connection.Conn().Exec(
			ctx, fmt.Sprintf("listen %s", channel),
		); err != nil {
			return nil, eris.Wrapf(err, "error creating listener for %s channel", channel)
		}
	}

	return &eventListener, nil
}

// NewNamespaceListener creates new database event listener for Namespace entity.
func NewNamespaceListener(ctx context.Context, db *gorm.DB) (*EventListener, error) {
	return NewEventListener(ctx, db, "namespace_update_events")
}

// Listen listens for incoming database events.
func (el *EventListener) Listen() {
	// if listener not nil, then listen for incoming events from database.
	// if listener is nil, then just return closed channel to do not do anything further.
	if el.connection != nil {
		go func() {
			for {
				select {
				case <-el.ctx.Done():
					log.Debugf("listener finished. exiting.")
					return
				default:
					notification, err := el.connection.Conn().WaitForNotification(el.ctx)
					if err != nil {
						log.Errorf("error occurred while listening for the event: %+v", err)
						return
					}
					for _, ch := range el.subscriptions[el.channel] {
						ch <- notification.Payload
					}
				}
			}
		}()
	}
}

// Subscribe subscribe to particular channel.
func (el *EventListener) Subscribe(subscriber chan<- string) {
	el.mu.Lock()
	defer el.mu.Unlock()
	if _, ok := el.subscriptions[el.channel]; !ok {
		el.subscriptions[el.channel] = []chan<- string{subscriber}
	} else {
		el.subscriptions[el.channel] = append(el.subscriptions[el.channel], subscriber)
	}
}

// GetChannelName returns current channel name.
func (el *EventListener) GetChannelName() string {
	return el.channel
}
