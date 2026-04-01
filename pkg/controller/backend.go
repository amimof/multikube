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
	proxyv2 "github.com/amimof/multikube/pkg/proxyv2"

	backendv1 "github.com/amimof/multikube/api/backend/v1"
	cav1 "github.com/amimof/multikube/api/ca/v1"
	certificatev1 "github.com/amimof/multikube/api/certificate/v1"
	credentialv1 "github.com/amimof/multikube/api/credential/v1"
	policyv1 "github.com/amimof/multikube/api/policy/v1"
	routev1 "github.com/amimof/multikube/api/route/v1"
)

type Controller struct {
	mu        sync.Mutex
	logger    logger.Logger
	clientset *client.ClientSet
	tracer    trace.Tracer
	exchange  *events.Exchange
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

func (c *Controller) onRouteUpdate(_ context.Context, r *routev1.Route) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger.Info("on update handler", "route", r.GetMeta().GetName())

	c.cache.Routes[r.GetMeta().GetName()] = r

	return c.compileRuntime()
}

func (c *Controller) onRouteDelete(_ context.Context, r *routev1.Route) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger.Info("on delete handler", "route", r.GetMeta().GetName())

	delete(c.cache.Routes, r.GetMeta().GetName())

	return c.compileRuntime()
}

func (c *Controller) onPolicyCreate(_ context.Context, p *policyv1.Policy) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger.Info("on create handler", "policy", p.GetMeta().GetName())

	c.cache.Policies[p.GetMeta().GetName()] = p

	return c.compileRuntime()
}

func (c *Controller) onPolicyUpdate(_ context.Context, p *policyv1.Policy) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger.Info("on update handler", "policy", p.GetMeta().GetName())

	c.cache.Policies[p.GetMeta().GetName()] = p

	return c.compileRuntime()
}

func (c *Controller) onPolicyDelete(_ context.Context, p *policyv1.Policy) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger.Info("on delete handler", "policy", p.GetMeta().GetName())

	delete(c.cache.Policies, p.GetMeta().GetName())

	return c.compileRuntime()
}

func (c *Controller) onCredentialCreate(_ context.Context, ctr *credentialv1.Credential) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger.Info("on create handler", "credential", ctr.GetMeta().GetName())

	c.cache.Credentials[ctr.GetMeta().GetName()] = ctr

	return c.compileRuntime()
}

func (c *Controller) onCredentialUpdate(_ context.Context, ctr *credentialv1.Credential) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger.Info("on update handler", "credential", ctr.GetMeta().GetName())

	c.cache.Credentials[ctr.GetMeta().GetName()] = ctr

	return c.compileRuntime()
}

func (c *Controller) onCredentialDelete(_ context.Context, ctr *credentialv1.Credential) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger.Info("on delete handler", "credential", ctr.GetMeta().GetName())

	delete(c.cache.Credentials, ctr.GetMeta().GetName())

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

	credentials, err := c.clientset.CredentialV1().List(ctx)
	if err != nil {
		return fmt.Errorf("error listing credentials: %v", err)
	}
	for _, credential := range credentials {
		c.cache.Credentials[credential.GetMeta().GetName()] = credential
	}

	routes, err := c.clientset.RouteV1().List(ctx)
	if err != nil {
		return fmt.Errorf("error listing routes: %v", err)
	}
	for _, route := range routes {
		c.cache.Routes[route.GetMeta().GetName()] = route
	}

	policies, err := c.clientset.PolicyV1().List(ctx)
	if err != nil {
		return fmt.Errorf("error listing policies: %v", err)
	}
	for _, policy := range policies {
		c.cache.Policies[policy.GetMeta().GetName()] = policy
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
	c.exchange.On(events.RouteUpdate, events.HandleErrors(c.logger, events.HandleRoutes(c.onRouteUpdate)))
	c.exchange.On(events.RoutePatch, events.HandleErrors(c.logger, events.HandleRoutes(c.onRouteUpdate)))
	c.exchange.On(events.RouteDelete, events.HandleErrors(c.logger, events.HandleRoutes(c.onRouteDelete)))
	c.exchange.On(events.CredentialCreate, events.HandleErrors(c.logger, events.HandleCredentials(c.onCredentialCreate)))
	c.exchange.On(events.CredentialUpdate, events.HandleErrors(c.logger, events.HandleCredentials(c.onCredentialUpdate)))
	c.exchange.On(events.CredentialPatch, events.HandleErrors(c.logger, events.HandleCredentials(c.onCredentialUpdate)))
	c.exchange.On(events.CredentialDelete, events.HandleErrors(c.logger, events.HandleCredentials(c.onCredentialDelete)))
	c.exchange.On(events.PolicyCreate, events.HandleErrors(c.logger, events.HandlePolicies(c.onPolicyCreate)))
	c.exchange.On(events.PolicyUpdate, events.HandleErrors(c.logger, events.HandlePolicies(c.onPolicyUpdate)))
	c.exchange.On(events.PolicyPatch, events.HandleErrors(c.logger, events.HandlePolicies(c.onPolicyUpdate)))
	c.exchange.On(events.PolicyDelete, events.HandleErrors(c.logger, events.HandlePolicies(c.onPolicyDelete)))

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
			Credentials:            map[string]*credentialv1.Credential{},
			Policies:               map[string]*policyv1.Policy{},
		},
	}
	for _, opt := range opts {
		opt(m)
	}

	return m
}
