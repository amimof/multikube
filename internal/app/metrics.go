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

func (n *MetricsService) MetricsSeries(ctx context.Context) ([]proxy.MetricSeries, error) {
	_ = ctx
	return n.Metrics.Series()
}
