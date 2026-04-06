package controller

import (
	"context"
	"time"

	proxy "github.com/amimof/multikube/pkg/proxyv2"
)

func (c *Controller) runHeartbeat(ctx context.Context, pool *proxy.BackendPool) {
	ticker := time.NewTicker(c.heartBeatInterval)

	select {
	case <-ticker.C:
	case <-ctx.Done():
		return
	}
}

func (c *Controller) heartbeat(ctx context.Context, pool *proxy.BackendPool) {
}
