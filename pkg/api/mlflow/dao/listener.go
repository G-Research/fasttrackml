package dao

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// EventListenerProvider provides an interface to work with database event listener.
type EventListenerProvider interface {
	// Listen listens for incoming database events.
	Listen() <-chan string
	// GetChannelName returns channel name.
	GetChannelName() string
}

// EventListener represents database event listener.
type EventListener struct {
	ctx        context.Context
	channel    string
	connection *stdlib.Conn
}

// NewEventListener creates new database event listener.
func NewEventListener(ctx context.Context, db *gorm.DB, channel string) (*EventListener, error) {
	eventListener := EventListener{
		ctx:     ctx,
		channel: channel,
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
func (l EventListener) Listen() <-chan string {
	ch := make(chan string)
	// if listener not nil, then listen for incoming events from database.
	// if listener is nil, then just return closed channel to do not do anything further.
	if l.connection != nil {
		go func() {
			defer close(ch)
			for {
				notification, err := l.connection.Conn().WaitForNotification(context.Background())
				if err != nil {
					log.Errorf("error occurred while listening for the event: %+v", err)
					return
				}
				ch <- notification.Payload
			}
		}()
	} else {
		close(ch)
	}
	return ch
}

// GetChannelName returns current channel name.
func (l EventListener) GetChannelName() string {
	return l.channel
}
