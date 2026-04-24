package proxy

import (
	"context"
	"testing"

	"github.com/nakabonne/tstorage"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/noop"
)

func TestInitMetrics_RestoresSeriesRegistryFromDisk(t *testing.T) {
	tempDir := t.TempDir()
	meter := noop.NewMeterProvider().Meter("test")

	storage, err := tstorage.NewStorage(tstorage.WithDataPath(tempDir))
	if err != nil {
		t.Fatalf("new storage: %v", err)
	}

	metrics, err := InitMetrics(meter, storage, tempDir)
	if err != nil {
		t.Fatalf("init metrics: %v", err)
	}

	metrics.RequestsTotal.Inc(context.Background(), 1,
		attribute.String("route", "demo"),
		attribute.String("method", "GET"),
	)

	if err := storage.Close(); err != nil {
		t.Fatalf("close storage: %v", err)
	}

	reopened, err := tstorage.NewStorage(tstorage.WithDataPath(tempDir))
	if err != nil {
		t.Fatalf("reopen storage: %v", err)
	}
	defer func() {
		_ = reopened.Close()
	}()

	restored, err := InitMetrics(meter, reopened, tempDir)
	if err != nil {
		t.Fatalf("re-init metrics: %v", err)
	}

	series, err := restored.Series()
	if err != nil {
		t.Fatalf("series: %v", err)
	}
	if len(series) == 0 {
		t.Fatal("expected restored series after restart")
	}

	var found bool
	for _, entry := range series {
		if entry.Metric != "proxy.http.requests.total" || entry.Kind != kindCounter {
			continue
		}
		for _, label := range entry.Labels {
			if label.Name == "route" && label.Value == "demo" {
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatal("expected restored labeled request series")
	}
}
