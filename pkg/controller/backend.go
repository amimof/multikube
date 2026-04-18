package controller

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

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
	mu                sync.Mutex
	logger            logger.Logger
	clientset         *client.ClientSet
	tracer            trace.Tracer
	exchange          *events.Exchange
	compiler          *compile.Compiler
	runtime           *proxyv2.RuntimeStore
	cache             *compile.State
	heartBeatInterval time.Duration
	heartBeatTimeout  time.Duration
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

func WithHeartBeatInterval(every time.Duration) NewOption {
	return func(c *Controller) {
		c.heartBeatInterval = every
	}
}

func WithHeartBeatTimeout(timeout time.Duration) NewOption {
	return func(c *Controller) {
		c.heartBeatTimeout = timeout
	}
}

func (c *Controller) onBackendCreate(ctx context.Context, b *backendv1.Backend) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger.Info("on create handler", "backend", b.GetMeta().GetName())

	// Update cache
	c.cache.Backends[b.GetMeta().GetName()] = b

	// Compile
	return c.compileRuntime(ctx)
}

func (c *Controller) onBackendUpdate(ctx context.Context, p *backendv1.Backend) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger.Info("on update handler", "backend", p.GetMeta().GetName())

	c.cache.Backends[p.GetMeta().GetName()] = p

	return c.compileRuntime(ctx)
}

func (c *Controller) onBackendDelete(ctx context.Context, p *backendv1.Backend) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger.Info("on delete handler", "backend", p.GetMeta().GetName())

	delete(c.cache.Backends, p.GetMeta().GetName())

	return c.compileRuntime(ctx)
}

func (c *Controller) onRouteCreate(ctx context.Context, r *routev1.Route) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger.Info("on create handler", "route", r.GetMeta().GetName())

	// Update cache
	c.cache.Routes[r.GetMeta().GetName()] = r

	// Compile
	return c.compileRuntime(ctx)
}

func (c *Controller) onRouteUpdate(ctx context.Context, r *routev1.Route) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger.Info("on update handler", "route", r.GetMeta().GetName())

	c.cache.Routes[r.GetMeta().GetName()] = r

	return c.compileRuntime(ctx)
}

func (c *Controller) onRouteDelete(ctx context.Context, r *routev1.Route) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger.Info("on delete handler", "route", r.GetMeta().GetName())

	delete(c.cache.Routes, r.GetMeta().GetName())

	return c.compileRuntime(ctx)
}

func (c *Controller) onPolicyCreate(ctx context.Context, p *policyv1.Policy) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger.Info("on create handler", "policy", p.GetMeta().GetName())

	c.cache.Policies[p.GetMeta().GetName()] = p

	return c.compileRuntime(ctx)
}

func (c *Controller) onPolicyUpdate(ctx context.Context, p *policyv1.Policy) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger.Info("on update handler", "policy", p.GetMeta().GetName())

	c.cache.Policies[p.GetMeta().GetName()] = p

	return c.compileRuntime(ctx)
}

func (c *Controller) onPolicyDelete(ctx context.Context, p *policyv1.Policy) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger.Info("on delete handler", "policy", p.GetMeta().GetName())

	delete(c.cache.Policies, p.GetMeta().GetName())

	return c.compileRuntime(ctx)
}

func (c *Controller) onCredentialCreate(ctx context.Context, ctr *credentialv1.Credential) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger.Info("on create handler", "credential", ctr.GetMeta().GetName())

	c.cache.Credentials[ctr.GetMeta().GetName()] = ctr

	return c.compileRuntime(ctx)
}

func (c *Controller) onCredentialUpdate(ctx context.Context, ctr *credentialv1.Credential) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger.Info("on update handler", "credential", ctr.GetMeta().GetName())

	c.cache.Credentials[ctr.GetMeta().GetName()] = ctr

	return c.compileRuntime(ctx)
}

func (c *Controller) onCredentialDelete(ctx context.Context, ctr *credentialv1.Credential) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger.Info("on delete handler", "credential", ctr.GetMeta().GetName())

	delete(c.cache.Credentials, ctr.GetMeta().GetName())

	return c.compileRuntime(ctx)
}

