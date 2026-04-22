package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	backendv1 "github.com/amimof/multikube/api/backend/v1"
	"github.com/amimof/multikube/pkg/logger"
	proxy "github.com/amimof/multikube/pkg/proxyv2"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	BackendHealthy   = "Healthy"
	BackendUnhealthy = "Unhealthy"
	BackendUnknown   = "Unknown"
)

type Heartbeat struct {
	Name             string
	Kind             string
	Runtime          *proxy.BackendRuntime
	Path             string
	Timeout          time.Duration
	Period           time.Duration
	InitialDelay     time.Duration
	FailureThreshold uint64
	SuccessThreshold uint64
	Logger           logger.Logger
	FailureCount     atomic.Uint64
	SuccessCount     atomic.Uint64
	stopCh           chan struct{}
	stopOnce         sync.Once
	Callbacks        Callbacks
}

type Callbacks struct {
	OnSuccess func(be *proxy.BackendRuntime) error
	OnFailure func(be *proxy.BackendRuntime, err error) error
}

func (c *Controller) runHeartbeat(ctx context.Context) {
	for _, backend := range c.cache.Backends {
		c.runHealthProbe(
			ctx,
			backend.GetMeta().GetName(),
			backend.GetConfig().GetProbes().GetHealthiness(),
			&Callbacks{
				OnSuccess: c.setTargetHealthy,
				OnFailure: c.setTargetUnhealthy,
			},
			"health",
		)
		c.runHealthProbe(
			ctx,
			backend.GetMeta().GetName(),
			backend.GetConfig().GetProbes().GetReadiness(),
			&Callbacks{
				OnSuccess: c.setTargetReady,
				OnFailure: c.setTargetNotReady,
			},
			"ready",
		)
	}
	<-ctx.Done()
}

func (c *Controller) NewHeartbeat(name, kind string, be *backendv1.Probe, target *proxy.BackendRuntime, cb *Callbacks) *Heartbeat {
	initialDelaySeconds := be.GetInitialDelaySeconds()
	periodSeconds := int(be.GetPeriodSeconds())
	timeoutSeconds := int(be.GetTimeoutSeconds())
	failureThreshold := be.GetFailureThreshold()
	successThreshold := be.GetSuccessThreshold()

	return &Heartbeat{
		Name:             name,
		Kind:             kind,
		Timeout:          heartbeatDuration(timeoutSeconds, c.heartBeatTimeout),
		Period:           heartbeatDuration(periodSeconds, c.heartBeatInterval),
		InitialDelay:     time.Duration(initialDelaySeconds) * time.Second,
		Path:             be.GetPath(),
		FailureThreshold: failureThreshold,
		SuccessThreshold: successThreshold,
		Runtime:          target,
		Logger:           c.logger,
		stopCh:           make(chan struct{}),
		Callbacks:        *cb,
	}
}

func (c *Controller) removeHealthProbe(ctx context.Context, be *backendv1.Backend) {
	backend := be.GetMeta().GetName()

	c.mu.Lock()
	probes := c.probes[backend]
	delete(c.probes, backend)
	c.mu.Unlock()

	for _, probe := range probes {
		probe.Stop()
	}
}

func (c *Controller) runHealthProbe(ctx context.Context, name string, be *backendv1.Probe, cb *Callbacks, kind string) {
	if be == nil || cb == nil {
		return
	}

	backend := name
	runtime := c.runtime.Load()

	c.mu.Lock()
	if c.probes[backend] == nil {
		c.probes[backend] = make(map[string]*Heartbeat)
	}
	c.mu.Unlock()

	if pool, ok := runtime.Backends[backend]; ok {
		for _, target := range pool.Targets {

			hb := c.NewHeartbeat(pool.Name, kind, be, target, cb)

			probeURL, err := url.JoinPath(target.URL.String(), hb.Path)
			if err != nil {
				c.logger.Error("build probe key", "error", err, "backend", backend, "target", target.Name)
				continue
			}
			key := kind + ":" + probeURL

			c.mu.Lock()
			c.probes[backend][key] = hb
			c.mu.Unlock()

			go func() {
				if !waitForHeartbeat(ctx, hb.stopCh, hb.InitialDelay) {
					return
				}

				c.runSingleHeartbeat(ctx, backend, target, hb)

				ticker := time.NewTicker(hb.Period)
				defer ticker.Stop()
				for {
					select {
					case <-ticker.C:
						c.runSingleHeartbeat(ctx, backend, target, hb)
					case <-ctx.Done():
						return
					case <-hb.stopCh:
						return
					}
				}
			}()

		}
	}
}

