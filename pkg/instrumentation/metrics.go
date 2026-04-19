package instrumentation

import (
	"context"

	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"google.golang.org/grpc"
	"google.golang.org/grpc/stats/opentelemetry"
)

func InitServerMetrics(ctx context.Context) (grpc.ServerOption, *metric.MeterProvider, error) {
	exporter, err := prometheus.New()
	if err != nil {
		return nil, nil, err
	}

	meterProvider := metric.NewMeterProvider(metric.WithReader(exporter))

	opts := opentelemetry.ServerOption(opentelemetry.Options{
		MetricsOptions: opentelemetry.MetricsOptions{
			MeterProvider: meterProvider,
		},
	})

	return opts, meterProvider, nil
}

func InitClientMetrics() (grpc.DialOption, *metric.MeterProvider, error) {
	exporter, err := prometheus.New()
	if err != nil {
		return nil, nil, err
	}
	meterProvider := metric.NewMeterProvider(metric.WithReader(exporter))

	opts := opentelemetry.DialOption(opentelemetry.Options{
		MetricsOptions: opentelemetry.MetricsOptions{
			MeterProvider: meterProvider,
		},
	})

	return opts, meterProvider, nil
}