func (c *Controller) onCertificateCreate(ctx context.Context, ctr *certificatev1.Certificate) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger.Info("on create handler", "certificate", ctr.GetMeta().GetName())

	c.cache.Certificates[ctr.GetMeta().GetName()] = ctr

	return c.compileRuntime(ctx)
}

func (c *Controller) onCertificateUpdate(ctx context.Context, ctr *certificatev1.Certificate) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger.Info("on update handler", "certificate", ctr.GetMeta().GetName())

	c.cache.Certificates[ctr.GetMeta().GetName()] = ctr

	return c.compileRuntime(ctx)
}

func (c *Controller) onCertificateDelete(ctx context.Context, ctr *certificatev1.Certificate) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger.Info("on delete handler", "certificate", ctr.GetMeta().GetName())

	delete(c.cache.Certificates, ctr.GetMeta().GetName())

	return c.compileRuntime(ctx)
}

func (c *Controller) onCertificateAuthorityCreate(ctx context.Context, ctr *cav1.CertificateAuthority) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger.Info("on create handler", "ca", ctr.GetMeta().GetName())

	c.cache.CertificateAuthorities[ctr.GetMeta().GetName()] = ctr

	return c.compileRuntime(ctx)
}

func (c *Controller) onCertificateAuthorityUpdate(ctx context.Context, ctr *cav1.CertificateAuthority) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger.Info("on update handler", "ca", ctr.GetMeta().GetName())

	c.cache.CertificateAuthorities[ctr.GetMeta().GetName()] = ctr

	return c.compileRuntime(ctx)
}

func (c *Controller) onCertificateAuthorityDelete(ctx context.Context, ctr *cav1.CertificateAuthority) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger.Info("on delete handler", "ca", ctr.GetMeta().GetName())

	delete(c.cache.CertificateAuthorities, ctr.GetMeta().GetName())

	return c.compileRuntime(ctx)
}

// Compiles into runtime types and stores in store.
func (c *Controller) compileRuntime(ctx context.Context) error {
	result, err := c.compiler.Compile(c.cache)
	if err != nil {
		return err
	}
	c.runtime.Store(result.Runtime)
	c.logger.Info("published runtime snapshot", "version", result.Runtime.Version)
	if err := c.reconcileRouteStatuses(ctx, result.RouteStatuses); err != nil {
		c.logger.Error("error reconciling route statuses", "error", err)
	}
	if err := c.reconcileBackendStatuses(ctx, result.BackendStatuses); err != nil {
		c.logger.Error("error reconciling route statuses", "error", err)
	}
	return nil
}

func (c *Controller) reconcileRouteStatuses(ctx context.Context, statuses map[string]compile.CompileStatus) error {
	for name, next := range statuses {
		route, ok := c.cache.Routes[name]
		if !ok || route == nil {
			continue
		}

		st := &routev1.RouteStatus{
			Phase:              wrapperspb.String(next.Phase),
			Reason:             wrapperspb.String(next.Reason),
			LastTransitionTime: timestamppb.Now(),
		}

		if err := c.clientset.RouteV1().UpdateStatus(ctx, name, st, "phase", "reason", "last_transition_time"); err != nil {
			return err
		}
		updated, err := c.clientset.RouteV1().Get(ctx, name)
		if err != nil {
			return err
		}
		c.cache.Routes[name] = updated
	}

	return nil
}

func (c *Controller) reconcileBackendStatuses(ctx context.Context, statuses map[string]compile.CompileStatus) error {
	for name, next := range statuses {
		backend, ok := c.cache.Backends[name]
		if !ok || backend == nil {
			continue
		}

		st := &backendv1.BackendStatus{
			Phase:              wrapperspb.String(next.Phase),
			Reason:             wrapperspb.String(next.Reason),
			LastTransitionTime: timestamppb.Now(),
		}

		if err := c.clientset.BackendV1().UpdateStatus(ctx, name, st, "phase", "reason", "last_transition_time"); err != nil {
			return err
		}
		updated, err := c.clientset.BackendV1().Get(ctx, name)
		if err != nil {
			return err
		}
		c.cache.Backends[name] = updated
	}

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

	return c.compileRuntime(ctx)
}

