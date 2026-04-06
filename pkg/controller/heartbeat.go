package controller

import (
	"context"
	"net/http"
	"net/url"
	"time"

	backendv1 "github.com/amimof/multikube/api/backend/v1"
	proxy "github.com/amimof/multikube/pkg/proxyv2"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	BackendHealthy   = "Healthy"
	BackendUnhealthy = "Unhealthy"
	BackendUnknown   = "Unknown"
)

func (c *Controller) runHeartbeat(ctx context.Context) {
	ticker := time.NewTicker(c.heartBeatInterval)
	for {
		select {
		case <-ticker.C:
			if err := c.heartbeat(); err != nil {
				c.logger.Error("heartbeat failed: %w", err)
			}
		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}

func (c *Controller) heartbeat() error {
	runtime := c.runtime.Load()

	for _, pool := range runtime.Backends {
		for _, target := range pool.Targets {
			go func(br *proxy.BackendRuntime) {
				if err := c.heartbeatSingle(br); err != nil {
					c.logger.Error("heartbeat failed for target %s: %w", target.URL, err)
				}
			}(target)
		}
	}

	return nil
}

func (c *Controller) heartbeatSingle(be *proxy.BackendRuntime) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.heartBeatTimeout)
	defer cancel()

	u, err := url.JoinPath(be.URL.String(), "/api/v1/")
	if err != nil {
		_ = c.setTargetUnhealthy(be, err.Error())
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		_ = c.setTargetUnhealthy(be, err.Error())
		return err
	}
	resp, err := be.Transport.RoundTrip(req)
	if err != nil {
		_ = c.setTargetUnhealthy(be, err.Error())
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		c.logger.Debug("heartbeat returned non-ok response: %d - %s", resp.StatusCode, resp.Status)
		_ = c.setTargetUnhealthy(be, resp.Status)
		return nil
	}
	return c.setTargetHealthy(be)
}

func (c *Controller) setTargetUnhealthy(be *proxy.BackendRuntime, reason string) error {
	key := be.URL.String()
	st := &backendv1.BackendStatus{
		TargetStatuses: map[string]*backendv1.TargetStatus{
			key: {
				Phase:              wrapperspb.String(BackendUnhealthy),
				Reason:             wrapperspb.String(reason),
				LastTransitionTime: timestamppb.Now(),
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
				Phase:              wrapperspb.String(BackendHealthy),
				Reason:             wrapperspb.String(""),
				LastTransitionTime: timestamppb.Now(),
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