func (h *Heartbeat) Stop() {
	h.stopOnce.Do(func() {
		close(h.stopCh)
	})
}

func (h *Heartbeat) Next(ctx context.Context) error {
	if h.Runtime == nil {
		return errors.New("runtime is nil")
	}
	if h.Timeout <= 0 {
		return errors.New("heartbeat timeout must be greater than zero")
	}

	ctx, cancel := context.WithTimeout(ctx, h.Timeout)
	defer cancel()

	u, err := url.JoinPath(h.Runtime.URL.String(), h.Path)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	resp, err := h.Runtime.Transport.RoundTrip(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		h.Logger.Debug("heartbeat returned non-ok response", "code", resp.StatusCode, "status", resp.Status)
		return fmt.Errorf("heartbeat returned status %s", resp.Status)
	}

	return nil
}

func heartbeatDuration(seconds int, fallback time.Duration) time.Duration {
	if seconds <= 0 {
		return fallback
	}
	return time.Duration(seconds) * time.Second
}

func waitForHeartbeat(ctx context.Context, stopCh <-chan struct{}, delay time.Duration) bool {
	if delay <= 0 {
		return true
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-timer.C:
		return true
	case <-ctx.Done():
		return false
	case <-stopCh:
		return false
	}
}

func (c *Controller) runSingleHeartbeat(ctx context.Context, backend string, target *proxy.BackendRuntime, hb *Heartbeat) {
	err := hb.Next(ctx)
	if err != nil {
		hb.SuccessCount.Store(0)
		failures := hb.FailureCount.Add(1)
		c.logger.Error("heartbeat target", "error", err, "backend", backend, "target", target.Name)
		if failures >= hb.FailureThreshold {
			if err := hb.Callbacks.OnFailure(target, err); err != nil {
				c.logger.Error("set target as unhealthy", "error", err, "backend", backend, "target", target.Name)
			}
		}
		return
	}

	hb.FailureCount.Store(0)
	successes := hb.SuccessCount.Add(1)
	if successes >= hb.SuccessThreshold {
		if err := hb.Callbacks.OnSuccess(target); err != nil {
			c.logger.Error("set target as healthy", "error", err, "backend", backend, "target", target.Name)
		}
	}
}

func (c *Controller) setTargetUnhealthy(be *proxy.BackendRuntime, err error) error {
	key := be.URL.String()
	st := &backendv1.BackendStatus{
		TargetStatuses: map[string]*backendv1.TargetStatus{
			key: {
				Healthiness: &backendv1.TargetHealthStatus{
					IsHealthy:          new(false),
					Reason:             wrapperspb.String(err.Error()),
					LastTransitionTime: timestamppb.Now(),
				},
			},
		},
	}
	return c.setBackendStatus(be, st)
}

func (c *Controller) setTargetHealthy(be *proxy.BackendRuntime) error {
	key := be.URL.String()
	st := &backendv1.BackendStatus{
		TargetStatuses: map[string]*backendv1.TargetStatus{
			key: {
				Healthiness: &backendv1.TargetHealthStatus{
					IsHealthy:          new(true),
					LastTransitionTime: timestamppb.Now(),
				},
			},
		},
	}
	return c.setBackendStatus(be, st)
}

func (c *Controller) setTargetNotReady(be *proxy.BackendRuntime, err error) error {
	key := be.URL.String()
	st := &backendv1.BackendStatus{
		TargetStatuses: map[string]*backendv1.TargetStatus{
			key: {
				Readiness: &backendv1.TargetReadyStatus{
					IsReady:            new(false),
					Reason:             wrapperspb.String(err.Error()),
					LastTransitionTime: timestamppb.Now(),
				},
			},
		},
	}
	return c.setBackendStatus(be, st)
}

func (c *Controller) setTargetReady(be *proxy.BackendRuntime) error {
	key := be.URL.String()
	st := &backendv1.BackendStatus{
		TargetStatuses: map[string]*backendv1.TargetStatus{
			key: {
				Readiness: &backendv1.TargetReadyStatus{
					IsReady:            new(true),
					LastTransitionTime: timestamppb.Now(),
				},
			},
		},
	}
	return c.setBackendStatus(be, st)
}

func (c *Controller) setBackendStatus(be *proxy.BackendRuntime, st *backendv1.BackendStatus) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.heartBeatTimeout)
	defer cancel()
	return c.clientset.BackendV1().UpdateStatus(ctx, be.Name, st, "target_statuses")
}