func (c *Controller) Run(ctx context.Context) {
	if err := c.onInit(ctx); err != nil {
		c.logger.Error("error initializing controller", "error", err)
		return
	}

	// Start heartbeats
	go c.runHeartbeat(ctx)

	// Subscribe to events via the exchange

	// Backends
	c.exchange.On(events.BackendCreate, events.HandleErrors(c.logger, events.HandleBackends(c.onBackendCreate)))
	c.exchange.On(events.BackendUpdate, events.HandleErrors(c.logger, events.HandleBackends(c.onBackendUpdate)))
	c.exchange.On(events.BackendPatch, events.HandleErrors(c.logger, events.HandleBackends(c.onBackendUpdate)))
	c.exchange.On(events.BackendDelete, events.HandleErrors(c.logger, events.HandleBackends(c.onBackendDelete)))

	// Routes
	c.exchange.On(events.RouteCreate, events.HandleErrors(c.logger, events.HandleRoutes(c.onRouteCreate)))
	c.exchange.On(events.RouteUpdate, events.HandleErrors(c.logger, events.HandleRoutes(c.onRouteUpdate)))
	c.exchange.On(events.RoutePatch, events.HandleErrors(c.logger, events.HandleRoutes(c.onRouteUpdate)))
	c.exchange.On(events.RouteDelete, events.HandleErrors(c.logger, events.HandleRoutes(c.onRouteDelete)))

	// Credentials
	c.exchange.On(events.CredentialCreate, events.HandleErrors(c.logger, events.HandleCredentials(c.onCredentialCreate)))
	c.exchange.On(events.CredentialUpdate, events.HandleErrors(c.logger, events.HandleCredentials(c.onCredentialUpdate)))
	c.exchange.On(events.CredentialPatch, events.HandleErrors(c.logger, events.HandleCredentials(c.onCredentialUpdate)))
	c.exchange.On(events.CredentialDelete, events.HandleErrors(c.logger, events.HandleCredentials(c.onCredentialDelete)))

	// Policies
	c.exchange.On(events.PolicyCreate, events.HandleErrors(c.logger, events.HandlePolicies(c.onPolicyCreate)))
	c.exchange.On(events.PolicyUpdate, events.HandleErrors(c.logger, events.HandlePolicies(c.onPolicyUpdate)))
	c.exchange.On(events.PolicyPatch, events.HandleErrors(c.logger, events.HandlePolicies(c.onPolicyUpdate)))
	c.exchange.On(events.PolicyDelete, events.HandleErrors(c.logger, events.HandlePolicies(c.onPolicyDelete)))

	// Certificates
	c.exchange.On(events.CertificateCreate, events.HandleErrors(c.logger, events.HandleCertificates(c.onCertificateCreate)))
	c.exchange.On(events.CertificateUpdate, events.HandleErrors(c.logger, events.HandleCertificates(c.onCertificateUpdate)))
	c.exchange.On(events.CertificatePatch, events.HandleErrors(c.logger, events.HandleCertificates(c.onCertificateUpdate)))
	c.exchange.On(events.CertificateDelete, events.HandleErrors(c.logger, events.HandleCertificates(c.onCertificateDelete)))

	// CAs
	c.exchange.On(events.CertificateAuthorityCreate, events.HandleErrors(c.logger, events.HandleCertificateAuthorities(c.onCertificateAuthorityCreate)))
	c.exchange.On(events.CertificateAuthorityUpdate, events.HandleErrors(c.logger, events.HandleCertificateAuthorities(c.onCertificateAuthorityUpdate)))
	c.exchange.On(events.CertificateAuthorityPatch, events.HandleErrors(c.logger, events.HandleCertificateAuthorities(c.onCertificateAuthorityUpdate)))
	c.exchange.On(events.CertificateAuthorityDelete, events.HandleErrors(c.logger, events.HandleCertificateAuthorities(c.onCertificateAuthorityDelete)))

	// Block until context is cancelled
	<-ctx.Done()
}

func New(cs *client.ClientSet, opts ...NewOption) *Controller {
	m := &Controller{
		clientset:         cs,
		logger:            logger.ConsoleLogger{},
		tracer:            otel.Tracer("controller"),
		heartBeatInterval: 15 * time.Second,
		heartBeatTimeout:  10 * time.Second,
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
