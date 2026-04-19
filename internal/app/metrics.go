package app

import (
	"context"

	"github.com/amimof/multikube/pkg/events"
	"github.com/amimof/multikube/pkg/logger"
	proxy "github.com/amimof/multikube/pkg/proxyv2"
)

type MetricsService struct {
	Logger   logger.Logger
	Exchange *events.Exchange
	Metrics  *proxy.ProxyMetrics
}

// MetricsLog streams metrics entries
func (n *MetricsService) MetricsLog(ctx context.Context) (*proxy.ProxyMetrics, error) {
	return n.Metrics, nil
}
