package dao

import (
	"context"
	"net/url"
	"time"

	"github.com/lib/pq"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"
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
	ctx      context.Context
	channel  string
	listener *pq.Listener
}

// NewEventListener creates new database event listener.
func NewEventListener(ctx context.Context, channel, dsn string) (*EventListener, error) {
	dsnURL, err := url.Parse(dsn)
	if err != nil {
		return nil, eris.Wrap(err, "invalid database URL")
	}

	var listener *pq.Listener
	switch dsnURL.Scheme {
	case "postgres", "postgresql":
		listener = pq.NewListener(
			dsn,
			10*time.Second,
			time.Minute,
			func(ev pq.ListenerEventType, err error) {
				if err != nil {
					log.Errorf(`error happened: %s`, err.Error())
				}
			},
		)
		if err := listener.Listen(channel); err != nil {
			return nil, eris.Wrapf(err, "error creating listener for %s channel", channel)
		}
	}
	return &EventListener{
		ctx:      ctx,
		channel:  channel,
		listener: listener,
	}, nil
}

// NewNamespaceListener creates new database event listener for Namespace entity.
func NewNamespaceListener(ctx context.Context, dsn string) (*EventListener, error) {
	return NewEventListener(ctx, "namespace-events-channel", dsn)
}

// Listen listens for incoming database events.
func (l EventListener) Listen() <-chan string {
	ch := make(chan string)
	// if listener not nil, the listen for incoming events from database.
	// if listener is nil, then just return closed channel to do not do anything further.
	if l.listener != nil {
		go func() {
			defer close(ch)
			for {
				select {
				case n := <-l.listener.Notify:
					log.Infof(`received event: %s from channel: '%s'`, n.Extra, n.Channel)
					ch <- n.Extra
				case <-l.ctx.Done():
					log.Info(`context has been canceled. exit`)
					return
				}
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
