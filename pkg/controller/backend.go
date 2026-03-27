package controller

import (
	"context"
	"fmt"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"github.com/amimof/multikube/pkg/client"
	"github.com/amimof/multikube/pkg/compile"
	"github.com/amimof/multikube/pkg/events"
	"github.com/amimof/multikube/pkg/logger"
	"github.com/amimof/multikube/pkg/proxy"
	proxyv2 "github.com/amimof/multikube/pkg/proxyv2"

	backendv1 "github.com/amimof/multikube/api/backend/v1"
	cav1 "github.com/amimof/multikube/api/ca/v1"
	certificatev1 "github.com/amimof/multikube/api/certificate/v1"
	routev1 "github.com/amimof/multikube/api/route/v1"
)

type Controller struct {
	mu        sync.Mutex
	logger    logger.Logger
	clientset *client.ClientSet
	tracer    trace.Tracer
	exchange  *events.Exchange
	proxy     *proxy.Proxy
	compiler  *compile.Compiler
	runtime   *proxyv2.RuntimeStore
	cache     *compile.State
}

type ControllerCache = compile.State

type NewOption func(c *Controller)

func WithCompiler(comp *compile.Compiler) NewOption {
	return func(c *Controller) {
		c.compiler = comp
	}
}

func WithRuntime(runtime *proxyv2.RuntimeStore) NewOption {
	return func(c *Controller) {
		c.runtime = runtime
	}
}

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

func (c *Controller) onBackendCreate(_ context.Context, b *backendv1.Backend) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger.Info("on create handler", "backend", b.GetMeta().GetName())

	// Update cache
	c.cache.Backends[b.GetMeta().GetName()] = b

	// Compile
	return c.compileRuntime()
}

func (c *Controller) onRouteCreate(_ context.Context, r *routev1.Route) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger.Info("on create handler", "route", r.GetMeta().GetName())

	// Update cache
	c.cache.Routes[r.GetMeta().GetName()] = r

	// Compile
	return c.compileRuntime()
}

// Compiles into runtime types and stores in store
func (c *Controller) compileRuntime() error {
	rt, err := c.compiler.Compile(c.cache)
	if err != nil {
		return err
	}
	rt.Version++
	c.runtime.Store(rt)
	c.logger.Info("published runtime snapshot", "version", rt.Version)
	return nil
}

func (c *Controller) onInit(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	backends, err := c.clientset.BackendV1().List(ctx)
	if err != nil {
		return fmt.Errorf("error listing backends: %v", err)
	}
	for _, backend := range backends {
		c.cache.Backends[backend.GetMeta().GetName()] = backend
	}

	cas, err := c.clientset.CAV1().List(ctx)
	if err != nil {
		return fmt.Errorf("error listing cas: %v", err)
	}
	for _, ca := range cas {
		c.cache.CertificateAuthorities[ca.GetMeta().GetName()] = ca
	}

	certs, err := c.clientset.CertificateV1().List(ctx)
	if err != nil {
		return fmt.Errorf("error listing certs: %v", err)
	}
	for _, cert := range certs {
		c.cache.Certificates[cert.GetMeta().GetName()] = cert
	}

	routes, err := c.clientset.RouteV1().List(ctx)
	if err != nil {
		return fmt.Errorf("error listing routes: %v", err)
	}
	for _, route := range routes {
		c.cache.Routes[route.GetMeta().GetName()] = route
	}

	return c.compileRuntime()
}

func (c *Controller) Run(ctx context.Context) {
	if err := c.onInit(ctx); err != nil {
		c.logger.Error("error initializing controller", "error", err)
		return
	}

	// Subscribe to events via the exchange
	c.exchange.On(events.BackendCreate, events.HandleErrors(c.logger, events.HandleBackends(c.onBackendCreate)))
	// c.exchange.On(events.BackendDelete, events.HandleErrors(c.logger, events.HandleBackends(c.onDelete)))
	// c.exchange.On(events.BackendUpdate, events.HandleErrors(c.logger, events.HandleBackends(c.onUpdate)))
	// c.exchange.On(events.BackendPatch, events.HandleErrors(c.logger, events.HandleBackends(c.onPatch)))
	c.exchange.On(events.RouteCreate, events.HandleErrors(c.logger, events.HandleRoutes(c.onRouteCreate)))

	// Block until context is cancelled
	<-ctx.Done()
}

func New(cs *client.ClientSet, opts ...NewOption) *Controller {
	m := &Controller{
		clientset: cs,
		logger:    logger.ConsoleLogger{},
		tracer:    otel.Tracer("controller"),
		cache: &compile.State{
			Backends:               map[string]*backendv1.Backend{},
			Routes:                 map[string]*routev1.Route{},
			Certificates:           map[string]*certificatev1.Certificate{},
			CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
		},
	}
	for _, opt := range opts {
		opt(m)
	}

	return m
}
