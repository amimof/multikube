package controller

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"github.com/amimof/multikube/pkg/client"
	"github.com/amimof/multikube/pkg/events"
	"github.com/amimof/multikube/pkg/logger"
	"github.com/amimof/multikube/pkg/proxy"

	backendv1 "github.com/amimof/multikube/api/backend/v1"
)

type Controller struct {
	logger    logger.Logger
	clientset *client.ClientSet
	tracer    trace.Tracer
	exchange  *events.Exchange
	proxy     *proxy.Proxy
}

type NewOption func(c *Controller)

func WithLogger(l logger.Logger) NewOption {
	return func(c *Controller) {
		c.logger = l
	}
}

func WithProxy(p *proxy.Proxy) NewOption {
	return func(c *Controller) {
		c.proxy = p
	}
}

func WithExchange(e *events.Exchange) NewOption {
	return func(c *Controller) {
		c.exchange = e
	}
}

func (c *Controller) onCreate(ctx context.Context, b *backendv1.Backend) error {
	c.logger.Info("on create handler", "backend", b.GetMeta().GetName())
	return nil
}

func (c *Controller) Run(ctx context.Context) {
	// Subscribe to events via the exchange
	c.exchange.On(events.BackendCreate, events.HandleErrors(c.logger, events.HandleBackends(c.onCreate)))
	// c.exchange.On(events.BackendDelete, events.HandleErrors(c.logger, events.HandleBackends(c.onDelete)))
	// c.exchange.On(events.BackendUpdate, events.HandleErrors(c.logger, events.HandleBackends(c.onUpdate)))
	// c.exchange.On(events.BackendPatch, events.HandleErrors(c.logger, events.HandleBackends(c.onPatch)))

	// Block until context is cancelled
	<-ctx.Done()
}

func New(cs *client.ClientSet, opts ...NewOption) *Controller {
	m := &Controller{
		clientset: cs,
		logger:    logger.ConsoleLogger{},
		tracer:    otel.Tracer("controller"),
	}
	for _, opt := range opts {
		opt(m)
	}

	return m
}
