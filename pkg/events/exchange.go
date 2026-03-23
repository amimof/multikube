package events

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/amimof/multikube/pkg/logger"

	eventv1 "github.com/amimof/multikube/api/event/v1"
)

var (
	_      Publisher  = &Exchange{}
	_      Forwarder  = &Exchange{}
	_      Subscriber = &Exchange{}
	tracer            = otel.Tracer("exchange")
)

type (
	NewExchangeOption func(*Exchange)
)

type Handler interface {
	Handle(context.Context, *eventv1.Envelope) error
}

type HandlerFunc func(context.Context, *eventv1.Envelope) error

type Exchange struct {
	topics             map[eventv1.Event][]chan *eventv1.Envelope
	persistentHandlers map[eventv1.Event][]HandlerFunc
	fireOnceHandlers   map[eventv1.Event][]HandlerFunc
	errChan            chan error
	mu                 sync.Mutex
	logger             logger.Logger
	forwarders         []Forwarder
}

func WithExchangeLogger(l logger.Logger) NewExchangeOption {
	return func(e *Exchange) {
		e.logger = l
	}
}

// AddForwarder adds a forwarder to this Exchange, forwarding any message on Publish()
func (e *Exchange) AddForwarder(forwarder Forwarder) {
	e.forwarders = append(e.forwarders, forwarder)
}

// On registers a handler func for a certain event type
func (e *Exchange) On(ev eventv1.Event, f HandlerFunc) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.persistentHandlers[ev] = append(e.persistentHandlers[ev], f)
}

// Once attaches a handler to the specified event type. The handler func is only executed once
func (e *Exchange) Once(ev eventv1.Event, f HandlerFunc) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.fireOnceHandlers[ev] = append(e.fireOnceHandlers[ev], f)
}

// Unsubscribe implements Subscriber.
func (e *Exchange) Unsubscribe(context.Context, eventv1.Event) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return nil
}

// Subscribe subscribes to events of a certain event type
func (e *Exchange) Subscribe(ctx context.Context, t ...eventv1.Event) chan *eventv1.Envelope {
	ch := make(chan *eventv1.Envelope, 10)
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, evType := range t {
		e.topics[evType] = append(e.topics[evType], ch)
	}
	return ch
}

// Forward publishes the event using the publishers added to this exchange. Implements Forwarder.
func (e *Exchange) Forward(ctx context.Context, ev *eventv1.Envelope) error {
	return e.publish(ctx, ev, true)
}

// Publish publishes an event of a certain type
func (e *Exchange) Publish(ctx context.Context, ev *eventv1.Envelope) error {
	return e.publish(ctx, ev, false)
}

// Publish publishes an event of a certain type
func (e *Exchange) publish(ctx context.Context, ev *eventv1.Envelope, persist bool) error {
	e.mu.Lock()
	t := ev.GetEvent()

	// Snapshot channels, handlers and forwarders under lock
	evChans := append([]chan *eventv1.Envelope(nil), e.topics[t]...)
	persistent := append([]HandlerFunc(nil), e.persistentHandlers[t]...)
	once := append([]HandlerFunc(nil), e.fireOnceHandlers[t]...)
	delete(e.fireOnceHandlers, t)
	forwarders := append([]Forwarder(nil), e.forwarders...)
	e.mu.Unlock()

	ctx, span := tracer.Start(ctx, "exchange.Publish")
	span.SetAttributes(
		attribute.String("event.type", t.String()),
	)
	defer span.End()

	// fan out to subscribers (lock-free)
	for _, evChan := range evChans {
		go func(ch chan *eventv1.Envelope) { ch <- ev }(evChan)
	}

	// call persistent handlers (lock-free — re-entrant Publish now safe)
	for _, handler := range persistent {
		if err := handler(ctx, ev); err != nil {
			go func(err error) { e.errChan <- err }(err)
		}
	}

	// call fire-once handlers (lock-free)
	for _, handler := range once {
		if err := handler(ctx, ev); err != nil {
			e.errChan <- err
		}
	}

	// call forwarders (lock-free)
	if persist {
		for _, forwarder := range forwarders {
			if err := forwarder.Forward(ctx, ev); err != nil {
				return err
			}
		}
	}

	return nil
}

func NewExchange(opts ...NewExchangeOption) *Exchange {
	e := &Exchange{
		topics:             make(map[eventv1.Event][]chan *eventv1.Envelope),
		persistentHandlers: make(map[eventv1.Event][]HandlerFunc),
		errChan:            make(chan error),
		logger:             logger.ConsoleLogger{},
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}
